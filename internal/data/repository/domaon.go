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
	"stocms/pkg/types"
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
	Update(ctx context.Context, slug string, updates types.JSON) (*ent.Domain, error)
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
	// create builder.
	builder := r.ec.Domain.Create()
	// Set values
	builder.SetNillableName(&body.Name)
	builder.SetNillableTitle(&body.Title)
	builder.SetNillableURL(&body.URL)
	builder.SetNillableLogo(&body.Logo)
	builder.SetNillableLogoAlt(&body.LogoAlt)
	builder.SetKeywords(strings.Join(body.Keywords, ","))
	builder.SetNillableCopyright(&body.Copyright)
	builder.SetNillableDescription(&body.Description)
	builder.SetDisabled(body.Disabled)
	builder.SetUserID(body.UserID)

	if body.Order > 0 {
		builder.SetOrder(int32(body.Order))
	}

	if !validator.IsNil(body.Extras) && len(body.Extras) > 0 {
		builder.SetExtras(body.Extras)
	}

	// execute the builder.
	row, err := builder.Save(ctx)
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
	row, err := r.FindDomain(ctx, &structs.FindDomain{ID: id})

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
	row, err := r.FindDomain(ctx, &structs.FindDomain{UserID: userID})

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
func (r *domainRepo) Update(ctx context.Context, id string, updates types.JSON) (*ent.Domain, error) {
	domain, err := r.FindDomain(ctx, &structs.FindDomain{ID: id})
	if err != nil {
		return nil, err
	}

	// create builder.
	builder := domain.Update()

	// Set values
	for field, value := range updates {
		switch field {
		case "name":
			builder.SetNillableName(types.ToPointer(value.(string)))
		case "title":
			builder.SetNillableTitle(types.ToPointer(value.(string)))
		case "url":
			builder.SetNillableURL(types.ToPointer(value.(string)))
		case "logo":
			builder.SetNillableLogo(types.ToPointer(value.(string)))
		case "logo_alt":
			builder.SetNillableLogoAlt(types.ToPointer(value.(string)))
		case "keywords":
			builder.SetKeywords(strings.Join(value.([]string), ","))
		case "copyright":
			builder.SetNillableCopyright(types.ToPointer(value.(string)))
		case "description":
			builder.SetNillableDescription(types.ToPointer(value.(string)))
		case "disabled":
			builder.SetDisabled(value.(bool))
		case "order":
			builder.SetOrder(int32(value.(float64)))
		case "extras":
			builder.SetExtras(value.(types.JSON))
		case "user_id":
			builder.SetNillableUserID(types.ToPointer(value.(string)))
		}
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(nil, "domainRepo.Update error: %v\n", err)
		return nil, err
	}

	// Remove from cache
	cacheKey := fmt.Sprintf("%s", domain.ID)
	err = r.c.Delete(ctx, cacheKey)
	if err != nil {
		log.Errorf(nil, "domainRepo.Update cache error: %v\n", err)
	}
	cacheUserKey := fmt.Sprintf("user:%s", domain.UserID)
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
	domain, err := r.FindDomain(ctx, &structs.FindDomain{ID: id})
	if err != nil {
		return err
	}
	// create builder.
	builder := r.ec.Domain.Delete()

	// execute the builder.
	_, err = builder.Where(domainEnt.IDEQ(id)).Exec(ctx)
	if err == nil {
		// Remove from cache
		cacheKey := fmt.Sprintf("%s", domain.ID)
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

func (r *domainRepo) FindDomain(ctx context.Context, p *structs.FindDomain) (*ent.Domain, error) {
	// create builder.
	builder := r.ec.Domain.Query()

	if validator.IsNotEmpty(p.ID) {
		builder = builder.Where(domainEnt.IDEQ(p.ID))
	}
	if validator.IsNotEmpty(p.UserID) {
		builder = builder.Where(domainEnt.UserIDEQ(p.UserID))
	}

	// execute the builder.
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// ****** Internal methods of repository
