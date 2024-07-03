package handler

import (
	"ncobase/app/data/structs"
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/types"
	"ncobase/helper"

	"github.com/gin-gonic/gin"
)

// CreateRoleHandler handles the creation of a new role.
//
// @Summary Create a new role
// @Description Create a new role with the provided data
// @Tags roles
// @Accept json
// @Produce json
// @Param body body structs.CreateRoleBody true "Role data"
// @Success 200 {object} structs.ReadRole "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/roles [post]
// @Security Bearer
func (h *Handler) CreateRoleHandler(c *gin.Context) {
	body := &structs.CreateRoleBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.svc.CreateRoleService(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetRoleHandler handles retrieving a role by slug.
//
// @Summary Get a role by slug or ID
// @Description Retrieve a role by its slug or ID
// @Tags roles
// @Produce json
// @Param slug path string true "Role slug or ID"
// @Success 200 {object} structs.ReadRole "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/roles/{slug} [get]
// @Security Bearer
func (h *Handler) GetRoleHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	result, err := h.svc.GetRoleByIDService(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// UpdateRoleHandler handles updating an existing role.
//
// @Summary Update an existing role
// @Description Update an existing role with the provided data
// @Tags roles
// @Accept json
// @Produce json
// @Param slug path string true "Role slug or ID"
// @Param body body types.JSON true "Role data"
// @Success 200 {object} structs.ReadRole "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/roles/{slug} [put]
// @Security Bearer
func (h *Handler) UpdateRoleHandler(c *gin.Context) {
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

	result, err := h.svc.UpdateRoleService(c.Request.Context(), slug, *updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// DeleteRoleHandler handles deleting a role.
//
// @Summary Delete a role by slug or ID
// @Description Delete a role by its slug or ID
// @Tags roles
// @Produce json
// @Param slug path string true "Role slug or ID"
// @Success 200 {object} structs.ReadRole "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/roles/{slug} [delete]
// @Security Bearer
func (h *Handler) DeleteRoleHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	result, err := h.svc.DeleteRoleService(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// ListRoleHandler handles listing all roles.
//
// @Summary List all roles
// @Description Retrieve a list of roles based on the provided query parameters
// @Tags roles
// @Produce json
// @Param params query structs.ListRoleParams true "List roles parameters"
// @Success 200 {array} structs.ReadRole "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/roles [get]
// @Security Bearer
func (h *Handler) ListRoleHandler(c *gin.Context) {
	params := &structs.ListRoleParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	roles, err := h.svc.ListRolesService(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, roles)
}

// ListRolePermissionHandler handles listing permissions for a role.
//
// @Summary List permissions for a role
// @Description Retrieve a list of permissions associated with a role by its ID
// @Tags roles
// @Produce json
// @Param slug path string true "Role ID"
// @Success 200 {array} structs.ReadPermission "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/roles/{slug}/permissions [get]
// @Security Bearer
func (h *Handler) ListRolePermissionHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.svc.GetRolePermissionsService(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// ListUserRoleHandler handles listing users for a role.
//
// @Summary List users for a role
// @Description Retrieve a list of users associated with a role by its ID
// @Tags roles
// @Produce json
// @Param slug path string true "Role ID"
// @Success 200 {array} structs.UserRole "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/roles/{slug}/users [get]
// @Security Bearer
func (h *Handler) ListUserRoleHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.svc.GetUsersByRoleIDService(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}
