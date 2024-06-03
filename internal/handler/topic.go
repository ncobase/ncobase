package handler

import (
	"stocms/internal/data/structs"
	"stocms/pkg/resp"

	"github.com/gin-gonic/gin"
)

// CreateTopicHandler handles creation of a topic
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

// UpdateTopicHandler handles updating a topic
func (h *Handler) UpdateTopicHandler(c *gin.Context) {
	var body *structs.UpdateTopicBody
	if err := c.ShouldBind(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.svc.UpdateTopicService(c, body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetTopicHandler handles getting a topic
func (h *Handler) GetTopicHandler(c *gin.Context) {
	slug := c.Param("slug")

	result, err := h.svc.GetTopicService(c, slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// DeleteTopicHandler handles deleting a topic
func (h *Handler) DeleteTopicHandler(c *gin.Context) {
	slug := c.Param("slug")

	result, err := h.svc.DeleteTopicService(c, slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// ListTopicHandler handles listing topics
func (h *Handler) ListTopicHandler(c *gin.Context) {
	params := &structs.ListTopicParams{}
	if err := c.ShouldBindQuery(&params); err != nil {
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
