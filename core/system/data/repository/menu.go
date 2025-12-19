package repository

import (
	"context"
	"fmt"
	"ncobase/system/data"
	"ncobase/system/data/ent"
	menuEnt "ncobase/system/data/ent/menu"
	"ncobase/system/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/ncobase/ncore/data/search"
)

// MenuRepositoryInterface represents the menu repository interface.
type MenuRepositoryInterface interface {
	Create(context.Context, *structs.MenuBody) (*ent.Menu, error)
	GetMenuTree(context.Context, *structs.FindMenu) ([]*ent.Menu, error)
	Get(context.Context, *structs.FindMenu) (*ent.Menu, error)
	Update(context.Context, *structs.UpdateMenuBody) (*ent.Menu, error)
	Delete(context.Context, *structs.FindMenu) error
	List(context.Context, *structs.ListMenuParams) ([]*ent.Menu, error)
	ListWithCount(ctx context.Context, params *structs.ListMenuParams) ([]*ent.Menu, int, error)
	CountX(context.Context, *structs.ListMenuParams) int
}

// menuRepository implements the MenuRepositoryInterface.
type menuRepository struct {
	data             *data.Data
	menuCache        cache.ICache[ent.Menu]
	slugMappingCache cache.ICache[string]     // Maps slug to menu ID
	menuTreeCache    cache.ICache[[]ent.Menu] // Cache menu trees
	menuTTL          time.Duration
}

// NewMenuRepository creates a new menu repository.
func NewMenuRepository(d *data.Data) MenuRepositoryInterface {
	redisClient := d.GetRedis()

	return &menuRepository{
		data:             d,
		menuCache:        cache.NewCache[ent.Menu](redisClient, "ncse_system:menus"),
		slugMappingCache: cache.NewCache[string](redisClient, "ncse_system:menu_mappings"),
		menuTreeCache:    cache.NewCache[[]ent.Menu](redisClient, "ncse_system:menu_trees"),
		menuTTL:          time.Hour * 6, // 6 hours cache TTL
	}
}

// Create creates a new menu
func (r *menuRepository) Create(ctx context.Context, body *structs.MenuBody) (*ent.Menu, error) {
	// Use master for writes
	builder := r.data.GetMasterEntClient().Menu.Create()

	// Set all the fields
	if validator.IsNotEmpty(body.Name) {
		builder.SetNillableName(&body.Name)
	}
	if validator.IsNotEmpty(body.Label) {
		builder.SetNillableLabel(&body.Label)
	}
	if validator.IsNotEmpty(body.Slug) {
		builder.SetNillableSlug(&body.Slug)
	}
	if validator.IsNotEmpty(body.Type) {
		builder.SetNillableType(&body.Type)
	}
	if validator.IsNotEmpty(body.Path) {
		builder.SetNillablePath(&body.Path)
	}
	if validator.IsNotEmpty(body.Target) {
		builder.SetNillableTarget(&body.Target)
	}
	if validator.IsNotEmpty(body.Icon) {
		builder.SetNillableIcon(&body.Icon)
	}
	if validator.IsNotEmpty(body.Perms) {
		builder.SetNillablePerms(&body.Perms)
	}
	if validator.IsNotNil(body.Hidden) {
		builder.SetNillableHidden(body.Hidden)
	}
	if validator.IsNotNil(body.Order) {
		builder.SetNillableOrder(body.Order)
	}
	if validator.IsNotNil(body.Disabled) {
		builder.SetNillableDisabled(body.Disabled)
	}
	if validator.IsNotEmpty(body.ParentID) {
		builder.SetNillableParentID(&body.ParentID)
	}
	if validator.IsNotEmpty(body.CreatedBy) {
		builder.SetNillableCreatedBy(body.CreatedBy)
	}
	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	menu, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "menuRepo.Create error: %v", err)
		return nil, err
	}

	// Create the menu in Meilisearch index
	if err = r.data.IndexDocument(ctx, &search.IndexRequest{Index: "menus", Document: menu}); err != nil {
		logger.Errorf(ctx, "menuRepo.Create error creating Meilisearch index: %v", err)
	}

	// Cache the menu and invalidate tree cache
	go func() {
		r.cacheMenu(context.Background(), menu)
		r.invalidateMenuTreeCache(context.Background(), body.Type)
	}()

	return menu, nil
}

