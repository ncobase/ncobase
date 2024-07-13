package handler

import (
	"ncobase/feature/linker/service"
)

// Handler represents the linker handler.
type Handler struct {
	s *service.Service
}

// New creates a new handler.
func New(svc *service.Service) *Handler {
	return &Handler{
		s: svc,
	}
}
