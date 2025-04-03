package repository

import (
	"context"
	"ncobase/core/tenant/data"
	"ncobase/core/tenant/data/ent"
	tenantEnt "ncobase/core/tenant/data/ent/tenant"
	userTenantEnt "ncobase/core/tenant/data/ent/usertenant"
	"ncobase/core/tenant/structs"

	"ncobase/ncore/data/cache"
	"ncobase/ncore/logger"

	"github.com/redis/go-redis/v9"
)

// UserTenantRepositoryInterface represents the user tenant repository interface.
type UserTenantRepositoryInterface interface {
	Create(ctx context.Context, body *structs.UserTenant) (*ent.UserTenant, error)
	GetByUserID(ctx context.Context, id string) (*ent.UserTenant, error)
	GetByTenantID(ctx context.Context, id string) (*ent.UserTenant, error)
	GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserTenant, error)
	GetByTenantIDs(ctx context.Context, ids []string) ([]*ent.UserTenant, error)
	Delete(ctx context.Context, uid, did string) error
	DeleteAllByUserID(ctx context.Context, id string) error
	DeleteAllByTenantID(ctx context.Context, id string) error
	GetTenantsByUserID(ctx context.Context, userID string) ([]*ent.Tenant, error)
	IsTenantInUser(ctx context.Context, userID string, tenantID string) (bool, error)
}

// userTenantRepository implements the UserTenantRepositoryInterface.
type userTenantRepository struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.UserTenant]
}

// NewUserTenantRepository creates a new user tenant repository.
func NewUserTenantRepository(d *data.Data) UserTenantRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &userTenantRepository{ec, rc, cache.NewCache[ent.UserTenant](rc, "ncse_user_tenant")}
}

// Create  creates a new user tenant
func (r *userTenantRepository) Create(ctx context.Context, body *structs.UserTenant) (*ent.UserTenant, error) {
	// create builder.
	builder := r.ec.UserTenant.Create()
	// set values.
	builder.SetNillableUserID(&body.UserID)
	builder.SetNillableTenantID(&body.TenantID)
	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRepo.Create error: %v", err)
		return nil, err
	}
	return row, nil
}

// GetByUserID find tenant by user id
func (r *userTenantRepository) GetByUserID(ctx context.Context, id string) (*ent.UserTenant, error) {
	// create builder.
	builder := r.ec.UserTenant.Query()
	// set conditions.
	builder.Where(userTenantEnt.UserIDEQ(id))
	// execute the builder.
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRepo.GetProfile error: %v", err)
		return nil, err
	}
	return row, nil
}

// GetByUserIDs find tenants by user ids
func (r *userTenantRepository) GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserTenant, error) {
	// create builder.
	builder := r.ec.UserTenant.Query()
	// set conditions.
	builder.Where(userTenantEnt.UserIDIn(ids...))
	// execute the builder.
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRepo.GetByUserIDs error: %v", err)
		return nil, err
	}
	return rows, nil
}

// GetByTenantID find tenant by tenant id
func (r *userTenantRepository) GetByTenantID(ctx context.Context, id string) (*ent.UserTenant, error) {
	// create builder.
	builder := r.ec.UserTenant.Query()
	// set conditions.
	builder.Where(userTenantEnt.TenantIDEQ(id))
	// execute the builder.
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRepo.GetProfile error: %v", err)
		return nil, err
	}
	return row, nil
}

// GetByTenantIDs find tenants by tenant ids
func (r *userTenantRepository) GetByTenantIDs(ctx context.Context, ids []string) ([]*ent.UserTenant, error) {
	// create builder.
	builder := r.ec.UserTenant.Query()
	// set conditions.
	builder.Where(userTenantEnt.TenantIDIn(ids...))
	// execute the builder.
	rows, err := builder.All(ctx)

	if err != nil {
		logger.Errorf(ctx, "userTenantRepo.GetByTenantIDs error: %v", err)
		return nil, err
	}
	return rows, nil
}

// Delete delete user tenant
func (r *userTenantRepository) Delete(ctx context.Context, uid, did string) error {
	if _, err := r.ec.UserTenant.Delete().Where(userTenantEnt.UserIDEQ(uid), userTenantEnt.TenantIDEQ(did)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userTenantRepo.Delete error: %v", err)
		return err
	}
	return nil
}

// DeleteAllByUserID delete all user tenant
func (r *userTenantRepository) DeleteAllByUserID(ctx context.Context, id string) error {
	if _, err := r.ec.UserTenant.Delete().Where(userTenantEnt.UserIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userTenantRepo.DeleteAllByUserID error: %v", err)
		return err
	}
	return nil
}

// DeleteAllByTenantID delete all user tenant
func (r *userTenantRepository) DeleteAllByTenantID(ctx context.Context, id string) error {
	if _, err := r.ec.UserTenant.Delete().Where(userTenantEnt.TenantIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userTenantRepo.DeleteAllByTenantID error: %v", err)
		return err
	}
	return nil
}

// GetTenantsByUserID retrieves all tenants a user belongs to.
func (r *userTenantRepository) GetTenantsByUserID(ctx context.Context, userID string) ([]*ent.Tenant, error) {
	userTenants, err := r.ec.UserTenant.Query().Where(userTenantEnt.UserIDEQ(userID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRepo.GetTenantsByUserID error: %v", err)
		return nil, err
	}

	var tenantIDs []string
	for _, userTenant := range userTenants {
		tenantIDs = append(tenantIDs, userTenant.TenantID)
	}

	tenants, err := r.ec.Tenant.Query().Where(tenantEnt.IDIn(tenantIDs...)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRepo.GetTenantsByUserID error: %v", err)
		return nil, err
	}

	return tenants, nil
}

// IsUserInTenant verifies if a user belongs to a specific tenant.
func (r *userTenantRepository) IsUserInTenant(ctx context.Context, userID string, tenantID string) (bool, error) {
	count, err := r.ec.UserTenant.Query().Where(userTenantEnt.UserIDEQ(userID), userTenantEnt.TenantIDEQ(tenantID)).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRepo.IsUserInTenant error: %v", err)
		return false, err
	}
	return count > 0, nil
}

// IsTenantInUser verifies if a tenant is assigned to a specific user.
func (r *userTenantRepository) IsTenantInUser(ctx context.Context, tenantID string, userID string) (bool, error) {
	count, err := r.ec.UserTenant.Query().Where(userTenantEnt.TenantIDEQ(tenantID), userTenantEnt.UserIDEQ(userID)).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "userTenantRepo.IsTenantInUser error: %v", err)
		return false, err
	}
	return count > 0, nil
}
