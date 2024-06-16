package service

import (
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/internal/data/ent"
	"ncobase/internal/data/structs"

	"github.com/gin-gonic/gin"
)

// CreateMenuService creates a new menu.
func (svc *Service) CreateMenuService(c *gin.Context, body *structs.MenuBody) (*resp.Exception, error) {
	if validator.IsEmpty(body.Name) {
		return resp.BadRequest(ecode.FieldIsRequired("name")), nil
	}

	menu, err := svc.menu.Create(c, body)
	if exception, err := handleError("Menu", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: menu,
	}, nil
}

// UpdateMenuService updates an existing menu (full and partial).
func (svc *Service) UpdateMenuService(c *gin.Context, updates *structs.UpdateMenuBody) (*resp.Exception, error) {
	if validator.IsEmpty(updates.ID) {
		return resp.BadRequest(ecode.FieldIsRequired("id")), nil
	}

	menu, err := svc.menu.Update(c, updates)
	if exception, err := handleError("Menu", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializeMenuReply(menu),
	}, nil
}

// GetMenuService retrieves a menu by ID.
func (svc *Service) GetMenuService(c *gin.Context, p *structs.FindMenu) (*resp.Exception, error) {

	if p.Children {
		return svc.GetMenuTreeService(c, p)
	}

	menu, err := svc.menu.Get(c, p)
	if exception, err := handleError("Menu", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializeMenuReply(menu),
	}, nil
}

// DeleteMenuService deletes a menu by ID.
func (svc *Service) DeleteMenuService(c *gin.Context, p *structs.FindMenu) (*resp.Exception, error) {
	err := svc.menu.Delete(c, p)
	if exception, err := handleError("Menu", err); exception != nil {
		return exception, err
	}

	return nil, nil
}

// ListMenusService lists all menus.
func (svc *Service) ListMenusService(c *gin.Context, p *structs.ListMenuParams) (*resp.Exception, error) {
	// with children menu
	if validator.IsTrue(p.Children) {
		return svc.GetMenuTreeService(c, &structs.FindMenu{
			Children: true,
			Tenant:   p.Tenant,
			Menu:     p.Parent,
			Type:     p.Type,
		})
	}

	// limit default value
	if validator.IsEmpty(p.Limit) {
		p.Limit = 20
	}
	// limit must be less than 100
	if p.Limit > 100 {
		return resp.BadRequest(ecode.FieldIsInvalid("limit")), nil
	}

	menus, err := svc.menu.List(c, p)
	if exception, err := handleError("Menu", err); exception != nil {
		return exception, err
	}

	total := svc.menu.CountX(c, p)

	return &resp.Exception{
		Data: types.JSON{
			"content": menus,
			"total":   total,
		},
	}, nil
}

// GetMenuTreeService retrieves the menu tree.
func (svc *Service) GetMenuTreeService(c *gin.Context, p *structs.FindMenu) (*resp.Exception, error) {
	menus, err := svc.menu.GetTree(c, p)
	if exception, err := handleError("MenuTree", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.buildMenuTree(menus),
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
	// // sort menus
	// sort.Slice(menus, func(i, j int) bool {
	// 	if menus[i].Order == menus[j].Order {
	// 		return menus[i].CreatedAt.Before(menus[j].CreatedAt)
	// 	}
	// 	return menus[i].Order < menus[j].Order
	// })

	menuNodes := make([]*structs.ReadMenu, len(menus))
	for i, menu := range menus {
		menuNodes[i] = svc.serializeMenuReply(menu)
	}

	tree := types.BuildTree(menuNodes)

	result := make([]*structs.ReadMenu, len(tree))
	for i, node := range tree {
		result[i] = node
	}

	// // sort result
	// sort.Slice(result, func(i, j int) bool {
	// 	if result[i].Order == result[j].Order {
	// 		return result[i].CreatedAt.Before(types.ToValue(result[j].CreatedAt))
	// 	}
	// 	return result[i].Order < result[j].Order
	// })

	return result
}
