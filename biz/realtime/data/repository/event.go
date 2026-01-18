package repository

import (
	"context"
	"fmt"
	"ncobase/biz/realtime/data"
	"ncobase/biz/realtime/data/ent"
	eventEnt "ncobase/biz/realtime/data/ent/event"
	"ncobase/biz/realtime/structs"
	"time"

	nd "github.com/ncobase/ncore/data"

	"github.com/ncobase/ncore/data/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/redis/go-redis/v9"

	"github.com/ncobase/ncore/data/search"
)

// EventRepositoryInterface defines event repository operations
type EventRepositoryInterface interface {
	Create(ctx context.Context, event *ent.EventCreate) (*ent.Event, error)
	Get(ctx context.Context, id string) (*ent.Event, error)
	Update(ctx context.Context, id string, event *ent.EventUpdateOne) (*ent.Event, error)
	Delete(ctx context.Context, id string) error

	FindByID(ctx context.Context, id string) (*ent.Event, error)
	List(ctx context.Context, params *structs.ListEventParams) ([]*ent.Event, error)
	Count(ctx context.Context, params *structs.ListEventParams) (int, error)
	CountX(ctx context.Context, params *structs.ListEventParams) int

	CreateBatch(ctx context.Context, events []*ent.EventCreate) ([]*ent.Event, error)
	DeleteBatch(ctx context.Context, ids []string) error

	UpdateStatus(ctx context.Context, id string, status string, errorMsg string) error
	UpdateStatusBatch(ctx context.Context, ids []string, status string) error
	GetEventsByStatus(ctx context.Context, status string, limit int) ([]*ent.Event, error)

	SearchEvents(ctx context.Context, query *structs.SearchQuery) ([]*ent.Event, error)
	GetEventsByTimeRange(ctx context.Context, start, end int64) ([]*ent.Event, error)
	GetEventsBySource(ctx context.Context, source string, limit int) ([]*ent.Event, error)
	GetEventsByType(ctx context.Context, eventType string, limit int) ([]*ent.Event, error)
	ListProcessedByTimeRange(ctx context.Context, start, end int64, limit int) ([]*ent.Event, error)
	CountByStatus(ctx context.Context) (map[string]int, error)
	CountByType(ctx context.Context) (map[string]int, error)
	CountBySource(ctx context.Context) (map[string]int, error)

	GetStatsData(ctx context.Context, params *structs.StatsParams) (map[string]any, error)
	GetEventCounts(ctx context.Context, timeRange *structs.TimeRange) (map[string]int64, error)

	GetFailedEvents(ctx context.Context, limit int) ([]*ent.Event, error)
	IncrementRetryCount(ctx context.Context, id string) error
}

type eventRepository struct {
	data *data.Data
	searchClient *search.Client
	ec   *ent.Client
	rc   *redis.Client
	c    *cache.Cache[ent.Event]
}

func NewEventRepository(d *data.Data) EventRepositoryInterface {
	searchClient := nd.NewSearchClient(d.Data)
	return &eventRepository{
		data: d,
		searchClient: searchClient,
		ec:   d.GetMasterEntClient(),
		rc:   d.GetRedis().(*redis.Client),
		c:    cache.NewCache[ent.Event](d.GetRedis().(*redis.Client), "rt_event"),
	}
}

// Create creates a new event
func (r *eventRepository) Create(ctx context.Context, event *ent.EventCreate) (*ent.Event, error) {
	row, err := event.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "eventRepo.Create error: %v", err)
		return nil, err
	}

	// Index in Meilisearch
	if err = r.searchClient.Index(ctx, &search.IndexRequest{Index: "realtime_events", Document: row}); err != nil {
		logger.Errorf(ctx, "eventRepo.Create error creating Meilisearch index: %v", err)
	}

	return row, nil
}

// Get gets an event by ID with cache
func (r *eventRepository) Get(ctx context.Context, id string) (*ent.Event, error) {
	cacheKey := fmt.Sprintf("event:%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	row, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := r.c.Set(ctx, cacheKey, row); err != nil {
		logger.Warnf(ctx, "Failed to set event cache: %v", err)
	}

	return row, nil
}

