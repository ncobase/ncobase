package service

import (
	"context"
	"errors"

	"ncobase/realtime/structs"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// EventHandlerService bridges extension events into realtime events.
type EventHandlerService interface {
	RegisterEventHandlers(em ext.ManagerInterface)
	HandleEvent(ctx context.Context, eventType string, payload map[string]any) error
}

type eventHandlerService struct {
	eventService EventService
}

// NewEventHandlerService creates a new event handler service.
func NewEventHandlerService(eventService EventService) EventHandlerService {
	return &eventHandlerService{
		eventService: eventService,
	}
}

func (s *eventHandlerService) RegisterEventHandlers(em ext.ManagerInterface) {
	if em == nil {
		logger.Warnf(context.Background(), "Extension manager is nil, cannot register event handlers")
		return
	}

	// Register custom event subscriptions here for domain-specific events.
}

func (s *eventHandlerService) HandleEvent(ctx context.Context, eventType string, payload map[string]any) error {
	if eventType == "" {
		return errors.New("event_type is required")
	}

	eventPayload := structs.EventBody{
		Type:    eventType,
		Source:  "extension",
		Payload: payload,
	}

	_, err := s.eventService.Publish(ctx, &structs.CreateEvent{Event: eventPayload})
	if err != nil {
		logger.Errorf(ctx, "Failed to publish event %s: %v", eventType, err)
		return err
	}

	return nil
}
