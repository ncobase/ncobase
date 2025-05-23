package repository

import (
	"context"
	"fmt"
	"ncobase/resource/data"
	"ncobase/resource/data/ent"
	fileEnt "ncobase/resource/data/ent/file"
	"ncobase/resource/structs"
	"time"

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

// FileRepositoryInterface represents the file repository interface.
type FileRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateFileBody) (*ent.File, error)
	GetByID(ctx context.Context, slug string) (*ent.File, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*ent.File, error)
	Delete(ctx context.Context, slug string) error
	List(ctx context.Context, params *structs.ListFileParams) ([]*ent.File, error)
	CountX(ctx context.Context, params *structs.ListFileParams) int
	SumSizeByTenant(ctx context.Context, tenantID string) (int64, error)
	GetAllTenants(ctx context.Context) ([]string, error)
	SearchByTags(ctx context.Context, tenantID string, tags []string, limit int) ([]*ent.File, error)
	GetByFolderPath(ctx context.Context, tenantID string, folderPath string) ([]*ent.File, error)
	GetByCategory(ctx context.Context, tenantID string, category string) ([]*ent.File, error)
	GetExpiredFiles(ctx context.Context) ([]*ent.File, error)
}

// fileRepostory implements the FileRepositoryInterface.
type fileRepostory struct {
	ec  *ent.Client
	ecr *ent.Client
	rc  *redis.Client
	ms  *meili.Client
	c   *cache.Cache[ent.File]
}

// NewFileRepository creates a new file repository.
func NewFileRepository(d *data.Data) FileRepositoryInterface {
	ec := d.GetMasterEntClient()
	ecr := d.GetSlaveEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &fileRepostory{ec, ecr, rc, ms, cache.NewCache[ent.File](rc, "ncse_file")}
}

// Create creates an file.
func (r *fileRepostory) Create(ctx context.Context, body *structs.CreateFileBody) (*ent.File, error) {

	// create builder.
	builder := r.ec.File.Create()
	// set values.

	builder.SetNillableName(&body.Name)
	builder.SetNillablePath(&body.Path)
	builder.SetNillableType(&body.Type)
	builder.SetNillableSize(body.Size)
	builder.SetNillableStorage(&body.Storage)
	builder.SetNillableBucket(&body.Bucket)
	builder.SetNillableEndpoint(&body.Endpoint)
	builder.SetNillableObjectID(&body.ObjectID)
	builder.SetNillableTenantID(&body.TenantID)
	builder.SetNillableCreatedBy(body.CreatedBy)

	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "fileRepo.Create error: %v", err)
		return nil, err
	}

	// create the file in Meilisearch index
	if err = r.ms.IndexDocuments("files", row); err != nil {
		logger.Errorf(ctx, "fileRepo.Create index error: %v", err)
		// return nil, err
	}

	return row, nil
}

// GetByID gets an file by ID.
func (r *fileRepostory) GetByID(ctx context.Context, slug string) (*ent.File, error) {
	// check cache
	cacheKey := fmt.Sprintf("%s", slug)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindFile(ctx, &structs.FindFile{File: slug})
	if err != nil {
		logger.Errorf(ctx, "fileRepo.GetByID error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "fileRepo.GetByID cache error: %v", err)
	}

	return row, nil
}

// Update updates an file by ID.
func (r *fileRepostory) Update(ctx context.Context, slug string, updates types.JSON) (*ent.File, error) {
	file, err := r.FindFile(ctx, &structs.FindFile{File: slug})
	if err != nil {
		return nil, err
	}

	// create builder.
	builder := r.ec.File.UpdateOne(file)

	// set values
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
		case "object_id":
			builder.SetNillableObjectID(convert.ToPointer(value.(string)))
		case "tenant_id":
			builder.SetNillableTenantID(convert.ToPointer(value.(string)))
		case "extras":
			builder.SetExtras(value.(types.JSON))
		case "updated_by":
			builder.SetNillableUpdatedBy(convert.ToPointer(value.(string)))
		}
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "fileRepo.Update error: %v", err)
		return nil, err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", file.ID)
	if err = r.c.Delete(ctx, cacheKey); err != nil {
		logger.Errorf(ctx, "fileRepo.Update cache error: %v", err)
	}

	// delete from Meilisearch index
	if err = r.ms.DeleteDocuments("files", file.ID); err != nil {
		logger.Errorf(ctx, "fileRepo.Update index error: %v", err)
		// return nil, err
	}

	return row, nil
}

// Delete deletes an file by ID.
func (r *fileRepostory) Delete(ctx context.Context, slug string) error {
	file, err := r.FindFile(ctx, &structs.FindFile{File: slug})
	if err != nil {
		return err
	}

	// create builder.
	builder := r.ec.File.Delete()

	// execute the builder and verify the result.
	if _, err = builder.Where(fileEnt.IDEQ(slug)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "fileRepo.Delete error: %v", err)
		return err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", file.ID)
	if err = r.c.Delete(ctx, cacheKey); err != nil {
		logger.Errorf(ctx, "fileRepo.Delete cache error: %v", err)
	}

	// delete from Meilisearch index
	if err = r.ms.DeleteDocuments("files", file.ID); err != nil {
		logger.Errorf(ctx, "fileRepo.Delete index error: %v", err)
		// return nil, err
	}

	return nil
}

