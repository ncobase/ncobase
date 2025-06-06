package service

import (
	"context"
	"errors"
	"ncobase/content/data"
	"ncobase/content/data/ent"
	"ncobase/content/data/repository"
	"ncobase/content/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/slug"
	"github.com/ncobase/ncore/validation/validator"
)

// TopicServiceInterface is the interface for the service.
type TopicServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateTopicBody) (*structs.ReadTopic, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*structs.ReadTopic, error)
	Get(ctx context.Context, slug string) (*structs.ReadTopic, error)
	List(ctx context.Context, params *structs.ListTopicParams) (paging.Result[*structs.ReadTopic], error)
	Delete(ctx context.Context, slug string) error
}

// topicService is the struct for the service.
type topicService struct {
	r repository.TopicRepositoryInterface
}

// NewTopicService creates a new service.
func NewTopicService(d *data.Data) TopicServiceInterface {
	return &topicService{
		r: repository.NewTopicRepository(d),
	}
}

// Create creates a new topic.
func (s *topicService) Create(ctx context.Context, body *structs.CreateTopicBody) (*structs.ReadTopic, error) {
	// set slug field.
	if validator.IsEmpty(body.Slug) {
		body.Slug = slug.Unicode(body.Name)
	}
	row, err := s.r.Create(ctx, body)
	if err := handleEntError(ctx, "Topic", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Update updates an existing topic (full and partial).
func (s *topicService) Update(ctx context.Context, slug string, updates types.JSON) (*structs.ReadTopic, error) {
	if validator.IsEmpty(slug) {
		return nil, errors.New(ecode.FieldIsRequired("slug / id"))
	}

	// Validate the updates map
	if len(updates) == 0 {
		return nil, errors.New(ecode.FieldIsEmpty("updates fields"))
	}

	row, err := s.r.Update(ctx, slug, updates)
	if err := handleEntError(ctx, "Topic", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Get retrieves a topic by ID.
func (s *topicService) Get(ctx context.Context, slug string) (*structs.ReadTopic, error) {
	row, err := s.r.GetBySlug(ctx, slug)
	if err := handleEntError(ctx, "Topic", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Delete deletes a topic by ID.
func (s *topicService) Delete(ctx context.Context, slug string) error {
	err := s.r.Delete(ctx, slug)
	if err := handleEntError(ctx, "Topic", err); err != nil {
		return err
	}

	return nil
}

// List lists all topics.
func (s *topicService) List(ctx context.Context, params *structs.ListTopicParams) (paging.Result[*structs.ReadTopic], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadTopic, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, err := s.r.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			logger.Errorf(ctx, "Error listing topics: %v", err)
			return nil, 0, err
		}

		total := s.r.CountX(ctx, params)

		return s.Serializes(rows), total, nil
	})
}

// Serializes serializes topics.
func (s *topicService) Serializes(rows []*ent.Topic) []*structs.ReadTopic {
	rs := make([]*structs.ReadTopic, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize serializes a topic.
func (s *topicService) Serialize(row *ent.Topic) *structs.ReadTopic {
	return &structs.ReadTopic{
		ID:         row.ID,
		Name:       row.Name,
		Title:      row.Title,
		Slug:       row.Slug,
		Content:    row.Content,
		Thumbnail:  row.Thumbnail,
		Temp:       row.Temp,
		Markdown:   row.Markdown,
		Private:    row.Private,
		Status:     row.Status,
		Released:   row.Released,
		TaxonomyID: row.TaxonomyID,
		TenantID:   row.TenantID,
		CreatedBy:  &row.CreatedBy,
		CreatedAt:  &row.CreatedAt,
		UpdatedBy:  &row.UpdatedBy,
		UpdatedAt:  &row.UpdatedAt,
	}
}
