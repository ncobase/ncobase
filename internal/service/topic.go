package service

import (
	"ncobase/internal/data/ent"
	"ncobase/internal/data/structs"

	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/slug"
	"ncobase/common/types"
	"ncobase/common/validator"

	"github.com/gin-gonic/gin"
)

// CreateTopicService creates a new topic.
func (svc *Service) CreateTopicService(c *gin.Context, body *structs.CreateTopicBody) (*resp.Exception, error) {
	// set slug field.
	if validator.IsEmpty(body.Slug) {
		body.Slug = slug.Unicode(body.Name)
	}
	topic, err := svc.topic.Create(c, body)
	if exception, err := handleError("Topic", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: topic,
	}, nil
}

// UpdateTopicService updates an existing topic (full and partial).
func (svc *Service) UpdateTopicService(c *gin.Context, slug string, updates types.JSON) (*resp.Exception, error) {
	if validator.IsEmpty(slug) {
		return resp.BadRequest(ecode.FieldIsRequired("slug / id")), nil
	}

	// Validate the updates map
	if len(updates) == 0 {
		return resp.BadRequest(ecode.FieldIsEmpty("updates fields")), nil
	}

	topic, err := svc.topic.Update(c, slug, updates)
	if exception, err := handleError("Topic", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: topic,
	}, nil
}

// GetTopicService retrieves a topic by ID.
func (svc *Service) GetTopicService(c *gin.Context, slug string) (*resp.Exception, error) {
	topic, err := svc.topic.GetBySlug(c, slug)
	if exception, err := handleError("Topic", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: topic,
	}, nil
}

// DeleteTopicService deletes a topic by ID.
func (svc *Service) DeleteTopicService(c *gin.Context, slug string) (*resp.Exception, error) {
	err := svc.topic.Delete(c, slug)
	if exception, err := handleError("Topic", err); exception != nil {
		return exception, err
	}

	return nil, nil
}

// ListTopicsService lists all topics.
func (svc *Service) ListTopicsService(c *gin.Context, params *structs.ListTopicParams) (*resp.Exception, error) {
	// limit default value
	if validator.IsEmpty(params.Limit) {
		params.Limit = 20
	}
	// limit must less than 100
	if params.Limit > 100 {
		return resp.BadRequest(ecode.FieldIsInvalid("limit")), nil
	}

	rows, err := svc.topic.List(c, params)

	if ent.IsNotFound(err) {
		return resp.NotFound(ecode.FieldIsInvalid("cursor")), nil
	}
	if validator.IsNotNil(err) {
		return resp.InternalServer(err.Error()), nil
	}

	total := svc.topic.CountX(c, params)

	return &resp.Exception{
		Data: &types.JSON{
			"content": rows,
			"total":   total,
		},
	}, nil
}
