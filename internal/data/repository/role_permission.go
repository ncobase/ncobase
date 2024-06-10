package repo

import (
	"context"
	"stocms/internal/data"
	"stocms/internal/data/ent"
	rolePermissionEnt "stocms/internal/data/ent/rolepermission"
	"stocms/internal/data/structs"
	"stocms/pkg/cache"
	"stocms/pkg/log"

	"github.com/redis/go-redis/v9"
)

// RolePermission represents the role permission repository interface.
type RolePermission interface {
	Create(ctx context.Context, body *structs.RolePermission) (*ent.RolePermission, error)
	GetByPermissionID(ctx context.Context, id string) (*ent.RolePermission, error)
	GetByRoleID(ctx context.Context, id string) (*ent.RolePermission, error)
	GetByPermissionIDs(ctx context.Context, ids []string) ([]*ent.RolePermission, error)
	GetByRoleIDs(ctx context.Context, ids []string) ([]*ent.RolePermission, error)
	DeleteByPermissionID(ctx context.Context, id string) error
	DeleteByRoleID(ctx context.Context, id string) error
	DeleteAllByPermissionID(ctx context.Context, id string) error
	DeleteAllByRoleID(ctx context.Context, id string) error
}

// rolePermissionRepo implements the Permission interface.
type rolePermissionRepo struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.RolePermission]
}

// NewRolePermission creates a new role permission repository.
func NewRolePermission(d *data.Data) RolePermission {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &rolePermissionRepo{ec, rc, cache.NewCache[ent.RolePermission](rc, cache.Key("sc_group_role"), true)}
}

// Create - Create role permission
func (r *rolePermissionRepo) Create(ctx context.Context, body *structs.RolePermission) (*ent.RolePermission, error) {

	// create builder.
	builder := r.ec.RolePermission.Create()
	// set values.
	builder.SetNillableID(&body.RoleID)
	builder.SetNillablePermissionID(&body.PermissionID)

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(nil, "rolePermissionRepo.Create error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByPermissionID - Find role permission by permission id
func (r *rolePermissionRepo) GetByPermissionID(ctx context.Context, id string) (*ent.RolePermission, error) {
	row, err := r.ec.RolePermission.
		Query().
		Where(rolePermissionEnt.IDEQ(id)).
		Only(ctx)

	if err != nil {
		log.Errorf(nil, "rolePermissionRepo.GetProfile error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// GetByPermissionIDs - Find role permissions by permission ids
func (r *rolePermissionRepo) GetByPermissionIDs(ctx context.Context, ids []string) ([]*ent.RolePermission, error) {
	rows, err := r.ec.RolePermission.
		Query().
		Where(rolePermissionEnt.IDIn(ids...)).
		All(ctx)

	if err != nil {
		log.Errorf(nil, "rolePermissionRepo.GetByPermissionIDs error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// GetByRoleID - Find role permission by role id
func (r *rolePermissionRepo) GetByRoleID(ctx context.Context, id string) (*ent.RolePermission, error) {
	row, err := r.ec.RolePermission.
		Query().
		Where(rolePermissionEnt.IDEQ(id)).
		Only(ctx)

	if err != nil {
		log.Errorf(nil, "rolePermissionRepo.GetProfile error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// GetByRoleIDs - Find role permissions by role ids
func (r *rolePermissionRepo) GetByRoleIDs(ctx context.Context, ids []string) ([]*ent.RolePermission, error) {
	rows, err := r.ec.RolePermission.
		Query().
		Where(rolePermissionEnt.IDIn(ids...)).
		All(ctx)

	if err != nil {
		log.Errorf(nil, "rolePermissionRepo.GetByRoleIDs error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// DeleteByPermissionID - Delete role permission
func (r *rolePermissionRepo) DeleteByPermissionID(ctx context.Context, id string) error {
	if _, err := r.ec.RolePermission.Delete().Where(rolePermissionEnt.IDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(nil, "rolePermissionRepo.DeleteByPermissionID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteByRoleID - Delete role permission
func (r *rolePermissionRepo) DeleteByRoleID(ctx context.Context, id string) error {
	if _, err := r.ec.RolePermission.Delete().Where(rolePermissionEnt.IDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(nil, "rolePermissionRepo.DeleteByRoleID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteAllByPermissionID - Delete all role permission
func (r *rolePermissionRepo) DeleteAllByPermissionID(ctx context.Context, id string) error {
	if _, err := r.ec.RolePermission.Delete().Where(rolePermissionEnt.IDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(nil, "rolePermissionRepo.DeleteAllByPermissionID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteAllByRoleID - Delete all role permission
func (r *rolePermissionRepo) DeleteAllByRoleID(ctx context.Context, id string) error {
	if _, err := r.ec.RolePermission.Delete().Where(rolePermissionEnt.IDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(nil, "rolePermissionRepo.DeleteAllByRoleID error: %v\n", err)
		return err
	}
	return nil
}
