package handler

import (
	"ncobase/tenant/service"
	"ncobase/tenant/structs"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// TenantQuotaHandlerInterface defines the interface for tenant quota handler
type TenantQuotaHandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	UpdateUsage(c *gin.Context)
	CheckLimit(c *gin.Context)
	GetSummary(c *gin.Context)
}

// tenantQuotaHandler implements TenantQuotaHandlerInterface
type tenantQuotaHandler struct {
	s *service.Service
}

// NewTenantQuotaHandler creates a new tenant quota handler
func NewTenantQuotaHandler(svc *service.Service) TenantQuotaHandlerInterface {
	return &tenantQuotaHandler{s: svc}
}

// Create handles creating a tenant quota
//
// @Summary Create tenant quota
// @Description Create a new tenant quota configuration
// @Tags sys
// @Accept json
// @Produce json
// @Param body body structs.CreateTenantQuotaBody true "Quota configuration"
// @Success 200 {object} structs.ReadTenantQuota "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/quotas [post]
// @Security Bearer
func (h *tenantQuotaHandler) Create(c *gin.Context) {
	body := &structs.CreateTenantQuotaBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.TenantQuota.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Update handles updating a tenant quota
//
// @Summary Update tenant quota
// @Description Update an existing tenant quota configuration
// @Tags sys
// @Accept json
// @Produce json
// @Param id path string true "Quota ID"
// @Param body body types.JSON true "Update data"
// @Success 200 {object} structs.ReadTenantQuota "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/quotas/{id} [put]
// @Security Bearer
func (h *tenantQuotaHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
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

	result, err := h.s.TenantQuota.Update(c.Request.Context(), id, *updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get handles retrieving a tenant quota
//
// @Summary Get tenant quota
// @Description Retrieve a tenant quota by ID
// @Tags sys
// @Produce json
// @Param id path string true "Quota ID"
// @Success 200 {object} structs.ReadTenantQuota "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/quotas/{id} [get]
// @Security Bearer
func (h *tenantQuotaHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.s.TenantQuota.Get(c.Request.Context(), id)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete handles deleting a tenant quota
//
// @Summary Delete tenant quota
// @Description Delete a tenant quota configuration
// @Tags sys
// @Produce json
// @Param id path string true "Quota ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/quotas/{id} [delete]
// @Security Bearer
func (h *tenantQuotaHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.s.TenantQuota.Delete(c.Request.Context(), id); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List handles listing tenant quotas
//
// @Summary List tenant quotas
// @Description Retrieve a list of tenant quotas
// @Tags sys
// @Produce json
// @Param params query structs.ListTenantQuotaParams true "List parameters"
// @Success 200 {array} structs.ReadTenantQuota "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/quotas [get]
// @Security Bearer
func (h *tenantQuotaHandler) List(c *gin.Context) {
	params := &structs.ListTenantQuotaParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.TenantQuota.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// UpdateUsage handles updating quota usage
//
// @Summary Update quota usage
// @Description Update the current usage of a quota
// @Tags sys
// @Accept json
// @Produce json
// @Param body body structs.QuotaUsageRequest true "Usage update request"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/quotas/usage [post]
// @Security Bearer
func (h *tenantQuotaHandler) UpdateUsage(c *gin.Context) {
	body := &structs.QuotaUsageRequest{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	if err := h.s.TenantQuota.UpdateUsage(c.Request.Context(), body); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// CheckLimit handles checking quota limits
//
// @Summary Check quota limit
// @Description Check if tenant can use additional quota
// @Tags sys
// @Produce json
// @Param tenantId query string true "Tenant ID"
// @Param quota_type query string true "Quota Type"
// @Param amount query int true "Requested Amount"
// @Success 200 {object} map[string]bool "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/quotas/check [get]
// @Security Bearer
func (h *tenantQuotaHandler) CheckLimit(c *gin.Context) {
	tenantID := c.Query("tenantId")
	quotaType := c.Query("quota_type")
	amountStr := c.Query("amount")

	if tenantID == "" || quotaType == "" || amountStr == "" {
		resp.Fail(c.Writer, resp.BadRequest("Missing required parameters"))
		return
	}

	amount := int64(0)
	if val, err := convert.StringToInt64(amountStr); err == nil {
		amount = val
	} else {
		resp.Fail(c.Writer, resp.BadRequest("Invalid amount"))
		return
	}

	allowed, err := h.s.TenantQuota.CheckQuotaLimit(c.Request.Context(), tenantID, structs.QuotaType(quotaType), amount)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]bool{"allowed": allowed})
}

// GetSummary handles retrieving tenant quota summary
//
// @Summary Get tenant quota summary
// @Description Retrieve all quotas for a tenant
// @Tags sys
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Success 200 {array} structs.ReadTenantQuota "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/tenants/{tenantId}/quotas [get]
// @Security Bearer
func (h *tenantQuotaHandler) GetSummary(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("tenant_id")))
		return
	}

	result, err := h.s.TenantQuota.GetTenantQuotaSummary(c.Request.Context(), tenantID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}
