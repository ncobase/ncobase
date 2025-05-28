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

// MediaHandlerInterface is the interface for the handler.
type MediaHandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	List(c *gin.Context)
	Delete(c *gin.Context)
}

// mediaHandler represents the handler.
type mediaHandler struct {
	s *service.Service
}

// NewMediaHandler creates a new handler.
func NewMediaHandler(s *service.Service) MediaHandlerInterface {
	return &mediaHandler{
		s: s,
	}
}

// Create handles the creation of a media.
//
// @Summary Create media
// @Description Create a new media resource.
// @Tags cms
// @Accept json
// @Produce json
// @Param body body structs.CreateMediaBody true "CreateMediaBody object"
// @Success 200 {object} structs.ReadMedia "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/media [post]
// @Security Bearer
func (h *mediaHandler) Create(c *gin.Context) {
	body := &structs.CreateMediaBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Media.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Update handles updating a media.
//
// @Summary Update media
// @Description Update an existing media resource.
// @Tags cms
// @Accept json
// @Produce json
// @Param id path string true "Media ID"
// @Param body body structs.UpdateMediaBody true "UpdateMediaBody object"
// @Success 200 {object} structs.ReadMedia "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/media/{id} [put]
// @Security Bearer
func (h *mediaHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
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

	result, err := h.s.Media.Update(c.Request.Context(), id, *updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get handles getting a media.
//
// @Summary Get media
// @Description Retrieve details of a media resource.
// @Tags cms
// @Produce json
// @Param id path string true "Media ID"
// @Success 200 {object} structs.ReadMedia "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/media/{id} [get]
func (h *mediaHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.s.Media.Get(c.Request.Context(), id)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete handles deleting a media.
//
// @Summary Delete media
// @Description Delete an existing media resource.
// @Tags cms
// @Produce json
// @Param id path string true "Media ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/media/{id} [delete]
// @Security Bearer
func (h *mediaHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.s.Media.Delete(c.Request.Context(), id); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List handles listing media resources.
//
// @Summary List media
// @Description Retrieve a list of media resources.
// @Tags cms
// @Produce json
// @Param params query structs.ListMediaParams true "List media parameters"
// @Success 200 {array} structs.ReadMedia "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/media [get]
func (h *mediaHandler) List(c *gin.Context) {
	params := &structs.ListMediaParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	media, err := h.s.Media.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, media)
}
