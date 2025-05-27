package repository

import (
	"context"
	"fmt"
	"ncobase/tenant/data"
	"ncobase/tenant/data/ent"
	tenantQuotaEnt "ncobase/tenant/data/ent/tenantquota"
	"ncobase/tenant/structs"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/redis/go-redis/v9"
)

// TenantQuotaRepositoryInterface defines the interface for tenant quota repository
type TenantQuotaRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateTenantQuotaBody) (*ent.TenantQuota, error)
	GetByID(ctx context.Context, id string) (*ent.TenantQuota, error)
	GetByTenantID(ctx context.Context, tenantID string) ([]*ent.TenantQuota, error)
	GetByTenantAndType(ctx context.Context, tenantID string, quotaType structs.QuotaType) (*ent.TenantQuota, error)
	Update(ctx context.Context, id string, updates types.JSON) (*ent.TenantQuota, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params *structs.ListTenantQuotaParams) ([]*ent.TenantQuota, error)
	ListWithCount(ctx context.Context, params *structs.ListTenantQuotaParams) ([]*ent.TenantQuota, int, error)
}

// tenantQuotaRepository implements TenantQuotaRepositoryInterface
type tenantQuotaRepository struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.TenantQuota]
}

// NewTenantQuotaRepository creates a new tenant quota repository
func NewTenantQuotaRepository(d *data.Data) TenantQuotaRepositoryInterface {
	ec := d.GetMasterEntClient()
	rc := d.GetRedis()
	return &tenantQuotaRepository{ec, rc, cache.NewCache[ent.TenantQuota](rc, "ncse_tenant_quota")}
}

// Create creates a new tenant quota
func (r *tenantQuotaRepository) Create(ctx context.Context, body *structs.CreateTenantQuotaBody) (*ent.TenantQuota, error) {
	builder := r.ec.TenantQuota.Create()

	builder.SetTenantID(body.TenantID)
	builder.SetQuotaType(string(body.QuotaType))
	builder.SetQuotaName(body.QuotaName)
	builder.SetMaxValue(body.MaxValue)
	builder.SetCurrentUsed(body.CurrentUsed)
	builder.SetUnit(string(body.Unit))
	builder.SetDescription(body.Description)
	builder.SetEnabled(body.Enabled)
	builder.SetNillableCreatedBy(body.CreatedBy)

	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantQuotaRepo.Create error: %v", err)
		return nil, err
	}

	return row, nil
}

// GetByID retrieves a tenant quota by ID
func (r *tenantQuotaRepository) GetByID(ctx context.Context, id string) (*ent.TenantQuota, error) {
	cacheKey := fmt.Sprintf("id:%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	row, err := r.ec.TenantQuota.Get(ctx, id)
	if err != nil {
		logger.Errorf(ctx, "tenantQuotaRepo.GetByID error: %v", err)
		return nil, err
	}

	_ = r.c.Set(ctx, cacheKey, row)
	return row, nil
}

// GetByTenantID retrieves all quotas for a tenant
func (r *tenantQuotaRepository) GetByTenantID(ctx context.Context, tenantID string) ([]*ent.TenantQuota, error) {
	rows, err := r.ec.TenantQuota.Query().
		Where(tenantQuotaEnt.TenantIDEQ(tenantID)).
		All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantQuotaRepo.GetByTenantID error: %v", err)
		return nil, err
	}

	return rows, nil
}

// GetByTenantAndType retrieves a specific quota for a tenant
func (r *tenantQuotaRepository) GetByTenantAndType(ctx context.Context, tenantID string, quotaType structs.QuotaType) (*ent.TenantQuota, error) {
	cacheKey := fmt.Sprintf("tenant:%s:type:%s", tenantID, quotaType)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	row, err := r.ec.TenantQuota.Query().
		Where(
			tenantQuotaEnt.TenantIDEQ(tenantID),
			tenantQuotaEnt.QuotaTypeEQ(string(quotaType)),
		).
		Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantQuotaRepo.GetByTenantAndType error: %v", err)
		return nil, err
	}

	_ = r.c.Set(ctx, cacheKey, row)
	return row, nil
}

