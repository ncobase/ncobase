package repository

import (
	"context"
	"fmt"
	"ncobase/space/data"
	"ncobase/space/data/ent"
	userGroupEnt "ncobase/space/data/ent/usergroup"
	"ncobase/space/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/logging/logger"
)

// UserGroupRepositoryInterface represents the user group repository interface.
type UserGroupRepositoryInterface interface {
	Create(ctx context.Context, body *structs.UserGroup) (*ent.UserGroup, error)
	GetByUserID(ctx context.Context, id string) ([]*ent.UserGroup, error)
	GetByGroupID(ctx context.Context, id string) ([]*ent.UserGroup, error)
	GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserGroup, error)
	GetByGroupIDs(ctx context.Context, ids []string) ([]*ent.UserGroup, error)
	GetByGroupIDAndRole(ctx context.Context, id string, role structs.UserRole) ([]*ent.UserGroup, error)
	GetUserGroup(ctx context.Context, uid, gid string) (*ent.UserGroup, error)
	Delete(ctx context.Context, uid, gid string) error
	DeleteAllByUserID(ctx context.Context, id string) error
	DeleteAllByGroupID(ctx context.Context, id string) error
	GetGroupsByUserID(ctx context.Context, userID string) ([]string, error)
	GetUsersByGroupID(ctx context.Context, groupID string) ([]string, error)
	IsUserInGroup(ctx context.Context, userID string, groupID string) (bool, error)
	UserHasRole(ctx context.Context, userID string, groupID string, role structs.UserRole) (bool, error)
}

// userGroupRepository implements the UserGroupRepositoryInterface.
type userGroupRepository struct {
	data                *data.Data
	userGroupCache      cache.ICache[ent.UserGroup]
	userGroupsCache     cache.ICache[[]string] // Maps user ID to group IDs
	groupUsersCache     cache.ICache[[]string] // Maps group ID to user IDs
	groupRoleUsersCache cache.ICache[[]string] // Maps group:role to user IDs
	relationshipTTL     time.Duration
}

// NewUserGroupRepository creates a new user group repository.
func NewUserGroupRepository(d *data.Data) UserGroupRepositoryInterface {
	redisClient := d.GetRedis()

	return &userGroupRepository{
		data:                d,
		userGroupCache:      cache.NewCache[ent.UserGroup](redisClient, "ncse_space:user_groups"),
		userGroupsCache:     cache.NewCache[[]string](redisClient, "ncse_space:user_group_mappings"),
		groupUsersCache:     cache.NewCache[[]string](redisClient, "ncse_space:group_user_mappings"),
		groupRoleUsersCache: cache.NewCache[[]string](redisClient, "ncse_space:group_role_user_mappings"),
		relationshipTTL:     time.Hour * 2, // 2 hours cache TTL
	}
}

// Create create user group
func (r *userGroupRepository) Create(ctx context.Context, body *structs.UserGroup) (*ent.UserGroup, error) {
	// Use master for writes
	builder := r.data.GetMasterEntClient().UserGroup.Create()

	// Set values
	builder.SetUserID(body.UserID)
	builder.SetGroupID(body.GroupID)

	// Set role if provided
	if body.Role != "" {
		builder.SetRole(string(body.Role))
	} else {
		builder.SetRole(string(structs.RoleMember)) // Default role
	}

	// Execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the relationship and invalidate related caches
	go func() {
		r.cacheUserGroup(context.Background(), row)
		r.invalidateUserGroupsCache(context.Background(), body.UserID)
		r.invalidateGroupUsersCache(context.Background(), body.GroupID)
		if body.Role != "" {
			r.invalidateGroupRoleUsersCache(context.Background(), body.GroupID, string(body.Role))
		}
	}()

	return row, nil
}

