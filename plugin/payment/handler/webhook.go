package handler

import (
	"io"
	"ncobase/payment/service"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
)

// WebhookHandlerInterface defines the interface for webhook handler operations
type WebhookHandlerInterface interface {
	ProcessWebhook(c *gin.Context)
}

// webhookHandler handles webhook requests from payment providers
type webhookHandler struct {
	svc service.OrderServiceInterface
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(svc service.OrderServiceInterface) WebhookHandlerInterface {
	return &webhookHandler{svc: svc}
}

// ProcessWebhook handles processing webhooks from payment providers
//
// @Summary Process payment webhook
// @Description Process webhook notifications from payment providers
// @Tags payment
// @Accept json
// @Produce json
// @Param channel path string true "Payment channel ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/webhooks/{channel} [post]
func (h *webhookHandler) ProcessWebhook(c *gin.Context) {
	channelID := c.Param("channel")
	if channelID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Channel ID is required"))
		return
	}

	// Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to read webhook body: %v", err)
		resp.Fail(c.Writer, resp.BadRequest("Failed to read webhook body", err))
		return
	}

	// Extract headers
	headers := make(map[string]string)
	for k, v := range c.Request.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	// Process webhook
	if err := h.svc.ProcessWebhook(c.Request.Context(), channelID, body, headers); err != nil {
		logger.Errorf(c.Request.Context(), "Failed to process webhook: %v", err)

		// Note: Many payment providers expect a 200 response even if we encounter errors,
		// so they don't retry unnecessarily. Log the error but still return success.
		resp.Success(c.Writer, gin.H{
			"message": "Webhook received (with processing error)",
		})
		return
	}

	resp.Success(c.Writer, gin.H{
		"message": "Webhook processed successfully",
	})
}
