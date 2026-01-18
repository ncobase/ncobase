package repository

import (
	"context"
	"fmt"
	"ncobase/core/organization/data"
	"ncobase/core/organization/data/ent"
	userOrganizationEnt "ncobase/core/organization/data/ent/userorganization"
	"ncobase/core/organization/structs"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ncobase/ncore/data/cache"
	"github.com/ncobase/ncore/logging/logger"
)

// UserOrganizationRepositoryInterface represents the user organization repository interface.
type UserOrganizationRepositoryInterface interface {
	Create(ctx context.Context, body *structs.UserOrganization) (*ent.UserOrganization, error)
	GetByUserID(ctx context.Context, id string) ([]*ent.UserOrganization, error)
	GetByOrgID(ctx context.Context, id string) ([]*ent.UserOrganization, error)
	GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserOrganization, error)
	GetByOrgIDs(ctx context.Context, ids []string) ([]*ent.UserOrganization, error)
	GetByOrgIDAndRole(ctx context.Context, id string, role structs.UserRole) ([]*ent.UserOrganization, error)
	GetUserOrganization(ctx context.Context, uid, oid string) (*ent.UserOrganization, error)
	Delete(ctx context.Context, uid, oid string) error
	DeleteAllByUserID(ctx context.Context, id string) error
	DeleteAllByOrgID(ctx context.Context, id string) error
	GetOrganizationsByUserID(ctx context.Context, userID string) ([]string, error)
	GetUsersByOrgID(ctx context.Context, organizationID string) ([]string, error)
	IsUserInOrganization(ctx context.Context, userID string, organizationID string) (bool, error)
	UserHasRole(ctx context.Context, userID string, organizationID string, role structs.UserRole) (bool, error)
}

// userOrganizationRepository implements the UserOrganizationRepositoryInterface.
type userOrganizationRepository struct {
	data                       *data.Data
	userOrganizationCache      cache.ICache[ent.UserOrganization]
	userOrganizationsCache     cache.ICache[[]string] // Maps user ID to organization IDs
	organizationUsersCache     cache.ICache[[]string] // Maps organization ID to user IDs
	organizationRoleUsersCache cache.ICache[[]string] // Maps organization:role to user IDs
	relationshipTTL            time.Duration
}

// NewUserOrganizationRepository creates a new user organization repository.
func NewUserOrganizationRepository(d *data.Data) UserOrganizationRepositoryInterface {
	redisClient := d.GetRedis().(*redis.Client)

	return &userOrganizationRepository{
		data:                       d,
		userOrganizationCache:      cache.NewCache[ent.UserOrganization](redisClient, "ncse_organization:user_organizations"),
		userOrganizationsCache:     cache.NewCache[[]string](redisClient, "ncse_organization:user_organization_mappings"),
		organizationUsersCache:     cache.NewCache[[]string](redisClient, "ncse_organization:organization_user_mappings"),
		organizationRoleUsersCache: cache.NewCache[[]string](redisClient, "ncse_organization:organization_role_user_mappings"),
		relationshipTTL:            time.Hour * 2, // 2 hours cache TTL
	}
}

// Create create user organization
func (r *userOrganizationRepository) Create(ctx context.Context, body *structs.UserOrganization) (*ent.UserOrganization, error) {
	// Use master for writes
	builder := r.data.GetMasterEntClient().UserOrganization.Create()

	// Set values
	builder.SetUserID(body.UserID)
	builder.SetOrgID(body.OrgID)

	// Set role if provided
	if body.Role != "" {
		builder.SetRole(string(body.Role))
	} else {
		builder.SetRole(string(structs.RoleMember)) // Default role
	}

	// Execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "userOrganizationRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the relationship and invalidate related caches
	go func() {
		r.cacheUserOrganization(context.Background(), row)
		r.invalidateUserOrganizationsCache(context.Background(), body.UserID)
		r.invalidateOrganizationUsersCache(context.Background(), body.OrgID)
		if body.Role != "" {
			r.invalidateOrganizationRoleUsersCache(context.Background(), body.OrgID, string(body.Role))
		}
	}()

	return row, nil
}

