package handler

import (
	"ncobase/content/service"
	"ncobase/content/structs"

	"github.com/ncobase/ncore/validation"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"

	"github.com/gin-gonic/gin"
)

// ChannelHandlerInterface is the interface for the handler.
type ChannelHandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	List(c *gin.Context)
	Delete(c *gin.Context)
}

// channelHandler represents the handler.
type channelHandler struct {
	s *service.Service
}

// NewChannelHandler creates a new handler.
func NewChannelHandler(s *service.Service) ChannelHandlerInterface {
	return &channelHandler{
		s: s,
	}
}

// Create handles the creation of a channel.
//
// @Summary Create channel
// @Description Create a new distribution channel.
// @Tags cms
// @Accept json
// @Produce json
// @Param body body structs.CreateChannelBody true "CreateChannelBody object"
// @Success 200 {object} structs.ReadChannel "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/channels [post]
// @Security Bearer
func (h *channelHandler) Create(c *gin.Context) {
	body := &structs.CreateChannelBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Channel.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Update handles updating a channel.
//
// @Summary Update channel
// @Description Update an existing distribution channel.
// @Tags cms
// @Accept json
// @Produce json
// @Param slug path string true "Channel slug"
// @Param body body structs.UpdateChannelBody true "UpdateChannelBody object"
// @Success 200 {object} structs.ReadChannel "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/channels/{slug} [put]
// @Security Bearer
func (h *channelHandler) Update(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	updates := &types.JSON{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, updates); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Channel.Update(c.Request.Context(), slug, *updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get handles getting a channel.
//
// @Summary Get channel
// @Description Retrieve details of a distribution channel.
// @Tags cms
// @Produce json
// @Param slug path string true "Channel slug"
// @Success 200 {object} structs.ReadChannel "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/channels/{slug} [get]
func (h *channelHandler) Get(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	result, err := h.s.Channel.Get(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete handles deleting a channel.
//
// @Summary Delete channel
// @Description Delete an existing distribution channel.
// @Tags cms
// @Produce json
// @Param slug path string true "Channel slug"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/channels/{slug} [delete]
// @Security Bearer
func (h *channelHandler) Delete(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	if err := h.s.Channel.Delete(c.Request.Context(), slug); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List handles listing channels.
//
// @Summary List channels
// @Description Retrieve a list of distribution channels.
// @Tags cms
// @Produce json
// @Param params query structs.ListChannelParams true "ListChannelParams object"
// @Success 200 {array} structs.ReadChannel "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/channels [get]
func (h *channelHandler) List(c *gin.Context) {
	params := &structs.ListChannelParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	channels, err := h.s.Channel.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, channels)
}
