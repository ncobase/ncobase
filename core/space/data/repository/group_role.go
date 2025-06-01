package repository

import (
	"context"
	"fmt"
	"ncobase/space/data"
	"ncobase/space/data/ent"
	groupEnt "ncobase/space/data/ent/group"
	groupRoleEnt "ncobase/space/data/ent/grouprole"
	"ncobase/space/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/logging/logger"
)

// GroupRoleRepositoryInterface represents the group role repository interface.
type GroupRoleRepositoryInterface interface {
	Create(ctx context.Context, body *structs.GroupRole) (*ent.GroupRole, error)
	GetByGroupID(ctx context.Context, id string) (*ent.GroupRole, error)
	GetByRoleID(ctx context.Context, id string) (*ent.GroupRole, error)
	GetByGroupIDs(ctx context.Context, ids []string) ([]*ent.GroupRole, error)
	GetByRoleIDs(ctx context.Context, ids []string) ([]*ent.GroupRole, error)
	Delete(ctx context.Context, gid, rid string) error
	DeleteAllByGroupID(ctx context.Context, id string) error
	DeleteAllByRoleID(ctx context.Context, id string) error
	GetRolesByGroupID(ctx context.Context, groupID string) ([]string, error)
	GetGroupsByRoleID(ctx context.Context, roleID string) ([]*ent.Group, error)
	IsRoleInGroup(ctx context.Context, groupID string, roleID string) (bool, error)
	IsGroupInRole(ctx context.Context, roleID string, groupID string) (bool, error)
}

// groupRoleRepository implements the GroupRoleRepositoryInterface.
type groupRoleRepository struct {
	data            *data.Data
	groupRoleCache  cache.ICache[ent.GroupRole]
	groupRolesCache cache.ICache[[]string] // Maps group ID to role IDs
	roleGroupsCache cache.ICache[[]string] // Maps role ID to group IDs
	relationshipTTL time.Duration
}

// NewGroupRoleRepository creates a new group role repository.
func NewGroupRoleRepository(d *data.Data) GroupRoleRepositoryInterface {
	redisClient := d.GetRedis()

	return &groupRoleRepository{
		data:            d,
		groupRoleCache:  cache.NewCache[ent.GroupRole](redisClient, "ncse_space:group_roles"),
		groupRolesCache: cache.NewCache[[]string](redisClient, "ncse_space:group_role_mappings"),
		roleGroupsCache: cache.NewCache[[]string](redisClient, "ncse_space:role_group_mappings"),
		relationshipTTL: time.Hour * 2, // 2 hours cache TTL
	}
}

// Create group role
func (r *groupRoleRepository) Create(ctx context.Context, body *structs.GroupRole) (*ent.GroupRole, error) {
	// Use master for writes
	builder := r.data.GetMasterEntClient().GroupRole.Create()

	// Set values
	builder.SetNillableGroupID(&body.GroupID)
	builder.SetNillableRoleID(&body.RoleID)

	// Execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "groupRoleRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the relationship and invalidate related caches
	go func() {
		r.cacheGroupRole(context.Background(), row)
		r.invalidateGroupRolesCache(context.Background(), body.GroupID)
		r.invalidateRoleGroupsCache(context.Background(), body.RoleID)
	}()

	return row, nil
}

// GetByGroupID Find role by group id
func (r *groupRoleRepository) GetByGroupID(ctx context.Context, id string) (*ent.GroupRole, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("group:%s", id)
	if cached, err := r.groupRoleCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Use slave for reads
	builder := r.data.GetSlaveEntClient().GroupRole.Query()

	// Set conditions
	builder.Where(groupRoleEnt.GroupIDEQ(id))

	// Execute the builder
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "groupRoleRepo.GetByGroupID error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheGroupRole(context.Background(), row)

	return row, nil
}

// GetByGroupIDs Find roles by group ids
func (r *groupRoleRepository) GetByGroupIDs(ctx context.Context, ids []string) ([]*ent.GroupRole, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().GroupRole.Query()

	// Set conditions
	builder.Where(groupRoleEnt.GroupIDIn(ids...))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "groupRoleRepo.GetByGroupIDs error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, gr := range rows {
			r.cacheGroupRole(context.Background(), gr)
		}
	}()

	return rows, nil
}

// GetByRoleID Find role by role id
func (r *groupRoleRepository) GetByRoleID(ctx context.Context, id string) (*ent.GroupRole, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("role:%s", id)
	if cached, err := r.groupRoleCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Use slave for reads
	builder := r.data.GetSlaveEntClient().GroupRole.Query()

	// Set conditions
	builder.Where(groupRoleEnt.RoleIDEQ(id))

	// Execute the builder
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "groupRoleRepo.GetByRoleID error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheGroupRole(context.Background(), row)

	return row, nil
}

