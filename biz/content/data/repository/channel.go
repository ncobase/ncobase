package repository

import (
	"context"
	"fmt"
	"ncobase/content/data"
	"ncobase/content/data/ent"
	channelEnt "ncobase/content/data/ent/cmschannel"
	"ncobase/content/structs"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/utils/slug"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/redis/go-redis/v9"
)

// ChannelRepositoryInterface represents the channel repository interface.
type ChannelRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateChannelBody) (*ent.CMSChannel, error)
	GetByID(ctx context.Context, id string) (*ent.CMSChannel, error)
	GetBySlug(ctx context.Context, slug string) (*ent.CMSChannel, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*ent.CMSChannel, error)
	List(ctx context.Context, params *structs.ListChannelParams) ([]*ent.CMSChannel, error)
	Count(ctx context.Context, params *structs.ListChannelParams) (int, error)
	ListWithCount(ctx context.Context, params *structs.ListChannelParams) ([]*ent.CMSChannel, int, error)
	Delete(ctx context.Context, slug string) error
	FindChannel(ctx context.Context, params *structs.FindChannel) (*ent.CMSChannel, error)
}

// channelRepository implements the ChannelRepositoryInterface.
type channelRepository struct {
	ec  *ent.Client
	ecr *ent.Client
	rc  *redis.Client
	c   *cache.Cache[ent.CMSChannel]
}

// NewChannelRepository creates a new channel repository.
func NewChannelRepository(d *data.Data) ChannelRepositoryInterface {
	ec := d.GetMasterEntClient()
	ecr := d.GetSlaveEntClient()
	rc := d.GetRedis()
	return &channelRepository{ec, ecr, rc, cache.NewCache[ent.CMSChannel](rc, "ncse_channel")}
}

// Create creates a new channel.
func (r *channelRepository) Create(ctx context.Context, body *structs.CreateChannelBody) (*ent.CMSChannel, error) {
	// create builder
	builder := r.ec.CMSChannel.Create()

	// set values
	builder.SetNillableName(&body.Name)
	builder.SetNillableType(&body.Type)

	// set slug or generate from name if empty
	if validator.IsEmpty(body.Slug) {
		body.Slug = slug.Unicode(body.Name)
	}
	builder.SetNillableSlug(&body.Slug)

	builder.SetNillableIcon(&body.Icon)
	builder.SetStatus(body.Status)

	if !validator.IsNil(body.Config) && !validator.IsEmpty(body.Config) {
		builder.SetExtras(*body.Config)
	}

	builder.SetNillableDescription(&body.Description)
	builder.SetNillableLogo(&body.Logo)
	builder.SetNillableWebhookURL(&body.WebhookURL)
	builder.SetAutoPublish(body.AutoPublish)
	builder.SetRequireReview(body.RequireReview)

	if len(body.AllowedTypes) > 0 {
		builder.SetAllowedTypes(body.AllowedTypes)
	}

	builder.SetNillableSpaceID(&body.SpaceID)
	builder.SetNillableCreatedBy(body.CreatedBy)

	// execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "channelRepo.Create error: %v", err)
		return nil, err
	}

	return row, nil
}

