package repository

import (
	"context"
	"ncobase/access/data"
	"ncobase/access/data/ent"
	userTenantRoleEnt "ncobase/access/data/ent/usertenantrole"
	"ncobase/access/structs"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/logging/logger"

	"github.com/redis/go-redis/v9"
)

// UserTenantRoleRepositoryInterface represents the user tenant role repository interface.
type UserTenantRoleRepositoryInterface interface {
	Create(ctx context.Context, body *structs.UserTenantRole) (*ent.UserTenantRole, error)
	GetByUserID(ctx context.Context, u string) (*ent.UserTenantRole, error)
	GetByTenantID(ctx context.Context, t string) ([]*ent.UserTenantRole, error)
	GetByRoleID(ctx context.Context, r string) ([]*ent.UserTenantRole, error)
	DeleteByUserIDAndTenantID(ctx context.Context, u, t string) error
	DeleteByUserIDAndRoleID(ctx context.Context, u, r string) error
	DeleteByTenantIDAndRoleID(ctx context.Context, t, r string) error
	DeleteByUserIDAndTenantIDAndRoleID(ctx context.Context, u, t, r string) error
	DeleteAllByUserID(ctx context.Context, u string) error
	DeleteAllByTenantID(ctx context.Context, t string) error
	DeleteAllByRoleID(ctx context.Context, r string) error
	GetRolesByUserAndTenant(ctx context.Context, u, t string) ([]string, error)
	IsUserInRoleInTenant(ctx context.Context, u, t, r string) (bool, error)
}

// userTenantRoleRepository implements the UserTenantRoleRepositoryInterface.
type userTenantRoleRepository struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.UserTenantRole]
}

// NewUserTenantRoleRepository creates a new user tenant role repository.
func NewUserTenantRoleRepository(d *data.Data) UserTenantRoleRepositoryInterface {
	ec := d.GetMasterEntClient()
	rc := d.GetRedis()
	return &userTenantRoleRepository{ec, rc, cache.NewCache[ent.UserTenantRole](rc, "ncse_user_tenant_role")}
}

// Create creates a new user tenant role.
func (r *userTenantRoleRepository) Create(ctx context.Context, body *structs.UserTenantRole) (*ent.UserTenantRole, error) {
	// create builder.
	builder := r.ec.UserTenantRole.Create()
	// set values.
	builder.SetNillableUserID(&body.UserID)
	builder.SetNillableTenantID(&body.TenantID)
	builder.SetNillableRoleID(&body.RoleID)
	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.Create error: %v", err)
		return nil, err
	}
	return row, nil
}

// GetByUserID retrieves user tenant role by user ID.
func (r *userTenantRoleRepository) GetByUserID(ctx context.Context, u string) (*ent.UserTenantRole, error) {
	// create builder.
	builder := r.ec.UserTenantRole.Query()
	// set conditions.
	builder.Where(userTenantRoleEnt.UserIDEQ(u))
	// execute the builder.
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.GetByUserID error: %v", err)
		return nil, err
	}
	return row, nil
}

// GetByTenantID retrieves user tenant roles by tenant ID.
func (r *userTenantRoleRepository) GetByTenantID(ctx context.Context, t string) ([]*ent.UserTenantRole, error) {
	// create builder.
	builder := r.ec.UserTenantRole.Query()
	// set conditions.
	builder.Where(userTenantRoleEnt.TenantID(t))
	// execute the builder.
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.GetByTenantID error: %v", err)
		return nil, err
	}
	return rows, nil
}

// GetByRoleID retrieves user tenant roles by role ID.
func (r *userTenantRoleRepository) GetByRoleID(ctx context.Context, rid string) ([]*ent.UserTenantRole, error) {
	// create builder.
	builder := r.ec.UserTenantRole.Query()
	// set conditions.
	builder.Where(userTenantRoleEnt.RoleID(rid))
	// execute the builder.
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.GetByRoleID error: %v", err)
		return nil, err
	}
	return rows, nil
}

