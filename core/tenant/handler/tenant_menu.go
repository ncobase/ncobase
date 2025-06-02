package handler

import (
	"ncobase/tenant/service"
	"ncobase/tenant/structs"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/net/resp"
)

// TenantMenuHandlerInterface represents the tenant menu handler interface.
type TenantMenuHandlerInterface interface {
	AddMenuToTenant(c *gin.Context)
	RemoveMenuFromTenant(c *gin.Context)
	GetTenantMenus(c *gin.Context)
	CheckMenuInTenant(c *gin.Context)
}

// tenantMenuHandler represents the tenant menu handler.
type tenantMenuHandler struct {
	s *service.Service
}

// NewTenantMenuHandler creates new tenant menu handler.
func NewTenantMenuHandler(svc *service.Service) TenantMenuHandlerInterface {
	return &tenantMenuHandler{
		s: svc,
	}
}

// AddMenuToTenant handles adding a menu to a tenant.
//
// @Summary Add menu to tenant
// @Description Add a menu to a tenant
// @Tags sys
// @Accept json
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Param body body structs.AddMenuToTenantRequest true "AddMenuToTenantRequest object"
// @Success 200 {object} structs.TenantMenu "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/{tenantId}/menus [post]
// @Security Bearer
func (h *tenantMenuHandler) AddMenuToTenant(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
		return
	}

	var req structs.AddMenuToTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	// Check if menu already in tenant
	exists, _ := h.s.TenantMenu.IsMenuInTenant(c.Request.Context(), tenantID, req.MenuID)
	if exists {
		resp.Fail(c.Writer, resp.BadRequest("Menu already belongs to this tenant"))
		return
	}

	result, err := h.s.TenantMenu.AddMenuToTenant(c.Request.Context(), tenantID, req.MenuID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// RemoveMenuFromTenant handles removing a menu from a tenant.
//
// @Summary Remove menu from tenant
// @Description Remove a menu from a tenant
// @Tags sys
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Param menuId path string true "Menu ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/{tenantId}/menus/{menuId} [delete]
// @Security Bearer
func (h *tenantMenuHandler) RemoveMenuFromTenant(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
		return
	}

	menuID := c.Param("menuId")
	if menuID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Menu ID is required"))
		return
	}

	err := h.s.TenantMenu.RemoveMenuFromTenant(c.Request.Context(), tenantID, menuID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]interface{}{
		"status":    "removed",
		"tenant_id": tenantID,
		"menu_id":   menuID,
	})
}

// GetTenantMenus handles getting all menus for a tenant.
//
// @Summary Get tenant menus
// @Description Get all menus for a tenant
// @Tags sys
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Success 200 {array} string "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/{tenantId}/menus [get]
// @Security Bearer
func (h *tenantMenuHandler) GetTenantMenus(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
		return
	}

	menuIDs, err := h.s.TenantMenu.GetTenantMenus(c.Request.Context(), tenantID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]interface{}{
		"tenant_id": tenantID,
		"menu_ids":  menuIDs,
		"count":     len(menuIDs),
	})
}

// CheckMenuInTenant handles checking if a menu belongs to a tenant.
//
// @Summary Check menu in tenant
// @Description Check if a menu belongs to a tenant
// @Tags sys
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Param menuId path string true "Menu ID"
// @Success 200 {object} map[string]bool "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/{tenantId}/menus/{menuId}/check [get]
// @Security Bearer
func (h *tenantMenuHandler) CheckMenuInTenant(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
		return
	}

	menuID := c.Param("menuId")
	if menuID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Menu ID is required"))
		return
	}

	exists, err := h.s.TenantMenu.IsMenuInTenant(c.Request.Context(), tenantID, menuID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]bool{"exists": exists})
}