// Update updates an event
func (r *eventRepository) Update(ctx context.Context, id string, event *ent.EventUpdateOne) (*ent.Event, error) {
	row, err := event.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "eventRepo.Update error: %v", err)
		return nil, err
	}

	// Update Meilisearch index
	if err = r.searchClient.Index(ctx, &search.IndexRequest{Index: "realtime_events", Document: row, DocumentID: row.ID}); err != nil {
		logger.Errorf(ctx, "eventRepo.Update error updating Meilisearch index: %v", err)
	}

	cacheKey := fmt.Sprintf("event:%s", id)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		logger.Warnf(ctx, "Failed to delete event cache: %v", err)
	}

	return row, nil
}

// Delete deletes an event
func (r *eventRepository) Delete(ctx context.Context, id string) error {
	err := r.ec.Event.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return err
	}

	// Delete from Meilisearch
	if err = r.searchClient.Delete(ctx, "realtime_events", id); err != nil {
		logger.Errorf(ctx, "eventRepo.Delete error deleting Meilisearch index: %v", err)
	}

	cacheKey := fmt.Sprintf("event:%s", id)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		logger.Warnf(ctx, "Failed to delete event cache: %v", err)
	}

	return nil
}

// FindByID finds an event by ID
func (r *eventRepository) FindByID(ctx context.Context, id string) (*ent.Event, error) {
	return r.ec.Event.Query().
		Where(eventEnt.ID(id)).
		Only(ctx)
}

