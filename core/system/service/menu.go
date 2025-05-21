package service

import (
	"context"
	"errors"
	"ncobase/system/data"
	"ncobase/system/data/ent"
	"ncobase/system/data/repository"
	"ncobase/system/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation/validator"
)

// MenuServiceInterface represents the menu service interface.
type MenuServiceInterface interface {
	Create(ctx context.Context, body *structs.MenuBody) (*structs.ReadMenu, error)
	Update(ctx context.Context, updates *structs.UpdateMenuBody) (*structs.ReadMenu, error)
	Get(ctx context.Context, params *structs.FindMenu) (any, error)
	GetByType(ctx context.Context, menuType string) ([]*structs.ReadMenu, error)
	GetBySlug(ctx context.Context, slug string) (*structs.ReadMenu, error)
	GetHeaderMenus(ctx context.Context) ([]*structs.ReadMenu, error)
	GetSidebarMenus(ctx context.Context, parentID string) ([]*structs.ReadMenu, error)
	GetMenuByPerms(ctx context.Context, perms string) ([]*structs.ReadMenu, error)
	GetActiveMenus(ctx context.Context) ([]*structs.ReadMenu, error)
	GetMenuByPath(ctx context.Context, path string) (*structs.ReadMenu, error)
	MoveMenu(ctx context.Context, menuID string, newParentID string, newOrder int) (*structs.ReadMenu, error)
	EnableMenu(ctx context.Context, menuID string) (*structs.ReadMenu, error)
	DisableMenu(ctx context.Context, menuID string) (*structs.ReadMenu, error)
	ShowMenu(ctx context.Context, menuID string) (*structs.ReadMenu, error)
	HideMenu(ctx context.Context, menuID string) (*structs.ReadMenu, error)
	GetDefaultMenuTree(ctx context.Context) (map[string][]*structs.ReadMenu, error)
	GetUserAuthorizedMenus(ctx context.Context, userID string) ([]*structs.ReadMenu, error)
	BatchGetByID(ctx context.Context, menuIDs []string) (map[string]*structs.ReadMenu, error)
	ReorderMenus(ctx context.Context, menuIDs []string) error
	Delete(ctx context.Context, params *structs.FindMenu) (*structs.ReadMenu, error)
	List(ctx context.Context, params *structs.ListMenuParams) (paging.Result[*structs.ReadMenu], error)
	CountX(ctx context.Context, params *structs.ListMenuParams) int
	GetTree(ctx context.Context, params *structs.FindMenu) (paging.Result[*structs.ReadMenu], error)
}

// MenuService represents the menu service.
type menuService struct {
	menu repository.MenuRepositoryInterface
	em   ext.ManagerInterface
}

// NewMenuService creates a new menu service.
func NewMenuService(d *data.Data, em ext.ManagerInterface) MenuServiceInterface {
	return &menuService{
		menu: repository.NewMenuRepository(d),
		em:   em,
	}
}

// Create creates a new menu.
func (s *menuService) Create(ctx context.Context, body *structs.MenuBody) (*structs.ReadMenu, error) {
	if validator.IsEmpty(body.Name) {
		return nil, errors.New(ecode.FieldIsInvalid("name"))
	}

	row, err := s.menu.Create(ctx, body)
	if err := handleEntError(ctx, "Menu", err); err != nil {
		return nil, err
	}

	// // publish event
	// s.em.PublishEvent("menu.created", s.Serialize(row))

	return s.Serialize(row), nil
}

