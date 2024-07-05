package service

import (
	"context"
	"ncobase/core/data/ent"
	"ncobase/core/data/structs"
	"ncobase/helper"
	"sort"

	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/types"
	"ncobase/common/validator"
)

// CreateMenuService creates a new menu.
func (svc *Service) CreateMenuService(ctx context.Context, body *structs.MenuBody) (*resp.Exception, error) {
	if validator.IsEmpty(body.Name) {
		return resp.BadRequest(ecode.FieldIsRequired("name")), nil
	}

	menu, err := svc.menu.Create(ctx, body)
	if exception, err := helper.HandleError("Menu", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: menu,
	}, nil
}

// UpdateMenuService updates an existing menu (full and partial).
func (svc *Service) UpdateMenuService(ctx context.Context, updates *structs.UpdateMenuBody) (*resp.Exception, error) {
	if validator.IsEmpty(updates.ID) {
		return resp.BadRequest(ecode.FieldIsRequired("id")), nil
	}

	menu, err := svc.menu.Update(ctx, updates)
	if exception, err := helper.HandleError("Menu", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializeMenuReply(menu),
	}, nil
}

// GetMenuService retrieves a menu by ID.
func (svc *Service) GetMenuService(ctx context.Context, params *structs.FindMenu) (*resp.Exception, error) {

	if params.Children {
		return svc.GetMenuTreeService(ctx, params)
	}

	menu, err := svc.menu.Get(ctx, params)
	if exception, err := helper.HandleError("Menu", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializeMenuReply(menu),
	}, nil
}

// DeleteMenuService deletes a menu by ID.
func (svc *Service) DeleteMenuService(ctx context.Context, params *structs.FindMenu) (*resp.Exception, error) {
	err := svc.menu.Delete(ctx, params)
	if exception, err := helper.HandleError("Menu", err); exception != nil {
		return exception, err
	}

	return nil, nil
}

// ListMenusService lists all menus.
func (svc *Service) ListMenusService(ctx context.Context, params *structs.ListMenuParams) (*resp.Exception, error) {
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
		return resp.BadRequest(ecode.FieldIsInvalid("limit")), nil
	}

	menus, err := svc.menu.List(ctx, params)
	if exception, err := helper.HandleError("Menu", err); exception != nil {
		return exception, err
	}

	total := svc.menu.CountX(ctx, params)

	return &resp.Exception{
		Data: types.JSON{
			"content": menus,
			"total":   total,
		},
	}, nil
}

// GetMenuTreeService retrieves the menu tree.
func (svc *Service) GetMenuTreeService(ctx context.Context, params *structs.FindMenu) (*resp.Exception, error) {
	menus, err := svc.menu.GetTree(ctx, params)
	if exception, err := helper.HandleError("MenuTree", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: &types.JSON{
			"content": svc.buildMenuTree(menus),
			"total":   len(menus),
		},
	}, nil
}

// serializeMenuReply serializes a menu.
func (svc *Service) serializeMenuReply(row *ent.Menu) *structs.ReadMenu {
	return &structs.ReadMenu{
		ID:       row.ID,
		Name:     row.Name,
		Label:    row.Label,
		Slug:     row.Slug,
		Type:     row.Type,
		Path:     row.Path,
		Target:   row.Target,
		Icon:     row.Icon,
		Perms:    row.Perms,
		Hidden:   row.Hidden,
		Order:    row.Order,
		Disabled: row.Disabled,
		Extras:   &row.Extras,
		ParentID: row.ParentID,
		TenantID: row.TenantID,
		BaseEntity: structs.BaseEntity{
			CreatedBy: &row.CreatedBy,
			CreatedAt: &row.CreatedAt,
			UpdatedBy: &row.UpdatedBy,
			UpdatedAt: &row.UpdatedAt,
		},
	}
}

// buildMenuTree builds a menu tree structure.
func (svc *Service) buildMenuTree(menus []*ent.Menu) []*structs.ReadMenu {
	// Convert menus to ReadMenu objects
	menuNodes := make([]types.TreeNode, len(menus))
	for i, menu := range menus {
		menuNodes[i] = svc.serializeMenuReply(menu)
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
