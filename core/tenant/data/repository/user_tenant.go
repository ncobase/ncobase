package repository

import (
	"context"
	"fmt"
	"ncobase/tenant/data"
	"ncobase/tenant/data/ent"
	tenantEnt "ncobase/tenant/data/ent/tenant"
	userTenantEnt "ncobase/tenant/data/ent/usertenant"
	"ncobase/tenant/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/logging/logger"
)

// UserTenantRepositoryInterface represents the user tenant repository interface.
type UserTenantRepositoryInterface interface {
	Create(ctx context.Context, body *structs.UserTenant) (*ent.UserTenant, error)
	GetByUserID(ctx context.Context, id string) (*ent.UserTenant, error)
	GetByTenantID(ctx context.Context, id string) (*ent.UserTenant, error)
	GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserTenant, error)
	GetByTenantIDs(ctx context.Context, ids []string) ([]*ent.UserTenant, error)
	Delete(ctx context.Context, uid, did string) error
	DeleteAllByUserID(ctx context.Context, id string) error
	DeleteAllByTenantID(ctx context.Context, id string) error
	GetTenantsByUserID(ctx context.Context, userID string) ([]*ent.Tenant, error)
	IsTenantInUser(ctx context.Context, tenantID, userID string) (bool, error)
}

// userTenantRepository implements the UserTenantRepositoryInterface.
type userTenantRepository struct {
	data             *data.Data
	userTenantCache  cache.ICache[ent.UserTenant]
	userTenantsCache cache.ICache[[]string] // Maps user ID to tenant IDs
	tenantUsersCache cache.ICache[[]string] // Maps tenant ID to user IDs
	relationshipTTL  time.Duration
}

// NewUserTenantRepository creates a new user tenant repository.
func NewUserTenantRepository(d *data.Data) UserTenantRepositoryInterface {
	redisClient := d.GetRedis()

	return &userTenantRepository{
		data:             d,
		userTenantCache:  cache.NewCache[ent.UserTenant](redisClient, "ncse_tenant:user_tenants"),
		userTenantsCache: cache.NewCache[[]string](redisClient, "ncse_tenant:user_tenant_mappings"),
		tenantUsersCache: cache.NewCache[[]string](redisClient, "ncse_tenant:tenant_user_mappings"),
		relationshipTTL:  time.Hour * 3, // 3 hours cache TTL (tenant relationships change less frequently)
	}
}

// Create creates a new user tenant
func (r *userTenantRepository) Create(ctx context.Context, body *structs.UserTenant) (*ent.UserTenant, error) {
	// Use master for writes
	builder := r.data.GetMasterEntClient().UserTenant.Create()

	// Set values
	builder.SetNillableUserID(&body.UserID)
	builder.SetNillableTenantID(&body.TenantID)

	// Execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the relationship and invalidate related caches
	go func() {
		r.cacheUserTenant(context.Background(), row)
		r.invalidateUserTenantsCache(context.Background(), body.UserID)
		r.invalidateTenantUsersCache(context.Background(), body.TenantID)
	}()

	return row, nil
}

// GetByUserID find tenant by user id
func (r *userTenantRepository) GetByUserID(ctx context.Context, id string) (*ent.UserTenant, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("user:%s", id)
	if cached, err := r.userTenantCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserTenant.Query()

	// Set conditions
	builder.Where(userTenantEnt.UserIDEQ(id))

	// Execute the builder
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRepo.GetByUserID error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheUserTenant(context.Background(), row)

	return row, nil
}

// GetByUserIDs find tenants by user ids
func (r *userTenantRepository) GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserTenant, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserTenant.Query()

	// Set conditions
	builder.Where(userTenantEnt.UserIDIn(ids...))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRepo.GetByUserIDs error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, ut := range rows {
			r.cacheUserTenant(context.Background(), ut)
		}
	}()

	return rows, nil
}