// Update updates an existing menu (full and partial).
func (s *menuService) Update(ctx context.Context, updates *structs.UpdateMenuBody) (*structs.ReadMenu, error) {
	if validator.IsEmpty(updates.ID) {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	row, err := s.menu.Update(ctx, updates)
	if err := handleEntError(ctx, "Menu", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Get retrieves a menu by ID.
func (s *menuService) Get(ctx context.Context, params *structs.FindMenu) (any, error) {

	if params.Children {
		return s.GetTree(ctx, params)
	}

	row, err := s.menu.Get(ctx, params)
	if err := handleEntError(ctx, "Menu", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// GetByType retrieves menus by type.
func (s *menuService) GetByType(ctx context.Context, menuType string) ([]*structs.ReadMenu, error) {
	if validator.IsEmpty(menuType) {
		return nil, errors.New(ecode.FieldIsRequired("type"))
	}

	params := &structs.ListMenuParams{
		Type: menuType,
	}

	result, err := s.List(ctx, params)
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}

// GetBySlug retrieves a menu by slug.
func (s *menuService) GetBySlug(ctx context.Context, slug string) (*structs.ReadMenu, error) {
	if validator.IsEmpty(slug) {
		return nil, errors.New(ecode.FieldIsRequired("slug"))
	}

	params := &structs.FindMenu{
		Menu: slug,
	}

	result, err := s.Get(ctx, params)
	if err != nil {
		return nil, err
	}

	menu, ok := result.(*structs.ReadMenu)
	if !ok {
		return nil, errors.New(ecode.AssertionFailed("menu"))
	}

	return menu, nil
}

// GetHeaderMenus retrieves all header menus.
func (s *menuService) GetHeaderMenus(ctx context.Context) ([]*structs.ReadMenu, error) {
	return s.GetByType(ctx, "header")
}

// GetSidebarMenus retrieves all sidebar menus for a given parent.
func (s *menuService) GetSidebarMenus(ctx context.Context, parentID string) ([]*structs.ReadMenu, error) {
	params := &structs.ListMenuParams{
		Type:   "sidebar",
		Parent: parentID,
	}

	result, err := s.List(ctx, params)
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}

// GetMenuByPerms retrieves menus by permission code.
func (s *menuService) GetMenuByPerms(ctx context.Context, perms string) ([]*structs.ReadMenu, error) {
	if validator.IsEmpty(perms) {
		return nil, errors.New(ecode.FieldIsRequired("perms"))
	}

	params := &structs.ListMenuParams{
		Perms: perms,
	}

	result, err := s.List(ctx, params)
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}

// GetActiveMenus retrieves all non-disabled, non-hidden menus.
func (s *menuService) GetActiveMenus(ctx context.Context) ([]*structs.ReadMenu, error) {
	// We'll need to enhance the repository to filter by hidden and disabled
	// For now, we'll filter the results in memory
	params := &structs.ListMenuParams{
		Limit: 1000, // Use a reasonable limit
	}

	result, err := s.List(ctx, params)
	if err != nil {
		return nil, err
	}

	activeMenus := make([]*structs.ReadMenu, 0, len(result.Items))
	for _, menu := range result.Items {
		if !menu.Hidden && !menu.Disabled {
			activeMenus = append(activeMenus, menu)
		}
	}

	return activeMenus, nil
}

// GetMenuByPath retrieves menus by path.
func (s *menuService) GetMenuByPath(ctx context.Context, path string) (*structs.ReadMenu, error) {
	if validator.IsEmpty(path) {
		return nil, errors.New(ecode.FieldIsRequired("path"))
	}

	// Note: This would ideally be implemented at the repository level
	// For now, we'll use List and filter
	params := &structs.ListMenuParams{
		Limit: 1000, // Use a reasonable limit
	}

	result, err := s.List(ctx, params)
	if err != nil {
		return nil, err
	}

	for _, menu := range result.Items {
		if menu.Path == path {
			return menu, nil
		}
	}

	return nil, errors.New(ecode.NotExist("Menu with path " + path))
}

// MoveMenu moves a menu to a new parent and/or changes its order.
func (s *menuService) MoveMenu(ctx context.Context, menuID string, newParentID string, newOrder int) (*structs.ReadMenu, error) {
	// Find the menu first
	params := &structs.FindMenu{
		Menu: menuID,
	}

	result, err := s.Get(ctx, params)
	if err != nil {
		return nil, err
	}

	menu, ok := result.(*structs.ReadMenu)
	if !ok {
		return nil, errors.New(ecode.AssertionFailed("menu"))
	}

	// Create update body
	updateBody := &structs.UpdateMenuBody{
		ID: menu.ID,
		MenuBody: structs.MenuBody{
			ParentID: newParentID,
			Order:    &newOrder,
		},
	}

	return s.Update(ctx, updateBody)
}

// EnableMenu enables a disabled menu.
func (s *menuService) EnableMenu(ctx context.Context, menuID string) (*structs.ReadMenu, error) {
	// Find the menu first
	params := &structs.FindMenu{
		Menu: menuID,
	}

	result, err := s.Get(ctx, params)
	if err != nil {
		return nil, err
	}

	menu, ok := result.(*structs.ReadMenu)
	if !ok {
		return nil, errors.New(ecode.AssertionFailed("menu"))
	}

	// Skip if already enabled
	if !menu.Disabled {
		return menu, nil
	}

	// Create update body
	disabled := false
	updateBody := &structs.UpdateMenuBody{
		ID: menu.ID,
		MenuBody: structs.MenuBody{
			Disabled: &disabled,
		},
	}

	return s.Update(ctx, updateBody)
}

// DisableMenu disables a menu.
func (s *menuService) DisableMenu(ctx context.Context, menuID string) (*structs.ReadMenu, error) {
	// Find the menu first
	params := &structs.FindMenu{
		Menu: menuID,
	}

	result, err := s.Get(ctx, params)
	if err != nil {
		return nil, err
	}

	menu, ok := result.(*structs.ReadMenu)
	if !ok {
		return nil, errors.New(ecode.AssertionFailed("menu"))
	}

	// Skip if already disabled
	if menu.Disabled {
		return menu, nil
	}

	// Create update body
	disabled := true
	updateBody := &structs.UpdateMenuBody{
		ID: menu.ID,
		MenuBody: structs.MenuBody{
			Disabled: &disabled,
		},
	}

	return s.Update(ctx, updateBody)
}

// ShowMenu makes a hidden menu visible.
func (s *menuService) ShowMenu(ctx context.Context, menuID string) (*structs.ReadMenu, error) {
	// Find the menu first
	params := &structs.FindMenu{
		Menu: menuID,
	}

	result, err := s.Get(ctx, params)
	if err != nil {
		return nil, err
	}

	menu, ok := result.(*structs.ReadMenu)
	if !ok {
		return nil, errors.New(ecode.AssertionFailed("menu"))
	}

	// Skip if already visible
	if !menu.Hidden {
		return menu, nil
	}

	// Create update body
	hidden := false
	updateBody := &structs.UpdateMenuBody{
		ID: menu.ID,
		MenuBody: structs.MenuBody{
			Hidden: &hidden,
		},
	}

	return s.Update(ctx, updateBody)
}

// HideMenu hides a menu.
func (s *menuService) HideMenu(ctx context.Context, menuID string) (*structs.ReadMenu, error) {
	// Find the menu first
	params := &structs.FindMenu{
		Menu: menuID,
	}

	result, err := s.Get(ctx, params)
	if err != nil {
		return nil, err
	}

	menu, ok := result.(*structs.ReadMenu)
	if !ok {
		return nil, errors.New(ecode.AssertionFailed("menu"))
	}

	// Skip if already hidden
	if menu.Hidden {
		return menu, nil
	}

	// Create update body
	hidden := true
	updateBody := &structs.UpdateMenuBody{
		ID: menu.ID,
		MenuBody: structs.MenuBody{
			Hidden: &hidden,
		},
	}

	return s.Update(ctx, updateBody)
}

// GetDefaultMenuTree retrieves the complete menu tree with proper structure.
// This can be used to get the full navigation structure for a UI.
func (s *menuService) GetDefaultMenuTree(ctx context.Context) (map[string][]*structs.ReadMenu, error) {
	// Get all headers
	headers, err := s.GetHeaderMenus(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]*structs.ReadMenu)
	result["headers"] = headers

	// For each header, get its sidebars
	sidebars := make([]*structs.ReadMenu, 0)
	for _, header := range headers {
		headerSidebars, err := s.GetSidebarMenus(ctx, header.ID)
		if err != nil {
			logger.Warnf(ctx, "Error getting sidebars for header %s: %v", header.ID, err)
			continue
		}
		sidebars = append(sidebars, headerSidebars...)
	}
	result["sidebars"] = sidebars

	// Get account menus
	accountMenus, err := s.GetByType(ctx, "account")
	if err != nil {
		logger.Warnf(ctx, "Error getting account menus: %v", err)
	} else {
		result["accounts"] = accountMenus
	}

	// Get tenant menus
	tenantMenus, err := s.GetByType(ctx, "tenant")
	if err != nil {
		logger.Warnf(ctx, "Error getting tenant menus: %v", err)
	} else {
		result["tenants"] = tenantMenus
	}

	return result, nil
}

// GetUserAuthorizedMenus retrieves menus that a user is authorized to see.
// This requires integration with the access control system.
func (s *menuService) GetUserAuthorizedMenus(ctx context.Context, userID string) ([]*structs.ReadMenu, error) {
	if validator.IsEmpty(userID) {
		return nil, errors.New(ecode.FieldIsRequired("userID"))
	}

	// This is a placeholder. The actual implementation would:
	// 1. Get the user's roles
	// 2. Get permissions associated with those roles
	// 3. Filter menus by those permissions

	// For now, we'll just get all active menus
	return s.GetActiveMenus(ctx)
}

// BatchGetByID retrieves multiple menus by their IDs.
func (s *menuService) BatchGetByID(ctx context.Context, menuIDs []string) (map[string]*structs.ReadMenu, error) {
	if len(menuIDs) == 0 {
		return make(map[string]*structs.ReadMenu), nil
	}

	result := make(map[string]*structs.ReadMenu)

	for _, id := range menuIDs {
		params := &structs.FindMenu{
			Menu: id,
		}

		menuResult, err := s.Get(ctx, params)
		if err != nil {
			logger.Warnf(ctx, "Failed to get menu %s: %v", id, err)
			continue
		}

		menu, ok := menuResult.(*structs.ReadMenu)
		if !ok {
			logger.Warnf(ctx, "Type assertion failed for menu %s", id)
			continue
		}

		result[id] = menu
	}

	return result, nil
}

// ReorderMenus reorders a set of sibling menus.
func (s *menuService) ReorderMenus(ctx context.Context, menuIDs []string) error {
	if len(menuIDs) == 0 {
		return nil
	}

	// Get all the menus first to validate they exist
	menus, err := s.BatchGetByID(ctx, menuIDs)
	if err != nil {
		return err
	}

	// Check that we got all menus
	if len(menus) != len(menuIDs) {
		return errors.New("not all menu IDs were found")
	}

	// Update each menu with its new order
	for i, id := range menuIDs {
		menu := menus[id]

		// Skip if order is already correct
		if menu.Order == i {
			continue
		}

		order := i
		updateBody := &structs.UpdateMenuBody{
			ID: id,
			MenuBody: structs.MenuBody{
				Order: &order,
			},
		}

		if _, err := s.Update(ctx, updateBody); err != nil {
			logger.Errorf(ctx, "Failed to update order for menu %s: %v", id, err)
			return err
		}
	}

	return nil
}

// Delete deletes a menu by ID.
func (s *menuService) Delete(ctx context.Context, params *structs.FindMenu) (*structs.ReadMenu, error) {
	err := s.menu.Delete(ctx, params)
	if err := handleEntError(ctx, "Menu", err); err != nil {
		return nil, err
	}

	return nil, nil
}

// List lists all menus.
func (s *menuService) List(ctx context.Context, params *structs.ListMenuParams) (paging.Result[*structs.ReadMenu], error) {
	if params.Children {
		return s.GetTree(ctx, &structs.FindMenu{
			Children: true,
			Tenant:   params.Tenant,
			Menu:     params.Parent,
			Type:     params.Type,
			SortBy:   params.SortBy,
		})
	}

	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadMenu, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, total, err := s.menu.ListWithCount(ctx, &lp)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
			}
			logger.Errorf(ctx, "Error listing menus: %v", err)
			return nil, 0, err
		}

		return s.Serializes(rows), total, nil
	})
}