// GetMenuTree retrieves the menu tree
func (r *menuRepository) GetMenuTree(ctx context.Context, params *structs.FindMenu) ([]*ent.Menu, error) {
	// Generate cache key based on parameters
	cacheKey := r.generateTreeCacheKey(params)

	// Try cache first
	var cachedMenus []*ent.Menu
	if err := r.menuTreeCache.GetArray(ctx, cacheKey, &cachedMenus); err == nil && len(cachedMenus) > 0 {
		return cachedMenus, nil
	}

	// Fallback to database
	builder := r.data.GetSlaveEntClient().Menu.Query()

	// Apply type filter if specified
	if validator.IsNotEmpty(params.Type) {
		builder.Where(menuEnt.TypeEQ(params.Type))
	}

	// Handle specific menu/parent requests
	if validator.IsNotEmpty(params.Menu) && params.Menu != "root" {
		menus, err := r.getSubMenu(ctx, params.Menu, builder)
		if err != nil {
			return nil, err
		}

		// Cache the result
		go func() {
			if err := r.menuTreeCache.SetArray(context.Background(), cacheKey, menus, r.menuTTL); err != nil {
				logger.Debugf(context.Background(), "Failed to cache menu tree %s: %v", cacheKey, err)
			}
		}()

		return menus, nil
	}

	// For tree building, get all matching records
	menus, err := r.executeArrayQuery(ctx, builder)
	if err != nil {
		return nil, err
	}

	// Cache the result
	go func() {
		if err := r.menuTreeCache.SetArray(context.Background(), cacheKey, menus, r.menuTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache menu tree %s: %v", cacheKey, err)
		}
	}()

	return menus, nil
}

// Get retrieves a specific menu
func (r *menuRepository) Get(ctx context.Context, params *structs.FindMenu) (*ent.Menu, error) {
	// Try to get menu ID from slug mapping cache if searching by slug
	if params.Menu != "" {
		if menuID, err := r.getMenuIDBySlug(ctx, params.Menu); err == nil && menuID != "" {
			// Try to get from menu cache
			cacheKey := fmt.Sprintf("id:%s", menuID)
			if cached, err := r.menuCache.Get(ctx, cacheKey); err == nil && cached != nil {
				return cached, nil
			}
		}
	}

	// Fallback to database
	row, err := r.getMenu(ctx, params)
	if err != nil {
		logger.Errorf(ctx, "menuRepo.Get error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheMenu(context.Background(), row)

	return row, nil
}

// Update updates an existing menu
func (r *menuRepository) Update(ctx context.Context, body *structs.UpdateMenuBody) (*ent.Menu, error) {
	// Query the menu
	menu, err := r.getMenu(ctx, &structs.FindMenu{Menu: body.ID})
	if validator.IsNotNil(err) {
		return nil, err
	}

	// Use master for writes
	builder := menu.Update()

	// Set values
	if validator.IsNotEmpty(body.Name) {
		builder.SetNillableName(&body.Name)
	}
	if validator.IsNotEmpty(body.Label) {
		builder.SetNillableLabel(&body.Label)
	}
	if validator.IsNotEmpty(body.Slug) {
		builder.SetNillableSlug(&body.Slug)
	}
	if validator.IsNotEmpty(body.Type) {
		builder.SetNillableType(&body.Type)
	}
	if validator.IsNotEmpty(body.Path) {
		builder.SetNillablePath(&body.Path)
	}
	if validator.IsNotEmpty(body.Target) {
		builder.SetNillableTarget(&body.Target)
	}
	if validator.IsNotEmpty(body.Icon) {
		builder.SetNillableIcon(&body.Icon)
	}
	if validator.IsNotEmpty(body.Perms) {
		builder.SetNillablePerms(&body.Perms)
	}
	if validator.IsNotNil(body.Hidden) {
		builder.SetNillableHidden(body.Hidden)
	}
	if validator.IsNotNil(body.Order) {
		builder.SetNillableOrder(body.Order)
	}
	if validator.IsNotNil(body.Disabled) {
		builder.SetNillableDisabled(body.Disabled)
	}
	if validator.IsNotEmpty(body.ParentID) {
		builder.SetNillableParentID(&body.ParentID)
	}
	if validator.IsNotEmpty(body.UpdatedBy) {
		builder.SetNillableUpdatedBy(body.UpdatedBy)
	}
	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	updatedMenu, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "menuRepo.Update error: %v", err)
		return nil, err
	}

	// Update Meilisearch index
	if err = r.data.IndexDocument(ctx, &search.IndexRequest{Index: "menus", Document: updatedMenu, DocumentID: updatedMenu.ID}); err != nil {
		logger.Errorf(ctx, "menuRepo.Update error updating Meilisearch index: %v", err)
	}

	// Invalidate and re-cache
	go func() {
		r.invalidateMenuCache(context.Background(), menu)
		r.cacheMenu(context.Background(), updatedMenu)

		// Invalidate tree caches for affected types
		r.invalidateMenuTreeCache(context.Background(), menu.Type)
		if updatedMenu.Type != menu.Type {
			r.invalidateMenuTreeCache(context.Background(), updatedMenu.Type)
		}
	}()

	return updatedMenu, nil
}

