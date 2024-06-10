package repo

import (
	"context"
	"stocms/internal/data"
	"stocms/internal/data/ent"
	userDomainRoleEnt "stocms/internal/data/ent/userdomainrole"
	"stocms/internal/data/structs"
	"stocms/pkg/cache"
	"stocms/pkg/log"

	"github.com/redis/go-redis/v9"
)

// UserDomainRole represents the user domain role repository interface.
type UserDomainRole interface {
	Create(ctx context.Context, body *structs.UserDomainRole) (*ent.UserDomainRole, error)
	GetByUserID(ctx context.Context, userID string) (*ent.UserDomainRole, error)
	GetByDomainID(ctx context.Context, domainID string) ([]*ent.UserDomainRole, error)
	GetByRoleID(ctx context.Context, roleID string) ([]*ent.UserDomainRole, error)
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteByDomainID(ctx context.Context, domainID string) error
	DeleteByRoleID(ctx context.Context, roleID string) error
	DeleteAllByUserID(ctx context.Context, userID string) error
	DeleteAllByDomainID(ctx context.Context, domainID string) error
	DeleteAllByRoleID(ctx context.Context, roleID string) error
}

// userDomainRoleRepo implements the UserDomainRole interface.
type userDomainRoleRepo struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.UserDomainRole]
}

// NewUserDomainRole creates a new user domain role repository.
func NewUserDomainRole(d *data.Data) UserDomainRole {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &userDomainRoleRepo{ec, rc, cache.NewCache[ent.UserDomainRole](rc, cache.Key("sc_user_domain_role"), true)}
}

// Create creates a new user domain role.
func (r *userDomainRoleRepo) Create(ctx context.Context, body *structs.UserDomainRole) (*ent.UserDomainRole, error) {

	builder := r.ec.UserDomainRole.Create()
	builder.SetID(body.UserID)
	builder.SetDomainID(body.DomainID)
	builder.SetRoleID(body.RoleID)

	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(nil, "userDomainRoleRepo.Create error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByUserID retrieves user domain role by user ID.
func (r *userDomainRoleRepo) GetByUserID(ctx context.Context, userID string) (*ent.UserDomainRole, error) {
	row, err := r.ec.UserDomainRole.
		Query().
		Where(userDomainRoleEnt.IDEQ(userID)).
		Only(ctx)

	if err != nil {
		log.Errorf(nil, "userDomainRoleRepo.GetByUserID error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByDomainID retrieves user domain roles by domain ID.
func (r *userDomainRoleRepo) GetByDomainID(ctx context.Context, domainID string) ([]*ent.UserDomainRole, error) {
	rows, err := r.ec.UserDomainRole.
		Query().
		Where(userDomainRoleEnt.DomainID(domainID)).
		All(ctx)

	if err != nil {
		log.Errorf(nil, "userDomainRoleRepo.GetByDomainID error: %v\n", err)
		return nil, err
	}

	return rows, nil
}

// GetByRoleID retrieves user domain roles by role ID.
func (r *userDomainRoleRepo) GetByRoleID(ctx context.Context, roleID string) ([]*ent.UserDomainRole, error) {
	rows, err := r.ec.UserDomainRole.
		Query().
		Where(userDomainRoleEnt.RoleID(roleID)).
		All(ctx)

	if err != nil {
		log.Errorf(nil, "userDomainRoleRepo.GetByRoleID error: %v\n", err)
		return nil, err
	}

	return rows, nil
}

// DeleteByUserID deletes user domain roles by user ID.
func (r *userDomainRoleRepo) DeleteByUserID(ctx context.Context, userID string) error {
	if _, err := r.ec.UserDomainRole.Delete().Where(userDomainRoleEnt.IDEQ(userID)).Exec(ctx); err != nil {
		log.Errorf(nil, "userDomainRoleRepo.DeleteByUserID error: %v\n", err)
		return err
	}

	return nil
}

// DeleteByDomainID deletes user domain roles by domain ID.
func (r *userDomainRoleRepo) DeleteByDomainID(ctx context.Context, domainID string) error {
	if _, err := r.ec.UserDomainRole.Delete().Where(userDomainRoleEnt.DomainID(domainID)).Exec(ctx); err != nil {
		log.Errorf(nil, "userDomainRoleRepo.DeleteByDomainID error: %v\n", err)
		return err
	}

	return nil
}

// DeleteByRoleID deletes user domain roles by role ID.
func (r *userDomainRoleRepo) DeleteByRoleID(ctx context.Context, roleID string) error {
	if _, err := r.ec.UserDomainRole.Delete().Where(userDomainRoleEnt.RoleID(roleID)).Exec(ctx); err != nil {
		log.Errorf(nil, "userDomainRoleRepo.DeleteByRoleID error: %v\n", err)
		return err
	}

	return nil
}

// DeleteAllByUserID deletes all user domain roles by user ID.
func (r *userDomainRoleRepo) DeleteAllByUserID(ctx context.Context, userID string) error {
	if _, err := r.ec.UserDomainRole.Delete().Where(userDomainRoleEnt.IDEQ(userID)).Exec(ctx); err != nil {
		log.Errorf(nil, "userDomainRoleRepo.DeleteAllByUserID error: %v\n", err)
		return err
	}

	return nil
}

// DeleteAllByDomainID deletes all user domain roles by domain ID.
func (r *userDomainRoleRepo) DeleteAllByDomainID(ctx context.Context, domainID string) error {
	if _, err := r.ec.UserDomainRole.Delete().Where(userDomainRoleEnt.DomainID(domainID)).Exec(ctx); err != nil {
		log.Errorf(nil, "userDomainRoleRepo.DeleteAllByDomainID error: %v\n", err)
		return err
	}

	return nil
}

// DeleteAllByRoleID deletes all user domain roles by role ID.
func (r *userDomainRoleRepo) DeleteAllByRoleID(ctx context.Context, roleID string) error {
	if _, err := r.ec.UserDomainRole.Delete().Where(userDomainRoleEnt.RoleID(roleID)).Exec(ctx); err != nil {
		log.Errorf(nil, "userDomainRoleRepo.DeleteAllByRoleID error: %v\n", err)
		return err
	}

	return nil
}
