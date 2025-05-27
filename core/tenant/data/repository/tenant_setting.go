package repository

import (
	"context"
	"fmt"
	"ncobase/tenant/data"
	"ncobase/tenant/data/ent"
	tenantSettingEnt "ncobase/tenant/data/ent/tenantsetting"
	"ncobase/tenant/structs"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/redis/go-redis/v9"
)

// TenantSettingRepositoryInterface defines the interface for tenant setting repository
type TenantSettingRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateTenantSettingBody) (*ent.TenantSetting, error)
	GetByID(ctx context.Context, id string) (*ent.TenantSetting, error)
	GetByKey(ctx context.Context, tenantID, key string) (*ent.TenantSetting, error)
	Update(ctx context.Context, id string, updates types.JSON) (*ent.TenantSetting, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params *structs.ListTenantSettingParams) ([]*ent.TenantSetting, error)
	ListWithCount(ctx context.Context, params *structs.ListTenantSettingParams) ([]*ent.TenantSetting, int, error)
}

// tenantSettingRepository implements TenantSettingRepositoryInterface
type tenantSettingRepository struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.TenantSetting]
}

// NewTenantSettingRepository creates a new tenant setting repository
func NewTenantSettingRepository(d *data.Data) TenantSettingRepositoryInterface {
	ec := d.GetMasterEntClient()
	rc := d.GetRedis()
	return &tenantSettingRepository{ec, rc, cache.NewCache[ent.TenantSetting](rc, "ncse_tenant_setting")}
}

// Create creates a new tenant setting
func (r *tenantSettingRepository) Create(ctx context.Context, body *structs.CreateTenantSettingBody) (*ent.TenantSetting, error) {
	builder := r.ec.TenantSetting.Create()

	builder.SetTenantID(body.TenantID)
	builder.SetSettingKey(body.SettingKey)
	builder.SetSettingName(body.SettingName)
	builder.SetSettingValue(body.SettingValue)
	builder.SetDefaultValue(body.DefaultValue)
	builder.SetSettingType(string(body.SettingType))
	builder.SetScope(string(body.Scope))
	builder.SetCategory(body.Category)
	builder.SetDescription(body.Description)
	builder.SetIsPublic(body.IsPublic)
	builder.SetIsRequired(body.IsRequired)
	builder.SetIsReadonly(body.IsReadonly)
	builder.SetNillableCreatedBy(body.CreatedBy)

	if !validator.IsNil(body.Validation) && !validator.IsEmpty(body.Validation) {
		builder.SetValidation(*body.Validation)
	}

	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantSettingRepo.Create error: %v", err)
		return nil, err
	}

	return row, nil
}

// GetByID retrieves a tenant setting by ID
func (r *tenantSettingRepository) GetByID(ctx context.Context, id string) (*ent.TenantSetting, error) {
	cacheKey := fmt.Sprintf("id:%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	row, err := r.ec.TenantSetting.Get(ctx, id)
	if err != nil {
		logger.Errorf(ctx, "tenantSettingRepo.GetByID error: %v", err)
		return nil, err
	}

	_ = r.c.Set(ctx, cacheKey, row)
	return row, nil
}

// GetByKey retrieves a tenant setting by tenant ID and key
func (r *tenantSettingRepository) GetByKey(ctx context.Context, tenantID, key string) (*ent.TenantSetting, error) {
	cacheKey := fmt.Sprintf("tenant:%s:key:%s", tenantID, key)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	row, err := r.ec.TenantSetting.Query().
		Where(
			tenantSettingEnt.TenantIDEQ(tenantID),
			tenantSettingEnt.SettingKeyEQ(key),
		).
		Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantSettingRepo.GetByKey error: %v", err)
		return nil, err
	}

	_ = r.c.Set(ctx, cacheKey, row)
	return row, nil
}

