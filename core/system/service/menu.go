package service

import (
	"context"
	"errors"
	"ncobase/core/system/data"
	"ncobase/core/system/data/ent"
	"ncobase/core/system/data/repository"
	"ncobase/core/system/structs"

	ext "github.com/ncobase/ncore/ext/types"
	"github.com/ncobase/ncore/pkg/ecode"
	"github.com/ncobase/ncore/pkg/logger"
	"github.com/ncobase/ncore/pkg/paging"
	"github.com/ncobase/ncore/pkg/types"
	"github.com/ncobase/ncore/pkg/validator"
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
