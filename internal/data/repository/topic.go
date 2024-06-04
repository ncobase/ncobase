package repo

import (
	"context"
	"fmt"
	"stocms/internal/data"
	"stocms/internal/data/ent"
	topicEnt "stocms/internal/data/ent/topic"
	"stocms/internal/data/structs"
	"stocms/pkg/cache"
	"stocms/pkg/log"
	"stocms/pkg/types"
	"stocms/pkg/validator"
	"time"

	"github.com/redis/go-redis/v9"
)

// Topic represents the topic repository interface.
type Topic interface {
	Create(ctx context.Context, body *structs.CreateTopicBody) (*ent.Topic, error)
	GetByID(ctx context.Context, id string) (*ent.Topic, error)
	GetBySlug(ctx context.Context, slug string) (*ent.Topic, error)
	Update(ctx context.Context, slug string, body types.JSON) (*ent.Topic, error)
	List(ctx context.Context, params *structs.ListTopicParams) ([]*ent.Topic, error)
	Delete(ctx context.Context, slug string) error
	FindTopic(ctx context.Context, p *structs.FindTopic) (*ent.Topic, error)
	ListBuilder(ctx context.Context, p *structs.ListTopicParams) (*ent.TopicQuery, error)
	CountX(ctx context.Context, p *structs.ListTopicParams) int
}

// topicRepo implements the Topic interface.
type topicRepo struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.Topic]
}

// NewTopic creates a new topic repository.
func NewTopic(d *data.Data) Topic {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &topicRepo{ec, rc, cache.NewCache[ent.Topic](rc, cache.Key("sc_topic"), true)}
}

// Create creates a new topic.
func (r *topicRepo) Create(ctx context.Context, body *structs.CreateTopicBody) (*ent.Topic, error) {

	// create builder.
	builder := r.ec.Topic.Create()

	// Set values
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
	builder.SetTaxonomyID(body.TaxonomyID)
	builder.SetDomainID(body.DomainID)
	builder.SetCreatedBy(body.CreatedBy)

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(nil, "topicRepo.Create error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByID gets a topic by ID.
func (r *topicRepo) GetByID(ctx context.Context, id string) (*ent.Topic, error) {
	cacheKey := fmt.Sprintf("%s", id)

	// Check cache first
	if cachedTopic, err := r.c.Get(ctx, cacheKey); err == nil {
		return cachedTopic, nil
	}

	// If not found in cache, query the database
	row, err := r.FindTopic(ctx, &structs.FindTopic{ID: id})

	if err != nil {
		log.Errorf(nil, "topicRepo.GetByID error: %v\n", err)
		return nil, err
	}

	// Cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(nil, "topicRepo.GetByID cache error: %v\n", err)
	}

	return row, nil
}

// GetBySlug gets a topic by slug.
func (r *topicRepo) GetBySlug(ctx context.Context, slug string) (*ent.Topic, error) {
	cacheKey := fmt.Sprintf("%s", slug)

	// Check cache first
	if cachedTopic, err := r.c.Get(ctx, cacheKey); err == nil {
		return cachedTopic, nil
	}

	// If not found in cache, query the database
	row, err := r.FindTopic(ctx, &structs.FindTopic{Slug: slug})

	if err != nil {
		log.Errorf(nil, "topicRepo.GetBySlug error: %v\n", err)
		return nil, err
	}

	// Cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(nil, "topicRepo.GetBySlug cache error: %v\n", err)
	}

	return row, nil
}

// Update updates a topic (full or partial).
func (r *topicRepo) Update(ctx context.Context, slug string, updates types.JSON) (*ent.Topic, error) {
	topic, err := r.FindTopic(ctx, &structs.FindTopic{Slug: slug})
	if err != nil {
		return nil, err
	}

	// create builder.
	builder := topic.Update()

	// Set values
	for field, value := range updates {
		switch field {
		case "name":
			builder.SetNillableName(types.ToPointer(value.(string)))
		case "title":
			builder.SetNillableTitle(types.ToPointer(value.(string)))
		case "slug":
			builder.SetNillableSlug(types.ToPointer(value.(string)))
		case "content":
			builder.SetNillableContent(types.ToPointer(value.(string)))
		case "thumbnail":
			builder.SetNillableThumbnail(types.ToPointer(value.(string)))
		case "temp":
			builder.SetTemp(value.(bool))
		case "markdown":
			builder.SetMarkdown(value.(bool))
		case "private":
			builder.SetPrivate(value.(bool))
		case "status":
			builder.SetStatus(value.(int32))
		case "released":
			builder.SetNillableReleased(types.ToPointer(value.(time.Time)))
		case "taxonomy_id":
			builder.SetNillableTaxonomyID(types.ToPointer(value.(string)))
		case "domain_id":
			builder.SetNillableDomainID(types.ToPointer(value.(string)))
		case "updated_by":
			builder.SetNillableUpdatedBy(types.ToPointer(value.(string)))
		}
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(nil, "topicRepo.Update error: %v\n", err)
		return nil, err
	}

	// Remove from cache
	cacheKey := fmt.Sprintf("%s", topic.ID)
	err = r.c.Delete(ctx, cacheKey)
	err = r.c.Delete(ctx, topic.Slug)
	if err != nil {
		log.Errorf(nil, "topicRepo.Update cache error: %v\n", err)
	}

	return row, nil
}