// GetByTenantID find tenant by tenant id
func (r *userTenantRepository) GetByTenantID(ctx context.Context, id string) (*ent.UserTenant, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("tenant:%s", id)
	if cached, err := r.userTenantCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserTenant.Query()

	// Set conditions
	builder.Where(userTenantEnt.TenantIDEQ(id))

	// Execute the builder
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRepo.GetByTenantID error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheUserTenant(context.Background(), row)

	return row, nil
}

// GetByTenantIDs find tenants by tenant ids
func (r *userTenantRepository) GetByTenantIDs(ctx context.Context, ids []string) ([]*ent.UserTenant, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserTenant.Query()

	// Set conditions
	builder.Where(userTenantEnt.TenantIDIn(ids...))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRepo.GetByTenantIDs error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, ut := range rows {
			r.cacheUserTenant(context.Background(), ut)
		}
	}()

	return rows, nil
}

// Delete delete user tenant
func (r *userTenantRepository) Delete(ctx context.Context, uid, did string) error {
	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserTenant.Delete().
		Where(userTenantEnt.UserIDEQ(uid), userTenantEnt.TenantIDEQ(did)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userTenantRepo.Delete error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateUserTenantCache(context.Background(), uid, did)
		r.invalidateUserTenantsCache(context.Background(), uid)
		r.invalidateTenantUsersCache(context.Background(), did)
	}()

	return nil
}

// DeleteAllByUserID delete all user tenant
func (r *userTenantRepository) DeleteAllByUserID(ctx context.Context, id string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.GetByUserIDs(ctx, []string{id})
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserTenant.Delete().
		Where(userTenantEnt.UserIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userTenantRepo.DeleteAllByUserID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateUserTenantsCache(context.Background(), id)
		for _, ut := range relationships {
			r.invalidateUserTenantCache(context.Background(), ut.UserID, ut.TenantID)
			r.invalidateTenantUsersCache(context.Background(), ut.TenantID)
		}
	}()

	return nil
}

// DeleteAllByTenantID delete all user tenant
func (r *userTenantRepository) DeleteAllByTenantID(ctx context.Context, id string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.GetByTenantIDs(ctx, []string{id})
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserTenant.Delete().
		Where(userTenantEnt.TenantIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userTenantRepo.DeleteAllByTenantID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateTenantUsersCache(context.Background(), id)
		for _, ut := range relationships {
			r.invalidateUserTenantCache(context.Background(), ut.UserID, ut.TenantID)
			r.invalidateUserTenantsCache(context.Background(), ut.UserID)
		}
	}()

	return nil
}

// GetTenantsByUserID retrieves all tenants a user belongs to.
func (r *userTenantRepository) GetTenantsByUserID(ctx context.Context, userID string) ([]*ent.Tenant, error) {
	// Try to get tenant IDs from cache
	cacheKey := fmt.Sprintf("user_tenants:%s", userID)
	var tenantIDs []string
	if err := r.userTenantsCache.GetArray(ctx, cacheKey, &tenantIDs); err == nil && len(tenantIDs) > 0 {
		// Get tenants by IDs from tenant repository
		return r.data.GetSlaveEntClient().Tenant.Query().Where(tenantEnt.IDIn(tenantIDs...)).All(ctx)
	}

	// Fallback to database
	userTenants, err := r.data.GetSlaveEntClient().UserTenant.Query().
		Where(userTenantEnt.UserIDEQ(userID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRepo.GetTenantsByUserID error: %v", err)
		return nil, err
	}

	// Extract tenant IDs from userTenants
	tenantIDs = make([]string, len(userTenants))
	for i, userTenant := range userTenants {
		tenantIDs[i] = userTenant.TenantID
	}

	// Query tenants based on extracted tenant IDs
	tenants, err := r.data.GetSlaveEntClient().Tenant.Query().Where(tenantEnt.IDIn(tenantIDs...)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRepo.GetTenantsByUserID error: %v", err)
		return nil, err
	}

	// Cache tenant IDs for future use
	go func() {
		if err := r.userTenantsCache.SetArray(context.Background(), cacheKey, tenantIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache user tenants %s: %v", userID, err)
		}
	}()

	return tenants, nil
}

// IsUserInTenant verifies if a user belongs to a specific tenant.
func (r *userTenantRepository) IsUserInTenant(ctx context.Context, userID string, tenantID string) (bool, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("relationship:%s:%s", userID, tenantID)
	if cached, err := r.userTenantCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return true, nil
	}

	// Use slave for reads
	count, err := r.data.GetSlaveEntClient().UserTenant.Query().
		Where(userTenantEnt.UserIDEQ(userID), userTenantEnt.TenantIDEQ(tenantID)).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRepo.IsUserInTenant error: %v", err)
		return false, err
	}

	exists := count > 0

	// Cache the result if it exists
	if exists {
		go func() {
			// Create a dummy relationship for caching
			relationship := &ent.UserTenant{
				UserID:   userID,
				TenantID: tenantID,
			}
			r.cacheUserTenant(context.Background(), relationship)
		}()
	}

	return exists, nil
}

