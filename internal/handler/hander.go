package handler

import (
	"stocms/internal/service"
)

// Handler represents a controller definition.
type Handler struct {
	svc *service.Service
}

// New creates a new Handler instance.
func New(svc *service.Service) *Handler {
	return &Handler{svc}
}
