package handler

import (
	"ncobase/tenant/service"
	"ncobase/tenant/structs"

	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// TenantGroupHandlerInterface represents the tenant group handler interface.
type TenantGroupHandlerInterface interface {
	AddGroupToTenant(c *gin.Context)
	RemoveGroupFromTenant(c *gin.Context)
	GetTenantGroups(c *gin.Context)
	GetGroupTenants(c *gin.Context)
	IsGroupInTenant(c *gin.Context)
}

// tenantGroupHandler represents the tenant group handler.
type tenantGroupHandler struct {
	s *service.Service
}

// NewTenantGroupHandler creates new tenant group handler.
func NewTenantGroupHandler(svc *service.Service) TenantGroupHandlerInterface {
	return &tenantGroupHandler{
		s: svc,
	}
}

// AddGroupToTenant handles adding a group to a tenant.
//
// @Summary Add group to tenant
// @Description Add a group to a specific tenant
// @Tags iam
// @Accept json
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Param body body structs.AddTenantGroupRequest true "AddTenantGroupRequest object"
// @Success 200 {object} structs.TenantGroupRelation "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/{tenantId}/groups [post]
// @Security Bearer
func (h *tenantGroupHandler) AddGroupToTenant(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
		return
	}

	var req structs.AddTenantGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	// Check if group already exists in tenant
	exists, _ := h.s.TenantGroup.IsGroupInTenant(c.Request.Context(), tenantID, req.GroupID)
	if exists {
		resp.Fail(c.Writer, resp.BadRequest("Group already exists in this tenant"))
		return
	}

	// Add the group to tenant
	relation, err := h.s.TenantGroup.AddGroupToTenant(c.Request.Context(), tenantID, req.GroupID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, relation)
}

// RemoveGroupFromTenant handles removing a group from a tenant.
//
// @Summary Remove group from tenant
// @Description Remove a group from a specific tenant
// @Tags iam
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Param groupId path string true "Group ID"
// @Success 200 {object} resp.Success "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/{tenantId}/groups/{groupId} [delete]
// @Security Bearer
func (h *tenantGroupHandler) RemoveGroupFromTenant(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
		return
	}

	groupID := c.Param("groupId")
	if groupID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Group ID is required"))
		return
	}

	err := h.s.TenantGroup.RemoveGroupFromTenant(c.Request.Context(), tenantID, groupID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]bool{"success": true})
}

// GetTenantGroups handles getting all groups for a tenant.
//
// @Summary Get tenant groups
// @Description Get all groups belonging to a specific tenant
// @Tags iam
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Param params query structs.ListGroupParams true "List group parameters"
// @Success 200 {array} structs.ReadGroup "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/{tenantId}/groups [get]
// @Security Bearer
func (h *tenantGroupHandler) GetTenantGroups(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
		return
	}

	params := &structs.ListGroupParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.TenantGroup.GetTenantGroups(c.Request.Context(), tenantID, params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetGroupTenants handles getting all tenants that have a specific group.
//
// @Summary Get group tenants
// @Description Get all tenants that have a specific group
// @Tags iam
// @Produce json
// @Param groupId path string true "Group ID"
// @Success 200 {array} string "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/groups/{groupId}/tenants [get]
// @Security Bearer
func (h *tenantGroupHandler) GetGroupTenants(c *gin.Context) {
	groupID := c.Param("groupId")
	if groupID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Group ID is required"))
		return
	}

	tenants, err := h.s.TenantGroup.GetGroupTenants(c.Request.Context(), groupID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, tenants)
}

// IsGroupInTenant handles checking if a group belongs to a tenant.
//
// @Summary Check if group is in tenant
// @Description Check if a group belongs to a specific tenant
// @Tags iam
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Param groupId path string true "Group ID"
// @Success 200 {object} map[string]bool "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/tenants/{tenantId}/groups/{groupId}/check [get]
// @Security Bearer
func (h *tenantGroupHandler) IsGroupInTenant(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
		return
	}

	groupID := c.Param("groupId")
	if groupID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Group ID is required"))
		return
	}

	exists, err := h.s.TenantGroup.IsGroupInTenant(c.Request.Context(), tenantID, groupID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]bool{"exists": exists})
}