// Update updates a tenant setting
func (r *tenantSettingRepository) Update(ctx context.Context, id string, updates types.JSON) (*ent.TenantSetting, error) {
	setting, err := r.ec.TenantSetting.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	builder := setting.Update()

	for field, value := range updates {
		switch field {
		case "setting_name":
			builder.SetSettingName(value.(string))
		case "setting_value":
			builder.SetSettingValue(value.(string))
		case "default_value":
			builder.SetDefaultValue(value.(string))
		case "setting_type":
			builder.SetSettingType(value.(string))
		case "scope":
			builder.SetScope(value.(string))
		case "category":
			builder.SetCategory(value.(string))
		case "description":
			builder.SetDescription(value.(string))
		case "is_public":
			builder.SetIsPublic(value.(bool))
		case "is_required":
			builder.SetIsRequired(value.(bool))
		case "is_readonly":
			builder.SetIsReadonly(value.(bool))
		case "validation":
			builder.SetValidation(value.(types.JSON))
		case "extras":
			builder.SetExtras(value.(types.JSON))
		case "updated_by":
			builder.SetUpdatedBy(value.(string))
		}
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantSettingRepo.Update error: %v", err)
		return nil, err
	}

	// Clear cache
	_ = r.c.Delete(ctx, fmt.Sprintf("id:%s", id))
	_ = r.c.Delete(ctx, fmt.Sprintf("tenant:%s:key:%s", setting.TenantID, setting.SettingKey))

	return row, nil
}

// Delete deletes a tenant setting
func (r *tenantSettingRepository) Delete(ctx context.Context, id string) error {
	setting, err := r.ec.TenantSetting.Get(ctx, id)
	if err != nil {
		return err
	}

	if err := r.ec.TenantSetting.DeleteOneID(id).Exec(ctx); err != nil {
		logger.Errorf(ctx, "tenantSettingRepo.Delete error: %v", err)
		return err
	}

	// Clear cache
	_ = r.c.Delete(ctx, fmt.Sprintf("id:%s", id))
	_ = r.c.Delete(ctx, fmt.Sprintf("tenant:%s:key:%s", setting.TenantID, setting.SettingKey))

	return nil
}

// List lists tenant settings
func (r *tenantSettingRepository) List(ctx context.Context, params *structs.ListTenantSettingParams) ([]*ent.TenantSetting, error) {
	builder := r.buildListQuery(params)

	if params.Cursor != "" {
		id, timestamp, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %v", err)
		}

		if !nanoid.IsPrimaryKey(id) {
			return nil, fmt.Errorf("invalid id in cursor: %s", id)
		}

		if params.Direction == "backward" {
			builder.Where(tenantSettingEnt.Or(
				tenantSettingEnt.CreatedAtGT(timestamp),
				tenantSettingEnt.And(
					tenantSettingEnt.CreatedAtEQ(timestamp),
					tenantSettingEnt.IDGT(id),
				),
			))
		} else {
			builder.Where(tenantSettingEnt.Or(
				tenantSettingEnt.CreatedAtLT(timestamp),
				tenantSettingEnt.And(
					tenantSettingEnt.CreatedAtEQ(timestamp),
					tenantSettingEnt.IDLT(id),
				),
			))
		}
	}

	if params.Direction == "backward" {
		builder.Order(ent.Asc(tenantSettingEnt.FieldCreatedAt), ent.Asc(tenantSettingEnt.FieldID))
	} else {
		builder.Order(ent.Desc(tenantSettingEnt.FieldCreatedAt), ent.Desc(tenantSettingEnt.FieldID))
	}

	builder.Limit(params.Limit)

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantSettingRepo.List error: %v", err)
		return nil, err
	}

	return rows, nil
}

// ListWithCount lists tenant settings with count
func (r *tenantSettingRepository) ListWithCount(ctx context.Context, params *structs.ListTenantSettingParams) ([]*ent.TenantSetting, int, error) {
	builder := r.buildListQuery(params)

	total, err := builder.Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantSettingRepo.ListWithCount count error: %v", err)
		return nil, 0, err
	}

	rows, err := r.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

// buildListQuery builds the list query based on parameters
func (r *tenantSettingRepository) buildListQuery(params *structs.ListTenantSettingParams) *ent.TenantSettingQuery {
	builder := r.ec.TenantSetting.Query()

	if params.TenantID != "" {
		builder.Where(tenantSettingEnt.TenantIDEQ(params.TenantID))
	}

	if params.Category != "" {
		builder.Where(tenantSettingEnt.CategoryEQ(params.Category))
	}

	if params.Scope != "" {
		builder.Where(tenantSettingEnt.ScopeEQ(string(params.Scope)))
	}

	if params.IsPublic != nil {
		builder.Where(tenantSettingEnt.IsPublicEQ(*params.IsPublic))
	}

	if params.IsRequired != nil {
		builder.Where(tenantSettingEnt.IsRequiredEQ(*params.IsRequired))
	}

	return builder
}
