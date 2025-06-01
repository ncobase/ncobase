package repository

import (
	"context"
	"fmt"
	"ncobase/access/data"
	"ncobase/access/data/ent"
	permissionEnt "ncobase/access/data/ent/permission"
	roleEnt "ncobase/access/data/ent/role"
	rolePermissionEnt "ncobase/access/data/ent/rolepermission"
	"ncobase/access/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/logging/logger"
)

// RolePermissionRepositoryInterface represents the role permission repository interface.
type RolePermissionRepositoryInterface interface {
	Create(ctx context.Context, body *structs.RolePermission) (*ent.RolePermission, error)
	GetByPermissionID(ctx context.Context, permissionID string) (*ent.RolePermission, error)
	GetByRoleID(ctx context.Context, roleID string) (*ent.RolePermission, error)
	GetByPermissionIDs(ctx context.Context, permissionIDs []string) ([]*ent.RolePermission, error)
	GetByRoleIDs(ctx context.Context, roleIDs []string) ([]*ent.RolePermission, error)
	Delete(ctx context.Context, roleID, permissionID string) error
	DeleteAllByPermissionID(ctx context.Context, permissionID string) error
	DeleteAllByRoleID(ctx context.Context, roleID string) error
	GetPermissionsByRoleID(ctx context.Context, roleID string) ([]*ent.Permission, error)
	GetRolesByPermissionID(ctx context.Context, permissionID string) ([]*ent.Role, error)
	IsPermissionInRole(ctx context.Context, roleID, permissionID string) (bool, error)
	IsRoleInPermission(ctx context.Context, permissionID, roleID string) (bool, error)
}

// rolePermissionRepository implements the RolePermissionRepositoryInterface.
type rolePermissionRepository struct {
	data                 *data.Data
	rolePermissionCache  cache.ICache[ent.RolePermission]
	rolePermissionsCache cache.ICache[[]string] // Maps role ID to permission IDs
	permissionRolesCache cache.ICache[[]string] // Maps permission ID to role IDs
	relationshipTTL      time.Duration
}

// NewRolePermissionRepository creates a new role permission repository.
func NewRolePermissionRepository(d *data.Data) RolePermissionRepositoryInterface {
	redisClient := d.GetRedis()

	return &rolePermissionRepository{
		data:                 d,
		rolePermissionCache:  cache.NewCache[ent.RolePermission](redisClient, "ncse_access:role_permissions"),
		rolePermissionsCache: cache.NewCache[[]string](redisClient, "ncse_access:role_permission_mappings"),
		permissionRolesCache: cache.NewCache[[]string](redisClient, "ncse_access:permission_role_mappings"),
		relationshipTTL:      time.Hour * 2, // 2 hours cache TTL
	}
}

// Create role permission
func (r *rolePermissionRepository) Create(ctx context.Context, body *structs.RolePermission) (*ent.RolePermission, error) {
	// Use master for writes
	builder := r.data.GetMasterEntClient().RolePermission.Create()

	// Set values
	builder.SetNillableRoleID(&body.RoleID)
	builder.SetNillablePermissionID(&body.PermissionID)

	// Execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "rolePermissionRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the relationship and invalidate related caches
	go func() {
		r.cacheRolePermission(context.Background(), row)
		r.invalidateRolePermissionsCache(context.Background(), body.RoleID)
		r.invalidatePermissionRolesCache(context.Background(), body.PermissionID)
	}()

	return row, nil
}

// GetByPermissionID Find role permission by permission id
func (r *rolePermissionRepository) GetByPermissionID(ctx context.Context, id string) (*ent.RolePermission, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("permission:%s", id)
	if cached, err := r.rolePermissionCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Use slave for reads
	builder := r.data.GetSlaveEntClient().RolePermission.Query()

	// Set conditions
	builder.Where(rolePermissionEnt.PermissionIDEQ(id))

	// Execute the builder
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "rolePermissionRepo.GetByPermissionID error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheRolePermission(context.Background(), row)

	return row, nil
}

// GetByPermissionIDs Find role permissions by permission ids
func (r *rolePermissionRepository) GetByPermissionIDs(ctx context.Context, ids []string) ([]*ent.RolePermission, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().RolePermission.Query()

	// Set conditions
	builder.Where(rolePermissionEnt.PermissionIDIn(ids...))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "rolePermissionRepo.GetByPermissionIDs error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, rp := range rows {
			r.cacheRolePermission(context.Background(), rp)
		}
	}()

	return rows, nil
}

// GetByRoleID Find role permission by role id
func (r *rolePermissionRepository) GetByRoleID(ctx context.Context, id string) (*ent.RolePermission, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("role:%s", id)
	if cached, err := r.rolePermissionCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Use slave for reads
	builder := r.data.GetSlaveEntClient().RolePermission.Query()

	// Set conditions
	builder.Where(rolePermissionEnt.RoleIDEQ(id))

	// Execute the builder
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "rolePermissionRepo.GetByRoleID error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheRolePermission(context.Background(), row)

	return row, nil
}

