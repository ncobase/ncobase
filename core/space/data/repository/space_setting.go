package repository

import (
	"context"
	"fmt"
	"ncobase/core/space/data"
	"ncobase/core/space/data/ent"
	spaceSettingEnt "ncobase/core/space/data/ent/spacesetting"
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

// SpaceSettingRepositoryInterface defines the interface for space setting repository
type SpaceSettingRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateSpaceSettingBody) (*ent.SpaceSetting, error)
	GetByID(ctx context.Context, id string) (*ent.SpaceSetting, error)
	GetByKey(ctx context.Context, spaceID, key string) (*ent.SpaceSetting, error)
	GetBySpaceID(ctx context.Context, spaceID string) ([]*ent.SpaceSetting, error)
	GetByCategory(ctx context.Context, spaceID, category string) ([]*ent.SpaceSetting, error)
	Update(ctx context.Context, id string, updates types.JSON) (*ent.SpaceSetting, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params *structs.ListSpaceSettingParams) ([]*ent.SpaceSetting, error)
	ListWithCount(ctx context.Context, params *structs.ListSpaceSettingParams) ([]*ent.SpaceSetting, int, error)
	CountX(ctx context.Context, params *structs.ListSpaceSettingParams) int
}

// spaceSettingRepository implements SpaceSettingRepositoryInterface
type spaceSettingRepository struct {
	data                  *data.Data
	settingCache          cache.ICache[ent.SpaceSetting]
	spaceSettingsCache    cache.ICache[[]string] // Maps space ID to setting IDs
	spaceKeySettingCache  cache.ICache[string]   // Maps space:key to setting ID
	categorySettingsCache cache.ICache[[]string] // Maps space:category to setting IDs
	scopeSettingsCache    cache.ICache[[]string] // Maps space:scope to setting IDs
	settingTTL            time.Duration
}

// NewSpaceSettingRepository creates a new space setting repository
func NewSpaceSettingRepository(d *data.Data) SpaceSettingRepositoryInterface {
	redisClient := d.GetRedis().(*redis.Client)

	return &spaceSettingRepository{
		data:                  d,
		settingCache:          cache.NewCache[ent.SpaceSetting](redisClient, "ncse_space:settings"),
		spaceSettingsCache:    cache.NewCache[[]string](redisClient, "ncse_space:space_setting_mappings"),
		spaceKeySettingCache:  cache.NewCache[string](redisClient, "ncse_space:space_key_setting_mappings"),
		categorySettingsCache: cache.NewCache[[]string](redisClient, "ncse_space:category_setting_mappings"),
		scopeSettingsCache:    cache.NewCache[[]string](redisClient, "ncse_space:scope_setting_mappings"),
		settingTTL:            time.Hour * 6, // 6 hours cache TTL (settings change less frequently)
	}
}

