package repository

import (
	"context"
	"fmt"
	"ncobase/tenant/data"
	"ncobase/tenant/data/ent"
	tenantGroupEnt "ncobase/tenant/data/ent/tenantgroup"
	"ncobase/tenant/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/logging/logger"
)

// TenantGroupRepositoryInterface represents the tenant group repository interface.
type TenantGroupRepositoryInterface interface {
	Create(ctx context.Context, body *structs.TenantGroup) (*ent.TenantGroup, error)
	GetByTenantID(ctx context.Context, id string) ([]*ent.TenantGroup, error)
	GetByGroupID(ctx context.Context, id string) ([]*ent.TenantGroup, error)
	GetByTenantIDs(ctx context.Context, ids []string) ([]*ent.TenantGroup, error)
	GetByGroupIDs(ctx context.Context, ids []string) ([]*ent.TenantGroup, error)
	Delete(ctx context.Context, tid, gid string) error
	DeleteAllByTenantID(ctx context.Context, id string) error
	DeleteAllByGroupID(ctx context.Context, id string) error
	GetGroupsByTenantID(ctx context.Context, tenantID string) ([]string, error)
	GetTenantsByGroupID(ctx context.Context, groupID string) ([]string, error)
	IsGroupInTenant(ctx context.Context, tenantID string, groupID string) (bool, error)
}

// tenantGroupRepository implements the TenantGroupRepositoryInterface.
type tenantGroupRepository struct {
	data              *data.Data
	tenantGroupCache  cache.ICache[ent.TenantGroup]
	tenantGroupsCache cache.ICache[[]string] // Maps tenant ID to group IDs
	groupTenantsCache cache.ICache[[]string] // Maps group ID to tenant IDs
	relationshipTTL   time.Duration
}

// NewTenantGroupRepository creates a new tenant group repository.
func NewTenantGroupRepository(d *data.Data) TenantGroupRepositoryInterface {
	redisClient := d.GetRedis()

	return &tenantGroupRepository{
		data:              d,
		tenantGroupCache:  cache.NewCache[ent.TenantGroup](redisClient, "ncse_tenant:tenant_groups"),
		tenantGroupsCache: cache.NewCache[[]string](redisClient, "ncse_tenant:tenant_group_mappings"),
		groupTenantsCache: cache.NewCache[[]string](redisClient, "ncse_tenant:group_tenant_mappings"),
		relationshipTTL:   time.Hour * 2, // 2 hours cache TTL
	}
}

// Create creates tenant group relationship
func (r *tenantGroupRepository) Create(ctx context.Context, body *structs.TenantGroup) (*ent.TenantGroup, error) {
	builder := r.data.GetMasterEntClient().TenantGroup.Create()

	builder.SetTenantID(body.TenantID)
	builder.SetGroupID(body.GroupID)

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantGroupRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the relationship and invalidate related caches
	go func() {
		r.cacheTenantGroup(context.Background(), row)
		r.invalidateTenantGroupsCache(context.Background(), body.TenantID)
		r.invalidateGroupTenantsCache(context.Background(), body.GroupID)
	}()

	return row, nil
}

// GetByTenantID finds groups by tenant id
func (r *tenantGroupRepository) GetByTenantID(ctx context.Context, id string) ([]*ent.TenantGroup, error) {
	builder := r.data.GetSlaveEntClient().TenantGroup.Query()
	builder.Where(tenantGroupEnt.TenantIDEQ(id))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantGroupRepo.GetByTenantID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, tg := range rows {
			r.cacheTenantGroup(context.Background(), tg)
		}
	}()

	return rows, nil
}

// GetByGroupID finds tenants by group id
func (r *tenantGroupRepository) GetByGroupID(ctx context.Context, id string) ([]*ent.TenantGroup, error) {
	builder := r.data.GetSlaveEntClient().TenantGroup.Query()
	builder.Where(tenantGroupEnt.GroupIDEQ(id))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantGroupRepo.GetByGroupID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, tg := range rows {
			r.cacheTenantGroup(context.Background(), tg)
		}
	}()

	return rows, nil
}