// List lists events with filters
func (r *eventRepository) List(ctx context.Context, params *structs.ListEventParams) ([]*ent.Event, error) {
	builder, err := r.buildQuery(ctx, params)
	if validator.IsNotNil(err) {
		return nil, err
	}

	if params.Cursor != "" {
		id, timestamp, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %v", err)
		}

		if !nanoid.IsPrimaryKey(id) {
			return nil, fmt.Errorf("invalid id in cursor: %s", id)
		}

		if params.Direction == "backward" {
			builder.Where(
				eventEnt.Or(
					eventEnt.CreatedAtGT(timestamp),
					eventEnt.And(
						eventEnt.CreatedAtEQ(timestamp),
						eventEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				eventEnt.Or(
					eventEnt.CreatedAtLT(timestamp),
					eventEnt.And(
						eventEnt.CreatedAtEQ(timestamp),
						eventEnt.IDLT(id),
					),
				),
			)
		}
	}

	if params.Direction == "backward" {
		builder.Order(ent.Asc(eventEnt.FieldCreatedAt), ent.Asc(eventEnt.FieldID))
	} else {
		builder.Order(ent.Desc(eventEnt.FieldCreatedAt), ent.Desc(eventEnt.FieldID))
	}

	if params.Limit > 0 {
		builder.Limit(params.Limit)
	}

	rows, err := builder.All(ctx)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// Count returns the total count of events with filters
func (r *eventRepository) Count(ctx context.Context, params *structs.ListEventParams) (int, error) {
	builder, err := r.buildQuery(ctx, params)
	if validator.IsNotNil(err) {
		return 0, err
	}
	return builder.Count(ctx)
}

// CountX gets a count of events
func (r *eventRepository) CountX(ctx context.Context, params *structs.ListEventParams) int {
	builder, err := r.buildQuery(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// CreateBatch creates multiple events in a transaction
func (r *eventRepository) CreateBatch(ctx context.Context, events []*ent.EventCreate) ([]*ent.Event, error) {
	var results []*ent.Event

	tx, err := r.ec.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, e := range events {
		event, err := e.Save(ctx)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		results = append(results, event)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return results, nil
}

// DeleteBatch deletes multiple events in a transaction
func (r *eventRepository) DeleteBatch(ctx context.Context, ids []string) error {
	tx, err := r.ec.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	_, err = tx.Event.Delete().
		Where(eventEnt.IDIn(ids...)).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// UpdateStatus updates an event's status and error message
func (r *eventRepository) UpdateStatus(ctx context.Context, id string, status string, errorMsg string) error {
	update := r.ec.Event.UpdateOneID(id).SetStatus(status)

	if errorMsg != "" {
		update = update.SetErrorMessage(errorMsg)
	}

	if status == "processed" {
		processedAt := time.Now().UnixMilli()
		if ts, ok := ctx.Value("timestamp").(int64); ok && ts > 0 {
			processedAt = ts
		}
		update = update.SetProcessedAt(processedAt)
	}

	err := update.Exec(ctx)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("event:%s", id)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		logger.Warnf(ctx, "Failed to delete event cache: %v", err)
	}

	return nil
}

// UpdateStatusBatch updates status for multiple events
func (r *eventRepository) UpdateStatusBatch(ctx context.Context, ids []string, status string) error {
	tx, err := r.ec.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	update := tx.Event.Update().
		Where(eventEnt.IDIn(ids...)).
		SetStatus(status)

	if status == "processed" {
		update = update.SetProcessedAt(ctx.Value("timestamp").(int64))
	}

	_, err = update.Save(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetEventsByStatus gets events by status
func (r *eventRepository) GetEventsByStatus(ctx context.Context, status string, limit int) ([]*ent.Event, error) {
	if limit <= 0 {
		limit = 100
	}

	return r.ec.Event.Query().
		Where(eventEnt.Status(status)).
		Order(ent.Desc(eventEnt.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
}

// SearchEvents performs search with complex queries
func (r *eventRepository) SearchEvents(ctx context.Context, query *structs.SearchQuery) ([]*ent.Event, error) {
	builder := r.ec.Event.Query()

	// Apply basic filters from search query
	if query.Filters != nil {
		if eventType, ok := query.Filters["type"].(string); ok {
			builder = builder.Where(eventEnt.Type(eventType))
		}
		if source, ok := query.Filters["source"].(string); ok {
			builder = builder.Where(eventEnt.Source(source))
		}
		if status, ok := query.Filters["status"].(string); ok {
			builder = builder.Where(eventEnt.Status(status))
		}
	}

	// Apply time range
	if query.TimeRange != nil {
		// Note: This is simplified. In real implementation, you'd parse ISO 8601 timestamps
		// and convert to Unix timestamps for the database query
	}

	// Apply sorting
	builder = builder.Order(ent.Desc(eventEnt.FieldCreatedAt))

	// Apply pagination
	if query.From > 0 {
		builder = builder.Offset(query.From)
	}
	if query.Size > 0 {
		builder = builder.Limit(query.Size)
	}

	return builder.All(ctx)
}

// GetEventsByTimeRange gets events within time range
func (r *eventRepository) GetEventsByTimeRange(ctx context.Context, start, end int64) ([]*ent.Event, error) {
	return r.ec.Event.Query().
		Where(
			eventEnt.CreatedAtGTE(start),
			eventEnt.CreatedAtLTE(end),
		).
		Order(ent.Desc(eventEnt.FieldCreatedAt)).
		All(ctx)
}

// GetEventsBySource gets events by source
func (r *eventRepository) GetEventsBySource(ctx context.Context, source string, limit int) ([]*ent.Event, error) {
	if limit <= 0 {
		limit = 100
	}

	return r.ec.Event.Query().
		Where(eventEnt.Source(source)).
		Order(ent.Desc(eventEnt.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
}

// GetEventsByType gets events by type
func (r *eventRepository) GetEventsByType(ctx context.Context, eventType string, limit int) ([]*ent.Event, error) {
	if limit <= 0 {
		limit = 100
	}

	return r.ec.Event.Query().
		Where(eventEnt.Type(eventType)).
		Order(ent.Desc(eventEnt.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
}

// ListProcessedByTimeRange returns processed events within the given time range.
func (r *eventRepository) ListProcessedByTimeRange(ctx context.Context, start, end int64, limit int) ([]*ent.Event, error) {
	builder := r.ec.Event.Query().
		Where(eventEnt.ProcessedAtGT(0))

	if start > 0 && end > 0 {
		builder = builder.Where(
			eventEnt.CreatedAtGTE(start),
			eventEnt.CreatedAtLTE(end),
		)
	}

	if limit > 0 {
		builder = builder.Limit(limit)
	}

	return builder.All(ctx)
}

// CountByStatus returns counts grouped by status.
func (r *eventRepository) CountByStatus(ctx context.Context) (map[string]int, error) {
	type row struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}
	var rows []row
	if err := r.ec.Event.Query().
		GroupBy(eventEnt.FieldStatus).
		Aggregate(ent.Count()).
		Scan(ctx, &rows); err != nil {
		return nil, err
	}
	return toCountMap(rows, func(r row) string { return r.Status }, func(r row) int { return r.Count }), nil
}

// CountByType returns counts grouped by type.
func (r *eventRepository) CountByType(ctx context.Context) (map[string]int, error) {
	type row struct {
		Type  string `json:"type"`
		Count int    `json:"count"`
	}
	var rows []row
	if err := r.ec.Event.Query().
		GroupBy(eventEnt.FieldType).
		Aggregate(ent.Count()).
		Scan(ctx, &rows); err != nil {
		return nil, err
	}
	return toCountMap(rows, func(r row) string { return r.Type }, func(r row) int { return r.Count }), nil
}

// CountBySource returns counts grouped by source.
func (r *eventRepository) CountBySource(ctx context.Context) (map[string]int, error) {
	type row struct {
		Source string `json:"source"`
		Count  int    `json:"count"`
	}
	var rows []row
	if err := r.ec.Event.Query().
		GroupBy(eventEnt.FieldSource).
		Aggregate(ent.Count()).
		Scan(ctx, &rows); err != nil {
		return nil, err
	}
	return toCountMap(rows, func(r row) string { return r.Source }, func(r row) int { return r.Count }), nil
}

// GetStatsData gets statistics data for real-time stats
func (r *eventRepository) GetStatsData(ctx context.Context, params *structs.StatsParams) (map[string]any, error) {
	stats := make(map[string]any)

	// Total events count
	total, err := r.ec.Event.Query().Count(ctx)
	if err != nil {
		return nil, err
	}
	stats["total_events"] = total

	// Events by status
	statusCounts := make(map[string]int)
	statuses := []string{"pending", "processed", "failed", "retry"}
	for _, status := range statuses {
		count, err := r.ec.Event.Query().
			Where(eventEnt.Status(status)).
			Count(ctx)
		if err != nil {
			return nil, err
		}
		statusCounts[status] = count
	}
	stats["by_status"] = statusCounts

	// Events by type
	typeCounts := make(map[string]int)
	// This is a simplified implementation. In practice, you'd want to
	// use a more efficient aggregation query
	stats["by_type"] = typeCounts

	return stats, nil
}

func toCountMap[T any](rows []T, keyFn func(T) string, countFn func(T) int) map[string]int {
	result := make(map[string]int, len(rows))
	for _, item := range rows {
		result[keyFn(item)] = countFn(item)
	}
	return result
}

// GetEventCounts gets event counts within time range
func (r *eventRepository) GetEventCounts(ctx context.Context, timeRange *structs.TimeRange) (map[string]int64, error) {
	counts := make(map[string]int64)

	// This is a simplified implementation
	// In practice, you'd parse the time range and perform aggregated queries
	total, err := r.ec.Event.Query().Count(ctx)
	if err != nil {
		return nil, err
	}

	counts["total"] = int64(total)
	return counts, nil
}

// GetFailedEvents gets events that failed processing
func (r *eventRepository) GetFailedEvents(ctx context.Context, limit int) ([]*ent.Event, error) {
	if limit <= 0 {
		limit = 100
	}

	return r.ec.Event.Query().
		Where(eventEnt.Status("failed")).
		Order(ent.Asc(eventEnt.FieldCreatedAt)). // Oldest first for retry
		Limit(limit).
		All(ctx)
}

// IncrementRetryCount increments the retry count for an event
func (r *eventRepository) IncrementRetryCount(ctx context.Context, id string) error {
	err := r.ec.Event.UpdateOneID(id).
		AddRetryCount(1).
		Exec(ctx)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("event:%s", id)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		logger.Warnf(ctx, "Failed to delete event cache: %v", err)
	}

	return nil
}

// buildQuery creates query builder with filters
func (r *eventRepository) buildQuery(ctx context.Context, params *structs.ListEventParams) (*ent.EventQuery, error) {
	builder := r.ec.Event.Query()

	if params.Type != "" {
		builder = builder.Where(eventEnt.Type(params.Type))
	}

	if params.Source != "" {
		builder = builder.Where(eventEnt.Source(params.Source))
	}

	if params.Status != "" {
		builder = builder.Where(eventEnt.Status(params.Status))
	}

	if len(params.TimeRange) == 2 {
		builder = builder.Where(
			eventEnt.CreatedAtGTE(params.TimeRange[0]),
			eventEnt.CreatedAtLTE(params.TimeRange[1]),
		)
	}

	return builder, nil
}
