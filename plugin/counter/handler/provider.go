package handler

import "ncobase/counter/service"

// Handler represents the counter handler.
type Handler struct {
	Counter CounterHandlerInterface
}

// New creates new counter handler.
func New(svc *service.Service) *Handler {
	return &Handler{
		Counter: NewCounterHandler(svc),
	}
}