// Update updates a tenant quota
func (r *tenantQuotaRepository) Update(ctx context.Context, id string, updates types.JSON) (*ent.TenantQuota, error) {
	quota, err := r.ec.TenantQuota.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	builder := quota.Update()

	for field, value := range updates {
		switch field {
		case "quota_name":
			builder.SetQuotaName(value.(string))
		case "max_value":
			builder.SetMaxValue(int64(value.(float64)))
		case "current_used":
			builder.SetCurrentUsed(int64(value.(float64)))
		case "unit":
			builder.SetUnit(value.(string))
		case "description":
			builder.SetDescription(value.(string))
		case "enabled":
			builder.SetEnabled(value.(bool))
		case "extras":
			builder.SetExtras(value.(types.JSON))
		case "updated_by":
			builder.SetUpdatedBy(value.(string))
		}
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantQuotaRepo.Update error: %v", err)
		return nil, err
	}

	// Clear cache
	_ = r.c.Delete(ctx, fmt.Sprintf("id:%s", id))
	_ = r.c.Delete(ctx, fmt.Sprintf("tenant:%s:type:%s", quota.TenantID, quota.QuotaType))

	return row, nil
}

// Delete deletes a tenant quota
func (r *tenantQuotaRepository) Delete(ctx context.Context, id string) error {
	quota, err := r.ec.TenantQuota.Get(ctx, id)
	if err != nil {
		return err
	}

	if err := r.ec.TenantQuota.DeleteOneID(id).Exec(ctx); err != nil {
		logger.Errorf(ctx, "tenantQuotaRepo.Delete error: %v", err)
		return err
	}

	// Clear cache
	_ = r.c.Delete(ctx, fmt.Sprintf("id:%s", id))
	_ = r.c.Delete(ctx, fmt.Sprintf("tenant:%s:type:%s", quota.TenantID, quota.QuotaType))

	return nil
}

// List lists tenant quotas
func (r *tenantQuotaRepository) List(ctx context.Context, params *structs.ListTenantQuotaParams) ([]*ent.TenantQuota, error) {
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
			builder.Where(tenantQuotaEnt.Or(
				tenantQuotaEnt.CreatedAtGT(timestamp),
				tenantQuotaEnt.And(
					tenantQuotaEnt.CreatedAtEQ(timestamp),
					tenantQuotaEnt.IDGT(id),
				),
			))
		} else {
			builder.Where(tenantQuotaEnt.Or(
				tenantQuotaEnt.CreatedAtLT(timestamp),
				tenantQuotaEnt.And(
					tenantQuotaEnt.CreatedAtEQ(timestamp),
					tenantQuotaEnt.IDLT(id),
				),
			))
		}
	}

	if params.Direction == "backward" {
		builder.Order(ent.Asc(tenantQuotaEnt.FieldCreatedAt), ent.Asc(tenantQuotaEnt.FieldID))
	} else {
		builder.Order(ent.Desc(tenantQuotaEnt.FieldCreatedAt), ent.Desc(tenantQuotaEnt.FieldID))
	}

	builder.Limit(params.Limit)

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantQuotaRepo.List error: %v", err)
		return nil, err
	}

	return rows, nil
}

// ListWithCount lists tenant quotas with count
func (r *tenantQuotaRepository) ListWithCount(ctx context.Context, params *structs.ListTenantQuotaParams) ([]*ent.TenantQuota, int, error) {
	builder := r.buildListQuery(params)

	total, err := builder.Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantQuotaRepo.ListWithCount count error: %v", err)
		return nil, 0, err
	}

	rows, err := r.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

// buildListQuery builds the list query based on parameters
func (r *tenantQuotaRepository) buildListQuery(params *structs.ListTenantQuotaParams) *ent.TenantQuotaQuery {
	builder := r.ec.TenantQuota.Query()

	if params.TenantID != "" {
		builder.Where(tenantQuotaEnt.TenantIDEQ(params.TenantID))
	}

	if params.QuotaType != "" {
		builder.Where(tenantQuotaEnt.QuotaTypeEQ(string(params.QuotaType)))
	}

	if params.Enabled != nil {
		builder.Where(tenantQuotaEnt.EnabledEQ(*params.Enabled))
	}

	return builder
}
