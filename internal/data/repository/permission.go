package repo

import (
	"context"
	"fmt"
	"stocms/internal/data"
	"stocms/internal/data/ent"
	permissionEnt "stocms/internal/data/ent/permission"
	"stocms/internal/data/structs"
	"stocms/pkg/cache"
	"stocms/pkg/log"
	"stocms/pkg/types"
	"stocms/pkg/validator"

	"github.com/redis/go-redis/v9"
)

// Permission represents the permission repository interface.
type Permission interface {
	Create(ctx context.Context, body *structs.CreatePermissionBody) (*ent.Permission, error)
	GetByID(ctx context.Context, id string) (*ent.Permission, error)
	GetByActionAndSubject(ctx context.Context, action, subject string) (*ent.Permission, error)
	Update(ctx context.Context, id string, updates types.JSON) (*ent.Permission, error)
	List(ctx context.Context, params *structs.ListPermissionParams) ([]*ent.Permission, error)
	Delete(ctx context.Context, id string) error
	FindPermission(ctx context.Context, p *structs.FindPermission) (*ent.Permission, error)
	ListBuilder(ctx context.Context, p *structs.ListPermissionParams) (*ent.PermissionQuery, error)
	CountX(ctx context.Context, p *structs.ListPermissionParams) int
}

// permissionRepo implements the Permission interface.
type permissionRepo struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.Permission]
}

// NewPermission creates a new permission repository.
func NewPermission(d *data.Data) Permission {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &permissionRepo{ec, rc, cache.NewCache[ent.Permission](rc, cache.Key("sc_permission"), true)}
}

// Create creates a new permission.
func (r *permissionRepo) Create(ctx context.Context, body *structs.CreatePermissionBody) (*ent.Permission, error) {

	// create builder.
	builder := r.ec.Permission.Create()
	// set values.
	builder.SetNillableName(&body.Name)
	builder.SetNillableAction(&body.Action)
	builder.SetNillableSubject(&body.Subject)
	builder.SetNillableDescription(&body.Description)
	builder.SetDefault(body.Default)
	builder.SetDisabled(body.Disabled)
	builder.SetExtras(body.ExtraProps)
	builder.SetNillableCreatedBy(&body.CreatedBy)

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(nil, "permissionRepo.Create error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByID gets a permission by ID.
func (r *permissionRepo) GetByID(ctx context.Context, id string) (*ent.Permission, error) {
	// check cache
	cacheKey := fmt.Sprintf("%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindPermission(ctx, &structs.FindPermission{ID: id})
	if err != nil {
		log.Errorf(nil, "permissionRepo.GetByID error: %v\n", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(nil, "permissionRepo.GetByID cache error: %v\n", err)
	}

	return row, nil
}

// GetByActionAndSubject gets a permission by action and subject.
func (r *permissionRepo) GetByActionAndSubject(ctx context.Context, action, subject string) (*ent.Permission, error) {
	// check cache
	cacheKey := fmt.Sprintf("%s_%s", action, subject)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindPermission(ctx, &structs.FindPermission{Action: action, Subject: subject})
	if err != nil {
		log.Errorf(nil, "permissionRepo.GetByActionAndSubject error: %v\n", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(nil, "permissionRepo.GetByActionAndSubject cache error: %v\n", err)
	}

	return row, nil
}

// Update updates a permission (full or partial).
func (r *permissionRepo) Update(ctx context.Context, id string, updates types.JSON) (*ent.Permission, error) {
	permission, err := r.FindPermission(ctx, &structs.FindPermission{ID: id})
	if err != nil {
		return nil, err
	}

	builder := permission.Update()

	for field, value := range updates {
		switch field {
		case "name":
			builder.SetNillableName(types.ToPointer(value.(string)))
		case "action":
			builder.SetNillableAction(types.ToPointer(value.(string)))
		case "subject":
			builder.SetNillableSubject(types.ToPointer(value.(string)))
		case "description":
			builder.SetNillableDescription(types.ToPointer(value.(string)))
		case "default":
			builder.SetDefault(value.(bool))
		case "disabled":
			builder.SetDisabled(value.(bool))
		case "extra_props":
			builder.SetExtras(value.(types.JSON))
		case "updated_by":
			builder.SetUpdatedBy(value.(string))
		}
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(nil, "permissionRepo.Update error: %v\n", err)
		return nil, err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", permission.ID)
	err = r.c.Delete(ctx, cacheKey)
	if err != nil {
		log.Errorf(nil, "permissionRepo.Update cache error: %v\n", err)
	}

	return row, nil
}

// List gets a list of permissions.
func (r *permissionRepo) List(ctx context.Context, p *structs.ListPermissionParams) ([]*ent.Permission, error) {
	// create list builder
	builder, err := r.ListBuilder(ctx, p)
	if validator.IsNotNil(err) {
		return nil, err
	}

	rows, err := builder.All(ctx)
	if err != nil {
		log.Errorf(nil, "permissionRepo.List error: %v\n", err)
		return nil, err
	}

	return rows, nil
}

// Delete deletes a permission.
func (r *permissionRepo) Delete(ctx context.Context, id string) error {
	permission, err := r.FindPermission(ctx, &structs.FindPermission{ID: id})
	if err != nil {
		return err
	}

	// create builder.
	builder := r.ec.Permission.Delete()

	// execute the builder and verify the result.
	if _, err = builder.Where(permissionEnt.IDEQ(permission.ID)).Exec(ctx); err != nil {
		log.Errorf(nil, "permissionRepo.Delete error: %v\n", err)
		return err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", permission.ID)
	err = r.c.Delete(ctx, cacheKey)
	if err != nil {
		log.Errorf(nil, "permissionRepo.Delete cache error: %v\n", err)
	}

	return nil
}

// FindPermission finds a permission.
func (r *permissionRepo) FindPermission(ctx context.Context, p *structs.FindPermission) (*ent.Permission, error) {

	// create builder.
	builder := r.ec.Permission.Query()

	if validator.IsNotEmpty(p.ID) {
		builder = builder.Where(permissionEnt.IDEQ(p.ID))
	}
	if validator.IsNotEmpty(p.Action) && validator.IsNotEmpty(p.Subject) {
		builder = builder.Where(permissionEnt.And(
			permissionEnt.ActionEQ(p.Action),
			permissionEnt.SubjectEQ(p.Subject),
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
func (r *permissionRepo) ListBuilder(ctx context.Context, p *structs.ListPermissionParams) (*ent.PermissionQuery, error) {
	// Here you can construct and return a builder for listing permissions based on the provided parameters.
	return nil, nil
}

// CountX gets a count of permissions.
func (r *permissionRepo) CountX(ctx context.Context, p *structs.ListPermissionParams) int {
	// Here you can implement the logic to count the number of permissions based on the provided parameters.
	return 0
}