// Create creates a new space setting
func (r *spaceSettingRepository) Create(ctx context.Context, body *structs.CreateSpaceSettingBody) (*ent.SpaceSetting, error) {
	// Use master for writes
	builder := r.data.GetMasterEntClient().SpaceSetting.Create()

	builder.SetSpaceID(body.SpaceID)
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
		logger.Errorf(ctx, "spaceSettingRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the setting and invalidate related caches
	go func() {
		r.cacheSetting(context.Background(), row)
		r.invalidateSpaceSettingsCache(context.Background(), body.SpaceID)
		r.invalidateCategorySettingsCache(context.Background(), body.SpaceID, body.Category)
		r.invalidateScopeSettingsCache(context.Background(), body.SpaceID, string(body.Scope))
	}()

	return row, nil
}

// GetByID retrieves a space setting by ID
func (r *spaceSettingRepository) GetByID(ctx context.Context, id string) (*ent.SpaceSetting, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("id:%s", id)
	if cached, err := r.settingCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Use slave for reads
	row, err := r.data.GetSlaveEntClient().SpaceSetting.Get(ctx, id)
	if err != nil {
		logger.Errorf(ctx, "spaceSettingRepo.GetByID error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheSetting(context.Background(), row)

	return row, nil
}

// GetByKey retrieves a space setting by space ID and key
func (r *spaceSettingRepository) GetByKey(ctx context.Context, spaceID, key string) (*ent.SpaceSetting, error) {
	// Try to get setting ID from space:key mapping cache
	if settingID, err := r.getSettingIDBySpaceAndKey(ctx, spaceID, key); err == nil && settingID != "" {
		return r.GetByID(ctx, settingID)
	}

	// Fallback to database
	row, err := r.data.GetSlaveEntClient().SpaceSetting.Query().
		Where(
			spaceSettingEnt.SpaceIDEQ(spaceID),
			spaceSettingEnt.SettingKeyEQ(key),
		).
		Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceSettingRepo.GetByKey error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheSetting(context.Background(), row)

	return row, nil
}

// GetBySpaceID retrieves all settings for a space
func (r *spaceSettingRepository) GetBySpaceID(ctx context.Context, spaceID string) ([]*ent.SpaceSetting, error) {
	// Try to get setting IDs from cache
	cacheKey := fmt.Sprintf("space_settings:%s", spaceID)
	var settingIDs []string
	if err := r.spaceSettingsCache.GetArray(ctx, cacheKey, &settingIDs); err == nil && len(settingIDs) > 0 {
		// Get settings by IDs
		settings := make([]*ent.SpaceSetting, 0, len(settingIDs))
		for _, settingID := range settingIDs {
			if setting, err := r.GetByID(ctx, settingID); err == nil {
				settings = append(settings, setting)
			}
		}
		return settings, nil
	}

	// Fallback to database
	rows, err := r.data.GetSlaveEntClient().SpaceSetting.Query().
		Where(spaceSettingEnt.SpaceIDEQ(spaceID)).
		All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceSettingRepo.GetBySpaceID error: %v", err)
		return nil, err
	}

	// Cache settings and setting IDs
	go func() {
		settingIDs := make([]string, 0, len(rows))
		for _, setting := range rows {
			r.cacheSetting(context.Background(), setting)
			settingIDs = append(settingIDs, setting.ID)
		}

		if err := r.spaceSettingsCache.SetArray(context.Background(), cacheKey, settingIDs, r.settingTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache space settings %s: %v", spaceID, err)
		}
	}()

	return rows, nil
}

// GetByCategory retrieves settings by space ID and category
func (r *spaceSettingRepository) GetByCategory(ctx context.Context, spaceID, category string) ([]*ent.SpaceSetting, error) {
	// Try to get setting IDs from cache
	cacheKey := fmt.Sprintf("category_settings:%s:%s", spaceID, category)
	var settingIDs []string
	if err := r.categorySettingsCache.GetArray(ctx, cacheKey, &settingIDs); err == nil && len(settingIDs) > 0 {
		// Get settings by IDs
		settings := make([]*ent.SpaceSetting, 0, len(settingIDs))
		for _, settingID := range settingIDs {
			if setting, err := r.GetByID(ctx, settingID); err == nil {
				settings = append(settings, setting)
			}
		}
		return settings, nil
	}

	// Fallback to database
	rows, err := r.data.GetSlaveEntClient().SpaceSetting.Query().
		Where(
			spaceSettingEnt.SpaceIDEQ(spaceID),
			spaceSettingEnt.CategoryEQ(category),
		).
		All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceSettingRepo.GetByCategory error: %v", err)
		return nil, err
	}

	// Cache settings and setting IDs
	go func() {
		settingIDs := make([]string, 0, len(rows))
		for _, setting := range rows {
			r.cacheSetting(context.Background(), setting)
			settingIDs = append(settingIDs, setting.ID)
		}

		if err := r.categorySettingsCache.SetArray(context.Background(), cacheKey, settingIDs, r.settingTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache category settings %s:%s: %v", spaceID, category, err)
		}
	}()

	return rows, nil
}

// Update updates a space setting
func (r *spaceSettingRepository) Update(ctx context.Context, id string, updates types.JSON) (*ent.SpaceSetting, error) {
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
		logger.Errorf(ctx, "spaceSettingRepo.Update error: %v", err)
		return nil, err
	}

	// Invalidate and re-cache
	go func() {
		r.invalidateSettingCache(context.Background(), originalSetting)
		r.cacheSetting(context.Background(), row)

		// Invalidate related caches
		r.invalidateSpaceSettingsCache(context.Background(), row.SpaceID)

		// Invalidate category cache for both old and new categories
		if originalSetting.Category != row.Category {
			r.invalidateCategorySettingsCache(context.Background(), originalSetting.SpaceID, originalSetting.Category)
			r.invalidateCategorySettingsCache(context.Background(), row.SpaceID, row.Category)
		} else {
			r.invalidateCategorySettingsCache(context.Background(), row.SpaceID, row.Category)
		}

		// Invalidate scope cache for both old and new scopes
		if originalSetting.Scope != row.Scope {
			r.invalidateScopeSettingsCache(context.Background(), originalSetting.SpaceID, originalSetting.Scope)
			r.invalidateScopeSettingsCache(context.Background(), row.SpaceID, row.Scope)
		} else {
			r.invalidateScopeSettingsCache(context.Background(), row.SpaceID, row.Scope)
		}
	}()

	return row, nil
}

