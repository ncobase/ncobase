package handler

import (
	"ncobase/proxy/service"
)

// Handler represents the proxy handler.
type Handler struct {
	Endpoint    EndpointHandlerInterface
	Route       RouteHandlerInterface
	Transformer TransformerHandlerInterface
	Dynamic     DynamicHandlerInterface
	WebSocket   WebSocketHandlerInterface
}

// New creates a new handler.
func New(s *service.Service) *Handler {
	return &Handler{
		Endpoint:    NewEndpointHandler(s),
		Route:       NewRouteHandler(s),
		Transformer: NewTransformerHandler(s),
		Dynamic:     NewDynamicHandler(s),
		WebSocket:   NewWebSocketHandler(s),
	}
}
