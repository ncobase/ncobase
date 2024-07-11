package repository

import (
	"context"
	"ncobase/feature/tenant/data"
	"ncobase/feature/tenant/data/ent"
	userTenantRoleEnt "ncobase/feature/tenant/data/ent/usertenantrole"
	"ncobase/feature/tenant/structs"

	"ncobase/common/cache"
	"ncobase/common/log"

	"github.com/redis/go-redis/v9"
)

// UserTenantRoleRepositoryInterface represents the user tenant role repository interface.
type UserTenantRoleRepositoryInterface interface {
	Create(ctx context.Context, body *structs.UserTenantRole) (*ent.UserTenantRole, error)
	GetByUserID(ctx context.Context, userID string) (*ent.UserTenantRole, error)
	GetByTenantID(ctx context.Context, tenantID string) ([]*ent.UserTenantRole, error)
	GetByRoleID(ctx context.Context, roleID string) ([]*ent.UserTenantRole, error)
	DeleteByUserIDAndTenantID(ctx context.Context, userID, tenantID string) error
	DeleteByUserIDAndRoleID(ctx context.Context, userID, roleID string) error
	DeleteByTenantIDAndRoleID(ctx context.Context, tenantID, roleID string) error
	DeleteByUserIDAndTenantIDAndRoleID(ctx context.Context, userID, tenantID, roleID string) error
	DeleteAllByUserID(ctx context.Context, userID string) error
	DeleteAllByTenantID(ctx context.Context, tenantID string) error
	DeleteAllByRoleID(ctx context.Context, roleID string) error
	GetRolesByUserAndTenant(ctx context.Context, userID, tenantID string) ([]string, error)
	IsUserInRoleInTenant(ctx context.Context, userID string, tenantID string, roleID string) (bool, error)
}

// userTenantRoleRepository implements the UserTenantRoleRepositoryInterface.
type userTenantRoleRepository struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.UserTenantRole]
}

