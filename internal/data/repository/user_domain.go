package repo

import (
	"context"
	"stocms/internal/data"
	"stocms/internal/data/ent"
	domainEnt "stocms/internal/data/ent/domain"
	userDomainEnt "stocms/internal/data/ent/userdomain"
	"stocms/internal/data/structs"
	"stocms/pkg/cache"
	"stocms/pkg/log"

	"github.com/redis/go-redis/v9"
)

// UserDomain represents the user domain repository interface.
type UserDomain interface {
	Create(ctx context.Context, body *structs.UserDomain) (*ent.UserDomain, error)
	GetByUserID(ctx context.Context, id string) (*ent.UserDomain, error)
	GetByDomainID(ctx context.Context, id string) (*ent.UserDomain, error)
	GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserDomain, error)
	GetByDomainIDs(ctx context.Context, ids []string) ([]*ent.UserDomain, error)
	DeleteByUserID(ctx context.Context, id string) error
	DeleteByDomainID(ctx context.Context, id string) error
	DeleteAllByUserID(ctx context.Context, id string) error
	DeleteAllByDomainID(ctx context.Context, id string) error
	GetDomainsByUserID(ctx context.Context, userID string) ([]*ent.Domain, error)
	IsDomainInUser(ctx context.Context, userID string, domainID string) (bool, error)
	IsUserInDomain(ctx context.Context, domainID string, userID string) (bool, error)
}

// userDomainRepo implements the User interface.
type userDomainRepo struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.UserDomain]
}

// NewUserDomain creates a new user domain repository.
func NewUserDomain(d *data.Data) UserDomain {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &userDomainRepo{ec, rc, cache.NewCache[ent.UserDomain](rc, cache.Key("sc_user_domain"), true)}
}

// Create - Create user domain
func (r *userDomainRepo) Create(ctx context.Context, body *structs.UserDomain) (*ent.UserDomain, error) {

	// create builder.
	builder := r.ec.UserDomain.Create()
	// set values.
	builder.SetNillableID(&body.UserID)
	builder.SetNillableDomainID(&body.DomainID)

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(nil, "userDomainRepo.Create error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByUserID - Find domain by user id
func (r *userDomainRepo) GetByUserID(ctx context.Context, id string) (*ent.UserDomain, error) {
	row, err := r.ec.UserDomain.
		Query().
		Where(userDomainEnt.IDEQ(id)).
		Only(ctx)

	if err != nil {
		log.Errorf(nil, "userDomainRepo.GetProfile error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// GetByUserIDs - Find domains by user ids
func (r *userDomainRepo) GetByUserIDs(ctx context.Context, ids []string) ([]*ent.UserDomain, error) {
	rows, err := r.ec.UserDomain.
		Query().
		Where(userDomainEnt.IDIn(ids...)).
		All(ctx)

	if err != nil {
		log.Errorf(nil, "userDomainRepo.GetByUserIDs error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// GetByDomainID - Find domain by domain id
func (r *userDomainRepo) GetByDomainID(ctx context.Context, id string) (*ent.UserDomain, error) {
	row, err := r.ec.UserDomain.
		Query().
		Where(userDomainEnt.DomainIDEQ(id)).
		Only(ctx)

	if err != nil {
		log.Errorf(nil, "userDomainRepo.GetProfile error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// GetByDomainIDs - Find domains by domain ids
func (r *userDomainRepo) GetByDomainIDs(ctx context.Context, ids []string) ([]*ent.UserDomain, error) {
	rows, err := r.ec.UserDomain.
		Query().
		Where(userDomainEnt.DomainIDIn(ids...)).
		All(ctx)

	if err != nil {
		log.Errorf(nil, "userDomainRepo.GetByDomainIDs error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// DeleteByUserID - Delete user domain
func (r *userDomainRepo) DeleteByUserID(ctx context.Context, id string) error {
	if _, err := r.ec.UserDomain.Delete().Where(userDomainEnt.IDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(nil, "userDomainRepo.DeleteByUserID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteByDomainID - Delete user domain
func (r *userDomainRepo) DeleteByDomainID(ctx context.Context, id string) error {
	if _, err := r.ec.UserDomain.Delete().Where(userDomainEnt.DomainIDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(nil, "userDomainRepo.DeleteByDomainID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteAllByUserID - Delete all user domain
func (r *userDomainRepo) DeleteAllByUserID(ctx context.Context, id string) error {
	if _, err := r.ec.UserDomain.Delete().Where(userDomainEnt.IDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(nil, "userDomainRepo.DeleteAllByUserID error: %v\n", err)
		return err
	}
	return nil
}

// DeleteAllByDomainID - Delete all user domain
func (r *userDomainRepo) DeleteAllByDomainID(ctx context.Context, id string) error {
	if _, err := r.ec.UserDomain.Delete().Where(userDomainEnt.DomainIDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(nil, "userDomainRepo.DeleteAllByDomainID error: %v\n", err)
		return err
	}
	return nil
}

// GetDomainsByUserID retrieves all domains a user belongs to.
func (r *userDomainRepo) GetDomainsByUserID(ctx context.Context, userID string) ([]*ent.Domain, error) {
	userDomains, err := r.ec.UserDomain.Query().Where(userDomainEnt.IDEQ(userID)).All(ctx)
	if err != nil {
		log.Errorf(nil, "userDomainRepo.GetDomainsByUserID error: %v\n", err)
		return nil, err
	}

	var domainIDs []string
	for _, userDomain := range userDomains {
		domainIDs = append(domainIDs, userDomain.DomainID)
	}

	domains, err := r.ec.Domain.Query().Where(domainEnt.IDIn(domainIDs...)).All(ctx)
	if err != nil {
		log.Errorf(nil, "userDomainRepo.GetDomainsByUserID error: %v\n", err)
		return nil, err
	}

	return domains, nil
}

// IsUserInDomain verifies if a user belongs to a specific domain.
func (r *userDomainRepo) IsUserInDomain(ctx context.Context, userID string, domainID string) (bool, error) {
	count, err := r.ec.UserDomain.Query().Where(userDomainEnt.IDEQ(userID), userDomainEnt.DomainIDEQ(domainID)).Count(ctx)
	if err != nil {
		log.Errorf(nil, "userDomainRepo.IsUserInDomain error: %v\n", err)
		return false, err
	}
	return count > 0, nil
}

// IsDomainInUser verifies if a domain is assigned to a specific user.
func (r *userDomainRepo) IsDomainInUser(ctx context.Context, domainID string, userID string) (bool, error) {
	count, err := r.ec.UserDomain.Query().Where(userDomainEnt.DomainIDEQ(domainID), userDomainEnt.IDEQ(userID)).Count(ctx)
	if err != nil {
		log.Errorf(nil, "userDomainRepo.IsDomainInUser error: %v\n", err)
		return false, err
	}
	return count > 0, nil
}
