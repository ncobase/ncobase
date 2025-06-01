package repository

import (
	"context"
	"fmt"
	"ncobase/tenant/data"
	"ncobase/tenant/data/ent"
	tenantSettingEnt "ncobase/tenant/data/ent/tenantsetting"
	"ncobase/tenant/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"
)

// TenantSettingRepositoryInterface defines the interface for tenant setting repository
type TenantSettingRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateTenantSettingBody) (*ent.TenantSetting, error)
	GetByID(ctx context.Context, id string) (*ent.TenantSetting, error)
	GetByKey(ctx context.Context, tenantID, key string) (*ent.TenantSetting, error)
	GetByTenantID(ctx context.Context, tenantID string) ([]*ent.TenantSetting, error)
	GetByCategory(ctx context.Context, tenantID, category string) ([]*ent.TenantSetting, error)
	Update(ctx context.Context, id string, updates types.JSON) (*ent.TenantSetting, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params *structs.ListTenantSettingParams) ([]*ent.TenantSetting, error)
	ListWithCount(ctx context.Context, params *structs.ListTenantSettingParams) ([]*ent.TenantSetting, int, error)
	CountX(ctx context.Context, params *structs.ListTenantSettingParams) int
}

// tenantSettingRepository implements TenantSettingRepositoryInterface
type tenantSettingRepository struct {
	data                  *data.Data
	settingCache          cache.ICache[ent.TenantSetting]
	tenantSettingsCache   cache.ICache[[]string] // Maps tenant ID to setting IDs
	tenantKeySettingCache cache.ICache[string]   // Maps tenant:key to setting ID
	categorySettingsCache cache.ICache[[]string] // Maps tenant:category to setting IDs
	scopeSettingsCache    cache.ICache[[]string] // Maps tenant:scope to setting IDs
	settingTTL            time.Duration
}

// NewTenantSettingRepository creates a new tenant setting repository
func NewTenantSettingRepository(d *data.Data) TenantSettingRepositoryInterface {
	redisClient := d.GetRedis()

	return &tenantSettingRepository{
		data:                  d,
		settingCache:          cache.NewCache[ent.TenantSetting](redisClient, "ncse_tenant:settings"),
		tenantSettingsCache:   cache.NewCache[[]string](redisClient, "ncse_tenant:tenant_setting_mappings"),
		tenantKeySettingCache: cache.NewCache[string](redisClient, "ncse_tenant:tenant_key_setting_mappings"),
		categorySettingsCache: cache.NewCache[[]string](redisClient, "ncse_tenant:category_setting_mappings"),
		scopeSettingsCache:    cache.NewCache[[]string](redisClient, "ncse_tenant:scope_setting_mappings"),
		settingTTL:            time.Hour * 6, // 6 hours cache TTL (settings change less frequently)
	}
}

// Create creates a new tenant setting
func (r *tenantSettingRepository) Create(ctx context.Context, body *structs.CreateTenantSettingBody) (*ent.TenantSetting, error) {
	// Use master for writes
	builder := r.data.GetMasterEntClient().TenantSetting.Create()

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

	// Cache the setting and invalidate related caches
	go func() {
		r.cacheSetting(context.Background(), row)
		r.invalidateTenantSettingsCache(context.Background(), body.TenantID)
		r.invalidateCategorySettingsCache(context.Background(), body.TenantID, body.Category)
		r.invalidateScopeSettingsCache(context.Background(), body.TenantID, string(body.Scope))
	}()

	return row, nil
}

