package service

import (
	"ncobase/core/realtime/data"
	"ncore/extension"
)

// Service represents all services for the realtime module
type Service struct {
	WebSocket    WebSocketService
	Notification NotificationService
	Event        EventService
	Channel      ChannelService
	em           *extension.Manager
}

// New creates a new service provider instance
func New(d *data.Data, em *extension.Manager) *Service {
	// Initialize WebSocket service, Other service depends on it
	wsService := NewWebSocketService(d)
	return &Service{
		WebSocket:    wsService,
		Notification: NewNotificationService(d, wsService),
		Event:        NewEventService(d, wsService),
		Channel:      NewChannelService(d, wsService),
		em:           em,
	}
}

// subscribeToEvents sets up event subscriptions
func (s *Service) subscribeToEvents() {
	// Subscribe to relevant events
}
