package repository

import (
	"context"
	"fmt"
	"ncobase/feature/access/data"
	"ncobase/feature/access/data/ent"
	permissionEnt "ncobase/feature/access/data/ent/permission"
	"ncobase/feature/access/structs"

	"ncobase/common/cache"
	"ncobase/common/log"
	"ncobase/common/types"
	"ncobase/common/validator"

	"github.com/redis/go-redis/v9"
)

// PermissionRepositoryInterface represents the permission repository interface.
type PermissionRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreatePermissionBody) (*ent.Permission, error)
	GetByID(ctx context.Context, id string) (*ent.Permission, error)
	GetByActionAndSubject(ctx context.Context, action, subject string) (*ent.Permission, error)
	Update(ctx context.Context, id string, updates types.JSON) (*ent.Permission, error)
	List(ctx context.Context, params *structs.ListPermissionParams) ([]*ent.Permission, error)
	Delete(ctx context.Context, id string) error
	FindPermission(ctx context.Context, params *structs.FindPermission) (*ent.Permission, error)
	CountX(ctx context.Context, params *structs.ListPermissionParams) int
}

// permissionRepository implements the PermissionRepositoryInterface.
type permissionRepository struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.Permission]
}

// NewPermissionRepository creates a new permission repository.
func NewPermissionRepository(d *data.Data) PermissionRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &permissionRepository{ec, rc, cache.NewCache[ent.Permission](rc, "nb_permission")}
}

// Create creates a new permission.
func (r *permissionRepository) Create(ctx context.Context, body *structs.CreatePermissionBody) (*ent.Permission, error) {

	// create builder.
	builder := r.ec.Permission.Create()
	// set values.
	builder.SetNillableName(&body.Name)
	builder.SetNillableAction(&body.Action)
	builder.SetNillableSubject(&body.Subject)
	builder.SetNillableDescription(&body.Description)
	builder.SetNillableDefault(body.Default)
	builder.SetNillableDisabled(body.Disabled)
	builder.SetNillableCreatedBy(body.CreatedBy)

	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(context.Background(), "permissionRepo.Create error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByID gets a permission by ID.
func (r *permissionRepository) GetByID(ctx context.Context, id string) (*ent.Permission, error) {
	// check cache
	cacheKey := fmt.Sprintf("%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindPermission(ctx, &structs.FindPermission{ID: id})
	if err != nil {
		log.Errorf(context.Background(), "permissionRepo.GetByID error: %v\n", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(context.Background(), "permissionRepo.GetByID cache error: %v\n", err)
	}

	return row, nil
}

// GetByActionAndSubject gets a permission by action and subject.
func (r *permissionRepository) GetByActionAndSubject(ctx context.Context, action, subject string) (*ent.Permission, error) {
	// check cache
	cacheKey := fmt.Sprintf("%s_%s", action, subject)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindPermission(ctx, &structs.FindPermission{Action: action, Subject: subject})
	if err != nil {
		log.Errorf(context.Background(), "permissionRepo.GetByActionAndSubject error: %v\n", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(context.Background(), "permissionRepo.GetByActionAndSubject cache error: %v\n", err)
	}

	return row, nil
}

// Update updates a permission (full or partial).
func (r *permissionRepository) Update(ctx context.Context, id string, updates types.JSON) (*ent.Permission, error) {
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
		log.Errorf(context.Background(), "permissionRepo.Update error: %v\n", err)
		return nil, err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", permission.ID)
	err = r.c.Delete(ctx, cacheKey)
	if err != nil {
		log.Errorf(context.Background(), "permissionRepo.Update cache error: %v\n", err)
	}

	return row, nil
}

// List gets a list of permissions.
func (r *permissionRepository) List(ctx context.Context, params *structs.ListPermissionParams) ([]*ent.Permission, error) {
	// create list builder
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return nil, err
	}

	// limit the result
	builder.Limit(int(params.Limit))

	rows, err := builder.All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "permissionRepo.List error: %v\n", err)
		return nil, err
	}

	return rows, nil
}

// Delete deletes a permission.
func (r *permissionRepository) Delete(ctx context.Context, id string) error {
	permission, err := r.FindPermission(ctx, &structs.FindPermission{ID: id})
	if err != nil {
		return err
	}

	// create builder.
	builder := r.ec.Permission.Delete()

	// execute the builder and verify the result.
	if _, err = builder.Where(permissionEnt.IDEQ(permission.ID)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "permissionRepo.Delete error: %v\n", err)
		return err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", permission.ID)
	err = r.c.Delete(ctx, cacheKey)
	if err != nil {
		log.Errorf(context.Background(), "permissionRepo.Delete cache error: %v\n", err)
	}

	return nil
}

// FindPermission finds a permission.
func (r *permissionRepository) FindPermission(ctx context.Context, params *structs.FindPermission) (*ent.Permission, error) {

	// create builder.
	builder := r.ec.Permission.Query()

	if validator.IsNotEmpty(params.ID) {
		builder = builder.Where(permissionEnt.IDEQ(params.ID))
	}
	if validator.IsNotEmpty(params.Action) && validator.IsNotEmpty(params.Subject) {
		builder = builder.Where(permissionEnt.And(
			permissionEnt.ActionEQ(params.Action),
			permissionEnt.SubjectEQ(params.Subject),
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
func (r *permissionRepository) listBuilder(ctx context.Context, params *structs.ListPermissionParams) (*ent.PermissionQuery, error) {
	// verify query params.
	var next *ent.Permission
	if validator.IsNotEmpty(params.Cursor) {
		// query the menu.
		row, err := r.GetByID(ctx, params.Cursor)
		if validator.IsNotNil(err) || validator.IsNil(row) {
			return nil, err
		}
		next = row
	}

	// create builder.
	builder := r.ec.Permission.Query()

	// lt the cursor create time
	if next != nil {
		builder.Where(permissionEnt.CreatedAtLT(next.CreatedAt))
	}

	return builder, nil
}

// CountX gets a count of permissions.
func (r *permissionRepository) CountX(ctx context.Context, params *structs.ListPermissionParams) int {
	// create list builder
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}