// GetByUserID find groups by user id
func (r *userGroupRepository) GetByUserID(ctx context.Context, id string) ([]*ent.UserGroup, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserGroup.Query()

	// Set conditions
	builder.Where(userGroupEnt.UserIDEQ(id))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.GetByUserID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, ug := range rows {
			r.cacheUserGroup(context.Background(), ug)
		}
	}()

	return rows, nil
}

// GetByUserIDs find groups by user ids
func (r *userGroupRepository) GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserGroup, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserGroup.Query()

	// Set conditions
	builder.Where(userGroupEnt.UserIDIn(ids...))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.GetByUserIDs error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, ug := range rows {
			r.cacheUserGroup(context.Background(), ug)
		}
	}()

	return rows, nil
}

// GetByGroupID find users by group id
func (r *userGroupRepository) GetByGroupID(ctx context.Context, id string) ([]*ent.UserGroup, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserGroup.Query()

	// Set conditions
	builder.Where(userGroupEnt.GroupIDEQ(id))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.GetByGroupID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, ug := range rows {
			r.cacheUserGroup(context.Background(), ug)
		}
	}()

	return rows, nil
}

// GetByGroupIDs find users by group ids
func (r *userGroupRepository) GetByGroupIDs(ctx context.Context, ids []string) ([]*ent.UserGroup, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserGroup.Query()

	// Set conditions
	builder.Where(userGroupEnt.GroupIDIn(ids...))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.GetByGroupIDs error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, ug := range rows {
			r.cacheUserGroup(context.Background(), ug)
		}
	}()

	return rows, nil
}

// GetByGroupIDAndRole find users by group id and role
func (r *userGroupRepository) GetByGroupIDAndRole(ctx context.Context, id string, role structs.UserRole) ([]*ent.UserGroup, error) {
	// Try to get user IDs from cache
	cacheKey := fmt.Sprintf("group_role_users:%s:%s", id, string(role))
	var userIDs []string
	if err := r.groupRoleUsersCache.GetArray(ctx, cacheKey, &userIDs); err == nil && len(userIDs) > 0 {
		// Get user groups by user IDs and group ID
		return r.data.GetSlaveEntClient().UserGroup.Query().
			Where(userGroupEnt.GroupIDEQ(id), userGroupEnt.UserIDIn(userIDs...)).All(ctx)
	}

	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserGroup.Query()

	// Set conditions
	builder.Where(
		userGroupEnt.GroupIDEQ(id),
		userGroupEnt.RoleEQ(string(role)),
	)

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.GetByGroupIDAndRole error: %v", err)
		return nil, err
	}

	// Cache user IDs for future use
	go func() {
		userIDs := make([]string, 0, len(rows))
		for _, ug := range rows {
			userIDs = append(userIDs, ug.UserID)
			r.cacheUserGroup(context.Background(), ug)
		}
		if err := r.groupRoleUsersCache.SetArray(context.Background(), cacheKey, userIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache group role users %s:%s: %v", id, string(role), err)
		}
	}()

	return rows, nil
}

// GetUserGroup gets a specific user-group relation
func (r *userGroupRepository) GetUserGroup(ctx context.Context, uid, gid string) (*ent.UserGroup, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("relationship:%s:%s", uid, gid)
	if cached, err := r.userGroupCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserGroup.Query()

	// Set conditions
	builder.Where(
		userGroupEnt.UserIDEQ(uid),
		userGroupEnt.GroupIDEQ(gid),
	)

	// Execute the builder
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.GetUserGroup error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheUserGroup(context.Background(), row)

	return row, nil
}

