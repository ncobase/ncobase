package repository

import (
	"context"
	"fmt"
	"ncobase/core/user/data"
	"ncobase/core/user/data/ent"
	apiKeyEnt "ncobase/core/user/data/ent/apikey"
	"ncobase/core/user/structs"
	"time"
	"github.com/redis/go-redis/v9"

	"github.com/ncobase/ncore/data/cache"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/security/crypto"
	"github.com/ncobase/ncore/utils/nanoid"
)

// ApiKeyRepositoryInterface defines API key repository operations
type ApiKeyRepositoryInterface interface {
	Create(ctx context.Context, userID string, request *structs.CreateApiKeyRequest) (*ent.ApiKey, string, error)
	GetByID(ctx context.Context, id string) (*ent.ApiKey, error)
	GetByKey(ctx context.Context, key string) (*ent.ApiKey, error)
	GetByUserID(ctx context.Context, userID string) ([]*ent.ApiKey, error)
	Delete(ctx context.Context, id string) error
	UpdateLastUsed(ctx context.Context, id string, timestamp int64) error
}

// apiKeyRepository implements ApiKeyRepositoryInterface
type apiKeyRepository struct {
	data            *data.Data
	apiKeyCache     cache.ICache[ent.ApiKey]
	keyMappingCache cache.ICache[string]   // Maps hashed key to API key ID
	userKeysCache   cache.ICache[[]string] // Maps user ID to API key IDs
	apiKeyTTL       time.Duration
}

// NewApiKeyRepository creates a new API key repository
func NewApiKeyRepository(d *data.Data) ApiKeyRepositoryInterface {
	redisClient := d.GetRedis().(*redis.Client)

	return &apiKeyRepository{
		data:            d,
		apiKeyCache:     cache.NewCache[ent.ApiKey](redisClient, "ncse_api_keys"),
		keyMappingCache: cache.NewCache[string](redisClient, "ncse_api_key_mappings"),
		userKeysCache:   cache.NewCache[[]string](redisClient, "ncse_user_api_keys"),
		apiKeyTTL:       time.Hour * 6, // 6 hours cache TTL
	}
}

// Create creates a new API key
func (r *apiKeyRepository) Create(ctx context.Context, userID string, request *structs.CreateApiKeyRequest) (*ent.ApiKey, string, error) {
	id := nanoid.PrimaryKey()()
	now := time.Now().UnixMilli()

	// Generate API key
	apiKeyValue := nanoid.String(32)
	hashedKey, err := crypto.HashPassword(ctx, apiKeyValue)
	if err != nil {
		return nil, "", err
	}

	client := r.data.GetMasterEntClient()
	apiKey, err := client.ApiKey.Create().
		SetID(id).
		SetName(request.Name).
		SetKey(hashedKey).
		SetUserID(userID).
		SetCreatedAt(now).
		SetLastUsed(now).
		Save(ctx)

	if err != nil {
		return nil, "", err
	}

	// Cache the API key and invalidate user keys cache
	go func() {
		r.cacheApiKey(context.Background(), apiKey)
		r.cacheKeyMapping(context.Background(), hashedKey, apiKey.ID)
		r.invalidateUserKeysCache(context.Background(), userID)
	}()

	return apiKey, apiKeyValue, nil
}

// GetByID retrieves an API key by ID
func (r *apiKeyRepository) GetByID(ctx context.Context, id string) (*ent.ApiKey, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("id:%s", id)
	if cached, err := r.apiKeyCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Fallback to database
	client := r.data.GetSlaveEntClient()
	apiKey, err := client.ApiKey.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache for future use
	go r.cacheApiKey(context.Background(), apiKey)

	return apiKey, nil
}

