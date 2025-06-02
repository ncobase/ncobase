package repository

import (
	"context"
	"fmt"
	"ncobase/tenant/data"
	"ncobase/tenant/data/ent"
	tenantOptionsEnt "ncobase/tenant/data/ent/tenantoptions"
	"ncobase/tenant/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/logging/logger"
)

// TenantOptionsRepositoryInterface represents the tenant options repository interface.
type TenantOptionsRepositoryInterface interface {
	Create(ctx context.Context, body *structs.TenantOptions) (*ent.TenantOptions, error)
	GetByTenantID(ctx context.Context, tenantID string) ([]*ent.TenantOptions, error)
	GetByOptionsID(ctx context.Context, optionsID string) ([]*ent.TenantOptions, error)
	DeleteByTenantIDAndOptionsID(ctx context.Context, tenantID, optionsID string) error
	DeleteAllByTenantID(ctx context.Context, tenantID string) error
	DeleteAllByOptionsID(ctx context.Context, optionsID string) error
	IsOptionsInTenant(ctx context.Context, tenantID, optionsID string) (bool, error)
	GetTenantOptions(ctx context.Context, tenantID string) ([]string, error)
}

// tenantOptionsRepository implements the TenantOptionsRepositoryInterface.
type tenantOptionsRepository struct {
	data                   *data.Data
	tenantOptionsCache     cache.ICache[ent.TenantOptions]
	tenantOptionsListCache cache.ICache[[]string] // Maps tenant to options IDs
	optionsTenantsCache    cache.ICache[[]string] // Maps options to tenant IDs
	relationshipTTL        time.Duration
}

// NewTenantOptionsRepository creates a new tenant options repository.
func NewTenantOptionsRepository(d *data.Data) TenantOptionsRepositoryInterface {
	redisClient := d.GetRedis()

	return &tenantOptionsRepository{
		data:                   d,
		tenantOptionsCache:     cache.NewCache[ent.TenantOptions](redisClient, "ncse_tenant:tenant_options"),
		tenantOptionsListCache: cache.NewCache[[]string](redisClient, "ncse_tenant:tenant_options_mappings"),
		optionsTenantsCache:    cache.NewCache[[]string](redisClient, "ncse_tenant:options_tenant_mappings"),
		relationshipTTL:        time.Hour * 2,
	}
}

// Create creates a new tenant options relationship.
func (r *tenantOptionsRepository) Create(ctx context.Context, body *structs.TenantOptions) (*ent.TenantOptions, error) {
	builder := r.data.GetMasterEntClient().TenantOptions.Create()

	builder.SetNillableTenantID(&body.TenantID)
	builder.SetNillableOptionsID(&body.OptionsID)

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantOptionsRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the relationship and invalidate related caches
	go func() {
		r.cacheTenantOptions(context.Background(), row)
		r.invalidateTenantOptionsListCache(context.Background(), body.TenantID)
		r.invalidateOptionsTenantsCache(context.Background(), body.OptionsID)
	}()

	return row, nil
}

// GetByTenantID retrieves tenant options by tenant ID.
func (r *tenantOptionsRepository) GetByTenantID(ctx context.Context, tenantID string) ([]*ent.TenantOptions, error) {
	builder := r.data.GetSlaveEntClient().TenantOptions.Query()
	builder.Where(tenantOptionsEnt.TenantIDEQ(tenantID))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantOptionsRepo.GetByTenantID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, to := range rows {
			r.cacheTenantOptions(context.Background(), to)
		}
	}()

	return rows, nil
}

// GetByOptionsID retrieves tenant options by options ID.
func (r *tenantOptionsRepository) GetByOptionsID(ctx context.Context, optionsID string) ([]*ent.TenantOptions, error) {
	builder := r.data.GetSlaveEntClient().TenantOptions.Query()
	builder.Where(tenantOptionsEnt.OptionsIDEQ(optionsID))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantOptionsRepo.GetByOptionsID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, to := range rows {
			r.cacheTenantOptions(context.Background(), to)
		}
	}()

	return rows, nil
}

