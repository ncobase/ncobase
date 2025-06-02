package handler

import (
	"ncobase/tenant/service"
	"ncobase/tenant/structs"

	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// UserTenantRoleHandlerInterface represents the user tenant role handler interface.
type UserTenantRoleHandlerInterface interface {
	AddUserToTenantRole(c *gin.Context)
	RemoveUserFromTenantRole(c *gin.Context)
	GetUserTenantRoles(c *gin.Context)
	GetTenantUsersByRole(c *gin.Context)
	CheckUserTenantRole(c *gin.Context)
	ListTenantUsers(c *gin.Context)
	UpdateUserTenantRole(c *gin.Context)
	BulkUpdateUserTenantRoles(c *gin.Context)
}

// userTenantRoleHandler represents the user tenant role handler.
type userTenantRoleHandler struct {
	s *service.Service
}

// NewUserTenantRoleHandler creates new user tenant role handler.
func NewUserTenantRoleHandler(svc *service.Service) UserTenantRoleHandlerInterface {
	return &userTenantRoleHandler{
		s: svc,
	}
}

// AddUserToTenantRole handles adding a user to a tenant
//
// @Summary Add user to tenant role
// @Description Add a user to a tenant
// @Tags sys
// @Accept json
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Param body body structs.AddUserToTenantRoleRequest true "AddUserToTenantRoleRequest object"
// @Success 200 {object} structs.UserTenantRoleResponse "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/{tenantId}/users/roles [post]
// @Security Bearer
func (h *userTenantRoleHandler) AddUserToTenantRole(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
		return
	}

	var req structs.AddUserToTenantRoleRequest
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

	// Check if user already has this role in tenant
	hasRole, _ := h.s.UserTenantRole.IsUserInRoleInTenant(c.Request.Context(), req.UserID, tenantID, req.RoleID)
	if hasRole {
		resp.Fail(c.Writer, resp.BadRequest("User already has this role in the tenant"))
		return
	}

	// Add the user to tenant role
	result, err := h.s.UserTenantRole.AddRoleToUserInTenant(c.Request.Context(), req.UserID, tenantID, req.RoleID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	response := &structs.UserTenantRoleResponse{
		UserID:   result.UserID,
		TenantID: result.TenantID,
		RoleID:   result.RoleID,
		Status:   "added",
	}

	resp.Success(c.Writer, response)
}

// RemoveUserFromTenantRole handles removing a user from a tenant role.
//
// @Summary Remove user from tenant role
// @Description Remove a user from a specific role in a tenant
// @Tags sys
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Param userId path string true "User ID"
// @Param roleId path string true "Role ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/{tenantId}/users/{userId}/roles/{roleId} [delete]
// @Security Bearer
func (h *userTenantRoleHandler) RemoveUserFromTenantRole(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
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

	err := h.s.UserTenantRole.RemoveRoleFromUserInTenant(c.Request.Context(), userID, tenantID, roleID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]interface{}{
		"status":    "removed",
		"user_id":   userID,
		"tenant_id": tenantID,
		"role_id":   roleID,
	})
}

// GetUserTenantRoles handles getting all roles a user has in a tenant.
//
// @Summary Get user tenant roles
// @Description Get all roles a user has in a specific tenant
// @Tags sys
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Param username path string true "User ID or username"
// @Success 200 {object} structs.UserTenantRolesResponse "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/{tenantId}/users/{username}/roles [get]
// @Security Bearer
func (h *userTenantRoleHandler) GetUserTenantRoles(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
		return
	}

	userID := c.Param("username")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest("User ID is required"))
		return
	}

	roleIDs, err := h.s.UserTenantRole.GetUserRolesInTenant(c.Request.Context(), userID, tenantID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	response := &structs.UserTenantRolesResponse{
		UserID:   userID,
		TenantID: tenantID,
		RoleIDs:  roleIDs,
		Count:    len(roleIDs),
	}

	resp.Success(c.Writer, response)
}

