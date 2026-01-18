package repository

import (
	"context"
	"fmt"
	"ncobase/core/space/data"
	"ncobase/core/space/data/ent"
	spaceOptionEnt "ncobase/core/space/data/ent/spaceoption"
	"ncobase/core/space/structs"
	"time"
	"github.com/redis/go-redis/v9"

	"github.com/ncobase/ncore/data/cache"
	"github.com/ncobase/ncore/logging/logger"
)

// SpaceOptionRepositoryInterface represents the space option repository interface.
type SpaceOptionRepositoryInterface interface {
	Create(ctx context.Context, body *structs.SpaceOption) (*ent.SpaceOption, error)
	GetBySpaceID(ctx context.Context, spaceID string) ([]*ent.SpaceOption, error)
	GetByOptionID(ctx context.Context, optionsID string) ([]*ent.SpaceOption, error)
	DeleteBySpaceIDAndOptionID(ctx context.Context, spaceID, optionsID string) error
	DeleteAllBySpaceID(ctx context.Context, spaceID string) error
	DeleteAllByOptionID(ctx context.Context, optionsID string) error
	IsOptionsInSpace(ctx context.Context, spaceID, optionsID string) (bool, error)
	GetSpaceOption(ctx context.Context, spaceID string) ([]string, error)
}

// spaceOptionRepository implements the SpaceOptionRepositoryInterface.
type spaceOptionRepository struct {
	data                 *data.Data
	spaceOptionCache     cache.ICache[ent.SpaceOption]
	spaceOptionListCache cache.ICache[[]string] // Maps space to options IDs
	optionsSpacesCache   cache.ICache[[]string] // Maps options to space IDs
	relationshipTTL      time.Duration
}

// NewSpaceOptionRepository creates a new space option repository.
func NewSpaceOptionRepository(d *data.Data) SpaceOptionRepositoryInterface {
	redisClient := d.GetRedis().(*redis.Client)

	return &spaceOptionRepository{
		data:                 d,
		spaceOptionCache:     cache.NewCache[ent.SpaceOption](redisClient, "ncse_space:space_options"),
		spaceOptionListCache: cache.NewCache[[]string](redisClient, "ncse_space:space_options_mappings"),
		optionsSpacesCache:   cache.NewCache[[]string](redisClient, "ncse_space:options_space_mappings"),
		relationshipTTL:      time.Hour * 2,
	}
}

// Create creates a new space option relationship.
func (r *spaceOptionRepository) Create(ctx context.Context, body *structs.SpaceOption) (*ent.SpaceOption, error) {
	builder := r.data.GetMasterEntClient().SpaceOption.Create()

	builder.SetNillableSpaceID(&body.SpaceID)
	builder.SetNillableOptionID(&body.OptionID)

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceOptionRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the relationship and invalidate related caches
	go func() {
		r.cacheSpaceOption(context.Background(), row)
		r.invalidateSpaceOptionListCache(context.Background(), body.SpaceID)
		r.invalidateOptionsSpacesCache(context.Background(), body.OptionID)
	}()

	return row, nil
}

// GetBySpaceID retrieves space option by space ID.
func (r *spaceOptionRepository) GetBySpaceID(ctx context.Context, spaceID string) ([]*ent.SpaceOption, error) {
	builder := r.data.GetSlaveEntClient().SpaceOption.Query()
	builder.Where(spaceOptionEnt.SpaceIDEQ(spaceID))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceOptionRepo.GetBySpaceID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, to := range rows {
			r.cacheSpaceOption(context.Background(), to)
		}
	}()

	return rows, nil
}

// GetByOptionID retrieves space option by options ID.
func (r *spaceOptionRepository) GetByOptionID(ctx context.Context, optionsID string) ([]*ent.SpaceOption, error) {
	builder := r.data.GetSlaveEntClient().SpaceOption.Query()
	builder.Where(spaceOptionEnt.OptionIDEQ(optionsID))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceOptionRepo.GetByOptionID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, to := range rows {
			r.cacheSpaceOption(context.Background(), to)
		}
	}()

	return rows, nil
}

