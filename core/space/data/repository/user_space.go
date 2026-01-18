package repository

import (
	"context"
	"fmt"
	"ncobase/core/space/data"
	"ncobase/core/space/data/ent"
	spaceEnt "ncobase/core/space/data/ent/space"
	userSpaceEnt "ncobase/core/space/data/ent/userspace"
	"ncobase/core/space/structs"
	"time"

	"github.com/ncobase/ncore/data/cache"
	"github.com/ncobase/ncore/logging/logger"

	"github.com/redis/go-redis/v9"
)

// UserSpaceRepositoryInterface represents the user space repository interface.
type UserSpaceRepositoryInterface interface {
	Create(ctx context.Context, body *structs.UserSpace) (*ent.UserSpace, error)
	GetByUserID(ctx context.Context, id string) (*ent.UserSpace, error)
	GetBySpaceID(ctx context.Context, id string) (*ent.UserSpace, error)
	GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserSpace, error)
	GetBySpaceIDs(ctx context.Context, ids []string) ([]*ent.UserSpace, error)
	Delete(ctx context.Context, uid, did string) error
	DeleteAllByUserID(ctx context.Context, id string) error
	DeleteAllBySpaceID(ctx context.Context, id string) error
	GetSpacesByUserID(ctx context.Context, userID string) ([]*ent.Space, error)
	IsSpaceInUser(ctx context.Context, spaceID, userID string) (bool, error)
}

// userSpaceRepository implements the UserSpaceRepositoryInterface.
type userSpaceRepository struct {
	data            *data.Data
	userSpaceCache  cache.ICache[ent.UserSpace]
	userSpacesCache cache.ICache[[]string] // Maps user ID to space IDs
	spaceUsersCache cache.ICache[[]string] // Maps space ID to user IDs
	relationshipTTL time.Duration
}

// NewUserSpaceRepository creates a new user space repository.
func NewUserSpaceRepository(d *data.Data) UserSpaceRepositoryInterface {
	redisClient := d.GetRedis().(*redis.Client)

	return &userSpaceRepository{
		data:            d,
		userSpaceCache:  cache.NewCache[ent.UserSpace](redisClient, "ncse_space:user_spaces"),
		userSpacesCache: cache.NewCache[[]string](redisClient, "ncse_space:user_space_mappings"),
		spaceUsersCache: cache.NewCache[[]string](redisClient, "ncse_space:space_user_mappings"),
		relationshipTTL: time.Hour * 3, // 3 hours cache TTL (space relationships change less frequently)
	}
}

// Create creates a new user space
func (r *userSpaceRepository) Create(ctx context.Context, body *structs.UserSpace) (*ent.UserSpace, error) {
	// Use master for writes
	builder := r.data.GetMasterEntClient().UserSpace.Create()

	// Set values
	builder.SetNillableUserID(&body.UserID)
	builder.SetNillableSpaceID(&body.SpaceID)

	// Execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "userSpaceRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the relationship and invalidate related caches
	go func() {
		r.cacheUserSpace(context.Background(), row)
		r.invalidateUserSpacesCache(context.Background(), body.UserID)
		r.invalidateSpaceUsersCache(context.Background(), body.SpaceID)
	}()

	return row, nil
}

// GetByUserID find space by user id
func (r *userSpaceRepository) GetByUserID(ctx context.Context, id string) (*ent.UserSpace, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("user:%s", id)
	if cached, err := r.userSpaceCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserSpace.Query()

	// Set conditions
	builder.Where(userSpaceEnt.UserIDEQ(id))

	// Execute the builder
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "userSpaceRepo.GetByUserID error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheUserSpace(context.Background(), row)

	return row, nil
}

// GetByUserIDs find spaces by user ids
func (r *userSpaceRepository) GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserSpace, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserSpace.Query()

	// Set conditions
	builder.Where(userSpaceEnt.UserIDIn(ids...))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userSpaceRepo.GetByUserIDs error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, ut := range rows {
			r.cacheUserSpace(context.Background(), ut)
		}
	}()

	return rows, nil
}

