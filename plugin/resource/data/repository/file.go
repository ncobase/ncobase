package repository

import (
	"context"
	"fmt"
	"ncobase/resource/data"
	"ncobase/resource/data/ent"
	fileEnt "ncobase/resource/data/ent/file"
	"ncobase/resource/structs"
	"strings"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/data/search/meili"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/redis/go-redis/v9"
)

type FileRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateFileBody) (*ent.File, error)
	GetByID(ctx context.Context, slug string) (*ent.File, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*ent.File, error)
	Delete(ctx context.Context, slug string) error
	List(ctx context.Context, params *structs.ListFileParams) ([]*ent.File, error)
	CountX(ctx context.Context, params *structs.ListFileParams) int
	SumSizeByOwner(ctx context.Context, ownerID string) (int64, error)
	GetAllOwners(ctx context.Context) ([]string, error)
	SearchByTags(ctx context.Context, ownerID string, tags []string, limit int) ([]*ent.File, error)
	GetTagsByOwner(ctx context.Context, ownerID string) ([]string, error)
	CheckNameExists(ctx context.Context, ownerID, name string) (bool, error)
}

type fileRepository struct {
	ec  *ent.Client
	ecr *ent.Client
	rc  *redis.Client
	ms  *meili.Client
	c   *cache.Cache[ent.File]
}

func NewFileRepository(d *data.Data) FileRepositoryInterface {
	ec := d.GetMasterEntClient()
	ecr := d.GetSlaveEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &fileRepository{ec, ecr, rc, ms, cache.NewCache[ent.File](rc, "ncse_file")}
}

// Create creates a file with complete field mapping
func (r *fileRepository) Create(ctx context.Context, body *structs.CreateFileBody) (*ent.File, error) {
	builder := r.ec.File.Create()

	// Set all basic fields with proper nil handling
	if body.Name != "" {
		builder.SetName(body.Name)
	}

	if body.OriginalName != "" {
		builder.SetOriginalName(body.OriginalName)
	}

	if body.Path != "" {
		builder.SetPath(body.Path)
	}

	if body.Type != "" {
		builder.SetType(body.Type)
	}

	if body.Size != nil {
		builder.SetSize(*body.Size)
	}

	if body.Storage != "" {
		builder.SetStorage(body.Storage)
	}

	if body.Bucket != "" {
		builder.SetBucket(body.Bucket)
	}

	if body.Endpoint != "" {
		builder.SetEndpoint(body.Endpoint)
	}

	if body.OwnerID != "" {
		builder.SetOwnerID(body.OwnerID)
	}

	// Set access level (with default)
	if body.AccessLevel != "" {
		builder.SetAccessLevel(string(body.AccessLevel))
	} else {
		builder.SetAccessLevel(string(structs.AccessLevelPrivate)) // Default
	}

	// Set expiration
	if body.ExpiresAt != nil {
		builder.SetExpiresAt(*body.ExpiresAt)
	}

	// Set tags
	if len(body.Tags) > 0 {
		builder.SetTags(body.Tags)
	}

	// Set public flag
	builder.SetIsPublic(body.IsPublic)

	// Set category based on file extension if not provided
	category := structs.GetFileCategory(body.Path)
	builder.SetCategory(string(category))

	// Set audit fields
	if body.CreatedBy != nil && *body.CreatedBy != "" {
		builder.SetCreatedBy(*body.CreatedBy)
	}

	if body.UpdatedBy != nil && *body.UpdatedBy != "" {
		builder.SetUpdatedBy(*body.UpdatedBy)
	}

	// Set extras with complete metadata
	extras := make(types.JSON)
	if body.Extras != nil {
		for k, v := range *body.Extras {
			extras[k] = v
		}
	}

	// Add processing options to extras
	if body.ProcessingOptions != nil {
		extras["processing_options"] = body.ProcessingOptions
	}

	// Add additional metadata
	if body.PathPrefix != "" {
		extras["path_prefix"] = body.PathPrefix
	}

	// Set computed hash if available
	if hashValue, ok := extras["hash"].(string); ok && hashValue != "" {
		builder.SetHash(hashValue)
	}

	// Set the complete extras
	if len(extras) > 0 {
		builder.SetExtras(extras)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "fileRepo.Create error: %v", err)
		return nil, err
	}

	// Index in Meilisearch
	if r.ms != nil {
		if err = r.ms.IndexDocuments("files", row); err != nil {
			logger.Errorf(ctx, "fileRepo.Create index error: %v", err)
		}
	}

	return row, nil
}

