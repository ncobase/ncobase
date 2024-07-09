package menu

import (
	"context"
	"fmt"
	"ncobase/common/cache"
	"ncobase/common/log"
	"ncobase/common/meili"
	"ncobase/common/validator"
	"ncobase/feature/system/data"
	"ncobase/feature/system/data/ent"
	menuEnt "ncobase/feature/system/data/ent/menu"
	"ncobase/feature/system/structs"
	"net/url"

	"github.com/redis/go-redis/v9"
)

// RepositoryInterface represents the menu repository interface.
type RepositoryInterface interface {
	Create(context.Context, *structs.MenuBody) (*ent.Menu, error)
	GetTree(context.Context, *structs.FindMenu) ([]*ent.Menu, error)
	Get(context.Context, *structs.FindMenu) (*ent.Menu, error)
	Update(context.Context, *structs.UpdateMenuBody) (*ent.Menu, error)
	Delete(context.Context, *structs.FindMenu) error
	List(context.Context, *structs.ListMenuParams) ([]*ent.Menu, error)
	CountX(context.Context, *structs.ListMenuParams) int
}

// menuRepo implements the Menu interface.
type menuRepo struct {
	ec *ent.Client
	rc *redis.Client
	ms *meili.Client
	c  *cache.Cache[ent.Menu]
}

// NewMenu creates a new menu repository.
func NewMenu(d *data.Data) RepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &menuRepo{
		ec: ec,
		rc: rc,
		ms: ms,
		c:  cache.NewCache[ent.Menu](rc, "nb_menu", false),
	}
}

// Create creates a new menu.
func (r *menuRepo) Create(ctx context.Context, body *structs.MenuBody) (*ent.Menu, error) {
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
		log.Errorf(ctx, "menuRepo.Create error: %v", err)
		return nil, err
	}

	// Create the menu in Meilisearch index
	if err = r.ms.IndexDocuments("menus", row); err != nil {
		log.Errorf(context.Background(), "menuRepo.Create error creating Meilisearch index: %v\n", err)
	}

	// delete cached menu tree
	// _ = r.c.Delete(ctx, cache.Key("menu=tree"))

	return row, nil
}

// GetTree retrieves the menu tree.
func (r *menuRepo) GetTree(ctx context.Context, params *structs.FindMenu) ([]*ent.Menu, error) {
	// create builder
	builder := r.ec.Menu.Query()

	// set where conditions
	// if validator.IsNotEmpty(params.Type) {
	// 	builder.Where(menuEnt.TypeEQ(params.Type))
	// }
	if validator.IsNotEmpty(params.Tenant) {
		builder.Where(menuEnt.TenantIDEQ(params.Tenant))
	}

	// handle sub menus
	if validator.IsNotEmpty(params.Menu) && params.Menu != "root" {
		return r.getSubMenu(ctx, params.Menu, builder)
	}

	// execute the builder
	return r.executeArrayQuery(ctx, builder)
}

// Get retrieves a specific menu.
func (r *menuRepo) Get(ctx context.Context, params *structs.FindMenu) (*ent.Menu, error) {
	cacheParams := url.Values{}
	if validator.IsNotEmpty(params.Menu) {
		cacheParams.Set("menu", params.Menu)
	}
	cacheKey := cacheParams.Encode()

	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	row, err := r.getMenu(ctx, params)
	if err != nil {
		log.Errorf(ctx, "menuRepo.Get error: %v", err)
		return nil, err
	}

	if err := r.c.Set(ctx, cacheKey, row); err != nil {
		log.Errorf(ctx, "menuRepo.Get cache error: %v", err)
	}

	return row, nil
}

// Update updates an existing menu.
func (r *menuRepo) Update(ctx context.Context, body *structs.UpdateMenuBody) (*ent.Menu, error) {
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
		log.Errorf(ctx, "menuRepo.Update error: %v", err)
		return nil, err
	}

	// update cache
	cacheKey := fmt.Sprintf("%s", row.ID)
	if err := r.c.Set(ctx, cacheKey, row); err != nil {
		log.Errorf(ctx, "menuRepo.Update cache error: %v", err)
	}

	// delete menu tree cache
	// if err := r.c.Delete(ctx, "menu=tree"); err != nil {
	// 	log.Errorf(ctx, "menuRepo.Update cache error: %v", err)
	// }

	return row, nil
}

