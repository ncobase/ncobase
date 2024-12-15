package repository

import (
	"context"
	"fmt"
	"ncobase/common/data/cache"
	"ncobase/common/data/meili"
	"ncobase/common/logger"
	"ncobase/common/nanoid"
	"ncobase/common/paging"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/core/tenant/data"
	"ncobase/core/tenant/data/ent"
	tenantEnt "ncobase/core/tenant/data/ent/tenant"
	"ncobase/core/tenant/structs"

	"github.com/redis/go-redis/v9"
)

// TenantRepositoryInterface represents the tenant repository interface.
type TenantRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateTenantBody) (*ent.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*ent.Tenant, error)
	GetByUser(ctx context.Context, user string) (*ent.Tenant, error)
	GetIDByUser(ctx context.Context, user string) (string, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*ent.Tenant, error)
	List(ctx context.Context, params *structs.ListTenantParams) ([]*ent.Tenant, error)
	Delete(ctx context.Context, id string) error
	DeleteByUser(ctx context.Context, id string) error
	CountX(ctx context.Context, params *structs.ListTenantParams) int
}

// tenantRepository implements the TenantRepositoryInterface.
type tenantRepository struct {
	ec *ent.Client
	rc *redis.Client
	ms *meili.Client
	c  *cache.Cache[ent.Tenant]
}

// NewTenantRepository creates a new tenant repository.
func NewTenantRepository(d *data.Data) TenantRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &tenantRepository{ec, rc, ms, cache.NewCache[ent.Tenant](rc, "ncse_tenant")}
}

// Create create tenant
func (r *tenantRepository) Create(ctx context.Context, body *structs.CreateTenantBody) (*ent.Tenant, error) {
	// create builder.
	builder := r.ec.Tenant.Create()
	// set values.
	builder.SetNillableName(&body.Name)
	builder.SetNillableSlug(&body.Slug)
	builder.SetNillableType(&body.Type)
	builder.SetNillableTitle(&body.Title)
	builder.SetNillableURL(&body.URL)
	builder.SetNillableLogo(&body.Logo)
	builder.SetNillableLogoAlt(&body.LogoAlt)
	builder.SetNillableKeywords(&body.Keywords)
	builder.SetNillableCopyright(&body.Copyright)
	builder.SetNillableDescription(&body.Description)
	builder.SetDisabled(body.Disabled)
	builder.SetNillableCreatedBy(body.CreatedBy)
	builder.SetNillableExpiredAt(body.ExpiredAt)

	if !validator.IsNil(body.Order) {
		builder.SetNillableOrder(body.Order)
	}

	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantRepo.Create error: %v", err)
		return nil, err
	}

	// Create the tenant in Meilisearch index
	if err = r.ms.IndexDocuments("tenants", row); err != nil {
		logger.Errorf(ctx, "tenantRepo.Create error creating Meilisearch index: %v", err)
		// return nil, err
	}

	return row, nil
}

