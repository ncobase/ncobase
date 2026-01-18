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

// RouteHandlerInterface is the interface for the route handler.
type RouteHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
}

// routeHandler represents the route handler.
type routeHandler struct {
	s *service.Service
}

// NewRouteHandler creates a new route handler.
func NewRouteHandler(svc *service.Service) RouteHandlerInterface {
	return &routeHandler{
		s: svc,
	}
}

// Create handles the creation of a new route.
//
// @Summary Create API route
// @Description Create a new proxy route for a third-party API endpoint
// @Tags proxy
// @Accept json
// @Produce json
// @Param body body structs.CreateRouteBody true "Route data"
// @Success 200 {object} structs.ReadRoute "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /tbp/routes [post]
// @Security Bearer
func (h *routeHandler) Create(c *gin.Context) {
	body := &structs.CreateRouteBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Route.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get handles retrieving a route by ID.
//
// @Summary Get a route by ID
// @Description Retrieve a route by its ID
// @Tags proxy
// @Produce json
// @Param id path string true "Route ID"
// @Success 200 {object} structs.ReadRoute "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /tbp/routes/{id} [get]
// @Security Bearer
func (h *routeHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.s.Route.GetByID(c.Request.Context(), id)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Update handles updating an existing route.
//
// @Summary Update an existing route
// @Description Update an existing route with the provided data
// @Tags proxy
// @Accept json
// @Produce json
// @Param id path string true "Route ID"
// @Param body body types.JSON true "Route data"
// @Success 200 {object} structs.ReadRoute "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /tbp/routes/{id} [put]
// @Security Bearer
func (h *routeHandler) Update(c *gin.Context) {
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

	result, err := h.s.Route.Update(c.Request.Context(), id, *updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete handles deleting a route.
//
// @Summary Delete a route by ID
// @Description Delete a route by its ID
// @Tags proxy
// @Produce json
// @Param id path string true "Route ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /tbp/routes/{id} [delete]
// @Security Bearer
func (h *routeHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	err := h.s.Route.Delete(c.Request.Context(), id)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List handles listing all routes.
//
// @Summary List all routes
// @Description Retrieve a list of routes based on the provided query parameters
// @Tags proxy
// @Produce json
// @Param params query structs.ListRouteParams true "List parameters"
// @Success 200 {array} structs.ReadRoute "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /tbp/routes [get]
// @Security Bearer
func (h *routeHandler) List(c *gin.Context) {
	params := &structs.ListRouteParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	routes, err := h.s.Route.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, routes)
}
