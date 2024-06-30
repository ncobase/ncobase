package handler

import (
	"context"
	"ncobase/app/service"

	"ncobase/common/log"
	"ncobase/common/oauth"
	"ncobase/common/resp"

	"github.com/gin-gonic/gin"
)

// Handler represents a handler definition.
type Handler struct {
	OAuthConfig map[string]oauth.ProviderConfig
	svc         *service.Service
}

// New creates a new Handler instance.
func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

// HealthHandler handles health status checks.
//
// @Summary Health status
// @Description Check the health status of the service.
// @Tags root
// @Produce json
// @Success 200 {object} resp.Exception "success"
// @Router /health [get]
func (h *Handler) HealthHandler(c *gin.Context) {
	if err := h.svc.Ping(c); err != nil {
		log.Fatalf(context.Background(), "ping error: %+v", err)
	}
	resp.Success(c.Writer, nil)
}
