package repository

import (
	"context"
	"fmt"
	"ncobase/tenant/data"
	"ncobase/tenant/data/ent"
	tenantMenuEnt "ncobase/tenant/data/ent/tenantmenu"
	"ncobase/tenant/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/logging/logger"
)

// TenantMenuRepositoryInterface represents the tenant menu repository interface.
type TenantMenuRepositoryInterface interface {
	Create(ctx context.Context, body *structs.TenantMenu) (*ent.TenantMenu, error)
	GetByTenantID(ctx context.Context, tenantID string) ([]*ent.TenantMenu, error)
	GetByMenuID(ctx context.Context, menuID string) ([]*ent.TenantMenu, error)
	DeleteByTenantIDAndMenuID(ctx context.Context, tenantID, menuID string) error
	DeleteAllByTenantID(ctx context.Context, tenantID string) error
	DeleteAllByMenuID(ctx context.Context, menuID string) error
	IsMenuInTenant(ctx context.Context, tenantID, menuID string) (bool, error)
	GetTenantMenus(ctx context.Context, tenantID string) ([]string, error)
}

// tenantMenuRepository implements the TenantMenuRepositoryInterface.
type tenantMenuRepository struct {
	data             *data.Data
	tenantMenuCache  cache.ICache[ent.TenantMenu]
	tenantMenusCache cache.ICache[[]string] // Maps tenant to menu IDs
	menuTenantsCache cache.ICache[[]string] // Maps menu to tenant IDs
	relationshipTTL  time.Duration
}

// NewTenantMenuRepository creates a new tenant menu repository.
func NewTenantMenuRepository(d *data.Data) TenantMenuRepositoryInterface {
	redisClient := d.GetRedis()

	return &tenantMenuRepository{
		data:             d,
		tenantMenuCache:  cache.NewCache[ent.TenantMenu](redisClient, "ncse_tenant:tenant_menus"),
		tenantMenusCache: cache.NewCache[[]string](redisClient, "ncse_tenant:tenant_menu_mappings"),
		menuTenantsCache: cache.NewCache[[]string](redisClient, "ncse_tenant:menu_tenant_mappings"),
		relationshipTTL:  time.Hour * 2,
	}
}

// Create creates a new tenant menu relationship.
func (r *tenantMenuRepository) Create(ctx context.Context, body *structs.TenantMenu) (*ent.TenantMenu, error) {
	builder := r.data.GetMasterEntClient().TenantMenu.Create()

	builder.SetNillableTenantID(&body.TenantID)
	builder.SetNillableMenuID(&body.MenuID)

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantMenuRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the relationship and invalidate related caches
	go func() {
		r.cacheTenantMenu(context.Background(), row)
		r.invalidateTenantMenusCache(context.Background(), body.TenantID)
		r.invalidateMenuTenantsCache(context.Background(), body.MenuID)
	}()

	return row, nil
}

// GetByTenantID retrieves tenant menus by tenant ID.
func (r *tenantMenuRepository) GetByTenantID(ctx context.Context, tenantID string) ([]*ent.TenantMenu, error) {
	builder := r.data.GetSlaveEntClient().TenantMenu.Query()
	builder.Where(tenantMenuEnt.TenantIDEQ(tenantID))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantMenuRepo.GetByTenantID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, tm := range rows {
			r.cacheTenantMenu(context.Background(), tm)
		}
	}()

	return rows, nil
}

// GetByMenuID retrieves tenant menus by menu ID.
func (r *tenantMenuRepository) GetByMenuID(ctx context.Context, menuID string) ([]*ent.TenantMenu, error) {
	builder := r.data.GetSlaveEntClient().TenantMenu.Query()
	builder.Where(tenantMenuEnt.MenuIDEQ(menuID))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantMenuRepo.GetByMenuID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, tm := range rows {
			r.cacheTenantMenu(context.Background(), tm)
		}
	}()

	return rows, nil
}

