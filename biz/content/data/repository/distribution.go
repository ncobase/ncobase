package repository

import (
	"context"
	"fmt"
	"ncobase/biz/content/data"
	"ncobase/biz/content/data/ent"
	distributionEnt "ncobase/biz/content/data/ent/distribution"
	"ncobase/biz/content/structs"

	"github.com/ncobase/ncore/data/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/redis/go-redis/v9"

	"github.com/ncobase/ncore/data/search"
)

// DistributionRepositoryInterface represents the distribution repository interface.
type DistributionRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateDistributionBody) (*ent.Distribution, error)
	GetByID(ctx context.Context, id string) (*ent.Distribution, error)
	Update(ctx context.Context, id string, updates types.JSON) (*ent.Distribution, error)
	List(ctx context.Context, params *structs.ListDistributionParams) ([]*ent.Distribution, error)
	Count(ctx context.Context, params *structs.ListDistributionParams) (int, error)
	ListWithCount(ctx context.Context, params *structs.ListDistributionParams) ([]*ent.Distribution, int, error)
	Delete(ctx context.Context, id string) error
	FindDistribution(ctx context.Context, params *structs.FindDistribution) (*ent.Distribution, error)
	GetByTopicAndChannel(ctx context.Context, topicID string, channelID string) (*ent.Distribution, error)
	GetPendingDistributions(ctx context.Context, limit int) ([]*ent.Distribution, error)
	GetScheduledDistributions(ctx context.Context, before int64, limit int) ([]*ent.Distribution, error)
}

// distributionRepository implements the DistributionRepositoryInterface.
type distributionRepository struct {
	data *data.Data
	sc   *search.Client
	ec   *ent.Client
	ecr  *ent.Client
	rc   *redis.Client
	c    *cache.Cache[ent.Distribution]
}

// NewDistributionRepository creates a new distribution repository.
func NewDistributionRepository(d *data.Data) DistributionRepositoryInterface {
	ec := d.GetMasterEntClient()
	ecr := d.GetSlaveEntClient()
	rc := d.GetRedis().(*redis.Client)
	return &distributionRepository{
		data: d,
		ec:   ec,
		ecr:  ecr,
		rc:   rc,
		c:    cache.NewCache[ent.Distribution](rc, "ncse_distribution"),
	}
}

// Create creates a new distribution.
func (r *distributionRepository) Create(ctx context.Context, body *structs.CreateDistributionBody) (*ent.Distribution, error) {
	// create builder
	builder := r.ec.Distribution.Create()

	// set values
	builder.SetTopicID(body.TopicID)
	builder.SetChannelID(body.ChannelID)
	builder.SetStatus(body.Status)
	builder.SetNillableScheduledAt(body.ScheduledAt)
	builder.SetNillablePublishedAt(body.PublishedAt)

	if !validator.IsNil(body.MetaData) && !validator.IsEmpty(body.MetaData) {
		builder.SetExtras(*body.MetaData)
	}

	builder.SetNillableExternalID(&body.ExternalID)
	builder.SetNillableExternalURL(&body.ExternalURL)

	if !validator.IsNil(body.CustomData) && !validator.IsEmpty(body.CustomData) {
		builder.SetExtras(*body.CustomData)
	}

	builder.SetNillableErrorDetails(&body.ErrorDetails)
	builder.SetNillableSpaceID(&body.SpaceID)
	builder.SetNillableCreatedBy(body.CreatedBy)

	// execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "distributionRepo.Create error: %v", err)
		return nil, err
	}

	// Index in Meilisearch
	if r.sc != nil {
		if err = r.sc.Index(ctx, &search.IndexRequest{Index: "content_distributions", Document: row}); err != nil {
			logger.Errorf(ctx, "distributionRepo.Create error creating Meilisearch index: %v", err)
		}
	}

	return row, nil
}

