package repo

import (
	"context"
	"stocms/internal/data"
	"stocms/internal/data/ent"
	userGroupEnt "stocms/internal/data/ent/usergroup"
	"stocms/internal/data/structs"
	"stocms/pkg/cache"
	"stocms/pkg/log"

	"github.com/redis/go-redis/v9"
)

// UserGroup represents the user group repository interface.
type UserGroup interface {
	Create(ctx context.Context, body *structs.UserGroup) (*ent.UserGroup, error)
	GetByUserID(ctx context.Context, id string) (*ent.UserGroup, error)
	GetByGroupID(ctx context.Context, id string) (*ent.UserGroup, error)
	GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserGroup, error)
	GetByGroupIDs(ctx context.Context, ids []string) ([]*ent.UserGroup, error)
	DeleteByUserID(ctx context.Context, id string) error
	DeleteByGroupID(ctx context.Context, id string) error
	DeleteAllByUserID(ctx context.Context, id string) error
	DeleteAllByGroupID(ctx context.Context, id string) error
}

// userGroupRepo implements the User interface.
type userGroupRepo struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.UserGroup]
}

// NewUserGroup creates a new user group repository.
func NewUserGroup(d *data.Data) UserGroup {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &userGroupRepo{ec, rc, cache.NewCache[ent.UserGroup](rc, cache.Key("sc_user_group"), true)}
}

// Create - Create user group
func (r *userGroupRepo) Create(ctx context.Context, body *structs.UserGroup) (*ent.UserGroup, error) {

	// create builder.
	builder := r.ec.UserGroup.Create()
	// set values.
	builder.SetNillableID(&body.UserID)
	builder.SetNillableGroupID(&body.GroupID)

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(nil, "userGroupRepo.Create error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByUserID - Find group by user id
func (r *userGroupRepo) GetByUserID(ctx context.Context, id string) (*ent.UserGroup, error) {
	row, err := r.ec.UserGroup.
		Query().
		Where(userGroupEnt.IDEQ(id)).
		Only(ctx)

	if err != nil {
		log.Errorf(nil, "userGroupRepo.GetProfile error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// GetByUserIDs - Find groups by user ids
func (r *userGroupRepo) GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserGroup, error) {
	rows, err := r.ec.UserGroup.
		Query().
		Where(userGroupEnt.IDIn(ids...)).
		All(ctx)

	if err != nil {
		log.Errorf(nil, "userGroupRepo.GetByUserIDs error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// GetByGroupID - Find group by group id
func (r *userGroupRepo) GetByGroupID(ctx context.Context, id string) (*ent.UserGroup, error) {
	row, err := r.ec.UserGroup.
		Query().
		Where(userGroupEnt.GroupIDEQ(id)).
		Only(ctx)

	if err != nil {
		log.Errorf(nil, "userGroupRepo.GetProfile error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// GetByGroupIDs - Find groups by group ids
func (r *userGroupRepo) GetByGroupIDs(ctx context.Context, ids []string) ([]*ent.UserGroup, error) {
	rows, err := r.ec.UserGroup.
		Query().
		Where(userGroupEnt.GroupIDIn(ids...)).
		All(ctx)

	if err != nil {
		log.Errorf(nil, "userGroupRepo.GetByGroupIDs error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// DeleteByUserID - Delete user group
func (r *userGroupRepo) DeleteByUserID(ctx context.Context, id string) error {
	if _, err := r.ec.UserGroup.Delete().Where(userGroupEnt.IDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(nil, "userGroupRepo.DeleteByUserID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteByGroupID - Delete user group
func (r *userGroupRepo) DeleteByGroupID(ctx context.Context, id string) error {
	if _, err := r.ec.UserGroup.Delete().Where(userGroupEnt.GroupIDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(nil, "userGroupRepo.DeleteByGroupID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteAllByUserID - Delete all user group
func (r *userGroupRepo) DeleteAllByUserID(ctx context.Context, id string) error {
	if _, err := r.ec.UserGroup.Delete().Where(userGroupEnt.IDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(nil, "userGroupRepo.DeleteAllByUserID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteAllByGroupID - Delete all user group
func (r *userGroupRepo) DeleteAllByGroupID(ctx context.Context, id string) error {
	if _, err := r.ec.UserGroup.Delete().Where(userGroupEnt.GroupIDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(nil, "userGroupRepo.DeleteAllByGroupID error: %v\n", err)
		return err
	}
	return nil
}
