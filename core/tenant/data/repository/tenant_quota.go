package repository

import (
	"context"
	"fmt"
	"ncobase/tenant/data"
	"ncobase/tenant/data/ent"
	tenantQuotaEnt "ncobase/tenant/data/ent/tenantquota"
	"ncobase/tenant/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"
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
	CountX(ctx context.Context, params *structs.ListTenantQuotaParams) int
}

// tenantQuotaRepository implements TenantQuotaRepositoryInterface
type tenantQuotaRepository struct {
	data                 *data.Data
	quotaCache           cache.ICache[ent.TenantQuota]
	tenantQuotasCache    cache.ICache[[]string] // Maps tenant ID to quota IDs
	tenantTypeQuotaCache cache.ICache[string]   // Maps tenant:type to quota ID
	quotaTypeCache       cache.ICache[[]string] // Maps quota type to quota IDs
	quotaTTL             time.Duration
}

// NewTenantQuotaRepository creates a new tenant quota repository
func NewTenantQuotaRepository(d *data.Data) TenantQuotaRepositoryInterface {
	redisClient := d.GetRedis()

	return &tenantQuotaRepository{
		data:                 d,
		quotaCache:           cache.NewCache[ent.TenantQuota](redisClient, "ncse_tenant:quotas"),
		tenantQuotasCache:    cache.NewCache[[]string](redisClient, "ncse_tenant:tenant_quota_mappings"),
		tenantTypeQuotaCache: cache.NewCache[string](redisClient, "ncse_tenant:tenant_type_quota_mappings"),
		quotaTypeCache:       cache.NewCache[[]string](redisClient, "ncse_tenant:quota_type_mappings"),
		quotaTTL:             time.Hour * 4, // 4 hours cache TTL (quotas change less frequently)
	}
}

// Create creates a new tenant quota
func (r *tenantQuotaRepository) Create(ctx context.Context, body *structs.CreateTenantQuotaBody) (*ent.TenantQuota, error) {
	// Use master for writes
	builder := r.data.GetMasterEntClient().TenantQuota.Create()

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

	// Cache the quota and invalidate related caches
	go func() {
		r.cacheQuota(context.Background(), row)
		r.invalidateTenantQuotasCache(context.Background(), body.TenantID)
		r.invalidateQuotaTypeCache(context.Background(), string(body.QuotaType))
	}()

	return row, nil
}

