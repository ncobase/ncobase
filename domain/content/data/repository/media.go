package repository

import (
	"context"
	"fmt"
	"ncobase/domain/content/data"
	"ncobase/domain/content/data/ent"
	mediaEnt "ncobase/domain/content/data/ent/media"
	"ncobase/domain/content/structs"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/redis/go-redis/v9"
)

// MediaRepositoryInterface represents the media repository interface.
type MediaRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateMediaBody) (*ent.Media, error)
	GetByID(ctx context.Context, id string) (*ent.Media, error)
	Update(ctx context.Context, id string, updates types.JSON) (*ent.Media, error)
	List(ctx context.Context, params *structs.ListMediaParams) ([]*ent.Media, error)
	Count(ctx context.Context, params *structs.ListMediaParams) (int, error)
	ListWithCount(ctx context.Context, params *structs.ListMediaParams) ([]*ent.Media, int, error)
	Delete(ctx context.Context, id string) error
	FindMedia(ctx context.Context, params *structs.FindMedia) (*ent.Media, error)
}

// mediaRepository implements the MediaRepositoryInterface.
type mediaRepository struct {
	ec  *ent.Client
	ecr *ent.Client
	rc  *redis.Client
	c   *cache.Cache[ent.Media]
}

// NewMediaRepository creates a new media repository.
func NewMediaRepository(d *data.Data) MediaRepositoryInterface {
	ec := d.GetEntClient()
	ecr := d.GetEntClientRead()
	rc := d.GetRedis()
	return &mediaRepository{ec, ecr, rc, cache.NewCache[ent.Media](rc, "ncse_media")}
}

// Create creates a new media.
func (r *mediaRepository) Create(ctx context.Context, body *structs.CreateMediaBody) (*ent.Media, error) {
	// create builder
	builder := r.ec.Media.Create()

	// set values
	builder.SetNillableTitle(&body.Title)
	builder.SetNillableType(&body.Type)
	builder.SetNillableURL(&body.URL)
	builder.SetNillablePath(&body.Path)
	builder.SetNillableMimeType(&body.MimeType)
	builder.SetSize(body.Size)
	builder.SetWidth(body.Width)
	builder.SetHeight(body.Height)
	builder.SetDuration(body.Duration)
	builder.SetNillableDescription(&body.Description)
	builder.SetNillableAlt(&body.Alt)

	if !validator.IsNil(body.Metadata) && !validator.IsEmpty(body.Metadata) {
		builder.SetExtras(*body.Metadata)
	}

	builder.SetNillableTenantID(&body.TenantID)
	builder.SetNillableCreatedBy(body.CreatedBy)

	// execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "mediaRepo.Create error: %v", err)
		return nil, err
	}

	return row, nil
}

