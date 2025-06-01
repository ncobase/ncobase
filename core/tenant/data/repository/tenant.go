package repository

import (
	"context"
	"fmt"
	"ncobase/tenant/data"
	"ncobase/tenant/data/ent"
	tenantEnt "ncobase/tenant/data/ent/tenant"
	"ncobase/tenant/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/data/search/meili"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"
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
	ec               *ent.Client
	ms               *meili.Client
	tenantCache      cache.ICache[ent.Tenant]
	slugMappingCache cache.ICache[string] // Maps slug to tenant ID
	userMappingCache cache.ICache[string] // Maps user ID to tenant ID
	tenantTTL        time.Duration
}

// NewTenantRepository creates a new tenant repository.
func NewTenantRepository(d *data.Data) TenantRepositoryInterface {
	redisClient := d.GetRedis()

	return &tenantRepository{
		ec:               d.GetMasterEntClient(),
		ms:               d.GetMeilisearch(),
		tenantCache:      cache.NewCache[ent.Tenant](redisClient, "ncse_tenant:tenants"),
		slugMappingCache: cache.NewCache[string](redisClient, "ncse_tenant:slug_mappings"),
		userMappingCache: cache.NewCache[string](redisClient, "ncse_tenant:user_mappings"),
		tenantTTL:        time.Hour * 4, // 4 hours cache TTL
	}
}

// Create create tenant
func (r *tenantRepository) Create(ctx context.Context, body *structs.CreateTenantBody) (*ent.Tenant, error) {
	builder := r.ec.Tenant.Create()
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

	tenant, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantRepo.Create error: %v", err)
		return nil, err
	}

	// Create the tenant in Meilisearch index
	if err = r.ms.IndexDocuments("tenants", tenant); err != nil {
		logger.Errorf(ctx, "tenantRepo.Create error creating Meilisearch index: %v", err)
	}

	// Cache the tenant
	go r.cacheTenant(context.Background(), tenant)

	return tenant, nil
}

