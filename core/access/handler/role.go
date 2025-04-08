package handler

import (
	"ncobase/core/access/service"
	"ncobase/core/access/structs"
	"ncore/pkg/ecode"
	"ncore/pkg/helper"
	"ncore/pkg/resp"
	"ncore/pkg/types"

	"github.com/gin-gonic/gin"
)

// RoleHandlerInterface is the interface for the handler.
type RoleHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
}

// roleHandler represents the handler.
type roleHandler struct {
	s *service.Service
}

// NewRoleHandler creates a new handler.
func NewRoleHandler(svc *service.Service) RoleHandlerInterface {
	return &roleHandler{
		s: svc,
	}
}

// Create handles the creation of a new role.
//
// @Summary Create a new role
// @Description Create a new role with the provided data
// @Tags iam
// @Accept json
// @Produce json
// @Param body body structs.CreateRoleBody true "Role data"
// @Success 200 {object} structs.ReadRole "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/roles [post]
// @Security Bearer
func (h *roleHandler) Create(c *gin.Context) {
	body := &structs.CreateRoleBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Role.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get handles retrieving a role by slug.
//
// @Summary Get a role by slug or ID
// @Description Retrieve a role by its slug or ID
// @Tags iam
// @Produce json
// @Param slug path string true "Role slug or ID"
// @Success 200 {object} structs.ReadRole "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/roles/{slug} [get]
// @Security Bearer
func (h *roleHandler) Get(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	result, err := h.s.Role.GetByID(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Update handles updating an existing role.
//
// @Summary Update an existing role
// @Description Update an existing role with the provided data
// @Tags iam
// @Accept json
// @Produce json
// @Param slug path string true "Role slug or ID"
// @Param body body types.JSON true "Role data"
// @Success 200 {object} structs.ReadRole "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/roles/{slug} [put]
// @Security Bearer
func (h *roleHandler) Update(c *gin.Context) {
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

	result, err := h.s.Role.Update(c.Request.Context(), slug, *updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete handles deleting a role.
//
// @Summary Delete a role by slug or ID
// @Description Delete a role by its slug or ID
// @Tags iam
// @Produce json
// @Param slug path string true "Role slug or ID"
// @Success 200 {object} structs.ReadRole "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/roles/{slug} [delete]
// @Security Bearer
func (h *roleHandler) Delete(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	if err := h.s.Role.Delete(c.Request.Context(), slug); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List handles listing all roles.
//
// @Summary List all roles
// @Description Retrieve a list of roles based on the provided query parameters
// @Tags iam
// @Produce json
// @Param params query structs.ListRoleParams true "List roles parameters"
// @Success 200 {array} structs.ReadRole "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/roles [get]
// @Security Bearer
func (h *roleHandler) List(c *gin.Context) {
	params := &structs.ListRoleParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	roles, err := h.s.Role.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, roles)
}

// // ListUserRoleHandler handles listing users for a role.
// //
// // @Summary List users for a role
// // @Description Retrieve a list of users associated with a role by its ID
// // @Tags iam
// // @Produce json
// // @Param slug path string true "Role ID"
// // @Success 200 {array} structs.UserRole "success"
// // @Failure 400 {object} resp.Exception "bad request"
// // @Router /iam/roles/{slug}/users [get]
// // @Security Bearer
// func (h *roleHandler) ListUser(c *gin.Context) {
// 	slug := c.Param("slug")
// 	if slug == "" {
// 		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
// 		return
// 	}
//
// 	result, err := h.s.RoleService.GetUsersByRoleID(c.Request.Context(), slug)
// 	if err != nil {
// 		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
// 		return
// 	}
//
// 	resp.Success(c.Writer, result)
// }