// Update updates file by ID with complete field mapping
func (r *fileRepository) Update(ctx context.Context, slug string, updates types.JSON) (*ent.File, error) {
	file, err := r.FindFile(ctx, &structs.FindFile{File: slug})
	if err != nil {
		return nil, err
	}

	builder := r.ec.File.UpdateOne(file)

	for field, value := range updates {
		switch field {
		case "name":
			if v, ok := value.(string); ok && v != "" {
				builder.SetName(v)
			}
		case "original_name":
			if v, ok := value.(string); ok {
				builder.SetOriginalName(v)
			}
		case "path":
			if v, ok := value.(string); ok && v != "" {
				builder.SetPath(v)
			}
		case "type":
			if v, ok := value.(string); ok && v != "" {
				builder.SetType(v)
			}
		case "size":
			if v, ok := value.(int); ok {
				builder.SetSize(v)
			}
		case "storage":
			if v, ok := value.(string); ok && v != "" {
				builder.SetStorage(v)
			}
		case "bucket":
			if v, ok := value.(string); ok {
				builder.SetBucket(v)
			}
		case "endpoint":
			if v, ok := value.(string); ok {
				builder.SetEndpoint(v)
			}
		case "owner_id":
			if v, ok := value.(string); ok && v != "" {
				builder.SetOwnerID(v)
			}
		case "access_level":
			if v, ok := value.(structs.AccessLevel); ok {
				builder.SetAccessLevel(string(v))
			} else if v, ok := value.(string); ok && v != "" {
				builder.SetAccessLevel(v)
			}
		case "expires_at":
			if v, ok := value.(int64); ok {
				builder.SetExpiresAt(v)
			} else if v, ok := value.(*int64); ok && v != nil {
				builder.SetExpiresAt(*v)
			}
		case "tags":
			if v, ok := value.([]string); ok {
				builder.SetTags(v)
			}
		case "is_public":
			if v, ok := value.(bool); ok {
				builder.SetIsPublic(v)
			}
		case "category":
			if v, ok := value.(structs.FileCategory); ok {
				builder.SetCategory(string(v))
			} else if v, ok := value.(string); ok && v != "" {
				builder.SetCategory(v)
			}
		case "hash":
			if v, ok := value.(string); ok {
				builder.SetHash(v)
			}
		case "processing_result":
			if v, ok := value.(types.JSON); ok {
				builder.SetProcessingResult(v)
			}
		case "extras":
			if v, ok := value.(types.JSON); ok {
				builder.SetExtras(v)
			}
		case "created_by":
			if v, ok := value.(string); ok {
				builder.SetCreatedBy(v)
			} else if v, ok := value.(*string); ok && v != nil {
				builder.SetCreatedBy(*v)
			}
		case "updated_by":
			if v, ok := value.(string); ok {
				builder.SetUpdatedBy(v)
			} else if v, ok := value.(*string); ok && v != nil {
				builder.SetUpdatedBy(*v)
			}
		case "updated_at":
			if v, ok := value.(int64); ok {
				builder.SetUpdatedAt(v)
			}
		default:
			// Handle unknown fields by adding them to extras
			logger.Warnf(ctx, "Unknown field in update: %s", field)
		}
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "fileRepo.Update error: %v", err)
		return nil, err
	}

	// Remove from cache
	if r.c != nil {
		cacheKey := fmt.Sprintf("%s", file.ID)
		if err = r.c.Delete(ctx, cacheKey); err != nil {
			logger.Errorf(ctx, "fileRepo.Update cache error: %v", err)
		}
	}

	// Update in Meilisearch
	if r.ms != nil {
		if err = r.ms.UpdateDocuments("files", row, row.ID); err != nil {
			logger.Errorf(ctx, "fileRepo.Update index error: %v", err)
		}
	}

	return row, nil
}

// CheckNameExists checks if a file name already exists for an owner
func (r *fileRepository) CheckNameExists(ctx context.Context, ownerID, name string) (bool, error) {
	count, err := r.ecr.File.Query().
		Where(
			fileEnt.OwnerIDEQ(ownerID),
			fileEnt.NameEQ(name),
		).
		Count(ctx)

	if err != nil {
		logger.Errorf(ctx, "fileRepo.CheckNameExists error: %v", err)
		return false, err
	}

	return count > 0, nil
}