// GetByID retrieves a tenant setting by ID
func (r *tenantSettingRepository) GetByID(ctx context.Context, id string) (*ent.TenantSetting, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("id:%s", id)
	if cached, err := r.settingCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Use slave for reads
	row, err := r.data.GetSlaveEntClient().TenantSetting.Get(ctx, id)
	if err != nil {
		logger.Errorf(ctx, "tenantSettingRepo.GetByID error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheSetting(context.Background(), row)

	return row, nil
}

// GetByKey retrieves a tenant setting by tenant ID and key
func (r *tenantSettingRepository) GetByKey(ctx context.Context, tenantID, key string) (*ent.TenantSetting, error) {
	// Try to get setting ID from tenant:key mapping cache
	if settingID, err := r.getSettingIDByTenantAndKey(ctx, tenantID, key); err == nil && settingID != "" {
		return r.GetByID(ctx, settingID)
	}

	// Fallback to database
	row, err := r.data.GetSlaveEntClient().TenantSetting.Query().
		Where(
			tenantSettingEnt.TenantIDEQ(tenantID),
			tenantSettingEnt.SettingKeyEQ(key),
		).
		Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantSettingRepo.GetByKey error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheSetting(context.Background(), row)

	return row, nil
}

// GetByTenantID retrieves all settings for a tenant
func (r *tenantSettingRepository) GetByTenantID(ctx context.Context, tenantID string) ([]*ent.TenantSetting, error) {
	// Try to get setting IDs from cache
	cacheKey := fmt.Sprintf("tenant_settings:%s", tenantID)
	var settingIDs []string
	if err := r.tenantSettingsCache.GetArray(ctx, cacheKey, &settingIDs); err == nil && len(settingIDs) > 0 {
		// Get settings by IDs
		settings := make([]*ent.TenantSetting, 0, len(settingIDs))
		for _, settingID := range settingIDs {
			if setting, err := r.GetByID(ctx, settingID); err == nil {
				settings = append(settings, setting)
			}
		}
		return settings, nil
	}

	// Fallback to database
	rows, err := r.data.GetSlaveEntClient().TenantSetting.Query().
		Where(tenantSettingEnt.TenantIDEQ(tenantID)).
		All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantSettingRepo.GetByTenantID error: %v", err)
		return nil, err
	}

	// Cache settings and setting IDs
	go func() {
		settingIDs := make([]string, len(rows))
		for i, setting := range rows {
			r.cacheSetting(context.Background(), setting)
			settingIDs[i] = setting.ID
		}

		if err := r.tenantSettingsCache.SetArray(context.Background(), cacheKey, settingIDs, r.settingTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache tenant settings %s: %v", tenantID, err)
		}
	}()

	return rows, nil
}

// GetByCategory retrieves settings by tenant ID and category
func (r *tenantSettingRepository) GetByCategory(ctx context.Context, tenantID, category string) ([]*ent.TenantSetting, error) {
	// Try to get setting IDs from cache
	cacheKey := fmt.Sprintf("category_settings:%s:%s", tenantID, category)
	var settingIDs []string
	if err := r.categorySettingsCache.GetArray(ctx, cacheKey, &settingIDs); err == nil && len(settingIDs) > 0 {
		// Get settings by IDs
		settings := make([]*ent.TenantSetting, 0, len(settingIDs))
		for _, settingID := range settingIDs {
			if setting, err := r.GetByID(ctx, settingID); err == nil {
				settings = append(settings, setting)
			}
		}
		return settings, nil
	}

	// Fallback to database
	rows, err := r.data.GetSlaveEntClient().TenantSetting.Query().
		Where(
			tenantSettingEnt.TenantIDEQ(tenantID),
			tenantSettingEnt.CategoryEQ(category),
		).
		All(ctx)
	if err != nil {
		logger.Errorf(ctx, "tenantSettingRepo.GetByCategory error: %v", err)
		return nil, err
	}

	// Cache settings and setting IDs
	go func() {
		settingIDs := make([]string, len(rows))
		for i, setting := range rows {
			r.cacheSetting(context.Background(), setting)
			settingIDs[i] = setting.ID
		}

		if err := r.categorySettingsCache.SetArray(context.Background(), cacheKey, settingIDs, r.settingTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache category settings %s:%s: %v", tenantID, category, err)
		}
	}()

	return rows, nil
}

// Update updates a tenant setting
func (r *tenantSettingRepository) Update(ctx context.Context, id string, updates types.JSON) (*ent.TenantSetting, error) {
	// Get original setting for cache invalidation
	originalSetting, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Use master for writes
	builder := originalSetting.Update()

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

	// Invalidate and re-cache
	go func() {
		r.invalidateSettingCache(context.Background(), originalSetting)
		r.cacheSetting(context.Background(), row)

		// Invalidate related caches
		r.invalidateTenantSettingsCache(context.Background(), row.TenantID)

		// Invalidate category cache for both old and new categories
		if originalSetting.Category != row.Category {
			r.invalidateCategorySettingsCache(context.Background(), originalSetting.TenantID, originalSetting.Category)
			r.invalidateCategorySettingsCache(context.Background(), row.TenantID, row.Category)
		} else {
			r.invalidateCategorySettingsCache(context.Background(), row.TenantID, row.Category)
		}

		// Invalidate scope cache for both old and new scopes
		if originalSetting.Scope != row.Scope {
			r.invalidateScopeSettingsCache(context.Background(), originalSetting.TenantID, originalSetting.Scope)
			r.invalidateScopeSettingsCache(context.Background(), row.TenantID, row.Scope)
		} else {
			r.invalidateScopeSettingsCache(context.Background(), row.TenantID, row.Scope)
		}
	}()

	return row, nil
}

// Delete deletes a tenant setting
func (r *tenantSettingRepository) Delete(ctx context.Context, id string) error {
	// Get setting first for cache invalidation
	setting, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Use master for writes
	if err := r.data.GetMasterEntClient().TenantSetting.DeleteOneID(id).Exec(ctx); err != nil {
		logger.Errorf(ctx, "tenantSettingRepo.Delete error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateSettingCache(context.Background(), setting)
		r.invalidateTenantSettingsCache(context.Background(), setting.TenantID)
		r.invalidateCategorySettingsCache(context.Background(), setting.TenantID, setting.Category)
		r.invalidateScopeSettingsCache(context.Background(), setting.TenantID, setting.Scope)
	}()

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

	// Cache settings in background
	go func() {
		for _, setting := range rows {
			r.cacheSetting(context.Background(), setting)
		}
	}()

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

// CountX counts tenant settings
func (r *tenantSettingRepository) CountX(ctx context.Context, params *structs.ListTenantSettingParams) int {
	builder := r.buildListQuery(params)
	return builder.CountX(ctx)
}

// buildListQuery builds the list query based on parameters
func (r *tenantSettingRepository) buildListQuery(params *structs.ListTenantSettingParams) *ent.TenantSettingQuery {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().TenantSetting.Query()

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

// cacheSetting caches a tenant setting
func (r *tenantSettingRepository) cacheSetting(ctx context.Context, setting *ent.TenantSetting) {
	// Cache by ID
	idKey := fmt.Sprintf("id:%s", setting.ID)
	if err := r.settingCache.Set(ctx, idKey, setting, r.settingTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache setting by ID %s: %v", setting.ID, err)
	}

	// Cache tenant:key to ID mapping
	tenantKeyKey := fmt.Sprintf("tenant_key:%s:%s", setting.TenantID, setting.SettingKey)
	if err := r.tenantKeySettingCache.Set(ctx, tenantKeyKey, &setting.ID, r.settingTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache tenant key mapping %s:%s: %v", setting.TenantID, setting.SettingKey, err)
	}
}

// invalidateSettingCache invalidates the cache for a tenant setting
func (r *tenantSettingRepository) invalidateSettingCache(ctx context.Context, setting *ent.TenantSetting) {
	// Invalidate ID cache
	idKey := fmt.Sprintf("id:%s", setting.ID)
	if err := r.settingCache.Delete(ctx, idKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate setting ID cache %s: %v", setting.ID, err)
	}

	// Invalidate tenant:key mapping
	tenantKeyKey := fmt.Sprintf("tenant_key:%s:%s", setting.TenantID, setting.SettingKey)
	if err := r.tenantKeySettingCache.Delete(ctx, tenantKeyKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate tenant key mapping cache %s:%s: %v", setting.TenantID, setting.SettingKey, err)
	}
}

// invalidateTenantSettingsCache invalidates the cache for tenant settings
func (r *tenantSettingRepository) invalidateTenantSettingsCache(ctx context.Context, tenantID string) {
	cacheKey := fmt.Sprintf("tenant_settings:%s", tenantID)
	if err := r.tenantSettingsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate tenant settings cache %s: %v", tenantID, err)
	}
}

// invalidateCategorySettingsCache invalidates the cache for category settings
func (r *tenantSettingRepository) invalidateCategorySettingsCache(ctx context.Context, tenantID, category string) {
	cacheKey := fmt.Sprintf("category_settings:%s:%s", tenantID, category)
	if err := r.categorySettingsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate category settings cache %s:%s: %v", tenantID, category, err)
	}
}

// invalidateScopeSettingsCache invalidates the cache for scope settings
func (r *tenantSettingRepository) invalidateScopeSettingsCache(ctx context.Context, tenantID, scope string) {
	cacheKey := fmt.Sprintf("scope_settings:%s:%s", tenantID, scope)
	if err := r.scopeSettingsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate scope settings cache %s:%s: %v", tenantID, scope, err)
	}
}

// getSettingIDByTenantAndKey gets the ID of a tenant setting by tenant ID and key
func (r *tenantSettingRepository) getSettingIDByTenantAndKey(ctx context.Context, tenantID, key string) (string, error) {
	cacheKey := fmt.Sprintf("tenant_key:%s:%s", tenantID, key)
	settingID, err := r.tenantKeySettingCache.Get(ctx, cacheKey)
	if err != nil || settingID == nil {
		return "", err
	}
	return *settingID, nil
}
