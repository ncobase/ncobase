package repository

import (
	"context"
	"fmt"
	"ncobase/biz/realtime/data"
	"ncobase/biz/realtime/data/ent"
	subscriptionEnt "ncobase/biz/realtime/data/ent/subscription"

	nd "github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/data/cache"
	"github.com/ncobase/ncore/logging/logger"

	"github.com/redis/go-redis/v9"

	"github.com/ncobase/ncore/data/search"
)

// SubscriptionRepositoryInterface defines subscription repository operations
type SubscriptionRepositoryInterface interface {
	Create(ctx context.Context, subscription *ent.SubscriptionCreate) (*ent.Subscription, error)
	Get(ctx context.Context, id string) (*ent.Subscription, error)
	Update(ctx context.Context, id string, subscription *ent.SubscriptionUpdateOne) (*ent.Subscription, error)
	Delete(ctx context.Context, id string) error

	FindByID(ctx context.Context, id string) (*ent.Subscription, error)
	List(ctx context.Context, offset, limit int, filters map[string]any) ([]*ent.Subscription, error)
	Count(ctx context.Context, filters map[string]any) (int, error)

	CreateBatch(ctx context.Context, subscriptions []*ent.SubscriptionCreate) ([]*ent.Subscription, error)
	DeleteBatch(ctx context.Context, ids []string) error

	FindByUserAndChannel(ctx context.Context, userID, channelID string) (*ent.Subscription, error)
	GetUserSubscriptions(ctx context.Context, userID string) ([]*ent.Subscription, error)
	GetChannelSubscribers(ctx context.Context, channelID string) ([]*ent.Subscription, error)
	UpdateStatus(ctx context.Context, id string, status int) error
	UpdateStatusBatch(ctx context.Context, ids []string, status int) error
	DeleteByUserAndChannel(ctx context.Context, userID, channelID string) error
	DeleteByChannel(ctx context.Context, channelID string) error
	DeleteByUser(ctx context.Context, userID string) error
}

type subscriptionRepository struct {
	data         *data.Data
	searchClient *search.Client
	ec           *ent.Client
	rc           *redis.Client
	c            *cache.Cache[ent.Subscription]
}

func NewSubscriptionRepository(d *data.Data) SubscriptionRepositoryInterface {
	redisClient := d.GetRedis().(*redis.Client)
	searchClient := nd.NewSearchClient(d.Data)

	return &subscriptionRepository{
		data:         d,
		searchClient: searchClient,
		ec:           d.GetMasterEntClient(),
		rc:           redisClient,
		c:            cache.NewCache[ent.Subscription](redisClient, "rt_subscription"),
	}
}

// Create creates a new subscription
func (r *subscriptionRepository) Create(ctx context.Context, subscription *ent.SubscriptionCreate) (*ent.Subscription, error) {
	// Check if subscription already exists

	userID, _ := subscription.Mutation().UserID()
	channelID, _ := subscription.Mutation().ChannelID()

	exists, err := r.ec.Subscription.Query().
		Where(
			subscriptionEnt.UserID(userID),
			subscriptionEnt.ChannelID(channelID),
		).
		Exist(ctx)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("subscription already exists")
	}

	row, err := subscription.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "subscriptionRepo.Create error: %v", err)
		return nil, err
	}

	// Index in Meilisearch
	if err = r.searchClient.Index(ctx, &search.IndexRequest{Index: "realtime_subscriptions", Document: row}); err != nil {
		logger.Errorf(ctx, "subscriptionRepo.Create error creating Meilisearch index: %v", err)
	}

	return row, nil
}

// Get gets a subscription by ID with cache
func (r *subscriptionRepository) Get(ctx context.Context, id string) (*ent.Subscription, error) {
	cacheKey := fmt.Sprintf("subscription:%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	row, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := r.c.Set(ctx, cacheKey, row); err != nil {
		logger.Warnf(ctx, "Failed to set subscription cache: %v", err)
	}

	return row, nil
}

// Update updates a subscription
func (r *subscriptionRepository) Update(ctx context.Context, id string, subscription *ent.SubscriptionUpdateOne) (*ent.Subscription, error) {
	row, err := subscription.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "subscriptionRepo.Update error: %v", err)
		return nil, err
	}

	// Update Meilisearch index
	if err = r.searchClient.Index(ctx, &search.IndexRequest{Index: "realtime_subscriptions", Document: row, DocumentID: row.ID}); err != nil {
		logger.Errorf(ctx, "subscriptionRepo.Update error updating Meilisearch index: %v", err)
	}

	cacheKey := fmt.Sprintf("subscription:%s", id)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		logger.Warnf(ctx, "Failed to delete subscription cache: %v", err)
	}

	return row, nil
}

