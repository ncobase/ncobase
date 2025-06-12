package repository

import (
	"context"
	"fmt"
	"ncobase/organization/data"
	"ncobase/organization/data/ent"
	organizationEnt "ncobase/organization/data/ent/organization"
	organizationRoleEnt "ncobase/organization/data/ent/organizationrole"
	"ncobase/organization/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/logging/logger"
)

// OrganizationRoleRepositoryInterface represents the organization role repository interface.
type OrganizationRoleRepositoryInterface interface {
	Create(ctx context.Context, body *structs.OrganizationRole) (*ent.OrganizationRole, error)
	GetByOrgID(ctx context.Context, id string) (*ent.OrganizationRole, error)
	GetByRoleID(ctx context.Context, id string) (*ent.OrganizationRole, error)
	GetByOrgIDs(ctx context.Context, ids []string) ([]*ent.OrganizationRole, error)
	GetByRoleIDs(ctx context.Context, ids []string) ([]*ent.OrganizationRole, error)
	Delete(ctx context.Context, oid, rid string) error
	DeleteAllByOrgID(ctx context.Context, id string) error
	DeleteAllByRoleID(ctx context.Context, id string) error
	GetRolesByOrgID(ctx context.Context, organizationID string) ([]string, error)
	GetOrganizationsByRoleID(ctx context.Context, roleID string) ([]*ent.Organization, error)
	IsRoleInOrganization(ctx context.Context, organizationID string, roleID string) (bool, error)
	IsOrganizationInRole(ctx context.Context, roleID string, organizationID string) (bool, error)
}

// organizationRoleRepository implements the OrganizationRoleRepositoryInterface.
type organizationRoleRepository struct {
	data                   *data.Data
	organizationRoleCache  cache.ICache[ent.OrganizationRole]
	organizationRolesCache cache.ICache[[]string] // Maps organization ID to role IDs
	roleOrganizationsCache cache.ICache[[]string] // Maps role ID to organization IDs
	relationshipTTL        time.Duration
}

// NewOrganizationRoleRepository creates a new organization role repository.
func NewOrganizationRoleRepository(d *data.Data) OrganizationRoleRepositoryInterface {
	redisClient := d.GetRedis()

	return &organizationRoleRepository{
		data:                   d,
		organizationRoleCache:  cache.NewCache[ent.OrganizationRole](redisClient, "ncse_organization:organization_roles"),
		organizationRolesCache: cache.NewCache[[]string](redisClient, "ncse_organization:organization_role_mappings"),
		roleOrganizationsCache: cache.NewCache[[]string](redisClient, "ncse_organization:role_organization_mappings"),
		relationshipTTL:        time.Hour * 2, // 2 hours cache TTL
	}
}

// Create organization role
func (r *organizationRoleRepository) Create(ctx context.Context, body *structs.OrganizationRole) (*ent.OrganizationRole, error) {
	// Use master for writes
	builder := r.data.GetMasterEntClient().OrganizationRole.Create()

	// Set values
	builder.SetNillableOrgID(&body.OrgID)
	builder.SetNillableRoleID(&body.RoleID)

	// Execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "organizationRoleRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the relationship and invalidate related caches
	go func() {
		r.cacheOrganizationRole(context.Background(), row)
		r.invalidateOrganizationRolesCache(context.Background(), body.OrgID)
		r.invalidateRoleOrganizationsCache(context.Background(), body.RoleID)
	}()

	return row, nil
}

// GetByOrgID Find role by organization id
func (r *organizationRoleRepository) GetByOrgID(ctx context.Context, id string) (*ent.OrganizationRole, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("organization:%s", id)
	if cached, err := r.organizationRoleCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Use slave for reads
	builder := r.data.GetSlaveEntClient().OrganizationRole.Query()

	// Set conditions
	builder.Where(organizationRoleEnt.OrgIDEQ(id))

	// Execute the builder
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "organizationRoleRepo.GetByOrgID error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheOrganizationRole(context.Background(), row)

	return row, nil
}

// GetByOrgIDs Find roles by organization ids
func (r *organizationRoleRepository) GetByOrgIDs(ctx context.Context, ids []string) ([]*ent.OrganizationRole, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().OrganizationRole.Query()

	// Set conditions
	builder.Where(organizationRoleEnt.OrgIDIn(ids...))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "organizationRoleRepo.GetByOrgIDs error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, or := range rows {
			r.cacheOrganizationRole(context.Background(), or)
		}
	}()

	return rows, nil
}

// GetByRoleID Find role by role id
func (r *organizationRoleRepository) GetByRoleID(ctx context.Context, id string) (*ent.OrganizationRole, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("role:%s", id)
	if cached, err := r.organizationRoleCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Use slave for reads
	builder := r.data.GetSlaveEntClient().OrganizationRole.Query()

	// Set conditions
	builder.Where(organizationRoleEnt.RoleIDEQ(id))

	// Execute the builder
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "organizationRoleRepo.GetByRoleID error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheOrganizationRole(context.Background(), row)

	return row, nil
}

