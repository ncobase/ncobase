package repository

import (
	"context"
	"fmt"
	"ncobase/access/data"
	"ncobase/access/data/ent"
	roleEnt "ncobase/access/data/ent/role"
	"ncobase/access/structs"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/redis/go-redis/v9"
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
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.Role]
}

// NewRoleRepository creates a new role repository.
func NewRoleRepository(d *data.Data) RoleRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &roleRepository{ec, rc, cache.NewCache[ent.Role](rc, "ncse_role")}
}

// Create creates a new role.
func (r *roleRepository) Create(ctx context.Context, body *structs.CreateRoleBody) (*ent.Role, error) {
	// create builder.
	builder := r.ec.Role.Create()
	// set values.
	builder.SetNillableName(&body.Name)
	builder.SetNillableSlug(&body.Slug)
	builder.SetDisabled(body.Disabled)
	builder.SetNillableDescription(&body.Description)
	builder.SetNillableCreatedBy(body.CreatedBy)

	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "roleRepo.Create error: %v", err)
		return nil, err
	}

	return row, nil
}

// GetByID gets a role by ID.
func (r *roleRepository) GetByID(ctx context.Context, id string) (*ent.Role, error) {
	// check cache
	cacheKey := fmt.Sprintf("%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindRole(ctx, &structs.FindRole{ID: id})
	if err != nil {
		logger.Errorf(ctx, "roleRepo.GetByID error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "roleRepo.GetByID cache error: %v", err)
	}

	return row, nil
}

// GetByIDs gets roles by IDs.
func (r *roleRepository) GetByIDs(ctx context.Context, ids []string) ([]*ent.Role, error) {
	// create builder.
	builder := r.ec.Role.Query()
	// set conditions.
	builder.Where(roleEnt.IDIn(ids...))
	// execute the builder.
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "roleRepo.GetByIDs error: %v", err)
		return nil, err
	}
	return rows, nil
}

// GetBySlug gets a role by slug.
func (r *roleRepository) GetBySlug(ctx context.Context, slug string) (*ent.Role, error) {
	// check cache
	cacheKey := fmt.Sprintf("%s", slug)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindRole(ctx, &structs.FindRole{Slug: slug})
	if err != nil {
		logger.Errorf(ctx, "roleRepo.GetBySlug error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "roleRepo.GetBySlug cache error: %v", err)
	}

	return row, nil
}

// Update updates a role (full or partial).
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

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "roleRepo.Update error: %v", err)
		return nil, err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", role.ID)
	err = r.c.Delete(ctx, cacheKey)
	err = r.c.Delete(ctx, fmt.Sprintf("role:slug:%s", role.Slug))
	if err != nil {
		logger.Errorf(ctx, "roleRepo.Update cache error: %v", err)
	}

	return row, nil
}

// List gets a list of roles.
func (r *roleRepository) List(ctx context.Context, params *structs.ListRoleParams) ([]*ent.Role, error) {
	// create list builder
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

	return rows, nil
}

// Delete deletes a role.
func (r *roleRepository) Delete(ctx context.Context, slug string) error {
	role, err := r.FindRole(ctx, &structs.FindRole{Slug: slug})
	if err != nil {
		return err
	}

	// create builder.
	builder := r.ec.Role.Delete()

	// execute the builder and verify the result.
	if _, err = builder.Where(roleEnt.IDEQ(role.ID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "roleRepo.Delete error: %v", err)
		return err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", role.ID)
	err = r.c.Delete(ctx, cacheKey)
	err = r.c.Delete(ctx, fmt.Sprintf("role:slug:%s", role.Slug))
	if err != nil {
		logger.Errorf(ctx, "roleRepo.Delete cache error: %v", err)
	}

	return nil
}

// FindRole finds a role.
func (r *roleRepository) FindRole(ctx context.Context, params *structs.FindRole) (*ent.Role, error) {
	// create builder.
	builder := r.ec.Role.Query()

	if validator.IsNotEmpty(params.ID) {
		builder = builder.Where(roleEnt.IDEQ(params.ID))
	}
	// support slug or ID
	if validator.IsNotEmpty(params.Slug) {
		builder = builder.Where(roleEnt.Or(
			roleEnt.ID(params.Slug),
			roleEnt.SlugEQ(params.Slug),
		))
	}

	// execute the builder.
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// listBuilder creates list builder.
func (r *roleRepository) listBuilder(_ context.Context, _ *structs.ListRoleParams) (*ent.RoleQuery, error) {
	// create builder.
	builder := r.ec.Role.Query()

	return builder, nil
}

// CountX gets a count of roles.
func (r *roleRepository) CountX(ctx context.Context, params *structs.ListRoleParams) int {
	// create list builder
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}