// Delete deletes a space setting
func (r *spaceSettingRepository) Delete(ctx context.Context, id string) error {
	// Get setting first for cache invalidation
	setting, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Use master for writes
	if err := r.data.GetMasterEntClient().SpaceSetting.DeleteOneID(id).Exec(ctx); err != nil {
		logger.Errorf(ctx, "spaceSettingRepo.Delete error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateSettingCache(context.Background(), setting)
		r.invalidateSpaceSettingsCache(context.Background(), setting.SpaceID)
		r.invalidateCategorySettingsCache(context.Background(), setting.SpaceID, setting.Category)
		r.invalidateScopeSettingsCache(context.Background(), setting.SpaceID, setting.Scope)
	}()

	return nil
}

// List lists space settings
func (r *spaceSettingRepository) List(ctx context.Context, params *structs.ListSpaceSettingParams) ([]*ent.SpaceSetting, error) {
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
			builder.Where(spaceSettingEnt.Or(
				spaceSettingEnt.CreatedAtGT(timestamp),
				spaceSettingEnt.And(
					spaceSettingEnt.CreatedAtEQ(timestamp),
					spaceSettingEnt.IDGT(id),
				),
			))
		} else {
			builder.Where(spaceSettingEnt.Or(
				spaceSettingEnt.CreatedAtLT(timestamp),
				spaceSettingEnt.And(
					spaceSettingEnt.CreatedAtEQ(timestamp),
					spaceSettingEnt.IDLT(id),
				),
			))
		}
	}

	if params.Direction == "backward" {
		builder.Order(ent.Asc(spaceSettingEnt.FieldCreatedAt), ent.Asc(spaceSettingEnt.FieldID))
	} else {
		builder.Order(ent.Desc(spaceSettingEnt.FieldCreatedAt), ent.Desc(spaceSettingEnt.FieldID))
	}

	builder.Limit(params.Limit)

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceSettingRepo.List error: %v", err)
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

// ListWithCount lists space settings with count
func (r *spaceSettingRepository) ListWithCount(ctx context.Context, params *structs.ListSpaceSettingParams) ([]*ent.SpaceSetting, int, error) {
	builder := r.buildListQuery(params)

	total, err := builder.Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceSettingRepo.ListWithCount count error: %v", err)
		return nil, 0, err
	}

	rows, err := r.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

// CountX counts space settings
func (r *spaceSettingRepository) CountX(ctx context.Context, params *structs.ListSpaceSettingParams) int {
	builder := r.buildListQuery(params)
	return builder.CountX(ctx)
}