// List gets a list of topics.
func (r *topicRepo) List(ctx context.Context, p *structs.ListTopicParams) ([]*ent.Topic, error) {
	// create list builder
	builder, err := r.ListBuilder(ctx, p)
	if validator.IsNotNil(err) {
		return nil, err
	}

	// limit the result
	builder.Limit(int(p.Limit))

	// belong domain
	if p.Domain != "" {
		builder.Where(topicEnt.DomainIDEQ(p.Domain))
	}

	// sort
	builder.Order(ent.Desc(topicEnt.FieldCreatedAt))

	rows, err := builder.All(ctx)
	if err != nil {
		log.Errorf(nil, "topicRepo.List error: %v\n", err)
		return nil, err
	}

	return rows, nil
}

// Delete deletes a topic.
func (r *topicRepo) Delete(ctx context.Context, slug string) error {
	topic, err := r.FindTopic(ctx, &structs.FindTopic{Slug: slug})
	if err != nil {
		return err
	}

	// create builder.
	builder := r.ec.Topic.Delete()

	// execute the builder.
	_, err = builder.Where(topicEnt.IDEQ(topic.ID)).Exec(ctx)
	if err == nil {
		// Remove from cache
		cacheKey := fmt.Sprintf("%s", topic.Slug)
		err := r.c.Delete(ctx, cacheKey)
		err = r.c.Delete(ctx, topic.ID)
		if err != nil {
			log.Errorf(nil, "topicRepo.Delete cache error: %v\n", err)
		}
	}

	return err
}

// FindTopic finds a topic
func (r *topicRepo) FindTopic(ctx context.Context, p *structs.FindTopic) (*ent.Topic, error) {
	// create builder.
	builder := r.ec.Topic.Query()

	if validator.IsNotEmpty(p.ID) {
		builder = builder.Where(topicEnt.IDEQ(p.ID))
	}
	// support slug or ID
	if validator.IsNotEmpty(p.Slug) {
		builder = builder.Where(topicEnt.Or(
			topicEnt.ID(p.Slug),
			topicEnt.SlugEQ(p.Slug),
		))
	}
	if validator.IsNotEmpty(p.DomainID) {
		builder = builder.Where(topicEnt.DomainIDEQ(p.DomainID))
	}

	// execute the builder.
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// ListBuilder create list builder
func (r *topicRepo) ListBuilder(ctx context.Context, p *structs.ListTopicParams) (*ent.TopicQuery, error) {
	// verify query params.
	var next *ent.Topic
	if validator.IsNotEmpty(p.Cursor) {
		// query the address.
		row, err := r.FindTopic(ctx, &structs.FindTopic{
			ID: p.Cursor,
		})
		if validator.IsNotNil(err) || validator.IsNil(row) {
			return nil, err
		}
		next = row
	}

	// create builder.
	builder := r.ec.Topic.Query()

	// lt the cursor create time
	if next != nil {
		builder.Where(topicEnt.CreatedAtLT(next.CreatedAt))
	}

	return builder, nil
}

// CountX gets a count of topics.
func (r *topicRepo) CountX(ctx context.Context, p *structs.ListTopicParams) int {
	// create list builder
	builder, err := r.ListBuilder(ctx, p)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// Count gets a count of topics.
func (r *topicRepo) Count(ctx context.Context, p *structs.ListTopicParams) (int, error) {
	// create list builder
	builder, err := r.ListBuilder(ctx, p)
	if validator.IsNotNil(err) {
		return 0, err
	}
	return builder.Count(ctx)
}