// GetByTenantIDs finds groups by tenant ids
func (r *tenantGroupRepository) GetByTenantIDs(ctx context.Context, ids []string) ([]*ent.TenantGroup, error) {
	builder := r.data.GetSlaveEntClient().TenantGroup.Query()
	builder.Where(tenantGroupEnt.TenantIDIn(ids...))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantGroupRepo.GetByTenantIDs error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, tg := range rows {
			r.cacheTenantGroup(context.Background(), tg)
		}
	}()

	return rows, nil
}

// GetByGroupIDs finds tenants by group ids
func (r *tenantGroupRepository) GetByGroupIDs(ctx context.Context, ids []string) ([]*ent.TenantGroup, error) {
	builder := r.data.GetSlaveEntClient().TenantGroup.Query()
	builder.Where(tenantGroupEnt.GroupIDIn(ids...))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantGroupRepo.GetByGroupIDs error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, tg := range rows {
			r.cacheTenantGroup(context.Background(), tg)
		}
	}()

	return rows, nil
}

// Delete deletes tenant group relationship
func (r *tenantGroupRepository) Delete(ctx context.Context, tid, gid string) error {
	if _, err := r.data.GetMasterEntClient().TenantGroup.Delete().
		Where(tenantGroupEnt.TenantIDEQ(tid), tenantGroupEnt.GroupIDEQ(gid)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "tenantGroupRepo.Delete error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateTenantGroupCache(context.Background(), tid, gid)
		r.invalidateTenantGroupsCache(context.Background(), tid)
		r.invalidateGroupTenantsCache(context.Background(), gid)
	}()

	return nil
}

// DeleteAllByTenantID deletes all tenant group by tenant id
func (r *tenantGroupRepository) DeleteAllByTenantID(ctx context.Context, id string) error {
	relationships, err := r.GetByTenantID(ctx, id)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	if _, err := r.data.GetMasterEntClient().TenantGroup.Delete().
		Where(tenantGroupEnt.TenantIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "tenantGroupRepo.DeleteAllByTenantID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateTenantGroupsCache(context.Background(), id)
		for _, tg := range relationships {
			r.invalidateTenantGroupCache(context.Background(), tg.TenantID, tg.GroupID)
			r.invalidateGroupTenantsCache(context.Background(), tg.GroupID)
		}
	}()

	return nil
}

// DeleteAllByGroupID deletes all tenant group by group id
func (r *tenantGroupRepository) DeleteAllByGroupID(ctx context.Context, id string) error {
	relationships, err := r.GetByGroupID(ctx, id)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	if _, err := r.data.GetMasterEntClient().TenantGroup.Delete().
		Where(tenantGroupEnt.GroupIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "tenantGroupRepo.DeleteAllByGroupID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateGroupTenantsCache(context.Background(), id)
		for _, tg := range relationships {
			r.invalidateTenantGroupCache(context.Background(), tg.TenantID, tg.GroupID)
			r.invalidateTenantGroupsCache(context.Background(), tg.TenantID)
		}
	}()

	return nil
}