// Delete deletes a subscription
func (r *subscriptionRepository) Delete(ctx context.Context, id string) error {
	err := r.ec.Subscription.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return err
	}

	// Delete from Meilisearch
	if err = r.searchClient.Delete(ctx, "realtime_subscriptions", id); err != nil {
		logger.Errorf(ctx, "subscriptionRepo.Delete error deleting Meilisearch index: %v", err)
	}

	cacheKey := fmt.Sprintf("subscription:%s", id)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		logger.Warnf(ctx, "Failed to delete subscription cache: %v", err)
	}

	return nil
}

// FindByID finds a subscription by ID
func (r *subscriptionRepository) FindByID(ctx context.Context, id string) (*ent.Subscription, error) {
	return r.ec.Subscription.Query().
		Where(subscriptionEnt.ID(id)).
		Only(ctx)
}

// FindByUserAndChannel finds a subscription by user ID and channel ID
func (r *subscriptionRepository) FindByUserAndChannel(ctx context.Context, userID, channelID string) (*ent.Subscription, error) {
	return r.ec.Subscription.Query().
		Where(
			subscriptionEnt.UserID(userID),
			subscriptionEnt.ChannelID(channelID),
		).
		Only(ctx)
}

// List lists subscriptions and filters
func (r *subscriptionRepository) List(ctx context.Context, offset, limit int, filters map[string]any) ([]*ent.Subscription, error) {
	query := r.buildQuery(filters)

	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	return query.Order(ent.Desc(subscriptionEnt.FieldCreatedAt)).All(ctx)
}

// Count returns the total count of subscriptions with filters
func (r *subscriptionRepository) Count(ctx context.Context, filters map[string]any) (int, error) {
	return r.buildQuery(filters).Count(ctx)
}

// CreateBatch creates multiple subscriptions in a transaction
func (r *subscriptionRepository) CreateBatch(ctx context.Context, subscriptions []*ent.SubscriptionCreate) ([]*ent.Subscription, error) {
	var results []*ent.Subscription

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

	for _, s := range subscriptions {
		subscription, err := s.Save(ctx)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		results = append(results, subscription)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	for _, row := range results {
		if msErr := r.searchClient.Index(ctx, &search.IndexRequest{Index: "realtime_subscriptions", Document: row}); msErr != nil {
			logger.Errorf(ctx, "subscriptionRepo.CreateBatch error creating Meilisearch index: %v", msErr)
		}
	}

	return results, nil
}

// DeleteBatch deletes multiple subscriptions in a transaction
func (r *subscriptionRepository) DeleteBatch(ctx context.Context, ids []string) error {
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

	_, err = tx.Subscription.Delete().
		Where(subscriptionEnt.IDIn(ids...)).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	for _, id := range ids {
		if msErr := r.searchClient.Delete(ctx, "realtime_subscriptions", id); msErr != nil {
			logger.Errorf(ctx, "subscriptionRepo.DeleteBatch error deleting Meilisearch index: %v", msErr)
		}
	}

	return nil
}

// GetUserSubscriptions gets all subscriptions for a user
func (r *subscriptionRepository) GetUserSubscriptions(ctx context.Context, userID string) ([]*ent.Subscription, error) {
	return r.ec.Subscription.Query().
		Where(
			subscriptionEnt.UserID(userID),
			subscriptionEnt.Status(1), // Only active subscriptions
		).
		All(ctx)
}

// GetChannelSubscribers gets all subscriptions for a channel
func (r *subscriptionRepository) GetChannelSubscribers(ctx context.Context, channelID string) ([]*ent.Subscription, error) {
	return r.ec.Subscription.Query().
		Where(
			subscriptionEnt.ChannelID(channelID),
			subscriptionEnt.Status(1), // Only active subscriptions
		).
		All(ctx)
}

// UpdateStatus updates a subscription's status
func (r *subscriptionRepository) UpdateStatus(ctx context.Context, id string, status int) error {
	row, err := r.ec.Subscription.UpdateOneID(id).
		SetStatus(status).
		Save(ctx)
	if err != nil {
		return err
	}

	if err = r.searchClient.Index(ctx, &search.IndexRequest{Index: "realtime_subscriptions", Document: row, DocumentID: row.ID}); err != nil {
		logger.Errorf(ctx, "subscriptionRepo.UpdateStatus error updating Meilisearch index: %v", err)
	}

	cacheKey := fmt.Sprintf("subscription:%s", id)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		logger.Warnf(ctx, "Failed to delete subscription cache: %v", err)
	}

	return nil
}

