package handler

import (
	"context"
	"stocms/internal/service"
	"stocms/pkg/log"
	"stocms/pkg/resp"

	"github.com/gin-gonic/gin"
)

// Handler represents a handler definition.
type Handler struct {
	svc *service.Service
}

// New creates a new Handler instance.
func New(svc *service.Service) *Handler {
	return &Handler{svc}
}

// Ping health status
func (h *Handler) Ping(ctx *gin.Context) {
	if err := h.svc.Ping(ctx); err != nil {
		log.Fatalf(context.Background(), "ping error: %+v", err)
	}
	resp.Success(ctx, nil)
}
