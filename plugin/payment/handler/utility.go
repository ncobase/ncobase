package handler

import (
	"ncobase/plugin/payment/service"
	"ncobase/plugin/payment/structs"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
)

// UtilityHandlerInterface defines the interface for utility handler operations
type UtilityHandlerInterface interface {
	ListProviders(c *gin.Context)
	GetStats(c *gin.Context)
}

// utilityHandler handles utility requests
type utilityHandler struct {
	svc *service.Service
}

// NewUtilityHandler creates a new utility handler
func NewUtilityHandler(svc *service.Service) UtilityHandlerInterface {
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
	providers := h.svc.Provider.GetAllProviders()
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
	endDate := parsePaymentStatsTime(c.Query("end_date"), time.Now())
	startDate := parsePaymentStatsTime(c.Query("start_date"), endDate.AddDate(0, 0, -30))
	if startDate.After(endDate) {
		resp.Fail(c.Writer, resp.BadRequest("start_date must be before end_date"))
		return
	}

	currency := c.DefaultQuery("currency", string(structs.CurrencyUSD))
	orderStats, err := h.svc.Order.GetOrderStats(
		c.Request.Context(),
		startDate.UnixMilli(),
		endDate.UnixMilli(),
		currency,
	)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to get payment order stats: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to get payment order stats", err))
		return
	}

	subscriptionStats, err := h.svc.Subscription.GetSubscriptionStats(c.Request.Context())
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to get payment subscription stats: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to get payment subscription stats", err))
		return
	}

	revenueByChannel, err := h.svc.Order.GetRevenueByChannel(
		c.Request.Context(),
		startDate.UnixMilli(),
		endDate.UnixMilli(),
		currency,
	)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to get payment channel revenue: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to get payment channel revenue", err))
		return
	}

	stats := map[string]any{
		"total_orders":         orderStats.TotalCount,
		"successful_payments":  orderStats.SuccessCount,
		"failed_payments":      orderStats.FailedCount,
		"total_refunds":        orderStats.RefundedCount,
		"total_revenue":        orderStats.SuccessAmount,
		"total_amount":         orderStats.TotalAmount,
		"active_subscriptions": subscriptionStats.ActiveCount,
		"revenue_by_channel":   revenueByChannel,
		"revenue_by_period":    []map[string]any{{"date": orderStats.PeriodEnd, "amount": orderStats.SuccessAmount}},
		"currency":             orderStats.Currency,
		"period_start":         orderStats.PeriodStart,
		"period_end":           orderStats.PeriodEnd,
		"subscription_summary": subscriptionStats,
		"registered_providers": h.svc.Provider.GetAllProviders(),
	}

	resp.Success(c.Writer, stats)
}

func parsePaymentStatsTime(value string, fallback time.Time) time.Time {
	if value == "" {
		return fallback
	}
	if timestamp, err := strconv.ParseInt(value, 10, 64); err == nil {
		if timestamp > 0 && timestamp < 1_000_000_000_000 {
			return time.Unix(timestamp, 0)
		}
		return time.UnixMilli(timestamp)
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed
	}
	return fallback
}