// NewUserTenantRoleRepository creates a new user tenant role repository.
func NewUserTenantRoleRepository(d *data.Data) UserTenantRoleRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &userTenantRoleRepository{ec, rc, cache.NewCache[ent.UserTenantRole](rc, "nb_user_tenant_role")}
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
		log.Errorf(context.Background(), "userTenantRoleRepo.Create error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// GetByUserID retrieves user tenant role by user ID.
func (r *userTenantRoleRepository) GetByUserID(ctx context.Context, userID string) (*ent.UserTenantRole, error) {
	// create builder.
	builder := r.ec.UserTenantRole.Query()
	// set conditions.
	builder.Where(userTenantRoleEnt.UserIDEQ(userID))
	// execute the builder.
	row, err := builder.Only(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.GetByUserID error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// GetByTenantID retrieves user tenant roles by tenant ID.
func (r *userTenantRoleRepository) GetByTenantID(ctx context.Context, tenantID string) ([]*ent.UserTenantRole, error) {
	// create builder.
	builder := r.ec.UserTenantRole.Query()
	// set conditions.
	builder.Where(userTenantRoleEnt.TenantID(tenantID))
	// execute the builder.
	rows, err := builder.All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.GetByTenantID error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// GetByRoleID retrieves user tenant roles by role ID.
func (r *userTenantRoleRepository) GetByRoleID(ctx context.Context, roleID string) ([]*ent.UserTenantRole, error) {
	// create builder.
	builder := r.ec.UserTenantRole.Query()
	// set conditions.
	builder.Where(userTenantRoleEnt.RoleID(roleID))
	// execute the builder.
	rows, err := builder.All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.GetByRoleID error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// DeleteByUserIDAndTenantID deletes user tenant role by user ID and tenant ID.
func (r *userTenantRoleRepository) DeleteByUserIDAndTenantID(ctx context.Context, userID, tenantID string) error {
	if _, err := r.ec.UserTenantRole.Delete().Where(userTenantRoleEnt.UserIDEQ(userID), userTenantRoleEnt.TenantID(tenantID)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.DeleteByUserIDAndTenantID error: %v\n", err)
		return err
	}

	return nil
}

// DeleteByUserIDAndRoleID deletes user tenant role by user ID and role ID.
func (r *userTenantRoleRepository) DeleteByUserIDAndRoleID(ctx context.Context, userID, roleID string) error {
	if _, err := r.ec.UserTenantRole.Delete().Where(userTenantRoleEnt.UserIDEQ(userID), userTenantRoleEnt.RoleID(roleID)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.DeleteByUserIDAndRoleID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteByTenantIDAndRoleID deletes user tenant role by tenant ID and role ID.
func (r *userTenantRoleRepository) DeleteByTenantIDAndRoleID(ctx context.Context, tenantID, roleID string) error {
	if _, err := r.ec.UserTenantRole.Delete().Where(userTenantRoleEnt.TenantID(tenantID), userTenantRoleEnt.RoleID(roleID)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.DeleteByTenantIDAndRoleID error: %v\n", err)
		return err
	}

	return nil
}

// DeleteByUserIDAndTenantIDAndRoleID deletes user tenant role by user ID, tenant ID and role ID.
func (r *userTenantRoleRepository) DeleteByUserIDAndTenantIDAndRoleID(ctx context.Context, userID, tenantID, roleID string) error {
	if _, err := r.ec.UserTenantRole.Delete().Where(userTenantRoleEnt.UserIDEQ(userID), userTenantRoleEnt.TenantID(tenantID), userTenantRoleEnt.RoleID(roleID)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.DeleteByUserIDAndTenantIDAndRoleID error: %v\n", err)
		return err
	}

	return nil
}

// DeleteAllByUserID deletes all user tenant roles by user ID.
func (r *userTenantRoleRepository) DeleteAllByUserID(ctx context.Context, userID string) error {
	if _, err := r.ec.UserTenantRole.Delete().Where(userTenantRoleEnt.UserIDEQ(userID)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.DeleteAllByUserID error: %v\n", err)
		return err
	}

	return nil
}

// DeleteAllByTenantID deletes all user tenant roles by tenant ID.
func (r *userTenantRoleRepository) DeleteAllByTenantID(ctx context.Context, tenantID string) error {
	if _, err := r.ec.UserTenantRole.Delete().Where(userTenantRoleEnt.TenantID(tenantID)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.DeleteAllByTenantID error: %v\n", err)
		return err
	}

	return nil
}

// DeleteAllByRoleID deletes all user tenant roles by role ID.
func (r *userTenantRoleRepository) DeleteAllByRoleID(ctx context.Context, roleID string) error {
	if _, err := r.ec.UserTenantRole.Delete().Where(userTenantRoleEnt.RoleID(roleID)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.DeleteAllByRoleID error: %v\n", err)
		return err
	}

	return nil
}

// GetRolesByUserAndTenant retrieves all roles a user has in a tenant.
func (r *userTenantRoleRepository) GetRolesByUserAndTenant(ctx context.Context, userID string, tenantID string) ([]string, error) {
	userTenantRoles, err := r.ec.UserTenantRole.Query().Where(userTenantRoleEnt.UserIDEQ(userID), userTenantRoleEnt.TenantIDEQ(tenantID)).All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.GetRolesByUserAndTenant error: %v\n", err)
		return nil, err
	}

	var roleIDs []string
	for _, userRole := range userTenantRoles {
		roleIDs = append(roleIDs, userRole.RoleID)
	}

	return roleIDs, nil
}

// IsUserInRoleInTenant verifies if a user has a specific role in a tenant.
func (r *userTenantRoleRepository) IsUserInRoleInTenant(ctx context.Context, userID string, tenantID string, roleID string) (bool, error) {
	count, err := r.ec.UserTenantRole.Query().Where(userTenantRoleEnt.UserIDEQ(userID), userTenantRoleEnt.TenantIDEQ(tenantID), userTenantRoleEnt.RoleIDEQ(roleID)).Count(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.IsUserInRoleInTenant error: %v\n", err)
		return false, err
	}
	return count > 0, nil
}
