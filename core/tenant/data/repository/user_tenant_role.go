package repository

import (
	"context"
	"fmt"
	"ncobase/tenant/data"
	"ncobase/tenant/data/ent"
	userTenantRoleEnt "ncobase/tenant/data/ent/usertenantrole"
	"ncobase/tenant/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/logging/logger"
)

// UserTenantRoleRepositoryInterface represents the user tenant role repository interface.
type UserTenantRoleRepositoryInterface interface {
	Create(ctx context.Context, body *structs.UserTenantRole) (*ent.UserTenantRole, error)
	GetByUserID(ctx context.Context, u string) (*ent.UserTenantRole, error)
	GetByTenantID(ctx context.Context, t string) ([]*ent.UserTenantRole, error)
	GetByRoleID(ctx context.Context, r string) ([]*ent.UserTenantRole, error)
	DeleteByUserIDAndTenantID(ctx context.Context, u, t string) error
	DeleteByUserIDAndRoleID(ctx context.Context, u, r string) error
	DeleteByTenantIDAndRoleID(ctx context.Context, t, r string) error
	DeleteByUserIDAndTenantIDAndRoleID(ctx context.Context, u, t, r string) error
	DeleteAllByUserID(ctx context.Context, u string) error
	DeleteAllByTenantID(ctx context.Context, t string) error
	DeleteAllByRoleID(ctx context.Context, r string) error
	GetRolesByUserAndTenant(ctx context.Context, u, t string) ([]string, error)
	IsUserInRoleInTenant(ctx context.Context, u, t, r string) (bool, error)
}

// userTenantRoleRepository implements the UserTenantRoleRepositoryInterface.
type userTenantRoleRepository struct {
	data                 *data.Data
	userTenantRoleCache  cache.ICache[ent.UserTenantRole]
	userTenantRolesCache cache.ICache[[]string] // Maps user:tenant to role IDs
	tenantUserRolesCache cache.ICache[[]string] // Maps tenant to user:role pairs
	roleUserTenantsCache cache.ICache[[]string] // Maps role to user:tenant pairs
	relationshipTTL      time.Duration
}

// NewUserTenantRoleRepository creates a new user tenant role repository.
func NewUserTenantRoleRepository(d *data.Data) UserTenantRoleRepositoryInterface {
	redisClient := d.GetRedis()

	return &userTenantRoleRepository{
		data:                 d,
		userTenantRoleCache:  cache.NewCache[ent.UserTenantRole](redisClient, "ncse_access:user_tenant_roles"),
		userTenantRolesCache: cache.NewCache[[]string](redisClient, "ncse_access:user_tenant_role_mappings"),
		tenantUserRolesCache: cache.NewCache[[]string](redisClient, "ncse_access:tenant_user_role_mappings"),
		roleUserTenantsCache: cache.NewCache[[]string](redisClient, "ncse_access:role_user_tenant_mappings"),
		relationshipTTL:      time.Hour * 2, // 2 hours cache TTL
	}
}

// Create creates a new user tenant role.
func (r *userTenantRoleRepository) Create(ctx context.Context, body *structs.UserTenantRole) (*ent.UserTenantRole, error) {
	// Use master for writes
	builder := r.data.GetMasterEntClient().UserTenantRole.Create()

	// Set values
	builder.SetNillableUserID(&body.UserID)
	builder.SetNillableTenantID(&body.TenantID)
	builder.SetNillableRoleID(&body.RoleID)

	// Execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the relationship and invalidate related caches
	go func() {
		r.cacheUserTenantRole(context.Background(), row)
		r.invalidateUserTenantRolesCache(context.Background(), body.UserID, body.TenantID)
		r.invalidateTenantUserRolesCache(context.Background(), body.TenantID)
		r.invalidateRoleUserTenantsCache(context.Background(), body.RoleID)
	}()

	return row, nil
}