// GetByUserID find organizations by user id
func (r *userOrganizationRepository) GetByUserID(ctx context.Context, id string) ([]*ent.UserOrganization, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserOrganization.Query()

	// Set conditions
	builder.Where(userOrganizationEnt.UserIDEQ(id))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userOrganizationRepo.GetByUserID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, uo := range rows {
			r.cacheUserOrganization(context.Background(), uo)
		}
	}()

	return rows, nil
}

// GetByUserIDs find organizations by user ids
func (r *userOrganizationRepository) GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserOrganization, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserOrganization.Query()

	// Set conditions
	builder.Where(userOrganizationEnt.UserIDIn(ids...))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userOrganizationRepo.GetByUserIDs error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, uo := range rows {
			r.cacheUserOrganization(context.Background(), uo)
		}
	}()

	return rows, nil
}

// GetByOrgID find users by organization id
func (r *userOrganizationRepository) GetByOrgID(ctx context.Context, id string) ([]*ent.UserOrganization, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserOrganization.Query()

	// Set conditions
	builder.Where(userOrganizationEnt.OrgIDEQ(id))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userOrganizationRepo.GetByOrgID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, uo := range rows {
			r.cacheUserOrganization(context.Background(), uo)
		}
	}()

	return rows, nil
}

// GetByOrgIDs find users by organization ids
func (r *userOrganizationRepository) GetByOrgIDs(ctx context.Context, ids []string) ([]*ent.UserOrganization, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserOrganization.Query()

	// Set conditions
	builder.Where(userOrganizationEnt.OrgIDIn(ids...))

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userOrganizationRepo.GetByOrgIDs error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, uo := range rows {
			r.cacheUserOrganization(context.Background(), uo)
		}
	}()

	return rows, nil
}

// GetByOrgIDAndRole find users by organization id and role
func (r *userOrganizationRepository) GetByOrgIDAndRole(ctx context.Context, id string, role structs.UserRole) ([]*ent.UserOrganization, error) {
	// Try to get user IDs from cache
	cacheKey := fmt.Sprintf("organization_role_users:%s:%s", id, string(role))
	var userIDs []string
	if err := r.organizationRoleUsersCache.GetArray(ctx, cacheKey, &userIDs); err == nil && len(userIDs) > 0 {
		// Get user organizations by user IDs and organization ID
		return r.data.GetSlaveEntClient().UserOrganization.Query().
			Where(userOrganizationEnt.OrgIDEQ(id), userOrganizationEnt.UserIDIn(userIDs...)).All(ctx)
	}

	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserOrganization.Query()

	// Set conditions
	builder.Where(
		userOrganizationEnt.OrgIDEQ(id),
		userOrganizationEnt.RoleEQ(string(role)),
	)

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userOrganizationRepo.GetByOrgIDAndRole error: %v", err)
		return nil, err
	}

	// Cache user IDs for future use
	go func() {
		userIDs := make([]string, 0, len(rows))
		for _, uo := range rows {
			userIDs = append(userIDs, uo.UserID)
			r.cacheUserOrganization(context.Background(), uo)
		}
		if err := r.organizationRoleUsersCache.SetArray(context.Background(), cacheKey, userIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache organization role users %s:%s: %v", id, string(role), err)
		}
	}()

	return rows, nil
}

// GetUserOrganization gets a specific user-organization relation
func (r *userOrganizationRepository) GetUserOrganization(ctx context.Context, uid, oid string) (*ent.UserOrganization, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("relationship:%s:%s", uid, oid)
	if cached, err := r.userOrganizationCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Use slave for reads
	builder := r.data.GetSlaveEntClient().UserOrganization.Query()

	// Set conditions
	builder.Where(
		userOrganizationEnt.UserIDEQ(uid),
		userOrganizationEnt.OrgIDEQ(oid),
	)

	// Execute the builder
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "userOrganizationRepo.GetUserOrganization error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheUserOrganization(context.Background(), row)

	return row, nil
}

