package repo

import (
	"context"
	"errors"
	"fmt"
	"stocms/internal/data"
	"stocms/internal/data/ent"
	domainEnt "stocms/internal/data/ent/domain"
	"stocms/internal/data/structs"
	"stocms/pkg/cache"
	"stocms/pkg/log"
	"stocms/pkg/validator"
	"strings"

	"github.com/redis/go-redis/v9"
)

// Domain represents the domain repository interface.
type Domain interface {
	Create(ctx context.Context, body *structs.CreateDomainBody) (*ent.Domain, error)
	GetByID(ctx context.Context, id string) (*ent.Domain, error)
	GetByUser(ctx context.Context, user string) (*ent.Domain, error)
	GetIDByUser(ctx context.Context, user string) (string, error)
	Update(ctx context.Context, body *structs.UpdateDomainBody) (*ent.Domain, error)
	List(ctx context.Context, params *structs.ListDomainParams) ([]*ent.Domain, error)
	Delete(ctx context.Context, id string) error
	DeleteByUser(ctx context.Context, id string) error
}

// domainRepo implements the Domain interface.
type domainRepo struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.Domain]
}

// NewDomain creates a new domain repository.
func NewDomain(d *data.Data) Domain {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &domainRepo{ec, rc, cache.NewCache[ent.Domain](rc, cache.Key("sc_domain"), true)}
}