// UpdateStatusBatch updates status for multiple subscriptions
func (r *subscriptionRepository) UpdateStatusBatch(ctx context.Context, ids []string, status int) error {
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

	_, err = tx.Subscription.Update().
		Where(subscriptionEnt.IDIn(ids...)).
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
		if row, err := r.ec.Subscription.Get(ctx, id); err == nil && row != nil {
			if msErr := r.searchClient.Index(ctx, &search.IndexRequest{Index: "realtime_subscriptions", Document: row, DocumentID: row.ID}); msErr != nil {
				logger.Errorf(ctx, "subscriptionRepo.UpdateStatusBatch error updating Meilisearch index: %v", msErr)
			}
		}
	}

	return nil
}

// DeleteByUserAndChannel deletes a subscription by user ID and channel ID
func (r *subscriptionRepository) DeleteByUserAndChannel(ctx context.Context, userID, channelID string) error {
	var targetID string
	if row, err := r.ec.Subscription.Query().
		Where(
			subscriptionEnt.UserID(userID),
			subscriptionEnt.ChannelID(channelID),
		).First(ctx); err == nil && row != nil {
		targetID = row.ID
	}

	_, err := r.ec.Subscription.Delete().
		Where(
			subscriptionEnt.UserID(userID),
			subscriptionEnt.ChannelID(channelID),
		).
		Exec(ctx)
	if err != nil {
		return err
	}

	if targetID != "" {
		if err = r.searchClient.Delete(ctx, "realtime_subscriptions", targetID); err != nil {
			logger.Errorf(ctx, "subscriptionRepo.DeleteByUserAndChannel error deleting Meilisearch index: %v", err)
		}
	}

	return nil
}

// DeleteByChannel deletes all subscriptions for a channel
func (r *subscriptionRepository) DeleteByChannel(ctx context.Context, channelID string) error {
	rows, _ := r.ec.Subscription.Query().Where(subscriptionEnt.ChannelID(channelID)).All(ctx)
	_, err := r.ec.Subscription.Delete().
		Where(subscriptionEnt.ChannelID(channelID)).
		Exec(ctx)
	if err != nil {
		return err
	}

	for _, row := range rows {
		if msErr := r.searchClient.Delete(ctx, "realtime_subscriptions", row.ID); msErr != nil {
			logger.Errorf(ctx, "subscriptionRepo.DeleteByChannel error deleting Meilisearch index: %v", msErr)
		}
	}

	return nil
}

// DeleteByUser deletes all subscriptions for a user
func (r *subscriptionRepository) DeleteByUser(ctx context.Context, userID string) error {
	rows, _ := r.ec.Subscription.Query().Where(subscriptionEnt.UserID(userID)).All(ctx)
	_, err := r.ec.Subscription.Delete().
		Where(subscriptionEnt.UserID(userID)).
		Exec(ctx)
	if err != nil {
		return err
	}

	for _, row := range rows {
		if msErr := r.searchClient.Delete(ctx, "realtime_subscriptions", row.ID); msErr != nil {
			logger.Errorf(ctx, "subscriptionRepo.DeleteByUser error deleting Meilisearch index: %v", msErr)
		}
	}

	return nil
}

// buildQuery builds a query with filters
func (r *subscriptionRepository) buildQuery(filters map[string]any) *ent.SubscriptionQuery {
	query := r.ec.Subscription.Query()

	for key, value := range filters {
		switch key {
		case "user_id":
			if userID, ok := value.(string); ok {
				query = query.Where(subscriptionEnt.UserID(userID))
			}
		case "channel_id":
			if channelID, ok := value.(string); ok {
				query = query.Where(subscriptionEnt.ChannelID(channelID))
			}
		case "status":
			if status, ok := value.(int); ok {
				query = query.Where(subscriptionEnt.Status(status))
			}
		}
	}

	return query
}
