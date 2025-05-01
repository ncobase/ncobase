package handler

import (
	"ncobase/domain/resource/service"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
)

// QuotaHandlerInterface defines the interface for quota handler operations
type QuotaHandlerInterface interface {
	GetQuota(c *gin.Context)
	SetQuota(c *gin.Context)
	GetUsage(c *gin.Context)
}

// quotaHandler handles quota operations
type quotaHandler struct {
	service service.QuotaServiceInterface
}

// NewQuotaHandler creates a new quota handler
func NewQuotaHandler(service service.QuotaServiceInterface) QuotaHandlerInterface {
	return &quotaHandler{
		service: service,
	}
}

// GetQuota handles retrieving the storage quota for a tenant
//
// @Summary Get storage quota
// @Description Get the storage quota for a tenant
// @Tags res
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Success 200 {object} map[string]int64 "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/quotas [get]
// @Security Bearer
func (h *quotaHandler) GetQuota(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("tenant_id")))
		return
	}

	quota, err := h.service.GetQuota(c.Request.Context(), tenantID)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error getting quota: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to get quota"))
		return
	}

	resp.Success(c.Writer, map[string]int64{
		"quota": quota,
	})
}

// SetQuota handles setting the storage quota for a tenant
//
// @Summary Set storage quota
// @Description Set the storage quota for a tenant
// @Tags res
// @Accept json
// @Produce json
// @Param body body map[string]interface{} true "Quota information"
// @Success 200 {object} map[string]int64 "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/quotas [put]
// @Security Bearer
func (h *quotaHandler) SetQuota(c *gin.Context) {
	var body struct {
		TenantID string `json:"tenant_id" binding:"required"`
		Quota    int64  `json:"quota" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Invalid request body"))
		return
	}

	if body.Quota <= 0 {
		resp.Fail(c.Writer, resp.BadRequest("Quota must be greater than zero"))
		return
	}

	err := h.service.SetQuota(c.Request.Context(), body.TenantID, body.Quota)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error setting quota: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to set quota"))
		return
	}

	resp.Success(c.Writer, map[string]int64{
		"quota": body.Quota,
	})
}

// GetUsage handles retrieving the current storage usage for a tenant
//
// @Summary Get storage usage
// @Description Get the current storage usage for a tenant
// @Tags res
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Success 200 {object} map[string]interface{} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/quotas/usage [get]
// @Security Bearer
func (h *quotaHandler) GetUsage(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("tenant_id")))
		return
	}

	usage, err := h.service.GetUsage(c.Request.Context(), tenantID)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error getting usage: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to get usage"))
		return
	}

	quota, err := h.service.GetQuota(c.Request.Context(), tenantID)
	if err != nil {
		logger.Warnf(c.Request.Context(), "Error getting quota: %v", err)
		// Continue with just usage if we can't get quota
	}

	// Calculate usage percentage
	var usagePercent float64 = 0
	if quota > 0 {
		usagePercent = float64(usage) / float64(quota) * 100
	}

	// Check if quota is exceeded
	isExceeded, err := h.service.IsQuotaExceeded(c.Request.Context(), tenantID)
	if err != nil {
		logger.Warnf(c.Request.Context(), "Error checking if quota is exceeded: %v", err)
		// Continue anyway
	}

	resp.Success(c.Writer, map[string]any{
		"usage":           usage,
		"quota":           quota,
		"usage_percent":   usagePercent,
		"quota_exceeded":  isExceeded,
		"formatted_usage": formatSize(usage),
		"formatted_quota": formatSize(quota),
	})
}
