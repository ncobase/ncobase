package repo

import (
	"context"
	"ncobase/internal/data"
	"ncobase/internal/data/ent"
	groupEnt "ncobase/internal/data/ent/group"
	userEnt "ncobase/internal/data/ent/user"
	userGroupEnt "ncobase/internal/data/ent/usergroup"
	"ncobase/internal/data/structs"
	"ncobase/pkg/cache"
	"ncobase/pkg/log"

	"github.com/redis/go-redis/v9"
)

// UserGroup represents the user group repository interface.
type UserGroup interface {
	Create(ctx context.Context, body *structs.UserGroup) (*ent.UserGroup, error)
	GetByUserID(ctx context.Context, id string) (*ent.UserGroup, error)
	GetByGroupID(ctx context.Context, id string) (*ent.UserGroup, error)
	GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserGroup, error)
	GetByGroupIDs(ctx context.Context, ids []string) ([]*ent.UserGroup, error)
	Delete(ctx context.Context, uid, gid string) error
	DeleteAllByUserID(ctx context.Context, id string) error
	DeleteAllByGroupID(ctx context.Context, id string) error
	GetGroupsByUserID(ctx context.Context, userID string) ([]*ent.Group, error)
	GetUsersByGroupID(ctx context.Context, groupID string) ([]*ent.User, error)
	IsUserInGroup(ctx context.Context, userID string, groupID string) (bool, error)
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
	return &userGroupRepo{ec, rc, cache.NewCache[ent.UserGroup](rc, cache.Key("nb_user_group"), true)}
}

// Create create user group
func (r *userGroupRepo) Create(ctx context.Context, body *structs.UserGroup) (*ent.UserGroup, error) {

	// create builder.
	builder := r.ec.UserGroup.Create()
	// set values.
	builder.SetNillableID(&body.UserID)
	builder.SetNillableGroupID(&body.GroupID)

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userGroupRepo.Create error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByUserID find group by user id
func (r *userGroupRepo) GetByUserID(ctx context.Context, id string) (*ent.UserGroup, error) {
	row, err := r.ec.UserGroup.
		Query().
		Where(userGroupEnt.IDEQ(id)).
		Only(ctx)

	if err != nil {
		log.Errorf(context.Background(), "userGroupRepo.GetProfile error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// GetByUserIDs find groups by user ids
func (r *userGroupRepo) GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserGroup, error) {
	rows, err := r.ec.UserGroup.
		Query().
		Where(userGroupEnt.IDIn(ids...)).
		All(ctx)

	if err != nil {
		log.Errorf(context.Background(), "userGroupRepo.GetByUserIDs error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// GetByGroupID find group by group id
func (r *userGroupRepo) GetByGroupID(ctx context.Context, id string) (*ent.UserGroup, error) {
	row, err := r.ec.UserGroup.
		Query().
		Where(userGroupEnt.GroupIDEQ(id)).
		Only(ctx)

	if err != nil {
		log.Errorf(context.Background(), "userGroupRepo.GetProfile error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// GetByGroupIDs find groups by group ids
func (r *userGroupRepo) GetByGroupIDs(ctx context.Context, ids []string) ([]*ent.UserGroup, error) {
	rows, err := r.ec.UserGroup.
		Query().
		Where(userGroupEnt.GroupIDIn(ids...)).
		All(ctx)

	if err != nil {
		log.Errorf(context.Background(), "userGroupRepo.GetByGroupIDs error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// Delete delete user group
func (r *userGroupRepo) Delete(ctx context.Context, uid, gid string) error {
	if _, err := r.ec.UserGroup.Delete().Where(userGroupEnt.IDEQ(uid), userGroupEnt.GroupIDEQ(gid)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userGroupRepo.DeleteByUserID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteAllByUserID delete all user group
func (r *userGroupRepo) DeleteAllByUserID(ctx context.Context, id string) error {
	if _, err := r.ec.UserGroup.Delete().Where(userGroupEnt.IDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userGroupRepo.DeleteAllByUserID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteAllByGroupID delete all user group
func (r *userGroupRepo) DeleteAllByGroupID(ctx context.Context, id string) error {
	if _, err := r.ec.UserGroup.Delete().Where(userGroupEnt.GroupIDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userGroupRepo.DeleteAllByGroupID error: %v\n", err)
		return err
	}
	return nil
}

// GetGroupsByUserID retrieves all groups a user belongs to.
func (r *userGroupRepo) GetGroupsByUserID(ctx context.Context, userID string) ([]*ent.Group, error) {
	userGroups, err := r.ec.UserGroup.Query().Where(userGroupEnt.IDEQ(userID)).All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userGroupRepo.GetGroupsByUserID error: %v\n", err)
		return nil, err
	}
	var groupIDs []string
	for _, group := range userGroups {
		groupIDs = append(groupIDs, group.GroupID)
	}

	groups, err := r.ec.Group.Query().Where(groupEnt.IDIn(groupIDs...)).All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userGroupRepo.GetGroupsByUserID error: %v\n", err)
		return nil, err
	}

	return groups, nil
}

// GetUsersByGroupID retrieves all users in a group.
func (r *userGroupRepo) GetUsersByGroupID(ctx context.Context, groupID string) ([]*ent.User, error) {
	userGroups, err := r.ec.UserGroup.Query().Where(userGroupEnt.GroupIDEQ(groupID)).All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userGroupRepo.GetUsersByGroupID error: %v\n", err)
		return nil, err
	}
	var userIDs []string
	for _, userGroup := range userGroups {
		userIDs = append(userIDs, userGroup.ID)
	}

	users, err := r.ec.User.Query().Where(userEnt.IDIn(userIDs...)).All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userGroupRepo.GetUsersByGroupID error: %v\n", err)
		return nil, err
	}
	return users, nil
}

// IsUserInGroup verifies if a user belongs to a specific group.
func (r *userGroupRepo) IsUserInGroup(ctx context.Context, userID string, groupID string) (bool, error) {
	count, err := r.ec.UserGroup.Query().Where(userGroupEnt.IDEQ(userID), userGroupEnt.GroupIDEQ(groupID)).Count(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userGroupRepo.IsUserInGroup error: %v\n", err)
		return false, err
	}
	return count > 0, nil
}
