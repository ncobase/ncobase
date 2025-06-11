package repository

import (
	"context"
	"fmt"
	"ncobase/content/data"
	"ncobase/content/data/ent"
	topicEnt "ncobase/content/data/ent/topic"
	"ncobase/content/structs"

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

// TopicRepositoryInterface represents the topic repository interface.
type TopicRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateTopicBody) (*ent.Topic, error)
	GetByID(ctx context.Context, id string) (*ent.Topic, error)
	GetBySlug(ctx context.Context, slug string) (*ent.Topic, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*ent.Topic, error)
	List(ctx context.Context, params *structs.ListTopicParams) ([]*ent.Topic, error)
	Delete(ctx context.Context, slug string) error
	FindTopic(ctx context.Context, params *structs.FindTopic) (*ent.Topic, error)
	ListBuilder(ctx context.Context, params *structs.ListTopicParams) (*ent.TopicQuery, error)
	CountX(ctx context.Context, params *structs.ListTopicParams) int
}

// topicRepository implements the TopicRepositoryInterface.
type topicRepository struct {
	ec  *ent.Client
	ecr *ent.Client
	rc  *redis.Client
	ms  *meili.Client
	c   *cache.Cache[ent.Topic]
}

// NewTopicRepository creates a new topic repository.
func NewTopicRepository(d *data.Data) TopicRepositoryInterface {
	ec := d.GetMasterEntClient()
	ecr := d.GetSlaveEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &topicRepository{ec, ecr, rc, ms, cache.NewCache[ent.Topic](rc, "ncse_topic")}
}

// Create creates a new topic.
func (r *topicRepository) Create(ctx context.Context, body *structs.CreateTopicBody) (*ent.Topic, error) {

	// create builder.
	builder := r.ec.Topic.Create()
	// set values.
	builder.SetNillableName(&body.Name)
	builder.SetNillableTitle(&body.Title)
	builder.SetNillableSlug(&body.Slug)
	builder.SetNillableContent(&body.Content)
	builder.SetNillableThumbnail(&body.Thumbnail)
	builder.SetTemp(body.Temp)
	builder.SetMarkdown(body.Markdown)
	builder.SetPrivate(body.Private)
	builder.SetStatus(body.Status)
	builder.SetNillableReleased(&body.Released)
	builder.SetNillableTaxonomyID(&body.TaxonomyID)
	builder.SetNillableSpaceID(&body.SpaceID)
	builder.SetNillableCreatedBy(body.CreatedBy)

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "topicRepo.Create error: %v", err)
		return nil, err
	}

	// Create the topic in Meilisearch index
	if err = r.ms.IndexDocuments("topics", row); err != nil {
		logger.Errorf(ctx, "topicRepo.Create error creating Meilisearch index: %v", err)
		// return nil, err
	}

	return row, nil
}

