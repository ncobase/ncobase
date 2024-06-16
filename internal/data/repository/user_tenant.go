package repo

import (
	"context"
	"ncobase/common/cache"
	"ncobase/common/log"
	"ncobase/internal/data"
	"ncobase/internal/data/ent"
	tenantEnt "ncobase/internal/data/ent/tenant"
	userTenantEnt "ncobase/internal/data/ent/usertenant"
	"ncobase/internal/data/structs"

	"github.com/redis/go-redis/v9"
)

// UserTenant represents the user tenant repository interface.
type UserTenant interface {
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
	IsUserInTenant(ctx context.Context, tenantID string, userID string) (bool, error)
}

// userTenantRepo implements the User interface.
type userTenantRepo struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.UserTenant]
}

// NewUserTenant creates a new user tenant repository.
func NewUserTenant(d *data.Data) UserTenant {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &userTenantRepo{ec, rc, cache.NewCache[ent.UserTenant](rc, cache.Key("nb_user_tenant"), true)}
}

// Create  creates a new user tenant
func (r *userTenantRepo) Create(ctx context.Context, body *structs.UserTenant) (*ent.UserTenant, error) {

	// create builder.
	builder := r.ec.UserTenant.Create()
	// set values.
	builder.SetNillableID(&body.UserID)
	builder.SetNillableTenantID(&body.TenantID)

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userTenantRepo.Create error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByUserID find tenant by user id
func (r *userTenantRepo) GetByUserID(ctx context.Context, id string) (*ent.UserTenant, error) {
	row, err := r.ec.UserTenant.
		Query().
		Where(userTenantEnt.IDEQ(id)).
		Only(ctx)

	if err != nil {
		log.Errorf(context.Background(), "userTenantRepo.GetProfile error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// GetByUserIDs find tenants by user ids
func (r *userTenantRepo) GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserTenant, error) {
	rows, err := r.ec.UserTenant.
		Query().
		Where(userTenantEnt.IDIn(ids...)).
		All(ctx)

	if err != nil {
		log.Errorf(context.Background(), "userTenantRepo.GetByUserIDs error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// GetByTenantID find tenant by tenant id
func (r *userTenantRepo) GetByTenantID(ctx context.Context, id string) (*ent.UserTenant, error) {
	row, err := r.ec.UserTenant.
		Query().
		Where(userTenantEnt.TenantIDEQ(id)).
		Only(ctx)

	if err != nil {
		log.Errorf(context.Background(), "userTenantRepo.GetProfile error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// GetByTenantIDs find tenants by tenant ids
func (r *userTenantRepo) GetByTenantIDs(ctx context.Context, ids []string) ([]*ent.UserTenant, error) {
	rows, err := r.ec.UserTenant.
		Query().
		Where(userTenantEnt.TenantIDIn(ids...)).
		All(ctx)

	if err != nil {
		log.Errorf(context.Background(), "userTenantRepo.GetByTenantIDs error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// Delete delete user tenant
func (r *userTenantRepo) Delete(ctx context.Context, uid, did string) error {
	if _, err := r.ec.UserTenant.Delete().Where(userTenantEnt.IDEQ(uid), userTenantEnt.TenantIDEQ(did)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userTenantRepo.Delete error: %v\n", err)
		return err
	}
	return nil
}

// DeleteAllByUserID delete all user tenant
func (r *userTenantRepo) DeleteAllByUserID(ctx context.Context, id string) error {
	if _, err := r.ec.UserTenant.Delete().Where(userTenantEnt.IDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userTenantRepo.DeleteAllByUserID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteAllByTenantID delete all user tenant
func (r *userTenantRepo) DeleteAllByTenantID(ctx context.Context, id string) error {
	if _, err := r.ec.UserTenant.Delete().Where(userTenantEnt.TenantIDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userTenantRepo.DeleteAllByTenantID error: %v\n", err)
		return err
	}
	return nil
}

// GetTenantsByUserID retrieves all tenants a user belongs to.
func (r *userTenantRepo) GetTenantsByUserID(ctx context.Context, userID string) ([]*ent.Tenant, error) {
	userTenants, err := r.ec.UserTenant.Query().Where(userTenantEnt.IDEQ(userID)).All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userTenantRepo.GetTenantsByUserID error: %v\n", err)
		return nil, err
	}

	var tenantIDs []string
	for _, userTenant := range userTenants {
		tenantIDs = append(tenantIDs, userTenant.TenantID)
	}

	tenants, err := r.ec.Tenant.Query().Where(tenantEnt.IDIn(tenantIDs...)).All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userTenantRepo.GetTenantsByUserID error: %v\n", err)
		return nil, err
	}

	return tenants, nil
}

// IsUserInTenant verifies if a user belongs to a specific tenant.
func (r *userTenantRepo) IsUserInTenant(ctx context.Context, userID string, tenantID string) (bool, error) {
	count, err := r.ec.UserTenant.Query().Where(userTenantEnt.IDEQ(userID), userTenantEnt.TenantIDEQ(tenantID)).Count(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userTenantRepo.IsUserInTenant error: %v\n", err)
		return false, err
	}
	return count > 0, nil
}

// IsTenantInUser verifies if a tenant is assigned to a specific user.
func (r *userTenantRepo) IsTenantInUser(ctx context.Context, tenantID string, userID string) (bool, error) {
	count, err := r.ec.UserTenant.Query().Where(userTenantEnt.TenantIDEQ(tenantID), userTenantEnt.IDEQ(userID)).Count(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userTenantRepo.IsTenantInUser error: %v\n", err)
		return false, err
	}
	return count > 0, nil
}