// buildListQuery builds the list query based on parameters
func (r *spaceSettingRepository) buildListQuery(params *structs.ListSpaceSettingParams) *ent.SpaceSettingQuery {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().SpaceSetting.Query()

	if params.SpaceID != "" {
		builder.Where(spaceSettingEnt.SpaceIDEQ(params.SpaceID))
	}

	if params.Category != "" {
		builder.Where(spaceSettingEnt.CategoryEQ(params.Category))
	}

	if params.Scope != "" {
		builder.Where(spaceSettingEnt.ScopeEQ(string(params.Scope)))
	}

	if params.IsPublic != nil {
		builder.Where(spaceSettingEnt.IsPublicEQ(*params.IsPublic))
	}

	if params.IsRequired != nil {
		builder.Where(spaceSettingEnt.IsRequiredEQ(*params.IsRequired))
	}

	return builder
}

// cacheSetting caches a space setting
func (r *spaceSettingRepository) cacheSetting(ctx context.Context, setting *ent.SpaceSetting) {
	// Cache by ID
	idKey := fmt.Sprintf("id:%s", setting.ID)
	if err := r.settingCache.Set(ctx, idKey, setting, r.settingTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache setting by ID %s: %v", setting.ID, err)
	}

	// Cache space:key to ID mapping
	spaceKeyKey := fmt.Sprintf("space_key:%s:%s", setting.SpaceID, setting.SettingKey)
	if err := r.spaceKeySettingCache.Set(ctx, spaceKeyKey, &setting.ID, r.settingTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache space key mapping %s:%s: %v", setting.SpaceID, setting.SettingKey, err)
	}
}

// invalidateSettingCache invalidates the cache for a space setting
func (r *spaceSettingRepository) invalidateSettingCache(ctx context.Context, setting *ent.SpaceSetting) {
	// Invalidate ID cache
	idKey := fmt.Sprintf("id:%s", setting.ID)
	if err := r.settingCache.Delete(ctx, idKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate setting ID cache %s: %v", setting.ID, err)
	}

	// Invalidate space:key mapping
	spaceKeyKey := fmt.Sprintf("space_key:%s:%s", setting.SpaceID, setting.SettingKey)
	if err := r.spaceKeySettingCache.Delete(ctx, spaceKeyKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate space key mapping cache %s:%s: %v", setting.SpaceID, setting.SettingKey, err)
	}
}

// invalidateSpaceSettingsCache invalidates the cache for space settings
func (r *spaceSettingRepository) invalidateSpaceSettingsCache(ctx context.Context, spaceID string) {
	cacheKey := fmt.Sprintf("space_settings:%s", spaceID)
	if err := r.spaceSettingsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate space settings cache %s: %v", spaceID, err)
	}
}

// invalidateCategorySettingsCache invalidates the cache for category settings
func (r *spaceSettingRepository) invalidateCategorySettingsCache(ctx context.Context, spaceID, category string) {
	cacheKey := fmt.Sprintf("category_settings:%s:%s", spaceID, category)
	if err := r.categorySettingsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate category settings cache %s:%s: %v", spaceID, category, err)
	}
}

// invalidateScopeSettingsCache invalidates the cache for scope settings
func (r *spaceSettingRepository) invalidateScopeSettingsCache(ctx context.Context, spaceID, scope string) {
	cacheKey := fmt.Sprintf("scope_settings:%s:%s", spaceID, scope)
	if err := r.scopeSettingsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate scope settings cache %s:%s: %v", spaceID, scope, err)
	}
}

// getSettingIDBySpaceAndKey gets the ID of a space setting by space ID and key
func (r *spaceSettingRepository) getSettingIDBySpaceAndKey(ctx context.Context, spaceID, key string) (string, error) {
	cacheKey := fmt.Sprintf("space_key:%s:%s", spaceID, key)
	settingID, err := r.spaceKeySettingCache.Get(ctx, cacheKey)
	if err != nil || settingID == nil {
		return "", err
	}
	return *settingID, nil
}