// IsTenantInUser verifies if a tenant is assigned to a specific user.
func (r *userTenantRepository) IsTenantInUser(ctx context.Context, tenantID, userID string) (bool, error) {
	return r.IsUserInTenant(ctx, userID, tenantID)
}

// cacheUserTenant caches a user-tenant relationship.
func (r *userTenantRepository) cacheUserTenant(ctx context.Context, ut *ent.UserTenant) {
	// Cache by relationship key
	relationshipKey := fmt.Sprintf("relationship:%s:%s", ut.UserID, ut.TenantID)
	if err := r.userTenantCache.Set(ctx, relationshipKey, ut, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache user tenant relationship %s:%s: %v", ut.UserID, ut.TenantID, err)
	}

	// Cache by user ID
	userKey := fmt.Sprintf("user:%s", ut.UserID)
	if err := r.userTenantCache.Set(ctx, userKey, ut, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache user tenant by user %s: %v", ut.UserID, err)
	}

	// Cache by tenant ID
	tenantKey := fmt.Sprintf("tenant:%s", ut.TenantID)
	if err := r.userTenantCache.Set(ctx, tenantKey, ut, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache user tenant by tenant %s: %v", ut.TenantID, err)
	}
}

// invalidateUserTenantCache invalidates the cache for a user-tenant relationship.
func (r *userTenantRepository) invalidateUserTenantCache(ctx context.Context, userID, tenantID string) {
	// Invalidate relationship cache
	relationshipKey := fmt.Sprintf("relationship:%s:%s", userID, tenantID)
	if err := r.userTenantCache.Delete(ctx, relationshipKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user tenant relationship cache %s:%s: %v", userID, tenantID, err)
	}

	// Invalidate user key
	userKey := fmt.Sprintf("user:%s", userID)
	if err := r.userTenantCache.Delete(ctx, userKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user tenant cache by user %s: %v", userID, err)
	}

	// Invalidate tenant key
	tenantKey := fmt.Sprintf("tenant:%s", tenantID)
	if err := r.userTenantCache.Delete(ctx, tenantKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user tenant cache by tenant %s: %v", tenantID, err)
	}
}

// invalidateUserTenantsCache invalidates the cache for all user tenants.
func (r *userTenantRepository) invalidateUserTenantsCache(ctx context.Context, userID string) {
	cacheKey := fmt.Sprintf("user_tenants:%s", userID)
	if err := r.userTenantsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user tenants cache %s: %v", userID, err)
	}
}

// invalidateTenantUsersCache invalidates the cache for all tenant users.
func (r *userTenantRepository) invalidateTenantUsersCache(ctx context.Context, tenantID string) {
	cacheKey := fmt.Sprintf("tenant_users:%s", tenantID)
	if err := r.tenantUsersCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate tenant users cache %s: %v", tenantID, err)
	}
}