// GetByRoleIDs Find roles by role ids
func (r *groupRoleRepository) GetByRoleIDs(ctx context.Context, ids []string) ([]*ent.GroupRole, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().GroupRole.Query()

	// Set conditions
	builder.Where(groupRoleEnt.RoleIDIn(ids...))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "groupRoleRepo.GetByRoleIDs error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, gr := range rows {
			r.cacheGroupRole(context.Background(), gr)
		}
	}()

	return rows, nil
}

// Delete group role
func (r *groupRoleRepository) Delete(ctx context.Context, gid, rid string) error {
	// Use master for writes
	if _, err := r.data.GetMasterEntClient().GroupRole.Delete().
		Where(groupRoleEnt.GroupIDEQ(gid), groupRoleEnt.RoleIDEQ(rid)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "groupRoleRepo.Delete error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateGroupRoleCache(context.Background(), gid, rid)
		r.invalidateGroupRolesCache(context.Background(), gid)
		r.invalidateRoleGroupsCache(context.Background(), rid)
	}()

	return nil
}

// DeleteAllByGroupID Delete all group role
func (r *groupRoleRepository) DeleteAllByGroupID(ctx context.Context, id string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.GetByGroupIDs(ctx, []string{id})
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().GroupRole.Delete().
		Where(groupRoleEnt.GroupIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "groupRoleRepo.DeleteAllByGroupID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateGroupRolesCache(context.Background(), id)
		for _, gr := range relationships {
			r.invalidateGroupRoleCache(context.Background(), gr.GroupID, gr.RoleID)
			r.invalidateRoleGroupsCache(context.Background(), gr.RoleID)
		}
	}()

	return nil
}

// DeleteAllByRoleID Delete all group role
func (r *groupRoleRepository) DeleteAllByRoleID(ctx context.Context, id string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.GetByRoleIDs(ctx, []string{id})
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().GroupRole.Delete().
		Where(groupRoleEnt.RoleIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "groupRoleRepo.DeleteAllByRoleID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateRoleGroupsCache(context.Background(), id)
		for _, gr := range relationships {
			r.invalidateGroupRoleCache(context.Background(), gr.GroupID, gr.RoleID)
			r.invalidateGroupRolesCache(context.Background(), gr.GroupID)
		}
	}()

	return nil
}

// GetRolesByGroupID retrieves all roles under a group.
func (r *groupRoleRepository) GetRolesByGroupID(ctx context.Context, groupID string) ([]string, error) {
	// Try to get role IDs from cache
	cacheKey := fmt.Sprintf("group_roles:%s", groupID)
	var roleIDs []string
	if err := r.groupRolesCache.GetArray(ctx, cacheKey, &roleIDs); err == nil && len(roleIDs) > 0 {
		return roleIDs, nil
	}

	// Fallback to database
	groupRoles, err := r.data.GetSlaveEntClient().GroupRole.Query().
		Where(groupRoleEnt.GroupIDEQ(groupID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "groupRoleRepo.GetRolesByGroupID error: %v", err)
		return nil, err
	}

	// Extract role IDs from groupRoles
	roleIDs = make([]string, len(groupRoles))
	for i, groupRole := range groupRoles {
		roleIDs[i] = groupRole.RoleID
	}

	// Cache role IDs for future use
	go func() {
		if err := r.groupRolesCache.SetArray(context.Background(), cacheKey, roleIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache group roles %s: %v", groupID, err)
		}
	}()

	return roleIDs, nil
}

// GetGroupsByRoleID retrieves all groups under a role.
func (r *groupRoleRepository) GetGroupsByRoleID(ctx context.Context, roleID string) ([]*ent.Group, error) {
	// Try to get group IDs from cache
	cacheKey := fmt.Sprintf("role_groups:%s", roleID)
	var groupIDs []string
	if err := r.roleGroupsCache.GetArray(ctx, cacheKey, &groupIDs); err == nil && len(groupIDs) > 0 {
		// Get groups by IDs from group repository
		return r.data.GetSlaveEntClient().Group.Query().Where(groupEnt.IDIn(groupIDs...)).All(ctx)
	}

	// Fallback to database
	groupRoles, err := r.data.GetSlaveEntClient().GroupRole.Query().
		Where(groupRoleEnt.RoleIDEQ(roleID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "groupRoleRepo.GetGroupsByRoleID error: %v", err)
		return nil, err
	}

	// Extract group IDs from groupRoles
	groupIDs = make([]string, len(groupRoles))
	for i, groupRole := range groupRoles {
		groupIDs[i] = groupRole.GroupID
	}

	// Query groups based on extracted group IDs
	groups, err := r.data.GetSlaveEntClient().Group.Query().Where(groupEnt.IDIn(groupIDs...)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "groupRoleRepo.GetGroupsByRoleID error: %v", err)
		return nil, err
	}

	// Cache group IDs for future use
	go func() {
		if err := r.roleGroupsCache.SetArray(context.Background(), cacheKey, groupIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache role groups %s: %v", roleID, err)
		}
	}()

	return groups, nil
}