// DeleteByUserIDAndTenantID deletes user tenant role by user ID and tenant ID.
func (r *userTenantRoleRepository) DeleteByUserIDAndTenantID(ctx context.Context, u, t string) error {
	if _, err := r.ec.UserTenantRole.Delete().Where(userTenantRoleEnt.UserIDEQ(u), userTenantRoleEnt.TenantID(t)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.DeleteByUserIDAndTenantID error: %v", err)
		return err
	}

	return nil
}

// DeleteByUserIDAndRoleID deletes user tenant role by user ID and role ID.
func (r *userTenantRoleRepository) DeleteByUserIDAndRoleID(ctx context.Context, u, rid string) error {
	if _, err := r.ec.UserTenantRole.Delete().Where(userTenantRoleEnt.UserIDEQ(u), userTenantRoleEnt.RoleID(rid)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.DeleteByUserIDAndRoleID error: %v", err)
		return err
	}
	return nil
}

// DeleteByTenantIDAndRoleID deletes user tenant role by tenant ID and role ID.
func (r *userTenantRoleRepository) DeleteByTenantIDAndRoleID(ctx context.Context, t, rid string) error {
	if _, err := r.ec.UserTenantRole.Delete().Where(userTenantRoleEnt.TenantID(t), userTenantRoleEnt.RoleID(rid)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.DeleteByTenantIDAndRoleID error: %v", err)
		return err
	}

	return nil
}

// DeleteByUserIDAndTenantIDAndRoleID deletes user tenant role by user ID, tenant ID and role ID.
func (r *userTenantRoleRepository) DeleteByUserIDAndTenantIDAndRoleID(ctx context.Context, u, t, rid string) error {
	if _, err := r.ec.UserTenantRole.Delete().Where(userTenantRoleEnt.UserIDEQ(u), userTenantRoleEnt.TenantID(t), userTenantRoleEnt.RoleID(rid)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.DeleteByUserIDAndTenantIDAndRoleID error: %v", err)
		return err
	}

	return nil
}

// DeleteAllByUserID deletes all user tenant roles by user ID.
func (r *userTenantRoleRepository) DeleteAllByUserID(ctx context.Context, u string) error {
	if _, err := r.ec.UserTenantRole.Delete().Where(userTenantRoleEnt.UserIDEQ(u)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.DeleteAllByUserID error: %v", err)
		return err
	}

	return nil
}

// DeleteAllByTenantID deletes all user tenant roles by tenant ID.
func (r *userTenantRoleRepository) DeleteAllByTenantID(ctx context.Context, t string) error {
	if _, err := r.ec.UserTenantRole.Delete().Where(userTenantRoleEnt.TenantID(t)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.DeleteAllByTenantID error: %v", err)
		return err
	}

	return nil
}

// DeleteAllByRoleID deletes all user tenant roles by role ID.
func (r *userTenantRoleRepository) DeleteAllByRoleID(ctx context.Context, rid string) error {
	if _, err := r.ec.UserTenantRole.Delete().Where(userTenantRoleEnt.RoleID(rid)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.DeleteAllByRoleID error: %v", err)
		return err
	}

	return nil
}

// GetRolesByUserAndTenant retrieves all roles a user has in a tenant.
func (r *userTenantRoleRepository) GetRolesByUserAndTenant(ctx context.Context, u string, t string) ([]string, error) {
	userTenantRoles, err := r.ec.UserTenantRole.Query().Where(userTenantRoleEnt.UserIDEQ(u), userTenantRoleEnt.TenantIDEQ(t)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.GetRolesByUserAndTenant error: %v", err)
		return nil, err
	}

	var roleIDs []string
	for _, userRole := range userTenantRoles {
		roleIDs = append(roleIDs, userRole.RoleID)
	}

	return roleIDs, nil
}

// IsUserInRoleInTenant verifies if a user has a specific role in a tenant.
func (r *userTenantRoleRepository) IsUserInRoleInTenant(ctx context.Context, u, t, rid string) (bool, error) {
	count, err := r.ec.UserTenantRole.Query().Where(userTenantRoleEnt.UserIDEQ(u), userTenantRoleEnt.TenantIDEQ(t), userTenantRoleEnt.RoleIDEQ(rid)).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRoleRepo.IsUserInRoleInTenant error: %v", err)
		return false, err
	}
	return count > 0, nil
}
