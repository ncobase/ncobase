package handler

import (
	"ncobase/tenant/service"
	"ncobase/tenant/structs"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/net/resp"
)

// TenantDictionaryHandlerInterface represents the tenant dictionary handler interface.
type TenantDictionaryHandlerInterface interface {
	AddDictionaryToTenant(c *gin.Context)
	RemoveDictionaryFromTenant(c *gin.Context)
	GetTenantDictionaries(c *gin.Context)
	CheckDictionaryInTenant(c *gin.Context)
}

// tenantDictionaryHandler represents the tenant dictionary handler.
type tenantDictionaryHandler struct {
	s *service.Service
}

// NewTenantDictionaryHandler creates new tenant dictionary handler.
func NewTenantDictionaryHandler(svc *service.Service) TenantDictionaryHandlerInterface {
	return &tenantDictionaryHandler{
		s: svc,
	}
}

// AddDictionaryToTenant handles adding a dictionary to a tenant.
//
// @Summary Add dictionary to tenant
// @Description Add a dictionary to a tenant
// @Tags sys
// @Accept json
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Param body body structs.AddDictionaryToTenantRequest true "AddDictionaryToTenantRequest object"
// @Success 200 {object} structs.TenantDictionary "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/{tenantId}/dictionaries [post]
// @Security Bearer
func (h *tenantDictionaryHandler) AddDictionaryToTenant(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
		return
	}

	var req structs.AddDictionaryToTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	// Check if dictionary already in tenant
	exists, _ := h.s.TenantDictionary.IsDictionaryInTenant(c.Request.Context(), tenantID, req.DictionaryID)
	if exists {
		resp.Fail(c.Writer, resp.BadRequest("Dictionary already belongs to this tenant"))
		return
	}

	result, err := h.s.TenantDictionary.AddDictionaryToTenant(c.Request.Context(), tenantID, req.DictionaryID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// RemoveDictionaryFromTenant handles removing a dictionary from a tenant.
//
// @Summary Remove dictionary from tenant
// @Description Remove a dictionary from a tenant
// @Tags sys
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Param dictionaryId path string true "Dictionary ID"
// @Success 200 {object} resp.Success "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/{tenantId}/dictionaries/{dictionaryId} [delete]
// @Security Bearer
func (h *tenantDictionaryHandler) RemoveDictionaryFromTenant(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
		return
	}

	dictionaryID := c.Param("dictionaryId")
	if dictionaryID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Dictionary ID is required"))
		return
	}

	err := h.s.TenantDictionary.RemoveDictionaryFromTenant(c.Request.Context(), tenantID, dictionaryID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]interface{}{
		"status":        "removed",
		"tenant_id":     tenantID,
		"dictionary_id": dictionaryID,
	})
}

// GetTenantDictionaries handles getting all dictionaries for a tenant.
//
// @Summary Get tenant dictionaries
// @Description Get all dictionaries for a tenant
// @Tags sys
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Success 200 {array} string "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/{tenantId}/dictionaries [get]
// @Security Bearer
func (h *tenantDictionaryHandler) GetTenantDictionaries(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
		return
	}

	dictionaryIDs, err := h.s.TenantDictionary.GetTenantDictionaries(c.Request.Context(), tenantID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]interface{}{
		"tenant_id":      tenantID,
		"dictionary_ids": dictionaryIDs,
		"count":          len(dictionaryIDs),
	})
}

// CheckDictionaryInTenant handles checking if a dictionary belongs to a tenant.
//
// @Summary Check dictionary in tenant
// @Description Check if a dictionary belongs to a tenant
// @Tags sys
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Param dictionaryId path string true "Dictionary ID"
// @Success 200 {object} map[string]bool "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/{tenantId}/dictionaries/{dictionaryId}/check [get]
// @Security Bearer
func (h *tenantDictionaryHandler) CheckDictionaryInTenant(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Tenant ID is required"))
		return
	}

	dictionaryID := c.Param("dictionaryId")
	if dictionaryID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Dictionary ID is required"))
		return
	}

	exists, err := h.s.TenantDictionary.IsDictionaryInTenant(c.Request.Context(), tenantID, dictionaryID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]bool{"exists": exists})
}