// IsRoleInGroup verifies if a role belongs to a specific group.
func (r *groupRoleRepository) IsRoleInGroup(ctx context.Context, groupID string, roleID string) (bool, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("relationship:%s:%s", groupID, roleID)
	if cached, err := r.groupRoleCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return true, nil
	}

	// Use slave for reads
	count, err := r.data.GetSlaveEntClient().GroupRole.Query().
		Where(groupRoleEnt.GroupIDEQ(groupID), groupRoleEnt.RoleIDEQ(roleID)).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "groupRoleRepo.IsRoleInGroup error: %v", err)
		return false, err
	}

	exists := count > 0

	// Cache the result if it exists
	if exists {
		go func() {
			// Create a dummy relationship for caching
			relationship := &ent.GroupRole{
				GroupID: groupID,
				RoleID:  roleID,
			}
			r.cacheGroupRole(context.Background(), relationship)
		}()
	}

	return exists, nil
}

// IsGroupInRole verifies if a group belongs to a specific role.
func (r *groupRoleRepository) IsGroupInRole(ctx context.Context, groupID string, roleID string) (bool, error) {
	return r.IsRoleInGroup(ctx, groupID, roleID)
}

// cacheGroupRole caches a group-role relationship.
func (r *groupRoleRepository) cacheGroupRole(ctx context.Context, gr *ent.GroupRole) {
	// Cache by relationship key
	relationshipKey := fmt.Sprintf("relationship:%s:%s", gr.GroupID, gr.RoleID)
	if err := r.groupRoleCache.Set(ctx, relationshipKey, gr, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache group role relationship %s:%s: %v", gr.GroupID, gr.RoleID, err)
	}

	// Cache by group ID
	groupKey := fmt.Sprintf("group:%s", gr.GroupID)
	if err := r.groupRoleCache.Set(ctx, groupKey, gr, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache group role by group %s: %v", gr.GroupID, err)
	}

	// Cache by role ID
	roleKey := fmt.Sprintf("role:%s", gr.RoleID)
	if err := r.groupRoleCache.Set(ctx, roleKey, gr, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache group role by role %s: %v", gr.RoleID, err)
	}
}

// invalidateGroupRoleCache invalidates the cache for a group-role relationship.
func (r *groupRoleRepository) invalidateGroupRoleCache(ctx context.Context, groupID, roleID string) {
	// Invalidate relationship cache
	relationshipKey := fmt.Sprintf("relationship:%s:%s", groupID, roleID)
	if err := r.groupRoleCache.Delete(ctx, relationshipKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate group role relationship cache %s:%s: %v", groupID, roleID, err)
	}

	// Invalidate group key
	groupKey := fmt.Sprintf("group:%s", groupID)
	if err := r.groupRoleCache.Delete(ctx, groupKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate group role cache by group %s: %v", groupID, err)
	}

	// Invalidate role key
	roleKey := fmt.Sprintf("role:%s", roleID)
	if err := r.groupRoleCache.Delete(ctx, roleKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate group role cache by role %s: %v", roleID, err)
	}
}

// invalidateGroupRolesCache invalidates the cache for a group's roles.
func (r *groupRoleRepository) invalidateGroupRolesCache(ctx context.Context, groupID string) {
	cacheKey := fmt.Sprintf("group_roles:%s", groupID)
	if err := r.groupRolesCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate group roles cache %s: %v", groupID, err)
	}
}

// invalidateRoleGroupsCache invalidates the cache for a role's groups.
func (r *groupRoleRepository) invalidateRoleGroupsCache(ctx context.Context, roleID string) {
	cacheKey := fmt.Sprintf("role_groups:%s", roleID)
	if err := r.roleGroupsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate role groups cache %s: %v", roleID, err)
	}
}