// DeleteByTenantIDAndMenuID deletes tenant menu by tenant ID and menu ID.
func (r *tenantMenuRepository) DeleteByTenantIDAndMenuID(ctx context.Context, tenantID, menuID string) error {
	if _, err := r.data.GetMasterEntClient().TenantMenu.Delete().
		Where(tenantMenuEnt.TenantIDEQ(tenantID), tenantMenuEnt.MenuIDEQ(menuID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "tenantMenuRepo.DeleteByTenantIDAndMenuID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateTenantMenuCache(context.Background(), tenantID, menuID)
		r.invalidateTenantMenusCache(context.Background(), tenantID)
		r.invalidateMenuTenantsCache(context.Background(), menuID)
	}()

	return nil
}

// DeleteAllByTenantID deletes all tenant menus by tenant ID.
func (r *tenantMenuRepository) DeleteAllByTenantID(ctx context.Context, tenantID string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().TenantMenu.Query().
		Where(tenantMenuEnt.TenantIDEQ(tenantID)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	if _, err := r.data.GetMasterEntClient().TenantMenu.Delete().
		Where(tenantMenuEnt.TenantIDEQ(tenantID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "tenantMenuRepo.DeleteAllByTenantID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateTenantMenusCache(context.Background(), tenantID)
		for _, tm := range relationships {
			r.invalidateTenantMenuCache(context.Background(), tm.TenantID, tm.MenuID)
			r.invalidateMenuTenantsCache(context.Background(), tm.MenuID)
		}
	}()

	return nil
}

// DeleteAllByMenuID deletes all tenant menus by menu ID.
func (r *tenantMenuRepository) DeleteAllByMenuID(ctx context.Context, menuID string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().TenantMenu.Query().
		Where(tenantMenuEnt.MenuIDEQ(menuID)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	if _, err := r.data.GetMasterEntClient().TenantMenu.Delete().
		Where(tenantMenuEnt.MenuIDEQ(menuID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "tenantMenuRepo.DeleteAllByMenuID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateMenuTenantsCache(context.Background(), menuID)
		for _, tm := range relationships {
			r.invalidateTenantMenuCache(context.Background(), tm.TenantID, tm.MenuID)
			r.invalidateTenantMenusCache(context.Background(), tm.TenantID)
		}
	}()

	return nil
}

// IsMenuInTenant verifies if a menu belongs to a tenant.
func (r *tenantMenuRepository) IsMenuInTenant(ctx context.Context, tenantID, menuID string) (bool, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("relationship:%s:%s", tenantID, menuID)
	if cached, err := r.tenantMenuCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return true, nil
	}

	count, err := r.data.GetSlaveEntClient().TenantMenu.Query().
		Where(tenantMenuEnt.TenantIDEQ(tenantID), tenantMenuEnt.MenuIDEQ(menuID)).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantMenuRepo.IsMenuInTenant error: %v", err)
		return false, err
	}

	exists := count > 0

	// Cache the result if it exists
	if exists {
		go func() {
			relationship := &ent.TenantMenu{
				TenantID: tenantID,
				MenuID:   menuID,
			}
			r.cacheTenantMenu(context.Background(), relationship)
		}()
	}

	return exists, nil
}

// GetTenantMenus retrieves all menu IDs for a tenant.
func (r *tenantMenuRepository) GetTenantMenus(ctx context.Context, tenantID string) ([]string, error) {
	// Try to get menu IDs from cache
	cacheKey := fmt.Sprintf("tenant_menus:%s", tenantID)
	var menuIDs []string
	if err := r.tenantMenusCache.GetArray(ctx, cacheKey, &menuIDs); err == nil && len(menuIDs) > 0 {
		return menuIDs, nil
	}

	// Fallback to database
	tenantMenus, err := r.data.GetSlaveEntClient().TenantMenu.Query().
		Where(tenantMenuEnt.TenantIDEQ(tenantID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantMenuRepo.GetTenantMenus error: %v", err)
		return nil, err
	}

	// Extract menu IDs
	menuIDs = make([]string, len(tenantMenus))
	for i, tm := range tenantMenus {
		menuIDs[i] = tm.MenuID
	}

	// Cache menu IDs for future use
	go func() {
		if err := r.tenantMenusCache.SetArray(context.Background(), cacheKey, menuIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache tenant menus %s: %v", tenantID, err)
		}
	}()

	return menuIDs, nil
}

// cacheTenantMenu caches a tenant menu relationship.
func (r *tenantMenuRepository) cacheTenantMenu(ctx context.Context, tm *ent.TenantMenu) {
	relationshipKey := fmt.Sprintf("relationship:%s:%s", tm.TenantID, tm.MenuID)
	if err := r.tenantMenuCache.Set(ctx, relationshipKey, tm, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache tenant menu relationship %s:%s: %v", tm.TenantID, tm.MenuID, err)
	}
}

// invalidateTenantMenuCache invalidates tenant menu cache
func (r *tenantMenuRepository) invalidateTenantMenuCache(ctx context.Context, tenantID, menuID string) {
	relationshipKey := fmt.Sprintf("relationship:%s:%s", tenantID, menuID)
	if err := r.tenantMenuCache.Delete(ctx, relationshipKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate tenant menu relationship cache %s:%s: %v", tenantID, menuID, err)
	}
}

// invalidateTenantMenusCache invalidates tenant menus cache
func (r *tenantMenuRepository) invalidateTenantMenusCache(ctx context.Context, tenantID string) {
	cacheKey := fmt.Sprintf("tenant_menus:%s", tenantID)
	if err := r.tenantMenusCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate tenant menus cache %s: %v", tenantID, err)
	}
}

// invalidateMenuTenantsCache invalidates menu tenants cache
func (r *tenantMenuRepository) invalidateMenuTenantsCache(ctx context.Context, menuID string) {
	cacheKey := fmt.Sprintf("menu_tenants:%s", menuID)
	if err := r.menuTenantsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate menu tenants cache %s: %v", menuID, err)
	}
}