// DeleteByTenantIDAndOptionsID deletes tenant options by tenant ID and options ID.
func (r *tenantOptionsRepository) DeleteByTenantIDAndOptionsID(ctx context.Context, tenantID, optionsID string) error {
	if _, err := r.data.GetMasterEntClient().TenantOptions.Delete().
		Where(tenantOptionsEnt.TenantIDEQ(tenantID), tenantOptionsEnt.OptionsIDEQ(optionsID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "tenantOptionsRepo.DeleteByTenantIDAndOptionsID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateTenantOptionsCache(context.Background(), tenantID, optionsID)
		r.invalidateTenantOptionsListCache(context.Background(), tenantID)
		r.invalidateOptionsTenantsCache(context.Background(), optionsID)
	}()

	return nil
}

// DeleteAllByTenantID deletes all tenant options by tenant ID.
func (r *tenantOptionsRepository) DeleteAllByTenantID(ctx context.Context, tenantID string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().TenantOptions.Query().
		Where(tenantOptionsEnt.TenantIDEQ(tenantID)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	if _, err := r.data.GetMasterEntClient().TenantOptions.Delete().
		Where(tenantOptionsEnt.TenantIDEQ(tenantID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "tenantOptionsRepo.DeleteAllByTenantID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateTenantOptionsListCache(context.Background(), tenantID)
		for _, to := range relationships {
			r.invalidateTenantOptionsCache(context.Background(), to.TenantID, to.OptionsID)
			r.invalidateOptionsTenantsCache(context.Background(), to.OptionsID)
		}
	}()

	return nil
}

// DeleteAllByOptionsID deletes all tenant options by options ID.
func (r *tenantOptionsRepository) DeleteAllByOptionsID(ctx context.Context, optionsID string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().TenantOptions.Query().
		Where(tenantOptionsEnt.OptionsIDEQ(optionsID)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	if _, err := r.data.GetMasterEntClient().TenantOptions.Delete().
		Where(tenantOptionsEnt.OptionsIDEQ(optionsID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "tenantOptionsRepo.DeleteAllByOptionsID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateOptionsTenantsCache(context.Background(), optionsID)
		for _, to := range relationships {
			r.invalidateTenantOptionsCache(context.Background(), to.TenantID, to.OptionsID)
			r.invalidateTenantOptionsListCache(context.Background(), to.TenantID)
		}
	}()

	return nil
}

// IsOptionsInTenant verifies if an options belongs to a tenant.
func (r *tenantOptionsRepository) IsOptionsInTenant(ctx context.Context, tenantID, optionsID string) (bool, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("relationship:%s:%s", tenantID, optionsID)
	if cached, err := r.tenantOptionsCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return true, nil
	}

	count, err := r.data.GetSlaveEntClient().TenantOptions.Query().
		Where(tenantOptionsEnt.TenantIDEQ(tenantID), tenantOptionsEnt.OptionsIDEQ(optionsID)).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantOptionsRepo.IsOptionsInTenant error: %v", err)
		return false, err
	}

	exists := count > 0

	// Cache the result if it exists
	if exists {
		go func() {
			relationship := &ent.TenantOptions{
				TenantID:  tenantID,
				OptionsID: optionsID,
			}
			r.cacheTenantOptions(context.Background(), relationship)
		}()
	}

	return exists, nil
}

// GetTenantOptions retrieves all options IDs for a tenant.
func (r *tenantOptionsRepository) GetTenantOptions(ctx context.Context, tenantID string) ([]string, error) {
	// Try to get options IDs from cache
	cacheKey := fmt.Sprintf("tenant_options:%s", tenantID)
	var optionsIDs []string
	if err := r.tenantOptionsListCache.GetArray(ctx, cacheKey, &optionsIDs); err == nil && len(optionsIDs) > 0 {
		return optionsIDs, nil
	}

	// Fallback to database
	tenantOptions, err := r.data.GetSlaveEntClient().TenantOptions.Query().
		Where(tenantOptionsEnt.TenantIDEQ(tenantID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantOptionsRepo.GetTenantOptions error: %v", err)
		return nil, err
	}

	// Extract options IDs
	optionsIDs = make([]string, len(tenantOptions))
	for i, to := range tenantOptions {
		optionsIDs[i] = to.OptionsID
	}

	// Cache options IDs for future use
	go func() {
		if err := r.tenantOptionsListCache.SetArray(context.Background(), cacheKey, optionsIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache tenant options %s: %v", tenantID, err)
		}
	}()

	return optionsIDs, nil
}

// cacheTenantOptions caches a tenant options relationship.
func (r *tenantOptionsRepository) cacheTenantOptions(ctx context.Context, to *ent.TenantOptions) {
	relationshipKey := fmt.Sprintf("relationship:%s:%s", to.TenantID, to.OptionsID)
	if err := r.tenantOptionsCache.Set(ctx, relationshipKey, to, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache tenant options relationship %s:%s: %v", to.TenantID, to.OptionsID, err)
	}
}

// invalidateTenantOptionsCache invalidates tenant options cache
func (r *tenantOptionsRepository) invalidateTenantOptionsCache(ctx context.Context, tenantID, optionsID string) {
	relationshipKey := fmt.Sprintf("relationship:%s:%s", tenantID, optionsID)
	if err := r.tenantOptionsCache.Delete(ctx, relationshipKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate tenant options relationship cache %s:%s: %v", tenantID, optionsID, err)
	}
}

// invalidateTenantOptionsListCache invalidates tenant options list cache
func (r *tenantOptionsRepository) invalidateTenantOptionsListCache(ctx context.Context, tenantID string) {
	cacheKey := fmt.Sprintf("tenant_options:%s", tenantID)
	if err := r.tenantOptionsListCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate tenant options cache %s: %v", tenantID, err)
	}
}

// invalidateOptionsTenantsCache invalidates options tenants cache
func (r *tenantOptionsRepository) invalidateOptionsTenantsCache(ctx context.Context, optionsID string) {
	cacheKey := fmt.Sprintf("options_tenants:%s", optionsID)
	if err := r.optionsTenantsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate options tenants cache %s: %v", optionsID, err)
	}
}
