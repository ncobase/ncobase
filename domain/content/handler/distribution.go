package handler

import (
	"ncobase/domain/content/service"
	"ncobase/domain/content/structs"

	"github.com/ncobase/ncore/validation"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"

	"github.com/gin-gonic/gin"
)

// DistributionHandlerInterface is the interface for the handler.
type DistributionHandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	List(c *gin.Context)
	Delete(c *gin.Context)
	Publish(c *gin.Context)
	Cancel(c *gin.Context)
}

// distributionHandler represents the handler.
type distributionHandler struct {
	s *service.Service
}

// NewDistributionHandler creates a new handler.
func NewDistributionHandler(s *service.Service) DistributionHandlerInterface {
	return &distributionHandler{
		s: s,
	}
}

// Create handles the creation of a distribution.
//
// @Summary Create distribution
// @Description Create a new content distribution.
// @Tags cms
// @Accept json
// @Produce json
// @Param body body structs.CreateDistributionBody true "CreateDistributionBody object"
// @Success 200 {object} structs.ReadDistribution "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/distributions [post]
// @Security Bearer
func (h *distributionHandler) Create(c *gin.Context) {
	body := &structs.CreateDistributionBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Distribution.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Update handles updating a distribution.
//
// @Summary Update distribution
// @Description Update an existing content distribution.
// @Tags cms
// @Accept json
// @Produce json
// @Param id path string true "Distribution ID"
// @Param body body structs.UpdateDistributionBody true "UpdateDistributionBody object"
// @Success 200 {object} structs.ReadDistribution "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/distributions/{id} [put]
// @Security Bearer
func (h *distributionHandler) Update(c *gin.Context) {
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

	result, err := h.s.Distribution.Update(c.Request.Context(), id, *updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get handles getting a distribution.
//
// @Summary Get distribution
// @Description Retrieve details of a content distribution.
// @Tags cms
// @Produce json
// @Param id path string true "Distribution ID"
// @Success 200 {object} structs.ReadDistribution "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/distributions/{id} [get]
func (h *distributionHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.s.Distribution.Get(c.Request.Context(), id)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete handles deleting a distribution.
//
// @Summary Delete distribution
// @Description Delete an existing content distribution.
// @Tags cms
// @Produce json
// @Param id path string true "Distribution ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/distributions/{id} [delete]
// @Security Bearer
func (h *distributionHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.s.Distribution.Delete(c.Request.Context(), id); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List handles listing distributions.
//
// @Summary List distributions
// @Description Retrieve a list of content distributions.
// @Tags cms
// @Produce json
// @Param params query structs.ListDistributionParams true "List distributions parameters"
// @Success 200 {array} structs.ReadDistribution "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/distributions [get]
func (h *distributionHandler) List(c *gin.Context) {
	params := &structs.ListDistributionParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	distributions, err := h.s.Distribution.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, distributions)
}

// Publish handles publishing a distribution.
//
// @Summary Publish distribution
// @Description Publish a content distribution immediately.
// @Tags cms
// @Produce json
// @Param id path string true "Distribution ID"
// @Success 200 {object} structs.ReadDistribution "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/distributions/{id}/publish [post]
// @Security Bearer
func (h *distributionHandler) Publish(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.s.Distribution.Publish(c.Request.Context(), id)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Cancel handles cancelling a distribution.
//
// @Summary Cancel distribution
// @Description Cancel a scheduled content distribution.
// @Tags cms
// @Accept json
// @Produce json
// @Param id path string true "Distribution ID"
// @Param reason body object{Reason string `json:"reason"`} true "Reason for cancellation"
// @Success 200 {object} structs.ReadDistribution "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/distributions/{id}/cancel [post]
// @Security Bearer
func (h *distributionHandler) Cancel(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	var body struct {
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.s.Distribution.Cancel(c.Request.Context(), id, body.Reason)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}
