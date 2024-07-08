package topic

import (
	"context"
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/slug"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/feature/content/data"
	"ncobase/feature/content/data/ent"
	"ncobase/feature/content/data/repository/topic"
	"ncobase/feature/content/structs"
	"ncobase/helper"
)

// ServiceInterface is the interface for the topic service.
type ServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateTopicBody) (*resp.Exception, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*resp.Exception, error)
	Get(ctx context.Context, slug string) (*resp.Exception, error)
	List(ctx context.Context, params *structs.ListTopicParams) (*resp.Exception, error)
	Delete(ctx context.Context, slug string) (*resp.Exception, error)
}

type Service struct {
	topic topic.RepositoryInterface
}

func New(d *data.Data) ServiceInterface {
	return &Service{
		topic: topic.NewTopicRepo(d),
	}
}

// Create creates a new topic.
func (svc *Service) Create(ctx context.Context, body *structs.CreateTopicBody) (*resp.Exception, error) {
	// set slug field.
	if validator.IsEmpty(body.Slug) {
		body.Slug = slug.Unicode(body.Name)
	}
	row, err := svc.topic.Create(ctx, body)
	if exception, err := helper.HandleError("Topic", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: row,
	}, nil
}

// Update updates an existing topic (full and partial).
func (svc *Service) Update(ctx context.Context, slug string, updates types.JSON) (*resp.Exception, error) {
	if validator.IsEmpty(slug) {
		return resp.BadRequest(ecode.FieldIsRequired("slug / id")), nil
	}

	// Validate the updates map
	if len(updates) == 0 {
		return resp.BadRequest(ecode.FieldIsEmpty("updates fields")), nil
	}

	row, err := svc.topic.Update(ctx, slug, updates)
	if exception, err := helper.HandleError("Topic", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: row,
	}, nil
}

// Get retrieves a topic by ID.
func (svc *Service) Get(ctx context.Context, slug string) (*resp.Exception, error) {
	row, err := svc.topic.GetBySlug(ctx, slug)
	if exception, err := helper.HandleError("Topic", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: row,
	}, nil
}

// Delete deletes a topic by ID.
func (svc *Service) Delete(ctx context.Context, slug string) (*resp.Exception, error) {
	err := svc.topic.Delete(ctx, slug)
	if exception, err := helper.HandleError("Topic", err); exception != nil {
		return exception, err
	}

	return nil, nil
}

// List lists all topics.
func (svc *Service) List(ctx context.Context, params *structs.ListTopicParams) (*resp.Exception, error) {
	// limit default value
	if validator.IsEmpty(params.Limit) {
		params.Limit = 20
	}
	// limit must less than 100
	if params.Limit > 100 {
		return resp.BadRequest(ecode.FieldIsInvalid("limit")), nil
	}

	rows, err := svc.topic.List(ctx, params)

	if ent.IsNotFound(err) {
		return resp.NotFound(ecode.FieldIsInvalid("cursor")), nil
	}
	if validator.IsNotNil(err) {
		return resp.InternalServer(err.Error()), nil
	}

	total := svc.topic.CountX(ctx, params)

	return &resp.Exception{
		Data: &types.JSON{
			"content": rows,
			"total":   total,
		},
	}, nil
}