// GetByID gets a topic by ID.
func (r *topicRepository) GetByID(ctx context.Context, id string) (*ent.Topic, error) {
	// // Search in Meilisearch index
	// if res, _ := r.ms.Search(ctx, "topics", id, &meilisearch.SearchRequest{Limit: 1}); res != nil && res.Hits != nil && len(res.Hits) > 0 {
	// 	if hit, ok := res.Hits[0].(*ent.Topic); ok {
	// 		return hit, nil
	// 	}
	// }
	// check cache
	cacheKey := fmt.Sprintf("%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindTopic(ctx, &structs.FindTopic{Topic: id})
	if err != nil {
		logger.Errorf(ctx, "topicRepo.GetByID error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "topicRepo.GetByID cache error: %v", err)
	}

	return row, nil
}

// GetBySlug gets a topic by slug.
func (r *topicRepository) GetBySlug(ctx context.Context, slug string) (*ent.Topic, error) {
	// // Search in Meilisearch index
	// if res, _ := r.ms.Search(ctx, "topics", slug, &meilisearch.SearchRequest{Limit: 1}); res != nil && res.Hits != nil && len(res.Hits) > 0 {
	// 	if hit, ok := res.Hits[0].(*ent.Topic); ok {
	// 		return hit, nil
	// 	}
	// }
	// check cache
	cacheKey := fmt.Sprintf("slug:%s", slug)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindTopic(ctx, &structs.FindTopic{Topic: slug})
	if err != nil {
		logger.Errorf(ctx, "topicRepo.GetBySlug error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "topicRepo.GetBySlug cache error: %v", err)
	}

	return row, nil
}

// Update updates a topic (full or partial).
func (r *topicRepository) Update(ctx context.Context, slug string, updates types.JSON) (*ent.Topic, error) {
	topic, err := r.FindTopic(ctx, &structs.FindTopic{Topic: slug})
	if err != nil {
		return nil, err
	}

	builder := topic.Update()

	for field, value := range updates {
		switch field {
		case "name":
			builder.SetNillableName(convert.ToPointer(value.(string)))
		case "title":
			builder.SetNillableTitle(convert.ToPointer(value.(string)))
		case "slug":
			builder.SetNillableSlug(convert.ToPointer(value.(string)))
		case "content":
			builder.SetNillableContent(convert.ToPointer(value.(string)))
		case "thumbnail":
			builder.SetNillableThumbnail(convert.ToPointer(value.(string)))
		case "temp":
			builder.SetTemp(value.(bool))
		case "markdown":
			builder.SetMarkdown(value.(bool))
		case "private":
			builder.SetPrivate(value.(bool))
		case "status":
			builder.SetStatus(value.(int))
		case "released":
			builder.SetNillableReleased(convert.ToPointer(value.(int64)))
		case "taxonomy_id":
			builder.SetNillableTaxonomyID(convert.ToPointer(value.(string)))
		case "tenant_id":
			builder.SetNillableSpaceID(convert.ToPointer(value.(string)))
		case "updated_by":
			builder.SetNillableUpdatedBy(convert.ToPointer(value.(string)))
		}
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "topicRepo.Update error: %v", err)
		return nil, err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", topic.ID)
	err = r.c.Delete(ctx, cacheKey)
	err = r.c.Delete(ctx, topic.Slug)
	if err != nil {
		logger.Errorf(ctx, "topicRepo.Update cache error: %v", err)
	}

	// Update the topic in Meilisearch index
	if err = r.ms.DeleteDocuments("topics", slug); err != nil {
		logger.Errorf(ctx, "topicRepo.Update error deleting Meilisearch index: %v", err)
		// return nil, err
	}
	if err = r.ms.IndexDocuments("topics", row); err != nil {
		logger.Errorf(ctx, "topicRepo.Update error updating Meilisearch index: %v", err)
		// return nil, err
	}

	return row, nil
}

// List gets a list of topics.
func (r *topicRepository) List(ctx context.Context, params *structs.ListTopicParams) ([]*ent.Topic, error) {
	// create list builder
	builder, err := r.ListBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return nil, err
	}

	// belong space / tenant
	if params.SpaceID != "" {
		builder.Where(topicEnt.SpaceIDEQ(params.SpaceID))
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
				topicEnt.Or(
					topicEnt.CreatedAtGT(timestamp),
					topicEnt.And(
						topicEnt.CreatedAtEQ(timestamp),
						topicEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				topicEnt.Or(
					topicEnt.CreatedAtLT(timestamp),
					topicEnt.And(
						topicEnt.CreatedAtEQ(timestamp),
						topicEnt.IDLT(id),
					),
				),
			)
		}
	}

	if params.Direction == "backward" {
		builder.Order(ent.Asc(topicEnt.FieldCreatedAt), ent.Asc(topicEnt.FieldID))
	} else {
		builder.Order(ent.Desc(topicEnt.FieldCreatedAt), ent.Desc(topicEnt.FieldID))
	}

	builder.Limit(params.Limit)

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "topicRepo.List error: %v", err)
		return nil, err
	}

	return rows, nil
}

// Delete deletes a topic.
func (r *topicRepository) Delete(ctx context.Context, slug string) error {
	topic, err := r.FindTopic(ctx, &structs.FindTopic{Topic: slug})
	if err != nil {
		return err
	}

	// create builder.
	builder := r.ec.Topic.Delete()

	// execute the builder and verify the result.
	if _, err = builder.Where(topicEnt.IDEQ(topic.ID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "topicRepo.Delete error: %v", err)
		return err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", topic.ID)
	err = r.c.Delete(ctx, cacheKey)
	err = r.c.Delete(ctx, fmt.Sprintf("topic:slug:%s", topic.Slug))
	if err != nil {
		logger.Errorf(ctx, "topicRepo.Delete cache error: %v", err)
	}

	// delete from Meilisearch index
	if err = r.ms.DeleteDocuments("topics", topic.ID); err != nil {
		logger.Errorf(ctx, "topicRepo.Delete index error: %v", err)
		// return nil, err
	}

	return nil
}

// FindTopic finds a topic.
func (r *topicRepository) FindTopic(ctx context.Context, params *structs.FindTopic) (*ent.Topic, error) {

	// create builder.
	builder := r.ecr.Topic.Query()

	if validator.IsNotEmpty(params.Topic) {
		builder = builder.Where(topicEnt.Or(
			topicEnt.ID(params.Topic),
			topicEnt.SlugEQ(params.Topic),
		))
	}
	if validator.IsNotEmpty(params.SpaceID) {
		builder = builder.Where(topicEnt.SpaceIDEQ(params.SpaceID))
	}

	// execute the builder.
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// ListBuilder creates list builder.
func (r *topicRepository) ListBuilder(_ context.Context, _ *structs.ListTopicParams) (*ent.TopicQuery, error) {
	// create builder.
	builder := r.ecr.Topic.Query()

	return builder, nil
}

// CountX gets a count of topics.
func (r *topicRepository) CountX(ctx context.Context, params *structs.ListTopicParams) int {
	// create list builder
	builder, err := r.ListBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// Count gets a count of topics.
func (r *topicRepository) Count(ctx context.Context, params *structs.ListTopicParams) (int, error) {
	// create list builder
	builder, err := r.ListBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0, err
	}
	return builder.Count(ctx)
}
