package repository

import (
	"context"
	"ncobase/common/cache"
	"ncobase/common/log"
	"ncobase/feature/group/data"
	"ncobase/feature/group/data/ent"
	groupEnt "ncobase/feature/group/data/ent/group"
	groupRoleEnt "ncobase/feature/group/data/ent/grouprole"
	"ncobase/feature/group/structs"

	"github.com/redis/go-redis/v9"
)

// GroupRoleRepositoryInterface represents the group role repository interface.
type GroupRoleRepositoryInterface interface {
	Create(ctx context.Context, body *structs.GroupRole) (*ent.GroupRole, error)
	GetByGroupID(ctx context.Context, id string) (*ent.GroupRole, error)
	GetByRoleID(ctx context.Context, id string) (*ent.GroupRole, error)
	GetByGroupIDs(ctx context.Context, ids []string) ([]*ent.GroupRole, error)
	GetByRoleIDs(ctx context.Context, ids []string) ([]*ent.GroupRole, error)
	Delete(ctx context.Context, gid, rid string) error
	DeleteAllByGroupID(ctx context.Context, id string) error
	DeleteAllByRoleID(ctx context.Context, id string) error
	GetRolesByGroupID(ctx context.Context, groupID string) ([]string, error)
	GetGroupsByRoleID(ctx context.Context, roleID string) ([]*ent.Group, error)
	IsRoleInGroup(ctx context.Context, groupID string, roleID string) (bool, error)
	IsGroupInRole(ctx context.Context, roleID string, groupID string) (bool, error)
}

// groupRoleRepository implements the GroupRoleRepositoryInterface.
type groupRoleRepository struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.GroupRole]
}

// NewGroupRoleRepository creates a new group role repository.
func NewGroupRoleRepository(d *data.Data) GroupRoleRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &groupRoleRepository{ec, rc, cache.NewCache[ent.GroupRole](rc, "ncse_group_role")}
}

// Create group role
func (r *groupRoleRepository) Create(ctx context.Context, body *structs.GroupRole) (*ent.GroupRole, error) {
	// create builder.
	builder := r.ec.GroupRole.Create()
	// set values.
	builder.SetNillableGroupID(&body.GroupID)
	builder.SetNillableRoleID(&body.RoleID)
	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(ctx, "groupRoleRepo.Create error: %v", err)
		return nil, err
	}
	return row, nil
}

// GetByGroupID Find role by group id
func (r *groupRoleRepository) GetByGroupID(ctx context.Context, id string) (*ent.GroupRole, error) {
	// create builder.
	builder := r.ec.GroupRole.Query()
	// set conditions.
	builder.Where(groupRoleEnt.GroupIDEQ(id))
	// execute the builder.
	row, err := builder.Only(ctx)
	if err != nil {
		log.Errorf(ctx, "groupRoleRepo.GetProfile error: %v", err)
		return nil, err
	}
	return row, nil
}

// GetByGroupIDs Find roles by group ids
func (r *groupRoleRepository) GetByGroupIDs(ctx context.Context, ids []string) ([]*ent.GroupRole, error) {
	// create builder.
	builder := r.ec.GroupRole.Query()
	// set conditions.
	builder.Where(groupRoleEnt.GroupIDIn(ids...))
	// execute the builder.
	rows, err := builder.All(ctx)
	if err != nil {
		log.Errorf(ctx, "groupRoleRepo.GetByGroupIDs error: %v", err)
		return nil, err
	}
	return rows, nil
}

// GetByRoleID Find role by role id
func (r *groupRoleRepository) GetByRoleID(ctx context.Context, id string) (*ent.GroupRole, error) {
	// create builder.
	builder := r.ec.GroupRole.Query()
	// set conditions.
	builder.Where(groupRoleEnt.RoleIDEQ(id))
	// execute the builder.
	row, err := builder.Only(ctx)
	if err != nil {
		log.Errorf(ctx, "groupRoleRepo.GetProfile error: %v", err)
		return nil, err
	}
	return row, nil
}

