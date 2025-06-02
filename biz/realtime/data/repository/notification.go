package repository

import (
	"context"
	"fmt"
	"ncobase/realtime/data"
	"ncobase/realtime/data/ent"
	notificationEnt "ncobase/realtime/data/ent/notification"
	"ncobase/realtime/structs"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/redis/go-redis/v9"
)

// NotificationRepositoryInterface represents the notification repository interface
type NotificationRepositoryInterface interface {
	Create(ctx context.Context, notification *ent.NotificationCreate) (*ent.Notification, error)
	Get(ctx context.Context, id string) (*ent.Notification, error)
	Update(ctx context.Context, id string, notification *ent.NotificationUpdateOne) (*ent.Notification, error)
	Delete(ctx context.Context, id string) error

	FindByID(ctx context.Context, id string) (*ent.Notification, error)
	List(ctx context.Context, params *structs.ListNotificationParams) ([]*ent.Notification, error)
	Count(ctx context.Context, params *structs.ListNotificationParams) (int, error)
	CountX(ctx context.Context, params *structs.ListNotificationParams) int

	CreateBatch(ctx context.Context, notifications []*ent.NotificationCreate) ([]*ent.Notification, error)
	DeleteBatch(ctx context.Context, ids []string) error

	UpdateStatus(ctx context.Context, id string, status int) error
	UpdateStatusBatch(ctx context.Context, userID string, status int) error
}

type notificationRepository struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.Notification]
}

func NewNotificationRepository(d *data.Data) NotificationRepositoryInterface {
	ec := d.GetMasterEntClient()
	rc := d.GetRedis()
	return &notificationRepository{
		ec: ec,
		rc: rc,
		c:  cache.NewCache[ent.Notification](rc, "rt_notification"),
	}
}

// Create creates a new notification
func (r *notificationRepository) Create(ctx context.Context, notification *ent.NotificationCreate) (*ent.Notification, error) {
	row, err := notification.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "notificationRepo.Create error: %v", err)
		return nil, err
	}
	return row, nil
}

// Get gets a notification by ID with cache
func (r *notificationRepository) Get(ctx context.Context, id string) (*ent.Notification, error) {
	// Check cache
	cacheKey := fmt.Sprintf("notification:%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Get from database
	row, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Set cache
	if err := r.c.Set(ctx, cacheKey, row); err != nil {
		logger.Warnf(ctx, "Failed to set notification cache: %v", err)
	}

	return row, nil
}

// Update updates a notification
func (r *notificationRepository) Update(ctx context.Context, id string, notification *ent.NotificationUpdateOne) (*ent.Notification, error) {
	row, err := notification.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "notificationRepo.Update error: %v", err)
		return nil, err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("notification:%s", id)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		logger.Warnf(ctx, "Failed to delete notification cache: %v", err)
	}

	return row, nil
}

// Delete deletes a notification
func (r *notificationRepository) Delete(ctx context.Context, id string) error {
	err := r.ec.Notification.DeleteOneID(id).Exec(ctx)
	if err != nil {
		logger.Errorf(ctx, "notificationRepo.Delete error: %v", err)
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("notification:%s", id)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		logger.Warnf(ctx, "Failed to delete notification cache: %v", err)
	}

	return nil
}

// FindByID finds a notification by ID
func (r *notificationRepository) FindByID(ctx context.Context, id string) (*ent.Notification, error) {
	return r.ec.Notification.Query().
		Where(notificationEnt.ID(id)).
		Only(ctx)
}

// List lists notifications and filters
func (r *notificationRepository) List(ctx context.Context, params *structs.ListNotificationParams) ([]*ent.Notification, error) {
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
				notificationEnt.Or(
					notificationEnt.CreatedAtGT(timestamp),
					notificationEnt.And(
						notificationEnt.CreatedAtEQ(timestamp),
						notificationEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				notificationEnt.Or(
					notificationEnt.CreatedAtLT(timestamp),
					notificationEnt.And(
						notificationEnt.CreatedAtEQ(timestamp),
						notificationEnt.IDLT(id),
					),
				),
			)
		}
	}

	if params.Direction == "backward" {
		builder.Order(ent.Asc(notificationEnt.FieldCreatedAt), ent.Asc(notificationEnt.FieldID))
	} else {
		builder.Order(ent.Desc(notificationEnt.FieldCreatedAt), ent.Desc(notificationEnt.FieldID))
	}

	builder.Limit(params.Limit)

	rows, err := builder.All(ctx)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// Count returns the total count of notifications with filters
func (r *notificationRepository) Count(ctx context.Context, params *structs.ListNotificationParams) (int, error) {
	builder, err := r.buildQuery(ctx, params)
	if validator.IsNotNil(err) {
		return 0, err
	}
	return builder.Count(ctx)
}

// CountX gets a count of notifications.
func (r *notificationRepository) CountX(ctx context.Context, params *structs.ListNotificationParams) int {
	builder, err := r.buildQuery(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// UpdateStatus updates a notification's status
func (r *notificationRepository) UpdateStatus(ctx context.Context, id string, status int) error {
	err := r.ec.Notification.UpdateOneID(id).
		SetStatus(status).
		Exec(ctx)

	if err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("notification:%s", id)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		logger.Warnf(ctx, "Failed to delete notification cache: %v", err)
	}

	return nil
}

// CreateBatch creates multiple notifications in a transaction
func (r *notificationRepository) CreateBatch(ctx context.Context, notifications []*ent.NotificationCreate) ([]*ent.Notification, error) {
	var results []*ent.Notification

	// Start transaction
	tx, err := r.ec.Tx(ctx)
	if err != nil {
		return nil, err
	}

	// Rollback on failure, commit on success
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	// Execute operations
	for _, n := range notifications {
		notification, err := n.Save(ctx)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		results = append(results, notification)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return results, nil
}

// DeleteBatch deletes multiple notifications in a transaction
func (r *notificationRepository) DeleteBatch(ctx context.Context, ids []string) error {
	// Start transaction
	tx, err := r.ec.Tx(ctx)
	if err != nil {
		return err
	}

	// Rollback on failure, commit on success
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	// Execute operation
	_, err = tx.Notification.Delete().
		Where(notificationEnt.IDIn(ids...)).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	return tx.Commit()
}

// UpdateStatusBatch updates status for multiple notifications of a user
func (r *notificationRepository) UpdateStatusBatch(ctx context.Context, userID string, status int) error {
	// Start transaction
	tx, err := r.ec.Tx(ctx)
	if err != nil {
		return err
	}

	// Rollback on failure, commit on success
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	// Execute operation
	_, err = tx.Notification.Update().
		Where(
			notificationEnt.UserID(userID),
			notificationEnt.StatusNEQ(status),
		).
		SetStatus(status).
		Save(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	return tx.Commit()
}

// buildQuery creates list builder.
func (r *notificationRepository) buildQuery(ctx context.Context, params *structs.ListNotificationParams) (*ent.NotificationQuery, error) {
	// create builder.
	builder := r.ec.Notification.Query()

	return builder, nil
}
