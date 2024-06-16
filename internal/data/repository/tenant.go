package repo

import (
	"context"
	"errors"
	"fmt"
	"ncobase/common/cache"
	"ncobase/common/log"
	"ncobase/common/meili"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/internal/data"
	"ncobase/internal/data/ent"
	groupEnt "ncobase/internal/data/ent/group"
	tenantEnt "ncobase/internal/data/ent/tenant"
	"ncobase/internal/data/structs"
	"strings"

	"github.com/redis/go-redis/v9"
)

// Tenant represents the tenant repository interface.
type Tenant interface {
	Create(ctx context.Context, body *structs.CreateTenantBody) (*ent.Tenant, error)
	GetByID(ctx context.Context, id string) (*ent.Tenant, error)
	GetByUser(ctx context.Context, user string) (*ent.Tenant, error)
	GetIDByUser(ctx context.Context, user string) (string, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*ent.Tenant, error)
	List(ctx context.Context, params *structs.ListTenantParams) ([]*ent.Tenant, error)
	Delete(ctx context.Context, id string) error
	DeleteByUser(ctx context.Context, id string) error
	GetGroupsByTenantID(ctx context.Context, tenantID string) ([]*ent.Group, error)
	IsGroupInTenant(ctx context.Context, groupID string, tenantID string) (bool, error)
}

// tenantRepo implements the Tenant interface.
type tenantRepo struct {
	ec *ent.Client
	rc *redis.Client
	ms *meili.Client
	c  *cache.Cache[ent.Tenant]
}

// NewTenant creates a new tenant repository.
func NewTenant(d *data.Data) Tenant {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &tenantRepo{ec, rc, ms, cache.NewCache[ent.Tenant](rc, cache.Key("nb_tenant"), true)}
}

// Create create tenant
func (r *tenantRepo) Create(ctx context.Context, body *structs.CreateTenantBody) (*ent.Tenant, error) {
	// create builder.
	builder := r.ec.Tenant.Create()
	// set values.
	builder.SetNillableName(&body.Name)
	builder.SetNillableTitle(&body.Title)
	builder.SetNillableURL(&body.URL)
	builder.SetNillableLogo(&body.Logo)
	builder.SetNillableLogoAlt(&body.LogoAlt)
	builder.SetKeywords(strings.Join(body.Keywords, ","))
	builder.SetNillableCopyright(&body.Copyright)
	builder.SetNillableDescription(&body.Description)
	builder.SetDisabled(body.Disabled)
	builder.SetNillableCreatedBy(&body.CreatedBy)

	if !validator.IsNil(body.Order) {
		builder.SetNillableOrder(body.Order)
	}

	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(context.Background(), "tenantRepo.Create error: %v\n", err)
		return nil, err
	}

	// Create the tenant in Meilisearch index
	if err = r.ms.IndexDocuments("tenants", row); err != nil {
		log.Errorf(context.Background(), "tenantRepo.Create error creating Meilisearch index: %v\n", err)
		// return nil, err
	}

	return row, nil
}

// GetByID get tenant by id
func (r *tenantRepo) GetByID(ctx context.Context, id string) (*ent.Tenant, error) {
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
	row, err := r.FindTenant(ctx, &structs.FindTenant{ID: id})

	if err != nil {
		log.Errorf(context.Background(), " tenantRepo.GetByID error: %v\n", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(context.Background(), "tenantRepo.GetByID cache error: %v\n", err)
	}

	return row, nil
}

// GetByUser get tenant by user
func (r *tenantRepo) GetByUser(ctx context.Context, userID string) (*ent.Tenant, error) {
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
		log.Errorf(context.Background(), " tenantRepo.GetByUser error: %v\n", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(context.Background(), "tenantRepo.GetByUser cache error: %v\n", err)
	}

	return row, nil
}

// GetIDByUser get tenant id by user id
func (r *tenantRepo) GetIDByUser(ctx context.Context, userID string) (string, error) {
	id, err := r.ec.Tenant.
		Query().
		Where(tenantEnt.CreatedByEQ(userID)).
		OnlyID(ctx)

	if err != nil {
		log.Errorf(context.Background(), " tenantRepo.FindTenantIDByUserID error: %v\n", err)
		return "", err
	}

	return id, nil
}

// Update update tenant
func (r *tenantRepo) Update(ctx context.Context, id string, updates types.JSON) (*ent.Tenant, error) {
	tenant, err := r.FindTenant(ctx, &structs.FindTenant{ID: id})
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
			builder.SetOrder(int(value.(float64)))
		case "extras":
			builder.SetExtras(value.(types.JSON))
		case "user_id":
			builder.SetNillableCreatedBy(types.ToPointer(value.(string)))
		}
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(context.Background(), "tenantRepo.Update error: %v\n", err)
		return nil, err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", tenant.ID)
	err = r.c.Delete(ctx, cacheKey)
	if err != nil {
		log.Errorf(context.Background(), "tenantRepo.Update cache error: %v\n", err)
	}
	cacheUserKey := fmt.Sprintf("user:%s", tenant.CreatedBy)
	err = r.c.Delete(ctx, cacheUserKey)
	if err != nil {
		log.Errorf(context.Background(), "tenantRepo.Update cache error: %v\n", err)
	}

	// doc := types.JSON{}
	// if err = copier.Copy(doc, row); err != nil {
	// 	log.Errorf(context.Background(), "tenantRepo.Update error copying data: %v\n", err)
	// 	// return nil, err
	// }
	if err = r.ms.UpdateDocuments("topics", row, row.ID); err != nil {
		log.Errorf(context.Background(), "tenantRepo.Update error updating Meilisearch index: %v\n", err)
	}

	return row, nil
}