// DeleteBySpaceIDAndOptionID deletes space option by space ID and options ID.
func (r *spaceOptionRepository) DeleteBySpaceIDAndOptionID(ctx context.Context, spaceID, optionsID string) error {
	if _, err := r.data.GetMasterEntClient().SpaceOption.Delete().
		Where(spaceOptionEnt.SpaceIDEQ(spaceID), spaceOptionEnt.OptionIDEQ(optionsID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "spaceOptionRepo.DeleteBySpaceIDAndOptionID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateSpaceOptionCache(context.Background(), spaceID, optionsID)
		r.invalidateSpaceOptionListCache(context.Background(), spaceID)
		r.invalidateOptionsSpacesCache(context.Background(), optionsID)
	}()

	return nil
}

// DeleteAllBySpaceID deletes all space option by space ID.
func (r *spaceOptionRepository) DeleteAllBySpaceID(ctx context.Context, spaceID string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().SpaceOption.Query().
		Where(spaceOptionEnt.SpaceIDEQ(spaceID)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	if _, err := r.data.GetMasterEntClient().SpaceOption.Delete().
		Where(spaceOptionEnt.SpaceIDEQ(spaceID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "spaceOptionRepo.DeleteAllBySpaceID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateSpaceOptionListCache(context.Background(), spaceID)
		for _, to := range relationships {
			r.invalidateSpaceOptionCache(context.Background(), to.SpaceID, to.OptionID)
			r.invalidateOptionsSpacesCache(context.Background(), to.OptionID)
		}
	}()

	return nil
}

// DeleteAllByOptionID deletes all space option by options ID.
func (r *spaceOptionRepository) DeleteAllByOptionID(ctx context.Context, optionsID string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().SpaceOption.Query().
		Where(spaceOptionEnt.OptionIDEQ(optionsID)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	if _, err := r.data.GetMasterEntClient().SpaceOption.Delete().
		Where(spaceOptionEnt.OptionIDEQ(optionsID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "spaceOptionRepo.DeleteAllByOptionID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateOptionsSpacesCache(context.Background(), optionsID)
		for _, to := range relationships {
			r.invalidateSpaceOptionCache(context.Background(), to.SpaceID, to.OptionID)
			r.invalidateSpaceOptionListCache(context.Background(), to.SpaceID)
		}
	}()

	return nil
}

// IsOptionsInSpace verifies if an options belongs to a space.
func (r *spaceOptionRepository) IsOptionsInSpace(ctx context.Context, spaceID, optionsID string) (bool, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("relationship:%s:%s", spaceID, optionsID)
	if cached, err := r.spaceOptionCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return true, nil
	}

	count, err := r.data.GetSlaveEntClient().SpaceOption.Query().
		Where(spaceOptionEnt.SpaceIDEQ(spaceID), spaceOptionEnt.OptionIDEQ(optionsID)).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceOptionRepo.IsOptionsInSpace error: %v", err)
		return false, err
	}

	exists := count > 0

	// Cache the result if it exists
	if exists {
		go func() {
			relationship := &ent.SpaceOption{
				SpaceID:  spaceID,
				OptionID: optionsID,
			}
			r.cacheSpaceOption(context.Background(), relationship)
		}()
	}

	return exists, nil
}

// GetSpaceOption retrieves all options IDs for a space.
func (r *spaceOptionRepository) GetSpaceOption(ctx context.Context, spaceID string) ([]string, error) {
	// Try to get options IDs from cache
	cacheKey := fmt.Sprintf("space_options:%s", spaceID)
	var optionsIDs []string
	if err := r.spaceOptionListCache.GetArray(ctx, cacheKey, &optionsIDs); err == nil && len(optionsIDs) > 0 {
		return optionsIDs, nil
	}

	// Fallback to database
	spaceOption, err := r.data.GetSlaveEntClient().SpaceOption.Query().
		Where(spaceOptionEnt.SpaceIDEQ(spaceID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceOptionRepo.GetSpaceOption error: %v", err)
		return nil, err
	}

	// Extract options IDs
	optionsIDs = make([]string, len(spaceOption))
	for i, to := range spaceOption {
		optionsIDs[i] = to.OptionID
	}

	// Cache options IDs for future use
	go func() {
		if err := r.spaceOptionListCache.SetArray(context.Background(), cacheKey, optionsIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache space option %s: %v", spaceID, err)
		}
	}()

	return optionsIDs, nil
}

// cacheSpaceOption caches a space option relationship.
func (r *spaceOptionRepository) cacheSpaceOption(ctx context.Context, to *ent.SpaceOption) {
	relationshipKey := fmt.Sprintf("relationship:%s:%s", to.SpaceID, to.OptionID)
	if err := r.spaceOptionCache.Set(ctx, relationshipKey, to, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache space option relationship %s:%s: %v", to.SpaceID, to.OptionID, err)
	}
}

// invalidateSpaceOptionCache invalidates space option cache
func (r *spaceOptionRepository) invalidateSpaceOptionCache(ctx context.Context, spaceID, optionsID string) {
	relationshipKey := fmt.Sprintf("relationship:%s:%s", spaceID, optionsID)
	if err := r.spaceOptionCache.Delete(ctx, relationshipKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate space option relationship cache %s:%s: %v", spaceID, optionsID, err)
	}
}

// invalidateSpaceOptionListCache invalidates space option list cache
func (r *spaceOptionRepository) invalidateSpaceOptionListCache(ctx context.Context, spaceID string) {
	cacheKey := fmt.Sprintf("space_options:%s", spaceID)
	if err := r.spaceOptionListCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate space option cache %s: %v", spaceID, err)
	}
}

// invalidateOptionsSpacesCache invalidates options spaces cache
func (r *spaceOptionRepository) invalidateOptionsSpacesCache(ctx context.Context, optionsID string) {
	cacheKey := fmt.Sprintf("options_spaces:%s", optionsID)
	if err := r.optionsSpacesCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate options spaces cache %s: %v", optionsID, err)
	}
}