// GetByRoleIDs Find role permissions by role ids
func (r *rolePermissionRepository) GetByRoleIDs(ctx context.Context, ids []string) ([]*ent.RolePermission, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().RolePermission.Query()

	// Set conditions
	builder.Where(rolePermissionEnt.RoleIDIn(ids...))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "rolePermissionRepo.GetByRoleIDs error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, rp := range rows {
			r.cacheRolePermission(context.Background(), rp)
		}
	}()

	return rows, nil
}

// Delete role permission
func (r *rolePermissionRepository) Delete(ctx context.Context, rid, pid string) error {
	// Use master for writes
	if _, err := r.data.GetMasterEntClient().RolePermission.Delete().Where(rolePermissionEnt.RoleIDEQ(rid), rolePermissionEnt.PermissionIDEQ(pid)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "rolePermissionRepo.Delete error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateRolePermissionCache(context.Background(), rid, pid)
		r.invalidateRolePermissionsCache(context.Background(), rid)
		r.invalidatePermissionRolesCache(context.Background(), pid)
	}()

	return nil
}

// DeleteAllByPermissionID Delete all role permission
func (r *rolePermissionRepository) DeleteAllByPermissionID(ctx context.Context, id string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.GetByPermissionIDs(ctx, []string{id})
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().RolePermission.Delete().Where(rolePermissionEnt.PermissionIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "rolePermissionRepo.DeleteAllByPermissionID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidatePermissionRolesCache(context.Background(), id)
		for _, rp := range relationships {
			r.invalidateRolePermissionCache(context.Background(), rp.RoleID, rp.PermissionID)
			r.invalidateRolePermissionsCache(context.Background(), rp.RoleID)
		}
	}()

	return nil
}

// DeleteAllByRoleID Delete all role permission
func (r *rolePermissionRepository) DeleteAllByRoleID(ctx context.Context, id string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.GetByRoleIDs(ctx, []string{id})
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().RolePermission.Delete().Where(rolePermissionEnt.RoleIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "rolePermissionRepo.DeleteAllByRoleID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateRolePermissionsCache(context.Background(), id)
		for _, rp := range relationships {
			r.invalidateRolePermissionCache(context.Background(), rp.RoleID, rp.PermissionID)
			r.invalidatePermissionRolesCache(context.Background(), rp.PermissionID)
		}
	}()

	return nil
}

// GetPermissionsByRoleID retrieves all permissions assigned to a role.
func (r *rolePermissionRepository) GetPermissionsByRoleID(ctx context.Context, rid string) ([]*ent.Permission, error) {
	// Try to get permission IDs from cache
	cacheKey := fmt.Sprintf("role_permissions:%s", rid)
	var permissionIDs []string
	if err := r.rolePermissionsCache.GetArray(ctx, cacheKey, &permissionIDs); err == nil && len(permissionIDs) > 0 {
		// Get permissions by IDs from permission repository
		return r.data.GetSlaveEntClient().Permission.Query().Where(permissionEnt.IDIn(permissionIDs...)).All(ctx)
	}

	// Fallback to database
	rolePermissions, err := r.data.GetSlaveEntClient().RolePermission.Query().Where(rolePermissionEnt.RoleIDEQ(rid)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "rolePermissionRepo.GetPermissionsByRoleID error: %v", err)
		return nil, err
	}

	// Extract permission ids from rolePermissions
	permissionIDs = make([]string, len(rolePermissions))
	for i, rolePermission := range rolePermissions {
		permissionIDs[i] = rolePermission.PermissionID
	}

	// Query permissions based on extracted permission ids
	permissions, err := r.data.GetSlaveEntClient().Permission.Query().Where(permissionEnt.IDIn(permissionIDs...)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "rolePermissionRepo.GetPermissionsByRoleID error: %v", err)
		return nil, err
	}

	// Cache permission IDs for future use
	go func() {
		if err := r.rolePermissionsCache.SetArray(context.Background(), cacheKey, permissionIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache role permissions %s: %v", rid, err)
		}
	}()

	return permissions, nil
}

