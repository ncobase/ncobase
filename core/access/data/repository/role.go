package repository

import (
	"context"
	"fmt"
	"ncobase/access/data"
	"ncobase/access/data/ent"
	roleEnt "ncobase/access/data/ent/role"
	"ncobase/access/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"
)

// RoleRepositoryInterface represents the role repository interface.
type RoleRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateRoleBody) (*ent.Role, error)
	GetByID(ctx context.Context, id string) (*ent.Role, error)
	GetByIDs(ctx context.Context, ids []string) ([]*ent.Role, error)
	GetBySlug(ctx context.Context, slug string) (*ent.Role, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*ent.Role, error)
	List(ctx context.Context, params *structs.ListRoleParams) ([]*ent.Role, error)
	Delete(ctx context.Context, slug string) error
	FindRole(ctx context.Context, params *structs.FindRole) (*ent.Role, error)
	CountX(ctx context.Context, params *structs.ListRoleParams) int
}

// roleRepository implements the RoleRepositoryInterface.
type roleRepository struct {
	ec               *ent.Client
	roleCache        cache.ICache[ent.Role]
	slugMappingCache cache.ICache[string] // Maps slug to role ID
	roleTTL          time.Duration
}

// NewRoleRepository creates a new role repository.
func NewRoleRepository(d *data.Data) RoleRepositoryInterface {
	redisClient := d.GetRedis()

	return &roleRepository{
		ec:               d.GetMasterEntClient(),
		roleCache:        cache.NewCache[ent.Role](redisClient, "ncse_access:roles"),
		slugMappingCache: cache.NewCache[string](redisClient, "ncse_access:role_mappings"),
		roleTTL:          time.Hour * 2, // 2 hours cache TTL
	}
}

// Create creates a new role
func (r *roleRepository) Create(ctx context.Context, body *structs.CreateRoleBody) (*ent.Role, error) {
	builder := r.ec.Role.Create()
	builder.SetNillableName(&body.Name)
	builder.SetNillableSlug(&body.Slug)
	builder.SetDisabled(body.Disabled)
	builder.SetNillableDescription(&body.Description)
	builder.SetNillableCreatedBy(body.CreatedBy)

	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	role, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "roleRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the role
	go r.cacheRole(context.Background(), role)

	return role, nil
}

// GetByID gets a role by ID
func (r *roleRepository) GetByID(ctx context.Context, id string) (*ent.Role, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("id:%s", id)
	if cached, err := r.roleCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Fallback to database
	row, err := r.FindRole(ctx, &structs.FindRole{ID: id})
	if err != nil {
		logger.Errorf(ctx, "roleRepo.GetByID error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheRole(context.Background(), row)

	return row, nil
}

// GetByIDs gets roles by IDs with batch caching
func (r *roleRepository) GetByIDs(ctx context.Context, ids []string) ([]*ent.Role, error) {
	// Try to get from cache first
	cacheKeys := make([]string, len(ids))
	for i, id := range ids {
		cacheKeys[i] = fmt.Sprintf("id:%s", id)
	}

	cachedRoles, err := r.roleCache.GetMultiple(ctx, cacheKeys)
	if err == nil && len(cachedRoles) == len(ids) {
		// All roles found in cache
		roles := make([]*ent.Role, len(ids))
		for i, key := range cacheKeys {
			if role, exists := cachedRoles[key]; exists {
				roles[i] = role
			}
		}
		return roles, nil
	}

	// Fallback to database
	builder := r.ec.Role.Query()
	builder.Where(roleEnt.IDIn(ids...))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "roleRepo.GetByIDs error: %v", err)
		return nil, err
	}

	// Cache roles in background
	go func() {
		for _, role := range rows {
			r.cacheRole(context.Background(), role)
		}
	}()

	return rows, nil
}

// GetBySlug gets a role by slug
func (r *roleRepository) GetBySlug(ctx context.Context, slug string) (*ent.Role, error) {
	// Try to get role ID from slug mapping cache
	if roleID, err := r.getRoleIDBySlug(ctx, slug); err == nil && roleID != "" {
		return r.GetByID(ctx, roleID)
	}

	// Fallback to database
	row, err := r.FindRole(ctx, &structs.FindRole{Slug: slug})
	if err != nil {
		logger.Errorf(ctx, "roleRepo.GetBySlug error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheRole(context.Background(), row)

	return row, nil
}

// Update updates a role
func (r *roleRepository) Update(ctx context.Context, slug string, updates types.JSON) (*ent.Role, error) {
	role, err := r.FindRole(ctx, &structs.FindRole{Slug: slug})
	if err != nil {
		return nil, err
	}

	builder := role.Update()

	for field, value := range updates {
		switch field {
		case "name":
			builder.SetNillableName(convert.ToPointer(value.(string)))
		case "slug":
			builder.SetNillableSlug(convert.ToPointer(value.(string)))
		case "disabled":
			builder.SetDisabled(value.(bool))
		case "description":
			builder.SetNillableDescription(convert.ToPointer(value.(string)))
		case "extra_props":
			builder.SetExtras(value.(types.JSON))
		case "updated_by":
			builder.SetUpdatedBy(value.(string))
		}
	}

	updatedRole, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "roleRepo.Update error: %v", err)
		return nil, err
	}

	// Invalidate and re-cache
	go func() {
		r.invalidateRoleCache(context.Background(), role)
		r.cacheRole(context.Background(), updatedRole)
	}()

	return updatedRole, nil
}

