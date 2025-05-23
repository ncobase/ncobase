package repository

import (
	"context"
	"ncobase/space/data"
	"ncobase/space/data/ent"
	userGroupEnt "ncobase/space/data/ent/usergroup"
	"ncobase/space/structs"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/logging/logger"

	"github.com/redis/go-redis/v9"
)

// UserGroupRepositoryInterface represents the user group repository interface.
type UserGroupRepositoryInterface interface {
	Create(ctx context.Context, body *structs.UserGroup) (*ent.UserGroup, error)
	GetByUserID(ctx context.Context, id string) ([]*ent.UserGroup, error)
	GetByGroupID(ctx context.Context, id string) ([]*ent.UserGroup, error)
	GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserGroup, error)
	GetByGroupIDs(ctx context.Context, ids []string) ([]*ent.UserGroup, error)
	GetByGroupIDAndRole(ctx context.Context, id string, role structs.UserRole) ([]*ent.UserGroup, error)
	GetUserGroup(ctx context.Context, uid, gid string) (*ent.UserGroup, error)
	Delete(ctx context.Context, uid, gid string) error
	DeleteAllByUserID(ctx context.Context, id string) error
	DeleteAllByGroupID(ctx context.Context, id string) error
	GetGroupsByUserID(ctx context.Context, userID string) ([]string, error)
	GetUsersByGroupID(ctx context.Context, groupID string) ([]string, error)
	IsUserInGroup(ctx context.Context, userID string, groupID string) (bool, error)
	UserHasRole(ctx context.Context, userID string, groupID string, role structs.UserRole) (bool, error)
}

// userGroupRepository implements the UserGroupRepositoryInterface.
type userGroupRepository struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.UserGroup]
}

// NewUserGroupRepository creates a new user group repository.
func NewUserGroupRepository(d *data.Data) UserGroupRepositoryInterface {
	ec := d.GetMasterEntClient()
	rc := d.GetRedis()
	return &userGroupRepository{ec, rc, cache.NewCache[ent.UserGroup](rc, "ncse_user_group")}
}

// Create create user group
func (r *userGroupRepository) Create(ctx context.Context, body *structs.UserGroup) (*ent.UserGroup, error) {
	// create builder.
	builder := r.ec.UserGroup.Create()
	// set values.
	builder.SetUserID(body.UserID)
	builder.SetGroupID(body.GroupID)

	// Set role if provided
	if body.Role != "" {
		builder.SetRole(string(body.Role))
	} else {
		builder.SetRole(string(structs.RoleMember)) // Default role
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.Create error: %v", err)
		return nil, err
	}

	return row, nil
}

// GetByUserID find groups by user id
func (r *userGroupRepository) GetByUserID(ctx context.Context, id string) ([]*ent.UserGroup, error) {
	// create builder.
	builder := r.ec.UserGroup.Query()
	// set conditions.
	builder.Where(userGroupEnt.UserIDEQ(id))
	// execute the builder.
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.GetByUserID error: %v", err)
		return nil, err
	}
	return rows, nil
}

// GetByUserIDs find groups by user ids
func (r *userGroupRepository) GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserGroup, error) {
	// create builder.
	builder := r.ec.UserGroup.Query()
	// set conditions.
	builder.Where(userGroupEnt.UserIDIn(ids...))
	// execute the builder.
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.GetByUserIDs error: %v", err)
		return nil, err
	}
	return rows, nil
}

// GetByGroupID find users by group id
func (r *userGroupRepository) GetByGroupID(ctx context.Context, id string) ([]*ent.UserGroup, error) {
	// create builder.
	builder := r.ec.UserGroup.Query()
	// set conditions.
	builder.Where(userGroupEnt.GroupIDEQ(id))
	// execute the builder.
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.GetByGroupID error: %v", err)
		return nil, err
	}
	return rows, nil
}

// GetByGroupIDs find users by group ids
func (r *userGroupRepository) GetByGroupIDs(ctx context.Context, ids []string) ([]*ent.UserGroup, error) {
	// create builder.
	builder := r.ec.UserGroup.Query()
	// set conditions.
	builder.Where(userGroupEnt.GroupIDIn(ids...))
	// execute the builder.
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.GetByGroupIDs error: %v", err)
		return nil, err
	}
	return rows, nil
}