// GetByID gets a distribution by ID.
func (r *distributionRepository) GetByID(ctx context.Context, id string) (*ent.Distribution, error) {
	// check cache
	cacheKey := fmt.Sprintf("%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindDistribution(ctx, &structs.FindDistribution{Distribution: id})
	if err != nil {
		logger.Errorf(ctx, "distributionRepo.GetByID error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "distributionRepo.GetByID cache error: %v", err)
	}

	return row, nil
}

// Update updates a distribution.
func (r *distributionRepository) Update(ctx context.Context, id string, updates types.JSON) (*ent.Distribution, error) {
	distribution, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// create builder
	builder := distribution.Update()

	// set values based on the updates map
	for field, value := range updates {
		switch field {
		case "topic_id":
			builder.SetNillableTopicID(convert.ToPointer(value.(string)))
		case "channel_id":
			builder.SetNillableChannelID(convert.ToPointer(value.(string)))
		case "status":
			builder.SetStatus(value.(int))
		case "scheduled_at":
			builder.SetNillableScheduledAt(convert.ToPointer(value.(int64)))
		case "published_at":
			builder.SetNillablePublishedAt(convert.ToPointer(value.(int64)))
		case "meta_data":
			builder.SetExtras(value.(types.JSON))
		case "external_id":
			builder.SetNillableExternalID(convert.ToPointer(value.(string)))
		case "external_url":
			builder.SetNillableExternalURL(convert.ToPointer(value.(string)))
		case "custom_data":
			builder.SetExtras(value.(types.JSON))
		case "error_details":
			builder.SetNillableErrorDetails(convert.ToPointer(value.(string)))
		case "space_id":
			builder.SetNillableSpaceID(convert.ToPointer(value.(string)))
		case "updated_by":
			builder.SetNillableUpdatedBy(convert.ToPointer(value.(string)))
		}
	}

	// execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "distributionRepo.Update error: %v", err)
		return nil, err
	}

	// Update Meilisearch index
	if r.sc != nil {
		if err = r.sc.Index(ctx, &search.IndexRequest{Index: "content_distributions", Document: row, DocumentID: row.ID}); err != nil {
			logger.Errorf(ctx, "distributionRepo.Update error updating Meilisearch index: %v", err)
		}
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", distribution.ID)
	err = r.c.Delete(ctx, cacheKey)
	if err != nil {
		logger.Errorf(ctx, "distributionRepo.Update cache error: %v", err)
	}

	return row, nil
}

// List gets a list of distributions.
func (r *distributionRepository) List(ctx context.Context, params *structs.ListDistributionParams) ([]*ent.Distribution, error) {
	// create builder
	builder := r.ecr.Distribution.Query()

	// apply filters
	if validator.IsNotEmpty(params.TopicID) {
		builder.Where(distributionEnt.TopicIDEQ(params.TopicID))
	}

	if validator.IsNotEmpty(params.ChannelID) {
		builder.Where(distributionEnt.ChannelIDEQ(params.ChannelID))
	}

	if params.Status > 0 {
		builder.Where(distributionEnt.StatusEQ(params.Status))
	}

	if validator.IsNotEmpty(params.SpaceID) {
		builder.Where(distributionEnt.SpaceIDEQ(params.SpaceID))
	}

	// load relations if requested
	if params.WithTopic {
		builder.WithTopic()
	}

	if params.WithChannel {
		builder.WithChannel()
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
				distributionEnt.Or(
					distributionEnt.CreatedAtGT(timestamp),
					distributionEnt.And(
						distributionEnt.CreatedAtEQ(timestamp),
						distributionEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				distributionEnt.Or(
					distributionEnt.CreatedAtLT(timestamp),
					distributionEnt.And(
						distributionEnt.CreatedAtEQ(timestamp),
						distributionEnt.IDLT(id),
					),
				),
			)
		}
	}

	// set ordering
	if params.Direction == "backward" {
		builder.Order(ent.Asc(distributionEnt.FieldCreatedAt), ent.Asc(distributionEnt.FieldID))
	} else {
		builder.Order(ent.Desc(distributionEnt.FieldCreatedAt), ent.Desc(distributionEnt.FieldID))
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
		logger.Errorf(ctx, "distributionRepo.List error: %v", err)
		return nil, err
	}

	return rows, nil
}

// Count gets a count of distributions.
func (r *distributionRepository) Count(ctx context.Context, params *structs.ListDistributionParams) (int, error) {
	// create builder
	builder := r.ecr.Distribution.Query()

	// apply filters
	if validator.IsNotEmpty(params.TopicID) {
		builder.Where(distributionEnt.TopicIDEQ(params.TopicID))
	}

	if validator.IsNotEmpty(params.ChannelID) {
		builder.Where(distributionEnt.ChannelIDEQ(params.ChannelID))
	}

	if params.Status > 0 {
		builder.Where(distributionEnt.StatusEQ(params.Status))
	}

	if validator.IsNotEmpty(params.SpaceID) {
		builder.Where(distributionEnt.SpaceIDEQ(params.SpaceID))
	}

	// execute count query
	return builder.Count(ctx)
}

// ListWithCount gets a list of distributions and their total count.
func (r *distributionRepository) ListWithCount(ctx context.Context, params *structs.ListDistributionParams) ([]*ent.Distribution, int, error) {
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

// Delete deletes a distribution.
func (r *distributionRepository) Delete(ctx context.Context, id string) error {
	// create builder
	builder := r.ec.Distribution.Delete()

	// execute the builder
	_, err := builder.Where(distributionEnt.IDEQ(id)).Exec(ctx)
	if err != nil {
		logger.Errorf(ctx, "distributionRepo.Delete error: %v", err)
		return err
	}

	// Delete from Meilisearch
	if r.sc != nil {
		if err = r.sc.Delete(ctx, "content_distributions", id); err != nil {
			logger.Errorf(ctx, "distributionRepo.Delete error deleting Meilisearch index: %v", err)
		}
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", id)
	err = r.c.Delete(ctx, cacheKey)
	if err != nil {
		logger.Errorf(ctx, "distributionRepo.Delete cache error: %v", err)
	}

	return nil
}

// FindDistribution finds a distribution by ID or other criteria.
func (r *distributionRepository) FindDistribution(ctx context.Context, params *structs.FindDistribution) (*ent.Distribution, error) {
	// create builder
	builder := r.ecr.Distribution.Query()

	// if distribution ID provided
	if validator.IsNotEmpty(params.Distribution) {
		builder = builder.Where(distributionEnt.IDEQ(params.Distribution))
	}

	// if topic ID provided
	if validator.IsNotEmpty(params.TopicID) {
		builder = builder.Where(distributionEnt.TopicIDEQ(params.TopicID))
	}

	// if channel ID provided
	if validator.IsNotEmpty(params.ChannelID) {
		builder = builder.Where(distributionEnt.ChannelIDEQ(params.ChannelID))
	}

	// if space / space provided
	if validator.IsNotEmpty(params.SpaceID) {
		builder = builder.Where(distributionEnt.SpaceIDEQ(params.SpaceID))
	}

	// execute the builder
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// GetByTopicAndChannel gets a distribution by topic ID and channel ID.
func (r *distributionRepository) GetByTopicAndChannel(ctx context.Context, topicID string, channelID string) (*ent.Distribution, error) {
	return r.FindDistribution(ctx, &structs.FindDistribution{
		TopicID:   topicID,
		ChannelID: channelID,
	})
}

// GetPendingDistributions gets a list of pending distributions.
func (r *distributionRepository) GetPendingDistributions(ctx context.Context, limit int) ([]*ent.Distribution, error) {
	// create builder
	builder := r.ecr.Distribution.Query().
		Where(distributionEnt.StatusEQ(structs.DistributionStatusDraft)).
		Order(ent.Asc(distributionEnt.FieldCreatedAt))

	if limit > 0 {
		builder.Limit(limit)
	}

	// execute query
	return builder.All(ctx)
}

// GetScheduledDistributions gets a list of scheduled distributions.
func (r *distributionRepository) GetScheduledDistributions(ctx context.Context, before int64, limit int) ([]*ent.Distribution, error) {
	// create builder
	builder := r.ecr.Distribution.Query().
		Where(
			distributionEnt.StatusEQ(structs.DistributionStatusScheduled),
			distributionEnt.ScheduledAtLTE(before),
		).
		Order(ent.Asc(distributionEnt.FieldScheduledAt))

	if limit > 0 {
		builder.Limit(limit)
	}

	// execute query
	return builder.All(ctx)
}