// GetByRoleIDs Find roles by role ids
func (r *organizationRoleRepository) GetByRoleIDs(ctx context.Context, ids []string) ([]*ent.OrganizationRole, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().OrganizationRole.Query()

	// Set conditions
	builder.Where(organizationRoleEnt.RoleIDIn(ids...))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "organizationRoleRepo.GetByRoleIDs error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, or := range rows {
			r.cacheOrganizationRole(context.Background(), or)
		}
	}()

	return rows, nil
}

// Delete organization role
func (r *organizationRoleRepository) Delete(ctx context.Context, oid, rid string) error {
	// Use master for writes
	if _, err := r.data.GetMasterEntClient().OrganizationRole.Delete().
		Where(organizationRoleEnt.OrgIDEQ(oid), organizationRoleEnt.RoleIDEQ(rid)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "organizationRoleRepo.Delete error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateOrganizationRoleCache(context.Background(), oid, rid)
		r.invalidateOrganizationRolesCache(context.Background(), oid)
		r.invalidateRoleOrganizationsCache(context.Background(), rid)
	}()

	return nil
}

// DeleteAllByOrgID Delete all organization role
func (r *organizationRoleRepository) DeleteAllByOrgID(ctx context.Context, id string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.GetByOrgIDs(ctx, []string{id})
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().OrganizationRole.Delete().
		Where(organizationRoleEnt.OrgIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "organizationRoleRepo.DeleteAllByOrgID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateOrganizationRolesCache(context.Background(), id)
		for _, or := range relationships {
			r.invalidateOrganizationRoleCache(context.Background(), or.OrgID, or.RoleID)
			r.invalidateRoleOrganizationsCache(context.Background(), or.RoleID)
		}
	}()

	return nil
}

// DeleteAllByRoleID Delete all organization role
func (r *organizationRoleRepository) DeleteAllByRoleID(ctx context.Context, id string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.GetByRoleIDs(ctx, []string{id})
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().OrganizationRole.Delete().
		Where(organizationRoleEnt.RoleIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "organizationRoleRepo.DeleteAllByRoleID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateRoleOrganizationsCache(context.Background(), id)
		for _, or := range relationships {
			r.invalidateOrganizationRoleCache(context.Background(), or.OrgID, or.RoleID)
			r.invalidateOrganizationRolesCache(context.Background(), or.OrgID)
		}
	}()

	return nil
}

// GetRolesByOrgID retrieves all roles under an organization.
func (r *organizationRoleRepository) GetRolesByOrgID(ctx context.Context, organizationID string) ([]string, error) {
	// Try to get role IDs from cache
	cacheKey := fmt.Sprintf("organization_roles:%s", organizationID)
	var roleIDs []string
	if err := r.organizationRolesCache.GetArray(ctx, cacheKey, &roleIDs); err == nil && len(roleIDs) > 0 {
		return roleIDs, nil
	}

	// Fallback to database
	organizationRoles, err := r.data.GetSlaveEntClient().OrganizationRole.Query().
		Where(organizationRoleEnt.OrgIDEQ(organizationID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "organizationRoleRepo.GetRolesByOrgID error: %v", err)
		return nil, err
	}

	// Extract role IDs from organizationRoles
	roleIDs = make([]string, len(organizationRoles))
	for i, organizationRole := range organizationRoles {
		roleIDs[i] = organizationRole.RoleID
	}

	// Cache role IDs for future use
	go func() {
		if err := r.organizationRolesCache.SetArray(context.Background(), cacheKey, roleIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache organization roles %s: %v", organizationID, err)
		}
	}()

	return roleIDs, nil
}

// GetOrganizationsByRoleID retrieves all organizations under a role.
func (r *organizationRoleRepository) GetOrganizationsByRoleID(ctx context.Context, roleID string) ([]*ent.Organization, error) {
	// Try to get organization IDs from cache
	cacheKey := fmt.Sprintf("role_organizations:%s", roleID)
	var organizationIDs []string
	if err := r.roleOrganizationsCache.GetArray(ctx, cacheKey, &organizationIDs); err == nil && len(organizationIDs) > 0 {
		// Get organizations by IDs from organization repository
		return r.data.GetSlaveEntClient().Organization.Query().Where(organizationEnt.IDIn(organizationIDs...)).All(ctx)
	}

	// Fallback to database
	organizationRoles, err := r.data.GetSlaveEntClient().OrganizationRole.Query().
		Where(organizationRoleEnt.RoleIDEQ(roleID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "organizationRoleRepo.GetOrganizationsByRoleID error: %v", err)
		return nil, err
	}

	// Extract organization IDs from organizationRoles
	organizationIDs = make([]string, len(organizationRoles))
	for i, organizationRole := range organizationRoles {
		organizationIDs[i] = organizationRole.OrgID
	}

	// Query organizations based on extracted organization IDs
	organizations, err := r.data.GetSlaveEntClient().Organization.Query().Where(organizationEnt.IDIn(organizationIDs...)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "organizationRoleRepo.GetOrganizationsByRoleID error: %v", err)
		return nil, err
	}

	// Cache organization IDs for future use
	go func() {
		if err := r.roleOrganizationsCache.SetArray(context.Background(), cacheKey, organizationIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache role organizations %s: %v", roleID, err)
		}
	}()

	return organizations, nil
}

