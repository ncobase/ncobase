package repo

import (
	"context"
	"stocms/internal/data"
	"stocms/internal/data/ent"
	groupRoleEnt "stocms/internal/data/ent/grouprole"
	"stocms/internal/data/structs"
	"stocms/pkg/cache"
	"stocms/pkg/log"

	"github.com/redis/go-redis/v9"
)

// GroupRole represents the group role repository interface.
type GroupRole interface {
	Create(ctx context.Context, body *structs.GroupRole) (*ent.GroupRole, error)
	GetByGroupID(ctx context.Context, id string) (*ent.GroupRole, error)
	GetByRoleID(ctx context.Context, id string) (*ent.GroupRole, error)
	GetByGroupIDs(ctx context.Context, ids []string) ([]*ent.GroupRole, error)
	GetByRoleIDs(ctx context.Context, ids []string) ([]*ent.GroupRole, error)
	DeleteByGroupID(ctx context.Context, id string) error
	DeleteByRoleID(ctx context.Context, id string) error
	DeleteAllByGroupID(ctx context.Context, id string) error
	DeleteAllByRoleID(ctx context.Context, id string) error
}

// groupRoleRepo implements the Group interface.
type groupRoleRepo struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.GroupRole]
}

// NewGroupRole creates a new group role repository.
func NewGroupRole(d *data.Data) GroupRole {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &groupRoleRepo{ec, rc, cache.NewCache[ent.GroupRole](rc, cache.Key("sc_group_role"), true)}
}

// Create - Create group role
func (r *groupRoleRepo) Create(ctx context.Context, body *structs.GroupRole) (*ent.GroupRole, error) {

	// create builder.
	builder := r.ec.GroupRole.Create()
	// set values.
	builder.SetNillableID(&body.GroupID)
	builder.SetNillableRoleID(&body.RoleID)

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(nil, "groupRoleRepo.Create error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByGroupID - Find role by group id
func (r *groupRoleRepo) GetByGroupID(ctx context.Context, id string) (*ent.GroupRole, error) {
	row, err := r.ec.GroupRole.
		Query().
		Where(groupRoleEnt.IDEQ(id)).
		Only(ctx)

	if err != nil {
		log.Errorf(nil, "groupRoleRepo.GetProfile error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// GetByGroupIDs - Find roles by group ids
func (r *groupRoleRepo) GetByGroupIDs(ctx context.Context, ids []string) ([]*ent.GroupRole, error) {
	rows, err := r.ec.GroupRole.
		Query().
		Where(groupRoleEnt.IDIn(ids...)).
		All(ctx)

	if err != nil {
		log.Errorf(nil, "groupRoleRepo.GetByGroupIDs error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// GetByRoleID - Find role by role id
func (r *groupRoleRepo) GetByRoleID(ctx context.Context, id string) (*ent.GroupRole, error) {
	row, err := r.ec.GroupRole.
		Query().
		Where(groupRoleEnt.RoleIDEQ(id)).
		Only(ctx)

	if err != nil {
		log.Errorf(nil, "groupRoleRepo.GetProfile error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// GetByRoleIDs - Find roles by role ids
func (r *groupRoleRepo) GetByRoleIDs(ctx context.Context, ids []string) ([]*ent.GroupRole, error) {
	rows, err := r.ec.GroupRole.
		Query().
		Where(groupRoleEnt.RoleIDIn(ids...)).
		All(ctx)

	if err != nil {
		log.Errorf(nil, "groupRoleRepo.GetByRoleIDs error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// DeleteByGroupID - Delete group role
func (r *groupRoleRepo) DeleteByGroupID(ctx context.Context, id string) error {
	if _, err := r.ec.GroupRole.Delete().Where(groupRoleEnt.IDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(nil, "groupRoleRepo.DeleteByGroupID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteByRoleID - Delete group role
func (r *groupRoleRepo) DeleteByRoleID(ctx context.Context, id string) error {
	if _, err := r.ec.GroupRole.Delete().Where(groupRoleEnt.RoleIDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(nil, "groupRoleRepo.DeleteByRoleID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteAllByGroupID - Delete all group role
func (r *groupRoleRepo) DeleteAllByGroupID(ctx context.Context, id string) error {
	if _, err := r.ec.GroupRole.Delete().Where(groupRoleEnt.IDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(nil, "groupRoleRepo.DeleteAllByGroupID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteAllByRoleID - Delete all group role
func (r *groupRoleRepo) DeleteAllByRoleID(ctx context.Context, id string) error {
	if _, err := r.ec.GroupRole.Delete().Where(groupRoleEnt.RoleIDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(nil, "groupRoleRepo.DeleteAllByRoleID error: %v\n", err)
		return err
	}
	return nil
}
