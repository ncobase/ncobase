package handler

import (
	"stocms/internal/data/structs"
	"stocms/internal/helper"
	"stocms/pkg/ecode"
	"stocms/pkg/resp"
	"stocms/pkg/types"

	"github.com/gin-gonic/gin"
)

// CreateTopicHandler handles the creation of a topic.
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
func (h *Handler) CreateTopicHandler(c *gin.Context) {
	body := &structs.CreateTopicBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.svc.CreateTopicService(c, body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// UpdateTopicHandler handles updating a topic (full and partial).
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
func (h *Handler) UpdateTopicHandler(c *gin.Context) {
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

	result, err := h.svc.UpdateTopicService(c, slug, *updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetTopicHandler handles getting a topic.
//
// @Summary Get topic
// @Description Retrieve details of a topic.
// @Tags topic
// @Produce json
// @Param slug path string true "Topic slug"
// @Success 200 {object} structs.ReadTopic "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/topics/{slug} [get]
func (h *Handler) GetTopicHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	result, err := h.svc.GetTopicService(c, slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// DeleteTopicHandler handles deleting a topic.
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
func (h *Handler) DeleteTopicHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	result, err := h.svc.DeleteTopicService(c, slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// ListTopicHandler handles listing topics.
//
// @Summary List topics
// @Description Retrieve a list of topics.
// @Tags topic
// @Produce json
// @Param params query structs.ListTopicParams true "List topics parameters"
// @Success 200 {array} structs.ReadTopic "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/topics [get]
func (h *Handler) ListTopicHandler(c *gin.Context) {
	params := &structs.ListTopicParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	validationErrors := structs.Validate(params)
	if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	topics, err := h.svc.ListTopicsService(c, params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, topics)
}
