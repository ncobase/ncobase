package handler

import (
	"ncobase/plugin/resource/service"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"
)

// QuotaHandlerInterface defines quota handler methods
type QuotaHandlerInterface interface {
	GetMyQuota(c *gin.Context)
	GetMyUsage(c *gin.Context)
}

type quotaHandler struct {
	service service.QuotaServiceInterface
}

// NewQuotaHandler creates new quota handler
func NewQuotaHandler(service service.QuotaServiceInterface) QuotaHandlerInterface {
	return &quotaHandler{
		service: service,
	}
}

// GetMyQuota handles retrieving user's storage quota
//
// @Summary Get my storage quota
// @Description Get storage quota for current user
// @Tags Resource
// @Produce json
// @Success 200 {object} map[string]int64 "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/quota [get]
// @Security Bearer
func (h *quotaHandler) GetMyQuota(c *gin.Context) {
	// Get user ID from context (assuming it's set by middleware)
	userID := c.GetString("user_id")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("user_id")))
		return
	}

	quota, err := h.service.GetQuota(c.Request.Context(), userID)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error getting quota: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to get quota"))
		return
	}

	resp.Success(c.Writer, map[string]int64{
		"quota": quota,
	})
}

// GetMyUsage handles retrieving user's current storage usage
//
// @Summary Get my storage usage
// @Description Get current storage usage for current user
// @Tags Resource
// @Produce json
// @Success 200 {object} map[string]interface{} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/usage [get]
// @Security Bearer
func (h *quotaHandler) GetMyUsage(c *gin.Context) {
	// Get user ID from context (assuming it's set by middleware)
	userID := c.GetString("user_id")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("user_id")))
		return
	}

	usage, err := h.service.GetUsage(c.Request.Context(), userID)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error getting usage: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to get usage"))
		return
	}

	quota, err := h.service.GetQuota(c.Request.Context(), userID)
	if err != nil {
		logger.Warnf(c.Request.Context(), "Error getting quota: %v", err)
	}

	// Calculate usage percentage
	var usagePercent float64 = 0
	if quota > 0 {
		usagePercent = float64(usage) / float64(quota) * 100
	}

	// Check if quota exceeded
	isExceeded, err := h.service.IsQuotaExceeded(c.Request.Context(), userID)
	if err != nil {
		logger.Warnf(c.Request.Context(), "Error checking if quota is exceeded: %v", err)
	}

	resp.Success(c.Writer, types.JSON{
		"usage":           usage,
		"quota":           quota,
		"usage_percent":   usagePercent,
		"quota_exceeded":  isExceeded,
		"formatted_usage": formatSize(usage),
		"formatted_quota": formatSize(quota),
	})
}