// GetBySlug get tenant by slug or id
func (r *tenantRepository) GetBySlug(ctx context.Context, id string) (*ent.Tenant, error) {
	// // Search in Meilisearch index
	// if res, _ := r.ms.Search(ctx, "taxonomies", id, &meilisearch.SearchRequest{Limit: 1}); res != nil && res.Hits != nil && len(res.Hits) > 0 {
	// 	if hit, ok := res.Hits[0].(*ent.Tenant); ok {
	// 		return hit, nil
	// 	}
	// }
	// check cache
	cacheKey := fmt.Sprintf("%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindTenant(ctx, &structs.FindTenant{Slug: id})

	if err != nil {
		logger.Errorf(ctx, " tenantRepo.GetByID error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "tenantRepo.GetByID cache error: %v", err)
	}

	return row, nil
}

// GetByUser get tenant by user
func (r *tenantRepository) GetByUser(ctx context.Context, userID string) (*ent.Tenant, error) {
	// // Search in Meilisearch index
	// if res, _ := r.ms.Search(ctx, "taxonomies", userID, &meilisearch.SearchRequest{Limit: 1}); res != nil && res.Hits != nil && len(res.Hits) > 0 {
	// 	if hit, ok := res.Hits[0].(*ent.Tenant); ok {
	// 		return hit, nil
	// 	}
	// }
	// check cache
	cacheKey := fmt.Sprintf("%s", userID)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindTenant(ctx, &structs.FindTenant{User: userID})

	if err != nil {
		logger.Errorf(ctx, " tenantRepo.GetByUser error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "tenantRepo.GetByUser cache error: %v", err)
	}

	return row, nil
}

// GetIDByUser get tenant id by user id
func (r *tenantRepository) GetIDByUser(ctx context.Context, userID string) (string, error) {
	id, err := r.ec.Tenant.
		Query().
		Where(tenantEnt.CreatedByEQ(userID)).
		OnlyID(ctx)

	if err != nil {
		logger.Errorf(ctx, " tenantRepo.FindTenantIDByUserID error: %v", err)
		return "", err
	}

	return id, nil
}

// Update update tenant
func (r *tenantRepository) Update(ctx context.Context, slug string, updates types.JSON) (*ent.Tenant, error) {
	tenant, err := r.FindTenant(ctx, &structs.FindTenant{Slug: slug})
	if err != nil {
		return nil, err
	}

	// create builder.
	builder := tenant.Update()

	// set values
	for field, value := range updates {
		switch field {
		case "name":
			builder.SetNillableName(types.ToPointer(value.(string)))
		case "slug":
			builder.SetNillableSlug(types.ToPointer(value.(string)))
		case "type":
			builder.SetNillableType(types.ToPointer(value.(string)))
		case "title":
			builder.SetNillableTitle(types.ToPointer(value.(string)))
		case "url":
			builder.SetNillableURL(types.ToPointer(value.(string)))
		case "logo":
			builder.SetNillableLogo(types.ToPointer(value.(string)))
		case "logo_alt":
			builder.SetNillableLogoAlt(types.ToPointer(value.(string)))
		case "keywords":
			builder.SetNillableKeywords(types.ToPointer(value.(string)))
		case "copyright":
			builder.SetNillableCopyright(types.ToPointer(value.(string)))
		case "description":
			builder.SetNillableDescription(types.ToPointer(value.(string)))
		case "disabled":
			builder.SetDisabled(value.(bool))
		case "order":
			builder.SetOrder(int(value.(float64)))
		case "extras":
			builder.SetExtras(value.(types.JSON))
		case "updated_by":
			builder.SetNillableUpdatedBy(types.ToPointer(value.(string)))
		case "expired_at":
			adjustedTime, _ := types.AdjustToEndOfDay(value)
			builder.SetNillableExpiredAt(&adjustedTime)
		}
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantRepo.Update error: %v", err)
		return nil, err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", tenant.ID)
	err = r.c.Delete(ctx, cacheKey)
	if err != nil {
		logger.Errorf(ctx, "tenantRepo.Update cache error: %v", err)
	}
	cacheUserKey := fmt.Sprintf("user:%s", tenant.CreatedBy)
	err = r.c.Delete(ctx, cacheUserKey)
	if err != nil {
		logger.Errorf(ctx, "tenantRepo.Update cache error: %v", err)
	}

	// Update Meilisearch index
	if err = r.ms.UpdateDocuments("topics", row, row.ID); err != nil {
		logger.Errorf(ctx, "tenantRepo.Update error updating Meilisearch index: %v", err)
	}

	return row, nil
}

// List get
func (r *tenantRepository) List(ctx context.Context, params *structs.ListTenantParams) ([]*ent.Tenant, error) {
	// create list builder
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return nil, err
	}

	// is disabled
	builder.Where(tenantEnt.DisabledEQ(false))

	if params.Cursor != "" {
		id, timestamp, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %v", err)
		}

		if !nanoid.IsPrimaryKey(id) {
			return nil, fmt.Errorf("invalid id in cursor: %s", id)
		}

		if params.Direction == "backward" {
			builder.Where(
				tenantEnt.Or(
					tenantEnt.CreatedAtGT(timestamp),
					tenantEnt.And(
						tenantEnt.CreatedAtEQ(timestamp),
						tenantEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				tenantEnt.Or(
					tenantEnt.CreatedAtLT(timestamp),
					tenantEnt.And(
						tenantEnt.CreatedAtEQ(timestamp),
						tenantEnt.IDLT(id),
					),
				),
			)
		}
	}

	if params.Direction == "backward" {
		builder.Order(ent.Asc(tenantEnt.FieldCreatedAt), ent.Asc(tenantEnt.FieldID))
	} else {
		builder.Order(ent.Desc(tenantEnt.FieldCreatedAt), ent.Desc(tenantEnt.FieldID))
	}

	builder.Limit(params.Limit)

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, " tenantRepo.GetTenantList error: %v", err)
		return nil, err
	}

	return rows, nil
}

// Delete delete tenant
func (r *tenantRepository) Delete(ctx context.Context, id string) error {
	tenant, err := r.FindTenant(ctx, &structs.FindTenant{Slug: id})
	if err != nil {
		return err
	}

	// create builder.
	builder := r.ec.Tenant.Delete()

	// execute the builder and verify the result.
	if _, err = builder.Where(tenantEnt.IDEQ(id)).Exec(ctx); err == nil {
		logger.Errorf(ctx, "tenantRepo.Delete error: %v", err)
		return err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", tenant.ID)
	err = r.c.Delete(ctx, cacheKey)
	if err != nil {
		logger.Errorf(ctx, "tenantRepo.Delete cache error: %v", err)
	}

	// delete from Meilisearch index
	if err = r.ms.DeleteDocuments("tenants", tenant.ID); err != nil {
		logger.Errorf(ctx, "tenantRepo.Delete index error: %v", err)
		// return nil, err
	}

	return err
}

// DeleteByUser delete tenant by user ID
func (r *tenantRepository) DeleteByUser(ctx context.Context, userID string) error {

	// create builder.
	builder := r.ec.Tenant.Delete()

	if _, err := builder.Where(tenantEnt.CreatedByEQ(userID)).Exec(ctx); err == nil {
		logger.Errorf(ctx, "tenantRepo.DeleteByUser error: %v", err)
		return err
	}

	// remove from cache
	cacheUserKey := fmt.Sprintf("user:%s", userID)
	err := r.c.Delete(ctx, cacheUserKey)
	if err != nil {
		logger.Errorf(ctx, "tenantRepo.DeleteByUser cache error: %v", err)
	}

	// delete from Meilisearch index
	if err = r.ms.DeleteDocuments("tenants", userID); err != nil {
		logger.Errorf(ctx, "tenantRepo.DeleteByUser index error: %v", err)
		// return nil, err
	}

	return err
}

// CountX gets a count of tenants.
func (r *tenantRepository) CountX(ctx context.Context, params *structs.ListTenantParams) int {
	// create list builder
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// FindTenant retrieves a tenant.
func (r *tenantRepository) FindTenant(ctx context.Context, params *structs.FindTenant) (*ent.Tenant, error) {

	// create builder.
	builder := r.ec.Tenant.Query()

	if validator.IsNotEmpty(params.Slug) {
		builder = builder.Where(tenantEnt.Or(
			tenantEnt.IDEQ(params.Slug),
			tenantEnt.SlugEQ(params.Slug),
		))
	}
	if validator.IsNotEmpty(params.User) {
		builder = builder.Where(tenantEnt.CreatedByEQ(params.User))
	}

	// execute the builder.
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// listBuilder - create list builder.
// internal method.
func (r *tenantRepository) listBuilder(_ context.Context, params *structs.ListTenantParams) (*ent.TenantQuery, error) {
	// create builder.
	builder := r.ec.Tenant.Query()

	// match belong user
	if validator.IsNotEmpty(params.User) {
		builder.Where(tenantEnt.CreatedByEQ(params.User))
	}

	return builder, nil
}