// Delete delete user organization
func (r *userOrganizationRepository) Delete(ctx context.Context, uid, oid string) error {
	// Get existing relationship for cache invalidation
	userOrganization, err := r.GetUserOrganization(ctx, uid, oid)
	if err != nil {
		logger.Debugf(ctx, "Failed to get user organization for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserOrganization.Delete().
		Where(userOrganizationEnt.UserIDEQ(uid), userOrganizationEnt.OrgIDEQ(oid)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userOrganizationRepo.Delete error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateUserOrganizationCache(context.Background(), uid, oid)
		r.invalidateUserOrganizationsCache(context.Background(), uid)
		r.invalidateOrganizationUsersCache(context.Background(), oid)
		if userOrganization != nil {
			r.invalidateOrganizationRoleUsersCache(context.Background(), oid, userOrganization.Role)
		}
	}()

	return nil
}

// DeleteAllByUserID delete all user organization by user id
func (r *userOrganizationRepository) DeleteAllByUserID(ctx context.Context, id string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.GetByUserID(ctx, id)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserOrganization.Delete().
		Where(userOrganizationEnt.UserIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userOrganizationRepo.DeleteAllByUserID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateUserOrganizationsCache(context.Background(), id)
		for _, uo := range relationships {
			r.invalidateUserOrganizationCache(context.Background(), uo.UserID, uo.OrgID)
			r.invalidateOrganizationUsersCache(context.Background(), uo.OrgID)
			r.invalidateOrganizationRoleUsersCache(context.Background(), uo.OrgID, uo.Role)
		}
	}()

	return nil
}

// DeleteAllByOrgID delete all user organization by organization id
func (r *userOrganizationRepository) DeleteAllByOrgID(ctx context.Context, id string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.GetByOrgID(ctx, id)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	// Use master for writes
	if _, err := r.data.GetMasterEntClient().UserOrganization.Delete().
		Where(userOrganizationEnt.OrgIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userOrganizationRepo.DeleteAllByOrgID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateOrganizationUsersCache(context.Background(), id)
		for _, uo := range relationships {
			r.invalidateUserOrganizationCache(context.Background(), uo.UserID, uo.OrgID)
			r.invalidateUserOrganizationsCache(context.Background(), uo.UserID)
			r.invalidateOrganizationRoleUsersCache(context.Background(), id, uo.Role)
		}
	}()

	return nil
}

// GetOrganizationsByUserID retrieves all organizations a user belongs to.
func (r *userOrganizationRepository) GetOrganizationsByUserID(ctx context.Context, userID string) ([]string, error) {
	// Try to get organization IDs from cache
	cacheKey := fmt.Sprintf("user_organizations:%s", userID)
	var organizationIDs []string
	if err := r.userOrganizationsCache.GetArray(ctx, cacheKey, &organizationIDs); err == nil && len(organizationIDs) > 0 {
		return organizationIDs, nil
	}

	// Fallback to database
	userOrganizations, err := r.data.GetSlaveEntClient().UserOrganization.Query().
		Where(userOrganizationEnt.UserIDEQ(userID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userOrganizationRepo.GetOrganizationsByUserID error: %v", err)
		return nil, err
	}

	// Extract organization IDs from userOrganizations
	organizationIDs = make([]string, len(userOrganizations))
	for i, organization := range userOrganizations {
		organizationIDs[i] = organization.OrgID
	}

	// Cache organization IDs for future use
	go func() {
		if err := r.userOrganizationsCache.SetArray(context.Background(), cacheKey, organizationIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache user organizations %s: %v", userID, err)
		}
	}()

	return organizationIDs, nil
}

// GetUsersByOrgID retrieves all users in an organization.
func (r *userOrganizationRepository) GetUsersByOrgID(ctx context.Context, organizationID string) ([]string, error) {
	// Try to get user IDs from cache
	cacheKey := fmt.Sprintf("organization_users:%s", organizationID)
	var userIDs []string
	if err := r.organizationUsersCache.GetArray(ctx, cacheKey, &userIDs); err == nil && len(userIDs) > 0 {
		return userIDs, nil
	}

	// Fallback to database
	userOrganizations, err := r.data.GetSlaveEntClient().UserOrganization.Query().
		Where(userOrganizationEnt.OrgIDEQ(organizationID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userOrganizationRepo.GetUsersByOrgID error: %v", err)
		return nil, err
	}

	// Extract user IDs from userOrganizations
	userIDs = make([]string, len(userOrganizations))
	for i, userOrganization := range userOrganizations {
		userIDs[i] = userOrganization.UserID
	}

	// Cache user IDs for future use
	go func() {
		if err := r.organizationUsersCache.SetArray(context.Background(), cacheKey, userIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache organization users %s: %v", organizationID, err)
		}
	}()

	return userIDs, nil
}

// IsUserInOrganization verifies if a user belongs to a specific organization.
func (r *userOrganizationRepository) IsUserInOrganization(ctx context.Context, userID string, organizationID string) (bool, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("relationship:%s:%s", userID, organizationID)
	if cached, err := r.userOrganizationCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return true, nil
	}

	// Use slave for reads
	count, err := r.data.GetSlaveEntClient().UserOrganization.Query().
		Where(userOrganizationEnt.UserIDEQ(userID), userOrganizationEnt.OrgIDEQ(organizationID)).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "userOrganizationRepo.IsUserInOrganization error: %v", err)
		return false, err
	}

	exists := count > 0

	// Cache the result if it exists
	if exists {
		go func() {
			// Get the actual relationship for caching
			if userOrganization, err := r.GetUserOrganization(context.Background(), userID, organizationID); err == nil {
				r.cacheUserOrganization(context.Background(), userOrganization)
			}
		}()
	}

	return exists, nil
}

// UserHasRole verifies if a user has a specific role in an organization.
func (r *userOrganizationRepository) UserHasRole(ctx context.Context, userID string, organizationID string, role structs.UserRole) (bool, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("relationship:%s:%s", userID, organizationID)
	if cached, err := r.userOrganizationCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached.Role == string(role), nil
	}

	// Use slave for reads
	count, err := r.data.GetSlaveEntClient().UserOrganization.Query().Where(
		userOrganizationEnt.UserIDEQ(userID),
		userOrganizationEnt.OrgIDEQ(organizationID),
		userOrganizationEnt.RoleEQ(string(role)),
	).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "userOrganizationRepo.UserHasRole error: %v", err)
		return false, err
	}

	exists := count > 0

	// Cache the result if it exists
	if exists {
		go func() {
			// Get the actual relationship for caching
			if userOrganization, err := r.GetUserOrganization(context.Background(), userID, organizationID); err == nil {
				r.cacheUserOrganization(context.Background(), userOrganization)
			}
		}()
	}

	return exists, nil
}