// GetByID gets a channel by ID.
func (r *channelRepository) GetByID(ctx context.Context, id string) (*ent.CMSChannel, error) {
	// check cache
	cacheKey := fmt.Sprintf("%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindChannel(ctx, &structs.FindChannel{Channel: id})
	if err != nil {
		logger.Errorf(ctx, "channelRepo.GetByID error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "channelRepo.GetByID cache error: %v", err)
	}

	return row, nil
}

// GetBySlug gets a channel by slug.
func (r *channelRepository) GetBySlug(ctx context.Context, slug string) (*ent.CMSChannel, error) {
	// check cache
	cacheKey := fmt.Sprintf("slug:%s", slug)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindChannel(ctx, &structs.FindChannel{Channel: slug})
	if err != nil {
		logger.Errorf(ctx, "channelRepo.GetBySlug error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "channelRepo.GetBySlug cache error: %v", err)
	}

	return row, nil
}

// Update updates a channel.
func (r *channelRepository) Update(ctx context.Context, slug string, updates types.JSON) (*ent.CMSChannel, error) {
	channel, err := r.FindChannel(ctx, &structs.FindChannel{Channel: slug})
	if err != nil {
		return nil, err
	}

	// create builder
	builder := channel.Update()

	// set values based on the updates map
	for field, value := range updates {
		switch field {
		case "name":
			builder.SetNillableName(convert.ToPointer(value.(string)))
		case "type":
			builder.SetNillableType(convert.ToPointer(value.(string)))
		case "slug":
			builder.SetNillableSlug(convert.ToPointer(value.(string)))
		case "icon":
			builder.SetNillableIcon(convert.ToPointer(value.(string)))
		case "status":
			builder.SetStatus(value.(int))
		case "allowed_types":
			builder.SetAllowedTypes(convert.ToStringArray(value))
		case "config":
			builder.SetExtras(value.(types.JSON))
		case "description":
			builder.SetNillableDescription(convert.ToPointer(value.(string)))
		case "logo":
			builder.SetNillableLogo(convert.ToPointer(value.(string)))
		case "webhook_url":
			builder.SetNillableWebhookURL(convert.ToPointer(value.(string)))
		case "auto_publish":
			builder.SetAutoPublish(value.(bool))
		case "require_review":
			builder.SetRequireReview(value.(bool))
		case "tenant_id":
			builder.SetNillableSpaceID(convert.ToPointer(value.(string)))
		case "updated_by":
			builder.SetNillableUpdatedBy(convert.ToPointer(value.(string)))
		}
	}

	// execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "channelRepo.Update error: %v", err)
		return nil, err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", channel.ID)
	err = r.c.Delete(ctx, cacheKey)
	err = r.c.Delete(ctx, fmt.Sprintf("slug:%s", channel.Slug))
	if err != nil {
		logger.Errorf(ctx, "channelRepo.Update cache error: %v", err)
	}

	return row, nil
}

// List gets a list of channels.
func (r *channelRepository) List(ctx context.Context, params *structs.ListChannelParams) ([]*ent.CMSChannel, error) {
	// create builder
	builder := r.ecr.CMSChannel.Query()

	// apply filters
	if validator.IsNotEmpty(params.Type) {
		builder.Where(channelEnt.TypeEQ(params.Type))
	}

	if params.Status > 0 {
		builder.Where(channelEnt.StatusEQ(params.Status))
	}

	if validator.IsNotEmpty(params.SpaceID) {
		builder.Where(channelEnt.SpaceIDEQ(params.SpaceID))
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
				channelEnt.Or(
					channelEnt.CreatedAtGT(timestamp),
					channelEnt.And(
						channelEnt.CreatedAtEQ(timestamp),
						channelEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				channelEnt.Or(
					channelEnt.CreatedAtLT(timestamp),
					channelEnt.And(
						channelEnt.CreatedAtEQ(timestamp),
						channelEnt.IDLT(id),
					),
				),
			)
		}
	}

	// set ordering
	if params.Direction == "backward" {
		builder.Order(ent.Asc(channelEnt.FieldCreatedAt), ent.Asc(channelEnt.FieldID))
	} else {
		builder.Order(ent.Desc(channelEnt.FieldCreatedAt), ent.Desc(channelEnt.FieldID))
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
		logger.Errorf(ctx, "channelRepo.List error: %v", err)
		return nil, err
	}

	return rows, nil
}

// Count gets a count of channels.
func (r *channelRepository) Count(ctx context.Context, params *structs.ListChannelParams) (int, error) {
	// create builder
	builder := r.ecr.CMSChannel.Query()

	// apply filters
	if validator.IsNotEmpty(params.Type) {
		builder.Where(channelEnt.TypeEQ(params.Type))
	}

	if params.Status > 0 {
		builder.Where(channelEnt.StatusEQ(params.Status))
	}

	if validator.IsNotEmpty(params.SpaceID) {
		builder.Where(channelEnt.SpaceIDEQ(params.SpaceID))
	}

	// execute count query
	return builder.Count(ctx)
}

// ListWithCount gets a list of channels and their total count.
func (r *channelRepository) ListWithCount(ctx context.Context, params *structs.ListChannelParams) ([]*ent.CMSChannel, int, error) {
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

// Delete deletes a channel.
func (r *channelRepository) Delete(ctx context.Context, slug string) error {
	channel, err := r.FindChannel(ctx, &structs.FindChannel{Channel: slug})
	if err != nil {
		return err
	}

	// create builder
	builder := r.ec.CMSChannel.Delete()

	// execute the builder
	_, err = builder.Where(channelEnt.IDEQ(channel.ID)).Exec(ctx)
	if err != nil {
		logger.Errorf(ctx, "channelRepo.Delete error: %v", err)
		return err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", channel.ID)
	err = r.c.Delete(ctx, cacheKey)
	err = r.c.Delete(ctx, fmt.Sprintf("slug:%s", channel.Slug))
	if err != nil {
		logger.Errorf(ctx, "channelRepo.Delete cache error: %v", err)
	}

	return nil
}

// FindChannel finds a channel by ID or slug.
func (r *channelRepository) FindChannel(ctx context.Context, params *structs.FindChannel) (*ent.CMSChannel, error) {
	// create builder
	builder := r.ecr.CMSChannel.Query()

	// if channel ID or slug provided
	if validator.IsNotEmpty(params.Channel) {
		builder = builder.Where(channelEnt.Or(
			channelEnt.ID(params.Channel),
			channelEnt.SlugEQ(params.Channel),
		))
	}

	// if type provided
	if validator.IsNotEmpty(params.Type) {
		builder = builder.Where(channelEnt.TypeEQ(params.Type))
	}

	// if space / tenant provided
	if validator.IsNotEmpty(params.SpaceID) {
		builder = builder.Where(channelEnt.SpaceIDEQ(params.SpaceID))
	}

	// execute the builder
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}
