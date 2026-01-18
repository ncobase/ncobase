package repository

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"ncobase/core/access/data"
	"ncobase/core/access/data/ent"
	roleEnt "ncobase/core/access/data/ent/role"
	userRoleEnt "ncobase/core/access/data/ent/userrole"
	"ncobase/core/access/structs"
	"time"

	"github.com/ncobase/ncore/data/cache"
	"github.com/ncobase/ncore/logging/logger"
)

// UserRoleRepositoryInterface represents the user role repository interface.
type UserRoleRepositoryInterface interface {
	Create(ctx context.Context, body *structs.UserRole) (*ent.UserRole, error)
	GetByIDAndRoleID(ctx context.Context, uid, rid string) (*ent.UserRole, error)
	GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserRole, error)
	GetByRoleIDs(ctx context.Context, ids []string) ([]*ent.UserRole, error)
	Delete(ctx context.Context, uid, rid string) error
	DeleteAllByUserID(ctx context.Context, id string) error
	DeleteAllByRoleID(ctx context.Context, id string) error
	VerifyUserRole(ctx context.Context, userID, roleID string) (bool, error)
	GetRolesByUserID(ctx context.Context, userID string) ([]*ent.Role, error)
	GetUsersByRoleID(ctx context.Context, roleID string) ([]string, error)
	IsUserInRole(ctx context.Context, userID string, roleID string) (bool, error)
}

// userRoleRepository implements the UserRoleRepositoryInterface.
type userRoleRepository struct {
	data            *data.Data
	userRoleCache   cache.ICache[ent.UserRole]
	userRolesCache  cache.ICache[[]string] // Maps user ID to role IDs
	roleUsersCache  cache.ICache[[]string] // Maps role ID to user IDs
	relationshipTTL time.Duration
}

// NewUserRoleRepository creates a new user role repository.
func NewUserRoleRepository(d *data.Data) UserRoleRepositoryInterface {
	redisClient := d.GetRedis().(*redis.Client)

	return &userRoleRepository{
		data:            d,
		userRoleCache:   cache.NewCache[ent.UserRole](redisClient, "ncse_access:user_roles"),
		userRolesCache:  cache.NewCache[[]string](redisClient, "ncse_access:user_role_mappings"),
		roleUsersCache:  cache.NewCache[[]string](redisClient, "ncse_access:role_user_mappings"),
		relationshipTTL: time.Hour * 2,
	}
}

// VerifyUserRole verifies if a user has a specific role.
func (r *userRoleRepository) VerifyUserRole(ctx context.Context, userID, roleID string) (bool, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("relationship:%s:%s", userID, roleID)
	if cached, err := r.userRoleCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return true, nil
	}

	// Use slave for reads
	count, err := r.data.GetSlaveEntClient().UserRole.Query().
		Where(
			userRoleEnt.UserIDEQ(userID),
			userRoleEnt.RoleIDEQ(roleID),
		).
		Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "userRoleRepo.VerifyUserRole error: %v", err)
		return false, err
	}

	exists := count > 0

	// Cache the result if it exists
	if exists {
		go func() {
			// Create a dummy relationship for caching
			relationship := &ent.UserRole{
				UserID: userID,
				RoleID: roleID,
			}
			r.cacheUserRole(context.Background(), relationship)
		}()
	}

	return exists, nil
}

// Create create user role
func (r *userRoleRepository) Create(ctx context.Context, body *structs.UserRole) (*ent.UserRole, error) {
	// Verify if the user already has the role
	exists, err := r.VerifyUserRole(ctx, body.UserID, body.RoleID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("user %s already has role %s", body.UserID, body.RoleID)
	}

	// Use master for writes
	builder := r.data.GetMasterEntClient().UserRole.Create()

	// Set values
	builder.SetNillableUserID(&body.UserID)
	builder.SetNillableRoleID(&body.RoleID)

	// Execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "userRoleRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the relationship and invalidate related caches
	go func() {
		r.cacheUserRole(context.Background(), row)
		r.invalidateUserRolesCache(context.Background(), body.UserID)
		r.invalidateRoleUsersCache(context.Background(), body.RoleID)
	}()

	return row, nil
}

