package handler

import (
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/types"
	"ncobase/helper"
	"ncobase/internal/data/structs"

	"github.com/gin-gonic/gin"
)

// CreatePermissionHandler handles the creation of a new permission.
//
// @Summary Create a new permission
// @Description Create a new permission with the provided data
// @Tags permissions
// @Accept json
// @Produce json
// @Param body body structs.CreatePermissionBody true "Permission data"
// @Success 200 {object} structs.ReadPermission "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/permissions [post]
// @Security Bearer
func (h *Handler) CreatePermissionHandler(c *gin.Context) {
	body := &structs.CreatePermissionBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.svc.CreatePermissionService(c, body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetPermissionHandler handles retrieving a permission by slug.
//
// @Summary Get a permission by slug or ID
// @Description Retrieve a permission by its slug or ID
// @Tags permissions
// @Produce json
// @Param slug path string true "Permission slug or ID"
// @Success 200 {object} structs.ReadPermission "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/permissions/{slug} [get]
// @Security Bearer
func (h *Handler) GetPermissionHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	result, err := h.svc.GetPermissionByIDService(c, slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// UpdatePermissionHandler handles updating an existing permission.
//
// @Summary Update an existing permission
// @Description Update an existing permission with the provided data
// @Tags permissions
// @Accept json
// @Produce json
// @Param slug path string true "Permission slug or ID"
// @Param body body types.JSON true "Permission data"
// @Success 200 {object} structs.ReadPermission "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/permissions/{slug} [put]
// @Security Bearer
func (h *Handler) UpdatePermissionHandler(c *gin.Context) {
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

	result, err := h.svc.UpdatePermissionService(c, slug, *updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// DeletePermissionHandler handles deleting a permission.
//
// @Summary Delete a permission by slug or ID
// @Description Delete a permission by its slug or ID
// @Tags permissions
// @Produce json
// @Param slug path string true "Permission slug or ID"
// @Success 200 {object} structs.ReadPermission "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/permissions/{slug} [delete]
// @Security Bearer
func (h *Handler) DeletePermissionHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	result, err := h.svc.DeletePermissionService(c, slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// ListPermissionHandler handles listing all permissions.
//
// @Summary List all permissions
// @Description Retrieve a list of permissions based on the provided query parameters
// @Tags permissions
// @Produce json
// @Param params query structs.ListPermissionParams true "List permissions parameters"
// @Success 200 {array} structs.ReadPermission "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/permissions [get]
// @Security Bearer
func (h *Handler) ListPermissionHandler(c *gin.Context) {
	params := &structs.ListPermissionParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	permissions, err := h.svc.ListPermissionsService(c, params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, permissions)
}