// FindFile finds an file.
func (r *fileRepostory) FindFile(ctx context.Context, params *structs.FindFile) (*ent.File, error) {
	// create builder.
	builder := r.ecr.File.Query()

	if validator.IsNotEmpty(params.File) {
		builder = builder.Where(fileEnt.Or(
			fileEnt.IDEQ(params.File),
			fileEnt.NameEQ(params.File),
		))
	}

	// execute the builder.
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// List gets a list of files.
func (r *fileRepostory) List(ctx context.Context, params *structs.ListFileParams) ([]*ent.File, error) {
	// create list builder
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

	// execute the builder.
	rows, err := builder.All(ctx)
	if validator.IsNotNil(err) {
		logger.Errorf(ctx, "fileRepo.List error: %v", err)
		return nil, err
	}

	return rows, nil
}

// ListBuilder creates list builder.
func (r *fileRepostory) ListBuilder(ctx context.Context, params *structs.ListFileParams) (*ent.FileQuery, error) {
	// create builder.
	builder := r.ecr.File.Query()

	// belong tenant
	if params.Tenant != "" {
		builder = builder.Where(fileEnt.TenantIDEQ(params.Tenant))
	}

	// belong user
	if params.User != "" {
		builder = builder.Where(fileEnt.CreatedByEQ(params.User))
	}

	// object id
	if params.Object != "" {
		builder = builder.Where(fileEnt.ObjectIDEQ(params.Object))
	}

	// file type
	if params.Type != "" {
		builder = builder.Where(fileEnt.TypeContains(params.Type))
	}

	// storage provider
	if params.Storage != "" {
		builder = builder.Where(fileEnt.StorageEQ(params.Storage))
	}

	return builder, nil
}

// CountX counts files based on given parameters.
func (r *fileRepostory) CountX(ctx context.Context, params *structs.ListFileParams) int {
	// create list builder
	builder, err := r.ListBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// SumSizeByTenant calculates the total storage used by a tenant
func (r *fileRepostory) SumSizeByTenant(ctx context.Context, tenantID string) (int64, error) {
	// Create query
	sum, err := r.ecr.File.Query().
		Where(fileEnt.TenantIDEQ(tenantID)).
		Aggregate(
			ent.Sum(fileEnt.FieldSize),
		).Int(ctx)

	if err != nil {
		logger.Errorf(ctx, "Error calculating storage usage for tenant %s: %v", tenantID, err)
		return 0, err
	}

	return int64(sum), nil
}

// GetAllTenants returns a list of all tenant IDs with files
func (r *fileRepostory) GetAllTenants(ctx context.Context) ([]string, error) {
	tenants, err := r.ecr.File.Query().
		GroupBy(fileEnt.FieldTenantID).
		Strings(ctx)

	if err != nil {
		logger.Errorf(ctx, "Error getting tenants: %v", err)
		return nil, err
	}

	return tenants, nil
}

// SearchByTags searches for files by tags in extras
func (r *fileRepostory) SearchByTags(
	ctx context.Context,
	tenantID string,
	tags []string,
	limit int,
) ([]*ent.File, error) {
	// Since tags are stored in extras JSON field, we can't do a direct query
	// In a real implementation, this would use a more efficient approach or a dedicated tags table

	// Get all files for tenant
	files, err := r.ecr.File.Query().
		Where(fileEnt.TenantIDEQ(tenantID)).
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
		// Check if tags exist
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

		// Check if file has any of the requested tags
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

// GetByFolderPath gets files by folder path
func (r *fileRepostory) GetByFolderPath(
	ctx context.Context,
	tenantID string,
	folderPath string,
) ([]*ent.File, error) {
	// Similar to tags, this requires filtering on JSON field
	// In a real implementation, consider adding a dedicated folder_path column

	// Get all files for tenant
	files, err := r.ecr.File.Query().
		Where(fileEnt.TenantIDEQ(tenantID)).
		Order(ent.Desc(fileEnt.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		logger.Errorf(ctx, "Error getting files by folder path: %v", err)
		return nil, err
	}

	// Filter by folder path
	filtered := make([]*ent.File, 0)
	for _, file := range files {
		// Check if folder path matches
		if path, ok := file.Extras["folder_path"].(string); ok && path == folderPath {
			filtered = append(filtered, file)
		}
	}

	return filtered, nil
}

// GetByCategory gets files by category
func (r *fileRepostory) GetByCategory(
	ctx context.Context,
	tenantID string,
	category string,
) ([]*ent.File, error) {
	// Get all files for tenant
	files, err := r.ecr.File.Query().
		Where(fileEnt.TenantIDEQ(tenantID)).
		Order(ent.Desc(fileEnt.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		logger.Errorf(ctx, "Error getting files by category: %v", err)
		return nil, err
	}

	// Filter by category
	filtered := make([]*ent.File, 0)
	for _, file := range files {
		// Check if metadata exists
		metadata, ok := file.Extras["metadata"].(map[string]any)
		if !ok {
			continue
		}

		// Check if category matches
		if cat, ok := metadata["category"].(string); ok && cat == category {
			filtered = append(filtered, file)
		}
	}

	return filtered, nil
}

// GetExpiredFiles gets files that have expired
func (r *fileRepostory) GetExpiredFiles(ctx context.Context) ([]*ent.File, error) {
	// Get all files with extras
	files, err := r.ecr.File.Query().
		Order(ent.Desc(fileEnt.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		logger.Errorf(ctx, "Error getting expired files: %v", err)
		return nil, err
	}

	// Current timestamp
	now := time.Now().Unix()

	// Filter expired files
	expired := make([]*ent.File, 0)
	for _, file := range files {
		// Check if expires_at exists and is in the past
		if exp, ok := file.Extras["expires_at"].(float64); ok && int64(exp) < now {
			expired = append(expired, file)
		} else if exp, ok := file.Extras["expires_at"].(int64); ok && exp < now {
			expired = append(expired, file)
		}
	}

	return expired, nil
}
