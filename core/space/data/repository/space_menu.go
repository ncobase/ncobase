package repository

import (
	"context"
	"fmt"
	"ncobase/space/data"
	"ncobase/space/data/ent"
	spaceMenuEnt "ncobase/space/data/ent/spacemenu"
	"ncobase/space/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/logging/logger"
)

// SpaceMenuRepositoryInterface represents the space menu repository interface.
type SpaceMenuRepositoryInterface interface {
	Create(ctx context.Context, body *structs.SpaceMenu) (*ent.SpaceMenu, error)
	GetBySpaceID(ctx context.Context, spaceID string) ([]*ent.SpaceMenu, error)
	GetByMenuID(ctx context.Context, menuID string) ([]*ent.SpaceMenu, error)
	DeleteBySpaceIDAndMenuID(ctx context.Context, spaceID, menuID string) error
	DeleteAllBySpaceID(ctx context.Context, spaceID string) error
	DeleteAllByMenuID(ctx context.Context, menuID string) error
	IsMenuInSpace(ctx context.Context, spaceID, menuID string) (bool, error)
	GetSpaceMenus(ctx context.Context, spaceID string) ([]string, error)
}

// spaceMenuRepository implements the SpaceMenuRepositoryInterface.
type spaceMenuRepository struct {
	data            *data.Data
	spaceMenuCache  cache.ICache[ent.SpaceMenu]
	spaceMenusCache cache.ICache[[]string] // Maps space to menu IDs
	menuSpacesCache cache.ICache[[]string] // Maps menu to space IDs
	relationshipTTL time.Duration
}

// NewSpaceMenuRepository creates a new space menu repository.
func NewSpaceMenuRepository(d *data.Data) SpaceMenuRepositoryInterface {
	redisClient := d.GetRedis()

	return &spaceMenuRepository{
		data:            d,
		spaceMenuCache:  cache.NewCache[ent.SpaceMenu](redisClient, "ncse_space:space_menus"),
		spaceMenusCache: cache.NewCache[[]string](redisClient, "ncse_space:space_menu_mappings"),
		menuSpacesCache: cache.NewCache[[]string](redisClient, "ncse_space:menu_space_mappings"),
		relationshipTTL: time.Hour * 2,
	}
}

// Create creates a new space menu relationship.
func (r *spaceMenuRepository) Create(ctx context.Context, body *structs.SpaceMenu) (*ent.SpaceMenu, error) {
	builder := r.data.GetMasterEntClient().SpaceMenu.Create()

	builder.SetNillableSpaceID(&body.SpaceID)
	builder.SetNillableMenuID(&body.MenuID)

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceMenuRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the relationship and invalidate related caches
	go func() {
		r.cacheSpaceMenu(context.Background(), row)
		r.invalidateSpaceMenusCache(context.Background(), body.SpaceID)
		r.invalidateMenuSpacesCache(context.Background(), body.MenuID)
	}()

	return row, nil
}

// GetBySpaceID retrieves space menus by space ID.
func (r *spaceMenuRepository) GetBySpaceID(ctx context.Context, spaceID string) ([]*ent.SpaceMenu, error) {
	builder := r.data.GetSlaveEntClient().SpaceMenu.Query()
	builder.Where(spaceMenuEnt.SpaceIDEQ(spaceID))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceMenuRepo.GetBySpaceID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, tm := range rows {
			r.cacheSpaceMenu(context.Background(), tm)
		}
	}()

	return rows, nil
}

// GetByMenuID retrieves space menus by menu ID.
func (r *spaceMenuRepository) GetByMenuID(ctx context.Context, menuID string) ([]*ent.SpaceMenu, error) {
	builder := r.data.GetSlaveEntClient().SpaceMenu.Query()
	builder.Where(spaceMenuEnt.MenuIDEQ(menuID))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceMenuRepo.GetByMenuID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, tm := range rows {
			r.cacheSpaceMenu(context.Background(), tm)
		}
	}()

	return rows, nil
}

