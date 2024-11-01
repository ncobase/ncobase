package handler

import "ncobase/core/space/service"

// Handler represents the group handler.
type Handler struct {
	Group GroupHandlerInterface
}

// New creates new group handler.
func New(svc *service.Service) *Handler {
	return &Handler{
		Group: NewGroupHandler(svc),
	}
}