// GetByID retrieves a tenant quota by ID
func (r *tenantQuotaRepository) GetByID(ctx context.Context, id string) (*ent.TenantQuota, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("id:%s", id)
	if cached, err := r.quotaCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Use slave for reads
	row, err := r.data.GetSlaveEntClient().TenantQuota.Get(ctx, id)
	if err != nil {
		logger.Errorf(ctx, "tenantQuotaRepo.GetByID error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheQuota(context.Background(), row)

	return row, nil
}

// GetByTenantID retrieves all quotas for a tenant
func (r *tenantQuotaRepository) GetByTenantID(ctx context.Context, tenantID string) ([]*ent.TenantQuota, error) {
	// Try to get quota IDs from cache
	cacheKey := fmt.Sprintf("tenant_quotas:%s", tenantID)
	var quotaIDs []string
	if err := r.tenantQuotasCache.GetArray(ctx, cacheKey, &quotaIDs); err == nil && len(quotaIDs) > 0 {
		// Get quotas by IDs
		quotas := make([]*ent.TenantQuota, 0, len(quotaIDs))
		for _, quotaID := range quotaIDs {
			if quota, err := r.GetByID(ctx, quotaID); err == nil {
				quotas = append(quotas, quota)
			}
		}
		return quotas, nil
	}

	// Fallback to database
	rows, err := r.data.GetSlaveEntClient().TenantQuota.Query().
		Where(tenantQuotaEnt.TenantIDEQ(tenantID)).
		All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantQuotaRepo.GetByTenantID error: %v", err)
		return nil, err
	}

	// Cache quotas and quota IDs
	go func() {
		quotaIDs := make([]string, len(rows))
		for i, quota := range rows {
			r.cacheQuota(context.Background(), quota)
			quotaIDs[i] = quota.ID
		}

		if err := r.tenantQuotasCache.SetArray(context.Background(), cacheKey, quotaIDs, r.quotaTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache tenant quotas %s: %v", tenantID, err)
		}
	}()

	return rows, nil
}

// GetByTenantAndType retrieves a specific quota for a tenant
func (r *tenantQuotaRepository) GetByTenantAndType(ctx context.Context, tenantID string, quotaType structs.QuotaType) (*ent.TenantQuota, error) {
	// Try to get quota ID from tenant:type mapping cache
	if quotaID, err := r.getQuotaIDByTenantAndType(ctx, tenantID, string(quotaType)); err == nil && quotaID != "" {
		return r.GetByID(ctx, quotaID)
	}

	// Fallback to database
	row, err := r.data.GetSlaveEntClient().TenantQuota.Query().
		Where(
			tenantQuotaEnt.TenantIDEQ(tenantID),
			tenantQuotaEnt.QuotaTypeEQ(string(quotaType)),
		).
		Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantQuotaRepo.GetByTenantAndType error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheQuota(context.Background(), row)

	return row, nil
}

// Update updates a tenant quota
func (r *tenantQuotaRepository) Update(ctx context.Context, id string, updates types.JSON) (*ent.TenantQuota, error) {
	// Get original quota for cache invalidation
	originalQuota, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Use master for writes
	builder := originalQuota.Update()

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

	// Invalidate and re-cache
	go func() {
		r.invalidateQuotaCache(context.Background(), originalQuota)
		r.cacheQuota(context.Background(), row)

		// Invalidate related caches if tenant or type changed
		if originalQuota.TenantID != row.TenantID {
			r.invalidateTenantQuotasCache(context.Background(), originalQuota.TenantID)
			r.invalidateTenantQuotasCache(context.Background(), row.TenantID)
		} else {
			r.invalidateTenantQuotasCache(context.Background(), row.TenantID)
		}

		if originalQuota.QuotaType != row.QuotaType {
			r.invalidateQuotaTypeCache(context.Background(), originalQuota.QuotaType)
			r.invalidateQuotaTypeCache(context.Background(), row.QuotaType)
		} else {
			r.invalidateQuotaTypeCache(context.Background(), row.QuotaType)
		}
	}()

	return row, nil
}

// Delete deletes a tenant quota
func (r *tenantQuotaRepository) Delete(ctx context.Context, id string) error {
	// Get quota first for cache invalidation
	quota, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Use master for writes
	if err := r.data.GetMasterEntClient().TenantQuota.DeleteOneID(id).Exec(ctx); err != nil {
		logger.Errorf(ctx, "tenantQuotaRepo.Delete error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateQuotaCache(context.Background(), quota)
		r.invalidateTenantQuotasCache(context.Background(), quota.TenantID)
		r.invalidateQuotaTypeCache(context.Background(), quota.QuotaType)
	}()

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

	// Cache quotas in background
	go func() {
		for _, quota := range rows {
			r.cacheQuota(context.Background(), quota)
		}
	}()

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

// CountX counts tenant quotas
func (r *tenantQuotaRepository) CountX(ctx context.Context, params *structs.ListTenantQuotaParams) int {
	builder := r.buildListQuery(params)
	return builder.CountX(ctx)
}

// buildListQuery builds the list query based on parameters
func (r *tenantQuotaRepository) buildListQuery(params *structs.ListTenantQuotaParams) *ent.TenantQuotaQuery {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().TenantQuota.Query()

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

// cacheQuota caches a tenant quota
func (r *tenantQuotaRepository) cacheQuota(ctx context.Context, quota *ent.TenantQuota) {
	// Cache by ID
	idKey := fmt.Sprintf("id:%s", quota.ID)
	if err := r.quotaCache.Set(ctx, idKey, quota, r.quotaTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache quota by ID %s: %v", quota.ID, err)
	}

	// Cache tenant:type to ID mapping
	tenantTypeKey := fmt.Sprintf("tenant_type:%s:%s", quota.TenantID, quota.QuotaType)
	if err := r.tenantTypeQuotaCache.Set(ctx, tenantTypeKey, &quota.ID, r.quotaTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache tenant type mapping %s:%s: %v", quota.TenantID, quota.QuotaType, err)
	}
}

// invalidateQuotaCache invalidates the cache for a tenant quota
func (r *tenantQuotaRepository) invalidateQuotaCache(ctx context.Context, quota *ent.TenantQuota) {
	// Invalidate ID cache
	idKey := fmt.Sprintf("id:%s", quota.ID)
	if err := r.quotaCache.Delete(ctx, idKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate quota ID cache %s: %v", quota.ID, err)
	}

	// Invalidate tenant:type mapping
	tenantTypeKey := fmt.Sprintf("tenant_type:%s:%s", quota.TenantID, quota.QuotaType)
	if err := r.tenantTypeQuotaCache.Delete(ctx, tenantTypeKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate tenant type mapping cache %s:%s: %v", quota.TenantID, quota.QuotaType, err)
	}
}

// invalidateTenantQuotasCache invalidates the cache for all tenant quotas
func (r *tenantQuotaRepository) invalidateTenantQuotasCache(ctx context.Context, tenantID string) {
	cacheKey := fmt.Sprintf("tenant_quotas:%s", tenantID)
	if err := r.tenantQuotasCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate tenant quotas cache %s: %v", tenantID, err)
	}
}

// invalidateQuotaTypeCache invalidates the cache for a quota type
func (r *tenantQuotaRepository) invalidateQuotaTypeCache(ctx context.Context, quotaType string) {
	cacheKey := fmt.Sprintf("quota_type:%s", quotaType)
	if err := r.quotaTypeCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate quota type cache %s: %v", quotaType, err)
	}
}

// getQuotaIDByTenantAndType gets the ID of a tenant quota by tenant ID and type
func (r *tenantQuotaRepository) getQuotaIDByTenantAndType(ctx context.Context, tenantID, quotaType string) (string, error) {
	cacheKey := fmt.Sprintf("tenant_type:%s:%s", tenantID, quotaType)
	quotaID, err := r.tenantTypeQuotaCache.Get(ctx, cacheKey)
	if err != nil || quotaID == nil {
		return "", err
	}
	return *quotaID, nil
}
