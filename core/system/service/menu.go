package service

import (
	"context"
	"errors"
	"ncobase/core/system/data"
	"ncobase/core/system/data/ent"
	"ncobase/core/system/data/repository"
	"ncobase/core/system/structs"
	"ncobase/core/system/wrapper"
	"sort"
	"strings"

	"github.com/ncobase/ncore/ctxutil"
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
	GetBySlug(ctx context.Context, slug string) (*structs.ReadMenu, error)
	GetByPath(ctx context.Context, path string) (*structs.ReadMenu, error)
	GetMenus(ctx context.Context, opts structs.MenuQueryParams) ([]*structs.ReadMenu, error)
	GetMenusByTypes(ctx context.Context, types []string, opts structs.MenuQueryParams) ([]*structs.ReadMenu, error)
	GetNavigationMenus(ctx context.Context, sortBy string) (*structs.NavigationMenus, error)
	GetUserAuthorizedMenus(ctx context.Context, userID string) ([]*structs.ReadMenu, error)
	BatchGetByID(ctx context.Context, menuIDs []string) (map[string]*structs.ReadMenu, error)
	MoveMenu(ctx context.Context, menuID string, newParentID string, newOrder int) (*structs.ReadMenu, error)
	ReorderMenus(ctx context.Context, menuIDs []string) error
	ToggleStatus(ctx context.Context, menuID string, action string) (*structs.ReadMenu, error)
	Delete(ctx context.Context, params *structs.FindMenu) (*structs.ReadMenu, error)
	List(ctx context.Context, params *structs.ListMenuParams) (paging.Result[*structs.ReadMenu], error)
	CountX(ctx context.Context, params *structs.ListMenuParams) int
	GetMenuTree(ctx context.Context, params *structs.FindMenu) (paging.Result[*structs.ReadMenu], error)
}

// MenuService represents the menu service.
type menuService struct {
	menu repository.MenuRepositoryInterface
	em   ext.ManagerInterface

	tsw *wrapper.SpaceServiceWrapper
}