// CountX gets a count of menus.
func (s *menuService) CountX(ctx context.Context, params *structs.ListMenuParams) int {
	return s.menu.CountX(ctx, params)
}

// GetTree retrieves the menu tree.
func (s *menuService) GetTree(ctx context.Context, params *structs.FindMenu) (paging.Result[*structs.ReadMenu], error) {
	rows, err := s.menu.GetTree(ctx, params)
	if err := handleEntError(ctx, "Menu", err); err != nil {
		return paging.Result[*structs.ReadMenu]{}, err
	}

	return paging.Result[*structs.ReadMenu]{
		Items: s.buildMenuTree(rows, string(determineSortField(params))),
		Total: len(rows),
	}, nil
}

// buildMenuTree builds a menu tree structure.
func (s *menuService) buildMenuTree(menus []*ent.Menu, sortField string) []*structs.ReadMenu {
	menuNodes := make([]*structs.ReadMenu, len(menus))
	for i, m := range menus {
		menuNodes[i] = s.Serialize(m)
	}

	tree := types.BuildTree(menuNodes, sortField)
	return tree
}

func determineSortField(params *structs.FindMenu) string {
	if params.SortBy != "" {
		return params.SortBy
	}
	return structs.SortByCreatedAt // Default
}

// Serializes menus.
func (s *menuService) Serializes(rows []*ent.Menu) []*structs.ReadMenu {
	rs := make([]*structs.ReadMenu, len(rows))
	for i, row := range rows {
		rs[i] = s.Serialize(row)
	}
	return rs
}

// Serialize serializes a menu.
func (s *menuService) Serialize(row *ent.Menu) *structs.ReadMenu {
	return &structs.ReadMenu{
		ID:        row.ID,
		Name:      row.Name,
		Label:     row.Label,
		Slug:      row.Slug,
		Type:      row.Type,
		Path:      row.Path,
		Target:    row.Target,
		Icon:      row.Icon,
		Perms:     row.Perms,
		Hidden:    row.Hidden,
		Order:     row.Order,
		Disabled:  row.Disabled,
		Extras:    &row.Extras,
		ParentID:  row.ParentID,
		TenantID:  row.TenantID,
		CreatedBy: &row.CreatedBy,
		CreatedAt: &row.CreatedAt,
		UpdatedBy: &row.UpdatedBy,
		UpdatedAt: &row.UpdatedAt,
	}
}