// IsRoleInOrganization verifies if a role belongs to a specific organization.
func (r *organizationRoleRepository) IsRoleInOrganization(ctx context.Context, organizationID string, roleID string) (bool, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("relationship:%s:%s", organizationID, roleID)
	if cached, err := r.organizationRoleCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return true, nil
	}

	// Use slave for reads
	count, err := r.data.GetSlaveEntClient().OrganizationRole.Query().
		Where(organizationRoleEnt.OrgIDEQ(organizationID), organizationRoleEnt.RoleIDEQ(roleID)).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "organizationRoleRepo.IsRoleInOrganization error: %v", err)
		return false, err
	}

	exists := count > 0

	// Cache the result if it exists
	if exists {
		go func() {
			// Create a dummy relationship for caching
			relationship := &ent.OrganizationRole{
				OrgID:  organizationID,
				RoleID: roleID,
			}
			r.cacheOrganizationRole(context.Background(), relationship)
		}()
	}

	return exists, nil
}

// IsOrganizationInRole verifies if an organization belongs to a specific role.
func (r *organizationRoleRepository) IsOrganizationInRole(ctx context.Context, organizationID string, roleID string) (bool, error) {
	return r.IsRoleInOrganization(ctx, organizationID, roleID)
}

// cacheOrganizationRole caches an organization-role relationship.
func (r *organizationRoleRepository) cacheOrganizationRole(ctx context.Context, or *ent.OrganizationRole) {
	// Cache by relationship key
	relationshipKey := fmt.Sprintf("relationship:%s:%s", or.OrgID, or.RoleID)
	if err := r.organizationRoleCache.Set(ctx, relationshipKey, or, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache organization role relationship %s:%s: %v", or.OrgID, or.RoleID, err)
	}

	// Cache by organization ID
	organizationKey := fmt.Sprintf("organization:%s", or.OrgID)
	if err := r.organizationRoleCache.Set(ctx, organizationKey, or, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache organization role by organization %s: %v", or.OrgID, err)
	}

	// Cache by role ID
	roleKey := fmt.Sprintf("role:%s", or.RoleID)
	if err := r.organizationRoleCache.Set(ctx, roleKey, or, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache organization role by role %s: %v", or.RoleID, err)
	}
}

// invalidateOrganizationRoleCache invalidates the cache for an organization-role relationship.
func (r *organizationRoleRepository) invalidateOrganizationRoleCache(ctx context.Context, organizationID, roleID string) {
	// Invalidate relationship cache
	relationshipKey := fmt.Sprintf("relationship:%s:%s", organizationID, roleID)
	if err := r.organizationRoleCache.Delete(ctx, relationshipKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate organization role relationship cache %s:%s: %v", organizationID, roleID, err)
	}

	// Invalidate organization key
	organizationKey := fmt.Sprintf("organization:%s", organizationID)
	if err := r.organizationRoleCache.Delete(ctx, organizationKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate organization role cache by organization %s: %v", organizationID, err)
	}

	// Invalidate role key
	roleKey := fmt.Sprintf("role:%s", roleID)
	if err := r.organizationRoleCache.Delete(ctx, roleKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate organization role cache by role %s: %v", roleID, err)
	}
}

// invalidateOrganizationRolesCache invalidates the cache for an organization's roles.
func (r *organizationRoleRepository) invalidateOrganizationRolesCache(ctx context.Context, organizationID string) {
	cacheKey := fmt.Sprintf("organization_roles:%s", organizationID)
	if err := r.organizationRolesCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate organization roles cache %s: %v", organizationID, err)
	}
}

// invalidateRoleOrganizationsCache invalidates the cache for a role's organizations.
func (r *organizationRoleRepository) invalidateRoleOrganizationsCache(ctx context.Context, roleID string) {
	cacheKey := fmt.Sprintf("role_organizations:%s", roleID)
	if err := r.roleOrganizationsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate role organizations cache %s: %v", roleID, err)
	}
}
