package repository

import (
	"context"
	"fmt"
	"ncobase/space/data"
	"ncobase/space/data/ent"
	userSpaceRoleEnt "ncobase/space/data/ent/userspacerole"
	"ncobase/space/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/logging/logger"
)

// UserSpaceRoleRepositoryInterface represents the user space role repository interface.
type UserSpaceRoleRepositoryInterface interface {
	Create(ctx context.Context, body *structs.UserSpaceRole) (*ent.UserSpaceRole, error)
	GetByUserID(ctx context.Context, u string) (*ent.UserSpaceRole, error)
	GetBySpaceID(ctx context.Context, t string) ([]*ent.UserSpaceRole, error)
	GetByRoleID(ctx context.Context, r string) ([]*ent.UserSpaceRole, error)
	DeleteByUserIDAndSpaceID(ctx context.Context, u, t string) error
	DeleteByUserIDAndRoleID(ctx context.Context, u, r string) error
	DeleteBySpaceIDAndRoleID(ctx context.Context, t, r string) error
	DeleteByUserIDAndSpaceIDAndRoleID(ctx context.Context, u, t, r string) error
	DeleteAllByUserID(ctx context.Context, u string) error
	DeleteAllBySpaceID(ctx context.Context, t string) error
	DeleteAllByRoleID(ctx context.Context, r string) error
	GetRolesByUserAndSpace(ctx context.Context, u, t string) ([]string, error)
	IsUserInRoleInSpace(ctx context.Context, u, t, r string) (bool, error)
}

// userSpaceRoleRepository implements the UserSpaceRoleRepositoryInterface.
type userSpaceRoleRepository struct {
	data                *data.Data
	userSpaceRoleCache  cache.ICache[ent.UserSpaceRole]
	userSpaceRolesCache cache.ICache[[]string] // Maps user:space to role IDs
	spaceUserRolesCache cache.ICache[[]string] // Maps space to user:role pairs
	roleUserSpacesCache cache.ICache[[]string] // Maps role to user:space pairs
	relationshipTTL     time.Duration
}

// NewUserSpaceRoleRepository creates a new user space role repository.
func NewUserSpaceRoleRepository(d *data.Data) UserSpaceRoleRepositoryInterface {
	redisClient := d.GetRedis()

	return &userSpaceRoleRepository{
		data:                d,
		userSpaceRoleCache:  cache.NewCache[ent.UserSpaceRole](redisClient, "ncse_access:user_space_roles"),
		userSpaceRolesCache: cache.NewCache[[]string](redisClient, "ncse_access:user_space_role_mappings"),
		spaceUserRolesCache: cache.NewCache[[]string](redisClient, "ncse_access:space_user_role_mappings"),
		roleUserSpacesCache: cache.NewCache[[]string](redisClient, "ncse_access:role_user_space_mappings"),
		relationshipTTL:     time.Hour * 2, // 2 hours cache TTL
	}
}

// Create creates a new user space role.
func (r *userSpaceRoleRepository) Create(ctx context.Context, body *structs.UserSpaceRole) (*ent.UserSpaceRole, error) {
	// Use master for writes
	builder := r.data.GetMasterEntClient().UserSpaceRole.Create()

	// Set values
	builder.SetNillableUserID(&body.UserID)
	builder.SetNillableSpaceID(&body.SpaceID)
	builder.SetNillableRoleID(&body.RoleID)

	// Execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "userSpaceRoleRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the relationship and invalidate related caches
	go func() {
		r.cacheUserSpaceRole(context.Background(), row)
		r.invalidateUserSpaceRolesCache(context.Background(), body.UserID, body.SpaceID)
		r.invalidateSpaceUserRolesCache(context.Background(), body.SpaceID)
		r.invalidateRoleUserSpacesCache(context.Background(), body.RoleID)
	}()

	return row, nil
}

