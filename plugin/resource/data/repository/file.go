package repository

import (
	"context"
	"fmt"
	"ncobase/resource/data"
	"ncobase/resource/data/ent"
	fileEnt "ncobase/resource/data/ent/file"
	"ncobase/resource/structs"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/data/search/meili"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/redis/go-redis/v9"
)

// FileRepositoryInterface defines file repository methods
type FileRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateFileBody) (*ent.File, error)
	GetByID(ctx context.Context, slug string) (*ent.File, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*ent.File, error)
	Delete(ctx context.Context, slug string) error
	List(ctx context.Context, params *structs.ListFileParams) ([]*ent.File, error)
	CountX(ctx context.Context, params *structs.ListFileParams) int
	SumSizeBySpace(ctx context.Context, spaceID string) (int64, error)
	GetAllSpaces(ctx context.Context) ([]string, error)
	SearchByTags(ctx context.Context, spaceID string, tags []string, limit int) ([]*ent.File, error)
}

type fileRepository struct {
	ec  *ent.Client
	ecr *ent.Client
	rc  *redis.Client
	ms  *meili.Client
	c   *cache.Cache[ent.File]
}

// NewFileRepository creates new file repository
func NewFileRepository(d *data.Data) FileRepositoryInterface {
	ec := d.GetMasterEntClient()
	ecr := d.GetSlaveEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &fileRepository{ec, ecr, rc, ms, cache.NewCache[ent.File](rc, "ncse_file")}
}

// Create creates a file
func (r *fileRepository) Create(ctx context.Context, body *structs.CreateFileBody) (*ent.File, error) {
	builder := r.ec.File.Create()

	builder.SetNillableName(&body.Name)
	builder.SetNillablePath(&body.Path)
	builder.SetNillableType(&body.Type)
	builder.SetNillableSize(body.Size)
	builder.SetNillableStorage(&body.Storage)
	builder.SetNillableBucket(&body.Bucket)
	builder.SetNillableEndpoint(&body.Endpoint)
	builder.SetNillableOwnerID(&body.OwnerID)
	builder.SetNillableSpaceID(&body.SpaceID)
	builder.SetNillableCreatedBy(body.CreatedBy)

	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "fileRepo.Create error: %v", err)
		return nil, err
	}

	// Index in Meilisearch
	if err = r.ms.IndexDocuments("files", row); err != nil {
		logger.Errorf(ctx, "fileRepo.Create index error: %v", err)
	}

	return row, nil
}

// GetByID gets file by ID
func (r *fileRepository) GetByID(ctx context.Context, slug string) (*ent.File, error) {
	cacheKey := fmt.Sprintf("%s", slug)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	row, err := r.FindFile(ctx, &structs.FindFile{File: slug})
	if err != nil {
		logger.Errorf(ctx, "fileRepo.GetByID error: %v", err)
		return nil, err
	}

	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "fileRepo.GetByID cache error: %v", err)
	}

	return row, nil
}

// Update updates file by ID
func (r *fileRepository) Update(ctx context.Context, slug string, updates types.JSON) (*ent.File, error) {
	file, err := r.FindFile(ctx, &structs.FindFile{File: slug})
	if err != nil {
		return nil, err
	}

	builder := r.ec.File.UpdateOne(file)

	for field, value := range updates {
		switch field {
		case "name":
			builder.SetNillableName(convert.ToPointer(value.(string)))
		case "path":
			builder.SetNillablePath(convert.ToPointer(value.(string)))
		case "type":
			builder.SetNillableType(convert.ToPointer(value.(string)))
		case "size":
			builder.SetNillableSize(convert.ToPointer(value.(int)))
		case "storage":
			builder.SetNillableStorage(convert.ToPointer(value.(string)))
		case "endpoint":
			builder.SetNillableEndpoint(convert.ToPointer(value.(string)))
		case "owner_id":
			builder.SetNillableOwnerID(convert.ToPointer(value.(string)))
		case "space_id":
			builder.SetNillableSpaceID(convert.ToPointer(value.(string)))
		case "extras":
			builder.SetExtras(value.(types.JSON))
		case "updated_by":
			builder.SetNillableUpdatedBy(convert.ToPointer(value.(string)))
		}
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "fileRepo.Update error: %v", err)
		return nil, err
	}

	// Remove from cache
	cacheKey := fmt.Sprintf("%s", file.ID)
	if err = r.c.Delete(ctx, cacheKey); err != nil {
		logger.Errorf(ctx, "fileRepo.Update cache error: %v", err)
	}

	// Update in Meilisearch
	if err = r.ms.UpdateDocuments("files", row, row.ID); err != nil {
		logger.Errorf(ctx, "fileRepo.Update index error: %v", err)
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
	cacheKey := fmt.Sprintf("%s", file.ID)
	if err = r.c.Delete(ctx, cacheKey); err != nil {
		logger.Errorf(ctx, "fileRepo.Delete cache error: %v", err)
	}

	// Delete from Meilisearch
	if err = r.ms.DeleteDocuments("files", file.ID); err != nil {
		logger.Errorf(ctx, "fileRepo.Delete index error: %v", err)
	}

	return nil
}

