package handler

import "ncobase/feature/user/service"

// Handler represents the user handler.
type Handler struct {
	User UserHandlerInterface
}

// New creates a new handler.
func New(svc *service.Service) *Handler {
	return &Handler{
		User: NewUserHandler(svc),
	}
}
