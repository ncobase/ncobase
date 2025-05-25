package repository

import (
	"context"
	"fmt"
	"ncobase/system/data"
	"ncobase/system/data/ent"
	menuEnt "ncobase/system/data/ent/menu"
	"ncobase/system/structs"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/data/search/meili"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/redis/go-redis/v9"
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
	ec *ent.Client
	rc *redis.Client
	ms *meili.Client
	c  *cache.Cache[ent.Menu]
}

// NewMenuRepository creates a new menu repository.
func NewMenuRepository(d *data.Data) MenuRepositoryInterface {
	ec := d.GetMasterEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &menuRepository{
		ec: ec,
		rc: rc,
		ms: ms,
		c:  cache.NewCache[ent.Menu](rc, "ncse_menu", false),
	}
}

// Create creates a new menu.
func (r *menuRepository) Create(ctx context.Context, body *structs.MenuBody) (*ent.Menu, error) {
	// create builder
	builder := r.ec.Menu.Create()

	// set values
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
	if validator.IsNotEmpty(body.TenantID) {
		builder.SetNillableTenantID(&body.TenantID)
	}
	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "menuRepo.Create error: %v", err)
		return nil, err
	}

	// Create the menu in Meilisearch index
	if err = r.ms.IndexDocuments("menus", row); err != nil {
		logger.Errorf(ctx, "menuRepo.Create error creating Meilisearch index: %v", err)
	}

	// delete cached menu tree
	// _ = r.c.Delete(ctx, cache.Key("menu=tree"))

	return row, nil
}

// GetMenuTree retrieves the menu tree.
func (r *menuRepository) GetMenuTree(ctx context.Context, params *structs.FindMenu) ([]*ent.Menu, error) {
	// create builder
	builder := r.ec.Menu.Query()

	// Apply type filter if specified
	if validator.IsNotEmpty(params.Type) {
		builder.Where(menuEnt.TypeEQ(params.Type))
	}

	// Apply tenant filter if specified
	if validator.IsNotEmpty(params.Tenant) {
		builder.Where(menuEnt.TenantIDEQ(params.Tenant))
	}

	// Handle specific menu/parent requests
	if validator.IsNotEmpty(params.Menu) && params.Menu != "root" {
		return r.getSubMenu(ctx, params.Menu, builder)
	}

	// For tree building, get all matching records
	// Don't apply parent filtering here as we need all records to build the tree
	return r.executeArrayQuery(ctx, builder)
}

// Get retrieves a specific menu.
func (r *menuRepository) Get(ctx context.Context, params *structs.FindMenu) (*ent.Menu, error) {
	cacheKey := fmt.Sprintf("%s", params.Menu)

	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	row, err := r.getMenu(ctx, params)
	if err != nil {
		logger.Errorf(ctx, "menuRepo.Get error: %v", err)
		return nil, err
	}

	if err := r.c.Set(ctx, cacheKey, row); err != nil {
		logger.Errorf(ctx, "menuRepo.Get cache error: %v", err)
	}

	return row, nil
}

// Update updates an existing menu.
func (r *menuRepository) Update(ctx context.Context, body *structs.UpdateMenuBody) (*ent.Menu, error) {
	// query the menu.
	row, err := r.getMenu(ctx, &structs.FindMenu{
		Menu: body.ID,
	})
	if validator.IsNotNil(err) {
		return nil, err
	}

	// create builder.
	builder := row.Update()

	// set values
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
	if validator.IsNotEmpty(body.TenantID) {
		builder.SetNillableTenantID(&body.TenantID)
	}
	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	row, err = builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "menuRepo.Update error: %v", err)
		return nil, err
	}

	// update cache
	cacheKey := fmt.Sprintf("%s", row.ID)
	if err := r.c.Set(ctx, cacheKey, row); err != nil {
		logger.Errorf(ctx, "menuRepo.Update cache error: %v", err)
	}

	// delete menu tree cache
	// if err := r.c.Delete(ctx, "menu=tree"); err != nil {
	// 	log.Errorf(ctx, "menuRepo.Update cache error: %v", err)
	// }

	return row, nil
}

// Delete deletes a menu.
func (r *menuRepository) Delete(ctx context.Context, params *structs.FindMenu) error {
	// create builder.
	builder := r.ec.Menu.Delete()

	// set where conditions.
	builder.Where(menuEnt.Or(
		menuEnt.IDEQ(params.Menu),
		menuEnt.SlugEQ(params.Menu),
	))

	// match tenant id.
	if validator.IsNotEmpty(params.Tenant) {
		builder.Where(menuEnt.TenantIDEQ(params.Tenant))
	}

	// execute the builder.
	_, err := builder.Exec(ctx)
	if validator.IsNotNil(err) {
		return err
	}

	cacheKey := fmt.Sprintf("%s", params.Menu)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		logger.Errorf(ctx, "menuRepo.Delete cache error: %v", err)
	}

	return nil
}

