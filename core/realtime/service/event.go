package service

import (
	"context"
	"errors"
	"fmt"
	"ncobase/common/ecode"
	"ncobase/common/log"
	"ncobase/common/paging"
	"ncobase/core/realtime/data"
	"ncobase/core/realtime/data/ent"
	"ncobase/core/realtime/data/repository"
	"ncobase/core/realtime/structs"
	"time"
)

type EventService interface {
	Publish(ctx context.Context, body *structs.CreateEvent) (*structs.ReadEvent, error)
	Get(ctx context.Context, params *structs.FindEvent) (*structs.ReadEvent, error)
	Delete(ctx context.Context, params *structs.FindEvent) error
	List(ctx context.Context, params *structs.ListEventParams) (paging.Result[*structs.ReadEvent], error)

	PublishBatch(ctx context.Context, events []*structs.CreateEvent) ([]*structs.ReadEvent, error)
	DeleteBatch(ctx context.Context, ids []string) error

	GetEventHistory(ctx context.Context, channelID string, eventType string) ([]*structs.ReadEvent, error)
	GetUserEvents(ctx context.Context, userID string) ([]*structs.ReadEvent, error)
	GetChannelEvents(ctx context.Context, channelID string, limit int) ([]*structs.ReadEvent, error)
	GetEventsByTimeRange(ctx context.Context, start, end time.Time) ([]*structs.ReadEvent, error)
}

type eventService struct {
	data        *data.Data
	eventRepo   repository.EventRepositoryInterface
	channelRepo repository.ChannelRepositoryInterface
	ws          WebSocketService
}

func NewEventService(
	d *data.Data,
	ws WebSocketService,
) EventService {
	return &eventService{
		data:        d,
		eventRepo:   repository.NewEventRepository(d),
		channelRepo: repository.NewChannelRepository(d),
		ws:          ws,
	}
}

// Publish publishes a new event
func (s *eventService) Publish(ctx context.Context, body *structs.CreateEvent) (*structs.ReadEvent, error) {
	e := body.Event
	if e.Type == "" || e.ChannelID == "" {
		return nil, errors.New("event type and channel_id are required")
	}

	// Check channel is existed and enabled
	channel, err := s.channelRepo.Get(ctx, e.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("invalid channel: %w", err)
	}
	if channel.Status != 1 {
		return nil, errors.New("channel is disabled")
	}

	// Create event
	event, err := s.eventRepo.Create(ctx, s.data.EC.Event.Create().
		SetType(e.Type).
		SetChannelID(e.ChannelID).
		SetNillableUserID(&e.UserID).
		SetPayload(e.Payload),
	)

	if err != nil {
		log.Errorf(ctx, "Failed to publish event: %v", err)
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	result := s.serializeEvent(event)

	// Broadcast event
	s.broadcastEvent(result)

	return result, nil
}

// Get retrieves an event by ID
func (s *eventService) Get(ctx context.Context, params *structs.FindEvent) (*structs.ReadEvent, error) {
	event, err := s.eventRepo.Get(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	return s.serializeEvent(event), nil
}

// Delete deletes an event
func (s *eventService) Delete(ctx context.Context, params *structs.FindEvent) error {
	return s.eventRepo.Delete(ctx, params.ID)
}

// List lists events with filters
func (s *eventService) List(ctx context.Context, params *structs.ListEventParams) (paging.Result[*structs.ReadEvent], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadEvent, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, err := s.eventRepo.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			log.Errorf(ctx, "Error listing permissions: %v", err)
			return nil, 0, err
		}

		total := s.eventRepo.CountX(ctx, params)

		return s.serializeEvents(rows), total, nil
	})
}

// PublishBatch publishes multiple events
func (s *eventService) PublishBatch(ctx context.Context, bodies []*structs.CreateEvent) ([]*structs.ReadEvent, error) {
	var creates []*ent.EventCreate

	// Preprocessing events
	for _, body := range bodies {
		e := body.Event
		if e.Type == "" || e.ChannelID == "" {
			return nil, errors.New("event type and channel_id are required for all events")
		}

		creates = append(creates, s.data.EC.Event.Create().
			SetType(e.Type).
			SetChannelID(e.ChannelID).
			SetNillableUserID(&e.UserID).
			SetPayload(e.Payload),
		)
	}

	// Batch create events
	events, err := s.eventRepo.CreateBatch(ctx, creates)
	if err != nil {
		return nil, err
	}

	results := s.serializeEvents(events)

	// Broadcast events
	for _, result := range results {
		s.broadcastEvent(result)
	}

	return results, nil
}

// DeleteBatch deletes multiple events
func (s *eventService) DeleteBatch(ctx context.Context, ids []string) error {
	return s.eventRepo.DeleteBatch(ctx, ids)
}

// GetEventHistory gets event history for a channel and event type
func (s *eventService) GetEventHistory(ctx context.Context, channelID string, eventType string) ([]*structs.ReadEvent, error) {
	events, err := s.eventRepo.GetEventHistory(ctx, channelID, eventType, -1)
	if err != nil {
		return nil, err
	}

	return s.serializeEvents(events), nil
}

// GetUserEvents gets events for a specific user
func (s *eventService) GetUserEvents(ctx context.Context, userID string) ([]*structs.ReadEvent, error) {
	events, err := s.eventRepo.GetEventsByUserID(ctx, userID, -1)
	if err != nil {
		return nil, err
	}

	return s.serializeEvents(events), nil
}

// GetChannelEvents gets events for a specific channel
func (s *eventService) GetChannelEvents(ctx context.Context, channelID string, limit int) ([]*structs.ReadEvent, error) {
	events, err := s.eventRepo.List(ctx, &structs.ListEventParams{ChannelID: channelID, Limit: limit})
	if err != nil {
		return nil, err
	}

	return s.serializeEvents(events), nil
}

// GetEventsByTimeRange gets events within a time range
func (s *eventService) GetEventsByTimeRange(ctx context.Context, start, end time.Time) ([]*structs.ReadEvent, error) {
	events, err := s.eventRepo.List(ctx, &structs.ListEventParams{TimeRange: []int64{start.Unix(), end.Unix()}})
	if err != nil {
		return nil, err
	}

	return s.serializeEvents(events), nil
}

// Serialization helpers
func (s *eventService) serializeEvent(e *ent.Event) *structs.ReadEvent {
	return &structs.ReadEvent{
		ID:        e.ID,
		Type:      e.Type,
		ChannelID: e.ChannelID,
		UserID:    e.UserID,
		Payload:   e.Payload,
		CreatedAt: e.CreatedAt,
	}
}

func (s *eventService) serializeEvents(events []*ent.Event) []*structs.ReadEvent {
	result := make([]*structs.ReadEvent, len(events))
	for i, e := range events {
		result[i] = s.serializeEvent(e)
	}
	return result
}

// broadcastEvent broadcasts an event through WebSocket
func (s *eventService) broadcastEvent(e *structs.ReadEvent) {
	message := &WebSocketMessage{
		Type:    "event",
		Channel: e.ChannelID,
		Data:    e,
	}

	err := s.ws.BroadcastToChannel(e.ChannelID, message)
	if err != nil {
		log.Errorf(context.Background(), "Failed to broadcast event: %v", err)
	}
}