// Delete delete user group
func (r *userGroupRepository) Delete(ctx context.Context, uid, gid string) error {
	// Get existing relationship for cache invalidation
	userGroup, err := r.GetUserGroup(ctx, uid, gid)
	if err != nil {
		logger.Debugf(ctx, "Failed to get user group for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserGroup.Delete().
		Where(userGroupEnt.UserIDEQ(uid), userGroupEnt.GroupIDEQ(gid)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userGroupRepo.Delete error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateUserGroupCache(context.Background(), uid, gid)
		r.invalidateUserGroupsCache(context.Background(), uid)
		r.invalidateGroupUsersCache(context.Background(), gid)
		if userGroup != nil {
			r.invalidateGroupRoleUsersCache(context.Background(), gid, userGroup.Role)
		}
	}()

	return nil
}

// DeleteAllByUserID delete all user group by user id
func (r *userGroupRepository) DeleteAllByUserID(ctx context.Context, id string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.GetByUserID(ctx, id)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserGroup.Delete().
		Where(userGroupEnt.UserIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userGroupRepo.DeleteAllByUserID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateUserGroupsCache(context.Background(), id)
		for _, ug := range relationships {
			r.invalidateUserGroupCache(context.Background(), ug.UserID, ug.GroupID)
			r.invalidateGroupUsersCache(context.Background(), ug.GroupID)
			r.invalidateGroupRoleUsersCache(context.Background(), ug.GroupID, ug.Role)
		}
	}()

	return nil
}

// DeleteAllByGroupID delete all user group by group id
func (r *userGroupRepository) DeleteAllByGroupID(ctx context.Context, id string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.GetByGroupID(ctx, id)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserGroup.Delete().
		Where(userGroupEnt.GroupIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userGroupRepo.DeleteAllByGroupID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateGroupUsersCache(context.Background(), id)
		for _, ug := range relationships {
			r.invalidateUserGroupCache(context.Background(), ug.UserID, ug.GroupID)
			r.invalidateUserGroupsCache(context.Background(), ug.UserID)
			r.invalidateGroupRoleUsersCache(context.Background(), id, ug.Role)
		}
	}()

	return nil
}

// GetGroupsByUserID retrieves all groups a user belongs to.
func (r *userGroupRepository) GetGroupsByUserID(ctx context.Context, userID string) ([]string, error) {
	// Try to get group IDs from cache
	cacheKey := fmt.Sprintf("user_groups:%s", userID)
	var groupIDs []string
	if err := r.userGroupsCache.GetArray(ctx, cacheKey, &groupIDs); err == nil && len(groupIDs) > 0 {
		return groupIDs, nil
	}

	// Fallback to database
	userGroups, err := r.data.GetSlaveEntClient().UserGroup.Query().
		Where(userGroupEnt.UserIDEQ(userID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.GetGroupsByUserID error: %v", err)
		return nil, err
	}

	// Extract group IDs from userGroups
	groupIDs = make([]string, len(userGroups))
	for i, group := range userGroups {
		groupIDs[i] = group.GroupID
	}

	// Cache group IDs for future use
	go func() {
		if err := r.userGroupsCache.SetArray(context.Background(), cacheKey, groupIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache user groups %s: %v", userID, err)
		}
	}()

	return groupIDs, nil
}

// GetUsersByGroupID retrieves all users in a group.
func (r *userGroupRepository) GetUsersByGroupID(ctx context.Context, groupID string) ([]string, error) {
	// Try to get user IDs from cache
	cacheKey := fmt.Sprintf("group_users:%s", groupID)
	var userIDs []string
	if err := r.groupUsersCache.GetArray(ctx, cacheKey, &userIDs); err == nil && len(userIDs) > 0 {
		return userIDs, nil
	}

	// Fallback to database
	userGroups, err := r.data.GetSlaveEntClient().UserGroup.Query().
		Where(userGroupEnt.GroupIDEQ(groupID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.GetUsersByGroupID error: %v", err)
		return nil, err
	}

	// Extract user IDs from userGroups
	userIDs = make([]string, len(userGroups))
	for i, userGroup := range userGroups {
		userIDs[i] = userGroup.UserID
	}

	// Cache user IDs for future use
	go func() {
		if err := r.groupUsersCache.SetArray(context.Background(), cacheKey, userIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache group users %s: %v", groupID, err)
		}
	}()

	return userIDs, nil
}

// IsUserInGroup verifies if a user belongs to a specific group.
func (r *userGroupRepository) IsUserInGroup(ctx context.Context, userID string, groupID string) (bool, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("relationship:%s:%s", userID, groupID)
	if cached, err := r.userGroupCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return true, nil
	}

	// Use slave for reads
	count, err := r.data.GetSlaveEntClient().UserGroup.Query().
		Where(userGroupEnt.UserIDEQ(userID), userGroupEnt.GroupIDEQ(groupID)).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.IsUserInGroup error: %v", err)
		return false, err
	}

	exists := count > 0

	// Cache the result if it exists
	if exists {
		go func() {
			// Get the actual relationship for caching
			if userGroup, err := r.GetUserGroup(context.Background(), userID, groupID); err == nil {
				r.cacheUserGroup(context.Background(), userGroup)
			}
		}()
	}

	return exists, nil
}

// UserHasRole verifies if a user has a specific role in a group.
func (r *userGroupRepository) UserHasRole(ctx context.Context, userID string, groupID string, role structs.UserRole) (bool, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("relationship:%s:%s", userID, groupID)
	if cached, err := r.userGroupCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached.Role == string(role), nil
	}

	// Use slave for reads
	count, err := r.data.GetSlaveEntClient().UserGroup.Query().Where(
		userGroupEnt.UserIDEQ(userID),
		userGroupEnt.GroupIDEQ(groupID),
		userGroupEnt.RoleEQ(string(role)),
	).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.UserHasRole error: %v", err)
		return false, err
	}

	exists := count > 0

	// Cache the result if it exists
	if exists {
		go func() {
			// Get the actual relationship for caching
			if userGroup, err := r.GetUserGroup(context.Background(), userID, groupID); err == nil {
				r.cacheUserGroup(context.Background(), userGroup)
			}
		}()
	}

	return exists, nil
}

