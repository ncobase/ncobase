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

// RoleHandlerInterface is the interface for the handler.
type RoleHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	GetBySlug(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	BulkUpdate(c *gin.Context)
	BulkDelete(c *gin.Context)
	ListUsers(c *gin.Context)
	AssignUsers(c *gin.Context)
	RemoveUsers(c *gin.Context)
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

type roleBulkUpdateBody struct {
	Updates []types.JSON `json:"updates"`
}

type roleBatchIDsBody struct {
	IDs     []string `json:"ids"`
	UserIDs []string `json:"user_ids"`
	UserIds []string `json:"userIds"`
}

func (b *roleBatchIDsBody) Users() []string {
	if len(b.UserIDs) > 0 {
		return b.UserIDs
	}
	return b.UserIds
}

// Create handles the creation of a new role.
//
// @Summary Create a new role
// @Description Create a new role with the provided data
// @Tags sys
// @Accept json
// @Produce json
// @Param body body structs.CreateRoleBody true "Role data"
// @Success 200 {object} structs.ReadRole "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/roles [post]
// @Security Bearer
func (h *roleHandler) Create(c *gin.Context) {
	body := &structs.CreateRoleBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
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
// @Tags sys
// @Produce json
// @Param slug path string true "Role slug or ID"
// @Success 200 {object} structs.ReadRole "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/roles/{slug} [get]
// @Security Bearer
func (h *roleHandler) Get(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	result, err := h.s.Role.Find(c.Request.Context(), &structs.FindRole{Slug: slug})
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetBySlug handles retrieving a role by slug.
func (h *roleHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}

	result, err := h.s.Role.GetBySlug(c.Request.Context(), slug)
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
// @Tags sys
// @Accept json
// @Produce json
// @Param slug path string true "Role slug or ID"
// @Param body body types.JSON true "Role data"
// @Success 200 {object} structs.ReadRole "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/roles/{slug} [put]
// @Security Bearer
func (h *roleHandler) Update(c *gin.Context) {
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
// @Tags sys
// @Produce json
// @Param slug path string true "Role slug or ID"
// @Success 200 {object} structs.ReadRole "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/roles/{slug} [delete]
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
// @Tags sys
// @Produce json
// @Param params query structs.ListRoleParams true "List roles parameters"
// @Success 200 {array} structs.ReadRole "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/roles [get]
// @Security Bearer
func (h *roleHandler) List(c *gin.Context) {
	params := &structs.ListRoleParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
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

// BulkUpdate handles updating multiple roles.
func (h *roleHandler) BulkUpdate(c *gin.Context) {
	body := &roleBulkUpdateBody{}
	if err := c.ShouldBindJSON(body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	if len(body.Updates) == 0 {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("updates")))
		return
	}

	results := make([]*structs.ReadRole, 0, len(body.Updates))
	for _, update := range body.Updates {
		roleID, _ := update["id"].(string)
		if roleID == "" {
			roleID, _ = update["slug"].(string)
		}
		if roleID == "" {
			resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
			return
		}

		result, err := h.s.Role.Update(c.Request.Context(), roleID, update)
		if err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			return
		}
		results = append(results, result)
	}

	resp.Success(c.Writer, results)
}

// BulkDelete handles deleting multiple roles.
func (h *roleHandler) BulkDelete(c *gin.Context) {
	body := &roleBatchIDsBody{}
	if err := c.ShouldBindJSON(body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	if len(body.IDs) == 0 {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("ids")))
		return
	}

	for _, id := range body.IDs {
		if id == "" {
			continue
		}
		if err := h.s.Role.Delete(c.Request.Context(), id); err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			return
		}
	}

	resp.Success(c.Writer, gin.H{"ids": body.IDs})
}

// ListUsers handles listing user IDs assigned to a role.
func (h *roleHandler) ListUsers(c *gin.Context) {
	roleID := c.Param("slug")
	if roleID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("role")))
		return
	}

	result, err := h.s.UserRole.GetUsersByRoleID(c.Request.Context(), roleID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// AssignUsers handles assigning users to a role.
func (h *roleHandler) AssignUsers(c *gin.Context) {
	roleID := c.Param("slug")
	if roleID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("role")))
		return
	}

	body := &roleBatchIDsBody{}
	if err := c.ShouldBindJSON(body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	userIDs := body.Users()
	if len(userIDs) == 0 {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("user_ids")))
		return
	}

	assigned := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		if userID == "" {
			continue
		}
		if err := h.s.UserRole.AddRoleToUser(c.Request.Context(), userID, roleID); err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			return
		}
		assigned = append(assigned, userID)
	}

	resp.Success(c.Writer, gin.H{"role_id": roleID, "user_ids": assigned})
}

// RemoveUsers handles removing users from a role.
func (h *roleHandler) RemoveUsers(c *gin.Context) {
	roleID := c.Param("slug")
	if roleID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("role")))
		return
	}

	body := &roleBatchIDsBody{}
	if err := c.ShouldBindJSON(body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	userIDs := body.Users()
	if len(userIDs) == 0 {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("user_ids")))
		return
	}

	removed := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		if userID == "" {
			continue
		}
		if err := h.s.UserRole.RemoveRoleFromUser(c.Request.Context(), userID, roleID); err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			return
		}
		removed = append(removed, userID)
	}

	resp.Success(c.Writer, gin.H{"role_id": roleID, "user_ids": removed})
}

// // ListUserRoleHandler handles listing users for a role.
// //
// // @Summary List users for a role
// // @Description Retrieve a list of users associated with a role by its ID
// // @Tags sys
// // @Produce json
// // @Param slug path string true "Role ID"
// // @Success 200 {array} structs.UserRole "success"
// // @Failure 400 {object} resp.Exception "bad request"
// // @Router /sys/roles/{slug}/users [get]
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
