package repo

import (
	"context"
	"fmt"
	"ncobase/internal/data"
	"ncobase/internal/data/ent"
	roleEnt "ncobase/internal/data/ent/role"
	userEnt "ncobase/internal/data/ent/user"
	userRoleEnt "ncobase/internal/data/ent/userrole"
	"ncobase/internal/data/structs"
	"ncobase/pkg/cache"
	"ncobase/pkg/log"

	"github.com/redis/go-redis/v9"
)

// UserRole represents the user role repository interface.
type UserRole interface {
	Create(ctx context.Context, body *structs.UserRole) (*ent.UserRole, error)
	GetByIDAndRoleID(ctx context.Context, uid, rid string) (*ent.UserRole, error)
	GetByIDs(ctx context.Context, ids []string) ([]*ent.UserRole, error)
	GetByRoleIDs(ctx context.Context, ids []string) ([]*ent.UserRole, error)
	Delete(ctx context.Context, uid, rid string) error
	DeleteAllByID(ctx context.Context, id string) error
	DeleteAllByRoleID(ctx context.Context, id string) error
	VerifyUserRole(ctx context.Context, userID, roleID string) (bool, error)
	GetRolesByUserID(ctx context.Context, userID string) ([]*ent.Role, error)
	GetUsersByRoleID(ctx context.Context, roleID string) ([]*ent.User, error)
	IsUserInRole(ctx context.Context, userID string, roleID string) (bool, error)
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

// VerifyUserRole verifies if a user has a specific role.
func (r *userRoleRepo) VerifyUserRole(ctx context.Context, userID, roleID string) (bool, error) {
	count, err := r.ec.UserRole.Query().
		Where(
			userRoleEnt.IDEQ(userID),
			userRoleEnt.RoleIDEQ(roleID),
		).
		Count(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userRoleRepo.VerifyUserRole error: %v\n", err)
		return false, err
	}
	return count > 0, nil
}

// Create create user role
func (r *userRoleRepo) Create(ctx context.Context, body *structs.UserRole) (*ent.UserRole, error) {
	// Verify if the user already has the role
	exists, err := r.VerifyUserRole(ctx, body.UserID, body.RoleID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("user %s already has role %s", body.UserID, body.RoleID)
	}

	// create builder.
	builder := r.ec.UserRole.Create()

	// set values.
	builder.SetID(body.UserID)
	builder.SetRoleID(body.RoleID)

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userRoleRepo.Create error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByIDAndRoleID find role by user id and role id
func (r *userRoleRepo) GetByIDAndRoleID(ctx context.Context, uid, rid string) (*ent.UserRole, error) {
	row, err := r.ec.UserRole.
		Query().
		Where(userRoleEnt.IDEQ(uid), userRoleEnt.RoleIDEQ(rid)).
		Only(ctx)

	if err != nil {
		log.Errorf(context.Background(), "userRoleRepo.GetByIDAndRoleID error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// GetByIDs find roles by user ids
func (r *userRoleRepo) GetByIDs(ctx context.Context, ids []string) ([]*ent.UserRole, error) {
	rows, err := r.ec.UserRole.
		Query().
		Where(userRoleEnt.IDIn(ids...)).
		All(ctx)

	if err != nil {
		log.Errorf(context.Background(), "userRoleRepo.GetByIDs error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// GetByRoleID find role by role id
func (r *userRoleRepo) GetByRoleID(ctx context.Context, id string) (*ent.UserRole, error) {
	row, err := r.ec.UserRole.
		Query().
		Where(userRoleEnt.RoleIDEQ(id)).
		Only(ctx)

	if err != nil {
		log.Errorf(context.Background(), "userRoleRepo.GetByRoleID error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// GetByRoleIDs find roles by role ids
func (r *userRoleRepo) GetByRoleIDs(ctx context.Context, ids []string) ([]*ent.UserRole, error) {
	rows, err := r.ec.UserRole.
		Query().
		Where(userRoleEnt.RoleIDIn(ids...)).
		All(ctx)

	if err != nil {
		log.Errorf(context.Background(), "userRoleRepo.GetByRoleIDs error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// Delete delete user role
func (r *userRoleRepo) Delete(ctx context.Context, uid, rid string) error {
	if _, err := r.ec.UserRole.Delete().Where(userRoleEnt.IDEQ(uid), userRoleEnt.RoleIDEQ(rid)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userRoleRepo.DeleteByID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteAllByID delete all user roles by user ID
func (r *userRoleRepo) DeleteAllByID(ctx context.Context, id string) error {
	if _, err := r.ec.UserRole.Delete().Where(userRoleEnt.IDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userRoleRepo.DeleteAllByID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteAllByRoleID delete all user roles by role ID
func (r *userRoleRepo) DeleteAllByRoleID(ctx context.Context, id string) error {
	if _, err := r.ec.UserRole.Delete().Where(userRoleEnt.RoleIDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userRoleRepo.DeleteAllByRoleID error: %v\n", err)
		return err
	}
	return nil
}

// GetRolesByUserID retrieves all roles assigned to a user.
func (r *userRoleRepo) GetRolesByUserID(ctx context.Context, userID string) ([]*ent.Role, error) {
	userRoles, err := r.ec.UserRole.Query().Where(userRoleEnt.IDEQ(userID)).All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userRoleRepo.GetRolesByUserID error: %v\n", err)
		return nil, err
	}

	var roleIDs []string
	for _, userRole := range userRoles {
		roleIDs = append(roleIDs, userRole.RoleID)
	}

	roles, err := r.ec.Role.Query().Where(roleEnt.IDIn(roleIDs...)).All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userRoleRepo.GetRolesByUserID error: %v\n", err)
		return nil, err
	}

	return roles, nil
}

// GetUsersByRoleID retrieves all users assigned to a role.
func (r *userRoleRepo) GetUsersByRoleID(ctx context.Context, roleID string) ([]*ent.User, error) {
	userRoles, err := r.ec.UserRole.Query().Where(userRoleEnt.RoleIDEQ(roleID)).All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userRoleRepo.GetUsersByRoleID error: %v\n", err)
		return nil, err
	}

	var userIDs []string
	for _, userRole := range userRoles {
		userIDs = append(userIDs, userRole.ID)
	}

	users, err := r.ec.User.Query().Where(userEnt.IDIn(userIDs...)).All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userRoleRepo.GetUsersByRoleID error: %v\n", err)
		return nil, err
	}

	return users, nil
}

// IsUserInRole verifies if a user has a specific role.
func (r *userRoleRepo) IsUserInRole(ctx context.Context, userID string, roleID string) (bool, error) {
	count, err := r.ec.UserRole.Query().Where(userRoleEnt.IDEQ(userID), userRoleEnt.RoleIDEQ(roleID)).Count(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userRoleRepo.IsUserInRole error: %v\n", err)
		return false, err
	}
	return count > 0, nil
}
