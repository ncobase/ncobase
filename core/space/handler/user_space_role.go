package handler

import (
	"ncobase/space/service"
	"ncobase/space/structs"

	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// UserSpaceRoleHandlerInterface represents the user space role handler interface.
type UserSpaceRoleHandlerInterface interface {
	AddUserToSpaceRole(c *gin.Context)
	RemoveUserFromSpaceRole(c *gin.Context)
	GetUserSpaceRoles(c *gin.Context)
	GetSpaceUsersByRole(c *gin.Context)
	CheckUserSpaceRole(c *gin.Context)
	ListSpaceUsers(c *gin.Context)
	UpdateUserSpaceRole(c *gin.Context)
	BulkUpdateUserSpaceRoles(c *gin.Context)
}

// userSpaceRoleHandler represents the user space role handler.
type userSpaceRoleHandler struct {
	s *service.Service
}

// NewUserSpaceRoleHandler creates new user space role handler.
func NewUserSpaceRoleHandler(svc *service.Service) UserSpaceRoleHandlerInterface {
	return &userSpaceRoleHandler{
		s: svc,
	}
}

// AddUserToSpaceRole handles adding a user to a space
//
// @Summary Add user to space role
// @Description Add a user to a space
// @Tags sys
// @Accept json
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param body body structs.AddUserToSpaceRoleRequest true "AddUserToSpaceRoleRequest object"
// @Success 200 {object} structs.UserSpaceRoleResponse "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/users/roles [post]
// @Security Bearer
func (h *userSpaceRoleHandler) AddUserToSpaceRole(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	var req structs.AddUserToSpaceRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	// Validate required fields
	if req.UserID == "" {
		resp.Fail(c.Writer, resp.BadRequest("User ID is required"))
		return
	}
	if req.RoleID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Role ID is required"))
		return
	}

	// Check if user already has this role in space
	hasRole, _ := h.s.UserSpaceRole.IsUserInRoleInSpace(c.Request.Context(), req.UserID, spaceID, req.RoleID)
	if hasRole {
		resp.Fail(c.Writer, resp.BadRequest("User already has this role in the space"))
		return
	}

	// Add the user to space role
	result, err := h.s.UserSpaceRole.AddRoleToUserInSpace(c.Request.Context(), req.UserID, spaceID, req.RoleID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	response := &structs.UserSpaceRoleResponse{
		UserID:  result.UserID,
		SpaceID: result.SpaceID,
		RoleID:  result.RoleID,
		Status:  "added",
	}

	resp.Success(c.Writer, response)
}

// RemoveUserFromSpaceRole handles removing a user from a space role.
//
// @Summary Remove user from space role
// @Description Remove a user from a specific role in a space
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param userId path string true "User ID"
// @Param roleId path string true "Role ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/users/{userId}/roles/{roleId} [delete]
// @Security Bearer
func (h *userSpaceRoleHandler) RemoveUserFromSpaceRole(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	userID := c.Param("userId")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest("User ID is required"))
		return
	}

	roleID := c.Param("roleId")
	if roleID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Role ID is required"))
		return
	}

	err := h.s.UserSpaceRole.RemoveRoleFromUserInSpace(c.Request.Context(), userID, spaceID, roleID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]any{
		"status":   "removed",
		"user_id":  userID,
		"space_id": spaceID,
		"role_id":  roleID,
	})
}

// GetUserSpaceRoles handles getting all roles a user has in a space.
//
// @Summary Get user space roles
// @Description Get all roles a user has in a specific space
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param userId path string true "User ID or username"
// @Success 200 {object} structs.UserSpaceRolesResponse "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/users/{userId}/roles [get]
// @Security Bearer
func (h *userSpaceRoleHandler) GetUserSpaceRoles(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	userID := c.Param("userId")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest("User ID is required"))
		return
	}

	roleIDs, err := h.s.UserSpaceRole.GetUserRolesInSpace(c.Request.Context(), userID, spaceID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	response := &structs.UserSpaceRolesResponse{
		UserID:  userID,
		SpaceID: spaceID,
		RoleIDs: roleIDs,
		Count:   len(roleIDs),
	}

	resp.Success(c.Writer, response)
}