// GetByGroupIDAndRole find users by group id and role
func (r *userGroupRepository) GetByGroupIDAndRole(ctx context.Context, id string, role structs.UserRole) ([]*ent.UserGroup, error) {
	// create builder.
	builder := r.ec.UserGroup.Query()
	// set conditions.
	builder.Where(
		userGroupEnt.GroupIDEQ(id),
		userGroupEnt.RoleEQ(string(role)),
	)
	// execute the builder.
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.GetByGroupIDAndRole error: %v", err)
		return nil, err
	}
	return rows, nil
}

// GetUserGroup gets a specific user-group relation
func (r *userGroupRepository) GetUserGroup(ctx context.Context, uid, gid string) (*ent.UserGroup, error) {
	// create builder.
	builder := r.ec.UserGroup.Query()
	// set conditions.
	builder.Where(
		userGroupEnt.UserIDEQ(uid),
		userGroupEnt.GroupIDEQ(gid),
	)
	// execute the builder.
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.GetUserGroup error: %v", err)
		return nil, err
	}
	return row, nil
}

// Delete delete user group
func (r *userGroupRepository) Delete(ctx context.Context, uid, gid string) error {
	if _, err := r.ec.UserGroup.Delete().Where(userGroupEnt.UserIDEQ(uid), userGroupEnt.GroupIDEQ(gid)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userGroupRepo.Delete error: %v", err)
		return err
	}
	return nil
}

// DeleteAllByUserID delete all user group by user id
func (r *userGroupRepository) DeleteAllByUserID(ctx context.Context, id string) error {
	if _, err := r.ec.UserGroup.Delete().Where(userGroupEnt.UserIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userGroupRepo.DeleteAllByUserID error: %v", err)
		return err
	}
	return nil
}

// DeleteAllByGroupID delete all user group by group id
func (r *userGroupRepository) DeleteAllByGroupID(ctx context.Context, id string) error {
	if _, err := r.ec.UserGroup.Delete().Where(userGroupEnt.GroupIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userGroupRepo.DeleteAllByGroupID error: %v", err)
		return err
	}
	return nil
}

// GetGroupsByUserID retrieves all groups a user belongs to.
func (r *userGroupRepository) GetGroupsByUserID(ctx context.Context, userID string) ([]string, error) {
	userGroups, err := r.ec.UserGroup.Query().Where(userGroupEnt.UserIDEQ(userID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.GetGroupsByUserID error: %v", err)
		return nil, err
	}
	var groupIDs []string
	for _, group := range userGroups {
		groupIDs = append(groupIDs, group.GroupID)
	}

	return groupIDs, nil
}

// GetUsersByGroupID retrieves all users in a group.
func (r *userGroupRepository) GetUsersByGroupID(ctx context.Context, groupID string) ([]string, error) {
	userGroups, err := r.ec.UserGroup.Query().Where(userGroupEnt.GroupIDEQ(groupID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.GetUsersByGroupID error: %v", err)
		return nil, err
	}
	var userIDs []string
	for _, userGroup := range userGroups {
		userIDs = append(userIDs, userGroup.UserID)
	}

	return userIDs, nil
}

// IsUserInGroup verifies if a user belongs to a specific group.
func (r *userGroupRepository) IsUserInGroup(ctx context.Context, userID string, groupID string) (bool, error) {
	count, err := r.ec.UserGroup.Query().Where(userGroupEnt.UserIDEQ(userID), userGroupEnt.GroupIDEQ(groupID)).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.IsUserInGroup error: %v", err)
		return false, err
	}
	return count > 0, nil
}

// UserHasRole verifies if a user has a specific role in a group.
func (r *userGroupRepository) UserHasRole(ctx context.Context, userID string, groupID string, role structs.UserRole) (bool, error) {
	count, err := r.ec.UserGroup.Query().Where(
		userGroupEnt.UserIDEQ(userID),
		userGroupEnt.GroupIDEQ(groupID),
		userGroupEnt.RoleEQ(string(role)),
	).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.UserHasRole error: %v", err)
		return false, err
	}
	return count > 0, nil
}
