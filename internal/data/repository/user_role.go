package repo

import (
	"context"
	"stocms/internal/data"
	"stocms/internal/data/ent"
	userRoleEnt "stocms/internal/data/ent/userrole"
	"stocms/internal/data/structs"
	"stocms/pkg/cache"
	"stocms/pkg/log"

	"github.com/redis/go-redis/v9"
)

// UserRole represents the user role repository interface.
type UserRole interface {
	Create(ctx context.Context, body *structs.UserRole) (*ent.UserRole, error)
	GetByUserID(ctx context.Context, id string) (*ent.UserRole, error)
	GetByRoleID(ctx context.Context, id string) (*ent.UserRole, error)
	GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserRole, error)
	GetByRoleIDs(ctx context.Context, ids []string) ([]*ent.UserRole, error)
	DeleteByUserID(ctx context.Context, id string) error
	DeleteByRoleID(ctx context.Context, id string) error
	DeleteAllByUserID(ctx context.Context, id string) error
	DeleteAllByRoleID(ctx context.Context, id string) error
}

// userRoleRepo implements the User interface.
type userRoleRepo struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.UserRole]
}

// NewUserRole creates a new user role repository.
func NewUserRole(d *data.Data) UserRole {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &userRoleRepo{ec, rc, cache.NewCache[ent.UserRole](rc, cache.Key("sc_user_role"), true)}
}

// Create - Create user role
func (r *userRoleRepo) Create(ctx context.Context, body *structs.UserRole) (*ent.UserRole, error) {

	// create builder.
	builder := r.ec.UserRole.Create()
	// set values.
	builder.SetNillableID(&body.UserID)
	builder.SetNillableRoleID(&body.RoleID)

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(nil, "userRoleRepo.Create error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByUserID - Find role by user id
func (r *userRoleRepo) GetByUserID(ctx context.Context, id string) (*ent.UserRole, error) {
	row, err := r.ec.UserRole.
		Query().
		Where(userRoleEnt.IDEQ(id)).
		Only(ctx)

	if err != nil {
		log.Errorf(nil, "userRoleRepo.GetProfile error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// GetByUserIDs - Find roles by user ids
func (r *userRoleRepo) GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserRole, error) {
	rows, err := r.ec.UserRole.
		Query().
		Where(userRoleEnt.IDIn(ids...)).
		All(ctx)

	if err != nil {
		log.Errorf(nil, "userRoleRepo.GetByUserIDs error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// GetByRoleID - Find role by role id
func (r *userRoleRepo) GetByRoleID(ctx context.Context, id string) (*ent.UserRole, error) {
	row, err := r.ec.UserRole.
		Query().
		Where(userRoleEnt.RoleIDEQ(id)).
		Only(ctx)

	if err != nil {
		log.Errorf(nil, "userRoleRepo.GetProfile error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// GetByRoleIDs - Find roles by role ids
func (r *userRoleRepo) GetByRoleIDs(ctx context.Context, ids []string) ([]*ent.UserRole, error) {
	rows, err := r.ec.UserRole.
		Query().
		Where(userRoleEnt.RoleIDIn(ids...)).
		All(ctx)

	if err != nil {
		log.Errorf(nil, "userRoleRepo.GetByRoleIDs error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// DeleteByUserID - Delete user role
func (r *userRoleRepo) DeleteByUserID(ctx context.Context, id string) error {
	if _, err := r.ec.UserRole.Delete().Where(userRoleEnt.IDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(nil, "userRoleRepo.DeleteByUserID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteByRoleID - Delete user role
func (r *userRoleRepo) DeleteByRoleID(ctx context.Context, id string) error {
	if _, err := r.ec.UserRole.Delete().Where(userRoleEnt.RoleIDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(nil, "userRoleRepo.DeleteByRoleID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteAllByUserID - Delete all user role
func (r *userRoleRepo) DeleteAllByUserID(ctx context.Context, id string) error {
	if _, err := r.ec.UserRole.Delete().Where(userRoleEnt.IDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(nil, "userRoleRepo.DeleteAllByUserID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteAllByRoleID - Delete all user role
func (r *userRoleRepo) DeleteAllByRoleID(ctx context.Context, id string) error {
	if _, err := r.ec.UserRole.Delete().Where(userRoleEnt.RoleIDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(nil, "userRoleRepo.DeleteAllByRoleID error: %v\n", err)
		return err
	}
	return nil
}
