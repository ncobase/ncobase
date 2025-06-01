package repository

import (
	"context"
	"fmt"
	"ncobase/access/data"
	"ncobase/access/data/ent"
	permissionEnt "ncobase/access/data/ent/permission"
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
	ec                 *ent.Client
	permissionCache    cache.ICache[ent.Permission]
	nameMappingCache   cache.ICache[string] // Maps name to permission ID
	actionSubjectCache cache.ICache[string] // Maps action:subject to permission ID
	permissionTTL      time.Duration
}

// NewPermissionRepository creates a new permission repository.
func NewPermissionRepository(d *data.Data) PermissionRepositoryInterface {
	redisClient := d.GetRedis()

	return &permissionRepository{
		ec:                 d.GetMasterEntClient(),
		permissionCache:    cache.NewCache[ent.Permission](redisClient, "ncse_access:permissions"),
		nameMappingCache:   cache.NewCache[string](redisClient, "ncse_access:permission_names"),
		actionSubjectCache: cache.NewCache[string](redisClient, "ncse_access:permission_actions"),
		permissionTTL:      time.Hour * 4, // 4 hours cache TTL (permissions change less frequently)
	}
}

// Create creates a new permission
func (r *permissionRepository) Create(ctx context.Context, body *structs.CreatePermissionBody) (*ent.Permission, error) {
	builder := r.ec.Permission.Create()
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

	permission, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "permissionRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the permission
	go r.cachePermission(context.Background(), permission)

	return permission, nil
}

