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

// EndpointHandlerInterface is the interface for the endpoint handler.
type EndpointHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
}

// endpointHandler represents the endpoint handler.
type endpointHandler struct {
	s *service.Service
}

// NewEndpointHandler creates a new endpoint handler.
func NewEndpointHandler(svc *service.Service) EndpointHandlerInterface {
	return &endpointHandler{
		s: svc,
	}
}

// Create handles the creation of a new endpoint.
//
// @Summary Create API endpoint
// @Description Create a new third-party API endpoint
// @Tags proxy
// @Accept json
// @Produce json
// @Param body body structs.CreateEndpointBody true "Endpoint data"
// @Success 200 {object} structs.ReadEndpoint "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /tbp/endpoints [post]
// @Security Bearer
func (h *endpointHandler) Create(c *gin.Context) {
	body := &structs.CreateEndpointBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Endpoint.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get handles retrieving an endpoint by ID.
//
// @Summary Get an endpoint by ID
// @Description Retrieve an endpoint by its ID
// @Tags proxy
// @Produce json
// @Param id path string true "Endpoint ID"
// @Success 200 {object} structs.ReadEndpoint "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /tbp/endpoints/{id} [get]
// @Security Bearer
func (h *endpointHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.s.Endpoint.GetByID(c.Request.Context(), id)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Update handles updating an existing endpoint.
//
// @Summary Update an existing endpoint
// @Description Update an existing endpoint with the provided data
// @Tags proxy
// @Accept json
// @Produce json
// @Param id path string true "Endpoint ID"
// @Param body body types.JSON true "Endpoint data"
// @Success 200 {object} structs.ReadEndpoint "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /tbp/endpoints/{id} [put]
// @Security Bearer
func (h *endpointHandler) Update(c *gin.Context) {
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

	result, err := h.s.Endpoint.Update(c.Request.Context(), id, *updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete handles deleting an endpoint.
//
// @Summary Delete an endpoint by ID
// @Description Delete an endpoint by its ID
// @Tags proxy
// @Produce json
// @Param id path string true "Endpoint ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /tbp/endpoints/{id} [delete]
// @Security Bearer
func (h *endpointHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	err := h.s.Endpoint.Delete(c.Request.Context(), id)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List handles listing all endpoints.
//
// @Summary List all endpoints
// @Description Retrieve a list of endpoints based on the provided query parameters
// @Tags proxy
// @Produce json
// @Param params query structs.ListEndpointParams true "List parameters"
// @Success 200 {array} structs.ReadEndpoint "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /tbp/endpoints [get]
// @Security Bearer
func (h *endpointHandler) List(c *gin.Context) {
	params := &structs.ListEndpointParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	endpoints, err := h.s.Endpoint.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, endpoints)
}
