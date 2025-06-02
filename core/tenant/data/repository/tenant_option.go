package repository

import (
	"context"
	"fmt"
	"ncobase/tenant/data"
	"ncobase/tenant/data/ent"
	tenantOptionEnt "ncobase/tenant/data/ent/tenantoption"
	"ncobase/tenant/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/logging/logger"
)

// TenantOptionRepositoryInterface represents the tenant option repository interface.
type TenantOptionRepositoryInterface interface {
	Create(ctx context.Context, body *structs.TenantOption) (*ent.TenantOption, error)
	GetByTenantID(ctx context.Context, tenantID string) ([]*ent.TenantOption, error)
	GetByOptionID(ctx context.Context, optionsID string) ([]*ent.TenantOption, error)
	DeleteByTenantIDAndOptionID(ctx context.Context, tenantID, optionsID string) error
	DeleteAllByTenantID(ctx context.Context, tenantID string) error
	DeleteAllByOptionID(ctx context.Context, optionsID string) error
	IsOptionsInTenant(ctx context.Context, tenantID, optionsID string) (bool, error)
	GetTenantOption(ctx context.Context, tenantID string) ([]string, error)
}

// tenantOptionRepository implements the TenantOptionRepositoryInterface.
type tenantOptionRepository struct {
	data                  *data.Data
	tenantOptionCache     cache.ICache[ent.TenantOption]
	tenantOptionListCache cache.ICache[[]string] // Maps tenant to options IDs
	optionsTenantsCache   cache.ICache[[]string] // Maps options to tenant IDs
	relationshipTTL       time.Duration
}

// NewTenantOptionRepository creates a new tenant option repository.
func NewTenantOptionRepository(d *data.Data) TenantOptionRepositoryInterface {
	redisClient := d.GetRedis()

	return &tenantOptionRepository{
		data:                  d,
		tenantOptionCache:     cache.NewCache[ent.TenantOption](redisClient, "ncse_tenant:tenant_options"),
		tenantOptionListCache: cache.NewCache[[]string](redisClient, "ncse_tenant:tenant_options_mappings"),
		optionsTenantsCache:   cache.NewCache[[]string](redisClient, "ncse_tenant:options_tenant_mappings"),
		relationshipTTL:       time.Hour * 2,
	}
}

// Create creates a new tenant option relationship.
func (r *tenantOptionRepository) Create(ctx context.Context, body *structs.TenantOption) (*ent.TenantOption, error) {
	builder := r.data.GetMasterEntClient().TenantOption.Create()

	builder.SetNillableTenantID(&body.TenantID)
	builder.SetNillableOptionID(&body.OptionID)

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantOptionRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the relationship and invalidate related caches
	go func() {
		r.cacheTenantOption(context.Background(), row)
		r.invalidateTenantOptionListCache(context.Background(), body.TenantID)
		r.invalidateOptionsTenantsCache(context.Background(), body.OptionID)
	}()

	return row, nil
}

// GetByTenantID retrieves tenant option by tenant ID.
func (r *tenantOptionRepository) GetByTenantID(ctx context.Context, tenantID string) ([]*ent.TenantOption, error) {
	builder := r.data.GetSlaveEntClient().TenantOption.Query()
	builder.Where(tenantOptionEnt.TenantIDEQ(tenantID))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantOptionRepo.GetByTenantID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, to := range rows {
			r.cacheTenantOption(context.Background(), to)
		}
	}()

	return rows, nil
}

// GetByOptionID retrieves tenant option by options ID.
func (r *tenantOptionRepository) GetByOptionID(ctx context.Context, optionsID string) ([]*ent.TenantOption, error) {
	builder := r.data.GetSlaveEntClient().TenantOption.Query()
	builder.Where(tenantOptionEnt.OptionIDEQ(optionsID))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantOptionRepo.GetByOptionID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, to := range rows {
			r.cacheTenantOption(context.Background(), to)
		}
	}()

	return rows, nil
}

