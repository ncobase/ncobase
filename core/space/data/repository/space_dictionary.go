package repository

import (
	"context"
	"fmt"
	"ncobase/space/data"
	"ncobase/space/data/ent"
	spaceDictionaryEnt "ncobase/space/data/ent/spacedictionary"
	"ncobase/space/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/logging/logger"
)

// SpaceDictionaryRepositoryInterface represents the space dictionary repository interface.
type SpaceDictionaryRepositoryInterface interface {
	Create(ctx context.Context, body *structs.SpaceDictionary) (*ent.SpaceDictionary, error)
	GetBySpaceID(ctx context.Context, spaceID string) ([]*ent.SpaceDictionary, error)
	GetByDictionaryID(ctx context.Context, dictionaryID string) ([]*ent.SpaceDictionary, error)
	DeleteBySpaceIDAndDictionaryID(ctx context.Context, spaceID, dictionaryID string) error
	DeleteAllBySpaceID(ctx context.Context, spaceID string) error
	DeleteAllByDictionaryID(ctx context.Context, dictionaryID string) error
	IsDictionaryInSpace(ctx context.Context, spaceID, dictionaryID string) (bool, error)
	GetSpaceDictionaries(ctx context.Context, spaceID string) ([]string, error)
}

// spaceDictionaryRepository implements the SpaceDictionaryRepositoryInterface.
type spaceDictionaryRepository struct {
	data                   *data.Data
	spaceDictionaryCache   cache.ICache[ent.SpaceDictionary]
	spaceDictionariesCache cache.ICache[[]string] // Maps space to dictionary IDs
	dictionarySpacesCache  cache.ICache[[]string] // Maps dictionary to space IDs
	relationshipTTL        time.Duration
}

// NewSpaceDictionaryRepository creates a new space dictionary repository.
func NewSpaceDictionaryRepository(d *data.Data) SpaceDictionaryRepositoryInterface {
	redisClient := d.GetRedis()

	return &spaceDictionaryRepository{
		data:                   d,
		spaceDictionaryCache:   cache.NewCache[ent.SpaceDictionary](redisClient, "ncse_space:space_dictionaries"),
		spaceDictionariesCache: cache.NewCache[[]string](redisClient, "ncse_space:space_dict_mappings"),
		dictionarySpacesCache:  cache.NewCache[[]string](redisClient, "ncse_space:dict_space_mappings"),
		relationshipTTL:        time.Hour * 2,
	}
}

// Create creates a new space dictionary relationship.
func (r *spaceDictionaryRepository) Create(ctx context.Context, body *structs.SpaceDictionary) (*ent.SpaceDictionary, error) {
	builder := r.data.GetMasterEntClient().SpaceDictionary.Create()

	builder.SetNillableSpaceID(&body.SpaceID)
	builder.SetNillableDictionaryID(&body.DictionaryID)

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceDictionaryRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the relationship and invalidate related caches
	go func() {
		r.cacheSpaceDictionary(context.Background(), row)
		r.invalidateSpaceDictionariesCache(context.Background(), body.SpaceID)
		r.invalidateDictionarySpacesCache(context.Background(), body.DictionaryID)
	}()

	return row, nil
}

// GetBySpaceID retrieves space dictionaries by space ID.
func (r *spaceDictionaryRepository) GetBySpaceID(ctx context.Context, spaceID string) ([]*ent.SpaceDictionary, error) {
	builder := r.data.GetSlaveEntClient().SpaceDictionary.Query()
	builder.Where(spaceDictionaryEnt.SpaceIDEQ(spaceID))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceDictionaryRepo.GetBySpaceID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, td := range rows {
			r.cacheSpaceDictionary(context.Background(), td)
		}
	}()

	return rows, nil
}

// GetByDictionaryID retrieves space dictionaries by dictionary ID.
func (r *spaceDictionaryRepository) GetByDictionaryID(ctx context.Context, dictionaryID string) ([]*ent.SpaceDictionary, error) {
	builder := r.data.GetSlaveEntClient().SpaceDictionary.Query()
	builder.Where(spaceDictionaryEnt.DictionaryIDEQ(dictionaryID))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceDictionaryRepo.GetByDictionaryID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, td := range rows {
			r.cacheSpaceDictionary(context.Background(), td)
		}
	}()

	return rows, nil
}