// GetByUserID retrieves user tenant role by user ID.
func (r *userTenantRoleRepository) GetByUserID(ctx context.Context, u string) (*ent.UserTenantRole, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("user:%s", u)
	if cached, err := r.userTenantRoleCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserTenantRole.Query()

	// Set conditions
	builder.Where(userTenantRoleEnt.UserIDEQ(u))

	// Execute the builder
	row, err := builder.First(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.GetByUserID error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheUserTenantRole(context.Background(), row)

	return row, nil
}

// GetByTenantID retrieves user tenant roles by tenant ID.
func (r *userTenantRoleRepository) GetByTenantID(ctx context.Context, t string) ([]*ent.UserTenantRole, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserTenantRole.Query()

	// Set conditions
	builder.Where(userTenantRoleEnt.TenantID(t))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.GetByTenantID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, utr := range rows {
			r.cacheUserTenantRole(context.Background(), utr)
		}
	}()

	return rows, nil
}

// GetByRoleID retrieves user tenant roles by role ID.
func (r *userTenantRoleRepository) GetByRoleID(ctx context.Context, rid string) ([]*ent.UserTenantRole, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserTenantRole.Query()

	// Set conditions
	builder.Where(userTenantRoleEnt.RoleID(rid))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.GetByRoleID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, utr := range rows {
			r.cacheUserTenantRole(context.Background(), utr)
		}
	}()

	return rows, nil
}

