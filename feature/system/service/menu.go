package service

import (
	"context"
	"errors"
	"ncobase/common/ecode"
	"ncobase/common/log"
	"ncobase/common/paging"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/feature"
	"ncobase/feature/system/data"
	"ncobase/feature/system/data/ent"
	"ncobase/feature/system/data/repository"
	"ncobase/feature/system/structs"
	"sort"
)

// MenuServiceInterface represents the menu service interface.
type MenuServiceInterface interface {
	Create(ctx context.Context, body *structs.MenuBody) (*structs.ReadMenu, error)
	Update(ctx context.Context, updates *structs.UpdateMenuBody) (*structs.ReadMenu, error)
	Get(ctx context.Context, params *structs.FindMenu) (any, error)
	Delete(ctx context.Context, params *structs.FindMenu) (*structs.ReadMenu, error)
	List(ctx context.Context, params *structs.ListMenuParams) (paging.Result[*structs.ReadMenu], error)
	GetTree(ctx context.Context, params *structs.FindMenu) (paging.Result[*structs.ReadMenu], error)
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

// Create creates a new menu.
func (s *menuService) Create(ctx context.Context, body *structs.MenuBody) (*structs.ReadMenu, error) {
	if validator.IsEmpty(body.Name) {
		return nil, errors.New(ecode.FieldIsInvalid("name"))
	}

	row, err := s.menu.Create(ctx, body)
	if err := handleEntError("Menu", err); err != nil {
		return nil, err
	}

	// // publish event
	// s.fm.PublishEvent("menu.created", s.Serialize(row))

	return s.Serialize(row), nil
}

// Update updates an existing menu (full and partial).
func (s *menuService) Update(ctx context.Context, updates *structs.UpdateMenuBody) (*structs.ReadMenu, error) {
	if validator.IsEmpty(updates.ID) {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	row, err := s.menu.Update(ctx, updates)
	if err := handleEntError("Menu", err); err != nil {
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
	if err := handleEntError("Menu", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Delete deletes a menu by ID.
func (s *menuService) Delete(ctx context.Context, params *structs.FindMenu) (*structs.ReadMenu, error) {
	err := s.menu.Delete(ctx, params)
	if err := handleEntError("Menu", err); err != nil {
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

		rows, err := s.menu.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			log.Errorf(ctx, "Error listing menus: %v\n", err)
			return nil, 0, err
		}

		total := s.menu.CountX(ctx, params)

		return s.Serializes(rows), total, nil
	})
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

// GetTree retrieves the menu tree.
func (s *menuService) GetTree(ctx context.Context, params *structs.FindMenu) (paging.Result[*structs.ReadMenu], error) {
	rows, err := s.menu.GetTree(ctx, params)
	if err := handleEntError("Menu", err); err != nil {
		return paging.Result[*structs.ReadMenu]{}, err
	}

	return paging.Result[*structs.ReadMenu]{
		Items: s.buildMenuTree(rows),
		Total: len(rows),
	}, nil
}

// buildMenuTree builds a menu tree structure.
func (s *menuService) buildMenuTree(menus []*ent.Menu) []*structs.ReadMenu {
	// Convert menus to ReadMenu objects
	menuNodes := make([]types.TreeNode, len(menus))
	for i, m := range menus {
		menuNodes[i] = s.Serialize(m)
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
