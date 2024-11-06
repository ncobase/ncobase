package handler

import "ncobase/core/realtime/service"

// Handler represents the socket handler.
type Handler struct {
	WebSocket    WebSocketHandler
	Notification NotificationHandler
	Event        EventHandler
	Channel      ChannelHandler
}

// New creates a new socket handler.
func New(s *service.Service) *Handler {
	return &Handler{
		WebSocket:    NewWebSocketHandler(s.WebSocket),
		Notification: NewNotificationHandler(s.Notification),
		Event:        NewEventHandler(s.Event),
		Channel:      NewChannelHandler(s.Channel),
	}
}
