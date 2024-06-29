package topic

import (
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/slug"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/internal/helper"
	"ncobase/plugin/content/data"
	"ncobase/plugin/content/data/ent"
	"ncobase/plugin/content/data/repository/topic"
	"ncobase/plugin/content/structs"

	"github.com/gin-gonic/gin"
)

// Interface is the interface for the topic service.
type Interface interface {
	Create(c *gin.Context, body *structs.CreateTopicBody) (*resp.Exception, error)
	Update(c *gin.Context, slug string, updates types.JSON) (*resp.Exception, error)
	Get(c *gin.Context, slug string) (*resp.Exception, error)
	List(c *gin.Context, params *structs.ListTopicParams) (*resp.Exception, error)
	Delete(c *gin.Context, slug string) (*resp.Exception, error)
}

type Service struct {
	topic topic.ITopic
}

func New(d *data.Data) Interface {
	return &Service{
		topic: topic.NewTopicRepo(d),
	}
}

// Create creates a new topic.
func (svc *Service) Create(c *gin.Context, body *structs.CreateTopicBody) (*resp.Exception, error) {
	// set slug field.
	if validator.IsEmpty(body.Slug) {
		body.Slug = slug.Unicode(body.Name)
	}
	row, err := svc.topic.Create(c, body)
	if exception, err := helper.HandleError("Topic", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: row,
	}, nil
}

// Update updates an existing topic (full and partial).
func (svc *Service) Update(c *gin.Context, slug string, updates types.JSON) (*resp.Exception, error) {
	if validator.IsEmpty(slug) {
		return resp.BadRequest(ecode.FieldIsRequired("slug / id")), nil
	}

	// Validate the updates map
	if len(updates) == 0 {
		return resp.BadRequest(ecode.FieldIsEmpty("updates fields")), nil
	}

	row, err := svc.topic.Update(c, slug, updates)
	if exception, err := helper.HandleError("Topic", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: row,
	}, nil
}

// Get retrieves a topic by ID.
func (svc *Service) Get(c *gin.Context, slug string) (*resp.Exception, error) {
	row, err := svc.topic.GetBySlug(c, slug)
	if exception, err := helper.HandleError("Topic", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: row,
	}, nil
}

// Delete deletes a topic by ID.
func (svc *Service) Delete(c *gin.Context, slug string) (*resp.Exception, error) {
	err := svc.topic.Delete(c, slug)
	if exception, err := helper.HandleError("Topic", err); exception != nil {
		return exception, err
	}

	return nil, nil
}

// List lists all topics.
func (svc *Service) List(c *gin.Context, params *structs.ListTopicParams) (*resp.Exception, error) {
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