// List get
func (r *tenantRepo) List(ctx context.Context, p *structs.ListTenantParams) ([]*ent.Tenant, error) {
	var nextTenant *ent.Tenant
	if p.Cursor != "" {
		tenant, err := r.ec.Tenant.
			Query().
			Where(
				tenantEnt.IDEQ(p.Cursor),
			).
			First(ctx)
		if err != nil || tenant == nil {
			return nil, errors.New("invalid cursor")
		}
		nextTenant = tenant
	}

	query := r.ec.Tenant.
		Query().
		Limit(int(p.Limit))

	// is disabled
	query.Where(tenantEnt.DisabledEQ(false))

	// lt the cursor create time
	if nextTenant != nil {
		query.Where(tenantEnt.CreatedAtLT(nextTenant.CreatedAt))
	}

	// belong tenant
	if p.User != "" {
		query.Where(tenantEnt.CreatedByEQ(p.User))
	}

	// sort
	query.Order(ent.Desc(tenantEnt.FieldCreatedAt))

	rows, err := query.All(ctx)
	if err != nil {
		log.Errorf(context.Background(), " tenantRepo.GetTenantList error: %v\n", err)
		return nil, err
	}

	return rows, nil
}

// Delete delete tenant
func (r *tenantRepo) Delete(ctx context.Context, id string) error {
	tenant, err := r.FindTenant(ctx, &structs.FindTenant{ID: id})
	if err != nil {
		return err
	}

	// create builder.
	builder := r.ec.Tenant.Delete()

	// execute the builder and verify the result.
	if _, err = builder.Where(tenantEnt.IDEQ(id)).Exec(ctx); err == nil {
		log.Errorf(context.Background(), "tenantRepo.Delete error: %v\n", err)
		return err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", tenant.ID)
	err = r.c.Delete(ctx, cacheKey)
	if err != nil {
		log.Errorf(context.Background(), "tenantRepo.Delete cache error: %v\n", err)
	}

	// delete from Meilisearch index
	if err = r.ms.DeleteDocuments("tenants", tenant.ID); err != nil {
		log.Errorf(context.Background(), "tenantRepo.Delete index error: %v\n", err)
		// return nil, err
	}

	return err
}

// DeleteByUser delete tenant by user ID
func (r *tenantRepo) DeleteByUser(ctx context.Context, userID string) error {

	// create builder.
	builder := r.ec.Tenant.Delete()

	if _, err := builder.Where(tenantEnt.CreatedByEQ(userID)).Exec(ctx); err == nil {
		log.Errorf(context.Background(), "tenantRepo.DeleteByUser error: %v\n", err)
		return err
	}

	// remove from cache
	cacheUserKey := fmt.Sprintf("user:%s", userID)
	err := r.c.Delete(ctx, cacheUserKey)
	if err != nil {
		log.Errorf(context.Background(), "tenantRepo.DeleteByUser cache error: %v\n", err)
	}

	// delete from Meilisearch index
	if err = r.ms.DeleteDocuments("tenants", userID); err != nil {
		log.Errorf(context.Background(), "tenantRepo.DeleteByUser index error: %v\n", err)
		// return nil, err
	}

	return err
}

// FindTenant retrieves a tenant.
func (r *tenantRepo) FindTenant(ctx context.Context, p *structs.FindTenant) (*ent.Tenant, error) {

	// create builder.
	builder := r.ec.Tenant.Query()

	if validator.IsNotEmpty(p.ID) {
		builder = builder.Where(tenantEnt.IDEQ(p.ID))
	}
	if validator.IsNotEmpty(p.User) {
		builder = builder.Where(tenantEnt.CreatedByEQ(p.User))
	}

	// execute the builder.
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// GetGroupsByTenantID retrieves all groups under a tenant.
func (r *tenantRepo) GetGroupsByTenantID(ctx context.Context, tenantID string) ([]*ent.Group, error) {
	groups, err := r.ec.Group.Query().Where(groupEnt.TenantIDEQ(tenantID)).All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "tenantRepo.GetGroupsByTenantID error: %v\n", err)
		return nil, err
	}
	return groups, nil
}

// IsGroupInTenant verifies if a group belongs to a specific tenant.
func (r *tenantRepo) IsGroupInTenant(ctx context.Context, tenantID string, groupID string) (bool, error) {
	count, err := r.ec.Group.Query().Where(groupEnt.TenantIDEQ(tenantID), groupEnt.IDEQ(groupID)).Count(ctx)
	if err != nil {
		log.Errorf(context.Background(), "tenantRepo.IsGroupInTenant error: %v\n", err)
		return false, err
	}
	return count > 0, nil
}
