package repo

import (
	"context"
	"ncobase/common/cache"
	"ncobase/common/log"
	"ncobase/internal/data"
	"ncobase/internal/data/ent"
	roleEnt "ncobase/internal/data/ent/role"
	userTenantRoleEnt "ncobase/internal/data/ent/usertenantrole"
	"ncobase/internal/data/structs"

	"github.com/redis/go-redis/v9"
)

// UserTenantRole represents the user tenant role repository interface.
type UserTenantRole interface {
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
	GetRolesByUserAndTenant(ctx context.Context, userID, tenantID string) ([]*ent.Role, error)
	IsUserInRoleInTenant(ctx context.Context, userID string, tenantID string, roleID string) (bool, error)
}

// userTenantRoleRepo implements the UserTenantRole interface.
type userTenantRoleRepo struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.UserTenantRole]
}

// NewUserTenantRole creates a new user tenant role repository.
func NewUserTenantRole(d *data.Data) UserTenantRole {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &userTenantRoleRepo{ec, rc, cache.NewCache[ent.UserTenantRole](rc, cache.Key("nb_user_tenant_role"), true)}
}

// Create creates a new user tenant role.
func (r *userTenantRoleRepo) Create(ctx context.Context, body *structs.UserTenantRole) (*ent.UserTenantRole, error) {

	builder := r.ec.UserTenantRole.Create()
	builder.SetID(body.UserID)
	builder.SetTenantID(body.TenantID)
	builder.SetRoleID(body.RoleID)

	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.Create error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByUserID retrieves user tenant role by user ID.
func (r *userTenantRoleRepo) GetByUserID(ctx context.Context, userID string) (*ent.UserTenantRole, error) {
	row, err := r.ec.UserTenantRole.
		Query().
		Where(userTenantRoleEnt.IDEQ(userID)).
		Only(ctx)

	if err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.GetByUserID error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByTenantID retrieves user tenant roles by tenant ID.
func (r *userTenantRoleRepo) GetByTenantID(ctx context.Context, tenantID string) ([]*ent.UserTenantRole, error) {
	rows, err := r.ec.UserTenantRole.
		Query().
		Where(userTenantRoleEnt.TenantID(tenantID)).
		All(ctx)

	if err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.GetByTenantID error: %v\n", err)
		return nil, err
	}

	return rows, nil
}

// GetByRoleID retrieves user tenant roles by role ID.
func (r *userTenantRoleRepo) GetByRoleID(ctx context.Context, roleID string) ([]*ent.UserTenantRole, error) {
	rows, err := r.ec.UserTenantRole.
		Query().
		Where(userTenantRoleEnt.RoleID(roleID)).
		All(ctx)

	if err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.GetByRoleID error: %v\n", err)
		return nil, err
	}

	return rows, nil
}

// DeleteByUserIDAndTenantID deletes user tenant role by user ID and tenant ID.
func (r *userTenantRoleRepo) DeleteByUserIDAndTenantID(ctx context.Context, userID, tenantID string) error {
	if _, err := r.ec.UserTenantRole.Delete().Where(userTenantRoleEnt.IDEQ(userID), userTenantRoleEnt.TenantID(tenantID)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.DeleteByUserIDAndTenantID error: %v\n", err)
		return err
	}

	return nil
}

// DeleteByUserIDAndRoleID deletes user tenant role by user ID and role ID.
func (r *userTenantRoleRepo) DeleteByUserIDAndRoleID(ctx context.Context, userID, roleID string) error {
	if _, err := r.ec.UserTenantRole.Delete().Where(userTenantRoleEnt.IDEQ(userID), userTenantRoleEnt.RoleID(roleID)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.DeleteByUserIDAndRoleID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteByTenantIDAndRoleID deletes user tenant role by tenant ID and role ID.
func (r *userTenantRoleRepo) DeleteByTenantIDAndRoleID(ctx context.Context, tenantID, roleID string) error {
	if _, err := r.ec.UserTenantRole.Delete().Where(userTenantRoleEnt.TenantID(tenantID), userTenantRoleEnt.RoleID(roleID)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.DeleteByTenantIDAndRoleID error: %v\n", err)
		return err
	}

	return nil
}

// DeleteByUserIDAndTenantIDAndRoleID deletes user tenant role by user ID, tenant ID and role ID.
func (r *userTenantRoleRepo) DeleteByUserIDAndTenantIDAndRoleID(ctx context.Context, userID, tenantID, roleID string) error {
	if _, err := r.ec.UserTenantRole.Delete().Where(userTenantRoleEnt.IDEQ(userID), userTenantRoleEnt.TenantID(tenantID), userTenantRoleEnt.RoleID(roleID)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.DeleteByUserIDAndTenantIDAndRoleID error: %v\n", err)
		return err
	}

	return nil
}

// DeleteAllByUserID deletes all user tenant roles by user ID.
func (r *userTenantRoleRepo) DeleteAllByUserID(ctx context.Context, userID string) error {
	if _, err := r.ec.UserTenantRole.Delete().Where(userTenantRoleEnt.IDEQ(userID)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.DeleteAllByUserID error: %v\n", err)
		return err
	}

	return nil
}

// DeleteAllByTenantID deletes all user tenant roles by tenant ID.
func (r *userTenantRoleRepo) DeleteAllByTenantID(ctx context.Context, tenantID string) error {
	if _, err := r.ec.UserTenantRole.Delete().Where(userTenantRoleEnt.TenantID(tenantID)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.DeleteAllByTenantID error: %v\n", err)
		return err
	}

	return nil
}

// DeleteAllByRoleID deletes all user tenant roles by role ID.
func (r *userTenantRoleRepo) DeleteAllByRoleID(ctx context.Context, roleID string) error {
	if _, err := r.ec.UserTenantRole.Delete().Where(userTenantRoleEnt.RoleID(roleID)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.DeleteAllByRoleID error: %v\n", err)
		return err
	}

	return nil
}

// GetRolesByUserAndTenant retrieves all roles a user has in a tenant.
func (r *userTenantRoleRepo) GetRolesByUserAndTenant(ctx context.Context, userID string, tenantID string) ([]*ent.Role, error) {
	userTenantRoles, err := r.ec.UserTenantRole.Query().Where(userTenantRoleEnt.IDEQ(userID), userTenantRoleEnt.TenantIDEQ(tenantID)).All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.GetRolesByUserAndTenant error: %v\n", err)
		return nil, err
	}

	var roleIDs []string
	for _, userRole := range userTenantRoles {
		roleIDs = append(roleIDs, userRole.RoleID)
	}

	roles, err := r.ec.Role.Query().Where(roleEnt.IDIn(roleIDs...)).All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.GetRolesByUserAndTenant error: %v\n", err)
		return nil, err
	}

	return roles, nil
}

// IsUserInRoleInTenant verifies if a user has a specific role in a tenant.
func (r *userTenantRoleRepo) IsUserInRoleInTenant(ctx context.Context, userID string, tenantID string, roleID string) (bool, error) {
	count, err := r.ec.UserTenantRole.Query().Where(userTenantRoleEnt.IDEQ(userID), userTenantRoleEnt.TenantIDEQ(tenantID), userTenantRoleEnt.RoleIDEQ(roleID)).Count(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userTenantRoleRepo.IsUserInRoleInTenant error: %v\n", err)
		return false, err
	}
	return count > 0, nil
}