// Delete deletes a menu
func (r *menuRepository) Delete(ctx context.Context, params *structs.FindMenu) error {
	// Get menu first for cache invalidation
	menu, err := r.getMenu(ctx, params)
	if err != nil {
		return err
	}

	// Use master for writes
	builder := r.data.GetMasterEntClient().Menu.Delete()
	builder.Where(menuEnt.Or(
		menuEnt.IDEQ(params.Menu),
		menuEnt.SlugEQ(params.Menu),
	))

	_, err = builder.Exec(ctx)
	if validator.IsNotNil(err) {
		return err
	}

	// Delete from Meilisearch index
	if err = r.data.DeleteDocument(ctx, "menus", menu.ID); err != nil {
		logger.Errorf(ctx, "menuRepo.Delete index error: %v", err)
	}

	// Invalidate cache
	go func() {
		r.invalidateMenuCache(context.Background(), menu)
		r.invalidateMenuTreeCache(context.Background(), menu.Type)
	}()

	return nil
}

// List returns a slice of menus based on the provided parameters
func (r *menuRepository) List(ctx context.Context, params *structs.ListMenuParams) ([]*ent.Menu, error) {
	builder, err := r.listBuilder(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("building list query: %w", err)
	}

	builder = r.applySorting(builder, params.SortBy)

	if params.Cursor != "" {
		id, value, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, fmt.Errorf("decoding cursor: %w", err)
		}

		if !nanoid.IsPrimaryKey(id) {
			return nil, fmt.Errorf("invalid id in cursor: %s", id)
		}

		builder = r.applyCursorCondition(builder, id, value, params.Direction, params.SortBy)
	}

	builder.Limit(params.Limit)

	menus, err := r.executeArrayQuery(ctx, builder)
	if err != nil {
		return nil, err
	}

	// Cache menus in background
	go func() {
		for _, menu := range menus {
			r.cacheMenu(context.Background(), menu)
		}
	}()

	return menus, nil
}

// CountX returns the total count of menus based on the provided parameters
func (r *menuRepository) CountX(ctx context.Context, params *structs.ListMenuParams) int {
	builder, err := r.listBuilder(ctx, params)
	if err != nil {
		logger.Errorf(ctx, "Error building count query: %v", err)
		return 0
	}
	return builder.CountX(ctx)
}

// ListWithCount returns both a slice of menus and the total count based on the provided parameters
func (r *menuRepository) ListWithCount(ctx context.Context, params *structs.ListMenuParams) ([]*ent.Menu, int, error) {
	builder, err := r.listBuilder(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("building list query: %w", err)
	}

	builder = r.applySorting(builder, params.SortBy)

	if params.Cursor != "" {
		id, value, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, 0, fmt.Errorf("decoding cursor: %w", err)
		}

		if !nanoid.IsPrimaryKey(id) {
			return nil, 0, fmt.Errorf("invalid id in cursor: %s", id)
		}

		builder = r.applyCursorCondition(builder, id, value, params.Direction, params.SortBy)
	}

	total, err := builder.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("counting menus: %w", err)
	}

	rows, err := builder.Limit(params.Limit).All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("fetching menus: %w", err)
	}

	// Cache menus in background
	go func() {
		for _, menu := range rows {
			r.cacheMenu(context.Background(), menu)
		}
	}()

	return rows, total, nil
}