// GetByUserID retrieves user space role by user ID.
func (r *userSpaceRoleRepository) GetByUserID(ctx context.Context, u string) (*ent.UserSpaceRole, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("user:%s", u)
	if cached, err := r.userSpaceRoleCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserSpaceRole.Query()

	// Set conditions
	builder.Where(userSpaceRoleEnt.UserIDEQ(u))

	// Execute the builder
	row, err := builder.First(ctx)
	if err != nil {
		logger.Errorf(ctx, "userSpaceRoleRepo.GetByUserID error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheUserSpaceRole(context.Background(), row)

	return row, nil
}

// GetBySpaceID retrieves user space roles by space ID.
func (r *userSpaceRoleRepository) GetBySpaceID(ctx context.Context, t string) ([]*ent.UserSpaceRole, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserSpaceRole.Query()

	// Set conditions
	builder.Where(userSpaceRoleEnt.SpaceID(t))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userSpaceRoleRepo.GetBySpaceID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, utr := range rows {
			r.cacheUserSpaceRole(context.Background(), utr)
		}
	}()

	return rows, nil
}

// GetByRoleID retrieves user space roles by role ID.
func (r *userSpaceRoleRepository) GetByRoleID(ctx context.Context, rid string) ([]*ent.UserSpaceRole, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserSpaceRole.Query()

	// Set conditions
	builder.Where(userSpaceRoleEnt.RoleID(rid))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userSpaceRoleRepo.GetByRoleID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, utr := range rows {
			r.cacheUserSpaceRole(context.Background(), utr)
		}
	}()

	return rows, nil
}

