package service

import (
	"context"
	"ncobase/core/space/data"
	"ncobase/core/space/data/ent"
	"ncobase/core/space/data/repository"
	"ncobase/core/space/structs"
)

// SpaceMenuServiceInterface is the interface for the service.
type SpaceMenuServiceInterface interface {
	AddMenuToSpace(ctx context.Context, spaceID, menuID string) (*structs.SpaceMenu, error)
	RemoveMenuFromSpace(ctx context.Context, spaceID, menuID string) error
	IsMenuInSpace(ctx context.Context, spaceID, menuID string) (bool, error)
	GetSpaceMenus(ctx context.Context, spaceID string) ([]string, error)
	GetMenuSpaces(ctx context.Context, menuID string) ([]string, error)
	RemoveAllMenusFromSpace(ctx context.Context, spaceID string) error
	RemoveMenuFromAllSpaces(ctx context.Context, menuID string) error
}

// spaceMenuService is the struct for the service.
type spaceMenuService struct {
	spaceMenu repository.SpaceMenuRepositoryInterface
}

// NewSpaceMenuService creates a new service.
func NewSpaceMenuService(d *data.Data) SpaceMenuServiceInterface {
	return &spaceMenuService{
		spaceMenu: repository.NewSpaceMenuRepository(d),
	}
}

// AddMenuToSpace adds a menu to a space.
func (s *spaceMenuService) AddMenuToSpace(ctx context.Context, spaceID, menuID string) (*structs.SpaceMenu, error) {
	row, err := s.spaceMenu.Create(ctx, &structs.SpaceMenu{
		SpaceID: spaceID,
		MenuID:  menuID,
	})
	if err := handleEntError(ctx, "SpaceMenu", err); err != nil {
		return nil, err
	}
	return s.SerializeSpaceMenu(row), nil
}

// RemoveMenuFromSpace removes a menu from a space.
func (s *spaceMenuService) RemoveMenuFromSpace(ctx context.Context, spaceID, menuID string) error {
	err := s.spaceMenu.DeleteBySpaceIDAndMenuID(ctx, spaceID, menuID)
	if err := handleEntError(ctx, "SpaceMenu", err); err != nil {
		return err
	}
	return nil
}

// IsMenuInSpace checks if a menu belongs to a space.
func (s *spaceMenuService) IsMenuInSpace(ctx context.Context, spaceID, menuID string) (bool, error) {
	exists, err := s.spaceMenu.IsMenuInSpace(ctx, spaceID, menuID)
	if err := handleEntError(ctx, "SpaceMenu", err); err != nil {
		return false, err
	}
	return exists, nil
}

// GetSpaceMenus retrieves all menu IDs for a space.
func (s *spaceMenuService) GetSpaceMenus(ctx context.Context, spaceID string) ([]string, error) {
	menuIDs, err := s.spaceMenu.GetSpaceMenus(ctx, spaceID)
	if err := handleEntError(ctx, "SpaceMenu", err); err != nil {
		return nil, err
	}
	return menuIDs, nil
}

// GetMenuSpaces retrieves all space IDs for a menu.
func (s *spaceMenuService) GetMenuSpaces(ctx context.Context, menuID string) ([]string, error) {
	spaceMenus, err := s.spaceMenu.GetByMenuID(ctx, menuID)
	if err := handleEntError(ctx, "SpaceMenu", err); err != nil {
		return nil, err
	}

	var spaceIDs []string
	for _, tm := range spaceMenus {
		spaceIDs = append(spaceIDs, tm.SpaceID)
	}

	return spaceIDs, nil
}

// RemoveAllMenusFromSpace removes all menus from a space.
func (s *spaceMenuService) RemoveAllMenusFromSpace(ctx context.Context, spaceID string) error {
	err := s.spaceMenu.DeleteAllBySpaceID(ctx, spaceID)
	if err := handleEntError(ctx, "SpaceMenu", err); err != nil {
		return err
	}
	return nil
}

// RemoveMenuFromAllSpaces removes a menu from all spaces.
func (s *spaceMenuService) RemoveMenuFromAllSpaces(ctx context.Context, menuID string) error {
	err := s.spaceMenu.DeleteAllByMenuID(ctx, menuID)
	if err := handleEntError(ctx, "SpaceMenu", err); err != nil {
		return err
	}
	return nil
}

// SerializeSpaceMenu serializes a space menu relationship.
func (s *spaceMenuService) SerializeSpaceMenu(row *ent.SpaceMenu) *structs.SpaceMenu {
	return &structs.SpaceMenu{
		SpaceID: row.SpaceID,
		MenuID:  row.MenuID,
	}
}