// FindFile finds a file
func (r *fileRepository) FindFile(ctx context.Context, params *structs.FindFile) (*ent.File, error) {
	builder := r.ecr.File.Query()

	if validator.IsNotEmpty(params.File) {
		builder = builder.Where(fileEnt.Or(
			fileEnt.IDEQ(params.File),
			fileEnt.NameEQ(params.File),
		))
	}

	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// List gets list of files
func (r *fileRepository) List(ctx context.Context, params *structs.ListFileParams) ([]*ent.File, error) {
	builder, err := r.ListBuilder(ctx, params)
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

// ListBuilder creates list builder
func (r *fileRepository) ListBuilder(ctx context.Context, params *structs.ListFileParams) (*ent.FileQuery, error) {
	builder := r.ecr.File.Query()

	// Filter by space
	if params.SpaceID != "" {
		builder = builder.Where(fileEnt.SpaceIDEQ(params.SpaceID))
	}

	// Filter by user
	if params.User != "" {
		builder = builder.Where(fileEnt.CreatedByEQ(params.User))
	}

	// Filter by owner
	if params.OwnerID != "" {
		builder = builder.Where(fileEnt.OwnerIDEQ(params.OwnerID))
	}

	// Filter by type
	if params.Type != "" {
		builder = builder.Where(fileEnt.TypeContains(params.Type))
	}

	// Filter by storage
	if params.Storage != "" {
		builder = builder.Where(fileEnt.StorageEQ(params.Storage))
	}

	return builder, nil
}

// CountX counts files
func (r *fileRepository) CountX(ctx context.Context, params *structs.ListFileParams) int {
	builder, err := r.ListBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// SumSizeBySpace calculates total storage used by a space
func (r *fileRepository) SumSizeBySpace(ctx context.Context, spaceID string) (int64, error) {
	sum, err := r.ecr.File.Query().
		Where(fileEnt.SpaceIDEQ(spaceID)).
		Aggregate(
			ent.Sum(fileEnt.FieldSize),
		).Int(ctx)

	if err != nil {
		logger.Errorf(ctx, "Error calculating storage usage for space %s: %v", spaceID, err)
		return 0, err
	}

	return int64(sum), nil
}

// GetAllSpaces returns list of all space IDs with files
func (r *fileRepository) GetAllSpaces(ctx context.Context) ([]string, error) {
	spaces, err := r.ecr.File.Query().
		GroupBy(fileEnt.FieldSpaceID).
		Strings(ctx)

	if err != nil {
		logger.Errorf(ctx, "Error getting spaces: %v", err)
		return nil, err
	}

	return spaces, nil
}

// SearchByTags searches files by tags
func (r *fileRepository) SearchByTags(ctx context.Context, spaceID string, tags []string, limit int) ([]*ent.File, error) {
	// Get all files for space
	files, err := r.ecr.File.Query().
		Where(fileEnt.SpaceIDEQ(spaceID)).
		Order(ent.Desc(fileEnt.FieldCreatedAt)).
		Limit(limit * 10). // Get more than needed to filter
		All(ctx)

	if err != nil {
		logger.Errorf(ctx, "Error searching files by tags: %v", err)
		return nil, err
	}

	// Filter by tags
	filtered := make([]*ent.File, 0)
	for _, file := range files {
		fileTags, ok := file.Extras["tags"].([]any)
		if !ok {
			continue
		}

		// Convert to strings
		tagStrings := make([]string, 0)
		for _, tag := range fileTags {
			if tagStr, ok := tag.(string); ok {
				tagStrings = append(tagStrings, tagStr)
			}
		}

		// Check if file has any requested tags
		for _, searchTag := range tags {
			for _, fileTag := range tagStrings {
				if searchTag == fileTag {
					filtered = append(filtered, file)
					break
				}
			}
			if len(filtered) == limit {
				break
			}
		}

		if len(filtered) == limit {
			break
		}
	}

	return filtered, nil
}
