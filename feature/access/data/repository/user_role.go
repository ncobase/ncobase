package repository

import (
	"context"
	"fmt"
	"ncobase/common/cache"
	"ncobase/common/log"
	"ncobase/feature/access/data"
	"ncobase/feature/access/data/ent"
	roleEnt "ncobase/feature/access/data/ent/role"
	userRoleEnt "ncobase/feature/access/data/ent/userrole"
	"ncobase/feature/access/structs"

	"github.com/redis/go-redis/v9"
)

// UserRoleRepositoryInterface represents the user role repository interface.
type UserRoleRepositoryInterface interface {
	Create(ctx context.Context, body *structs.UserRole) (*ent.UserRole, error)
	GetByIDAndRoleID(ctx context.Context, uid, rid string) (*ent.UserRole, error)
	GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserRole, error)
	GetByRoleIDs(ctx context.Context, ids []string) ([]*ent.UserRole, error)
	Delete(ctx context.Context, uid, rid string) error
	DeleteAllByUserID(ctx context.Context, id string) error
	DeleteAllByRoleID(ctx context.Context, id string) error
	VerifyUserRole(ctx context.Context, userID, roleID string) (bool, error)
	GetRolesByUserID(ctx context.Context, userID string) ([]*ent.Role, error)
	GetUsersByRoleID(ctx context.Context, roleID string) ([]string, error)
	IsUserInRole(ctx context.Context, userID string, roleID string) (bool, error)
}

// userRoleRepository implements the UserRoleRepositoryInterface.
type userRoleRepository struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.UserRole]
}

// NewUserRoleRepository creates a new user role repository.
func NewUserRoleRepository(d *data.Data) UserRoleRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &userRoleRepository{ec, rc, cache.NewCache[ent.UserRole](rc, "ncse_user_role")}
}

// VerifyUserRole verifies if a user has a specific role.
func (r *userRoleRepository) VerifyUserRole(ctx context.Context, userID, roleID string) (bool, error) {
	count, err := r.ec.UserRole.Query().
		Where(
			userRoleEnt.UserIDEQ(userID),
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
func (r *userRoleRepository) Create(ctx context.Context, body *structs.UserRole) (*ent.UserRole, error) {
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
	builder.SetNillableUserID(&body.UserID)
	builder.SetNillableRoleID(&body.RoleID)

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userRoleRepo.Create error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByIDAndRoleID find role by user id and role id
func (r *userRoleRepository) GetByIDAndRoleID(ctx context.Context, uid, rid string) (*ent.UserRole, error) {
	// create builder.
	builder := r.ec.UserRole.Query()
	// set conditions.
	builder.Where(userRoleEnt.UserIDEQ(uid), userRoleEnt.RoleIDEQ(rid))
	// execute the builder.
	row, err := builder.Only(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userRoleRepo.GetByIDAndRoleID error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// GetByUserIDs find roles by user ids
func (r *userRoleRepository) GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserRole, error) {
	// create builder.
	builder := r.ec.UserRole.Query()
	// set conditions.
	builder.Where(userRoleEnt.UserIDIn(ids...))
	// execute the builder.
	rows, err := builder.All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userRoleRepo.GetByIDs error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// GetByRoleID find role by role id
func (r *userRoleRepository) GetByRoleID(ctx context.Context, id string) (*ent.UserRole, error) {
	// create builder.
	builder := r.ec.UserRole.Query()
	// set condition.
	builder.Where(userRoleEnt.RoleIDEQ(id))
	// execute the builder.
	row, err := builder.Only(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userRoleRepo.GetByRoleID error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// GetByRoleIDs find roles by role ids
func (r *userRoleRepository) GetByRoleIDs(ctx context.Context, ids []string) ([]*ent.UserRole, error) {
	// create builder.
	builder := r.ec.UserRole.Query()
	// set conditions.
	builder.Where(userRoleEnt.RoleIDIn(ids...))
	// execute the builder.
	rows, err := builder.All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userRoleRepo.GetByRoleIDs error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// Delete delete user role
func (r *userRoleRepository) Delete(ctx context.Context, uid, rid string) error {
	if _, err := r.ec.UserRole.Delete().Where(userRoleEnt.UserIDEQ(uid), userRoleEnt.RoleIDEQ(rid)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userRoleRepo.DeleteByID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteAllByUserID delete all user roles by user ID
func (r *userRoleRepository) DeleteAllByUserID(ctx context.Context, id string) error {
	if _, err := r.ec.UserRole.Delete().Where(userRoleEnt.UserIDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userRoleRepo.DeleteAllByID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteAllByRoleID delete all user roles by role ID
func (r *userRoleRepository) DeleteAllByRoleID(ctx context.Context, id string) error {
	if _, err := r.ec.UserRole.Delete().Where(userRoleEnt.RoleIDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userRoleRepo.DeleteAllByRoleID error: %v\n", err)
		return err
	}
	return nil
}

// GetRolesByUserID retrieves all roles assigned to a user.
func (r *userRoleRepository) GetRolesByUserID(ctx context.Context, userID string) ([]*ent.Role, error) {
	userRoles, err := r.ec.UserRole.Query().Where(userRoleEnt.UserIDEQ(userID)).All(ctx)
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
func (r *userRoleRepository) GetUsersByRoleID(ctx context.Context, roleID string) ([]string, error) {
	userRoles, err := r.ec.UserRole.Query().Where(userRoleEnt.RoleIDEQ(roleID)).All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userRoleRepo.GetUsersByRoleID error: %v\n", err)
		return nil, err
	}

	var userIDs []string
	for _, userRole := range userRoles {
		userIDs = append(userIDs, userRole.UserID)
	}
	return userIDs, nil
}

// IsUserInRole verifies if a user has a specific role.
func (r *userRoleRepository) IsUserInRole(ctx context.Context, userID string, roleID string) (bool, error) {
	count, err := r.ec.UserRole.Query().Where(userRoleEnt.UserIDEQ(userID), userRoleEnt.RoleIDEQ(roleID)).Count(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userRoleRepo.IsUserInRole error: %v\n", err)
		return false, err
	}
	return count > 0, nil
}
