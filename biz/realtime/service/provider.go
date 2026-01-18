package service

import (
	"ncobase/biz/realtime/data"

	ext "github.com/ncobase/ncore/extension/types"
)

// Service represents all services for the realtime module
type Service struct {
	WebSocket      WebSocketService
	Notification   NotificationService
	Event          EventService
	Channel        ChannelService
	EventHandler   EventHandlerService
	ChannelManager ChannelManagerService
	em             ext.ManagerInterface
}

// New creates a new service provider instance
func New(d *data.Data, em ext.ManagerInterface) *Service {
	// Initialize WebSocket service, Other service depends on it
	wsService := NewWebSocketService(d)
	evtService := NewEventService(d, wsService)
	chService := NewChannelService(d, wsService)

	svc := &Service{
		WebSocket:      wsService,
		Notification:   NewNotificationService(d, wsService),
		Event:          evtService,
		Channel:        chService,
		EventHandler:   NewEventHandlerService(evtService),
		ChannelManager: NewChannelManagerService(chService),
		em:             em,
	}

	svc.subscribeToEvents()

	return svc
}

// subscribeToEvents sets up event subscriptions
func (s *Service) subscribeToEvents() {
	// Subscribe to relevant events
	if s.EventHandler != nil {
		s.EventHandler.RegisterEventHandlers(s.em)
	}
	if s.ChannelManager != nil {
		s.ChannelManager.RegisterChannelManagers(s.em)
	}
}
