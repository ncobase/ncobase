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
	channelEnt "ncobase/core/realtime/data/ent/channel"
	subscriptionEnt "ncobase/core/realtime/data/ent/subscription"
	"ncobase/core/realtime/structs"

	"github.com/redis/go-redis/v9"
)

// ChannelRepositoryInterface defines channel repository operations
type ChannelRepositoryInterface interface {
	Create(ctx context.Context, channel *ent.ChannelCreate) (*ent.Channel, error)
	Get(ctx context.Context, id string) (*ent.Channel, error)
	Update(ctx context.Context, id string, channel *ent.ChannelUpdateOne) (*ent.Channel, error)
	Delete(ctx context.Context, id string) error

	FindByID(ctx context.Context, id string) (*ent.Channel, error)
	FindByName(ctx context.Context, name string) (*ent.Channel, error)
	List(ctx context.Context, params *structs.ListChannelParams) ([]*ent.Channel, error)
	Count(ctx context.Context, params *structs.ListChannelParams) (int, error)
	CountX(ctx context.Context, params *structs.ListChannelParams) int

	CreateBatch(ctx context.Context, channels []*ent.ChannelCreate) ([]*ent.Channel, error)
	DeleteBatch(ctx context.Context, ids []string) error

	UpdateStatus(ctx context.Context, id string, status int) error
	UpdateStatusBatch(ctx context.Context, ids []string, status int) error
}

type channelRepository struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.Channel]
}

func NewChannelRepository(d *data.Data) ChannelRepositoryInterface {
	return &channelRepository{
		ec: d.GetEntClient(),
		rc: d.GetRedis(),
		c:  cache.NewCache[ent.Channel](d.GetRedis(), "rt_channel"),
	}
}

// Create creates a new channel
func (r *channelRepository) Create(ctx context.Context, channel *ent.ChannelCreate) (*ent.Channel, error) {
	row, err := channel.Save(ctx)
	if err != nil {
		log.Errorf(ctx, "channelRepo.Create error: %v", err)
		return nil, err
	}
	return row, nil
}

// Get gets a channel by ID with cache
func (r *channelRepository) Get(ctx context.Context, id string) (*ent.Channel, error) {
	cacheKey := fmt.Sprintf("channel:%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	row, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := r.c.Set(ctx, cacheKey, row); err != nil {
		log.Warnf(ctx, "Failed to set channel cache: %v", err)
	}

	return row, nil
}

// FindByID finds a channel by ID
func (r *channelRepository) FindByID(ctx context.Context, id string) (*ent.Channel, error) {
	return r.ec.Channel.Query().
		Where(channelEnt.ID(id)).
		Only(ctx)
}

// FindByName finds a channel by name
func (r *channelRepository) FindByName(ctx context.Context, name string) (*ent.Channel, error) {
	return r.ec.Channel.Query().
		Where(channelEnt.Name(name)).
		Only(ctx)
}

// Update updates a channel
func (r *channelRepository) Update(ctx context.Context, id string, channel *ent.ChannelUpdateOne) (*ent.Channel, error) {
	row, err := channel.Save(ctx)
	if err != nil {
		log.Errorf(ctx, "channelRepo.Update error: %v", err)
		return nil, err
	}

	cacheKey := fmt.Sprintf("channel:%s", id)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		log.Warnf(ctx, "Failed to delete channel cache: %v", err)
	}

	return row, nil
}

// Delete deletes a channel
func (r *channelRepository) Delete(ctx context.Context, id string) error {
	// Start transaction
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

	// Delete subscriptions first
	_, err = tx.Subscription.Delete().
		Where(subscriptionEnt.ChannelID(id)).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Then delete the channel
	err = tx.Channel.DeleteOneID(id).Exec(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	// Clear cache
	cacheKey := fmt.Sprintf("channel:%s", id)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		log.Warnf(ctx, "Failed to delete channel cache: %v", err)
	}

	return nil
}

// List lists channels with pagination and filters
func (r *channelRepository) List(ctx context.Context, params *structs.ListChannelParams) ([]*ent.Channel, error) {
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
				channelEnt.Or(
					channelEnt.CreatedAtGT(timestamp),
					channelEnt.And(
						channelEnt.CreatedAtEQ(timestamp),
						channelEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				channelEnt.Or(
					channelEnt.CreatedAtLT(timestamp),
					channelEnt.And(
						channelEnt.CreatedAtEQ(timestamp),
						channelEnt.IDLT(id),
					),
				),
			)
		}
	}

	if params.Direction == "backward" {
		builder.Order(ent.Asc(channelEnt.FieldCreatedAt), ent.Asc(channelEnt.FieldID))
	} else {
		builder.Order(ent.Desc(channelEnt.FieldCreatedAt), ent.Desc(channelEnt.FieldID))
	}

	builder.Limit(params.Limit)

	rows, err := builder.All(ctx)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// Count returns the total count of channels with filters
func (r *channelRepository) Count(ctx context.Context, params *structs.ListChannelParams) (int, error) {
	builder, err := r.buildQuery(ctx, params)
	if validator.IsNotNil(err) {
		return 0, err
	}
	return builder.Count(ctx)
}

// CountX gets a count of channels.
func (r *channelRepository) CountX(ctx context.Context, params *structs.ListChannelParams) int {
	builder, err := r.buildQuery(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// CreateBatch creates multiple channels in a transaction
func (r *channelRepository) CreateBatch(ctx context.Context, channels []*ent.ChannelCreate) ([]*ent.Channel, error) {
	var results []*ent.Channel

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

	for _, c := range channels {
		channel, err := c.Save(ctx)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		results = append(results, channel)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return results, nil
}

// DeleteBatch deletes multiple channels in a transaction
func (r *channelRepository) DeleteBatch(ctx context.Context, ids []string) error {
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

	// Delete subscriptions first
	_, err = tx.Subscription.Delete().
		Where(subscriptionEnt.ChannelIDIn(ids...)).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Then delete channels
	_, err = tx.Channel.Delete().
		Where(channelEnt.IDIn(ids...)).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// UpdateStatus updates a channel's status
func (r *channelRepository) UpdateStatus(ctx context.Context, id string, status int) error {
	err := r.ec.Channel.UpdateOneID(id).
		SetStatus(status).
		Exec(ctx)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("channel:%s", id)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		log.Warnf(ctx, "Failed to delete channel cache: %v", err)
	}

	return nil
}

// UpdateStatusBatch updates status for multiple channels
func (r *channelRepository) UpdateStatusBatch(ctx context.Context, ids []string, status int) error {
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

	_, err = tx.Channel.Update().
		Where(channelEnt.IDIn(ids...)).
		SetStatus(status).
		Save(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// buildQuery creates list builder.
func (r *channelRepository) buildQuery(ctx context.Context, params *structs.ListChannelParams) (*ent.ChannelQuery, error) {
	// create builder.
	builder := r.ec.Channel.Query()
	return builder, nil
}
