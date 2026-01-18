package handler

import (
	"ncobase/core/access/service"
	"ncobase/core/access/structs"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// PermissionHandlerInterface is the interface for the handler.
type PermissionHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
}

// permissionHandler represents the handler.
type permissionHandler struct {
	s *service.Service
}

// NewPermissionHandler creates a new handler.
func NewPermissionHandler(svc *service.Service) PermissionHandlerInterface {
	return &permissionHandler{
		s: svc,
	}
}

// Create handles the creation of a new permission.
//
// @Summary Create a new permission
// @Description Create a new permission with the provided data
// @Tags sys
// @Accept json
// @Produce json
// @Param body body structs.CreatePermissionBody true "Permission data"
// @Success 200 {object} structs.ReadPermission "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/permissions [post]
// @Security Bearer
func (h *permissionHandler) Create(c *gin.Context) {
	body := &structs.CreatePermissionBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Permission.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get handles retrieving a permission by slug.
//
// @Summary Get a permission by slug or ID
// @Description Retrieve a permission by its slug or ID
// @Tags sys
// @Produce json
// @Param slug path string true "Permission slug or ID"
// @Success 200 {object} structs.ReadPermission "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/permissions/{slug} [get]
// @Security Bearer
func (h *permissionHandler) Get(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	result, err := h.s.Permission.GetByID(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Update handles updating an existing permission.
//
// @Summary Update an existing permission
// @Description Update an existing permission with the provided data
// @Tags sys
// @Accept json
// @Produce json
// @Param slug path string true "Permission slug or ID"
// @Param body body types.JSON true "Permission data"
// @Success 200 {object} structs.ReadPermission "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/permissions/{slug} [put]
// @Security Bearer
func (h *permissionHandler) Update(c *gin.Context) {
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

	result, err := h.s.Permission.Update(c.Request.Context(), slug, *updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete handles deleting a permission.
//
// @Summary Delete a permission by slug or ID
// @Description Delete a permission by its slug or ID
// @Tags sys
// @Produce json
// @Param slug path string true "Permission slug or ID"
// @Success 200 {object} structs.ReadPermission "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/permissions/{slug} [delete]
// @Security Bearer
func (h *permissionHandler) Delete(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	if err := h.s.Permission.Delete(c.Request.Context(), slug); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List handles listing all permissions.
//
// @Summary List all permissions
// @Description Retrieve a list of permissions based on the provided query parameters
// @Tags sys
// @Produce json
// @Param params query structs.ListPermissionParams true "List permissions parameters"
// @Success 200 {array} structs.ReadPermission "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/permissions [get]
// @Security Bearer
func (h *permissionHandler) List(c *gin.Context) {
	params := &structs.ListPermissionParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	permissions, err := h.s.Permission.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, permissions)
}