// GetByIDAndRoleID find role by user id and role id
func (r *userRoleRepository) GetByIDAndRoleID(ctx context.Context, uid, rid string) (*ent.UserRole, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("relationship:%s:%s", uid, rid)
	if cached, err := r.userRoleCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserRole.Query()

	// Set conditions
	builder.Where(userRoleEnt.UserIDEQ(uid), userRoleEnt.RoleIDEQ(rid))

	// Execute the builder
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "userRoleRepo.GetByIDAndRoleID error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheUserRole(context.Background(), row)

	return row, nil
}

// GetByUserIDs find roles by user ids
func (r *userRoleRepository) GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserRole, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserRole.Query()

	// Set conditions
	builder.Where(userRoleEnt.UserIDIn(ids...))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userRoleRepo.GetByUserIDs error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, ur := range rows {
			r.cacheUserRole(context.Background(), ur)
		}
	}()

	return rows, nil
}

// GetByRoleID find role by role id (kept for compatibility, returns first match)
func (r *userRoleRepository) GetByRoleID(ctx context.Context, id string) (*ent.UserRole, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserRole.Query()

	// Set condition
	builder.Where(userRoleEnt.RoleIDEQ(id))

	// Execute the builder
	row, err := builder.First(ctx)
	if err != nil {
		logger.Errorf(ctx, "userRoleRepo.GetByRoleID error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheUserRole(context.Background(), row)

	return row, nil
}

// GetByRoleIDs find roles by role ids
func (r *userRoleRepository) GetByRoleIDs(ctx context.Context, ids []string) ([]*ent.UserRole, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserRole.Query()

	// Set conditions
	builder.Where(userRoleEnt.RoleIDIn(ids...))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userRoleRepo.GetByRoleIDs error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, ur := range rows {
			r.cacheUserRole(context.Background(), ur)
		}
	}()

	return rows, nil
}

// Delete delete user role
func (r *userRoleRepository) Delete(ctx context.Context, uid, rid string) error {
	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserRole.Delete().Where(userRoleEnt.UserIDEQ(uid), userRoleEnt.RoleIDEQ(rid)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userRoleRepo.Delete error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateUserRoleCache(context.Background(), uid, rid)
		r.invalidateUserRolesCache(context.Background(), uid)
		r.invalidateRoleUsersCache(context.Background(), rid)
	}()

	return nil
}

// DeleteAllByUserID delete all user roles by user ID
func (r *userRoleRepository) DeleteAllByUserID(ctx context.Context, id string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.GetByUserIDs(ctx, []string{id})
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserRole.Delete().Where(userRoleEnt.UserIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userRoleRepo.DeleteAllByUserID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateUserRolesCache(context.Background(), id)
		for _, ur := range relationships {
			r.invalidateUserRoleCache(context.Background(), ur.UserID, ur.RoleID)
			r.invalidateRoleUsersCache(context.Background(), ur.RoleID)
		}
	}()

	return nil
}

// DeleteAllByRoleID delete all user roles by role ID
func (r *userRoleRepository) DeleteAllByRoleID(ctx context.Context, id string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.GetByRoleIDs(ctx, []string{id})
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserRole.Delete().Where(userRoleEnt.RoleIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userRoleRepo.DeleteAllByRoleID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateRoleUsersCache(context.Background(), id)
		for _, ur := range relationships {
			r.invalidateUserRoleCache(context.Background(), ur.UserID, ur.RoleID)
			r.invalidateUserRolesCache(context.Background(), ur.UserID)
		}
	}()

	return nil
}

// GetRolesByUserID retrieves all roles assigned to a user.
func (r *userRoleRepository) GetRolesByUserID(ctx context.Context, userID string) ([]*ent.Role, error) {
	// Try to get role IDs from cache
	cacheKey := fmt.Sprintf("user_roles:%s", userID)
	var roleIDs []string
	if err := r.userRolesCache.GetArray(ctx, cacheKey, &roleIDs); err == nil && len(roleIDs) > 0 {
		// Get roles by IDs from role repository
		return r.data.GetSlaveEntClient().Role.Query().Where(roleEnt.IDIn(roleIDs...)).All(ctx)
	}

	// Fallback to database
	userRoles, err := r.data.GetSlaveEntClient().UserRole.Query().Where(userRoleEnt.UserIDEQ(userID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userRoleRepo.GetRolesByUserID error: %v", err)
		return nil, err
	}

	// Extract role IDs from userRoles
	roleIDs = make([]string, len(userRoles))
	for i, userRole := range userRoles {
		roleIDs[i] = userRole.RoleID
	}

	// Query roles based on extracted role IDs
	roles, err := r.data.GetSlaveEntClient().Role.Query().Where(roleEnt.IDIn(roleIDs...)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userRoleRepo.GetRolesByUserID error: %v", err)
		return nil, err
	}

	// Cache role IDs for future use
	go func() {
		if err := r.userRolesCache.SetArray(context.Background(), cacheKey, roleIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache user roles %s: %v", userID, err)
		}
	}()

	return roles, nil
}

// GetUsersByRoleID retrieves all users assigned to a role.
func (r *userRoleRepository) GetUsersByRoleID(ctx context.Context, roleID string) ([]string, error) {
	// Try to get user IDs from cache
	cacheKey := fmt.Sprintf("role_users:%s", roleID)
	var userIDs []string
	if err := r.roleUsersCache.GetArray(ctx, cacheKey, &userIDs); err == nil && len(userIDs) > 0 {
		return userIDs, nil
	}

	// Fallback to database
	userRoles, err := r.data.GetSlaveEntClient().UserRole.Query().Where(userRoleEnt.RoleIDEQ(roleID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userRoleRepo.GetUsersByRoleID error: %v", err)
		return nil, err
	}

	// Extract user IDs from userRoles
	userIDs = make([]string, len(userRoles))
	for i, userRole := range userRoles {
		userIDs[i] = userRole.UserID
	}

	// Cache user IDs for future use
	go func() {
		if err := r.roleUsersCache.SetArray(context.Background(), cacheKey, userIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache role users %s: %v", roleID, err)
		}
	}()

	return userIDs, nil
}

// IsUserInRole verifies if a user has a specific role.
func (r *userRoleRepository) IsUserInRole(ctx context.Context, userID string, roleID string) (bool, error) {
	return r.VerifyUserRole(ctx, userID, roleID)
}

// cacheUserRole caches a user role relationship
func (r *userRoleRepository) cacheUserRole(ctx context.Context, ur *ent.UserRole) {
	// Cache by relationship key
	relationshipKey := fmt.Sprintf("relationship:%s:%s", ur.UserID, ur.RoleID)
	if err := r.userRoleCache.Set(ctx, relationshipKey, ur, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache user role relationship %s:%s: %v", ur.UserID, ur.RoleID, err)
	}
}

// invalidateUserRoleCache invalidates user role cache
func (r *userRoleRepository) invalidateUserRoleCache(ctx context.Context, userID, roleID string) {
	// Invalidate relationship cache
	relationshipKey := fmt.Sprintf("relationship:%s:%s", userID, roleID)
	if err := r.userRoleCache.Delete(ctx, relationshipKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user role relationship cache %s:%s: %v", userID, roleID, err)
	}
}

// invalidateUserRolesCache invalidates user roles cache
func (r *userRoleRepository) invalidateUserRolesCache(ctx context.Context, userID string) {
	cacheKey := fmt.Sprintf("user_roles:%s", userID)
	if err := r.userRolesCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user roles cache %s: %v", userID, err)
	}
}

// invalidateRoleUsersCache invalidates role users cache
func (r *userRoleRepository) invalidateRoleUsersCache(ctx context.Context, roleID string) {
	cacheKey := fmt.Sprintf("role_users:%s", roleID)
	if err := r.roleUsersCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate role users cache %s: %v", roleID, err)
	}
}
