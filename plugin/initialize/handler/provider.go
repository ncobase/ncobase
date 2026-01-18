package handler

import "ncobase/plugin/initialize/service"

// Handler represents the initialize handler.
type Handler struct {
	Initialize InitializeHandlerInterface
}

// New creates new initialize handler.
func New(svc *service.Service) *Handler {
	return &Handler{
		Initialize: NewInitializeHandler(svc),
	}
}