// DeleteBySpaceIDAndMenuID deletes space menu by space ID and menu ID.
func (r *spaceMenuRepository) DeleteBySpaceIDAndMenuID(ctx context.Context, spaceID, menuID string) error {
	if _, err := r.data.GetMasterEntClient().SpaceMenu.Delete().
		Where(spaceMenuEnt.SpaceIDEQ(spaceID), spaceMenuEnt.MenuIDEQ(menuID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "spaceMenuRepo.DeleteBySpaceIDAndMenuID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateSpaceMenuCache(context.Background(), spaceID, menuID)
		r.invalidateSpaceMenusCache(context.Background(), spaceID)
		r.invalidateMenuSpacesCache(context.Background(), menuID)
	}()

	return nil
}

// DeleteAllBySpaceID deletes all space menus by space ID.
func (r *spaceMenuRepository) DeleteAllBySpaceID(ctx context.Context, spaceID string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().SpaceMenu.Query().
		Where(spaceMenuEnt.SpaceIDEQ(spaceID)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	if _, err := r.data.GetMasterEntClient().SpaceMenu.Delete().
		Where(spaceMenuEnt.SpaceIDEQ(spaceID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "spaceMenuRepo.DeleteAllBySpaceID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateSpaceMenusCache(context.Background(), spaceID)
		for _, tm := range relationships {
			r.invalidateSpaceMenuCache(context.Background(), tm.SpaceID, tm.MenuID)
			r.invalidateMenuSpacesCache(context.Background(), tm.MenuID)
		}
	}()

	return nil
}

// DeleteAllByMenuID deletes all space menus by menu ID.
func (r *spaceMenuRepository) DeleteAllByMenuID(ctx context.Context, menuID string) error {
	// Get existing relationships for cache invalidation
	relationships, err := r.data.GetSlaveEntClient().SpaceMenu.Query().
		Where(spaceMenuEnt.MenuIDEQ(menuID)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	if _, err := r.data.GetMasterEntClient().SpaceMenu.Delete().
		Where(spaceMenuEnt.MenuIDEQ(menuID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "spaceMenuRepo.DeleteAllByMenuID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateMenuSpacesCache(context.Background(), menuID)
		for _, tm := range relationships {
			r.invalidateSpaceMenuCache(context.Background(), tm.SpaceID, tm.MenuID)
			r.invalidateSpaceMenusCache(context.Background(), tm.SpaceID)
		}
	}()

	return nil
}

// IsMenuInSpace verifies if a menu belongs to a space.
func (r *spaceMenuRepository) IsMenuInSpace(ctx context.Context, spaceID, menuID string) (bool, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("relationship:%s:%s", spaceID, menuID)
	if cached, err := r.spaceMenuCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return true, nil
	}

	count, err := r.data.GetSlaveEntClient().SpaceMenu.Query().
		Where(spaceMenuEnt.SpaceIDEQ(spaceID), spaceMenuEnt.MenuIDEQ(menuID)).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceMenuRepo.IsMenuInSpace error: %v", err)
		return false, err
	}

	exists := count > 0

	// Cache the result if it exists
	if exists {
		go func() {
			relationship := &ent.SpaceMenu{
				SpaceID: spaceID,
				MenuID:  menuID,
			}
			r.cacheSpaceMenu(context.Background(), relationship)
		}()
	}

	return exists, nil
}

// GetSpaceMenus retrieves all menu IDs for a space.
func (r *spaceMenuRepository) GetSpaceMenus(ctx context.Context, spaceID string) ([]string, error) {
	// Try to get menu IDs from cache
	cacheKey := fmt.Sprintf("space_menus:%s", spaceID)
	var menuIDs []string
	if err := r.spaceMenusCache.GetArray(ctx, cacheKey, &menuIDs); err == nil && len(menuIDs) > 0 {
		return menuIDs, nil
	}

	// Fallback to database
	spaceMenus, err := r.data.GetSlaveEntClient().SpaceMenu.Query().
		Where(spaceMenuEnt.SpaceIDEQ(spaceID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceMenuRepo.GetSpaceMenus error: %v", err)
		return nil, err
	}

	// Extract menu IDs
	menuIDs = make([]string, len(spaceMenus))
	for i, tm := range spaceMenus {
		menuIDs[i] = tm.MenuID
	}

	// Cache menu IDs for future use
	go func() {
		if err := r.spaceMenusCache.SetArray(context.Background(), cacheKey, menuIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache space menus %s: %v", spaceID, err)
		}
	}()

	return menuIDs, nil
}

// cacheSpaceMenu caches a space menu relationship.
func (r *spaceMenuRepository) cacheSpaceMenu(ctx context.Context, tm *ent.SpaceMenu) {
	relationshipKey := fmt.Sprintf("relationship:%s:%s", tm.SpaceID, tm.MenuID)
	if err := r.spaceMenuCache.Set(ctx, relationshipKey, tm, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache space menu relationship %s:%s: %v", tm.SpaceID, tm.MenuID, err)
	}
}

// invalidateSpaceMenuCache invalidates space menu cache
func (r *spaceMenuRepository) invalidateSpaceMenuCache(ctx context.Context, spaceID, menuID string) {
	relationshipKey := fmt.Sprintf("relationship:%s:%s", spaceID, menuID)
	if err := r.spaceMenuCache.Delete(ctx, relationshipKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate space menu relationship cache %s:%s: %v", spaceID, menuID, err)
	}
}

// invalidateSpaceMenusCache invalidates space menus cache
func (r *spaceMenuRepository) invalidateSpaceMenusCache(ctx context.Context, spaceID string) {
	cacheKey := fmt.Sprintf("space_menus:%s", spaceID)
	if err := r.spaceMenusCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate space menus cache %s: %v", spaceID, err)
	}
}

// invalidateMenuSpacesCache invalidates menu spaces cache
func (r *spaceMenuRepository) invalidateMenuSpacesCache(ctx context.Context, menuID string) {
	cacheKey := fmt.Sprintf("menu_spaces:%s", menuID)
	if err := r.menuSpacesCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate menu spaces cache %s: %v", menuID, err)
	}
}
