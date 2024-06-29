package topic

import (
	"ncobase/internal/helper"
	"ncobase/plugin/content/service"
	"ncobase/plugin/content/structs"

	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/types"

	"github.com/gin-gonic/gin"
)

// HandlerInterface is the interface for the topic handler.
type HandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	List(c *gin.Context)
	Delete(c *gin.Context)
}

type Handler struct {
	s *service.Service
}

func New(s *service.Service) HandlerInterface {
	return &Handler{
		s: s,
	}
}

// Create  handles the creation of a topic.
//
// @Summary Create topic
// @Description Create a new topic.
// @Tags topic
// @Accept json
// @Produce json
// @Param body body structs.CreateTopicBody true "CreateTopicBody object"
// @Success 200 {object} structs.ReadTopic "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/topics [post]
// @Security Bearer
func (h *Handler) Create(c *gin.Context) {
	body := &structs.CreateTopicBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Topic.Create(c, body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Update  handles updating a topic (full and partial).
//
// @Summary Update topic
// @Description Update an existing topic, either fully or partially.
// @Tags topic
// @Accept json
// @Produce json
// @Param slug path string true "Topic slug"
// @Param body body structs.UpdateTopicBody true "UpdateTopicBody object"
// @Success 200 {object} structs.ReadTopic "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/topics/{slug} [put]
// @Security Bearer
func (h *Handler) Update(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	updates := &types.JSON{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, updates); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Topic.Update(c, slug, *updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get  handles getting a topic.
//
// @Summary Get topic
// @Description Retrieve details of a topic.
// @Tags topic
// @Produce json
// @Param slug path string true "Topic slug"
// @Success 200 {object} structs.ReadTopic "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/topics/{slug} [get]
func (h *Handler) Get(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	result, err := h.s.Topic.Get(c, slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete  handles deleting a topic.
//
// @Summary Delete topic
// @Description Delete an existing topic.
// @Tags topic
// @Produce json
// @Param slug path string true "Topic slug"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/topics/{slug} [delete]
// @Security Bearer
func (h *Handler) Delete(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	result, err := h.s.Topic.Delete(c, slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// List  handles listing topics.
//
// @Summary List topics
// @Description Retrieve a list of topics.
// @Tags topic
// @Produce json
// @Param params query structs.ListTopicParams true "List topics parameters"
// @Success 200 {array} structs.ReadTopic "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/topics [get]
func (h *Handler) List(c *gin.Context) {
	params := &structs.ListTopicParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	topics, err := h.s.Topic.List(c, params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, topics)
}