// cacheUserOrganization caches a user organization relationship.
func (r *userOrganizationRepository) cacheUserOrganization(ctx context.Context, uo *ent.UserOrganization) {
	// Cache by relationship key
	relationshipKey := fmt.Sprintf("relationship:%s:%s", uo.UserID, uo.OrgID)
	if err := r.userOrganizationCache.Set(ctx, relationshipKey, uo, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache user organization relationship %s:%s: %v", uo.UserID, uo.OrgID, err)
	}
}

// invalidateUserOrganizationCache invalidates a user organization relationship cache.
func (r *userOrganizationRepository) invalidateUserOrganizationCache(ctx context.Context, userID, organizationID string) {
	// Invalidate relationship cache
	relationshipKey := fmt.Sprintf("relationship:%s:%s", userID, organizationID)
	if err := r.userOrganizationCache.Delete(ctx, relationshipKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user organization relationship cache %s:%s: %v", userID, organizationID, err)
	}
}

// invalidateUserOrganizationsCache invalidates the cache for a user's organizations.
func (r *userOrganizationRepository) invalidateUserOrganizationsCache(ctx context.Context, userID string) {
	cacheKey := fmt.Sprintf("user_organizations:%s", userID)
	if err := r.userOrganizationsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user organizations cache %s: %v", userID, err)
	}
}

// invalidateOrganizationUsersCache invalidates the cache for an organization's users.
func (r *userOrganizationRepository) invalidateOrganizationUsersCache(ctx context.Context, organizationID string) {
	cacheKey := fmt.Sprintf("organization_users:%s", organizationID)
	if err := r.organizationUsersCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate organization users cache %s: %v", organizationID, err)
	}
}

// invalidateOrganizationRoleUsersCache invalidates the cache for an organization-role relationship.
func (r *userOrganizationRepository) invalidateOrganizationRoleUsersCache(ctx context.Context, organizationID, role string) {
	cacheKey := fmt.Sprintf("organization_role_users:%s:%s", organizationID, role)
	if err := r.organizationRoleUsersCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate organization role users cache %s:%s: %v", organizationID, role, err)
	}
}