// DeleteByUserIDAndTenantID deletes user tenant role by user ID and tenant ID.
func (r *userTenantRoleRepository) DeleteByUserIDAndTenantID(ctx context.Context, u, t string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().UserTenantRole.Query().
		Where(userTenantRoleEnt.UserIDEQ(u), userTenantRoleEnt.TenantID(t)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserTenantRole.Delete().
		Where(userTenantRoleEnt.UserIDEQ(u), userTenantRoleEnt.TenantID(t)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.DeleteByUserIDAndTenantID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateUserTenantRolesCache(context.Background(), u, t)
		r.invalidateTenantUserRolesCache(context.Background(), t)
		for _, utr := range relationships {
			r.invalidateUserTenantRoleCache(context.Background(), utr)
			r.invalidateRoleUserTenantsCache(context.Background(), utr.RoleID)
		}
	}()

	return nil
}

// DeleteByUserIDAndRoleID deletes user tenant role by user ID and role ID.
func (r *userTenantRoleRepository) DeleteByUserIDAndRoleID(ctx context.Context, u, rid string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().UserTenantRole.Query().
		Where(userTenantRoleEnt.UserIDEQ(u), userTenantRoleEnt.RoleID(rid)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserTenantRole.Delete().
		Where(userTenantRoleEnt.UserIDEQ(u), userTenantRoleEnt.RoleID(rid)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.DeleteByUserIDAndRoleID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateRoleUserTenantsCache(context.Background(), rid)
		for _, utr := range relationships {
			r.invalidateUserTenantRoleCache(context.Background(), utr)
			r.invalidateUserTenantRolesCache(context.Background(), u, utr.TenantID)
			r.invalidateTenantUserRolesCache(context.Background(), utr.TenantID)
		}
	}()

	return nil
}

// DeleteByTenantIDAndRoleID deletes user tenant role by tenant ID and role ID.
func (r *userTenantRoleRepository) DeleteByTenantIDAndRoleID(ctx context.Context, t, rid string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().UserTenantRole.Query().
		Where(userTenantRoleEnt.TenantID(t), userTenantRoleEnt.RoleID(rid)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserTenantRole.Delete().
		Where(userTenantRoleEnt.TenantID(t), userTenantRoleEnt.RoleID(rid)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.DeleteByTenantIDAndRoleID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateTenantUserRolesCache(context.Background(), t)
		r.invalidateRoleUserTenantsCache(context.Background(), rid)
		for _, utr := range relationships {
			r.invalidateUserTenantRoleCache(context.Background(), utr)
			r.invalidateUserTenantRolesCache(context.Background(), utr.UserID, t)
		}
	}()

	return nil
}

// DeleteByUserIDAndTenantIDAndRoleID deletes user tenant role by user ID, tenant ID and role ID.
func (r *userTenantRoleRepository) DeleteByUserIDAndTenantIDAndRoleID(ctx context.Context, u, t, rid string) error {
	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserTenantRole.Delete().
		Where(userTenantRoleEnt.UserIDEQ(u), userTenantRoleEnt.TenantID(t), userTenantRoleEnt.RoleID(rid)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.DeleteByUserIDAndTenantIDAndRoleID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		// Create a dummy relationship for cache invalidation
		utr := &ent.UserTenantRole{
			UserID:   u,
			TenantID: t,
			RoleID:   rid,
		}
		r.invalidateUserTenantRoleCache(context.Background(), utr)
		r.invalidateUserTenantRolesCache(context.Background(), u, t)
		r.invalidateTenantUserRolesCache(context.Background(), t)
		r.invalidateRoleUserTenantsCache(context.Background(), rid)
	}()

	return nil
}

// DeleteAllByUserID deletes all user tenant roles by user ID.
func (r *userTenantRoleRepository) DeleteAllByUserID(ctx context.Context, u string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().UserTenantRole.Query().
		Where(userTenantRoleEnt.UserIDEQ(u)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserTenantRole.Delete().
		Where(userTenantRoleEnt.UserIDEQ(u)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.DeleteAllByUserID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		for _, utr := range relationships {
			r.invalidateUserTenantRoleCache(context.Background(), utr)
			r.invalidateUserTenantRolesCache(context.Background(), u, utr.TenantID)
			r.invalidateTenantUserRolesCache(context.Background(), utr.TenantID)
			r.invalidateRoleUserTenantsCache(context.Background(), utr.RoleID)
		}
	}()

	return nil
}

// DeleteAllByTenantID deletes all user tenant roles by tenant ID.
func (r *userTenantRoleRepository) DeleteAllByTenantID(ctx context.Context, t string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().UserTenantRole.Query().
		Where(userTenantRoleEnt.TenantID(t)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserTenantRole.Delete().
		Where(userTenantRoleEnt.TenantID(t)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.DeleteAllByTenantID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateTenantUserRolesCache(context.Background(), t)
		for _, utr := range relationships {
			r.invalidateUserTenantRoleCache(context.Background(), utr)
			r.invalidateUserTenantRolesCache(context.Background(), utr.UserID, t)
			r.invalidateRoleUserTenantsCache(context.Background(), utr.RoleID)
		}
	}()

	return nil
}

// DeleteAllByRoleID deletes all user tenant roles by role ID.
func (r *userTenantRoleRepository) DeleteAllByRoleID(ctx context.Context, rid string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().UserTenantRole.Query().
		Where(userTenantRoleEnt.RoleID(rid)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserTenantRole.Delete().
		Where(userTenantRoleEnt.RoleID(rid)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.DeleteAllByRoleID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateRoleUserTenantsCache(context.Background(), rid)
		for _, utr := range relationships {
			r.invalidateUserTenantRoleCache(context.Background(), utr)
			r.invalidateUserTenantRolesCache(context.Background(), utr.UserID, utr.TenantID)
			r.invalidateTenantUserRolesCache(context.Background(), utr.TenantID)
		}
	}()

	return nil
}

// GetRolesByUserAndTenant retrieves all roles a user has in a tenant.
func (r *userTenantRoleRepository) GetRolesByUserAndTenant(ctx context.Context, u string, t string) ([]string, error) {
	// Try to get role IDs from cache
	cacheKey := fmt.Sprintf("user_tenant_roles:%s:%s", u, t)
	var roleIDs []string
	if err := r.userTenantRolesCache.GetArray(ctx, cacheKey, &roleIDs); err == nil && len(roleIDs) > 0 {
		return roleIDs, nil
	}

	// Fallback to database
	userTenantRoles, err := r.data.GetSlaveEntClient().UserTenantRole.Query().
		Where(userTenantRoleEnt.UserIDEQ(u), userTenantRoleEnt.TenantIDEQ(t)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.GetRolesByUserAndTenant error: %v", err)
		return nil, err
	}

	// Extract role IDs from userTenantRoles
	roleIDs = make([]string, len(userTenantRoles))
	for i, userRole := range userTenantRoles {
		roleIDs[i] = userRole.RoleID
	}

	// Cache role IDs for future use
	go func() {
		if err := r.userTenantRolesCache.SetArray(context.Background(), cacheKey, roleIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache user tenant roles %s:%s: %v", u, t, err)
		}
	}()

	return roleIDs, nil
}

// IsUserInRoleInTenant verifies if a user has a specific role in a tenant.
func (r *userTenantRoleRepository) IsUserInRoleInTenant(ctx context.Context, u, t, rid string) (bool, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("relationship:%s:%s:%s", u, t, rid)
	if cached, err := r.userTenantRoleCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return true, nil
	}

	// Use slave for reads
	count, err := r.data.GetSlaveEntClient().UserTenantRole.Query().
		Where(userTenantRoleEnt.UserIDEQ(u), userTenantRoleEnt.TenantIDEQ(t), userTenantRoleEnt.RoleIDEQ(rid)).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.IsUserInRoleInTenant error: %v", err)
		return false, err
	}

	exists := count > 0

	// Cache the result if it exists
	if exists {
		go func() {
			// Create a dummy relationship for caching
			relationship := &ent.UserTenantRole{
				UserID:   u,
				TenantID: t,
				RoleID:   rid,
			}
			r.cacheUserTenantRole(context.Background(), relationship)
		}()
	}

	return exists, nil
}

// cacheUserTenantRole caches a user tenant role relationship.
func (r *userTenantRoleRepository) cacheUserTenantRole(ctx context.Context, utr *ent.UserTenantRole) {
	// Cache by relationship key
	relationshipKey := fmt.Sprintf("relationship:%s:%s:%s", utr.UserID, utr.TenantID, utr.RoleID)
	if err := r.userTenantRoleCache.Set(ctx, relationshipKey, utr, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache user tenant role relationship %s:%s:%s: %v", utr.UserID, utr.TenantID, utr.RoleID, err)
	}

	// Cache by user ID
	userKey := fmt.Sprintf("user:%s", utr.UserID)
	if err := r.userTenantRoleCache.Set(ctx, userKey, utr, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache user tenant role by user %s: %v", utr.UserID, err)
	}
}

// invalidateUserTenantRoleCache invalidates user tenant role cache
func (r *userTenantRoleRepository) invalidateUserTenantRoleCache(ctx context.Context, utr *ent.UserTenantRole) {
	// Invalidate relationship cache
	relationshipKey := fmt.Sprintf("relationship:%s:%s:%s", utr.UserID, utr.TenantID, utr.RoleID)
	if err := r.userTenantRoleCache.Delete(ctx, relationshipKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user tenant role relationship cache %s:%s:%s: %v", utr.UserID, utr.TenantID, utr.RoleID, err)
	}

	// Invalidate user key
	userKey := fmt.Sprintf("user:%s", utr.UserID)
	if err := r.userTenantRoleCache.Delete(ctx, userKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user tenant role cache by user %s: %v", utr.UserID, err)
	}
}

// invalidateUserTenantRolesCache invalidates user tenant roles cache
func (r *userTenantRoleRepository) invalidateUserTenantRolesCache(ctx context.Context, userID, tenantID string) {
	cacheKey := fmt.Sprintf("user_tenant_roles:%s:%s", userID, tenantID)
	if err := r.userTenantRolesCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user tenant roles cache %s:%s: %v", userID, tenantID, err)
	}
}

// invalidateTenantUserRolesCache invalidates tenant user roles cache
func (r *userTenantRoleRepository) invalidateTenantUserRolesCache(ctx context.Context, tenantID string) {
	cacheKey := fmt.Sprintf("tenant_user_roles:%s", tenantID)
	if err := r.tenantUserRolesCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate tenant user roles cache %s: %v", tenantID, err)
	}
}

// invalidateRoleUserTenantsCache invalidates role user tenants cache
func (r *userTenantRoleRepository) invalidateRoleUserTenantsCache(ctx context.Context, roleID string) {
	cacheKey := fmt.Sprintf("role_user_tenants:%s", roleID)
	if err := r.roleUserTenantsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate role user tenants cache %s: %v", roleID, err)
	}
}