// DeleteByUserIDAndSpaceID deletes user space role by user ID and space ID.
func (r *userSpaceRoleRepository) DeleteByUserIDAndSpaceID(ctx context.Context, u, t string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().UserSpaceRole.Query().
		Where(userSpaceRoleEnt.UserIDEQ(u), userSpaceRoleEnt.SpaceID(t)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserSpaceRole.Delete().
		Where(userSpaceRoleEnt.UserIDEQ(u), userSpaceRoleEnt.SpaceID(t)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userSpaceRoleRepo.DeleteByUserIDAndSpaceID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateUserSpaceRolesCache(context.Background(), u, t)
		r.invalidateSpaceUserRolesCache(context.Background(), t)
		for _, utr := range relationships {
			r.invalidateUserSpaceRoleCache(context.Background(), utr)
			r.invalidateRoleUserSpacesCache(context.Background(), utr.RoleID)
		}
	}()

	return nil
}

// DeleteByUserIDAndRoleID deletes user space role by user ID and role ID.
func (r *userSpaceRoleRepository) DeleteByUserIDAndRoleID(ctx context.Context, u, rid string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().UserSpaceRole.Query().
		Where(userSpaceRoleEnt.UserIDEQ(u), userSpaceRoleEnt.RoleID(rid)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserSpaceRole.Delete().
		Where(userSpaceRoleEnt.UserIDEQ(u), userSpaceRoleEnt.RoleID(rid)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userSpaceRoleRepo.DeleteByUserIDAndRoleID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateRoleUserSpacesCache(context.Background(), rid)
		for _, utr := range relationships {
			r.invalidateUserSpaceRoleCache(context.Background(), utr)
			r.invalidateUserSpaceRolesCache(context.Background(), u, utr.SpaceID)
			r.invalidateSpaceUserRolesCache(context.Background(), utr.SpaceID)
		}
	}()

	return nil
}

// DeleteBySpaceIDAndRoleID deletes user space role by space ID and role ID.
func (r *userSpaceRoleRepository) DeleteBySpaceIDAndRoleID(ctx context.Context, t, rid string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().UserSpaceRole.Query().
		Where(userSpaceRoleEnt.SpaceID(t), userSpaceRoleEnt.RoleID(rid)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserSpaceRole.Delete().
		Where(userSpaceRoleEnt.SpaceID(t), userSpaceRoleEnt.RoleID(rid)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userSpaceRoleRepo.DeleteBySpaceIDAndRoleID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateSpaceUserRolesCache(context.Background(), t)
		r.invalidateRoleUserSpacesCache(context.Background(), rid)
		for _, utr := range relationships {
			r.invalidateUserSpaceRoleCache(context.Background(), utr)
			r.invalidateUserSpaceRolesCache(context.Background(), utr.UserID, t)
		}
	}()

	return nil
}

// DeleteByUserIDAndSpaceIDAndRoleID deletes user space role by user ID, space ID and role ID.
func (r *userSpaceRoleRepository) DeleteByUserIDAndSpaceIDAndRoleID(ctx context.Context, u, t, rid string) error {
	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserSpaceRole.Delete().
		Where(userSpaceRoleEnt.UserIDEQ(u), userSpaceRoleEnt.SpaceID(t), userSpaceRoleEnt.RoleID(rid)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userSpaceRoleRepo.DeleteByUserIDAndSpaceIDAndRoleID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		// Create a dummy relationship for cache invalidation
		utr := &ent.UserSpaceRole{
			UserID:  u,
			SpaceID: t,
			RoleID:  rid,
		}
		r.invalidateUserSpaceRoleCache(context.Background(), utr)
		r.invalidateUserSpaceRolesCache(context.Background(), u, t)
		r.invalidateSpaceUserRolesCache(context.Background(), t)
		r.invalidateRoleUserSpacesCache(context.Background(), rid)
	}()

	return nil
}

// DeleteAllByUserID deletes all user space roles by user ID.
func (r *userSpaceRoleRepository) DeleteAllByUserID(ctx context.Context, u string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().UserSpaceRole.Query().
		Where(userSpaceRoleEnt.UserIDEQ(u)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserSpaceRole.Delete().
		Where(userSpaceRoleEnt.UserIDEQ(u)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userSpaceRoleRepo.DeleteAllByUserID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		for _, utr := range relationships {
			r.invalidateUserSpaceRoleCache(context.Background(), utr)
			r.invalidateUserSpaceRolesCache(context.Background(), u, utr.SpaceID)
			r.invalidateSpaceUserRolesCache(context.Background(), utr.SpaceID)
			r.invalidateRoleUserSpacesCache(context.Background(), utr.RoleID)
		}
	}()

	return nil
}

// DeleteAllBySpaceID deletes all user space roles by space ID.
func (r *userSpaceRoleRepository) DeleteAllBySpaceID(ctx context.Context, t string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().UserSpaceRole.Query().
		Where(userSpaceRoleEnt.SpaceID(t)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserSpaceRole.Delete().
		Where(userSpaceRoleEnt.SpaceID(t)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userSpaceRoleRepo.DeleteAllBySpaceID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateSpaceUserRolesCache(context.Background(), t)
		for _, utr := range relationships {
			r.invalidateUserSpaceRoleCache(context.Background(), utr)
			r.invalidateUserSpaceRolesCache(context.Background(), utr.UserID, t)
			r.invalidateRoleUserSpacesCache(context.Background(), utr.RoleID)
		}
	}()

	return nil
}

// DeleteAllByRoleID deletes all user space roles by role ID.
func (r *userSpaceRoleRepository) DeleteAllByRoleID(ctx context.Context, rid string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().UserSpaceRole.Query().
		Where(userSpaceRoleEnt.RoleID(rid)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserSpaceRole.Delete().
		Where(userSpaceRoleEnt.RoleID(rid)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userSpaceRoleRepo.DeleteAllByRoleID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateRoleUserSpacesCache(context.Background(), rid)
		for _, utr := range relationships {
			r.invalidateUserSpaceRoleCache(context.Background(), utr)
			r.invalidateUserSpaceRolesCache(context.Background(), utr.UserID, utr.SpaceID)
			r.invalidateSpaceUserRolesCache(context.Background(), utr.SpaceID)
		}
	}()

	return nil
}

// GetRolesByUserAndSpace retrieves all roles a user has in a space.
func (r *userSpaceRoleRepository) GetRolesByUserAndSpace(ctx context.Context, u string, t string) ([]string, error) {
	// Try to get role IDs from cache
	cacheKey := fmt.Sprintf("user_space_roles:%s:%s", u, t)
	var roleIDs []string
	if err := r.userSpaceRolesCache.GetArray(ctx, cacheKey, &roleIDs); err == nil && len(roleIDs) > 0 {
		return roleIDs, nil
	}

	// Fallback to database
	userSpaceRoles, err := r.data.GetSlaveEntClient().UserSpaceRole.Query().
		Where(userSpaceRoleEnt.UserIDEQ(u), userSpaceRoleEnt.SpaceIDEQ(t)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userSpaceRoleRepo.GetRolesByUserAndSpace error: %v", err)
		return nil, err
	}

	// Extract role IDs from userSpaceRoles
	roleIDs = make([]string, len(userSpaceRoles))
	for i, userRole := range userSpaceRoles {
		roleIDs[i] = userRole.RoleID
	}

	// Cache role IDs for future use
	go func() {
		if err := r.userSpaceRolesCache.SetArray(context.Background(), cacheKey, roleIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache user space roles %s:%s: %v", u, t, err)
		}
	}()

	return roleIDs, nil
}

// IsUserInRoleInSpace verifies if a user has a specific role in a space.
func (r *userSpaceRoleRepository) IsUserInRoleInSpace(ctx context.Context, u, t, rid string) (bool, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("relationship:%s:%s:%s", u, t, rid)
	if cached, err := r.userSpaceRoleCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return true, nil
	}

	// Use slave for reads
	count, err := r.data.GetSlaveEntClient().UserSpaceRole.Query().
		Where(userSpaceRoleEnt.UserIDEQ(u), userSpaceRoleEnt.SpaceIDEQ(t), userSpaceRoleEnt.RoleIDEQ(rid)).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "userSpaceRoleRepo.IsUserInRoleInSpace error: %v", err)
		return false, err
	}

	exists := count > 0

	// Cache the result if it exists
	if exists {
		go func() {
			// Create a dummy relationship for caching
			relationship := &ent.UserSpaceRole{
				UserID:  u,
				SpaceID: t,
				RoleID:  rid,
			}
			r.cacheUserSpaceRole(context.Background(), relationship)
		}()
	}

	return exists, nil
}

// cacheUserSpaceRole caches a user space role relationship.
func (r *userSpaceRoleRepository) cacheUserSpaceRole(ctx context.Context, utr *ent.UserSpaceRole) {
	// Cache by relationship key
	relationshipKey := fmt.Sprintf("relationship:%s:%s:%s", utr.UserID, utr.SpaceID, utr.RoleID)
	if err := r.userSpaceRoleCache.Set(ctx, relationshipKey, utr, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache user space role relationship %s:%s:%s: %v", utr.UserID, utr.SpaceID, utr.RoleID, err)
	}

	// Cache by user ID
	userKey := fmt.Sprintf("user:%s", utr.UserID)
	if err := r.userSpaceRoleCache.Set(ctx, userKey, utr, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache user space role by user %s: %v", utr.UserID, err)
	}
}

// invalidateUserSpaceRoleCache invalidates user space role cache
func (r *userSpaceRoleRepository) invalidateUserSpaceRoleCache(ctx context.Context, utr *ent.UserSpaceRole) {
	// Invalidate relationship cache
	relationshipKey := fmt.Sprintf("relationship:%s:%s:%s", utr.UserID, utr.SpaceID, utr.RoleID)
	if err := r.userSpaceRoleCache.Delete(ctx, relationshipKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user space role relationship cache %s:%s:%s: %v", utr.UserID, utr.SpaceID, utr.RoleID, err)
	}

	// Invalidate user key
	userKey := fmt.Sprintf("user:%s", utr.UserID)
	if err := r.userSpaceRoleCache.Delete(ctx, userKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user space role cache by user %s: %v", utr.UserID, err)
	}
}

// invalidateUserSpaceRolesCache invalidates user space roles cache
func (r *userSpaceRoleRepository) invalidateUserSpaceRolesCache(ctx context.Context, userID, spaceID string) {
	cacheKey := fmt.Sprintf("user_space_roles:%s:%s", userID, spaceID)
	if err := r.userSpaceRolesCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user space roles cache %s:%s: %v", userID, spaceID, err)
	}
}

// invalidateSpaceUserRolesCache invalidates space user roles cache
func (r *userSpaceRoleRepository) invalidateSpaceUserRolesCache(ctx context.Context, spaceID string) {
	cacheKey := fmt.Sprintf("space_user_roles:%s", spaceID)
	if err := r.spaceUserRolesCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate space user roles cache %s: %v", spaceID, err)
	}
}

// invalidateRoleUserSpacesCache invalidates role user spaces cache
func (r *userSpaceRoleRepository) invalidateRoleUserSpacesCache(ctx context.Context, roleID string) {
	cacheKey := fmt.Sprintf("role_user_spaces:%s", roleID)
	if err := r.roleUserSpacesCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate role user spaces cache %s: %v", roleID, err)
	}
}