// applySorting applies the specified sorting to the query builder.
func (r *menuRepository) applySorting(builder *ent.MenuQuery, sortBy string) *ent.MenuQuery {
	switch sortBy {
	case structs.SortByOrder:
		return builder.Order(
			ent.Desc(menuEnt.FieldOrder),
			ent.Desc(menuEnt.FieldCreatedAt),
			ent.Asc(menuEnt.FieldID),
		)
	case structs.SortByCreatedAt:
		return builder.Order(
			ent.Desc(menuEnt.FieldCreatedAt),
			ent.Desc(menuEnt.FieldOrder),
			ent.Asc(menuEnt.FieldID),
		)
	case structs.SortByName:
		return builder.Order(
			ent.Asc(menuEnt.FieldName),
			ent.Desc(menuEnt.FieldOrder),
			ent.Asc(menuEnt.FieldID),
		)
	default:
		return builder.Order(
			ent.Desc(menuEnt.FieldOrder),
			ent.Desc(menuEnt.FieldCreatedAt),
			ent.Asc(menuEnt.FieldID),
		)
	}
}

// applyCursorCondition applies the cursor-based condition to the query builder.
func (r *menuRepository) applyCursorCondition(builder *ent.MenuQuery, id string, value any, direction string, sortBy string) *ent.MenuQuery {
	switch sortBy {
	case structs.SortByCreatedAt:
		timestamp, ok := value.(int64)
		if !ok {
			logger.Errorf(context.Background(), "Invalid timestamp value for cursor")
			return builder
		}
		if direction == "backward" {
			return builder.Where(
				menuEnt.Or(
					menuEnt.CreatedAtGT(timestamp),
					menuEnt.And(
						menuEnt.CreatedAtEQ(timestamp),
						menuEnt.IDGT(id),
					),
				),
			)
		}
		return builder.Where(
			menuEnt.Or(
				menuEnt.CreatedAtLT(timestamp),
				menuEnt.And(
					menuEnt.CreatedAtEQ(timestamp),
					menuEnt.IDLT(id),
				),
			),
		)
	case structs.SortByOrder:
		order, ok := value.(int)
		if !ok {
			logger.Errorf(context.Background(), "Invalid order value for cursor")
			return builder
		}
		if direction == "backward" {
			return builder.Where(
				menuEnt.Or(
					menuEnt.OrderGT(order),
					menuEnt.And(
						menuEnt.OrderEQ(order),
						menuEnt.IDGT(id),
					),
				),
			)
		}
		return builder.Where(
			menuEnt.Or(
				menuEnt.OrderLT(order),
				menuEnt.And(
					menuEnt.OrderEQ(order),
					menuEnt.IDLT(id),
				),
			),
		)
	case structs.SortByName:
		name, ok := value.(string)
		if !ok {
			logger.Errorf(context.Background(), "Invalid name value for cursor")
			return builder
		}
		if direction == "backward" {
			return builder.Where(
				menuEnt.Or(
					menuEnt.NameGT(name),
					menuEnt.And(
						menuEnt.NameEQ(name),
						menuEnt.IDGT(id),
					),
				),
			)
		}
		return builder.Where(
			menuEnt.Or(
				menuEnt.NameLT(name),
				menuEnt.And(
					menuEnt.NameEQ(name),
					menuEnt.IDLT(id),
				),
			),
		)
	default:
		return r.applyCursorCondition(builder, id, value, direction, structs.SortByCreatedAt)
	}
}

// listBuilder - create list builder.
func (r *menuRepository) listBuilder(_ context.Context, params *structs.ListMenuParams) (*ent.MenuQuery, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().Menu.Query()

	// match type.
	if params.Type != "" {
		builder.Where(menuEnt.TypeEQ(params.Type))
	}

	// match permission.
	if params.Perms != "" {
		builder.Where(menuEnt.PermsContains(params.Perms))
	}

	if params.Parent != "" { // Explicit parent specified
		builder.Where(menuEnt.ParentIDEQ(params.Parent))
	} else {
		// Determine if we should apply root filter
		shouldApplyRootFilter := !params.Children && // Not building tree structure
			params.Type == "" // No specific type requested

		if shouldApplyRootFilter {
			// Only get root level items for general list queries
			builder.Where(menuEnt.Or(
				menuEnt.ParentIDIsNil(),
				menuEnt.ParentIDEQ(""),
				menuEnt.ParentIDEQ("root"),
			))
		}
		// If Children=true or Type is specified, get all records for tree building or type filtering
	}

	return builder, nil
}

