package service

import (
	"context"
	"errors"
	"fmt"
	"ncobase/realtime/data"
	"ncobase/realtime/data/ent"
	"ncobase/realtime/data/repository"
	"ncobase/realtime/structs"
	"time"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils/nanoid"
)

type EventService interface {
	Publish(ctx context.Context, body *structs.CreateEvent) (*structs.ReadEvent, error)
	Get(ctx context.Context, params *structs.FindEvent) (*structs.ReadEvent, error)
	Delete(ctx context.Context, params *structs.FindEvent) error
	List(ctx context.Context, params *structs.ListEventParams) (paging.Result[*structs.ReadEvent], error)

	PublishBatch(ctx context.Context, events []*structs.CreateEvent) ([]*structs.ReadEvent, error)
	DeleteBatch(ctx context.Context, ids []string) error

	Search(ctx context.Context, query *structs.SearchQuery) (*structs.SearchResult, error)
	GetRealtimeStats(ctx context.Context, params *structs.StatsParams) (*structs.RealtimeStats, error)

	RetryEvent(ctx context.Context, eventID string, params *structs.RetryParams) (*structs.RetryResult, error)
	GetFailedEvents(ctx context.Context, limit int) ([]*structs.ReadEvent, error)

	UpdateEventStatus(ctx context.Context, eventID string, status string, errorMsg string) error
	ProcessPendingEvents(ctx context.Context, limit int) ([]*structs.ReadEvent, error)
}

type eventService struct {
	data      *data.Data
	eventRepo repository.EventRepositoryInterface
	ws        WebSocketService
}

func NewEventService(d *data.Data, ws WebSocketService) EventService {
	return &eventService{
		data:      d,
		eventRepo: repository.NewEventRepository(d),
		ws:        ws,
	}
}

// Publish publishes a new event
func (s *eventService) Publish(ctx context.Context, body *structs.CreateEvent) (*structs.ReadEvent, error) {
	e := body.Event
	if e.Type == "" {
		return nil, errors.New("event type is required")
	}

	// Set default values
	if e.Priority == "" {
		e.Priority = "normal"
	}

	// Create event
	event, err := s.eventRepo.Create(ctx, s.data.EC.Event.Create().
		SetType(e.Type).
		SetSource(e.Source).
		SetPayload(e.Payload).
		SetPriority(e.Priority).
		SetStatus("pending"),
	)

	if err != nil {
		logger.Errorf(ctx, "Failed to publish event: %v", err)
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	result := s.serializeEvent(event)

	// Broadcast event via WebSocket
	s.broadcastEvent(result)

	// Process event asynchronously
	go s.processEventAsync(context.Background(), event.ID)

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
			logger.Errorf(ctx, "Error listing events: %v", err)
			return nil, 0, err
		}

		total := s.eventRepo.CountX(ctx, params)

		return s.serializeEvents(rows), total, nil
	})
}

// PublishBatch publishes multiple events
func (s *eventService) PublishBatch(ctx context.Context, bodies []*structs.CreateEvent) ([]*structs.ReadEvent, error) {
	var creates []*ent.EventCreate

	for _, body := range bodies {
		e := body.Event
		if e.Type == "" {
			return nil, errors.New("event type is required for all events")
		}

		if e.Priority == "" {
			e.Priority = "normal"
		}

		creates = append(creates, s.data.EC.Event.Create().
			SetType(e.Type).
			SetSource(e.Source).
			SetPayload(e.Payload).
			SetPriority(e.Priority).
			SetStatus("pending"),
		)
	}

	events, err := s.eventRepo.CreateBatch(ctx, creates)
	if err != nil {
		return nil, err
	}

	results := s.serializeEvents(events)

	// Broadcast events
	for _, result := range results {
		s.broadcastEvent(result)
	}

	// Process events asynchronously
	for _, event := range events {
		go s.processEventAsync(context.Background(), event.ID)
	}

	return results, nil
}

// DeleteBatch deletes multiple events
func (s *eventService) DeleteBatch(ctx context.Context, ids []string) error {
	return s.eventRepo.DeleteBatch(ctx, ids)
}

// Search performs complex search queries
func (s *eventService) Search(ctx context.Context, query *structs.SearchQuery) (*structs.SearchResult, error) {
	// For basic implementation, we'll use the repository search
	// In production, this would integrate with Elasticsearch
	events, err := s.eventRepo.SearchEvents(ctx, query)
	if err != nil {
		return nil, err
	}

	// Get total count for pagination
	totalCount, err := s.eventRepo.Count(ctx, &structs.ListEventParams{
		Type:   getStringFromFilters(query.Filters, "type"),
		Source: getStringFromFilters(query.Filters, "source"),
		Status: getStringFromFilters(query.Filters, "status"),
	})
	if err != nil {
		logger.Warnf(ctx, "Failed to get total count: %v", err)
		totalCount = len(events)
	}

	result := &structs.SearchResult{
		Total:  int64(totalCount),
		Events: s.serializeEvents(events),
	}

	// Handle aggregations if requested
	if query.Aggregations != nil {
		result.Aggregations = s.performAggregations(ctx, query.Aggregations)
	}

	return result, nil
}

