package repo

import (
	"context"
	"fmt"
	"ncobase/common/cache"
	"ncobase/common/log"
	"ncobase/common/meili"
	"ncobase/common/validator"
	"ncobase/internal/data"
	"ncobase/internal/data/ent"
	menuEnt "ncobase/internal/data/ent/menu"
	"ncobase/internal/data/structs"

	"github.com/redis/go-redis/v9"
)

// Menu - menu repository interface.
type Menu interface {
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
func NewMenu(d *data.Data) Menu {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &menuRepo{ec, rc, ms, cache.NewCache[ent.Menu](rc, cache.Key("nb_menu"), true)}
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
		// return nil, err
	}

	return row, nil
}

// GetTree retrieves the menu tree.
func (r *menuRepo) GetTree(ctx context.Context, p *structs.FindMenu) ([]*ent.Menu, error) {
	// use internal get method.
	subMenuIds := r.getSubMenuIds(ctx, p)

	// create builder.
	builder := r.ec.Menu.Query()
	builder.Where(menuEnt.IDIn(subMenuIds...))

	// order by order field (ascending)
	builder.Order(ent.Asc(menuEnt.FieldOrder))

	// If multiple menus have the same Order, add secondary sort by CreatedAt.
	builder.Order(ent.Asc(menuEnt.FieldCreatedAt))

	// execute the builder.
	rows, err := builder.All(ctx)
	if err != nil {
		log.Errorf(ctx, "menuRepo.GetTree error: %v", err)
		return nil, err
	}

	return rows, nil
}

// Get retrieves a specific menu.
func (r *menuRepo) Get(ctx context.Context, p *structs.FindMenu) (*ent.Menu, error) {
	cacheKey := fmt.Sprintf("%s", p.Menu)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	row, err := r.getMenu(ctx, p)
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
	// use internal get method.
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

	cacheKey := fmt.Sprintf("%s", row.ID)
	if err := r.c.Reset(ctx, cacheKey, row); err != nil {
		log.Errorf(ctx, "menuRepo.Update cache error: %v", err)
	}

	return row, nil
}

// Delete deletes a menu.
func (r *menuRepo) Delete(ctx context.Context, p *structs.FindMenu) error {

	// create builder.
	builder := r.ec.Menu.Delete()

	// set where conditions.
	builder.Where(menuEnt.Or(
		menuEnt.IDEQ(p.Menu),
		menuEnt.SlugEQ(p.Menu),
	))

	// match tenant id.
	if validator.IsNotEmpty(p.Tenant) {
		builder.Where(menuEnt.TenantIDEQ(p.Tenant))
	}

	// execute the builder.
	_, err := builder.Exec(ctx)
	if validator.IsNotNil(err) {
		return err
	}

	cacheKey := fmt.Sprintf("%s", p.Menu)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		log.Errorf(ctx, "menuRepo.Delete cache error: %v", err)
	}

	return nil
}

// List lists menus based on given parameters.
func (r *menuRepo) List(ctx context.Context, p *structs.ListMenuParams) ([]*ent.Menu, error) {
	// create list builder
	builder, err := r.listBuilder(ctx, p)
	if validator.IsNotNil(err) {
		return nil, err
	}

	// limit the result
	builder.Limit(p.Limit)

	// order by order field
	builder.Order(ent.Desc(menuEnt.FieldOrder))

	// execute the builder.
	rows, err := builder.All(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return rows, nil
}

// CountX counts menus based on given parameters.
func (r *menuRepo) CountX(ctx context.Context, p *structs.ListMenuParams) int {
	// create list builder
	builder, err := r.listBuilder(ctx, p)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// listBuilder - create list builder.
// internal method.
func (r *menuRepo) listBuilder(ctx context.Context, p *structs.ListMenuParams) (*ent.MenuQuery, error) {
	// verify query p.
	var nextMenu *ent.Menu
	if validator.IsNotEmpty(p.Cursor) {
		// query the menu.
		// use internal get method.
		row, err := r.getMenu(ctx, &structs.FindMenu{
			Menu:   p.Cursor,
			Tenant: p.Tenant,
			Type:   p.Type,
		})
		if validator.IsNotNil(err) || validator.IsNil(row) {
			return nil, err
		}
		nextMenu = row
	}

	// create builder.
	builder := r.ec.Menu.Query()

	// lt the cursor create time
	if nextMenu != nil {
		builder.Where(menuEnt.CreatedAtLT(nextMenu.CreatedAt))
	}

	// match tenant id.
	if validator.IsNotEmpty(p.Tenant) {
		builder.Where(menuEnt.TenantIDEQ(p.Tenant))
	}

	// match type.
	if validator.IsNotEmpty(p.Type) {
		builder.Where(menuEnt.TypeEQ(p.Type))
	}
	// match permission.
	if validator.IsNotEmpty(p.Perms) {
		builder.Where(menuEnt.PermsContains(p.Perms))
	}
	// match parent id.
	// default is root.
	if validator.IsEmpty(p.Parent) {
		builder.Where(menuEnt.Or(
			menuEnt.ParentIDIsNil(),
			menuEnt.ParentIDEQ(""),
			menuEnt.ParentIDEQ("root"),
		))
	} else {
		builder.Where(menuEnt.ParentIDEQ(p.Parent))
	}
	return builder, nil
}

// getMenu - get menu.
// internal method.
func (r *menuRepo) getMenu(ctx context.Context, p *structs.FindMenu) (*ent.Menu, error) {
	// create builder.
	builder := r.ec.Menu.Query()

	// set where conditions.
	if validator.IsNotEmpty(p.Menu) {
		builder.Where(menuEnt.Or(
			menuEnt.IDEQ(p.Menu),
			menuEnt.SlugEQ(p.Menu),
		))
	}
	// match tenant id.
	if validator.IsNotEmpty(p.Tenant) {
		builder.Where(menuEnt.TenantIDEQ(p.Tenant))
	}

	// execute the builder.
	row, err := builder.First(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// getSubMenuIds recursively retrieves sub-menu IDs.
// Internal method.
func (r *menuRepo) getSubMenuIds(ctx context.Context, p *structs.FindMenu) []string {
	var subMenuIds []string

	// Create a builder to query menu IDs.
	builder := r.ec.Menu.Query()

	// Build the query based on the parameters.
	if validator.IsEmpty(p.Menu) || p.Menu == "root" {
		builder.Where(menuEnt.Or(menuEnt.ParentIDIsNil(), menuEnt.ParentIDEQ("")))
	} else {
		builder.Where(menuEnt.ParentIDEQ(p.Menu))
	}

	if validator.IsNotEmpty(p.Type) {
		builder.Where(menuEnt.TypeEQ(p.Type))
	}

	if validator.IsNotEmpty(p.Tenant) {
		builder.Where(menuEnt.TenantIDEQ(p.Tenant))
	}

	// Retrieve menu IDs.
	menuIDs := builder.IDsX(ctx)

	// Iterate through each menu ID to get its sub-menu IDs recursively.
	for _, id := range menuIDs {
		subIds := r.getSubMenuIds(ctx, &structs.FindMenu{Menu: id, Tenant: p.Tenant})
		subMenuIds = append(subMenuIds, subIds...)
	}

	// Include the current menu ID itself in the result.
	subMenuIds = append(subMenuIds, menuIDs...)

	return subMenuIds
}
