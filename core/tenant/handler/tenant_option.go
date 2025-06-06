package handler

import (
	"ncobase/tenant/service"
	"ncobase/tenant/structs"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/net/resp"
)

// TenantOptionHandlerInterface represents the tenant option handler interface.
type TenantOptionHandlerInterface interface {
	AddOptionsToTenant(c *gin.Context)
	RemoveOptionsFromTenant(c *gin.Context)
	GetTenantOption(c *gin.Context)
	CheckOptionsInTenant(c *gin.Context)
}

// tenantOptionHandler represents the tenant option handler.
type tenantOptionHandler struct {
	s *service.Service
}

// NewTenantOptionHandler creates new tenant option handler.
func NewTenantOptionHandler(svc *service.Service) TenantOptionHandlerInterface {
	return &tenantOptionHandler{
		s: svc,
	}
}

// AddOptionsToTenant handles adding options to a tenant.
//
// @Summary Add options to tenant
// @Description Add options to a tenant
// @Tags sys
// @Accept json
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Param body body structs.AddOptionsToTenantRequest true "AddOptionsToTenantRequest object"
// @Success 200 {object} structs.TenantOption "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/{tenantId}/options [post]
// @Security Bearer
func (h *tenantOptionHandler) AddOptionsToTenant(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
		return
	}

	var req structs.AddOptionsToTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	// Check if options already in tenant
	exists, _ := h.s.TenantOption.IsOptionsInTenant(c.Request.Context(), tenantID, req.OptionID)
	if exists {
		resp.Fail(c.Writer, resp.BadRequest("Options already belong to this tenant"))
		return
	}

	result, err := h.s.TenantOption.AddOptionsToTenant(c.Request.Context(), tenantID, req.OptionID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// RemoveOptionsFromTenant handles removing options from a tenant.
//
// @Summary Remove options from tenant
// @Description Remove options from a tenant
// @Tags sys
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Param optionsId path string true "Options ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/{tenantId}/options/{optionsId} [delete]
// @Security Bearer
func (h *tenantOptionHandler) RemoveOptionsFromTenant(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
		return
	}

	optionsID := c.Param("optionsId")
	if optionsID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Options ID is required"))
		return
	}

	err := h.s.TenantOption.RemoveOptionsFromTenant(c.Request.Context(), tenantID, optionsID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]interface{}{
		"status":    "removed",
		"tenant_id": tenantID,
		"option_id": optionsID,
	})
}

// GetTenantOption handles getting all options for a tenant.
//
// @Summary Get tenant option
// @Description Get all options for a tenant
// @Tags sys
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Success 200 {array} string "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/{tenantId}/options [get]
// @Security Bearer
func (h *tenantOptionHandler) GetTenantOption(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
		return
	}

	optionsIDs, err := h.s.TenantOption.GetTenantOption(c.Request.Context(), tenantID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]interface{}{
		"tenant_id":  tenantID,
		"option_ids": optionsIDs,
		"count":      len(optionsIDs),
	})
}

// CheckOptionsInTenant handles checking if options belong to a tenant.
//
// @Summary Check options in tenant
// @Description Check if options belong to a tenant
// @Tags sys
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Param optionsId path string true "Options ID"
// @Success 200 {object} map[string]bool "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/{tenantId}/options/{optionsId}/check [get]
// @Security Bearer
func (h *tenantOptionHandler) CheckOptionsInTenant(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
		return
	}

	optionsID := c.Param("optionsId")
	if optionsID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Options ID is required"))
		return
	}

	exists, err := h.s.TenantOption.IsOptionsInTenant(c.Request.Context(), tenantID, optionsID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]bool{"exists": exists})
}