// List gets a list of roles
func (r *roleRepository) List(ctx context.Context, params *structs.ListRoleParams) ([]*ent.Role, error) {
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return nil, err
	}

	if params.Cursor != "" {
		id, timestamp, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %v", err)
		}

		if !nanoid.IsPrimaryKey(id) {
			return nil, fmt.Errorf("invalid id in cursor: %s", id)
		}

		if params.Direction == "backward" {
			builder.Where(
				roleEnt.Or(
					roleEnt.CreatedAtGT(timestamp),
					roleEnt.And(
						roleEnt.CreatedAtEQ(timestamp),
						roleEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				roleEnt.Or(
					roleEnt.CreatedAtLT(timestamp),
					roleEnt.And(
						roleEnt.CreatedAtEQ(timestamp),
						roleEnt.IDLT(id),
					),
				),
			)
		}
	}

	if params.Direction == "backward" {
		builder.Order(ent.Asc(roleEnt.FieldCreatedAt), ent.Asc(roleEnt.FieldID))
	} else {
		builder.Order(ent.Desc(roleEnt.FieldCreatedAt), ent.Desc(roleEnt.FieldID))
	}

	builder.Limit(params.Limit)

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "roleRepo.List error: %v", err)
		return nil, err
	}

	// Cache roles in background
	go func() {
		for _, role := range rows {
			r.cacheRole(context.Background(), role)
		}
	}()

	return rows, nil
}

// Delete deletes a role
func (r *roleRepository) Delete(ctx context.Context, slug string) error {
	role, err := r.FindRole(ctx, &structs.FindRole{Slug: slug})
	if err != nil {
		return err
	}

	builder := r.ec.Role.Delete()
	if _, err = builder.Where(roleEnt.IDEQ(role.ID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "roleRepo.Delete error: %v", err)
		return err
	}

	// Invalidate cache
	go r.invalidateRoleCache(context.Background(), role)

	return nil
}

// FindRole finds a role
func (r *roleRepository) FindRole(ctx context.Context, params *structs.FindRole) (*ent.Role, error) {
	builder := r.ec.Role.Query()

	if validator.IsNotEmpty(params.ID) {
		builder = builder.Where(roleEnt.IDEQ(params.ID))
	}
	if validator.IsNotEmpty(params.Slug) {
		builder = builder.Where(roleEnt.Or(
			roleEnt.ID(params.Slug),
			roleEnt.SlugEQ(params.Slug),
		))
	}

	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// listBuilder creates list builder
func (r *roleRepository) listBuilder(_ context.Context, _ *structs.ListRoleParams) (*ent.RoleQuery, error) {
	return r.ec.Role.Query(), nil
}

// CountX gets a count of roles
func (r *roleRepository) CountX(ctx context.Context, params *structs.ListRoleParams) int {
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// cacheRole caches a role
func (r *roleRepository) cacheRole(ctx context.Context, role *ent.Role) {
	// Cache by ID
	idKey := fmt.Sprintf("id:%s", role.ID)
	if err := r.roleCache.Set(ctx, idKey, role, r.roleTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache role by ID %s: %v", role.ID, err)
	}

	// Cache slug to ID mapping
	if role.Slug != "" {
		slugKey := fmt.Sprintf("slug:%s", role.Slug)
		if err := r.slugMappingCache.Set(ctx, slugKey, &role.ID, r.roleTTL); err != nil {
			logger.Debugf(ctx, "Failed to cache slug mapping %s: %v", role.Slug, err)
		}
	}
}

// invalidateRoleCache invalidates a role cache
func (r *roleRepository) invalidateRoleCache(ctx context.Context, role *ent.Role) {
	// Invalidate ID cache
	idKey := fmt.Sprintf("id:%s", role.ID)
	if err := r.roleCache.Delete(ctx, idKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate role ID cache %s: %v", role.ID, err)
	}

	// Invalidate slug mapping
	if role.Slug != "" {
		slugKey := fmt.Sprintf("slug:%s", role.Slug)
		if err := r.slugMappingCache.Delete(ctx, slugKey); err != nil {
			logger.Debugf(ctx, "Failed to invalidate slug mapping cache %s: %v", role.Slug, err)
		}
	}
}

// getRoleIDBySlug gets a role ID by slug
func (r *roleRepository) getRoleIDBySlug(ctx context.Context, slug string) (string, error) {
	cacheKey := fmt.Sprintf("slug:%s", slug)
	roleID, err := r.slugMappingCache.Get(ctx, cacheKey)
	if err != nil || roleID == nil {
		return "", err
	}
	return *roleID, nil
}