// GetSpaceUsersByRole handles getting all users with a specific role
//
// @Summary Get space users by role
// @Description Get all users that have a specific role in a space
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param roleId path string true "Role ID"
// @Success 200 {object} structs.SpaceRoleUsersResponse "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/roles/{roleId}/users [get]
// @Security Bearer
func (h *userSpaceRoleHandler) GetSpaceUsersByRole(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	roleID := c.Param("roleId")
	if roleID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Role ID is required"))
		return
	}

	userIDs, err := h.s.UserSpaceRole.GetSpaceUsersByRole(c.Request.Context(), spaceID, roleID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	response := &structs.SpaceRoleUsersResponse{
		SpaceID: spaceID,
		RoleID:  roleID,
		UserIDs: userIDs,
		Count:   len(userIDs),
	}

	resp.Success(c.Writer, response)
}

// CheckUserSpaceRole handles checking if a user has a specific role in a space.
//
// @Summary Check user space role
// @Description Check if a user has a specific role in a space
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param userId path string true "User ID or username"
// @Param roleId path string true "Role ID"
// @Success 200 {object} map[string]bool "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/users/{userId}/roles/{roleId}/check [get]
// @Security Bearer
func (h *userSpaceRoleHandler) CheckUserSpaceRole(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	userID := c.Param("userId")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest("User ID is required"))
		return
	}

	roleID := c.Param("roleId")
	if roleID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Role ID is required"))
		return
	}

	hasRole, err := h.s.UserSpaceRole.IsUserInRoleInSpace(c.Request.Context(), userID, spaceID, roleID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]bool{"hasRole": hasRole})
}

// ListSpaceUsers handles listing all users in a space with their roles.
//
// @Summary List space users
// @Description List all users in a space with their roles
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param params query structs.ListSpaceUsersParams true "List parameters"
// @Success 200 {object} structs.SpaceUsersListResponse "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/users [get]
// @Security Bearer
func (h *userSpaceRoleHandler) ListSpaceUsers(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	params := &structs.ListSpaceUsersParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.UserSpaceRole.ListSpaceUsers(c.Request.Context(), spaceID, params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// UpdateUserSpaceRole handles updating a user's role in a space.
//
// @Summary Update user space role
// @Description Update a user's role in a space
// @Tags sys
// @Accept json
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param userId path string true "User ID"
// @Param body body structs.UpdateUserSpaceRoleRequest true "UpdateUserSpaceRoleRequest object"
// @Success 200 {object} structs.UserSpaceRoleResponse "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/users/{userId}/roles [put]
// @Security Bearer
func (h *userSpaceRoleHandler) UpdateUserSpaceRole(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	userID := c.Param("userId")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest("User ID is required"))
		return
	}

	var req structs.UpdateUserSpaceRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.s.UserSpaceRole.UpdateUserSpaceRole(c.Request.Context(), userID, spaceID, &req)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// BulkUpdateUserSpaceRoles handles bulk updating user space roles.
//
// @Summary Bulk update user space roles
// @Description Bulk update multiple user space roles
// @Tags sys
// @Accept json
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param body body structs.BulkUpdateUserSpaceRolesRequest true "BulkUpdateUserSpaceRolesRequest object"
// @Success 200 {object} structs.BulkUpdateResponse "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/users/roles/bulk [put]
// @Security Bearer
func (h *userSpaceRoleHandler) BulkUpdateUserSpaceRoles(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	var req structs.BulkUpdateUserSpaceRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.s.UserSpaceRole.BulkUpdateUserSpaceRoles(c.Request.Context(), spaceID, &req)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}
