package repository

import (
	"context"
	"fmt"
	"ncobase/content/data"
	"ncobase/content/data/ent"
	mediaEnt "ncobase/content/data/ent/media"
	topicMediaEnt "ncobase/content/data/ent/topicmedia"
	"ncobase/content/structs"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/redis/go-redis/v9"
)

// TopicMediaRepositoryInterface represents the topic media repository interface.
type TopicMediaRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateTopicMediaBody) (*ent.TopicMedia, error)
	GetByID(ctx context.Context, id string) (*ent.TopicMedia, error)
	Update(ctx context.Context, id string, topicID string, mediaID string, mediaType string, order int) (*ent.TopicMedia, error)
	List(ctx context.Context, params *structs.ListTopicMediaParams) ([]*ent.TopicMedia, error)
	Count(ctx context.Context, params *structs.ListTopicMediaParams) (int, error)
	Delete(ctx context.Context, id string) error
	FindTopicMedia(ctx context.Context, params *structs.FindTopicMedia) (*ent.TopicMedia, error)
	GetByTopicAndMedia(ctx context.Context, topicID string, mediaID string) (*ent.TopicMedia, error)
	ListWithMedia(ctx context.Context, params *structs.ListTopicMediaParams) ([]*ent.TopicMedia, error)
}

// topicMediaRepository implements the TopicMediaRepositoryInterface.
type topicMediaRepository struct {
	ec  *ent.Client
	ecr *ent.Client
	rc  *redis.Client
	c   *cache.Cache[ent.TopicMedia]
}

// NewTopicMediaRepository creates a new topic media repository.
func NewTopicMediaRepository(d *data.Data) TopicMediaRepositoryInterface {
	ec := d.GetEntClient()
	ecr := d.GetEntClientRead()
	rc := d.GetRedis()
	return &topicMediaRepository{ec, ecr, rc, cache.NewCache[ent.TopicMedia](rc, "ncse_topic_media")}
}

// Create creates a new topic media relation.
func (r *topicMediaRepository) Create(ctx context.Context, body *structs.CreateTopicMediaBody) (*ent.TopicMedia, error) {
	// create builder
	builder := r.ec.TopicMedia.Create()

	// set values
	builder.SetTopicID(body.TopicID)
	builder.SetMediaID(body.MediaID)
	builder.SetType(body.Type)
	builder.SetOrder(body.Order)
	builder.SetNillableCreatedBy(body.CreatedBy)

	// execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "topicMediaRepo.Create error: %v", err)
		return nil, err
	}

	return row, nil
}

// GetByID gets a topic media relation by ID.
func (r *topicMediaRepository) GetByID(ctx context.Context, id string) (*ent.TopicMedia, error) {
	// check cache
	cacheKey := fmt.Sprintf("%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindTopicMedia(ctx, &structs.FindTopicMedia{TopicMedia: id})
	if err != nil {
		logger.Errorf(ctx, "topicMediaRepo.GetByID error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "topicMediaRepo.GetByID cache error: %v", err)
	}

	return row, nil
}

// Update updates a topic media relation.
func (r *topicMediaRepository) Update(ctx context.Context, id string, topicID string, mediaID string, mediaType string, order int) (*ent.TopicMedia, error) {
	topicMedia, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// create builder
	builder := topicMedia.Update()

	// set values
	if topicID != "" {
		builder.SetTopicID(topicID)
	}

	if mediaID != "" {
		builder.SetMediaID(mediaID)
	}

	if mediaType != "" {
		builder.SetType(mediaType)
	}

	if order >= 0 {
		builder.SetOrder(order)
	}

	// execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "topicMediaRepo.Update error: %v", err)
		return nil, err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", topicMedia.ID)
	err = r.c.Delete(ctx, cacheKey)
	if err != nil {
		logger.Errorf(ctx, "topicMediaRepo.Update cache error: %v", err)
	}

	return row, nil
}

// List gets a list of topic media relations.
func (r *topicMediaRepository) List(ctx context.Context, params *structs.ListTopicMediaParams) ([]*ent.TopicMedia, error) {
	// create builder
	builder := r.ecr.TopicMedia.Query()

	// apply filters
	if validator.IsNotEmpty(params.TopicID) {
		builder.Where(topicMediaEnt.TopicIDEQ(params.TopicID))
	}

	if validator.IsNotEmpty(params.MediaID) {
		builder.Where(topicMediaEnt.MediaIDEQ(params.MediaID))
	}

	if validator.IsNotEmpty(params.Type) {
		builder.Where(topicMediaEnt.TypeEQ(params.Type))
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
				topicMediaEnt.Or(
					topicMediaEnt.CreatedAtGT(timestamp),
					topicMediaEnt.And(
						topicMediaEnt.CreatedAtEQ(timestamp),
						topicMediaEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				topicMediaEnt.Or(
					topicMediaEnt.CreatedAtLT(timestamp),
					topicMediaEnt.And(
						topicMediaEnt.CreatedAtEQ(timestamp),
						topicMediaEnt.IDLT(id),
					),
				),
			)
		}
	}

	// set ordering first by order field, then by creation time
	builder.Order(ent.Asc(topicMediaEnt.FieldOrder))

	if params.Direction == "backward" {
		builder.Order(ent.Asc(topicMediaEnt.FieldCreatedAt), ent.Asc(topicMediaEnt.FieldID))
	} else {
		builder.Order(ent.Desc(topicMediaEnt.FieldCreatedAt), ent.Desc(topicMediaEnt.FieldID))
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
		logger.Errorf(ctx, "topicMediaRepo.List error: %v", err)
		return nil, err
	}

	return rows, nil
}

// ListWithMedia gets a list of topic media relations with their associated media.
func (r *topicMediaRepository) ListWithMedia(ctx context.Context, params *structs.ListTopicMediaParams) ([]*ent.TopicMedia, error) {
	// create builder
	builder := r.ecr.TopicMedia.Query()

	// apply filters
	if validator.IsNotEmpty(params.TopicID) {
		builder.Where(topicMediaEnt.TopicIDEQ(params.TopicID))
	}

	if validator.IsNotEmpty(params.MediaID) {
		builder.Where(topicMediaEnt.MediaIDEQ(params.MediaID))
	}

	if validator.IsNotEmpty(params.Type) {
		builder.Where(topicMediaEnt.TypeEQ(params.Type))
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
				topicMediaEnt.Or(
					topicMediaEnt.CreatedAtGT(timestamp),
					topicMediaEnt.And(
						topicMediaEnt.CreatedAtEQ(timestamp),
						topicMediaEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				topicMediaEnt.Or(
					topicMediaEnt.CreatedAtLT(timestamp),
					topicMediaEnt.And(
						topicMediaEnt.CreatedAtEQ(timestamp),
						topicMediaEnt.IDLT(id),
					),
				),
			)
		}
	}

	// set ordering first by order field, then by creation time
	builder.Order(ent.Asc(topicMediaEnt.FieldOrder))

	if params.Direction == "backward" {
		builder.Order(ent.Asc(topicMediaEnt.FieldCreatedAt), ent.Asc(topicMediaEnt.FieldID))
	} else {
		builder.Order(ent.Desc(topicMediaEnt.FieldCreatedAt), ent.Desc(topicMediaEnt.FieldID))
	}

	// set limit
	if params.Limit > 0 {
		builder.Limit(params.Limit)
	} else {
		builder.Limit(10) // default limit
	}

	// Manually resolve the media relationships
	topicMediaList, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "topicMediaRepo.ListWithMedia error: %v", err)
		return nil, err
	}

	// Now load the media for each topic-media relation
	mediaIDs := make([]string, 0, len(topicMediaList))
	for _, tm := range topicMediaList {
		mediaIDs = append(mediaIDs, tm.MediaID)
	}

	// If there are no media IDs, return the original list
	if len(mediaIDs) == 0 {
		return topicMediaList, nil
	}

	// Get all media in one query
	mediaMap := make(map[string]*ent.Media)
	mediaList, err := r.ecr.Media.Query().
		Where(mediaEnt.IDIn(mediaIDs...)).
		All(ctx)
	if err != nil {
		logger.Errorf(ctx, "topicMediaRepo.ListWithMedia error getting media: %v", err)
		// Return the list without media
		return topicMediaList, nil
	}

	// Create a map of media by ID for easy lookup
	for _, media := range mediaList {
		mediaMap[media.ID] = media
	}

	// Attach media to topic-media relations
	for _, tm := range topicMediaList {
		if media, ok := mediaMap[tm.MediaID]; ok {
			if tm.Edges.Media == nil {
				tm.Edges.Media = media
			}
		}
	}

	return topicMediaList, nil
}

