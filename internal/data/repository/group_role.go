package repo

import (
	"context"
	"ncobase/internal/data"
	"ncobase/internal/data/ent"
	groupEnt "ncobase/internal/data/ent/group"
	groupRoleEnt "ncobase/internal/data/ent/grouprole"
	roleEnt "ncobase/internal/data/ent/role"
	"ncobase/internal/data/structs"
	"ncobase/pkg/cache"
	"ncobase/pkg/log"

	"github.com/redis/go-redis/v9"
)

// GroupRole represents the group role repository interface.
type GroupRole interface {
	Create(ctx context.Context, body *structs.GroupRole) (*ent.GroupRole, error)
	GetByGroupID(ctx context.Context, id string) (*ent.GroupRole, error)
	GetByRoleID(ctx context.Context, id string) (*ent.GroupRole, error)
	GetByGroupIDs(ctx context.Context, ids []string) ([]*ent.GroupRole, error)
	GetByRoleIDs(ctx context.Context, ids []string) ([]*ent.GroupRole, error)
	Delete(ctx context.Context, gid, rid string) error
	DeleteAllByGroupID(ctx context.Context, id string) error
	DeleteAllByRoleID(ctx context.Context, id string) error
	GetRolesByGroupID(ctx context.Context, groupID string) ([]*ent.Role, error)
	GetGroupsByRoleID(ctx context.Context, roleID string) ([]*ent.Group, error)
	IsRoleInGroup(ctx context.Context, groupID string, roleID string) (bool, error)
	IsGroupInRole(ctx context.Context, roleID string, groupID string) (bool, error)
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
	return &groupRoleRepo{ec, rc, cache.NewCache[ent.GroupRole](rc, cache.Key("nb_group_role"), true)}
}

// Create group role
func (r *groupRoleRepo) Create(ctx context.Context, body *structs.GroupRole) (*ent.GroupRole, error) {

	// create builder.
	builder := r.ec.GroupRole.Create()
	// set values.
	builder.SetNillableID(&body.GroupID)
	builder.SetNillableRoleID(&body.RoleID)

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(context.Background(), "groupRoleRepo.Create error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByGroupID Find role by group id
func (r *groupRoleRepo) GetByGroupID(ctx context.Context, id string) (*ent.GroupRole, error) {
	row, err := r.ec.GroupRole.
		Query().
		Where(groupRoleEnt.IDEQ(id)).
		Only(ctx)

	if err != nil {
		log.Errorf(context.Background(), "groupRoleRepo.GetProfile error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// GetByGroupIDs Find roles by group ids
func (r *groupRoleRepo) GetByGroupIDs(ctx context.Context, ids []string) ([]*ent.GroupRole, error) {
	rows, err := r.ec.GroupRole.
		Query().
		Where(groupRoleEnt.IDIn(ids...)).
		All(ctx)

	if err != nil {
		log.Errorf(context.Background(), "groupRoleRepo.GetByGroupIDs error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// GetByRoleID Find role by role id
func (r *groupRoleRepo) GetByRoleID(ctx context.Context, id string) (*ent.GroupRole, error) {
	row, err := r.ec.GroupRole.
		Query().
		Where(groupRoleEnt.RoleIDEQ(id)).
		Only(ctx)

	if err != nil {
		log.Errorf(context.Background(), "groupRoleRepo.GetProfile error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// GetByRoleIDs Find roles by role ids
func (r *groupRoleRepo) GetByRoleIDs(ctx context.Context, ids []string) ([]*ent.GroupRole, error) {
	rows, err := r.ec.GroupRole.
		Query().
		Where(groupRoleEnt.RoleIDIn(ids...)).
		All(ctx)

	if err != nil {
		log.Errorf(context.Background(), "groupRoleRepo.GetByRoleIDs error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// Delete group role
func (r *groupRoleRepo) Delete(ctx context.Context, gid, rid string) error {
	if _, err := r.ec.GroupRole.Delete().Where(groupRoleEnt.IDEQ(gid), groupRoleEnt.RoleIDEQ(rid)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "groupRoleRepo.Delete error: %v\n", err)
		return err
	}
	return nil
}

// DeleteAllByGroupID Delete all group role
func (r *groupRoleRepo) DeleteAllByGroupID(ctx context.Context, id string) error {
	if _, err := r.ec.GroupRole.Delete().Where(groupRoleEnt.IDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "groupRoleRepo.DeleteAllByGroupID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteAllByRoleID Delete all group role
func (r *groupRoleRepo) DeleteAllByRoleID(ctx context.Context, id string) error {
	if _, err := r.ec.GroupRole.Delete().Where(groupRoleEnt.RoleIDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "groupRoleRepo.DeleteAllByRoleID error: %v\n", err)
		return err
	}
	return nil
}

// GetRolesByGroupID retrieves all roles under a group.
func (r *groupRoleRepo) GetRolesByGroupID(ctx context.Context, groupID string) ([]*ent.Role, error) {
	groupRoles, err := r.ec.GroupRole.Query().Where(groupRoleEnt.IDEQ(groupID)).All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "groupRoleRepo.GetRolesByGroupID error: %v\n", err)
		return nil, err
	}

	var roleIDs []string
	for _, groupRole := range groupRoles {
		roleIDs = append(roleIDs, groupRole.RoleID)
	}

	roles, err := r.ec.Role.Query().Where(roleEnt.IDIn(roleIDs...)).All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "groupRoleRepo.GetRolesByGroupID error: %v\n", err)
		return nil, err
	}

	return roles, nil
}

// GetGroupsByRoleID retrieves all groups under a role.
func (r *groupRoleRepo) GetGroupsByRoleID(ctx context.Context, roleID string) ([]*ent.Group, error) {
	groupRoles, err := r.ec.GroupRole.Query().Where(groupRoleEnt.RoleIDEQ(roleID)).All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "groupRoleRepo.GetGroupsByRoleID error: %v\n", err)
		return nil, err
	}

	var groupIDs []string
	for _, groupRole := range groupRoles {
		groupIDs = append(groupIDs, groupRole.ID)
	}

	groups, err := r.ec.Group.Query().Where(groupEnt.IDIn(groupIDs...)).All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "groupRoleRepo.GetGroupsByRoleID error: %v\n", err)
		return nil, err
	}

	return groups, nil
}

// IsRoleInGroup verifies if a role belongs to a specific group.
func (r *groupRoleRepo) IsRoleInGroup(ctx context.Context, groupID string, roleID string) (bool, error) {
	count, err := r.ec.GroupRole.Query().Where(groupRoleEnt.IDEQ(groupID), groupRoleEnt.RoleIDEQ(roleID)).Count(ctx)
	if err != nil {
		log.Errorf(context.Background(), "groupRoleRepo.IsRoleInGroup error: %v\n", err)
		return false, err
	}
	return count > 0, nil
}

// IsGroupInRole verifies if a group belongs to a specific role.
func (r *groupRoleRepo) IsGroupInRole(ctx context.Context, groupID string, roleID string) (bool, error) {
	count, err := r.ec.GroupRole.Query().Where(groupRoleEnt.IDEQ(groupID), groupRoleEnt.RoleIDEQ(roleID)).Count(ctx)
	if err != nil {
		log.Errorf(context.Background(), "groupRoleRepo.IsGroupInRole error: %v\n", err)
		return false, err
	}
	return count > 0, nil
}