// getMenu - get menu.
// internal method.
func (r *menuRepository) getMenu(ctx context.Context, params *structs.FindMenu) (*ent.Menu, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().Menu.Query()

	// set where conditions.
	if validator.IsNotEmpty(params.Menu) {
		builder.Where(menuEnt.Or(
			menuEnt.IDEQ(params.Menu),
			menuEnt.SlugEQ(params.Menu),
		))
	}

	// execute the builder.
	row, err := builder.First(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// getSubMenu - get sub menus.
func (r *menuRepository) getSubMenu(ctx context.Context, rootID string, builder *ent.MenuQuery) ([]*ent.Menu, error) {
	// set where conditions
	builder.Where(
		menuEnt.Or(
			menuEnt.ID(rootID),
			menuEnt.ParentIDHasPrefix(rootID),
		),
	)

	// execute the builder
	return r.executeArrayQuery(ctx, builder)
}

// executeArrayQuery - execute the builder query and return results.
func (r *menuRepository) executeArrayQuery(ctx context.Context, builder *ent.MenuQuery) ([]*ent.Menu, error) {
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "menuRepo.executeArrayQuery error: %v", err)
		return nil, err
	}
	return rows, nil
}

// cacheMenu caches a menu
func (r *menuRepository) cacheMenu(ctx context.Context, menu *ent.Menu) {
	// Cache by ID
	idKey := fmt.Sprintf("id:%s", menu.ID)
	if err := r.menuCache.Set(ctx, idKey, menu, r.menuTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache menu by ID %s: %v", menu.ID, err)
	}

	// Cache slug to ID mapping
	if menu.Slug != "" {
		slugKey := fmt.Sprintf("slug:%s", menu.Slug)
		if err := r.slugMappingCache.Set(ctx, slugKey, &menu.ID, r.menuTTL); err != nil {
			logger.Debugf(ctx, "Failed to cache slug mapping %s: %v", menu.Slug, err)
		}
	}
}

// invalidateMenuCache invalidates a menu cache
func (r *menuRepository) invalidateMenuCache(ctx context.Context, menu *ent.Menu) {
	// Invalidate ID cache
	idKey := fmt.Sprintf("id:%s", menu.ID)
	if err := r.menuCache.Delete(ctx, idKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate menu ID cache %s: %v", menu.ID, err)
	}

	// Invalidate slug mapping
	if menu.Slug != "" {
		slugKey := fmt.Sprintf("slug:%s", menu.Slug)
		if err := r.slugMappingCache.Delete(ctx, slugKey); err != nil {
			logger.Debugf(ctx, "Failed to invalidate slug mapping cache %s: %v", menu.Slug, err)
		}
	}
}

// invalidateMenuTreeCache invalidates a menu tree cache
func (r *menuRepository) invalidateMenuTreeCache(ctx context.Context, menuType string) {
	// Invalidate various tree cache combinations
	keys := []string{
		"tree:all",
		fmt.Sprintf("tree:type:%s", menuType),
	}

	for _, key := range keys {
		if err := r.menuTreeCache.Delete(ctx, key); err != nil {
			logger.Debugf(ctx, "Failed to invalidate menu tree cache %s: %v", key, err)
		}
	}
}

// generateTreeCacheKey generates a cache key for the menu tree
func (r *menuRepository) generateTreeCacheKey(params *structs.FindMenu) string {
	if params.Type != "" {
		return fmt.Sprintf("tree:type:%s", params.Type)
	}
	return "tree:all"
}

// getMenuIDBySlug gets a menu ID by slug
func (r *menuRepository) getMenuIDBySlug(ctx context.Context, slug string) (string, error) {
	cacheKey := fmt.Sprintf("slug:%s", slug)
	menuID, err := r.slugMappingCache.Get(ctx, cacheKey)
	if err != nil || menuID == nil {
		return "", err
	}
	return *menuID, nil
}
