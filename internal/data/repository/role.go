package repo

import (
	"context"
	"fmt"
	"ncobase/internal/data"
	"ncobase/internal/data/ent"
	roleEnt "ncobase/internal/data/ent/role"
	"ncobase/internal/data/structs"

	"ncobase/common/cache"
	"ncobase/common/log"
	"ncobase/common/types"
	"ncobase/common/validator"

	"github.com/redis/go-redis/v9"
)

// Role represents the role repository interface.
type Role interface {
	Create(ctx context.Context, body *structs.CreateRoleBody) (*ent.Role, error)
	CreateSuperAdminRole(ctx context.Context) (*ent.Role, error)
	GetByID(ctx context.Context, id string) (*ent.Role, error)
	GetBySlug(ctx context.Context, slug string) (*ent.Role, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*ent.Role, error)
	List(ctx context.Context, params *structs.ListRoleParams) ([]*ent.Role, error)
	Delete(ctx context.Context, slug string) error
	FindRole(ctx context.Context, params *structs.FindRole) (*ent.Role, error)
	ListBuilder(ctx context.Context, params *structs.ListRoleParams) (*ent.RoleQuery, error)
	CountX(ctx context.Context, params *structs.ListRoleParams) int
}

// roleRepo implements the Role interface.
type roleRepo struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.Role]
}

// NewRole creates a new role repository.
func NewRole(d *data.Data) Role {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &roleRepo{ec, rc, cache.NewCache[ent.Role](rc, cache.Key("nb_role"))}
}

// Create creates a new role.
func (r *roleRepo) Create(ctx context.Context, body *structs.CreateRoleBody) (*ent.Role, error) {
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
		log.Errorf(context.Background(), "roleRepo.Create error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// CreateSuperAdminRole creates a new super admin role.
func (r *roleRepo) CreateSuperAdminRole(ctx context.Context) (*ent.Role, error) {
	return r.Create(ctx, &structs.CreateRoleBody{
		RoleBody: structs.RoleBody{
			Name:        "Super Admin",
			Slug:        "super-admin",
			Disabled:    false,
			Description: "Super Admin Role",
			Extras:      &types.JSON{},
		},
	})
}

// GetByID gets a role by ID.
func (r *roleRepo) GetByID(ctx context.Context, id string) (*ent.Role, error) {
	// check cache
	cacheKey := fmt.Sprintf("%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindRole(ctx, &structs.FindRole{ID: id})
	if err != nil {
		log.Errorf(context.Background(), "roleRepo.GetByID error: %v\n", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(context.Background(), "roleRepo.GetByID cache error: %v\n", err)
	}

	return row, nil
}

// GetBySlug gets a role by slug.
func (r *roleRepo) GetBySlug(ctx context.Context, slug string) (*ent.Role, error) {
	// check cache
	cacheKey := fmt.Sprintf("%s", slug)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindRole(ctx, &structs.FindRole{Slug: slug})
	if err != nil {
		log.Errorf(context.Background(), "roleRepo.GetBySlug error: %v\n", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(context.Background(), "roleRepo.GetBySlug cache error: %v\n", err)
	}

	return row, nil
}

// Update updates a role (full or partial).
func (r *roleRepo) Update(ctx context.Context, slug string, updates types.JSON) (*ent.Role, error) {
	role, err := r.FindRole(ctx, &structs.FindRole{Slug: slug})
	if err != nil {
		return nil, err
	}

	builder := role.Update()

	for field, value := range updates {
		switch field {
		case "name":
			builder.SetNillableName(types.ToPointer(value.(string)))
		case "slug":
			builder.SetNillableSlug(types.ToPointer(value.(string)))
		case "disabled":
			builder.SetDisabled(value.(bool))
		case "description":
			builder.SetNillableDescription(types.ToPointer(value.(string)))
		case "extra_props":
			builder.SetExtras(value.(types.JSON))
		case "updated_by":
			builder.SetUpdatedBy(value.(string))
		}
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(context.Background(), "roleRepo.Update error: %v\n", err)
		return nil, err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", role.ID)
	err = r.c.Delete(ctx, cacheKey)
	err = r.c.Delete(ctx, fmt.Sprintf("role:slug:%s", role.Slug))
	if err != nil {
		log.Errorf(context.Background(), "roleRepo.Update cache error: %v\n", err)
	}

	return row, nil
}

// List gets a list of roles.
func (r *roleRepo) List(ctx context.Context, params *structs.ListRoleParams) ([]*ent.Role, error) {
	// create list builder
	builder, err := r.ListBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return nil, err
	}

	// limit the result
	builder.Limit(int(params.Limit))

	rows, err := builder.All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "roleRepo.List error: %v\n", err)
		return nil, err
	}

	return rows, nil
}

// Delete deletes a role.
func (r *roleRepo) Delete(ctx context.Context, slug string) error {
	role, err := r.FindRole(ctx, &structs.FindRole{Slug: slug})
	if err != nil {
		return err
	}

	// create builder.
	builder := r.ec.Role.Delete()

	// execute the builder and verify the result.
	if _, err = builder.Where(roleEnt.IDEQ(role.ID)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "roleRepo.Delete error: %v\n", err)
		return err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", role.ID)
	err = r.c.Delete(ctx, cacheKey)
	err = r.c.Delete(ctx, fmt.Sprintf("role:slug:%s", role.Slug))
	if err != nil {
		log.Errorf(context.Background(), "roleRepo.Delete cache error: %v\n", err)
	}

	return nil
}

// FindRole finds a role.
func (r *roleRepo) FindRole(ctx context.Context, params *structs.FindRole) (*ent.Role, error) {
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

// ListBuilder creates list builder.
func (r *roleRepo) ListBuilder(ctx context.Context, params *structs.ListRoleParams) (*ent.RoleQuery, error) {
	// Here you can construct and return a builder for listing roles based on the provided parameters.
	// Similar to the ListBuilder method in the groupRepo.
	return nil, nil
}

// CountX gets a count of roles.
func (r *roleRepo) CountX(ctx context.Context, params *structs.ListRoleParams) int {
	// Here you can implement the logic to count the number of roles based on the provided parameters.
	return 0
}
