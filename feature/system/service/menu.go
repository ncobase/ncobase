package service

import (
	"context"
	"errors"
	"ncobase/feature"
	"ncobase/feature/system/data"
	"ncobase/feature/system/data/ent"
	"ncobase/feature/system/data/repository"
	"ncobase/feature/system/structs"
	"sort"

	"ncobase/common/ecode"
	"ncobase/common/types"
	"ncobase/common/validator"
)

// MenuServiceInterface represents the menu service interface.
type MenuServiceInterface interface {
	CreateMenuService(ctx context.Context, body *structs.MenuBody) (*structs.ReadMenu, error)
	UpdateMenuService(ctx context.Context, updates *structs.UpdateMenuBody) (*structs.ReadMenu, error)
	GetMenuService(ctx context.Context, params *structs.FindMenu) (any, error)
	DeleteMenuService(ctx context.Context, params *structs.FindMenu) (*structs.ReadMenu, error)
	ListMenusService(ctx context.Context, params *structs.ListMenuParams) (*types.JSON, error)
	GetMenuTreeService(ctx context.Context, params *structs.FindMenu) (*types.JSON, error)
}

// MenuService represents the menu service.
type menuService struct {
	menu repository.MenuRepositoryInterface
	fm   *feature.Manager
}

// NewMenuService creates a new menu service.
func NewMenuService(d *data.Data, fm *feature.Manager) MenuServiceInterface {
	return &menuService{
		menu: repository.NewMenuRepository(d),
		fm:   fm,
	}
}

// CreateMenuService creates a new menu.
func (svc *menuService) CreateMenuService(ctx context.Context, body *structs.MenuBody) (*structs.ReadMenu, error) {
	if validator.IsEmpty(body.Name) {
		return nil, errors.New(ecode.FieldIsInvalid("name"))
	}

	row, err := svc.menu.Create(ctx, body)
	if err := handleEntError("Menu", err); err != nil {
		return nil, err
	}

	// // publish event
	// svc.fm.PublishEvent("menu.created", svc.serializeMenuReply(row))

	return svc.serializeMenuReply(row), nil
}

// UpdateMenuService updates an existing menu (full and partial).
func (svc *menuService) UpdateMenuService(ctx context.Context, updates *structs.UpdateMenuBody) (*structs.ReadMenu, error) {
	if validator.IsEmpty(updates.ID) {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	row, err := svc.menu.Update(ctx, updates)
	if err := handleEntError("Menu", err); err != nil {
		return nil, err
	}

	return svc.serializeMenuReply(row), nil
}

// GetMenuService retrieves a menu by ID.
func (svc *menuService) GetMenuService(ctx context.Context, params *structs.FindMenu) (any, error) {

	if params.Children {
		return svc.GetMenuTreeService(ctx, params)
	}

	row, err := svc.menu.Get(ctx, params)
	if err := handleEntError("Menu", err); err != nil {
		return nil, err
	}

	return svc.serializeMenuReply(row), nil
}

// DeleteMenuService deletes a menu by ID.
func (svc *menuService) DeleteMenuService(ctx context.Context, params *structs.FindMenu) (*structs.ReadMenu, error) {
	err := svc.menu.Delete(ctx, params)
	if err := handleEntError("Menu", err); err != nil {
		return nil, err
	}

	return nil, nil
}

// ListMenusService lists all menus.
func (svc *menuService) ListMenusService(ctx context.Context, params *structs.ListMenuParams) (*types.JSON, error) {
	// with children menu
	if validator.IsTrue(params.Children) {
		return svc.GetMenuTreeService(ctx, &structs.FindMenu{
			Children: true,
			Tenant:   params.Tenant,
			Menu:     params.Parent,
			Type:     params.Type,
		})
	}

	// limit default value
	if validator.IsEmpty(params.Limit) {
		params.Limit = 20
	}
	// limit must be less than 100
	if params.Limit > 100 {
		return nil, errors.New(ecode.FieldIsInvalid("limit"))
	}

	rows, err := svc.menu.List(ctx, params)
	if err := handleEntError("Menu", err); err != nil {
		return nil, err
	}

	total := svc.menu.CountX(ctx, params)

	return &types.JSON{
		"content": rows,
		"total":   total,
	}, nil
}

// GetMenuTreeService retrieves the menu tree.
func (svc *menuService) GetMenuTreeService(ctx context.Context, params *structs.FindMenu) (*types.JSON, error) {
	rows, err := svc.menu.GetTree(ctx, params)
	if err := handleEntError("Menu", err); err != nil {
		return nil, err
	}

	return &types.JSON{
		"content": svc.buildMenuTree(rows),
		"total":   len(rows),
	}, nil
}

// serializeMenuReply serializes a menu.
func (svc *menuService) serializeMenuReply(row *ent.Menu) *structs.ReadMenu {
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

// buildMenuTree builds a menu tree structure.
func (svc *menuService) buildMenuTree(menus []*ent.Menu) []*structs.ReadMenu {
	// Convert menus to ReadMenu objects
	menuNodes := make([]types.TreeNode, len(menus))
	for i, m := range menus {
		menuNodes[i] = svc.serializeMenuReply(m)
	}

	// Sort menu nodes
	sortMenuNodes(menuNodes)

	// Build tree structure
	tree := types.BuildTree(menuNodes)

	result := make([]*structs.ReadMenu, len(tree))
	for i, node := range tree {
		result[i] = node.(*structs.ReadMenu)
	}

	return result
}

// sortMenuNodes sorts menu nodes.
func sortMenuNodes(menuNodes []types.TreeNode) {
	// Recursively sort children nodes first
	for _, node := range menuNodes {
		children := node.GetChildren()
		sortMenuNodes(children)

		// Sort children and set back to node
		sort.SliceStable(children, func(i, j int) bool {
			nodeI := children[i].(*structs.ReadMenu)
			nodeJ := children[j].(*structs.ReadMenu)
			// if nodeI.Order == nodeJ.Order {
			// 	return nodeI.CreatedAt.Before(types.ToValue(nodeJ.CreatedAt))
			// }
			return nodeI.Order < nodeJ.Order
		})
		node.SetChildren(children)
	}

	// Sort the immediate children of the current level
	sort.SliceStable(menuNodes, func(i, j int) bool {
		nodeI := menuNodes[i].(*structs.ReadMenu)
		nodeJ := menuNodes[j].(*structs.ReadMenu)
		// if nodeI.Order == nodeJ.Order {
		// 	return nodeI.CreatedAt.Before(types.ToValue(nodeJ.CreatedAt))
		// }
		return nodeI.Order < nodeJ.Order
	})
}