// GetBySlug get tenant by slug or id
func (r *tenantRepository) GetBySlug(ctx context.Context, slug string) (*ent.Tenant, error) {
	// Try to get tenant ID from slug mapping cache
	if tenantID, err := r.getTenantIDBySlug(ctx, slug); err == nil && tenantID != "" {
		// Try to get from tenant cache
		cacheKey := fmt.Sprintf("id:%s", tenantID)
		if cached, err := r.tenantCache.Get(ctx, cacheKey); err == nil && cached != nil {
			return cached, nil
		}
	}

	// Fallback to database
	row, err := r.FindTenant(ctx, &structs.FindTenant{Slug: slug})
	if err != nil {
		logger.Errorf(ctx, "tenantRepo.GetBySlug error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheTenant(context.Background(), row)

	return row, nil
}

// GetByUser get tenant by user
func (r *tenantRepository) GetByUser(ctx context.Context, userID string) (*ent.Tenant, error) {
	// Try to get tenant ID from user mapping cache
	if tenantID, err := r.getTenantIDByUser(ctx, userID); err == nil && tenantID != "" {
		// Try to get from tenant cache
		cacheKey := fmt.Sprintf("id:%s", tenantID)
		if cached, err := r.tenantCache.Get(ctx, cacheKey); err == nil && cached != nil {
			return cached, nil
		}
	}

	// Fallback to database
	row, err := r.FindTenant(ctx, &structs.FindTenant{User: userID})
	if err != nil {
		logger.Errorf(ctx, "tenantRepo.GetByUser error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheTenant(context.Background(), row)

	return row, nil
}

// GetIDByUser get tenant id by user id
func (r *tenantRepository) GetIDByUser(ctx context.Context, userID string) (string, error) {
	// Try cache first
	if tenantID, err := r.getTenantIDByUser(ctx, userID); err == nil && tenantID != "" {
		return tenantID, nil
	}

	// Fallback to database
	id, err := r.ec.Tenant.
		Query().
		Where(tenantEnt.CreatedByEQ(userID)).
		OnlyID(ctx)

	if err != nil {
		logger.Errorf(ctx, "tenantRepo.GetIDByUser error: %v", err)
		return "", err
	}

	// Cache user to tenant mapping
	go func() {
		userKey := fmt.Sprintf("user:%s", userID)
		if err := r.userMappingCache.Set(context.Background(), userKey, &id, r.tenantTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache user mapping %s: %v", userID, err)
		}
	}()

	return id, nil
}

// Update update tenant
func (r *tenantRepository) Update(ctx context.Context, slug string, updates types.JSON) (*ent.Tenant, error) {
	tenant, err := r.FindTenant(ctx, &structs.FindTenant{Slug: slug})
	if err != nil {
		return nil, err
	}

	builder := tenant.Update()

	// Set values as in original implementation
	for field, value := range updates {
		switch field {
		case "name":
			builder.SetNillableName(convert.ToPointer(value.(string)))
		case "slug":
			builder.SetNillableSlug(convert.ToPointer(value.(string)))
		case "type":
			builder.SetNillableType(convert.ToPointer(value.(string)))
		case "title":
			builder.SetNillableTitle(convert.ToPointer(value.(string)))
		case "url":
			builder.SetNillableURL(convert.ToPointer(value.(string)))
		case "logo":
			builder.SetNillableLogo(convert.ToPointer(value.(string)))
		case "logo_alt":
			builder.SetNillableLogoAlt(convert.ToPointer(value.(string)))
		case "keywords":
			builder.SetNillableKeywords(convert.ToPointer(value.(string)))
		case "copyright":
			builder.SetNillableCopyright(convert.ToPointer(value.(string)))
		case "description":
			builder.SetNillableDescription(convert.ToPointer(value.(string)))
		case "disabled":
			builder.SetDisabled(value.(bool))
		case "order":
			builder.SetOrder(int(value.(float64)))
		case "extras":
			builder.SetExtras(value.(types.JSON))
		case "updated_by":
			builder.SetNillableUpdatedBy(convert.ToPointer(value.(string)))
		case "expired_at":
			adjustedTime, _ := convert.AdjustToEndOfDay(value)
			builder.SetNillableExpiredAt(&adjustedTime)
		}
	}

	updatedTenant, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantRepo.Update error: %v", err)
		return nil, err
	}

	// Update Meilisearch index
	if err = r.ms.UpdateDocuments("tenants", updatedTenant, updatedTenant.ID); err != nil {
		logger.Errorf(ctx, "tenantRepo.Update error updating Meilisearch index: %v", err)
	}

	// Invalidate and re-cache
	go func() {
		r.invalidateTenantCache(context.Background(), tenant)
		r.cacheTenant(context.Background(), updatedTenant)
	}()

	return updatedTenant, nil
}

// List get tenant list
func (r *tenantRepository) List(ctx context.Context, params *structs.ListTenantParams) ([]*ent.Tenant, error) {
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return nil, err
	}

	// Is disabled
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
		logger.Errorf(ctx, "tenantRepo.List error: %v", err)
		return nil, err
	}

	// Cache tenants in background
	go func() {
		for _, tenant := range rows {
			r.cacheTenant(context.Background(), tenant)
		}
	}()

	return rows, nil
}

// Delete delete tenant
func (r *tenantRepository) Delete(ctx context.Context, id string) error {
	tenant, err := r.FindTenant(ctx, &structs.FindTenant{Slug: id})
	if err != nil {
		return err
	}

	builder := r.ec.Tenant.Delete()
	if _, err = builder.Where(tenantEnt.IDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "tenantRepo.Delete error: %v", err)
		return err
	}

	// Delete from Meilisearch index
	if err = r.ms.DeleteDocuments("tenants", tenant.ID); err != nil {
		logger.Errorf(ctx, "tenantRepo.Delete index error: %v", err)
	}

	// Invalidate cache
	go r.invalidateTenantCache(context.Background(), tenant)

	return nil
}

