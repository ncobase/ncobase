package handler

import (
	"github.com/ncobase/ncore/pkg/helper"
	"ncobase/domain/content/service"
	"ncobase/domain/content/structs"

	"github.com/ncobase/ncore/pkg/ecode"
	"github.com/ncobase/ncore/pkg/resp"
	"github.com/ncobase/ncore/pkg/types"

	"github.com/gin-gonic/gin"
)

// TopicHandlerInterface is the interface for the handler.
type TopicHandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	List(c *gin.Context)
	Delete(c *gin.Context)
}

// topicHandler represents the handler.
type topicHandler struct {
	s *service.Service
}

// NewTopicHandler creates a new handler.
func NewTopicHandler(s *service.Service) TopicHandlerInterface {
	return &topicHandler{
		s: s,
	}
}

// Create  handles the creation of a topic.
//
// @Summary Create topic
// @Description Create a new topic.
// @Tags cms
// @Accept json
// @Produce json
// @Param body body structs.CreateTopicBody true "CreateTopicBody object"
// @Success 200 {object} structs.ReadTopic "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/topics [post]
// @Security Bearer
func (h *topicHandler) Create(c *gin.Context) {
	body := &structs.CreateTopicBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Topic.Create(c.Request.Context(), body)
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
// @Tags cms
// @Accept json
// @Produce json
// @Param slug path string true "Topic slug"
// @Param body body structs.UpdateTopicBody true "UpdateTopicBody object"
// @Success 200 {object} structs.ReadTopic "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/topics/{slug} [put]
// @Security Bearer
func (h *topicHandler) Update(c *gin.Context) {
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

	result, err := h.s.Topic.Update(c.Request.Context(), slug, *updates)
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
// @Tags cms
// @Produce json
// @Param slug path string true "Topic slug"
// @Success 200 {object} structs.ReadTopic "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/topics/{slug} [get]
func (h *topicHandler) Get(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	result, err := h.s.Topic.Get(c.Request.Context(), slug)
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
// @Tags cms
// @Produce json
// @Param slug path string true "Topic slug"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/topics/{slug} [delete]
// @Security Bearer
func (h *topicHandler) Delete(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	if err := h.s.Topic.Delete(c.Request.Context(), slug); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List  handles listing topics.
//
// @Summary List topics
// @Description Retrieve a list of topics.
// @Tags cms
// @Produce json
// @Param params query structs.ListTopicParams true "List topics parameters"
// @Success 200 {array} structs.ReadTopic "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/topics [get]
func (h *topicHandler) List(c *gin.Context) {
	params := &structs.ListTopicParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	topics, err := h.s.Topic.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, topics)
}