// GetByName gets a permission by name
func (r *permissionRepository) GetByName(ctx context.Context, name string) (*ent.Permission, error) {
	// Try to get permission ID from name mapping cache
	if permissionID, err := r.getPermissionIDByName(ctx, name); err == nil && permissionID != "" {
		return r.GetByID(ctx, permissionID)
	}

	// Fallback to database
	row, err := r.FindPermission(ctx, &structs.FindPermission{Name: name})
	if err != nil {
		logger.Errorf(ctx, "permissionRepo.GetByName error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cachePermission(context.Background(), row)

	return row, nil
}

// GetByID gets a permission by ID
func (r *permissionRepository) GetByID(ctx context.Context, id string) (*ent.Permission, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("id:%s", id)
	if cached, err := r.permissionCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Fallback to database
	row, err := r.FindPermission(ctx, &structs.FindPermission{ID: id})
	if err != nil {
		logger.Errorf(ctx, "permissionRepo.GetByID error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cachePermission(context.Background(), row)

	return row, nil
}

// GetByActionAndSubject gets a permission by action and subject
func (r *permissionRepository) GetByActionAndSubject(ctx context.Context, action, subject string) (*ent.Permission, error) {
	// Try to get permission ID from action:subject mapping cache
	if permissionID, err := r.getPermissionIDByActionSubject(ctx, action, subject); err == nil && permissionID != "" {
		return r.GetByID(ctx, permissionID)
	}

	// Fallback to database
	row, err := r.FindPermission(ctx, &structs.FindPermission{Action: action, Subject: subject})
	if err != nil {
		logger.Errorf(ctx, "permissionRepo.GetByActionAndSubject error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cachePermission(context.Background(), row)

	return row, nil
}

// Update updates a permission
func (r *permissionRepository) Update(ctx context.Context, id string, updates types.JSON) (*ent.Permission, error) {
	permission, err := r.FindPermission(ctx, &structs.FindPermission{ID: id})
	if err != nil {
		return nil, err
	}

	builder := permission.Update()

	for field, value := range updates {
		switch field {
		case "name":
			builder.SetNillableName(convert.ToPointer(value.(string)))
		case "action":
			builder.SetNillableAction(convert.ToPointer(value.(string)))
		case "subject":
			builder.SetNillableSubject(convert.ToPointer(value.(string)))
		case "description":
			builder.SetNillableDescription(convert.ToPointer(value.(string)))
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

	updatedPermission, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "permissionRepo.Update error: %v", err)
		return nil, err
	}

	// Invalidate and re-cache
	go func() {
		r.invalidatePermissionCache(context.Background(), permission)
		r.cachePermission(context.Background(), updatedPermission)
	}()

	return updatedPermission, nil
}

// Delete deletes a permission
func (r *permissionRepository) Delete(ctx context.Context, id string) error {
	permission, err := r.FindPermission(ctx, &structs.FindPermission{ID: id})
	if err != nil {
		return err
	}

	builder := r.ec.Permission.Delete()
	if _, err = builder.Where(permissionEnt.IDEQ(permission.ID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "permissionRepo.Delete error: %v", err)
		return err
	}

	// Invalidate cache
	go r.invalidatePermissionCache(context.Background(), permission)

	return nil
}

// FindPermission finds a permission
func (r *permissionRepository) FindPermission(ctx context.Context, params *structs.FindPermission) (*ent.Permission, error) {
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

	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// List gets a list of permissions
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

	// Cache permissions in background
	go func() {
		for _, permission := range rows {
			r.cachePermission(context.Background(), permission)
		}
	}()

	return rows, nil
}

// CountX gets a count of permissions
func (r *permissionRepository) CountX(ctx context.Context, params *structs.ListPermissionParams) int {
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// listBuilder creates list builder
func (r *permissionRepository) listBuilder(ctx context.Context, params *structs.ListPermissionParams) (*ent.PermissionQuery, error) {
	return r.ec.Permission.Query(), nil
}

// cachePermission caches a permission
func (r *permissionRepository) cachePermission(ctx context.Context, permission *ent.Permission) {
	// Cache by ID
	idKey := fmt.Sprintf("id:%s", permission.ID)
	if err := r.permissionCache.Set(ctx, idKey, permission, r.permissionTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache permission by ID %s: %v", permission.ID, err)
	}

	// Cache name to ID mapping
	if permission.Name != "" {
		nameKey := fmt.Sprintf("name:%s", permission.Name)
		if err := r.nameMappingCache.Set(ctx, nameKey, &permission.ID, r.permissionTTL); err != nil {
			logger.Debugf(ctx, "Failed to cache name mapping %s: %v", permission.Name, err)
		}
	}

	// Cache action:subject to ID mapping
	if permission.Action != "" && permission.Subject != "" {
		actionSubjectKey := fmt.Sprintf("action:%s:subject:%s", permission.Action, permission.Subject)
		if err := r.actionSubjectCache.Set(ctx, actionSubjectKey, &permission.ID, r.permissionTTL); err != nil {
			logger.Debugf(ctx, "Failed to cache action:subject mapping %s:%s: %v", permission.Action, permission.Subject, err)
		}
	}
}

// invalidatePermissionCache invalidates permission cache
func (r *permissionRepository) invalidatePermissionCache(ctx context.Context, permission *ent.Permission) {
	// Invalidate ID cache
	idKey := fmt.Sprintf("id:%s", permission.ID)
	if err := r.permissionCache.Delete(ctx, idKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate permission ID cache %s: %v", permission.ID, err)
	}

	// Invalidate name mapping
	if permission.Name != "" {
		nameKey := fmt.Sprintf("name:%s", permission.Name)
		if err := r.nameMappingCache.Delete(ctx, nameKey); err != nil {
			logger.Debugf(ctx, "Failed to invalidate name mapping cache %s: %v", permission.Name, err)
		}
	}

	// Invalidate action:subject mapping
	if permission.Action != "" && permission.Subject != "" {
		actionSubjectKey := fmt.Sprintf("action:%s:subject:%s", permission.Action, permission.Subject)
		if err := r.actionSubjectCache.Delete(ctx, actionSubjectKey); err != nil {
			logger.Debugf(ctx, "Failed to invalidate action:subject mapping cache %s:%s: %v", permission.Action, permission.Subject, err)
		}
	}
}

// getPermissionIDByName gets permission ID by name
func (r *permissionRepository) getPermissionIDByName(ctx context.Context, name string) (string, error) {
	cacheKey := fmt.Sprintf("name:%s", name)
	permissionID, err := r.nameMappingCache.Get(ctx, cacheKey)
	if err != nil || permissionID == nil {
		return "", err
	}
	return *permissionID, nil
}

// getPermissionIDByActionSubject gets permission ID by action and subject
func (r *permissionRepository) getPermissionIDByActionSubject(ctx context.Context, action, subject string) (string, error) {
	cacheKey := fmt.Sprintf("action:%s:subject:%s", action, subject)
	permissionID, err := r.actionSubjectCache.Get(ctx, cacheKey)
	if err != nil || permissionID == nil {
		return "", err
	}
	return *permissionID, nil
}
