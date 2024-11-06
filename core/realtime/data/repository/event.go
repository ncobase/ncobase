package repository

import (
	"context"
	"fmt"
	"ncobase/common/cache"
	"ncobase/common/log"
	"ncobase/common/nanoid"
	"ncobase/common/paging"
	"ncobase/common/validator"
	"ncobase/core/realtime/data"
	"ncobase/core/realtime/data/ent"
	eventEnt "ncobase/core/realtime/data/ent/event"
	"ncobase/core/realtime/structs"

	"github.com/redis/go-redis/v9"
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

	GetEventHistory(ctx context.Context, channelID, eventType string, limit int) ([]*ent.Event, error)
	GetEventsByUserID(ctx context.Context, userID string, limit int) ([]*ent.Event, error)
}

type eventRepository struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.Event]
}

func NewEventRepository(d *data.Data) EventRepositoryInterface {
	return &eventRepository{
		ec: d.GetEntClient(),
		rc: d.GetRedis(),
		c:  cache.NewCache[ent.Event](d.GetRedis(), "rt_event"),
	}
}

// Create creates a new event
func (r *eventRepository) Create(ctx context.Context, event *ent.EventCreate) (*ent.Event, error) {
	row, err := event.Save(ctx)
	if err != nil {
		log.Errorf(ctx, "eventRepo.Create error: %v", err)
		return nil, err
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
		log.Warnf(ctx, "Failed to set event cache: %v", err)
	}

	return row, nil
}

// Update updates an event
func (r *eventRepository) Update(ctx context.Context, id string, event *ent.EventUpdateOne) (*ent.Event, error) {
	row, err := event.Save(ctx)
	if err != nil {
		log.Errorf(ctx, "eventRepo.Update error: %v", err)
		return nil, err
	}

	cacheKey := fmt.Sprintf("event:%s", id)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		log.Warnf(ctx, "Failed to delete event cache: %v", err)
	}

	return row, nil
}

// Delete deletes an event
func (r *eventRepository) Delete(ctx context.Context, id string) error {
	err := r.ec.Event.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("event:%s", id)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		log.Warnf(ctx, "Failed to delete event cache: %v", err)
	}

	return nil
}

// FindByID finds an event by ID
func (r *eventRepository) FindByID(ctx context.Context, id string) (*ent.Event, error) {
	return r.ec.Event.Query().
		Where(eventEnt.ID(id)).
		Only(ctx)
}

// List lists events with pagination and filters
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

	builder.Limit(params.Limit)

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

// CountX gets a count of channels.
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

// GetEventHistory gets event history for a channel and event type
func (r *eventRepository) GetEventHistory(ctx context.Context, channelID, eventType string, limit int) ([]*ent.Event, error) {
	if limit <= 0 {
		limit = 100 // default limit
	}

	return r.ec.Event.Query().
		Where(
			eventEnt.ChannelID(channelID),
			eventEnt.Type(eventType),
		).
		Order(ent.Desc(eventEnt.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
}

// GetEventsByUserID gets events for a specific user
func (r *eventRepository) GetEventsByUserID(ctx context.Context, userID string, limit int) ([]*ent.Event, error) {
	if limit <= 0 {
		limit = 100 // default limit
	}

	return r.ec.Event.Query().
		Where(eventEnt.UserID(userID)).
		Order(ent.Desc(eventEnt.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
}

// buildQuery creates list builder.
func (r *eventRepository) buildQuery(ctx context.Context, params *structs.ListEventParams) (*ent.EventQuery, error) {
	// create builder.
	builder := r.ec.Event.Query()

	return builder, nil
}

// // buildQuery builds a query with filters
// func (r *eventRepository) buildQuery(filters map[string]any) *ent.EventQuery {
// 	query := r.ec.Event.Query()
//
// 	for key, value := range filters {
// 		switch key {
// 		case "type":
// 			if typ, ok := value.(string); ok {
// 				query = query.Where(eventEnt.Type(typ))
// 			}
// 		case "channel_id":
// 			if channelID, ok := value.(string); ok {
// 				query = query.Where(eventEnt.ChannelID(channelID))
// 			}
// 		case "user_id":
// 			if userID, ok := value.(string); ok {
// 				query = query.Where(eventEnt.UserID(userID))
// 			}
// 		case "time_range":
// 			if timeRange, ok := value.([]int64); ok && len(timeRange) == 2 {
// 				query = query.Where(
// 					eventEnt.CreatedAtGTE(timeRange[0]),
// 					eventEnt.CreatedAtLTE(timeRange[1]),
// 				)
// 			}
// 		}
// 	}
//
// 	return query
// }