// cacheUserGroup caches a user group relationship.
func (r *userGroupRepository) cacheUserGroup(ctx context.Context, ug *ent.UserGroup) {
	// Cache by relationship key
	relationshipKey := fmt.Sprintf("relationship:%s:%s", ug.UserID, ug.GroupID)
	if err := r.userGroupCache.Set(ctx, relationshipKey, ug, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache user group relationship %s:%s: %v", ug.UserID, ug.GroupID, err)
	}
}

// invalidateUserGroupCache invalidates a user group relationship cache.
func (r *userGroupRepository) invalidateUserGroupCache(ctx context.Context, userID, groupID string) {
	// Invalidate relationship cache
	relationshipKey := fmt.Sprintf("relationship:%s:%s", userID, groupID)
	if err := r.userGroupCache.Delete(ctx, relationshipKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user group relationship cache %s:%s: %v", userID, groupID, err)
	}
}

// invalidateUserGroupsCache invalidates the cache for a user's groups.
func (r *userGroupRepository) invalidateUserGroupsCache(ctx context.Context, userID string) {
	cacheKey := fmt.Sprintf("user_groups:%s", userID)
	if err := r.userGroupsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user groups cache %s: %v", userID, err)
	}
}

// invalidateGroupUsersCache invalidates the cache for a group's users.
func (r *userGroupRepository) invalidateGroupUsersCache(ctx context.Context, groupID string) {
	cacheKey := fmt.Sprintf("group_users:%s", groupID)
	if err := r.groupUsersCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate group users cache %s: %v", groupID, err)
	}
}

// invalidateGroupRoleUsersCache invalidates the cache for a group-role relationship.
func (r *userGroupRepository) invalidateGroupRoleUsersCache(ctx context.Context, groupID, role string) {
	cacheKey := fmt.Sprintf("group_role_users:%s:%s", groupID, role)
	if err := r.groupRoleUsersCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate group role users cache %s:%s: %v", groupID, role, err)
	}
}
