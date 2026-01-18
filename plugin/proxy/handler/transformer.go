package handler

import (
	"ncobase/plugin/proxy/service"
	"ncobase/plugin/proxy/structs"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// TransformerHandlerInterface is the interface for the transformer handler.
type TransformerHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
}

// transformerHandler represents the transformer handler.
type transformerHandler struct {
	s *service.Service
}

// NewTransformerHandler creates a new transformer handler.
func NewTransformerHandler(svc *service.Service) TransformerHandlerInterface {
	return &transformerHandler{
		s: svc,
	}
}

// Create handles the creation of a new transformer.
//
// @Summary Create data transformer
// @Description Create a new data transformer for API requests/responses
// @Tags proxy
// @Accept json
// @Produce json
// @Param body body structs.CreateTransformerBody true "Transformer data"
// @Success 200 {object} structs.ReadTransformer "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /tbp/transformers [post]
// @Security Bearer
func (h *transformerHandler) Create(c *gin.Context) {
	body := &structs.CreateTransformerBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Transformer.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get handles retrieving a transformer by ID.
//
// @Summary Get a transformer by ID
// @Description Retrieve a transformer by its ID
// @Tags proxy
// @Produce json
// @Param id path string true "Transformer ID"
// @Success 200 {object} structs.ReadTransformer "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /tbp/transformers/{id} [get]
// @Security Bearer
func (h *transformerHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.s.Transformer.GetByID(c.Request.Context(), id)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Update handles updating an existing transformer.
//
// @Summary Update an existing transformer
// @Description Update an existing transformer with the provided data
// @Tags proxy
// @Accept json
// @Produce json
// @Param id path string true "Transformer ID"
// @Param body body types.JSON true "Transformer data"
// @Success 200 {object} structs.ReadTransformer "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /tbp/transformers/{id} [put]
// @Security Bearer
func (h *transformerHandler) Update(c *gin.Context) {
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

	result, err := h.s.Transformer.Update(c.Request.Context(), id, *updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete handles deleting a transformer.
//
// @Summary Delete a transformer by ID
// @Description Delete a transformer by its ID
// @Tags proxy
// @Produce json
// @Param id path string true "Transformer ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /tbp/transformers/{id} [delete]
// @Security Bearer
func (h *transformerHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	err := h.s.Transformer.Delete(c.Request.Context(), id)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List handles listing all transformers.
//
// @Summary List all transformers
// @Description Retrieve a list of transformers based on the provided query parameters
// @Tags proxy
// @Produce json
// @Param params query structs.ListTransformerParams true "List parameters"
// @Success 200 {array} structs.ReadTransformer "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /tbp/transformers [get]
// @Security Bearer
func (h *transformerHandler) List(c *gin.Context) {
	params := &structs.ListTransformerParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	transformers, err := h.s.Transformer.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, transformers)
}