// GetBySpaceID find space by space id
func (r *userSpaceRepository) GetBySpaceID(ctx context.Context, id string) (*ent.UserSpace, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("space:%s", id)
	if cached, err := r.userSpaceCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserSpace.Query()

	// Set conditions
	builder.Where(userSpaceEnt.SpaceIDEQ(id))

	// Execute the builder
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "userSpaceRepo.GetBySpaceID error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheUserSpace(context.Background(), row)

	return row, nil
}

// GetBySpaceIDs find spaces by space ids
func (r *userSpaceRepository) GetBySpaceIDs(ctx context.Context, ids []string) ([]*ent.UserSpace, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserSpace.Query()

	// Set conditions
	builder.Where(userSpaceEnt.SpaceIDIn(ids...))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userSpaceRepo.GetBySpaceIDs error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, ut := range rows {
			r.cacheUserSpace(context.Background(), ut)
		}
	}()

	return rows, nil
}

// Delete delete user space
func (r *userSpaceRepository) Delete(ctx context.Context, uid, did string) error {
	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserSpace.Delete().
		Where(userSpaceEnt.UserIDEQ(uid), userSpaceEnt.SpaceIDEQ(did)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userSpaceRepo.Delete error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateUserSpaceCache(context.Background(), uid, did)
		r.invalidateUserSpacesCache(context.Background(), uid)
		r.invalidateSpaceUsersCache(context.Background(), did)
	}()

	return nil
}

// DeleteAllByUserID delete all user space
func (r *userSpaceRepository) DeleteAllByUserID(ctx context.Context, id string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.GetByUserIDs(ctx, []string{id})
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserSpace.Delete().
		Where(userSpaceEnt.UserIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userSpaceRepo.DeleteAllByUserID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateUserSpacesCache(context.Background(), id)
		for _, ut := range relationships {
			r.invalidateUserSpaceCache(context.Background(), ut.UserID, ut.SpaceID)
			r.invalidateSpaceUsersCache(context.Background(), ut.SpaceID)
		}
	}()

	return nil
}

// DeleteAllBySpaceID delete all user space
func (r *userSpaceRepository) DeleteAllBySpaceID(ctx context.Context, id string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.GetBySpaceIDs(ctx, []string{id})
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserSpace.Delete().
		Where(userSpaceEnt.SpaceIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userSpaceRepo.DeleteAllBySpaceID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateSpaceUsersCache(context.Background(), id)
		for _, ut := range relationships {
			r.invalidateUserSpaceCache(context.Background(), ut.UserID, ut.SpaceID)
			r.invalidateUserSpacesCache(context.Background(), ut.UserID)
		}
	}()

	return nil
}

// GetSpacesByUserID retrieves all spaces a user belongs to.
func (r *userSpaceRepository) GetSpacesByUserID(ctx context.Context, userID string) ([]*ent.Space, error) {
	// Try to get space IDs from cache
	cacheKey := fmt.Sprintf("user_spaces:%s", userID)
	var spaceIDs []string
	if err := r.userSpacesCache.GetArray(ctx, cacheKey, &spaceIDs); err == nil && len(spaceIDs) > 0 {
		// Get spaces by IDs from space repository
		return r.data.GetSlaveEntClient().Space.Query().Where(spaceEnt.IDIn(spaceIDs...)).All(ctx)
	}

	// Fallback to database
	userSpaces, err := r.data.GetSlaveEntClient().UserSpace.Query().
		Where(userSpaceEnt.UserIDEQ(userID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userSpaceRepo.GetSpacesByUserID error: %v", err)
		return nil, err
	}

	// Extract space IDs from userSpaces
	spaceIDs = make([]string, len(userSpaces))
	for i, userSpace := range userSpaces {
		spaceIDs[i] = userSpace.SpaceID
	}

	// Query spaces based on extracted space IDs
	spaces, err := r.data.GetSlaveEntClient().Space.Query().Where(spaceEnt.IDIn(spaceIDs...)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userSpaceRepo.GetSpacesByUserID error: %v", err)
		return nil, err
	}

	// Cache space IDs for future use
	go func() {
		if err := r.userSpacesCache.SetArray(context.Background(), cacheKey, spaceIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache user spaces %s: %v", userID, err)
		}
	}()

	return spaces, nil
}

// IsUserInSpace verifies if a user belongs to a specific space.
func (r *userSpaceRepository) IsUserInSpace(ctx context.Context, userID string, spaceID string) (bool, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("relationship:%s:%s", userID, spaceID)
	if cached, err := r.userSpaceCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return true, nil
	}

	// Use slave for reads
	count, err := r.data.GetSlaveEntClient().UserSpace.Query().
		Where(userSpaceEnt.UserIDEQ(userID), userSpaceEnt.SpaceIDEQ(spaceID)).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "userSpaceRepo.IsUserInSpace error: %v", err)
		return false, err
	}

	exists := count > 0

	// Cache the result if it exists
	if exists {
		go func() {
			// Create a dummy relationship for caching
			relationship := &ent.UserSpace{
				UserID:  userID,
				SpaceID: spaceID,
			}
			r.cacheUserSpace(context.Background(), relationship)
		}()
	}

	return exists, nil
}