// GetByID gets file by ID with improved caching
func (r *fileRepository) GetByID(ctx context.Context, slug string) (*ent.File, error) {
	cacheKey := fmt.Sprintf("%s", slug)
	if r.c != nil {
		if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
			return cached, nil
		}
	}

	row, err := r.FindFile(ctx, &structs.FindFile{File: slug})
	if err != nil {
		logger.Errorf(ctx, "fileRepo.GetByID error: %v", err)
		return nil, err
	}

	if r.c != nil {
		err = r.c.Set(ctx, cacheKey, row)
		if err != nil {
			logger.Errorf(ctx, "fileRepo.GetByID cache error: %v", err)
		}
	}

	return row, nil
}

// Delete deletes file by ID
func (r *fileRepository) Delete(ctx context.Context, slug string) error {
	file, err := r.FindFile(ctx, &structs.FindFile{File: slug})
	if err != nil {
		return err
	}

	builder := r.ec.File.Delete()

	if _, err = builder.Where(fileEnt.IDEQ(slug)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "fileRepo.Delete error: %v", err)
		return err
	}

	// Remove from cache
	if r.c != nil {
		cacheKey := fmt.Sprintf("%s", file.ID)
		if err = r.c.Delete(ctx, cacheKey); err != nil {
			logger.Errorf(ctx, "fileRepo.Delete cache error: %v", err)
		}
	}

	// Delete from Meilisearch
	if r.ms != nil {
		if err = r.ms.DeleteDocuments("files", file.ID); err != nil {
			logger.Errorf(ctx, "fileRepo.Delete index error: %v", err)
		}
	}

	return nil
}

// FindFile finds a file with improved query
func (r *fileRepository) FindFile(ctx context.Context, params *structs.FindFile) (*ent.File, error) {
	builder := r.ecr.File.Query()

	if validator.IsNotEmpty(params.File) {
		builder = builder.Where(fileEnt.Or(
			fileEnt.IDEQ(params.File),
			fileEnt.NameEQ(params.File),
		))
	}

	if validator.IsNotEmpty(params.OwnerID) {
		builder = builder.Where(fileEnt.OwnerIDEQ(params.OwnerID))
	}

	if validator.IsNotEmpty(params.User) {
		builder = builder.Where(fileEnt.CreatedByEQ(params.User))
	}

	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// List gets list of files with improved pagination and filtering
func (r *fileRepository) List(ctx context.Context, params *structs.ListFileParams) ([]*ent.File, error) {
	builder, err := r.ListBuilder(params)
	if validator.IsNotNil(err) {
		return nil, err
	}

	if params.Cursor != "" {
		id, timestamp, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %v", err)
		}

		if !nanoid.IsPrimaryKey(id) {
			return nil, fmt.Errorf("invalid id in cursor: %s", id)
		}

		if params.Direction == "backward" {
			builder.Where(
				fileEnt.Or(
					fileEnt.CreatedAtGT(timestamp),
					fileEnt.And(
						fileEnt.CreatedAtEQ(timestamp),
						fileEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				fileEnt.Or(
					fileEnt.CreatedAtLT(timestamp),
					fileEnt.And(
						fileEnt.CreatedAtEQ(timestamp),
						fileEnt.IDLT(id),
					),
				),
			)
		}
	}

	if params.Direction == "backward" {
		builder.Order(ent.Asc(fileEnt.FieldCreatedAt), ent.Asc(fileEnt.FieldID))
	} else {
		builder.Order(ent.Desc(fileEnt.FieldCreatedAt), ent.Desc(fileEnt.FieldID))
	}

	builder.Limit(params.Limit)

	rows, err := builder.All(ctx)
	if validator.IsNotNil(err) {
		logger.Errorf(ctx, "fileRepo.List error: %v", err)
		return nil, err
	}

	return rows, nil
}

// ListBuilder creates list builder with enhanced filtering
func (r *fileRepository) ListBuilder(params *structs.ListFileParams) (*ent.FileQuery, error) {
	builder := r.ecr.File.Query()

	// Filter by owner
	if params.OwnerID != "" {
		builder = builder.Where(fileEnt.OwnerIDEQ(params.OwnerID))
	}

	// Filter by user
	if params.User != "" {
		builder = builder.Where(fileEnt.CreatedByEQ(params.User))
	}

	// Filter by type
	if params.Type != "" {
		builder = builder.Where(fileEnt.TypeContains(params.Type))
	}

	// Filter by storage
	if params.Storage != "" {
		builder = builder.Where(fileEnt.StorageEQ(params.Storage))
	}

	// Filter by category
	if params.Category != "" {
		builder = builder.Where(fileEnt.CategoryEQ(string(params.Category)))
	}

	// Filter by access level
	if params.AccessLevel != "" {
		builder = builder.Where(fileEnt.AccessLevelEQ(string(params.AccessLevel)))
	}

	// Filter by public flag
	if params.IsPublic != nil {
		builder = builder.Where(fileEnt.IsPublicEQ(*params.IsPublic))
	}

	// Filter by path prefix
	if params.PathPrefix != "" {
		builder = builder.Where(fileEnt.PathHasPrefix(params.PathPrefix))
	}

	// Filter by date range
	if params.CreatedAfter > 0 {
		builder = builder.Where(fileEnt.CreatedAtGT(params.CreatedAfter))
	}

	if params.CreatedBefore > 0 {
		builder = builder.Where(fileEnt.CreatedAtLT(params.CreatedBefore))
	}

	// Filter by size range
	if params.SizeMin > 0 {
		builder = builder.Where(fileEnt.SizeGTE(int(params.SizeMin)))
	}

	if params.SizeMax > 0 {
		builder = builder.Where(fileEnt.SizeLTE(int(params.SizeMax)))
	}

	// Search query (name and tags)
	if params.SearchQuery != "" {
		builder = builder.Where(
			fileEnt.Or(
				fileEnt.NameContains(params.SearchQuery),
				fileEnt.OriginalNameContains(params.SearchQuery),
				func(s *sql.Selector) {
					s.Where(sqljson.ValueContains(fileEnt.FieldTags, params.SearchQuery))
				},
			),
		)
	}

	// Filter by tags
	if params.Tags != "" {
		tags := strings.Split(params.Tags, ",")
		cleanTags := make([]string, 0, len(tags))
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				cleanTags = append(cleanTags, tag)
			}
		}
		if len(cleanTags) > 0 {
			builder = builder.Where(func(s *sql.Selector) {
				for _, tag := range cleanTags {
					s.Where(sqljson.ValueContains(fileEnt.FieldTags, tag))
				}
			})
		}
	}

	return builder, nil
}

