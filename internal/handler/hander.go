package handler

import (
	"stocms/internal/service"
	"stocms/pkg/log"
	"stocms/pkg/oauth"
	"stocms/pkg/resp"

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

// HealthHandler health status
func (h *Handler) HealthHandler(c *gin.Context) {
	if err := h.svc.Ping(c); err != nil {
		log.Fatalf(nil, "ping error: %+v", err)
	}
	resp.Success(c.Writer, nil)
}
