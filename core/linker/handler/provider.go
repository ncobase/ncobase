package handler

import "ncobase/core/linker/service"

// Handler represents the socket handler.
type Handler struct {
	WebSocket WebSocketHandlerInterface
}

// New creates a new socket handler.
func New(s *service.Service) *Handler {
	return &Handler{
		WebSocket: NewWebSocketHandler(s),
	}
}