// CountX counts files
func (r *fileRepository) CountX(ctx context.Context, params *structs.ListFileParams) int {
	builder, err := r.ListBuilder(params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// SumSizeByOwner calculates total storage used by an owner
func (r *fileRepository) SumSizeByOwner(ctx context.Context, ownerID string) (int64, error) {
	builder := r.ecr.File.Query()

	if ownerID != "" {
		builder = builder.Where(fileEnt.OwnerIDEQ(ownerID))
	}

	// Get all files to calculate sum manually
	files, err := builder.Select(fileEnt.FieldSize).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "Error querying files for size calculation for owner %s: %v", ownerID, err)
		return 0, err
	}

	// Calculate total size manually
	var totalSize int64 = 0
	for _, file := range files {
		totalSize += int64(file.Size)
	}

	return totalSize, nil
}

// GetAllOwners gets all unique owners
func (r *fileRepository) GetAllOwners(ctx context.Context) ([]string, error) {
	owners, err := r.ecr.File.Query().
		Select(fileEnt.FieldOwnerID).
		GroupBy(fileEnt.FieldOwnerID).
		Strings(ctx)

	if err != nil {
		logger.Errorf(ctx, "Error getting owners: %v", err)
		return nil, err
	}

	return owners, nil
}

// SearchByTags searches files by tags with improved performance
func (r *fileRepository) SearchByTags(ctx context.Context, ownerID string, tags []string, limit int) ([]*ent.File, error) {
	builder := r.ecr.File.Query().
		Where(fileEnt.OwnerIDEQ(ownerID)).
		Order(ent.Desc(fileEnt.FieldCreatedAt)).
		Limit(limit)

	if len(tags) > 0 {
		builder = builder.Where(func(s *sql.Selector) {
			for _, tag := range tags {
				s.Where(sqljson.ValueContains(fileEnt.FieldTags, tag))
			}
		})
	}

	files, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "Error searching files by tags: %v", err)
		return nil, err
	}

	return files, nil
}

// GetTagsByOwner retrieves all tags used by an owner
func (r *fileRepository) GetTagsByOwner(ctx context.Context, ownerID string) ([]string, error) {
	files, err := r.ecr.File.Query().
		Where(fileEnt.OwnerIDEQ(ownerID)).
		Select(fileEnt.FieldTags).
		All(ctx)

	if err != nil {
		logger.Errorf(ctx, "Error getting files for tags: %v", err)
		return nil, err
	}

	tagSet := make(map[string]bool)
	for _, file := range files {
		for _, tag := range file.Tags {
			if tag != "" {
				tagSet[tag] = true
			}
		}
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	return tags, nil
}