// DeleteByTenantIDAndOptionID deletes tenant option by tenant ID and options ID.
func (r *tenantOptionRepository) DeleteByTenantIDAndOptionID(ctx context.Context, tenantID, optionsID string) error {
	if _, err := r.data.GetMasterEntClient().TenantOption.Delete().
		Where(tenantOptionEnt.TenantIDEQ(tenantID), tenantOptionEnt.OptionIDEQ(optionsID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "tenantOptionRepo.DeleteByTenantIDAndOptionID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateTenantOptionCache(context.Background(), tenantID, optionsID)
		r.invalidateTenantOptionListCache(context.Background(), tenantID)
		r.invalidateOptionsTenantsCache(context.Background(), optionsID)
	}()

	return nil
}

// DeleteAllByTenantID deletes all tenant option by tenant ID.
func (r *tenantOptionRepository) DeleteAllByTenantID(ctx context.Context, tenantID string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().TenantOption.Query().
		Where(tenantOptionEnt.TenantIDEQ(tenantID)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	if _, err := r.data.GetMasterEntClient().TenantOption.Delete().
		Where(tenantOptionEnt.TenantIDEQ(tenantID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "tenantOptionRepo.DeleteAllByTenantID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateTenantOptionListCache(context.Background(), tenantID)
		for _, to := range relationships {
			r.invalidateTenantOptionCache(context.Background(), to.TenantID, to.OptionID)
			r.invalidateOptionsTenantsCache(context.Background(), to.OptionID)
		}
	}()

	return nil
}

// DeleteAllByOptionID deletes all tenant option by options ID.
func (r *tenantOptionRepository) DeleteAllByOptionID(ctx context.Context, optionsID string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().TenantOption.Query().
		Where(tenantOptionEnt.OptionIDEQ(optionsID)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	if _, err := r.data.GetMasterEntClient().TenantOption.Delete().
		Where(tenantOptionEnt.OptionIDEQ(optionsID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "tenantOptionRepo.DeleteAllByOptionID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateOptionsTenantsCache(context.Background(), optionsID)
		for _, to := range relationships {
			r.invalidateTenantOptionCache(context.Background(), to.TenantID, to.OptionID)
			r.invalidateTenantOptionListCache(context.Background(), to.TenantID)
		}
	}()

	return nil
}

// IsOptionsInTenant verifies if an options belongs to a tenant.
func (r *tenantOptionRepository) IsOptionsInTenant(ctx context.Context, tenantID, optionsID string) (bool, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("relationship:%s:%s", tenantID, optionsID)
	if cached, err := r.tenantOptionCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return true, nil
	}

	count, err := r.data.GetSlaveEntClient().TenantOption.Query().
		Where(tenantOptionEnt.TenantIDEQ(tenantID), tenantOptionEnt.OptionIDEQ(optionsID)).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantOptionRepo.IsOptionsInTenant error: %v", err)
		return false, err
	}

	exists := count > 0

	// Cache the result if it exists
	if exists {
		go func() {
			relationship := &ent.TenantOption{
				TenantID: tenantID,
				OptionID: optionsID,
			}
			r.cacheTenantOption(context.Background(), relationship)
		}()
	}

	return exists, nil
}

// GetTenantOption retrieves all options IDs for a tenant.
func (r *tenantOptionRepository) GetTenantOption(ctx context.Context, tenantID string) ([]string, error) {
	// Try to get options IDs from cache
	cacheKey := fmt.Sprintf("tenant_options:%s", tenantID)
	var optionsIDs []string
	if err := r.tenantOptionListCache.GetArray(ctx, cacheKey, &optionsIDs); err == nil && len(optionsIDs) > 0 {
		return optionsIDs, nil
	}

	// Fallback to database
	tenantOption, err := r.data.GetSlaveEntClient().TenantOption.Query().
		Where(tenantOptionEnt.TenantIDEQ(tenantID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantOptionRepo.GetTenantOption error: %v", err)
		return nil, err
	}

	// Extract options IDs
	optionsIDs = make([]string, len(tenantOption))
	for i, to := range tenantOption {
		optionsIDs[i] = to.OptionID
	}

	// Cache options IDs for future use
	go func() {
		if err := r.tenantOptionListCache.SetArray(context.Background(), cacheKey, optionsIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache tenant option %s: %v", tenantID, err)
		}
	}()

	return optionsIDs, nil
}

// cacheTenantOption caches a tenant option relationship.
func (r *tenantOptionRepository) cacheTenantOption(ctx context.Context, to *ent.TenantOption) {
	relationshipKey := fmt.Sprintf("relationship:%s:%s", to.TenantID, to.OptionID)
	if err := r.tenantOptionCache.Set(ctx, relationshipKey, to, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache tenant option relationship %s:%s: %v", to.TenantID, to.OptionID, err)
	}
}

// invalidateTenantOptionCache invalidates tenant option cache
func (r *tenantOptionRepository) invalidateTenantOptionCache(ctx context.Context, tenantID, optionsID string) {
	relationshipKey := fmt.Sprintf("relationship:%s:%s", tenantID, optionsID)
	if err := r.tenantOptionCache.Delete(ctx, relationshipKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate tenant option relationship cache %s:%s: %v", tenantID, optionsID, err)
	}
}

// invalidateTenantOptionListCache invalidates tenant option list cache
func (r *tenantOptionRepository) invalidateTenantOptionListCache(ctx context.Context, tenantID string) {
	cacheKey := fmt.Sprintf("tenant_options:%s", tenantID)
	if err := r.tenantOptionListCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate tenant option cache %s: %v", tenantID, err)
	}
}

// invalidateOptionsTenantsCache invalidates options tenants cache
func (r *tenantOptionRepository) invalidateOptionsTenantsCache(ctx context.Context, optionsID string) {
	cacheKey := fmt.Sprintf("options_tenants:%s", optionsID)
	if err := r.optionsTenantsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate options tenants cache %s: %v", optionsID, err)
	}
}
