package repository

import (
	"context"
	"fmt"
	"ncobase/core/access/data"
	"ncobase/core/access/data/ent"
	permissionEnt "ncobase/core/access/data/ent/permission"
	"ncobase/core/access/structs"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/redis/go-redis/v9"
)

// PermissionRepositoryInterface represents the permission repository interface.
type PermissionRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreatePermissionBody) (*ent.Permission, error)
	GetByName(ctx context.Context, name string) (*ent.Permission, error)
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
	return &permissionRepository{ec, rc, cache.NewCache[ent.Permission](rc, "ncse_permission")}
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
		logger.Errorf(ctx, "permissionRepo.Create error: %v", err)
		return nil, err
	}

	return row, nil
}

// GetByName gets a permission by name.
func (r *permissionRepository) GetByName(ctx context.Context, name string) (*ent.Permission, error) {
	// check cache
	cacheKey := fmt.Sprintf("%s", name)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindPermission(ctx, &structs.FindPermission{Name: name})
	if err != nil {
		logger.Errorf(ctx, "permissionRepo.GetByName error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "permissionRepo.GetByName cache error: %v", err)
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
		logger.Errorf(ctx, "permissionRepo.GetByID error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "permissionRepo.GetByID cache error: %v", err)
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
		logger.Errorf(ctx, "permissionRepo.GetByActionAndSubject error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "permissionRepo.GetByActionAndSubject cache error: %v", err)
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
		logger.Errorf(ctx, "permissionRepo.Update error: %v", err)
		return nil, err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", permission.ID)
	err = r.c.Delete(ctx, cacheKey)
	if err != nil {
		logger.Errorf(ctx, "permissionRepo.Update cache error: %v", err)
	}

	return row, nil
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
		logger.Errorf(ctx, "permissionRepo.Delete error: %v", err)
		return err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", permission.ID)
	err = r.c.Delete(ctx, cacheKey)
	if err != nil {
		logger.Errorf(ctx, "permissionRepo.Delete cache error: %v", err)
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
	if validator.IsNotEmpty(params.Name) {
		builder = builder.Where(permissionEnt.NameEQ(params.Name))
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

// List gets a list of permissions.
func (r *permissionRepository) List(ctx context.Context, params *structs.ListPermissionParams) ([]*ent.Permission, error) {
	builder := r.ec.Permission.Query()

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
				permissionEnt.Or(
					permissionEnt.CreatedAtGT(timestamp),
					permissionEnt.And(
						permissionEnt.CreatedAtEQ(timestamp),
						permissionEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				permissionEnt.Or(
					permissionEnt.CreatedAtLT(timestamp),
					permissionEnt.And(
						permissionEnt.CreatedAtEQ(timestamp),
						permissionEnt.IDLT(id),
					),
				),
			)
		}
	}

	if params.Direction == "backward" {
		builder.Order(ent.Asc(permissionEnt.FieldCreatedAt), ent.Asc(permissionEnt.FieldID))
	} else {
		builder.Order(ent.Desc(permissionEnt.FieldCreatedAt), ent.Desc(permissionEnt.FieldID))
	}

	builder.Limit(params.Limit)

	rows, err := builder.All(ctx)
	if err != nil {
		return nil, err
	}

	return rows, nil
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

// listBuilder creates list builder.
func (r *permissionRepository) listBuilder(ctx context.Context, params *structs.ListPermissionParams) (*ent.PermissionQuery, error) {
	// create builder.
	builder := r.ec.Permission.Query()

	return builder, nil
}