// DeleteBySpaceIDAndDictionaryID deletes space dictionary by space ID and dictionary ID.
func (r *spaceDictionaryRepository) DeleteBySpaceIDAndDictionaryID(ctx context.Context, spaceID, dictionaryID string) error {
	if _, err := r.data.GetMasterEntClient().SpaceDictionary.Delete().
		Where(spaceDictionaryEnt.SpaceIDEQ(spaceID), spaceDictionaryEnt.DictionaryIDEQ(dictionaryID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "spaceDictionaryRepo.DeleteBySpaceIDAndDictionaryID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateSpaceDictionaryCache(context.Background(), spaceID, dictionaryID)
		r.invalidateSpaceDictionariesCache(context.Background(), spaceID)
		r.invalidateDictionarySpacesCache(context.Background(), dictionaryID)
	}()

	return nil
}

// DeleteAllBySpaceID deletes all space dictionaries by space ID.
func (r *spaceDictionaryRepository) DeleteAllBySpaceID(ctx context.Context, spaceID string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().SpaceDictionary.Query().
		Where(spaceDictionaryEnt.SpaceIDEQ(spaceID)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	if _, err := r.data.GetMasterEntClient().SpaceDictionary.Delete().
		Where(spaceDictionaryEnt.SpaceIDEQ(spaceID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "spaceDictionaryRepo.DeleteAllBySpaceID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateSpaceDictionariesCache(context.Background(), spaceID)
		for _, td := range relationships {
			r.invalidateSpaceDictionaryCache(context.Background(), td.SpaceID, td.DictionaryID)
			r.invalidateDictionarySpacesCache(context.Background(), td.DictionaryID)
		}
	}()

	return nil
}

// DeleteAllByDictionaryID deletes all space dictionaries by dictionary ID.
func (r *spaceDictionaryRepository) DeleteAllByDictionaryID(ctx context.Context, dictionaryID string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().SpaceDictionary.Query().
		Where(spaceDictionaryEnt.DictionaryIDEQ(dictionaryID)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	if _, err := r.data.GetMasterEntClient().SpaceDictionary.Delete().
		Where(spaceDictionaryEnt.DictionaryIDEQ(dictionaryID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "spaceDictionaryRepo.DeleteAllByDictionaryID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateDictionarySpacesCache(context.Background(), dictionaryID)
		for _, td := range relationships {
			r.invalidateSpaceDictionaryCache(context.Background(), td.SpaceID, td.DictionaryID)
			r.invalidateSpaceDictionariesCache(context.Background(), td.SpaceID)
		}
	}()

	return nil
}

// IsDictionaryInSpace verifies if a dictionary belongs to a space.
func (r *spaceDictionaryRepository) IsDictionaryInSpace(ctx context.Context, spaceID, dictionaryID string) (bool, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("relationship:%s:%s", spaceID, dictionaryID)
	if cached, err := r.spaceDictionaryCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return true, nil
	}

	count, err := r.data.GetSlaveEntClient().SpaceDictionary.Query().
		Where(spaceDictionaryEnt.SpaceIDEQ(spaceID), spaceDictionaryEnt.DictionaryIDEQ(dictionaryID)).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceDictionaryRepo.IsDictionaryInSpace error: %v", err)
		return false, err
	}

	exists := count > 0

	// Cache the result if it exists
	if exists {
		go func() {
			relationship := &ent.SpaceDictionary{
				SpaceID:      spaceID,
				DictionaryID: dictionaryID,
			}
			r.cacheSpaceDictionary(context.Background(), relationship)
		}()
	}

	return exists, nil
}

// GetSpaceDictionaries retrieves all dictionary IDs for a space.
func (r *spaceDictionaryRepository) GetSpaceDictionaries(ctx context.Context, spaceID string) ([]string, error) {
	// Try to get dictionary IDs from cache
	cacheKey := fmt.Sprintf("space_dictionaries:%s", spaceID)
	var dictionaryIDs []string
	if err := r.spaceDictionariesCache.GetArray(ctx, cacheKey, &dictionaryIDs); err == nil && len(dictionaryIDs) > 0 {
		return dictionaryIDs, nil
	}

	// Fallback to database
	spaceDictionaries, err := r.data.GetSlaveEntClient().SpaceDictionary.Query().
		Where(spaceDictionaryEnt.SpaceIDEQ(spaceID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceDictionaryRepo.GetSpaceDictionaries error: %v", err)
		return nil, err
	}

	// Extract dictionary IDs
	dictionaryIDs = make([]string, len(spaceDictionaries))
	for i, td := range spaceDictionaries {
		dictionaryIDs[i] = td.DictionaryID
	}

	// Cache dictionary IDs for future use
	go func() {
		if err := r.spaceDictionariesCache.SetArray(context.Background(), cacheKey, dictionaryIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache space dictionaries %s: %v", spaceID, err)
		}
	}()

	return dictionaryIDs, nil
}

// cacheSpaceDictionary caches a space dictionary relationship.
func (r *spaceDictionaryRepository) cacheSpaceDictionary(ctx context.Context, td *ent.SpaceDictionary) {
	relationshipKey := fmt.Sprintf("relationship:%s:%s", td.SpaceID, td.DictionaryID)
	if err := r.spaceDictionaryCache.Set(ctx, relationshipKey, td, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache space dictionary relationship %s:%s: %v", td.SpaceID, td.DictionaryID, err)
	}
}

// invalidateSpaceDictionaryCache invalidates space dictionary cache
func (r *spaceDictionaryRepository) invalidateSpaceDictionaryCache(ctx context.Context, spaceID, dictionaryID string) {
	relationshipKey := fmt.Sprintf("relationship:%s:%s", spaceID, dictionaryID)
	if err := r.spaceDictionaryCache.Delete(ctx, relationshipKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate space dictionary relationship cache %s:%s: %v", spaceID, dictionaryID, err)
	}
}

// invalidateSpaceDictionariesCache invalidates space dictionaries cache
func (r *spaceDictionaryRepository) invalidateSpaceDictionariesCache(ctx context.Context, spaceID string) {
	cacheKey := fmt.Sprintf("space_dictionaries:%s", spaceID)
	if err := r.spaceDictionariesCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate space dictionaries cache %s: %v", spaceID, err)
	}
}

// invalidateDictionarySpacesCache invalidates dictionary spaces cache
func (r *spaceDictionaryRepository) invalidateDictionarySpacesCache(ctx context.Context, dictionaryID string) {
	cacheKey := fmt.Sprintf("dictionary_spaces:%s", dictionaryID)
	if err := r.dictionarySpacesCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate dictionary spaces cache %s: %v", dictionaryID, err)
	}
}