// GetTenantUsersByRole handles getting all users with a specific role
//
// @Summary Get tenant users by role
// @Description Get all users that have a specific role in a tenant
// @Tags sys
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Param roleId path string true "Role ID"
// @Success 200 {object} structs.TenantRoleUsersResponse "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/{tenantId}/roles/{roleId}/users [get]
// @Security Bearer
func (h *userTenantRoleHandler) GetTenantUsersByRole(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
		return
	}

	roleID := c.Param("roleId")
	if roleID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Role ID is required"))
		return
	}

	userIDs, err := h.s.UserTenantRole.GetTenantUsersByRole(c.Request.Context(), tenantID, roleID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	response := &structs.TenantRoleUsersResponse{
		TenantID: tenantID,
		RoleID:   roleID,
		UserIDs:  userIDs,
		Count:    len(userIDs),
	}

	resp.Success(c.Writer, response)
}

// CheckUserTenantRole handles checking if a user has a specific role in a tenant.
//
// @Summary Check user tenant role
// @Description Check if a user has a specific role in a tenant
// @Tags sys
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Param username path string true "User ID or username"
// @Param roleId path string true "Role ID"
// @Success 200 {object} map[string]bool "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/{tenantId}/users/{username}/roles/{roleId}/check [get]
// @Security Bearer
func (h *userTenantRoleHandler) CheckUserTenantRole(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
		return
	}

	userID := c.Param("username")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest("User ID is required"))
		return
	}

	roleID := c.Param("roleId")
	if roleID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Role ID is required"))
		return
	}

	hasRole, err := h.s.UserTenantRole.IsUserInRoleInTenant(c.Request.Context(), userID, tenantID, roleID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]bool{"hasRole": hasRole})
}

// ListTenantUsers handles listing all users in a tenant with their roles.
//
// @Summary List tenant users
// @Description List all users in a tenant with their roles
// @Tags sys
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Param params query structs.ListTenantUsersParams true "List parameters"
// @Success 200 {object} structs.TenantUsersListResponse "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/{tenantId}/users [get]
// @Security Bearer
func (h *userTenantRoleHandler) ListTenantUsers(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
		return
	}

	params := &structs.ListTenantUsersParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.UserTenantRole.ListTenantUsers(c.Request.Context(), tenantID, params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// UpdateUserTenantRole handles updating a user's role in a tenant.
//
// @Summary Update user tenant role
// @Description Update a user's role in a tenant
// @Tags sys
// @Accept json
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Param userId path string true "User ID"
// @Param body body structs.UpdateUserTenantRoleRequest true "UpdateUserTenantRoleRequest object"
// @Success 200 {object} structs.UserTenantRoleResponse "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/{tenantId}/users/{userId}/roles [put]
// @Security Bearer
func (h *userTenantRoleHandler) UpdateUserTenantRole(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
		return
	}

	userID := c.Param("userId")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest("User ID is required"))
		return
	}

	var req structs.UpdateUserTenantRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.s.UserTenantRole.UpdateUserTenantRole(c.Request.Context(), userID, tenantID, &req)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// BulkUpdateUserTenantRoles handles bulk updating user tenant roles.
//
// @Summary Bulk update user tenant roles
// @Description Bulk update multiple user tenant roles
// @Tags sys
// @Accept json
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Param body body structs.BulkUpdateUserTenantRolesRequest true "BulkUpdateUserTenantRolesRequest object"
// @Success 200 {object} structs.BulkUpdateResponse "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/{tenantId}/users/roles/bulk [put]
// @Security Bearer
func (h *userTenantRoleHandler) BulkUpdateUserTenantRoles(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
		return
	}

	var req structs.BulkUpdateUserTenantRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.s.UserTenantRole.BulkUpdateUserTenantRoles(c.Request.Context(), tenantID, &req)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}
