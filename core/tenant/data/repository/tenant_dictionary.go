package repository

import (
	"context"
	"fmt"
	"ncobase/tenant/data"
	"ncobase/tenant/data/ent"
	tenantDictionaryEnt "ncobase/tenant/data/ent/tenantdictionary"
	"ncobase/tenant/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/logging/logger"
)

// TenantDictionaryRepositoryInterface represents the tenant dictionary repository interface.
type TenantDictionaryRepositoryInterface interface {
	Create(ctx context.Context, body *structs.TenantDictionary) (*ent.TenantDictionary, error)
	GetByTenantID(ctx context.Context, tenantID string) ([]*ent.TenantDictionary, error)
	GetByDictionaryID(ctx context.Context, dictionaryID string) ([]*ent.TenantDictionary, error)
	DeleteByTenantIDAndDictionaryID(ctx context.Context, tenantID, dictionaryID string) error
	DeleteAllByTenantID(ctx context.Context, tenantID string) error
	DeleteAllByDictionaryID(ctx context.Context, dictionaryID string) error
	IsDictionaryInTenant(ctx context.Context, tenantID, dictionaryID string) (bool, error)
	GetTenantDictionaries(ctx context.Context, tenantID string) ([]string, error)
}

// tenantDictionaryRepository implements the TenantDictionaryRepositoryInterface.
type tenantDictionaryRepository struct {
	data                    *data.Data
	tenantDictionaryCache   cache.ICache[ent.TenantDictionary]
	tenantDictionariesCache cache.ICache[[]string] // Maps tenant to dictionary IDs
	dictionaryTenantsCache  cache.ICache[[]string] // Maps dictionary to tenant IDs
	relationshipTTL         time.Duration
}

// NewTenantDictionaryRepository creates a new tenant dictionary repository.
func NewTenantDictionaryRepository(d *data.Data) TenantDictionaryRepositoryInterface {
	redisClient := d.GetRedis()

	return &tenantDictionaryRepository{
		data:                    d,
		tenantDictionaryCache:   cache.NewCache[ent.TenantDictionary](redisClient, "ncse_tenant:tenant_dictionaries"),
		tenantDictionariesCache: cache.NewCache[[]string](redisClient, "ncse_tenant:tenant_dict_mappings"),
		dictionaryTenantsCache:  cache.NewCache[[]string](redisClient, "ncse_tenant:dict_tenant_mappings"),
		relationshipTTL:         time.Hour * 2,
	}
}

// Create creates a new tenant dictionary relationship.
func (r *tenantDictionaryRepository) Create(ctx context.Context, body *structs.TenantDictionary) (*ent.TenantDictionary, error) {
	builder := r.data.GetMasterEntClient().TenantDictionary.Create()

	builder.SetNillableTenantID(&body.TenantID)
	builder.SetNillableDictionaryID(&body.DictionaryID)

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantDictionaryRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the relationship and invalidate related caches
	go func() {
		r.cacheTenantDictionary(context.Background(), row)
		r.invalidateTenantDictionariesCache(context.Background(), body.TenantID)
		r.invalidateDictionaryTenantsCache(context.Background(), body.DictionaryID)
	}()

	return row, nil
}

// GetByTenantID retrieves tenant dictionaries by tenant ID.
func (r *tenantDictionaryRepository) GetByTenantID(ctx context.Context, tenantID string) ([]*ent.TenantDictionary, error) {
	builder := r.data.GetSlaveEntClient().TenantDictionary.Query()
	builder.Where(tenantDictionaryEnt.TenantIDEQ(tenantID))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantDictionaryRepo.GetByTenantID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, td := range rows {
			r.cacheTenantDictionary(context.Background(), td)
		}
	}()

	return rows, nil
}

// GetByDictionaryID retrieves tenant dictionaries by dictionary ID.
func (r *tenantDictionaryRepository) GetByDictionaryID(ctx context.Context, dictionaryID string) ([]*ent.TenantDictionary, error) {
	builder := r.data.GetSlaveEntClient().TenantDictionary.Query()
	builder.Where(tenantDictionaryEnt.DictionaryIDEQ(dictionaryID))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantDictionaryRepo.GetByDictionaryID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, td := range rows {
			r.cacheTenantDictionary(context.Background(), td)
		}
	}()

	return rows, nil
}