// GetByKey retrieves an API key by the actual key value
func (r *apiKeyRepository) GetByKey(ctx context.Context, key string) (*ent.ApiKey, error) {
	// First, we need to find which API key this belongs to
	// We'll need to check all API keys (this is expensive, but API key validation should be cached)
	client := r.data.GetSlaveEntClient()
	apiKeys, err := client.ApiKey.Query().All(ctx)
	if err != nil {
		return nil, err
	}

	// Check each API key hash
	for _, apiKey := range apiKeys {
		if crypto.ComparePassword(apiKey.Key, key) {
			// Cache this API key for future use
			go r.cacheApiKey(context.Background(), apiKey)
			return apiKey, nil
		}
	}

	return nil, fmt.Errorf("API key not found")
}

// GetByUserID retrieves all API keys for a user
func (r *apiKeyRepository) GetByUserID(ctx context.Context, userID string) ([]*ent.ApiKey, error) {
	client := r.data.GetSlaveEntClient()
	apiKeys, err := client.ApiKey.Query().
		Where(apiKeyEnt.UserIDEQ(userID)).
		Order(ent.Desc(apiKeyEnt.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		return nil, err
	}

	// Cache API keys in background
	go func() {
		for _, apiKey := range apiKeys {
			r.cacheApiKey(context.Background(), apiKey)
		}
	}()

	return apiKeys, nil
}

// Delete deletes an API key
func (r *apiKeyRepository) Delete(ctx context.Context, id string) error {
	// Get API key first for cache invalidation
	apiKey, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	client := r.data.GetMasterEntClient()
	err = client.ApiKey.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateApiKeyCache(context.Background(), id)
		r.invalidateKeyMappingCache(context.Background(), apiKey.Key)
		r.invalidateUserKeysCache(context.Background(), apiKey.UserID)
	}()

	return nil
}

// UpdateLastUsed updates the last used timestamp
func (r *apiKeyRepository) UpdateLastUsed(ctx context.Context, id string, timestamp int64) error {
	client := r.data.GetMasterEntClient()
	apiKey, err := client.ApiKey.UpdateOneID(id).
		SetLastUsed(timestamp).
		Save(ctx)

	if err != nil {
		return err
	}

	// Update cache
	go func() {
		r.invalidateApiKeyCache(context.Background(), id)
		r.cacheApiKey(context.Background(), apiKey)
	}()

	return nil
}

// cacheApiKey - cache API key.
func (r *apiKeyRepository) cacheApiKey(ctx context.Context, apiKey *ent.ApiKey) {
	cacheKey := fmt.Sprintf("id:%s", apiKey.ID)
	if err := r.apiKeyCache.Set(ctx, cacheKey, apiKey, r.apiKeyTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache API key %s: %v", apiKey.ID, err)
	}
}

// cacheKeyMapping - cache key mapping.
func (r *apiKeyRepository) cacheKeyMapping(ctx context.Context, hashedKey, apiKeyID string) {
	cacheKey := fmt.Sprintf("key:%s", hashedKey)
	if err := r.keyMappingCache.Set(ctx, cacheKey, &apiKeyID, r.apiKeyTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache key mapping %s: %v", hashedKey, err)
	}
}

// invalidateApiKeyCache invalidates API key cache
func (r *apiKeyRepository) invalidateApiKeyCache(ctx context.Context, apiKeyID string) {
	cacheKey := fmt.Sprintf("id:%s", apiKeyID)
	if err := r.apiKeyCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate API key cache %s: %v", apiKeyID, err)
	}
}

// invalidateKeyMappingCache invalidates key mapping cache
func (r *apiKeyRepository) invalidateKeyMappingCache(ctx context.Context, hashedKey string) {
	cacheKey := fmt.Sprintf("key:%s", hashedKey)
	if err := r.keyMappingCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate key mapping cache %s: %v", hashedKey, err)
	}
}

// invalidateUserKeysCache invalidates user keys cache
func (r *apiKeyRepository) invalidateUserKeysCache(ctx context.Context, userID string) {
	cacheKey := fmt.Sprintf("user:%s", userID)
	if err := r.userKeysCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user keys cache %s: %v", userID, err)
	}
}