// IsSpaceInUser verifies if a space is assigned to a specific user.
func (r *userSpaceRepository) IsSpaceInUser(ctx context.Context, spaceID, userID string) (bool, error) {
	return r.IsUserInSpace(ctx, userID, spaceID)
}

// cacheUserSpace caches a user-space relationship.
func (r *userSpaceRepository) cacheUserSpace(ctx context.Context, ut *ent.UserSpace) {
	// Cache by relationship key
	relationshipKey := fmt.Sprintf("relationship:%s:%s", ut.UserID, ut.SpaceID)
	if err := r.userSpaceCache.Set(ctx, relationshipKey, ut, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache user space relationship %s:%s: %v", ut.UserID, ut.SpaceID, err)
	}

	// Cache by user ID
	userKey := fmt.Sprintf("user:%s", ut.UserID)
	if err := r.userSpaceCache.Set(ctx, userKey, ut, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache user space by user %s: %v", ut.UserID, err)
	}

	// Cache by space ID
	spaceKey := fmt.Sprintf("space:%s", ut.SpaceID)
	if err := r.userSpaceCache.Set(ctx, spaceKey, ut, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache user space by space %s: %v", ut.SpaceID, err)
	}
}

// invalidateUserSpaceCache invalidates the cache for a user-space relationship.
func (r *userSpaceRepository) invalidateUserSpaceCache(ctx context.Context, userID, spaceID string) {
	// Invalidate relationship cache
	relationshipKey := fmt.Sprintf("relationship:%s:%s", userID, spaceID)
	if err := r.userSpaceCache.Delete(ctx, relationshipKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user space relationship cache %s:%s: %v", userID, spaceID, err)
	}

	// Invalidate user key
	userKey := fmt.Sprintf("user:%s", userID)
	if err := r.userSpaceCache.Delete(ctx, userKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user space cache by user %s: %v", userID, err)
	}

	// Invalidate space key
	spaceKey := fmt.Sprintf("space:%s", spaceID)
	if err := r.userSpaceCache.Delete(ctx, spaceKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user space cache by space %s: %v", spaceID, err)
	}
}

// invalidateUserSpacesCache invalidates the cache for all user spaces.
func (r *userSpaceRepository) invalidateUserSpacesCache(ctx context.Context, userID string) {
	cacheKey := fmt.Sprintf("user_spaces:%s", userID)
	if err := r.userSpacesCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user spaces cache %s: %v", userID, err)
	}
}

// invalidateSpaceUsersCache invalidates the cache for all space users.
func (r *userSpaceRepository) invalidateSpaceUsersCache(ctx context.Context, spaceID string) {
	cacheKey := fmt.Sprintf("space_users:%s", spaceID)
	if err := r.spaceUsersCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate space users cache %s: %v", spaceID, err)
	}
}