// Create create domain
func (r *domainRepo) Create(ctx context.Context, body *structs.CreateDomainBody) (*ent.Domain, error) {
	query := r.ec.Domain.
		Create().
		SetNillableName(&body.Name).
		SetNillableTitle(&body.Title).
		SetNillableURL(&body.URL).
		SetNillableLogo(&body.Logo).
		SetNillableLogoAlt(&body.LogoAlt).
		SetKeywords(strings.Join(body.Keywords, ",")).
		SetNillableCopyright(&body.Copyright).
		SetNillableDescription(&body.Description).
		SetDisabled(body.Disabled).
		SetUserID(body.UserID)

	if body.Order > 0 {
		query.SetOrder(int32(body.Order))
	}

	if !validator.IsNil(body.Extras) && len(body.Extras) > 0 {
		query.SetExtras(body.Extras)
	}

	row, err := query.Save(ctx)
	if err != nil {
		log.Errorf(nil, "domainRepo.Create error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByID get domain by id
func (r *domainRepo) GetByID(ctx context.Context, id string) (*ent.Domain, error) {
	cacheKey := fmt.Sprintf("%s", id)

	// Check cache first
	if cachedDomain, err := r.c.Get(ctx, cacheKey); err == nil {
		return cachedDomain, nil
	}

	// If not found in cache, query the database
	row, err := r.ec.Domain.
		Query().
		Where(domainEnt.IDEQ(id)).
		Only(ctx)

	if err != nil {
		log.Errorf(nil, " domainRepo.GetByID error: %v\n", err)
		return nil, err
	}

	// Cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(nil, "domainRepo.GetByID cache error: %v\n", err)
	}

	return row, nil
}

// GetByUser get domain by user
func (r *domainRepo) GetByUser(ctx context.Context, userID string) (*ent.Domain, error) {
	cacheKey := fmt.Sprintf("user:%s", userID)

	// Check cache first
	if cachedDomain, err := r.c.Get(ctx, cacheKey); err == nil {
		return cachedDomain, nil
	}

	// If not found in cache, query the database
	row, err := r.ec.Domain.
		Query().
		Where(domainEnt.UserIDEQ(userID)).
		Only(ctx)

	if err != nil {
		log.Errorf(nil, " domainRepo.GetByUser error: %v\n", err)
		return nil, err
	}

	// Cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(nil, "domainRepo.GetByUser cache error: %v\n", err)
	}

	return row, nil
}

// GetIDByUser get domain id by user id
func (r *domainRepo) GetIDByUser(ctx context.Context, userID string) (string, error) {
	id, err := r.ec.Domain.
		Query().
		Where(domainEnt.UserIDEQ(userID)).
		OnlyID(ctx)

	if err != nil {
		log.Errorf(nil, " domainRepo.FindDomainIDByUserID error: %v\n", err)
		return "", err
	}

	return id, nil
}

// Update update domain
func (r *domainRepo) Update(ctx context.Context, body *structs.UpdateDomainBody) (*ent.Domain, error) {
	domain, err := r.GetByID(ctx, body.ID)
	if err != nil {
		return nil, err
	}

	query := domain.
		Update().
		SetNillableName(&body.Name).
		SetNillableTitle(&body.Title).
		SetNillableURL(&body.URL).
		SetNillableLogo(&body.Logo).
		SetNillableLogoAlt(&body.LogoAlt).
		SetKeywords(strings.Join(body.Keywords, ",")).
		SetNillableCopyright(&body.Copyright).
		SetNillableDescription(&body.Description).
		SetDisabled(body.Disabled)

	if body.Order > 0 {
		query.SetOrder(int32(body.Order))
	}

	if !validator.IsNil(body.Extras) && len(body.Extras) > 0 {
		query.SetExtras(body.Extras)
	}

	if body.UserID != "" {
		query.SetUserID(body.UserID)
	}

	row, err := query.Save(ctx)
	if err != nil {
		log.Errorf(nil, "domainRepo.Update error: %v\n", err)
		return nil, err
	}

	// Remove from cache
	cacheKey := fmt.Sprintf("%s", body.ID)
	err = r.c.Delete(ctx, cacheKey)
	if err != nil {
		log.Errorf(nil, "domainRepo.Update cache error: %v\n", err)
	}
	cacheUserKey := fmt.Sprintf("user:%s", body.UserID)
	err = r.c.Delete(ctx, cacheUserKey)
	if err != nil {
		log.Errorf(nil, "domainRepo.Update cache error: %v\n", err)
	}

	return row, nil
}

// List get
func (r *domainRepo) List(ctx context.Context, p *structs.ListDomainParams) ([]*ent.Domain, error) {
	var nextDomain *ent.Domain
	if p.Cursor != "" {
		domain, err := r.ec.Domain.
			Query().
			Where(
				domainEnt.IDEQ(p.Cursor),
			).
			First(ctx)
		if err != nil || domain == nil {
			return nil, errors.New("invalid cursor")
		}
		nextDomain = domain
	}

	query := r.ec.Domain.
		Query().
		Limit(int(p.Limit))

	// is disabled
	query.Where(domainEnt.DisabledEQ(false))

	// lt the cursor create time
	if nextDomain != nil {
		query.Where(domainEnt.CreatedAtLT(nextDomain.CreatedAt))
	}

	// belong domain
	if p.UserID != "" {
		query.Where(domainEnt.UserIDEQ(p.UserID))
	}

	// sort
	query.Order(ent.Desc(domainEnt.FieldCreatedAt))

	rows, err := query.All(ctx)
	if err != nil {
		log.Errorf(nil, " domainRepo.GetDomainList error: %v\n", err)
		return nil, err
	}

	return rows, nil
}

// Delete delete domain
func (r *domainRepo) Delete(ctx context.Context, id string) error {
	_, err := r.ec.Domain.
		Delete().
		Where(domainEnt.IDEQ(id)).
		Exec(ctx)

	if err == nil {
		// Remove from cache
		cacheKey := fmt.Sprintf("%s", id)
		err := r.c.Delete(ctx, cacheKey)
		if err != nil {
			log.Errorf(nil, "domainRepo.Delete cache error: %v\n", err)
		}
	}

	return err
}

// DeleteByUser delete domain by user ID
func (r *domainRepo) DeleteByUser(ctx context.Context, userID string) error {
	_, err := r.ec.Domain.
		Delete().
		Where(domainEnt.UserIDEQ(userID)).
		Exec(ctx)

	if err == nil {
		// Remove from cache
		cacheUserKey := fmt.Sprintf("user:%s", userID)
		err = r.c.Delete(ctx, cacheUserKey)
		if err != nil {
			log.Errorf(nil, "domainRepo.DeleteByUser cache error: %v\n", err)
		}
	}

	return err
}