// GetRolesByPermissionID retrieves all roles assigned to a permission.
func (r *rolePermissionRepository) GetRolesByPermissionID(ctx context.Context, pid string) ([]*ent.Role, error) {
	// Try to get role IDs from cache
	cacheKey := fmt.Sprintf("permission_roles:%s", pid)
	var roleIDs []string
	if err := r.permissionRolesCache.GetArray(ctx, cacheKey, &roleIDs); err == nil && len(roleIDs) > 0 {
		// Get roles by IDs from role repository
		return r.data.GetSlaveEntClient().Role.Query().Where(roleEnt.IDIn(roleIDs...)).All(ctx)
	}

	// Fallback to database
	rolePermissions, err := r.data.GetSlaveEntClient().RolePermission.Query().Where(rolePermissionEnt.PermissionIDEQ(pid)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "rolePermissionRepo.GetRolesByPermissionID error: %v", err)
		return nil, err
	}

	// Extract role IDs from rolePermissions
	roleIDs = make([]string, len(rolePermissions))
	for i, rolePermission := range rolePermissions {
		roleIDs[i] = rolePermission.RoleID
	}

	// Query roles based on extracted role IDs
	roles, err := r.data.GetSlaveEntClient().Role.Query().Where(roleEnt.IDIn(roleIDs...)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "rolePermissionRepo.GetRolesByPermissionID error: %v", err)
		return nil, err
	}

	// Cache role IDs for future use
	go func() {
		if err := r.permissionRolesCache.SetArray(context.Background(), cacheKey, roleIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache permission roles %s: %v", pid, err)
		}
	}()

	return roles, nil
}

// IsPermissionInRole verifies if a permission is assigned to a specific role.
func (r *rolePermissionRepository) IsPermissionInRole(ctx context.Context, rid, pid string) (bool, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("relationship:%s:%s", rid, pid)
	if cached, err := r.rolePermissionCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return true, nil
	}

	// Use slave for reads
	count, err := r.data.GetSlaveEntClient().RolePermission.Query().Where(rolePermissionEnt.RoleIDEQ(rid), rolePermissionEnt.PermissionIDEQ(pid)).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "rolePermissionRepo.IsPermissionInRole error: %v", err)
		return false, err
	}

	exists := count > 0

	// Cache the result if it exists
	if exists {
		go func() {
			// Create a dummy relationship for caching
			relationship := &ent.RolePermission{
				RoleID:       rid,
				PermissionID: pid,
			}
			r.cacheRolePermission(context.Background(), relationship)
		}()
	}

	return exists, nil
}

// IsRoleInPermission verifies if a role is assigned to a specific permission.
func (r *rolePermissionRepository) IsRoleInPermission(ctx context.Context, rid, pid string) (bool, error) {
	return r.IsPermissionInRole(ctx, rid, pid)
}

// cacheRolePermission caches a role permission relationship
func (r *rolePermissionRepository) cacheRolePermission(ctx context.Context, rp *ent.RolePermission) {
	// Cache by relationship key
	relationshipKey := fmt.Sprintf("relationship:%s:%s", rp.RoleID, rp.PermissionID)
	if err := r.rolePermissionCache.Set(ctx, relationshipKey, rp, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache role permission relationship %s:%s: %v", rp.RoleID, rp.PermissionID, err)
	}

	// Cache by role ID
	roleKey := fmt.Sprintf("role:%s", rp.RoleID)
	if err := r.rolePermissionCache.Set(ctx, roleKey, rp, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache role permission by role %s: %v", rp.RoleID, err)
	}

	// Cache by permission ID
	permissionKey := fmt.Sprintf("permission:%s", rp.PermissionID)
	if err := r.rolePermissionCache.Set(ctx, permissionKey, rp, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache role permission by permission %s: %v", rp.PermissionID, err)
	}
}

// invalidateRolePermissionCache invalidates role permission cache
func (r *rolePermissionRepository) invalidateRolePermissionCache(ctx context.Context, roleID, permissionID string) {
	// Invalidate relationship cache
	relationshipKey := fmt.Sprintf("relationship:%s:%s", roleID, permissionID)
	if err := r.rolePermissionCache.Delete(ctx, relationshipKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate role permission relationship cache %s:%s: %v", roleID, permissionID, err)
	}

	// Invalidate role key
	roleKey := fmt.Sprintf("role:%s", roleID)
	if err := r.rolePermissionCache.Delete(ctx, roleKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate role permission cache by role %s: %v", roleID, err)
	}

	// Invalidate permission key
	permissionKey := fmt.Sprintf("permission:%s", permissionID)
	if err := r.rolePermissionCache.Delete(ctx, permissionKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate role permission cache by permission %s: %v", permissionID, err)
	}
}

// invalidateRolePermissionsCache invalidates role permissions cache
func (r *rolePermissionRepository) invalidateRolePermissionsCache(ctx context.Context, roleID string) {
	cacheKey := fmt.Sprintf("role_permissions:%s", roleID)
	if err := r.rolePermissionsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate role permissions cache %s: %v", roleID, err)
	}
}

// invalidatePermissionRolesCache invalidates permission roles cache
func (r *rolePermissionRepository) invalidatePermissionRolesCache(ctx context.Context, permissionID string) {
	cacheKey := fmt.Sprintf("permission_roles:%s", permissionID)
	if err := r.permissionRolesCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate permission roles cache %s: %v", permissionID, err)
	}
}
