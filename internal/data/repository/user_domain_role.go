package repo

import (
	"context"
	"stocms/internal/data"
	"stocms/internal/data/ent"
	roleEnt "stocms/internal/data/ent/role"
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
	DeleteByUserIDAndDomainID(ctx context.Context, userID, domainID string) error
	DeleteByUserIDAndRoleID(ctx context.Context, userID, roleID string) error
	DeleteByDomainIDAndRoleID(ctx context.Context, domainID, roleID string) error
	DeleteByUserIDAndDomainIDAndRoleID(ctx context.Context, userID, domainID, roleID string) error
	DeleteAllByUserID(ctx context.Context, userID string) error
	DeleteAllByDomainID(ctx context.Context, domainID string) error
	DeleteAllByRoleID(ctx context.Context, roleID string) error
	GetRolesByUserAndDomain(ctx context.Context, userID, domainID string) ([]*ent.Role, error)
	IsUserInRoleInDomain(ctx context.Context, userID string, domainID string, roleID string) (bool, error)
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

// DeleteByUserIDAndDomainID deletes user domain role by user ID and domain ID.
func (r *userDomainRoleRepo) DeleteByUserIDAndDomainID(ctx context.Context, userID, domainID string) error {
	if _, err := r.ec.UserDomainRole.Delete().Where(userDomainRoleEnt.IDEQ(userID), userDomainRoleEnt.DomainID(domainID)).Exec(ctx); err != nil {
		log.Errorf(nil, "userDomainRoleRepo.DeleteByUserIDAndDomainID error: %v\n", err)
		return err
	}

	return nil
}

// DeleteByUserIDAndRoleID deletes user domain role by user ID and role ID.
func (r *userDomainRoleRepo) DeleteByUserIDAndRoleID(ctx context.Context, userID, roleID string) error {
	if _, err := r.ec.UserDomainRole.Delete().Where(userDomainRoleEnt.IDEQ(userID), userDomainRoleEnt.RoleID(roleID)).Exec(ctx); err != nil {
		log.Errorf(nil, "userDomainRoleRepo.DeleteByUserIDAndRoleID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteByDomainIDAndRoleID deletes user domain role by domain ID and role ID.
func (r *userDomainRoleRepo) DeleteByDomainIDAndRoleID(ctx context.Context, domainID, roleID string) error {
	if _, err := r.ec.UserDomainRole.Delete().Where(userDomainRoleEnt.DomainID(domainID), userDomainRoleEnt.RoleID(roleID)).Exec(ctx); err != nil {
		log.Errorf(nil, "userDomainRoleRepo.DeleteByDomainIDAndRoleID error: %v\n", err)
		return err
	}

	return nil
}

// DeleteByUserIDAndDomainIDAndRoleID deletes user domain role by user ID, domain ID and role ID.
func (r *userDomainRoleRepo) DeleteByUserIDAndDomainIDAndRoleID(ctx context.Context, userID, domainID, roleID string) error {
	if _, err := r.ec.UserDomainRole.Delete().Where(userDomainRoleEnt.IDEQ(userID), userDomainRoleEnt.DomainID(domainID), userDomainRoleEnt.RoleID(roleID)).Exec(ctx); err != nil {
		log.Errorf(nil, "userDomainRoleRepo.DeleteByUserIDAndDomainIDAndRoleID error: %v\n", err)
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

// GetRolesByUserAndDomain retrieves all roles a user has in a domain.
func (r *userDomainRoleRepo) GetRolesByUserAndDomain(ctx context.Context, userID string, domainID string) ([]*ent.Role, error) {
	userDomainRoles, err := r.ec.UserDomainRole.Query().Where(userDomainRoleEnt.IDEQ(userID), userDomainRoleEnt.DomainIDEQ(domainID)).All(ctx)
	if err != nil {
		log.Errorf(nil, "userDomainRoleRepo.GetRolesByUserAndDomain error: %v\n", err)
		return nil, err
	}

	var roleIDs []string
	for _, userRole := range userDomainRoles {
		roleIDs = append(roleIDs, userRole.RoleID)
	}

	roles, err := r.ec.Role.Query().Where(roleEnt.IDIn(roleIDs...)).All(ctx)
	if err != nil {
		log.Errorf(nil, "userDomainRoleRepo.GetRolesByUserAndDomain error: %v\n", err)
		return nil, err
	}

	return roles, nil
}

// IsUserInRoleInDomain verifies if a user has a specific role in a domain.
func (r *userDomainRoleRepo) IsUserInRoleInDomain(ctx context.Context, userID string, domainID string, roleID string) (bool, error) {
	count, err := r.ec.UserDomainRole.Query().Where(userDomainRoleEnt.IDEQ(userID), userDomainRoleEnt.DomainIDEQ(domainID), userDomainRoleEnt.RoleIDEQ(roleID)).Count(ctx)
	if err != nil {
		log.Errorf(nil, "userDomainRoleRepo.IsUserInRoleInDomain error: %v\n", err)
		return false, err
	}
	return count > 0, nil
}
