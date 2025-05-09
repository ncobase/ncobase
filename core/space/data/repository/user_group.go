package repository

import (
	"context"
	"ncobase/core/space/data"
	"ncobase/core/space/data/ent"
	userGroupEnt "ncobase/core/space/data/ent/usergroup"
	"ncobase/core/space/structs"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/logging/logger"

	"github.com/redis/go-redis/v9"
)

// UserGroupRepositoryInterface represents the user group repository interface.
type UserGroupRepositoryInterface interface {
	Create(ctx context.Context, body *structs.UserGroup) (*ent.UserGroup, error)
	GetByUserID(ctx context.Context, id string) (*ent.UserGroup, error)
	GetByGroupID(ctx context.Context, id string) (*ent.UserGroup, error)
	GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserGroup, error)
	GetByGroupIDs(ctx context.Context, ids []string) ([]*ent.UserGroup, error)
	Delete(ctx context.Context, uid, gid string) error
	DeleteAllByUserID(ctx context.Context, id string) error
	DeleteAllByGroupID(ctx context.Context, id string) error
	GetGroupsByUserID(ctx context.Context, userID string) ([]string, error)
	GetUsersByGroupID(ctx context.Context, groupID string) ([]string, error)
	IsUserInGroup(ctx context.Context, userID string, groupID string) (bool, error)
}

// userGroupRepository implements the UserGroupRepositoryInterface.
type userGroupRepository struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.UserGroup]
}

// NewUserGroupRepository creates a new user group repository.
func NewUserGroupRepository(d *data.Data) UserGroupRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &userGroupRepository{ec, rc, cache.NewCache[ent.UserGroup](rc, "ncse_user_group")}
}

// Create create user group
func (r *userGroupRepository) Create(ctx context.Context, body *structs.UserGroup) (*ent.UserGroup, error) {

	// create builder.
	builder := r.ec.UserGroup.Create()
	// set values.
	builder.SetNillableUserID(&body.UserID)
	builder.SetNillableGroupID(&body.GroupID)

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.Create error: %v", err)
		return nil, err
	}

	return row, nil
}

// GetByUserID find group by user id
func (r *userGroupRepository) GetByUserID(ctx context.Context, id string) (*ent.UserGroup, error) {
	// create builder.
	builder := r.ec.UserGroup.Query()
	// set conditions.
	builder.Where(userGroupEnt.UserIDEQ(id))
	// execute the builder.
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.GetProfile error: %v", err)
		return nil, err
	}
	return row, nil
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

// GetByGroupID find group by group id
func (r *userGroupRepository) GetByGroupID(ctx context.Context, id string) (*ent.UserGroup, error) {
	// create builder.
	builder := r.ec.UserGroup.Query()
	// set conditions.
	builder.Where(userGroupEnt.GroupIDEQ(id))
	// execute the builder.
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "userGroupRepo.GetProfile error: %v", err)
		return nil, err
	}
	return row, nil
}

// GetByGroupIDs find groups by group ids
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

// Delete delete user group
func (r *userGroupRepository) Delete(ctx context.Context, uid, gid string) error {
	if _, err := r.ec.UserGroup.Delete().Where(userGroupEnt.UserIDEQ(uid), userGroupEnt.GroupIDEQ(gid)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userGroupRepo.DeleteByUserID error: %v", err)
		return err
	}
	return nil
}

// DeleteAllByUserID delete all user group
func (r *userGroupRepository) DeleteAllByUserID(ctx context.Context, id string) error {
	if _, err := r.ec.UserGroup.Delete().Where(userGroupEnt.UserIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userGroupRepo.DeleteAllByUserID error: %v", err)
		return err
	}
	return nil
}

// DeleteAllByGroupID delete all user group
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
