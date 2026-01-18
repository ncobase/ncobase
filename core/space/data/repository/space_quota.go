package repository

import (
	"context"
	"fmt"
	"ncobase/core/space/data"
	"ncobase/core/space/data/ent"
	spaceQuotaEnt "ncobase/core/space/data/ent/spacequota"
	"ncobase/core/space/structs"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ncobase/ncore/data/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"
)

// SpaceQuotaRepositoryInterface defines the interface for space quota repository
type SpaceQuotaRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateSpaceQuotaBody) (*ent.SpaceQuota, error)
	GetByID(ctx context.Context, id string) (*ent.SpaceQuota, error)
	GetBySpaceID(ctx context.Context, spaceID string) ([]*ent.SpaceQuota, error)
	GetBySpaceAndType(ctx context.Context, spaceID string, quotaType structs.QuotaType) (*ent.SpaceQuota, error)
	Update(ctx context.Context, id string, updates types.JSON) (*ent.SpaceQuota, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params *structs.ListSpaceQuotaParams) ([]*ent.SpaceQuota, error)
	ListWithCount(ctx context.Context, params *structs.ListSpaceQuotaParams) ([]*ent.SpaceQuota, int, error)
	CountX(ctx context.Context, params *structs.ListSpaceQuotaParams) int
}

// spaceQuotaRepository implements SpaceQuotaRepositoryInterface
type spaceQuotaRepository struct {
	data                *data.Data
	quotaCache          cache.ICache[ent.SpaceQuota]
	spaceQuotasCache    cache.ICache[[]string] // Maps space ID to quota IDs
	spaceTypeQuotaCache cache.ICache[string]   // Maps space:type to quota ID
	quotaTypeCache      cache.ICache[[]string] // Maps quota type to quota IDs
	quotaTTL            time.Duration
}

// NewSpaceQuotaRepository creates a new space quota repository
func NewSpaceQuotaRepository(d *data.Data) SpaceQuotaRepositoryInterface {
	redisClient := d.GetRedis().(*redis.Client)

	return &spaceQuotaRepository{
		data:                d,
		quotaCache:          cache.NewCache[ent.SpaceQuota](redisClient, "ncse_space:quotas"),
		spaceQuotasCache:    cache.NewCache[[]string](redisClient, "ncse_space:space_quota_mappings"),
		spaceTypeQuotaCache: cache.NewCache[string](redisClient, "ncse_space:space_type_quota_mappings"),
		quotaTypeCache:      cache.NewCache[[]string](redisClient, "ncse_space:quota_type_mappings"),
		quotaTTL:            time.Hour * 4, // 4 hours cache TTL (quotas change less frequently)
	}
}

