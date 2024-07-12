package service

import (
	"context"
	"errors"
	"ncobase/common/ecode"
	"ncobase/common/slug"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/feature/content/data"
	"ncobase/feature/content/data/ent"
	"ncobase/feature/content/data/repository"
	"ncobase/feature/content/structs"
	"ncobase/helper"
)

// TopicServiceInterface is the interface for the service.
type TopicServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateTopicBody) (*structs.ReadTopic, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*structs.ReadTopic, error)
	Get(ctx context.Context, slug string) (*structs.ReadTopic, error)
	List(ctx context.Context, params *structs.ListTopicParams) (*types.JSON, error)
	Delete(ctx context.Context, slug string) error
}

// topicService is the struct for the service.
type topicService struct {
	topic repository.TopicRepositoryInterface
}

// NewTopicService creates a new service.
func NewTopicService(d *data.Data) TopicServiceInterface {
	return &topicService{
		topic: repository.NewTopicRepository(d),
	}
}

// Create creates a new topic.
func (s *topicService) Create(ctx context.Context, body *structs.CreateTopicBody) (*structs.ReadTopic, error) {
	// set slug field.
	if validator.IsEmpty(body.Slug) {
		body.Slug = slug.Unicode(body.Name)
	}
	row, err := s.topic.Create(ctx, body)
	if err := handleEntError("Topic", err); err != nil {
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

	row, err := s.topic.Update(ctx, slug, updates)
	if err := handleEntError("Topic", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Get retrieves a topic by ID.
func (s *topicService) Get(ctx context.Context, slug string) (*structs.ReadTopic, error) {
	row, err := s.topic.GetBySlug(ctx, slug)
	if err := handleEntError("Topic", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Delete deletes a topic by ID.
func (s *topicService) Delete(ctx context.Context, slug string) error {
	err := s.topic.Delete(ctx, slug)
	if err := handleEntError("Topic", err); err != nil {
		return err
	}

	return nil
}

// List lists all topics.
func (s *topicService) List(ctx context.Context, params *structs.ListTopicParams) (*types.JSON, error) {
	// limit default value
	if validator.IsEmpty(params.Limit) {
		params.Limit = 20
	}
	// limit must less than 100
	if params.Limit > 100 {
		return nil, errors.New(ecode.FieldIsInvalid("limit"))
	}

	rows, err := s.topic.List(ctx, params)

	if ent.IsNotFound(err) {
		return nil, errors.New(ecode.FieldIsInvalid("cursor"))
	}
	if validator.IsNotNil(err) {
		return nil, err
	}

	total := s.topic.CountX(ctx, params)

	return &types.JSON{
		"current_tenant_id": helper.GetTenantID(ctx),
		"current_user_id":   helper.GetUserID(ctx),
		"content":           rows,
		"total":             total,
	}, nil
}

// Serializes serializes topics.
func (s *topicService) Serializes(rows []*ent.Topic) []*structs.ReadTopic {
	var rs []*structs.ReadTopic
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
