package repository

import (
	"context"
	"fmt"
	"ncobase/content/data"
	"ncobase/content/data/ent"
	mediaEnt "ncobase/content/data/ent/media"
	"ncobase/content/structs"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/redis/go-redis/v9"
)

// MediaRepositoryInterface for media repository operations
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

type mediaRepository struct {
	ec  *ent.Client
	ecr *ent.Client
	rc  *redis.Client
	c   *cache.Cache[ent.Media]
}

// NewMediaRepository creates new media repository
func NewMediaRepository(d *data.Data) MediaRepositoryInterface {
	ec := d.GetMasterEntClient()
	ecr := d.GetSlaveEntClient()
	rc := d.GetRedis()
	return &mediaRepository{ec, ecr, rc, cache.NewCache[ent.Media](rc, "ncse_media")}
}

// Create creates new media
func (r *mediaRepository) Create(ctx context.Context, body *structs.CreateMediaBody) (*ent.Media, error) {
	builder := r.ec.Media.Create()

	builder.SetNillableTitle(&body.Title)
	builder.SetNillableType(&body.Type)
	builder.SetNillableURL(&body.URL)
	builder.SetNillableSpaceID(&body.SpaceID)
	builder.SetNillableCreatedBy(body.CreatedBy)

	// Set extras with additional fields
	extras := make(types.JSON)
	if body.ResourceID != "" {
		extras["resource_id"] = body.ResourceID
	}
	if body.Description != "" {
		extras["description"] = body.Description
	}
	if body.Alt != "" {
		extras["alt"] = body.Alt
	}
	if body.OwnerID != "" {
		extras["owner_id"] = body.OwnerID
	}
	if body.Metadata != nil {
		for k, v := range *body.Metadata {
			extras[k] = v
		}
	}

	if len(extras) > 0 {
		builder.SetExtras(extras)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "mediaRepo.Create error: %v", err)
		return nil, err
	}

	return row, nil
}

// GetByID gets media by ID
func (r *mediaRepository) GetByID(ctx context.Context, id string) (*ent.Media, error) {
	cacheKey := fmt.Sprintf("%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	row, err := r.FindMedia(ctx, &structs.FindMedia{Media: id})
	if err != nil {
		logger.Errorf(ctx, "mediaRepo.GetByID error: %v", err)
		return nil, err
	}

	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "mediaRepo.GetByID cache error: %v", err)
	}

	return row, nil
}

// Update updates media
func (r *mediaRepository) Update(ctx context.Context, id string, updates types.JSON) (*ent.Media, error) {
	media, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	builder := media.Update()

	// Handle basic fields
	for field, value := range updates {
		switch field {
		case "title":
			builder.SetNillableTitle(convert.ToPointer(value.(string)))
		case "type":
			builder.SetNillableType(convert.ToPointer(value.(string)))
		case "url":
			builder.SetNillableURL(convert.ToPointer(value.(string)))
		case "space_id":
			builder.SetNillableSpaceID(convert.ToPointer(value.(string)))
		case "updated_by":
			builder.SetNillableUpdatedBy(convert.ToPointer(value.(string)))
		}
	}

	// Handle extras fields
	extras := media.Extras
	if extras == nil {
		extras = make(types.JSON)
	}

	extrasFields := []string{"resource_id", "description", "alt", "owner_id", "metadata"}
	for _, field := range extrasFields {
		if value, ok := updates[field]; ok {
			extras[field] = value
		}
	}

	builder.SetExtras(extras)

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "mediaRepo.Update error: %v", err)
		return nil, err
	}

	// Remove from cache
	cacheKey := fmt.Sprintf("%s", media.ID)
	err = r.c.Delete(ctx, cacheKey)
	if err != nil {
		logger.Errorf(ctx, "mediaRepo.Update cache error: %v", err)
	}

	return row, nil
}

// List gets list of media
func (r *mediaRepository) List(ctx context.Context, params *structs.ListMediaParams) ([]*ent.Media, error) {
	builder := r.ecr.Media.Query()

	// Apply filters
	if validator.IsNotEmpty(params.Type) {
		builder.Where(mediaEnt.TypeEQ(params.Type))
	}

	if validator.IsNotEmpty(params.Search) {
		builder.Where(mediaEnt.Or(
			mediaEnt.TitleContains(params.Search),
		))
	}

	if validator.IsNotEmpty(params.SpaceID) {
		builder.Where(mediaEnt.SpaceIDEQ(params.SpaceID))
	}

	// Apply cursor-based pagination
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

	// Set ordering
	if params.Direction == "backward" {
		builder.Order(ent.Asc(mediaEnt.FieldCreatedAt), ent.Asc(mediaEnt.FieldID))
	} else {
		builder.Order(ent.Desc(mediaEnt.FieldCreatedAt), ent.Desc(mediaEnt.FieldID))
	}

	// Set limit
	if params.Limit > 0 {
		builder.Limit(params.Limit)
	} else {
		builder.Limit(10)
	}

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "mediaRepo.List error: %v", err)
		return nil, err
	}

	return rows, nil
}

// Count gets count of media
func (r *mediaRepository) Count(ctx context.Context, params *structs.ListMediaParams) (int, error) {
	builder := r.ecr.Media.Query()

	if validator.IsNotEmpty(params.Type) {
		builder.Where(mediaEnt.TypeEQ(params.Type))
	}

	if validator.IsNotEmpty(params.Search) {
		builder.Where(mediaEnt.Or(
			mediaEnt.TitleContains(params.Search),
		))
	}

	if validator.IsNotEmpty(params.SpaceID) {
		builder.Where(mediaEnt.SpaceIDEQ(params.SpaceID))
	}

	return builder.Count(ctx)
}

// ListWithCount gets list of media and their total count
func (r *mediaRepository) ListWithCount(ctx context.Context, params *structs.ListMediaParams) ([]*ent.Media, int, error) {
	count, err := r.Count(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	return rows, count, nil
}

// Delete deletes media
func (r *mediaRepository) Delete(ctx context.Context, id string) error {
	builder := r.ec.Media.Delete()

	_, err := builder.Where(mediaEnt.IDEQ(id)).Exec(ctx)
	if err != nil {
		logger.Errorf(ctx, "mediaRepo.Delete error: %v", err)
		return err
	}

	cacheKey := fmt.Sprintf("%s", id)
	err = r.c.Delete(ctx, cacheKey)
	if err != nil {
		logger.Errorf(ctx, "mediaRepo.Delete cache error: %v", err)
	}

	return nil
}

// FindMedia finds media by ID
func (r *mediaRepository) FindMedia(ctx context.Context, params *structs.FindMedia) (*ent.Media, error) {
	builder := r.ecr.Media.Query()

	if validator.IsNotEmpty(params.Media) {
		builder = builder.Where(mediaEnt.IDEQ(params.Media))
	}

	if validator.IsNotEmpty(params.SpaceID) {
		builder = builder.Where(mediaEnt.SpaceIDEQ(params.SpaceID))
	}

	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}