// GetByRoleIDs Find roles by role ids
func (r *groupRoleRepository) GetByRoleIDs(ctx context.Context, ids []string) ([]*ent.GroupRole, error) {
	// create builder.
	builder := r.ec.GroupRole.Query()
	// set conditions.
	builder.Where(groupRoleEnt.RoleIDIn(ids...))
	// execute the builder.
	rows, err := builder.All(ctx)
	if err != nil {
		log.Errorf(ctx, "groupRoleRepo.GetByRoleIDs error: %v", err)
		return nil, err
	}
	return rows, nil
}

// Delete group role
func (r *groupRoleRepository) Delete(ctx context.Context, gid, rid string) error {
	if _, err := r.ec.GroupRole.Delete().Where(groupRoleEnt.GroupIDEQ(gid), groupRoleEnt.RoleIDEQ(rid)).Exec(ctx); err != nil {
		log.Errorf(ctx, "groupRoleRepo.Delete error: %v", err)
		return err
	}
	return nil
}

// DeleteAllByGroupID Delete all group role
func (r *groupRoleRepository) DeleteAllByGroupID(ctx context.Context, id string) error {
	if _, err := r.ec.GroupRole.Delete().Where(groupRoleEnt.GroupIDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(ctx, "groupRoleRepo.DeleteAllByGroupID error: %v", err)
		return err
	}
	return nil
}

// DeleteAllByRoleID Delete all group role
func (r *groupRoleRepository) DeleteAllByRoleID(ctx context.Context, id string) error {
	if _, err := r.ec.GroupRole.Delete().Where(groupRoleEnt.RoleIDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(ctx, "groupRoleRepo.DeleteAllByRoleID error: %v", err)
		return err
	}
	return nil
}

// GetRolesByGroupID retrieves all roles under a group.
func (r *groupRoleRepository) GetRolesByGroupID(ctx context.Context, groupID string) ([]string, error) {
	groupRoles, err := r.ec.GroupRole.Query().Where(groupRoleEnt.GroupIDEQ(groupID)).All(ctx)
	if err != nil {
		log.Errorf(ctx, "groupRoleRepo.GetRolesByGroupID error: %v", err)
		return nil, err
	}

	var roleIDs []string
	for _, groupRole := range groupRoles {
		roleIDs = append(roleIDs, groupRole.RoleID)
	}

	return roleIDs, nil
}

// GetGroupsByRoleID retrieves all groups under a role.
func (r *groupRoleRepository) GetGroupsByRoleID(ctx context.Context, roleID string) ([]*ent.Group, error) {
	groupRoles, err := r.ec.GroupRole.Query().Where(groupRoleEnt.RoleIDEQ(roleID)).All(ctx)
	if err != nil {
		log.Errorf(ctx, "groupRoleRepo.GetGroupsByRoleID error: %v", err)
		return nil, err
	}

	var groupIDs []string
	for _, groupRole := range groupRoles {
		groupIDs = append(groupIDs, groupRole.GroupID)
	}

	groups, err := r.ec.Group.Query().Where(groupEnt.IDIn(groupIDs...)).All(ctx)
	if err != nil {
		log.Errorf(ctx, "groupRoleRepo.GetGroupsByRoleID error: %v", err)
		return nil, err
	}

	return groups, nil
}

// IsRoleInGroup verifies if a role belongs to a specific group.
func (r *groupRoleRepository) IsRoleInGroup(ctx context.Context, groupID string, roleID string) (bool, error) {
	count, err := r.ec.GroupRole.Query().Where(groupRoleEnt.GroupIDEQ(groupID), groupRoleEnt.RoleIDEQ(roleID)).Count(ctx)
	if err != nil {
		log.Errorf(ctx, "groupRoleRepo.IsRoleInGroup error: %v", err)
		return false, err
	}
	return count > 0, nil
}

// IsGroupInRole verifies if a group belongs to a specific role.
func (r *groupRoleRepository) IsGroupInRole(ctx context.Context, groupID string, roleID string) (bool, error) {
	count, err := r.ec.GroupRole.Query().Where(groupRoleEnt.GroupIDEQ(groupID), groupRoleEnt.RoleIDEQ(roleID)).Count(ctx)
	if err != nil {
		log.Errorf(ctx, "groupRoleRepo.IsGroupInRole error: %v", err)
		return false, err
	}
	return count > 0, nil
}