// NewMenuService creates a new menu service.
func NewMenuService(d *data.Data, em ext.ManagerInterface, tsw *wrapper.SpaceServiceWrapper) MenuServiceInterface {
	return &menuService{
		menu: repository.NewMenuRepository(d),
		em:   em,
		tsw:  tsw,
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

	return s.Serialize(row), nil
}

// Update updates an existing menu.
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

// Get retrieves a menu by ID with optional tree structure.
func (s *menuService) Get(ctx context.Context, params *structs.FindMenu) (any, error) {
	if params.Children {
		return s.GetMenuTree(ctx, params)
	}

	row, err := s.menu.Get(ctx, params)
	if err := handleEntError(ctx, "Menu", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// GetBySlug retrieves a menu by slug.
func (s *menuService) GetBySlug(ctx context.Context, slug string) (*structs.ReadMenu, error) {
	if validator.IsEmpty(slug) {
		return nil, errors.New(ecode.FieldIsRequired("slug"))
	}

	result, err := s.Get(ctx, &structs.FindMenu{Menu: slug})
	if err != nil {
		return nil, err
	}

	menu, ok := result.(*structs.ReadMenu)
	if !ok {
		return nil, errors.New(ecode.AssertionFailed("menu"))
	}

	return menu, nil
}

// GetByPath retrieves a menu by path.
func (s *menuService) GetByPath(ctx context.Context, path string) (*structs.ReadMenu, error) {
	if validator.IsEmpty(path) {
		return nil, errors.New(ecode.FieldIsRequired("path"))
	}

	menus, err := s.GetMenus(ctx, structs.MenuQueryParams{
		Path:  path,
		Limit: 1,
	})
	if err != nil {
		return nil, err
	}

	if len(menus) == 0 {
		return nil, errors.New(ecode.NotExist("Menu with path " + path))
	}

	return menus[0], nil
}

// GetMenus retrieves menus based on flexible query options.
func (s *menuService) GetMenus(ctx context.Context, opts structs.MenuQueryParams) ([]*structs.ReadMenu, error) {
	params := &structs.ListMenuParams{
		Type:     opts.Type,
		Parent:   opts.ParentID,
		Perms:    opts.Perms,
		Children: opts.Children,
		SortBy:   s.getDefaultSort(opts.SortBy),
		Limit:    s.getDefaultLimit(opts.Limit),
	}

	result, err := s.List(ctx, params)
	if err != nil {
		return nil, err
	}

	menus := result.Items

	// Apply path filter if specified
	if opts.Path != "" {
		menus = s.filterByPath(menus, opts.Path)
	}

	// Apply active filter if specified
	if opts.ActiveOnly {
		menus = s.filterActiveMenus(menus)
	}

	return menus, nil
}

// GetMenusByTypes retrieves and merges menus from multiple types, with optional tree building
func (s *menuService) GetMenusByTypes(ctx context.Context, types []string, opts structs.MenuQueryParams) ([]*structs.ReadMenu, error) {
	if len(types) == 0 {
		return []*structs.ReadMenu{}, nil
	}

	// Query all menus in a single database call and filter by type
	params := &structs.ListMenuParams{
		Type:     "",
		Parent:   opts.ParentID,
		Perms:    opts.Perms,
		Children: opts.Children,
		SortBy:   s.getDefaultSort(opts.SortBy),
		Limit:    10000,
	}

	result, err := s.List(ctx, params)
	if err != nil {
		return nil, err
	}

	// Filter menus by requested types
	typeSet := make(map[string]bool)
	for _, t := range types {
		typeSet[t] = true
	}

	allMenus := make([]*structs.ReadMenu, 0)
	for _, menu := range result.Items {
		if typeSet[menu.Type] {
			allMenus = append(allMenus, menu)
		}
	}

	// Apply path filter if specified
	if opts.Path != "" {
		allMenus = s.filterByPath(allMenus, opts.Path)
	}

	// Apply active filter if specified
	if opts.ActiveOnly {
		allMenus = s.filterActiveMenus(allMenus)
	}

	// Build tree structure if requested
	if opts.Children {
		allMenus = s.buildMenuTree(allMenus)
		s.sortMenusByField(allMenus, s.getDefaultSort(opts.SortBy))
	}

	return allMenus, nil
}

// GetNavigationMenus retrieves all navigation menus grouped by type.
func (s *menuService) GetNavigationMenus(ctx context.Context, sortBy string) (*structs.NavigationMenus, error) {
	sortBy = s.getDefaultSort(sortBy)
	params := structs.MenuQueryParams{
		SortBy:     sortBy,
		ActiveOnly: true,
		Limit:      1000,
	}

	nav := &structs.NavigationMenus{}
	userPermissions := ctxutil.GetUserPermissions(ctx)
	isAdmin := ctxutil.GetUserIsAdmin(ctx)
	spaceID := ctxutil.GetSpaceID(ctx)

	// Get space-specific menus if space context exists
	var spaceMenuIDs []string
	if spaceID != "" {
		if s.tsw != nil && s.tsw.HasSpaceMenuService() {
			var err error
			spaceMenuIDs, err = s.tsw.GetSpaceMenus(ctx, spaceID)
			if err != nil {
				logger.Warnf(ctx, "Failed to get space menus: %v", err)
			}
		}
	}

	// Get headers
	if headers, err := s.getMenusWithSpaceAndPermissionFilter(ctx, "header", params, spaceMenuIDs, userPermissions, isAdmin); err != nil {
		logger.Warnf(ctx, "Failed to get header menus: %v", err)
		nav.Headers = []*structs.ReadMenu{}
	} else {
		nav.Headers = headers
	}

	// Get sidebars and submenus with tree structure
	if sidebars, err := s.GetMenusByTypesWithSpaceAndPermissionCheck(ctx, []string{"sidebar", "submenu"}, structs.MenuQueryParams{
		SortBy:     params.SortBy,
		ActiveOnly: params.ActiveOnly,
		Children:   true,
		Limit:      params.Limit,
	}, spaceMenuIDs, userPermissions, isAdmin); err != nil {
		logger.Warnf(ctx, "Failed to get sidebar/submenu menus: %v", err)
		nav.Sidebars = []*structs.ReadMenu{}
	} else {
		nav.Sidebars = sidebars
	}

	// Get accounts - special handling for space context
	if accounts, err := s.getAccountMenusWithSpaceContext(ctx, params, spaceID, userPermissions, isAdmin); err != nil {
		logger.Warnf(ctx, "Failed to get account menus: %v", err)
		nav.Accounts = []*structs.ReadMenu{}
	} else {
		nav.Accounts = accounts
	}

	// Get spaces - show space switching menus only if user has multiple spaces
	if spaces, err := s.getSpaceMenusWithContext(ctx, params, userPermissions, isAdmin); err != nil {
		logger.Warnf(ctx, "Failed to get space menus: %v", err)
		nav.Spaces = []*structs.ReadMenu{}
	} else {
		nav.Spaces = spaces
	}

	return nav, nil
}

// getMenusWithSpaceAndPermissionFilter filters menus by space and permissions
func (s *menuService) getMenusWithSpaceAndPermissionFilter(ctx context.Context, menuType string, params structs.MenuQueryParams, spaceMenuIDs []string, userPermissions []string, isAdmin bool) ([]*structs.ReadMenu, error) {
	typeParams := params
	typeParams.Type = menuType

	menus, err := s.GetMenus(ctx, typeParams)
	if err != nil {
		return nil, err
	}

	// Filter by space if space context exists
	if len(spaceMenuIDs) > 0 {
		menus = s.filterMenusBySpace(menus, spaceMenuIDs)
	}

	// Filter by permissions
	return s.filterMenusByPermission(ctx, menus, userPermissions, isAdmin), nil
}

// GetMenusByTypesWithSpaceAndPermissionCheck retrieves menus with space and permission filtering
func (s *menuService) GetMenusByTypesWithSpaceAndPermissionCheck(ctx context.Context, types []string, opts structs.MenuQueryParams, spaceMenuIDs []string, userPermissions []string, isAdmin bool) ([]*structs.ReadMenu, error) {
	allMenus, err := s.getMenusByTypesWithParents(ctx, types, opts)
	if err != nil {
		return nil, err
	}

	// Filter by space if space context exists
	if len(spaceMenuIDs) > 0 {
		allMenus = s.filterMenusBySpace(allMenus, spaceMenuIDs)
	}

	// Filter by permissions
	filteredMenus := s.filterMenusByPermission(ctx, allMenus, userPermissions, isAdmin)

	if opts.Children {
		filteredMenus = s.buildMenuTree(filteredMenus)
		s.sortMenusByField(filteredMenus, s.getDefaultSort(opts.SortBy))
	}

	return filteredMenus, nil
}

// getAccountMenusWithSpaceContext handles account menus with special space considerations
func (s *menuService) getAccountMenusWithSpaceContext(ctx context.Context, params structs.MenuQueryParams, spaceID string, userPermissions []string, isAdmin bool) ([]*structs.ReadMenu, error) {
	typeParams := params
	typeParams.Type = "account"

	menus, err := s.GetMenus(ctx, typeParams)
	if err != nil {
		return nil, err
	}

	// Apply permission filtering
	filteredMenus := s.filterMenusByPermission(ctx, menus, userPermissions, isAdmin)

	// Add dynamic account menus based on space context
	if spaceID != "" {
		dynamicMenus := s.generateSpaceAccountMenus(ctx, spaceID)
		filteredMenus = append(filteredMenus, dynamicMenus...)
	}

	return filteredMenus, nil
}

// getSpaceMenusWithContext handles space switching and management menus
func (s *menuService) getSpaceMenusWithContext(ctx context.Context, params structs.MenuQueryParams, userPermissions []string, isAdmin bool) ([]*structs.ReadMenu, error) {
	typeParams := params
	typeParams.Type = "space"

	menus, err := s.GetMenus(ctx, typeParams)
	if err != nil {
		return nil, err
	}

	// Check if user has access to multiple spaces
	userSpaceIDs := ctxutil.GetUserSpaceIDs(ctx)

	// Only show space switching menus if user has multiple spaces or is admin
	if len(userSpaceIDs) <= 1 && !isAdmin {
		return []*structs.ReadMenu{}, nil
	}

	return s.filterMenusByPermission(ctx, menus, userPermissions, isAdmin), nil
}

// generateSpaceAccountMenus creates dynamic account menus based on space context
func (s *menuService) generateSpaceAccountMenus(ctx context.Context, spaceID string) []*structs.ReadMenu {
	var dynamicMenus []*structs.ReadMenu

	// // Add space-specific token management menu
	// dynamicMenus = append(dynamicMenus, &structs.ReadMenu{
	// 	ID:       "space-tokens-" + spaceID,
	// 	Name:     "Space API Keys",
	// 	Label:    "API Keys",
	// 	Type:     "account",
	// 	Path:     "/account/api-keys?space=" + spaceID,
	// 	Icon:     "key",
	// 	Perms:    "read:tokens",
	// 	Order:    100,
	// 	Hidden:   false,
	// 	Disabled: false,
	// })
	//
	// // Add space-specific session management menu
	// dynamicMenus = append(dynamicMenus, &structs.ReadMenu{
	// 	ID:       "space-sessions-" + spaceID,
	// 	Name:     "Active Sessions",
	// 	Label:    "Sessions",
	// 	Type:     "account",
	// 	Path:     "/account/sessions?space=" + spaceID,
	// 	Icon:     "monitor",
	// 	Perms:    "read:sessions",
	// 	Order:    101,
	// 	Hidden:   false,
	// 	Disabled: false,
	// })

	return dynamicMenus
}

// filterMenusBySpace filters menus based on space access
func (s *menuService) filterMenusBySpace(menus []*structs.ReadMenu, spaceMenuIDs []string) []*structs.ReadMenu {
	if len(spaceMenuIDs) == 0 {
		return menus
	}

	spaceMenuMap := make(map[string]bool)
	for _, id := range spaceMenuIDs {
		spaceMenuMap[id] = true
	}

	var filteredMenus []*structs.ReadMenu
	for _, menu := range menus {
		// Include menu if it's in space's allowed menus or if it's a system menu
		if spaceMenuMap[menu.ID] || menu.Type == "system" {
			filteredMenus = append(filteredMenus, menu)
		}
	}

	return filteredMenus
}

// getMenusWithPermissionFilter is a helper method for single type menu retrieval with permission filtering
func (s *menuService) getMenusWithPermissionFilter(ctx context.Context, menuType string, params structs.MenuQueryParams, userPermissions []string, isAdmin bool) ([]*structs.ReadMenu, error) {
	typeParams := params
	typeParams.Type = menuType

	menus, err := s.GetMenus(ctx, typeParams)
	if err != nil {
		return nil, err
	}

	return s.filterMenusByPermission(ctx, menus, userPermissions, isAdmin), nil
}

// GetMenusWithPermissionCheck retrieves menus with permission validation
func (s *menuService) GetMenusWithPermissionCheck(ctx context.Context, opts structs.MenuQueryParams, userPermissions []string, isAdmin bool) ([]*structs.ReadMenu, error) {
	menus, err := s.GetMenus(ctx, opts)
	if err != nil {
		return nil, err
	}

	return s.filterMenusByPermission(ctx, menus, userPermissions, isAdmin), nil
}

// GetMenusByTypesWithPermissionCheck retrieves and merges menus from multiple types with permission filtering
func (s *menuService) GetMenusByTypesWithPermissionCheck(ctx context.Context, types []string, opts structs.MenuQueryParams, userPermissions []string, isAdmin bool) ([]*structs.ReadMenu, error) {
	allMenus, err := s.getMenusByTypesWithParents(ctx, types, opts)
	if err != nil {
		return nil, err
	}

	filteredMenus := s.filterMenusByPermission(ctx, allMenus, userPermissions, isAdmin)

	if opts.Children {
		filteredMenus = s.buildMenuTree(filteredMenus)
		s.sortMenusByField(filteredMenus, s.getDefaultSort(opts.SortBy))
	}

	return filteredMenus, nil
}

// getMenusByTypesWithParents gets menus of specified types including necessary parent menus within type scope
func (s *menuService) getMenusByTypesWithParents(ctx context.Context, types []string, opts structs.MenuQueryParams) ([]*structs.ReadMenu, error) {
	typeSet := make(map[string]bool)
	for _, t := range types {
		typeSet[t] = true
	}

	var allMenus []*structs.ReadMenu
	menuIDs := make(map[string]bool)

	// Get menus of specified types
	for _, menuType := range types {
		typeOpts := opts
		typeOpts.Type = menuType
		typeOpts.Children = false

		menus, err := s.GetMenus(ctx, typeOpts)
		if err != nil {
			logger.Warnf(ctx, "Failed to get %s menus: %v", menuType, err)
			continue
		}

		for _, menu := range menus {
			if !menuIDs[menu.ID] {
				allMenus = append(allMenus, menu)
				menuIDs[menu.ID] = true
			}
		}
	}

	// Get necessary parent menus recursively within type scope
	parentIDs := make(map[string]bool)
	for _, menu := range allMenus {
		if menu.ParentID != "" && menu.ParentID != "root" && !menuIDs[menu.ParentID] {
			parentIDs[menu.ParentID] = true
		}
	}

	for len(parentIDs) > 0 {
		var currentParentIDs []string
		for parentID := range parentIDs {
			currentParentIDs = append(currentParentIDs, parentID)
		}

		parentMenus, err := s.BatchGetByID(ctx, currentParentIDs)
		if err != nil {
			logger.Warnf(ctx, "Failed to get parent menus: %v", err)
			break
		}

		parentIDs = make(map[string]bool)

		for _, parentMenu := range parentMenus {
			// Only include parent menus within specified types
			if !menuIDs[parentMenu.ID] && typeSet[parentMenu.Type] {
				allMenus = append(allMenus, parentMenu)
				menuIDs[parentMenu.ID] = true

				if parentMenu.ParentID != "" && parentMenu.ParentID != "root" && !menuIDs[parentMenu.ParentID] {
					parentIDs[parentMenu.ParentID] = true
				}
			}
		}
	}

	return allMenus, nil
}

// Permission filtering methods

// filterMenusByPermission filters menus based on user permissions while preserving tree structure
func (s *menuService) filterMenusByPermission(ctx context.Context, menus []*structs.ReadMenu, userPermissions []string, isAdmin bool) []*structs.ReadMenu {
	if isAdmin {
		return menus
	}

	menuMap := make(map[string]*structs.ReadMenu)
	childrenMap := make(map[string][]*structs.ReadMenu)

	for _, menu := range menus {
		menuMap[menu.ID] = menu
		if menu.ParentID != "" && menu.ParentID != "root" {
			childrenMap[menu.ParentID] = append(childrenMap[menu.ParentID], menu)
		}
	}

	var filteredMenus []*structs.ReadMenu
	for _, menu := range menus {
		if menu.ParentID == "" || menu.ParentID == "root" {
			if filteredMenu := s.buildFilteredMenuTree(ctx, menu, menuMap, childrenMap, userPermissions, isAdmin); filteredMenu != nil {
				filteredMenus = append(filteredMenus, filteredMenu)
			}
		}
	}

	return filteredMenus
}

// buildFilteredMenuTree recursively builds filtered menu tree
func (s *menuService) buildFilteredMenuTree(ctx context.Context, menu *structs.ReadMenu, menuMap map[string]*structs.ReadMenu, childrenMap map[string][]*structs.ReadMenu, userPermissions []string, isAdmin bool) *structs.ReadMenu {
	hasDirectPermission := s.hasMenuPermission(ctx, menu, userPermissions, isAdmin)

	var filteredChildren []types.TreeNode
	if children, exists := childrenMap[menu.ID]; exists {
		for _, child := range children {
			if childMenu := s.buildFilteredMenuTree(ctx, child, menuMap, childrenMap, userPermissions, isAdmin); childMenu != nil {
				filteredChildren = append(filteredChildren, childMenu)
			}
		}
	}

	if hasDirectPermission || len(filteredChildren) > 0 {
		filteredMenu := *menu
		filteredMenu.Children = filteredChildren
		return &filteredMenu
	}

	return nil
}

// hasMenuPermission checks if user has permission to access a menu
func (s *menuService) hasMenuPermission(ctx context.Context, menu *structs.ReadMenu, userPermissions []string, isAdmin bool) bool {
	if isAdmin || menu.Perms == "" {
		return true
	}
	return s.checkUserPermission(userPermissions, menu.Perms)
}

// checkUserPermission checks if user has a specific permission
func (s *menuService) checkUserPermission(userPermissions []string, requiredPermission string) bool {
	if requiredPermission == "" {
		return true
	}

	// Check for exact match
	for _, perm := range userPermissions {
		if perm == requiredPermission {
			return true
		}
	}

	// Check for wildcard permissions
	parts := strings.Split(requiredPermission, ":")
	if len(parts) == 2 {
		action, resource := parts[0], parts[1]

		for _, perm := range userPermissions {
			permParts := strings.Split(perm, ":")
			if len(permParts) == 2 {
				permAction, permResource := permParts[0], permParts[1]

				if (permAction == "*" && permResource == resource) ||
					(permAction == action && permResource == "*") ||
					(permAction == "*" && permResource == "*") {
					return true
				}
			}
		}
	}

	return false
}

// GetUserAuthorizedMenus retrieves menus that a user is authorized to access
func (s *menuService) GetUserAuthorizedMenus(ctx context.Context, userID string) ([]*structs.ReadMenu, error) {
	if validator.IsEmpty(userID) {
		return nil, errors.New(ecode.FieldIsRequired("userID"))
	}

	userPermissions := ctxutil.GetUserPermissions(ctx)
	isAdmin := ctxutil.GetUserIsAdmin(ctx)

	allMenus, err := s.GetMenus(ctx, structs.MenuQueryParams{
		ActiveOnly: true,
		Children:   true,
		SortBy:     structs.SortByOrder,
	})
	if err != nil {
		return nil, err
	}

	return s.filterMenusByPermission(ctx, allMenus, userPermissions, isAdmin), nil
}

// BatchGetByID retrieves multiple menus by their IDs.
func (s *menuService) BatchGetByID(ctx context.Context, menuIDs []string) (map[string]*structs.ReadMenu, error) {
	if len(menuIDs) == 0 {
		return make(map[string]*structs.ReadMenu), nil
	}

	result := make(map[string]*structs.ReadMenu)

	for _, id := range menuIDs {
		menuResult, err := s.Get(ctx, &structs.FindMenu{Menu: id})
		if err != nil {
			logger.Warnf(ctx, "Failed to get menu %s: %v", id, err)
			continue
		}

		if menu, ok := menuResult.(*structs.ReadMenu); ok {
			result[id] = menu
		}
	}

	return result, nil
}

// MoveMenu moves a menu to a new parent and/or changes its order.
func (s *menuService) MoveMenu(ctx context.Context, menuID string, newParentID string, newOrder int) (*structs.ReadMenu, error) {
	result, err := s.Get(ctx, &structs.FindMenu{Menu: menuID})
	if err != nil {
		return nil, err
	}

	menu, ok := result.(*structs.ReadMenu)
	if !ok {
		return nil, errors.New(ecode.AssertionFailed("menu"))
	}

	return s.Update(ctx, &structs.UpdateMenuBody{
		ID: menu.ID,
		MenuBody: structs.MenuBody{
			ParentID: newParentID,
			Order:    &newOrder,
		},
	})
}

// ReorderMenus reorders a set of sibling menus.
func (s *menuService) ReorderMenus(ctx context.Context, menuIDs []string) error {
	if len(menuIDs) == 0 {
		return nil
	}

	menus, err := s.BatchGetByID(ctx, menuIDs)
	if err != nil {
		return err
	}

	if len(menus) != len(menuIDs) {
		return errors.New("not all menu IDs were found")
	}

	for i, id := range menuIDs {
		menu := menus[id]
		if menu.Order == i {
			continue
		}

		order := i
		if _, err := s.Update(ctx, &structs.UpdateMenuBody{
			ID:       id,
			MenuBody: structs.MenuBody{Order: &order},
		}); err != nil {
			logger.Errorf(ctx, "Failed to update order for menu %s: %v", id, err)
			return err
		}
	}

	return nil
}

// ToggleStatus handles enable/disable/show/hide operations.
func (s *menuService) ToggleStatus(ctx context.Context, menuID string, action string) (*structs.ReadMenu, error) {
	result, err := s.Get(ctx, &structs.FindMenu{Menu: menuID})
	if err != nil {
		return nil, err
	}

	menu, ok := result.(*structs.ReadMenu)
	if !ok {
		return nil, errors.New(ecode.AssertionFailed("menu"))
	}

	updateBody := &structs.UpdateMenuBody{
		ID:       menu.ID,
		MenuBody: structs.MenuBody{},
	}

	switch action {
	case "enable":
		if !menu.Disabled {
			return menu, nil
		}
		disabled := false
		updateBody.Disabled = &disabled
	case "disable":
		if menu.Disabled {
			return menu, nil
		}
		disabled := true
		updateBody.Disabled = &disabled
	case "show":
		if !menu.Hidden {
			return menu, nil
		}
		hidden := false
		updateBody.Hidden = &hidden
	case "hide":
		if menu.Hidden {
			return menu, nil
		}
		hidden := true
		updateBody.Hidden = &hidden
	default:
		return nil, errors.New("invalid action: " + action)
	}

	return s.Update(ctx, updateBody)
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
		return s.GetMenuTree(ctx, &structs.FindMenu{
			Children: true,
			Menu:     params.Parent,
			Type:     params.Type,
			SortBy:   params.SortBy,
		})
	}

	if params.SortBy == "" {
		params.SortBy = structs.SortByOrder
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

// GetMenuTree retrieves the menu tree with hierarchy.
func (s *menuService) GetMenuTree(ctx context.Context, params *structs.FindMenu) (paging.Result[*structs.ReadMenu], error) {
	rows, err := s.menu.GetMenuTree(ctx, params)
	if err := handleEntError(ctx, "Menu", err); err != nil {
		return paging.Result[*structs.ReadMenu]{}, err
	}

	serializedMenus := s.Serializes(rows)
	treeMenus := s.buildMenuTree(serializedMenus)

	sortBy := s.getDefaultSort(params.SortBy)
	s.sortMenusByField(treeMenus, sortBy)

	return paging.Result[*structs.ReadMenu]{
		Items: treeMenus,
		Total: len(treeMenus),
	}, nil
}

// Helper methods

func (s *menuService) getDefaultSort(sortBy string) string {
	if sortBy == "" {
		return structs.SortByOrder
	}
	return sortBy
}

func (s *menuService) getDefaultLimit(limit int) int {
	if limit <= 0 {
		return 1000
	}
	return limit
}

func (s *menuService) filterByPath(menus []*structs.ReadMenu, path string) []*structs.ReadMenu {
	var filtered []*structs.ReadMenu
	for _, menu := range menus {
		if menu.Path == path {
			filtered = append(filtered, menu)
		}
	}
	return filtered
}

func (s *menuService) buildMenuTree(menus []*structs.ReadMenu) []*structs.ReadMenu {
	if len(menus) == 0 {
		return menus
	}

	menuMap := make(map[string]*structs.ReadMenu)
	for _, menu := range menus {
		menuMap[menu.ID] = menu
		menu.Children = []types.TreeNode{}
	}

	var rootMenus []*structs.ReadMenu

	for _, menu := range menus {
		if menu.ParentID == "" || menu.ParentID == "root" {
			rootMenus = append(rootMenus, menu)
		} else if parent, exists := menuMap[menu.ParentID]; exists {
			parent.Children = append(parent.Children, menu)
		} else {
			rootMenus = append(rootMenus, menu)
		}
	}

	return rootMenus
}

func (s *menuService) sortMenusByField(menus []*structs.ReadMenu, sortBy string) {
	if len(menus) <= 1 {
		return
	}

	sort.Slice(menus, func(i, j int) bool {
		return s.compareMenus(menus[i], menus[j], sortBy)
	})

	for _, menu := range menus {
		if len(menu.Children) > 0 {
			childMenus := make([]*structs.ReadMenu, 0, len(menu.Children))
			for _, child := range menu.Children {
				if childMenu, ok := child.(*structs.ReadMenu); ok {
					childMenus = append(childMenus, childMenu)
				}
			}

			if len(childMenus) > 0 {
				s.sortMenusByField(childMenus, sortBy)
				menu.Children = make([]types.TreeNode, len(childMenus))
				for i, child := range childMenus {
					menu.Children[i] = child
				}
			}
		}
	}
}

func (s *menuService) compareMenus(a, b *structs.ReadMenu, sortBy string) bool {
	switch sortBy {
	case structs.SortByOrder:
		if a.Order != b.Order {
			return a.Order > b.Order
		}
		if a.CreatedAt != nil && b.CreatedAt != nil && *a.CreatedAt != *b.CreatedAt {
			return *a.CreatedAt > *b.CreatedAt
		}
		return a.ID < b.ID
	case structs.SortByCreatedAt:
		if a.CreatedAt != nil && b.CreatedAt != nil && *a.CreatedAt != *b.CreatedAt {
			return *a.CreatedAt > *b.CreatedAt
		}
		if a.Order != b.Order {
			return a.Order > b.Order
		}
		return a.ID < b.ID
	case structs.SortByName:
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		if a.Order != b.Order {
			return a.Order > b.Order
		}
		return a.ID < b.ID
	default:
		if a.Order != b.Order {
			return a.Order > b.Order
		}
		if a.CreatedAt != nil && b.CreatedAt != nil && *a.CreatedAt != *b.CreatedAt {
			return *a.CreatedAt > *b.CreatedAt
		}
		return a.ID < b.ID
	}
}

func (s *menuService) filterActiveMenus(menus []*structs.ReadMenu) []*structs.ReadMenu {
	var activeMenus []*structs.ReadMenu

	for _, menu := range menus {
		if !menu.Hidden && !menu.Disabled {
			activeMenu := *menu

			if len(menu.Children) > 0 {
				childMenus := make([]*structs.ReadMenu, 0)
				for _, child := range menu.Children {
					if childMenu, ok := child.(*structs.ReadMenu); ok {
						childMenus = append(childMenus, childMenu)
					}
				}
				filteredChildren := s.filterActiveMenus(childMenus)
				activeMenu.Children = make([]types.TreeNode, len(filteredChildren))
				for i, child := range filteredChildren {
					activeMenu.Children[i] = child
				}
			}

			activeMenus = append(activeMenus, &activeMenu)
		}
	}

	return activeMenus
}

func (s *menuService) Serializes(rows []*ent.Menu) []*structs.ReadMenu {
	rs := make([]*structs.ReadMenu, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

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
		CreatedBy: &row.CreatedBy,
		CreatedAt: &row.CreatedAt,
		UpdatedBy: &row.UpdatedBy,
		UpdatedAt: &row.UpdatedAt,
	}
}