// DeleteByUser delete tenant by user ID
func (r *tenantRepository) DeleteByUser(ctx context.Context, userID string) error {
	// Get tenant first for cache invalidation
	tenant, err := r.GetByUser(ctx, userID)
	if err != nil {
		return err
	}

	builder := r.ec.Tenant.Delete()
	if _, err := builder.Where(tenantEnt.CreatedByEQ(userID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "tenantRepo.DeleteByUser error: %v", err)
		return err
	}

	// Delete from Meilisearch index
	if err = r.ms.DeleteDocuments("tenants", tenant.ID); err != nil {
		logger.Errorf(ctx, "tenantRepo.DeleteByUser index error: %v", err)
	}

	// Invalidate cache
	go r.invalidateTenantCache(context.Background(), tenant)

	return nil
}

// CountX gets a count of tenants
func (r *tenantRepository) CountX(ctx context.Context, params *structs.ListTenantParams) int {
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// FindTenant retrieves a tenant
func (r *tenantRepository) FindTenant(ctx context.Context, params *structs.FindTenant) (*ent.Tenant, error) {
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

	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// listBuilder - create list builder
func (r *tenantRepository) listBuilder(_ context.Context, params *structs.ListTenantParams) (*ent.TenantQuery, error) {
	builder := r.ec.Tenant.Query()

	// Match belong user
	if validator.IsNotEmpty(params.User) {
		builder.Where(tenantEnt.CreatedByEQ(params.User))
	}

	return builder, nil
}

func (r *tenantRepository) cacheTenant(ctx context.Context, tenant *ent.Tenant) {
	// Cache by ID
	idKey := fmt.Sprintf("id:%s", tenant.ID)
	if err := r.tenantCache.Set(ctx, idKey, tenant, r.tenantTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache tenant by ID %s: %v", tenant.ID, err)
	}

	// Cache slug to ID mapping
	if tenant.Slug != "" {
		slugKey := fmt.Sprintf("slug:%s", tenant.Slug)
		if err := r.slugMappingCache.Set(ctx, slugKey, &tenant.ID, r.tenantTTL); err != nil {
			logger.Debugf(ctx, "Failed to cache slug mapping %s: %v", tenant.Slug, err)
		}
	}

	// Cache user to tenant ID mapping
	if tenant.CreatedBy != "" {
		userKey := fmt.Sprintf("user:%s", tenant.CreatedBy)
		if err := r.userMappingCache.Set(ctx, userKey, &tenant.ID, r.tenantTTL); err != nil {
			logger.Debugf(ctx, "Failed to cache user mapping %s: %v", tenant.CreatedBy, err)
		}
	}
}

func (r *tenantRepository) invalidateTenantCache(ctx context.Context, tenant *ent.Tenant) {
	// Invalidate ID cache
	idKey := fmt.Sprintf("id:%s", tenant.ID)
	if err := r.tenantCache.Delete(ctx, idKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate tenant ID cache %s: %v", tenant.ID, err)
	}

	// Invalidate slug mapping
	if tenant.Slug != "" {
		slugKey := fmt.Sprintf("slug:%s", tenant.Slug)
		if err := r.slugMappingCache.Delete(ctx, slugKey); err != nil {
			logger.Debugf(ctx, "Failed to invalidate slug mapping cache %s: %v", tenant.Slug, err)
		}
	}

	// Invalidate user mapping
	if tenant.CreatedBy != "" {
		userKey := fmt.Sprintf("user:%s", tenant.CreatedBy)
		if err := r.userMappingCache.Delete(ctx, userKey); err != nil {
			logger.Debugf(ctx, "Failed to invalidate user mapping cache %s: %v", tenant.CreatedBy, err)
		}
	}
}

func (r *tenantRepository) getTenantIDBySlug(ctx context.Context, slug string) (string, error) {
	cacheKey := fmt.Sprintf("slug:%s", slug)
	tenantID, err := r.slugMappingCache.Get(ctx, cacheKey)
	if err != nil || tenantID == nil {
		return "", err
	}
	return *tenantID, nil
}

func (r *tenantRepository) getTenantIDByUser(ctx context.Context, userID string) (string, error) {
	cacheKey := fmt.Sprintf("user:%s", userID)
	tenantID, err := r.userMappingCache.Get(ctx, cacheKey)
	if err != nil || tenantID == nil {
		return "", err
	}
	return *tenantID, nil
}