// DeleteByTenantIDAndDictionaryID deletes tenant dictionary by tenant ID and dictionary ID.
func (r *tenantDictionaryRepository) DeleteByTenantIDAndDictionaryID(ctx context.Context, tenantID, dictionaryID string) error {
	if _, err := r.data.GetMasterEntClient().TenantDictionary.Delete().
		Where(tenantDictionaryEnt.TenantIDEQ(tenantID), tenantDictionaryEnt.DictionaryIDEQ(dictionaryID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "tenantDictionaryRepo.DeleteByTenantIDAndDictionaryID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateTenantDictionaryCache(context.Background(), tenantID, dictionaryID)
		r.invalidateTenantDictionariesCache(context.Background(), tenantID)
		r.invalidateDictionaryTenantsCache(context.Background(), dictionaryID)
	}()

	return nil
}

// DeleteAllByTenantID deletes all tenant dictionaries by tenant ID.
func (r *tenantDictionaryRepository) DeleteAllByTenantID(ctx context.Context, tenantID string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().TenantDictionary.Query().
		Where(tenantDictionaryEnt.TenantIDEQ(tenantID)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	if _, err := r.data.GetMasterEntClient().TenantDictionary.Delete().
		Where(tenantDictionaryEnt.TenantIDEQ(tenantID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "tenantDictionaryRepo.DeleteAllByTenantID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateTenantDictionariesCache(context.Background(), tenantID)
		for _, td := range relationships {
			r.invalidateTenantDictionaryCache(context.Background(), td.TenantID, td.DictionaryID)
			r.invalidateDictionaryTenantsCache(context.Background(), td.DictionaryID)
		}
	}()

	return nil
}

// DeleteAllByDictionaryID deletes all tenant dictionaries by dictionary ID.
func (r *tenantDictionaryRepository) DeleteAllByDictionaryID(ctx context.Context, dictionaryID string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().TenantDictionary.Query().
		Where(tenantDictionaryEnt.DictionaryIDEQ(dictionaryID)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	if _, err := r.data.GetMasterEntClient().TenantDictionary.Delete().
		Where(tenantDictionaryEnt.DictionaryIDEQ(dictionaryID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "tenantDictionaryRepo.DeleteAllByDictionaryID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateDictionaryTenantsCache(context.Background(), dictionaryID)
		for _, td := range relationships {
			r.invalidateTenantDictionaryCache(context.Background(), td.TenantID, td.DictionaryID)
			r.invalidateTenantDictionariesCache(context.Background(), td.TenantID)
		}
	}()

	return nil
}

// IsDictionaryInTenant verifies if a dictionary belongs to a tenant.
func (r *tenantDictionaryRepository) IsDictionaryInTenant(ctx context.Context, tenantID, dictionaryID string) (bool, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("relationship:%s:%s", tenantID, dictionaryID)
	if cached, err := r.tenantDictionaryCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return true, nil
	}

	count, err := r.data.GetSlaveEntClient().TenantDictionary.Query().
		Where(tenantDictionaryEnt.TenantIDEQ(tenantID), tenantDictionaryEnt.DictionaryIDEQ(dictionaryID)).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantDictionaryRepo.IsDictionaryInTenant error: %v", err)
		return false, err
	}

	exists := count > 0

	// Cache the result if it exists
	if exists {
		go func() {
			relationship := &ent.TenantDictionary{
				TenantID:     tenantID,
				DictionaryID: dictionaryID,
			}
			r.cacheTenantDictionary(context.Background(), relationship)
		}()
	}

	return exists, nil
}

// GetTenantDictionaries retrieves all dictionary IDs for a tenant.
func (r *tenantDictionaryRepository) GetTenantDictionaries(ctx context.Context, tenantID string) ([]string, error) {
	// Try to get dictionary IDs from cache
	cacheKey := fmt.Sprintf("tenant_dictionaries:%s", tenantID)
	var dictionaryIDs []string
	if err := r.tenantDictionariesCache.GetArray(ctx, cacheKey, &dictionaryIDs); err == nil && len(dictionaryIDs) > 0 {
		return dictionaryIDs, nil
	}

	// Fallback to database
	tenantDictionaries, err := r.data.GetSlaveEntClient().TenantDictionary.Query().
		Where(tenantDictionaryEnt.TenantIDEQ(tenantID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantDictionaryRepo.GetTenantDictionaries error: %v", err)
		return nil, err
	}

	// Extract dictionary IDs
	dictionaryIDs = make([]string, len(tenantDictionaries))
	for i, td := range tenantDictionaries {
		dictionaryIDs[i] = td.DictionaryID
	}

	// Cache dictionary IDs for future use
	go func() {
		if err := r.tenantDictionariesCache.SetArray(context.Background(), cacheKey, dictionaryIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache tenant dictionaries %s: %v", tenantID, err)
		}
	}()

	return dictionaryIDs, nil
}

// cacheTenantDictionary caches a tenant dictionary relationship.
func (r *tenantDictionaryRepository) cacheTenantDictionary(ctx context.Context, td *ent.TenantDictionary) {
	relationshipKey := fmt.Sprintf("relationship:%s:%s", td.TenantID, td.DictionaryID)
	if err := r.tenantDictionaryCache.Set(ctx, relationshipKey, td, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache tenant dictionary relationship %s:%s: %v", td.TenantID, td.DictionaryID, err)
	}
}

// invalidateTenantDictionaryCache invalidates tenant dictionary cache
func (r *tenantDictionaryRepository) invalidateTenantDictionaryCache(ctx context.Context, tenantID, dictionaryID string) {
	relationshipKey := fmt.Sprintf("relationship:%s:%s", tenantID, dictionaryID)
	if err := r.tenantDictionaryCache.Delete(ctx, relationshipKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate tenant dictionary relationship cache %s:%s: %v", tenantID, dictionaryID, err)
	}
}

// invalidateTenantDictionariesCache invalidates tenant dictionaries cache
func (r *tenantDictionaryRepository) invalidateTenantDictionariesCache(ctx context.Context, tenantID string) {
	cacheKey := fmt.Sprintf("tenant_dictionaries:%s", tenantID)
	if err := r.tenantDictionariesCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate tenant dictionaries cache %s: %v", tenantID, err)
	}
}

// invalidateDictionaryTenantsCache invalidates dictionary tenants cache
func (r *tenantDictionaryRepository) invalidateDictionaryTenantsCache(ctx context.Context, dictionaryID string) {
	cacheKey := fmt.Sprintf("dictionary_tenants:%s", dictionaryID)
	if err := r.dictionaryTenantsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate dictionary tenants cache %s: %v", dictionaryID, err)
	}
}
