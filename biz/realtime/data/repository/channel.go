package repository

import (
	"context"
	"fmt"
	"ncobase/realtime/data"
	"ncobase/realtime/data/ent"
	channelEnt "ncobase/realtime/data/ent/rtchannel"
	subscriptionEnt "ncobase/realtime/data/ent/subscription"
	"ncobase/realtime/structs"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/redis/go-redis/v9"

	"github.com/ncobase/ncore/data/search"
)

// ChannelRepositoryInterface defines channel repository operations
type ChannelRepositoryInterface interface {
	Create(ctx context.Context, channel *ent.RTChannelCreate) (*ent.RTChannel, error)
	Get(ctx context.Context, id string) (*ent.RTChannel, error)
	Update(ctx context.Context, id string, channel *ent.RTChannelUpdateOne) (*ent.RTChannel, error)
	Delete(ctx context.Context, id string) error

	FindByID(ctx context.Context, id string) (*ent.RTChannel, error)
	FindByName(ctx context.Context, name string) (*ent.RTChannel, error)
	List(ctx context.Context, params *structs.ListChannelParams) ([]*ent.RTChannel, error)
	Count(ctx context.Context, params *structs.ListChannelParams) (int, error)
	CountX(ctx context.Context, params *structs.ListChannelParams) int

	CreateBatch(ctx context.Context, channels []*ent.RTChannelCreate) ([]*ent.RTChannel, error)
	DeleteBatch(ctx context.Context, ids []string) error

	UpdateStatus(ctx context.Context, id string, status int) error
	UpdateStatusBatch(ctx context.Context, ids []string, status int) error
}

type channelRepository struct {
	data *data.Data
	ec   *ent.Client
	rc   *redis.Client
	c    *cache.Cache[ent.RTChannel]
}

func NewChannelRepository(d *data.Data) ChannelRepositoryInterface {
	return &channelRepository{
		data: d,
		ec:   d.GetMasterEntClient(),
		rc:   d.GetRedis(),
		c:    cache.NewCache[ent.RTChannel](d.GetRedis(), "rt_channel"),
	}
}

// Create creates a new channel
func (r *channelRepository) Create(ctx context.Context, channel *ent.RTChannelCreate) (*ent.RTChannel, error) {
	row, err := channel.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "channelRepo.Create error: %v", err)
		return nil, err
	}

	// Index in Meilisearch
	if err = r.data.IndexDocument(ctx, &search.IndexRequest{Index: "realtime_channels", Document: row}); err != nil {
		logger.Errorf(ctx, "channelRepo.Create error creating Meilisearch index: %v", err)
	}

	return row, nil
}

// Get gets a channel by ID with cache
func (r *channelRepository) Get(ctx context.Context, id string) (*ent.RTChannel, error) {
	cacheKey := fmt.Sprintf("channel:%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	row, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := r.c.Set(ctx, cacheKey, row); err != nil {
		logger.Warnf(ctx, "Failed to set channel cache: %v", err)
	}

	return row, nil
}

// FindByID finds a channel by ID
func (r *channelRepository) FindByID(ctx context.Context, id string) (*ent.RTChannel, error) {
	return r.ec.RTChannel.Query().
		Where(channelEnt.ID(id)).
		Only(ctx)
}

// FindByName finds a channel by name
func (r *channelRepository) FindByName(ctx context.Context, name string) (*ent.RTChannel, error) {
	return r.ec.RTChannel.Query().
		Where(channelEnt.Name(name)).
		Only(ctx)
}

// Update updates a channel
func (r *channelRepository) Update(ctx context.Context, id string, channel *ent.RTChannelUpdateOne) (*ent.RTChannel, error) {
	row, err := channel.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "channelRepo.Update error: %v", err)
		return nil, err
	}

	// Update Meilisearch index
	if err = r.data.IndexDocument(ctx, &search.IndexRequest{Index: "realtime_channels", Document: row, DocumentID: row.ID}); err != nil {
		logger.Errorf(ctx, "channelRepo.Update error updating Meilisearch index: %v", err)
	}

	cacheKey := fmt.Sprintf("channel:%s", id)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		logger.Warnf(ctx, "Failed to delete channel cache: %v", err)
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
	err = tx.RTChannel.DeleteOneID(id).Exec(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	// Delete from Meilisearch
	if err = r.data.DeleteDocument(ctx, "realtime_channels", id); err != nil {
		logger.Errorf(ctx, "channelRepo.Delete error deleting Meilisearch index: %v", err)
	}

	// Clear cache
	cacheKey := fmt.Sprintf("channel:%s", id)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		logger.Warnf(ctx, "Failed to delete channel cache: %v", err)
	}

	return nil
}

// List lists channels and filters
func (r *channelRepository) List(ctx context.Context, params *structs.ListChannelParams) ([]*ent.RTChannel, error) {
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
func (r *channelRepository) CreateBatch(ctx context.Context, channels []*ent.RTChannelCreate) ([]*ent.RTChannel, error) {
	var results []*ent.RTChannel

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
	_, err = tx.RTChannel.Delete().
		Where(channelEnt.IDIn(ids...)).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// Delete from Meilisearch
	for _, id := range ids {
		if msErr := r.data.DeleteDocument(ctx, "realtime_channels", id); msErr != nil {
			logger.Errorf(ctx, "channelRepo.DeleteBatch error deleting Meilisearch index: %v", msErr)
		}
	}

	return nil
}

// UpdateStatus updates a channel's status
func (r *channelRepository) UpdateStatus(ctx context.Context, id string, status int) error {
	row, err := r.ec.RTChannel.UpdateOneID(id).
		SetStatus(status).
		Save(ctx)
	if err != nil {
		return err
	}

	if err = r.data.IndexDocument(ctx, &search.IndexRequest{Index: "realtime_channels", Document: row, DocumentID: row.ID}); err != nil {
		logger.Errorf(ctx, "channelRepo.UpdateStatus error updating Meilisearch index: %v", err)
	}

	cacheKey := fmt.Sprintf("channel:%s", id)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		logger.Warnf(ctx, "Failed to delete channel cache: %v", err)
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

	_, err = tx.RTChannel.Update().
		Where(channelEnt.IDIn(ids...)).
		SetStatus(status).
		Save(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	for _, id := range ids {
		if row, err := r.ec.RTChannel.Get(ctx, id); err == nil && row != nil {
			if msErr := r.data.IndexDocument(ctx, &search.IndexRequest{Index: "realtime_channels", Document: row, DocumentID: row.ID}); msErr != nil {
				logger.Errorf(ctx, "channelRepo.UpdateStatusBatch error updating Meilisearch index: %v", msErr)
			}
		}
	}

	return nil
}

// buildQuery creates list builder.
func (r *channelRepository) buildQuery(ctx context.Context, params *structs.ListChannelParams) (*ent.RTChannelQuery, error) {
	// create builder.
	builder := r.ec.RTChannel.Query()
	return builder, nil
}