// GetRealtimeStats gets real-time statistics
func (s *eventService) GetRealtimeStats(ctx context.Context, params *structs.StatsParams) (*structs.RealtimeStats, error) {
	now := time.Now()

	// Get basic stats from repository
	statsData, err := s.eventRepo.GetStatsData(ctx, params)
	if err != nil {
		return nil, err
	}

	// Calculate real-time metrics
	metrics := make(map[string]any)

	// Total events
	if total, ok := statsData["total_events"]; ok {
		metrics["total_events"] = total
	}

	// Events per second (simplified calculation)
	// In production, this would use time-series data
	metrics["events_per_second"] = s.calculateEventsPerSecond(ctx, params.Interval)

	// Error rate
	if statusCounts, ok := statsData["by_status"].(map[string]int); ok {
		total := 0
		failed := 0
		for status, count := range statusCounts {
			total += count
			if status == "failed" {
				failed = count
			}
		}
		if total > 0 {
			metrics["error_rate"] = float64(failed) / float64(total)
		} else {
			metrics["error_rate"] = 0.0
		}
	}

	// Average processing time (placeholder)
	metrics["avg_processing_time_ms"] = 45.0

	breakdown := make(map[string]map[string]any)
	if statusCounts, ok := statsData["by_status"]; ok {
		breakdown["by_status"] = map[string]any{}
		if counts, ok := statusCounts.(map[string]int); ok {
			for k, v := range counts {
				breakdown["by_status"][k] = v
			}
		}
	}

	return &structs.RealtimeStats{
		Timestamp: now.Format(time.RFC3339),
		Interval:  params.Interval,
		Metrics:   metrics,
		Breakdown: breakdown,
	}, nil
}