// Create creates a new space quota
func (r *spaceQuotaRepository) Create(ctx context.Context, body *structs.CreateSpaceQuotaBody) (*ent.SpaceQuota, error) {
	// Use master for writes
	builder := r.data.GetMasterEntClient().SpaceQuota.Create()

	builder.SetSpaceID(body.SpaceID)
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
		logger.Errorf(ctx, "spaceQuotaRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the quota and invalidate related caches
	go func() {
		r.cacheQuota(context.Background(), row)
		r.invalidateSpaceQuotasCache(context.Background(), body.SpaceID)
		r.invalidateQuotaTypeCache(context.Background(), string(body.QuotaType))
	}()

	return row, nil
}

// GetByID retrieves a space quota by ID
func (r *spaceQuotaRepository) GetByID(ctx context.Context, id string) (*ent.SpaceQuota, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("id:%s", id)
	if cached, err := r.quotaCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Use slave for reads
	row, err := r.data.GetSlaveEntClient().SpaceQuota.Get(ctx, id)
	if err != nil {
		logger.Errorf(ctx, "spaceQuotaRepo.GetByID error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheQuota(context.Background(), row)

	return row, nil
}

// GetBySpaceID retrieves all quotas for a space
func (r *spaceQuotaRepository) GetBySpaceID(ctx context.Context, spaceID string) ([]*ent.SpaceQuota, error) {
	// Try to get quota IDs from cache
	cacheKey := fmt.Sprintf("space_quotas:%s", spaceID)
	var quotaIDs []string
	if err := r.spaceQuotasCache.GetArray(ctx, cacheKey, &quotaIDs); err == nil && len(quotaIDs) > 0 {
		// Get quotas by IDs
		quotas := make([]*ent.SpaceQuota, 0, len(quotaIDs))
		for _, quotaID := range quotaIDs {
			if quota, err := r.GetByID(ctx, quotaID); err == nil {
				quotas = append(quotas, quota)
			}
		}
		return quotas, nil
	}

	// Fallback to database
	rows, err := r.data.GetSlaveEntClient().SpaceQuota.Query().
		Where(spaceQuotaEnt.SpaceIDEQ(spaceID)).
		All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceQuotaRepo.GetBySpaceID error: %v", err)
		return nil, err
	}

	// Cache quotas and quota IDs
	go func() {
		quotaIDs := make([]string, 0, len(rows))
		for _, quota := range rows {
			r.cacheQuota(context.Background(), quota)
			quotaIDs = append(quotaIDs, quota.ID)
		}

		if err := r.spaceQuotasCache.SetArray(context.Background(), cacheKey, quotaIDs, r.quotaTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache space quotas %s: %v", spaceID, err)
		}
	}()

	return rows, nil
}

// GetBySpaceAndType retrieves a specific quota for a space
func (r *spaceQuotaRepository) GetBySpaceAndType(ctx context.Context, spaceID string, quotaType structs.QuotaType) (*ent.SpaceQuota, error) {
	// Try to get quota ID from space:type mapping cache
	if quotaID, err := r.getQuotaIDBySpaceAndType(ctx, spaceID, string(quotaType)); err == nil && quotaID != "" {
		return r.GetByID(ctx, quotaID)
	}

	// Fallback to database
	row, err := r.data.GetSlaveEntClient().SpaceQuota.Query().
		Where(
			spaceQuotaEnt.SpaceIDEQ(spaceID),
			spaceQuotaEnt.QuotaTypeEQ(string(quotaType)),
		).
		Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceQuotaRepo.GetBySpaceAndType error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheQuota(context.Background(), row)

	return row, nil
}

// Update updates a space quota
func (r *spaceQuotaRepository) Update(ctx context.Context, id string, updates types.JSON) (*ent.SpaceQuota, error) {
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
		logger.Errorf(ctx, "spaceQuotaRepo.Update error: %v", err)
		return nil, err
	}

	// Invalidate and re-cache
	go func() {
		r.invalidateQuotaCache(context.Background(), originalQuota)
		r.cacheQuota(context.Background(), row)

		// Invalidate related caches if space or type changed
		if originalQuota.SpaceID != row.SpaceID {
			r.invalidateSpaceQuotasCache(context.Background(), originalQuota.SpaceID)
			r.invalidateSpaceQuotasCache(context.Background(), row.SpaceID)
		} else {
			r.invalidateSpaceQuotasCache(context.Background(), row.SpaceID)
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

// Delete deletes a space quota
func (r *spaceQuotaRepository) Delete(ctx context.Context, id string) error {
	// Get quota first for cache invalidation
	quota, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Use master for writes
	if err := r.data.GetMasterEntClient().SpaceQuota.DeleteOneID(id).Exec(ctx); err != nil {
		logger.Errorf(ctx, "spaceQuotaRepo.Delete error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateQuotaCache(context.Background(), quota)
		r.invalidateSpaceQuotasCache(context.Background(), quota.SpaceID)
		r.invalidateQuotaTypeCache(context.Background(), quota.QuotaType)
	}()

	return nil
}

// List lists space quotas
func (r *spaceQuotaRepository) List(ctx context.Context, params *structs.ListSpaceQuotaParams) ([]*ent.SpaceQuota, error) {
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
			builder.Where(spaceQuotaEnt.Or(
				spaceQuotaEnt.CreatedAtGT(timestamp),
				spaceQuotaEnt.And(
					spaceQuotaEnt.CreatedAtEQ(timestamp),
					spaceQuotaEnt.IDGT(id),
				),
			))
		} else {
			builder.Where(spaceQuotaEnt.Or(
				spaceQuotaEnt.CreatedAtLT(timestamp),
				spaceQuotaEnt.And(
					spaceQuotaEnt.CreatedAtEQ(timestamp),
					spaceQuotaEnt.IDLT(id),
				),
			))
		}
	}

	if params.Direction == "backward" {
		builder.Order(ent.Asc(spaceQuotaEnt.FieldCreatedAt), ent.Asc(spaceQuotaEnt.FieldID))
	} else {
		builder.Order(ent.Desc(spaceQuotaEnt.FieldCreatedAt), ent.Desc(spaceQuotaEnt.FieldID))
	}

	builder.Limit(params.Limit)

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceQuotaRepo.List error: %v", err)
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

// ListWithCount lists space quotas with count
func (r *spaceQuotaRepository) ListWithCount(ctx context.Context, params *structs.ListSpaceQuotaParams) ([]*ent.SpaceQuota, int, error) {
	builder := r.buildListQuery(params)

	total, err := builder.Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceQuotaRepo.ListWithCount count error: %v", err)
		return nil, 0, err
	}

	rows, err := r.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

// CountX counts space quotas
func (r *spaceQuotaRepository) CountX(ctx context.Context, params *structs.ListSpaceQuotaParams) int {
	builder := r.buildListQuery(params)
	return builder.CountX(ctx)
}

// buildListQuery builds the list query based on parameters
func (r *spaceQuotaRepository) buildListQuery(params *structs.ListSpaceQuotaParams) *ent.SpaceQuotaQuery {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().SpaceQuota.Query()

	if params.SpaceID != "" {
		builder.Where(spaceQuotaEnt.SpaceIDEQ(params.SpaceID))
	}

	if params.QuotaType != "" {
		builder.Where(spaceQuotaEnt.QuotaTypeEQ(string(params.QuotaType)))
	}

	if params.Enabled != nil {
		builder.Where(spaceQuotaEnt.EnabledEQ(*params.Enabled))
	}

	return builder
}

// cacheQuota caches a space quota
func (r *spaceQuotaRepository) cacheQuota(ctx context.Context, quota *ent.SpaceQuota) {
	// Cache by ID
	idKey := fmt.Sprintf("id:%s", quota.ID)
	if err := r.quotaCache.Set(ctx, idKey, quota, r.quotaTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache quota by ID %s: %v", quota.ID, err)
	}

	// Cache space:type to ID mapping
	spaceTypeKey := fmt.Sprintf("space_type:%s:%s", quota.SpaceID, quota.QuotaType)
	if err := r.spaceTypeQuotaCache.Set(ctx, spaceTypeKey, &quota.ID, r.quotaTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache space type mapping %s:%s: %v", quota.SpaceID, quota.QuotaType, err)
	}
}

// invalidateQuotaCache invalidates the cache for a space quota
func (r *spaceQuotaRepository) invalidateQuotaCache(ctx context.Context, quota *ent.SpaceQuota) {
	// Invalidate ID cache
	idKey := fmt.Sprintf("id:%s", quota.ID)
	if err := r.quotaCache.Delete(ctx, idKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate quota ID cache %s: %v", quota.ID, err)
	}

	// Invalidate space:type mapping
	spaceTypeKey := fmt.Sprintf("space_type:%s:%s", quota.SpaceID, quota.QuotaType)
	if err := r.spaceTypeQuotaCache.Delete(ctx, spaceTypeKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate space type mapping cache %s:%s: %v", quota.SpaceID, quota.QuotaType, err)
	}
}

// invalidateSpaceQuotasCache invalidates the cache for all space quotas
func (r *spaceQuotaRepository) invalidateSpaceQuotasCache(ctx context.Context, spaceID string) {
	cacheKey := fmt.Sprintf("space_quotas:%s", spaceID)
	if err := r.spaceQuotasCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate space quotas cache %s: %v", spaceID, err)
	}
}

// invalidateQuotaTypeCache invalidates the cache for a quota type
func (r *spaceQuotaRepository) invalidateQuotaTypeCache(ctx context.Context, quotaType string) {
	cacheKey := fmt.Sprintf("quota_type:%s", quotaType)
	if err := r.quotaTypeCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate quota type cache %s: %v", quotaType, err)
	}
}

// getQuotaIDBySpaceAndType gets the ID of a space quota by space ID and type
func (r *spaceQuotaRepository) getQuotaIDBySpaceAndType(ctx context.Context, spaceID, quotaType string) (string, error) {
	cacheKey := fmt.Sprintf("space_type:%s:%s", spaceID, quotaType)
	quotaID, err := r.spaceTypeQuotaCache.Get(ctx, cacheKey)
	if err != nil || quotaID == nil {
		return "", err
	}
	return *quotaID, nil
}