// Delete deletes a menu.
func (r *menuRepo) Delete(ctx context.Context, params *structs.FindMenu) error {
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
		log.Errorf(ctx, "menuRepo.Delete cache error: %v", err)
	}

	return nil
}

// List lists menus based on given parameters.
func (r *menuRepo) List(ctx context.Context, params *structs.ListMenuParams) ([]*ent.Menu, error) {
	// cacheParams := url.Values{}
	// if validator.IsNotEmpty(params.Cursor) {
	// 	cacheParams.Set("cursor", params.Cursor)
	// }
	// if validator.IsNotEmpty(params.Tenant) {
	// 	cacheParams.Set("tenant", params.Tenant)
	// }
	// if validator.IsNotEmpty(params.Parent) {
	// 	cacheParams.Set("parent", params.Parent)
	// }
	// if len(cacheParams) == 0 {
	// 	cacheParams.Set("menu", "list")
	// }
	// cacheKey := cacheParams.Encode()

	// // Attempt to retrieve the menu list from cache
	// var rows []*ent.Menu
	// if err := r.c.GetArray(ctx, cacheKey, &rows); err == nil && len(rows) > 0 {
	// 	return rows, nil
	// }

	// create list builder
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return nil, err
	}

	// limit the result
	builder.Limit(params.Limit)

	// order by order field
	builder.Order(ent.Desc(menuEnt.FieldOrder))

	// execute the builder.
	return r.executeArrayQuery(ctx, builder)
}

// CountX counts menus based on given parameters.
func (r *menuRepo) CountX(ctx context.Context, params *structs.ListMenuParams) int {
	// create list builder
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// listBuilder - create list builder.
// internal method.
func (r *menuRepo) listBuilder(ctx context.Context, params *structs.ListMenuParams) (*ent.MenuQuery, error) {
	// verify query params.
	var next *ent.Menu
	if validator.IsNotEmpty(params.Cursor) {
		// query the menu.
		row, err := r.getMenu(ctx, &structs.FindMenu{
			Menu:   params.Cursor,
			Tenant: params.Tenant,
			Type:   params.Type,
		})
		if validator.IsNotNil(err) || validator.IsNil(row) {
			return nil, err
		}
		next = row
	}

	// create builder.
	builder := r.ec.Menu.Query()

	// lt the cursor create time
	if next != nil {
		builder.Where(menuEnt.CreatedAtLT(next.CreatedAt))
	}

	// match tenant id.
	if validator.IsNotEmpty(params.Tenant) {
		builder.Where(menuEnt.TenantIDEQ(params.Tenant))
	}

	// match type.
	if validator.IsNotEmpty(params.Type) {
		builder.Where(menuEnt.TypeEQ(params.Type))
	}
	// match permission.
	if validator.IsNotEmpty(params.Perms) {
		builder.Where(menuEnt.PermsContains(params.Perms))
	}
	// match parent id.
	// default is root.
	if validator.IsEmpty(params.Parent) {
		builder.Where(menuEnt.Or(
			menuEnt.ParentIDIsNil(),
			menuEnt.ParentIDEQ(""),
			menuEnt.ParentIDEQ("root"),
		))
	} else {
		builder.Where(menuEnt.ParentIDEQ(params.Parent))
	}
	return builder, nil
}

// getMenu - get menu.
// internal method.
func (r *menuRepo) getMenu(ctx context.Context, params *structs.FindMenu) (*ent.Menu, error) {
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
func (r *menuRepo) getSubMenu(ctx context.Context, rootID string, builder *ent.MenuQuery) ([]*ent.Menu, error) {
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
func (r *menuRepo) executeArrayQuery(ctx context.Context, builder *ent.MenuQuery) ([]*ent.Menu, error) {
	rows, err := builder.All(ctx)
	if err != nil {
		log.Errorf(ctx, "menuRepo.executeArrayQuery error: %v", err)
		return nil, err
	}
	return rows, nil
}