// RetryEvent retries a failed event
func (s *eventService) RetryEvent(ctx context.Context, eventID string, params *structs.RetryParams) (*structs.RetryResult, error) {
	// Get the event
	event, err := s.eventRepo.Get(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Check if event can be retried
	if event.Status != "failed" && event.Status != "retry" {
		return nil, fmt.Errorf("event status is %s, cannot retry", event.Status)
	}

	// Check retry limits
	maxRetries := 3
	if params.RetryOptions != nil && params.RetryOptions.MaxAttempts > 0 {
		maxRetries = params.RetryOptions.MaxAttempts
	}

	if event.RetryCount >= maxRetries {
		return nil, fmt.Errorf("maximum retry attempts (%d) exceeded", maxRetries)
	}

	// Increment retry count
	err = s.eventRepo.IncrementRetryCount(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Update status to retry
	priority := params.Priority
	if priority == "" {
		priority = event.Priority
	}

	err = s.eventRepo.UpdateStatus(ctx, eventID, "retry", "")
	if err != nil {
		return nil, err
	}

	// Calculate schedule time
	delay := 60 // default 60 seconds
	if params.RetryOptions != nil && params.RetryOptions.DelaySeconds > 0 {
		delay = params.RetryOptions.DelaySeconds
	}

	scheduledAt := time.Now().Add(time.Duration(delay) * time.Second)

	// Schedule retry (in production, this would use a job queue)
	go s.scheduleRetry(context.Background(), eventID, scheduledAt)

	return &structs.RetryResult{
		RetryID:     nanoid.String(),
		ScheduledAt: scheduledAt.Format(time.RFC3339),
	}, nil
}

// GetFailedEvents gets events that failed processing
func (s *eventService) GetFailedEvents(ctx context.Context, limit int) ([]*structs.ReadEvent, error) {
	events, err := s.eventRepo.GetFailedEvents(ctx, limit)
	if err != nil {
		return nil, err
	}

	return s.serializeEvents(events), nil
}

// UpdateEventStatus updates an event's status
func (s *eventService) UpdateEventStatus(ctx context.Context, eventID string, status string, errorMsg string) error {
	return s.eventRepo.UpdateStatus(ctx, eventID, status, errorMsg)
}

// ProcessPendingEvents processes pending events
func (s *eventService) ProcessPendingEvents(ctx context.Context, limit int) ([]*structs.ReadEvent, error) {
	events, err := s.eventRepo.GetEventsByStatus(ctx, "pending", limit)
	if err != nil {
		return nil, err
	}

	var results []*structs.ReadEvent
	for _, event := range events {
		// Process event
		err := s.processEvent(ctx, event)
		if err != nil {
			logger.Errorf(ctx, "Failed to process event %s: %v", event.ID, err)
			// Update status to failed
			s.eventRepo.UpdateStatus(ctx, event.ID, "failed", err.Error())
		} else {
			// Update status to processed
			s.eventRepo.UpdateStatus(ctx, event.ID, "processed", "")
		}

		results = append(results, s.serializeEvent(event))
	}

	return results, nil
}

// Helper methods

// serializeEvent converts ent.Event to structs.ReadEvent
func (s *eventService) serializeEvent(e *ent.Event) *structs.ReadEvent {
	result := &structs.ReadEvent{
		ID:         e.ID,
		Type:       e.Type,
		Source:     e.Source,
		Payload:    e.Payload,
		Status:     e.Status,
		Priority:   e.Priority,
		CreatedAt:  e.CreatedAt,
		RetryCount: e.RetryCount,
	}

	if e.ProcessedAt != 0 {
		result.ProcessedAt = &e.ProcessedAt
	}

	if e.ErrorMessage != "" {
		result.ErrorMessage = e.ErrorMessage
	}

	return result
}

// serializeEvents converts []*ent.Event to []*structs.ReadEvent
func (s *eventService) serializeEvents(events []*ent.Event) []*structs.ReadEvent {
	result := make([]*structs.ReadEvent, len(events))
	for i, e := range events {
		result[i] = s.serializeEvent(e)
	}
	return result
}

// broadcastEvent broadcasts an event through WebSocket
func (s *eventService) broadcastEvent(e *structs.ReadEvent) {
	if s.ws == nil {
		return
	}

	message := &WebSocketMessage{
		Type: "event",
		Data: e,
	}

	err := s.ws.BroadcastToAll(message)
	if err != nil {
		logger.Errorf(context.Background(), "Failed to broadcast event: %v", err)
	}
}

// processEvent processes a single event
func (s *eventService) processEvent(ctx context.Context, event *ent.Event) error {
	// Implement event processing logic here
	// This is where you'd add business logic for different event types

	logger.Infof(ctx, "Processing event %s of type %s", event.ID, event.Type)

	// Simulate processing
	time.Sleep(10 * time.Millisecond)

	return nil
}

// processEventAsync processes an event asynchronously
func (s *eventService) processEventAsync(ctx context.Context, eventID string) {
	event, err := s.eventRepo.Get(ctx, eventID)
	if err != nil {
		logger.Errorf(ctx, "Failed to get event for async processing: %v", err)
		return
	}

	err = s.processEvent(ctx, event)
	if err != nil {
		s.eventRepo.UpdateStatus(ctx, eventID, "failed", err.Error())
	} else {
		currentTime := time.Now().Unix()
		ctxWithTime := context.WithValue(ctx, "timestamp", currentTime)
		s.eventRepo.UpdateStatus(ctxWithTime, eventID, "processed", "")
	}
}

// scheduleRetry schedules a retry for later execution
func (s *eventService) scheduleRetry(ctx context.Context, eventID string, scheduledAt time.Time) {
	// In production, this would use a proper job queue like RabbitMQ delayed messages
	time.Sleep(time.Until(scheduledAt))

	event, err := s.eventRepo.Get(ctx, eventID)
	if err != nil {
		logger.Errorf(ctx, "Failed to get event for retry: %v", err)
		return
	}

	err = s.processEvent(ctx, event)
	if err != nil {
		s.eventRepo.UpdateStatus(ctx, eventID, "failed", err.Error())
	} else {
		currentTime := time.Now().Unix()
		ctxWithTime := context.WithValue(ctx, "timestamp", currentTime)
		s.eventRepo.UpdateStatus(ctxWithTime, eventID, "processed", "")
	}
}

// calculateEventsPerSecond calculates events per second
func (s *eventService) calculateEventsPerSecond(ctx context.Context, interval string) float64 {
	// Simplified calculation - in production, use time-series data
	// This would query events from the last minute and calculate rate
	return 1250.0 // placeholder
}

// performAggregations performs aggregations on search results
func (s *eventService) performAggregations(ctx context.Context, aggregations map[string]any) map[string]any {
	// Simplified aggregation - in production, use Elasticsearch aggregations
	result := make(map[string]any)

	// Placeholder aggregation results
	result["event_trends"] = map[string]any{
		"buckets": []map[string]any{
			{"key": "2025-08-03T10:00:00Z", "doc_count": 100},
			{"key": "2025-08-03T11:00:00Z", "doc_count": 150},
		},
	}

	return result
}

// getStringFromFilters gets string value from filters map
func getStringFromFilters(filters map[string]any, key string) string {
	if filters == nil {
		return ""
	}
	if value, ok := filters[key].(string); ok {
		return value
	}
	return ""
}
