package handler

import (
	"stocms/internal/data/structs"
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
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /topic [post]
func (h *Handler) CreateTopicHandler(c *gin.Context) {
	var body *structs.CreateTopicBody
	if err := c.ShouldBind(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
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
// @Param body body object true "Update data"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /topic/{slug} [put]
func (h *Handler) UpdateTopicHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	var updates types.JSON
	if err := c.ShouldBindJSON(&updates); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.svc.UpdateTopicService(c, slug, updates)
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
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /topic/{slug} [get]
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
// @Router /topic/{slug} [delete]
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
// @Param category query string false "Category filter"
// @Param limit query int false "Result limit"
// @Param offset query int false "Result offset"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /topic [get]
func (h *Handler) ListTopicHandler(c *gin.Context) {
	params := &structs.ListTopicParams{}
	if err := c.ShouldBindQuery(params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	if err := params.Validate(); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	topics, err := h.svc.ListTopicsService(c, params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, topics)
}