// GetByID gets a media by ID.
func (r *mediaRepository) GetByID(ctx context.Context, id string) (*ent.Media, error) {
	// check cache
	cacheKey := fmt.Sprintf("%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindMedia(ctx, &structs.FindMedia{Media: id})
	if err != nil {
		logger.Errorf(ctx, "mediaRepo.GetByID error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "mediaRepo.GetByID cache error: %v", err)
	}

	return row, nil
}

// Update updates a media.
func (r *mediaRepository) Update(ctx context.Context, id string, updates types.JSON) (*ent.Media, error) {
	media, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// create builder
	builder := media.Update()

	// set values based on the updates map
	for field, value := range updates {
		switch field {
		case "title":
			builder.SetNillableTitle(convert.ToPointer(value.(string)))
		case "type":
			builder.SetNillableType(convert.ToPointer(value.(string)))
		case "url":
			builder.SetNillableURL(convert.ToPointer(value.(string)))
		case "path":
			builder.SetNillablePath(convert.ToPointer(value.(string)))
		case "mime_type":
			builder.SetNillableMimeType(convert.ToPointer(value.(string)))
		case "size":
			builder.SetSize(value.(int64))
		case "width":
			builder.SetWidth(value.(int))
		case "height":
			builder.SetHeight(value.(int))
		case "duration":
			builder.SetDuration(value.(float64))
		case "description":
			builder.SetNillableDescription(convert.ToPointer(value.(string)))
		case "alt":
			builder.SetNillableAlt(convert.ToPointer(value.(string)))
		case "metadata":
			builder.SetExtras(value.(types.JSON))
		case "tenant_id":
			builder.SetNillableTenantID(convert.ToPointer(value.(string)))
		case "updated_by":
			builder.SetNillableUpdatedBy(convert.ToPointer(value.(string)))
		}
	}

	// execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "mediaRepo.Update error: %v", err)
		return nil, err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", media.ID)
	err = r.c.Delete(ctx, cacheKey)
	if err != nil {
		logger.Errorf(ctx, "mediaRepo.Update cache error: %v", err)
	}

	return row, nil
}

// List gets a list of media.
func (r *mediaRepository) List(ctx context.Context, params *structs.ListMediaParams) ([]*ent.Media, error) {
	// create builder
	builder := r.ecr.Media.Query()

	// apply filters
	if validator.IsNotEmpty(params.Type) {
		builder.Where(mediaEnt.TypeEQ(params.Type))
	}

	if validator.IsNotEmpty(params.Search) {
		builder.Where(mediaEnt.Or(
			mediaEnt.TitleContains(params.Search),
			mediaEnt.DescriptionContains(params.Search),
			mediaEnt.AltContains(params.Search),
		))
	}

	if validator.IsNotEmpty(params.Tenant) {
		builder.Where(mediaEnt.TenantIDEQ(params.Tenant))
	}

	// apply cursor-based pagination
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
				mediaEnt.Or(
					mediaEnt.CreatedAtGT(timestamp),
					mediaEnt.And(
						mediaEnt.CreatedAtEQ(timestamp),
						mediaEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				mediaEnt.Or(
					mediaEnt.CreatedAtLT(timestamp),
					mediaEnt.And(
						mediaEnt.CreatedAtEQ(timestamp),
						mediaEnt.IDLT(id),
					),
				),
			)
		}
	}

	// set ordering
	if params.Direction == "backward" {
		builder.Order(ent.Asc(mediaEnt.FieldCreatedAt), ent.Asc(mediaEnt.FieldID))
	} else {
		builder.Order(ent.Desc(mediaEnt.FieldCreatedAt), ent.Desc(mediaEnt.FieldID))
	}

	// set limit
	if params.Limit > 0 {
		builder.Limit(params.Limit)
	} else {
		builder.Limit(10) // default limit
	}

	// execute query
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "mediaRepo.List error: %v", err)
		return nil, err
	}

	return rows, nil
}

// Count gets a count of media.
func (r *mediaRepository) Count(ctx context.Context, params *structs.ListMediaParams) (int, error) {
	// create builder
	builder := r.ecr.Media.Query()

	// apply filters
	if validator.IsNotEmpty(params.Type) {
		builder.Where(mediaEnt.TypeEQ(params.Type))
	}

	if validator.IsNotEmpty(params.Search) {
		builder.Where(mediaEnt.Or(
			mediaEnt.TitleContains(params.Search),
			mediaEnt.DescriptionContains(params.Search),
			mediaEnt.AltContains(params.Search),
		))
	}

	if validator.IsNotEmpty(params.Tenant) {
		builder.Where(mediaEnt.TenantIDEQ(params.Tenant))
	}

	// execute count query
	return builder.Count(ctx)
}

// ListWithCount gets a list of media and their total count.
func (r *mediaRepository) ListWithCount(ctx context.Context, params *structs.ListMediaParams) ([]*ent.Media, int, error) {
	// Get count first
	count, err := r.Count(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	// Then get list
	rows, err := r.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	return rows, count, nil
}

// Delete deletes a media.
func (r *mediaRepository) Delete(ctx context.Context, id string) error {
	// create builder
	builder := r.ec.Media.Delete()

	// execute the builder
	_, err := builder.Where(mediaEnt.IDEQ(id)).Exec(ctx)
	if err != nil {
		logger.Errorf(ctx, "mediaRepo.Delete error: %v", err)
		return err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", id)
	err = r.c.Delete(ctx, cacheKey)
	if err != nil {
		logger.Errorf(ctx, "mediaRepo.Delete cache error: %v", err)
	}

	return nil
}

// FindMedia finds a media by ID.
func (r *mediaRepository) FindMedia(ctx context.Context, params *structs.FindMedia) (*ent.Media, error) {
	// create builder
	builder := r.ecr.Media.Query()

	// if media ID provided
	if validator.IsNotEmpty(params.Media) {
		builder = builder.Where(mediaEnt.IDEQ(params.Media))
	}

	// if tenant provided
	if validator.IsNotEmpty(params.Tenant) {
		builder = builder.Where(mediaEnt.TenantIDEQ(params.Tenant))
	}

	// execute the builder
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}