// GetGroupsByTenantID retrieves all groups under a tenant
func (r *tenantGroupRepository) GetGroupsByTenantID(ctx context.Context, tenantID string) ([]string, error) {
	cacheKey := fmt.Sprintf("tenant_groups:%s", tenantID)
	var groupIDs []string
	if err := r.tenantGroupsCache.GetArray(ctx, cacheKey, &groupIDs); err == nil && len(groupIDs) > 0 {
		return groupIDs, nil
	}

	// Fallback to database
	tenantGroups, err := r.data.GetSlaveEntClient().TenantGroup.Query().
		Where(tenantGroupEnt.TenantIDEQ(tenantID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantGroupRepo.GetGroupsByTenantID error: %v", err)
		return nil, err
	}

	groupIDs = make([]string, len(tenantGroups))
	for i, tg := range tenantGroups {
		groupIDs[i] = tg.GroupID
	}

	// Cache for future use
	go func() {
		if err := r.tenantGroupsCache.SetArray(context.Background(), cacheKey, groupIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache tenant groups %s: %v", tenantID, err)
		}
	}()

	return groupIDs, nil
}

// GetTenantsByGroupID retrieves all tenants that have a group
func (r *tenantGroupRepository) GetTenantsByGroupID(ctx context.Context, groupID string) ([]string, error) {
	cacheKey := fmt.Sprintf("group_tenants:%s", groupID)
	var tenantIDs []string
	if err := r.groupTenantsCache.GetArray(ctx, cacheKey, &tenantIDs); err == nil && len(tenantIDs) > 0 {
		return tenantIDs, nil
	}

	// Fallback to database
	tenantGroups, err := r.data.GetSlaveEntClient().TenantGroup.Query().
		Where(tenantGroupEnt.GroupIDEQ(groupID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantGroupRepo.GetTenantsByGroupID error: %v", err)
		return nil, err
	}

	tenantIDs = make([]string, len(tenantGroups))
	for i, tg := range tenantGroups {
		tenantIDs[i] = tg.TenantID
	}

	// Cache for future use
	go func() {
		if err := r.groupTenantsCache.SetArray(context.Background(), cacheKey, tenantIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache group tenants %s: %v", groupID, err)
		}
	}()

	return tenantIDs, nil
}

// IsGroupInTenant verifies if a group belongs to a specific tenant
func (r *tenantGroupRepository) IsGroupInTenant(ctx context.Context, tenantID string, groupID string) (bool, error) {
	cacheKey := fmt.Sprintf("relationship:%s:%s", tenantID, groupID)
	if cached, err := r.tenantGroupCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return true, nil
	}

	count, err := r.data.GetSlaveEntClient().TenantGroup.Query().
		Where(tenantGroupEnt.TenantIDEQ(tenantID), tenantGroupEnt.GroupIDEQ(groupID)).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantGroupRepo.IsGroupInTenant error: %v", err)
		return false, err
	}

	exists := count > 0

	if exists {
		go func() {
			// Create dummy relationship for caching
			relationship := &ent.TenantGroup{
				TenantID: tenantID,
				GroupID:  groupID,
			}
			r.cacheTenantGroup(context.Background(), relationship)
		}()
	}

	return exists, nil
}

// Cache management methods
func (r *tenantGroupRepository) cacheTenantGroup(ctx context.Context, tg *ent.TenantGroup) {
	relationshipKey := fmt.Sprintf("relationship:%s:%s", tg.TenantID, tg.GroupID)
	if err := r.tenantGroupCache.Set(ctx, relationshipKey, tg, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache tenant group relationship %s:%s: %v", tg.TenantID, tg.GroupID, err)
	}
}

func (r *tenantGroupRepository) invalidateTenantGroupCache(ctx context.Context, tenantID, groupID string) {
	relationshipKey := fmt.Sprintf("relationship:%s:%s", tenantID, groupID)
	if err := r.tenantGroupCache.Delete(ctx, relationshipKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate tenant group relationship cache %s:%s: %v", tenantID, groupID, err)
	}
}

func (r *tenantGroupRepository) invalidateTenantGroupsCache(ctx context.Context, tenantID string) {
	cacheKey := fmt.Sprintf("tenant_groups:%s", tenantID)
	if err := r.tenantGroupsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate tenant groups cache %s: %v", tenantID, err)
	}
}

func (r *tenantGroupRepository) invalidateGroupTenantsCache(ctx context.Context, groupID string) {
	cacheKey := fmt.Sprintf("group_tenants:%s", groupID)
	if err := r.groupTenantsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate group tenants cache %s: %v", groupID, err)
	}
}