// List returns a slice of menus based on the provided parameters.
func (r *menuRepository) List(ctx context.Context, params *structs.ListMenuParams) ([]*ent.Menu, error) {
	builder, err := r.listBuilder(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("building list query: %w", err)
	}

	builder = menuSorting(builder, params.SortBy)

	if params.Cursor != "" {
		id, value, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, fmt.Errorf("decoding cursor: %w", err)
		}
		builder = menuCondition(builder, id, value, params.Direction, params.SortBy)
	}

	builder.Limit(params.Limit)

	return r.executeArrayQuery(ctx, builder)
}

// CountX returns the total count of menus based on the provided parameters.
func (r *menuRepository) CountX(ctx context.Context, params *structs.ListMenuParams) int {
	builder, err := r.listBuilder(ctx, params)
	if err != nil {
		logger.Errorf(ctx, "Error building count query: %v", err)
		return 0
	}
	return builder.CountX(ctx)
}

// ListWithCount returns both a slice of menus and the total count based on the provided parameters.
func (r *menuRepository) ListWithCount(ctx context.Context, params *structs.ListMenuParams) ([]*ent.Menu, int, error) {
	builder, err := r.listBuilder(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("building list query: %w", err)
	}

	builder = menuSorting(builder, params.SortBy)

	if params.Cursor != "" {
		id, value, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, 0, fmt.Errorf("decoding cursor: %w", err)
		}
		builder = menuCondition(builder, id, value, params.Direction, params.SortBy)
	}

	total, err := builder.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("counting menus: %w", err)
	}

	rows, err := builder.Limit(params.Limit).All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("fetching menus: %w", err)
	}

	return rows, total, nil
}

// menuSorting applies the specified sorting to the query builder.
func menuSorting(builder *ent.MenuQuery, sortBy string) *ent.MenuQuery {
	switch sortBy {
	case structs.SortByOrder:
		// Primary: Order DESC, Fallback: Created time DESC, Final: ID ASC
		return builder.Order(
			ent.Desc(menuEnt.FieldOrder),
			ent.Desc(menuEnt.FieldCreatedAt),
			ent.Asc(menuEnt.FieldID),
		)
	case structs.SortByCreatedAt:
		// Primary: Created time DESC, Fallback: Order DESC, Final: ID ASC
		return builder.Order(
			ent.Desc(menuEnt.FieldCreatedAt),
			ent.Desc(menuEnt.FieldOrder),
			ent.Asc(menuEnt.FieldID),
		)
	case structs.SortByName:
		// Primary: Name ASC, Fallback: Order DESC, Final: ID ASC
		return builder.Order(
			ent.Asc(menuEnt.FieldName),
			ent.Desc(menuEnt.FieldOrder),
			ent.Asc(menuEnt.FieldID),
		)
	default:
		// Default: Order DESC, Created time DESC, ID ASC
		return builder.Order(
			ent.Desc(menuEnt.FieldOrder),
			ent.Desc(menuEnt.FieldCreatedAt),
			ent.Asc(menuEnt.FieldID),
		)
	}
}

// menuCondition applies the cursor-based condition to the query builder.
func menuCondition(builder *ent.MenuQuery, id string, value any, direction string, sortBy string) *ent.MenuQuery {
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
	default:
		return menuCondition(builder, id, value, direction, structs.SortByCreatedAt)
	}
}

// listBuilder - create list builder.
func (r *menuRepository) listBuilder(_ context.Context, params *structs.ListMenuParams) (*ent.MenuQuery, error) {
	// create builder.
	builder := r.ec.Menu.Query()

	// match tenant id.
	if params.Tenant != "" {
		builder.Where(menuEnt.TenantIDEQ(params.Tenant))
	}

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
	// create builder.
	builder := r.ec.Menu.Query()

	// set where conditions.
	if validator.IsNotEmpty(params.Menu) {
		builder.Where(menuEnt.Or(
			menuEnt.IDEQ(params.Menu),
			menuEnt.SlugEQ(params.Menu),
		))
	}
	// match tenant id.
	if validator.IsNotEmpty(params.Tenant) {
		builder.Where(menuEnt.TenantIDEQ(params.Tenant))
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
