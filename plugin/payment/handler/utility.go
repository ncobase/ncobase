package handler

import (
	"ncobase/payment/service"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/net/resp"
)

// UtilityHandlerInterface defines the interface for utility handler operations
type UtilityHandlerInterface interface {
	ListProviders(c *gin.Context)
	GetStats(c *gin.Context)
}

// utilityHandler handles utility requests
type utilityHandler struct {
	svc service.ProviderServiceInterface
}

// NewUtilityHandler creates a new utility handler
func NewUtilityHandler(svc service.ProviderServiceInterface) UtilityHandlerInterface {
	return &utilityHandler{svc: svc}
}

// ListProviders lists all available payment providers
//
// @Summary List payment providers
// @Description Get a list of all available payment providers
// @Tags payment
// @Produce json
// @Success 200 {object} resp.Exception{items=[]map[string]any} "success"
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/providers [get]
// @Security Bearer
func (h *utilityHandler) ListProviders(c *gin.Context) {
	providers := h.svc.GetAllProviders()
	resp.Success(c.Writer, providers)
}

// GetStats gets payment statistics
//
// @Summary Get payment statistics
// @Description Get payment statistics and metrics
// @Tags payment
// @Produce json
// @Success 200 {object} resp.Exception{data=map[string]any} "success"
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/stats [get]
// @Security Bearer
func (h *utilityHandler) GetStats(c *gin.Context) {
	// This would typically come from a stats service that aggregates payment data
	// For now, return a placeholder
	stats := map[string]any{
		"total_transactions":           0,
		"total_revenue":                0.0,
		"active_subscriptions":         0,
		"popular_products":             []map[string]any{},
		"payment_methods_distribution": map[string]any{},
	}

	resp.Success(c.Writer, stats)
}