// Count gets a count of topic media relations.
func (r *topicMediaRepository) Count(ctx context.Context, params *structs.ListTopicMediaParams) (int, error) {
	// create builder
	builder := r.ecr.TopicMedia.Query()

	// apply filters
	if validator.IsNotEmpty(params.TopicID) {
		builder.Where(topicMediaEnt.TopicIDEQ(params.TopicID))
	}

	if validator.IsNotEmpty(params.MediaID) {
		builder.Where(topicMediaEnt.MediaIDEQ(params.MediaID))
	}

	if validator.IsNotEmpty(params.Type) {
		builder.Where(topicMediaEnt.TypeEQ(params.Type))
	}

	// execute count query
	return builder.Count(ctx)
}

// Delete deletes a topic media relation.
func (r *topicMediaRepository) Delete(ctx context.Context, id string) error {
	// create builder
	builder := r.ec.TopicMedia.Delete()

	// execute the builder
	_, err := builder.Where(topicMediaEnt.IDEQ(id)).Exec(ctx)
	if err != nil {
		logger.Errorf(ctx, "topicMediaRepo.Delete error: %v", err)
		return err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", id)
	err = r.c.Delete(ctx, cacheKey)
	if err != nil {
		logger.Errorf(ctx, "topicMediaRepo.Delete cache error: %v", err)
	}

	return nil
}

// FindTopicMedia finds a topic media relation by ID.
func (r *topicMediaRepository) FindTopicMedia(ctx context.Context, params *structs.FindTopicMedia) (*ent.TopicMedia, error) {
	// create builder
	builder := r.ecr.TopicMedia.Query()

	// if topic media ID provided
	if validator.IsNotEmpty(params.TopicMedia) {
		builder = builder.Where(topicMediaEnt.IDEQ(params.TopicMedia))
	}

	// if topic ID provided
	if validator.IsNotEmpty(params.TopicID) {
		builder = builder.Where(topicMediaEnt.TopicIDEQ(params.TopicID))
	}

	// if media ID provided
	if validator.IsNotEmpty(params.MediaID) {
		builder = builder.Where(topicMediaEnt.MediaIDEQ(params.MediaID))
	}

	// execute the builder
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// GetByTopicAndMedia gets a topic media relation by topic ID and media ID.
func (r *topicMediaRepository) GetByTopicAndMedia(ctx context.Context, topicID string, mediaID string) (*ent.TopicMedia, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("topic:%s:media:%s", topicID, mediaID)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	row, err := r.FindTopicMedia(ctx, &structs.FindTopicMedia{
		TopicID: topicID,
		MediaID: mediaID,
	})
	if err != nil {
		return nil, err
	}

	// Cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "topicMediaRepo.GetByTopicAndMedia cache error: %v", err)
	}

	return row, nil
}
