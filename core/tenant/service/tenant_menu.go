package service

import (
	"context"
	"ncobase/tenant/data"
	"ncobase/tenant/data/ent"
	"ncobase/tenant/data/repository"
	"ncobase/tenant/structs"
)

// TenantMenuServiceInterface is the interface for the service.
type TenantMenuServiceInterface interface {
	AddMenuToTenant(ctx context.Context, tenantID, menuID string) (*structs.TenantMenu, error)
	RemoveMenuFromTenant(ctx context.Context, tenantID, menuID string) error
	IsMenuInTenant(ctx context.Context, tenantID, menuID string) (bool, error)
	GetTenantMenus(ctx context.Context, tenantID string) ([]string, error)
	GetMenuTenants(ctx context.Context, menuID string) ([]string, error)
	RemoveAllMenusFromTenant(ctx context.Context, tenantID string) error
	RemoveMenuFromAllTenants(ctx context.Context, menuID string) error
}

// tenantMenuService is the struct for the service.
type tenantMenuService struct {
	tenantMenu repository.TenantMenuRepositoryInterface
}

// NewTenantMenuService creates a new service.
func NewTenantMenuService(d *data.Data) TenantMenuServiceInterface {
	return &tenantMenuService{
		tenantMenu: repository.NewTenantMenuRepository(d),
	}
}

// AddMenuToTenant adds a menu to a tenant.
func (s *tenantMenuService) AddMenuToTenant(ctx context.Context, tenantID, menuID string) (*structs.TenantMenu, error) {
	row, err := s.tenantMenu.Create(ctx, &structs.TenantMenu{
		TenantID: tenantID,
		MenuID:   menuID,
	})
	if err := handleEntError(ctx, "TenantMenu", err); err != nil {
		return nil, err
	}
	return s.SerializeTenantMenu(row), nil
}

// RemoveMenuFromTenant removes a menu from a tenant.
func (s *tenantMenuService) RemoveMenuFromTenant(ctx context.Context, tenantID, menuID string) error {
	err := s.tenantMenu.DeleteByTenantIDAndMenuID(ctx, tenantID, menuID)
	if err := handleEntError(ctx, "TenantMenu", err); err != nil {
		return err
	}
	return nil
}

// IsMenuInTenant checks if a menu belongs to a tenant.
func (s *tenantMenuService) IsMenuInTenant(ctx context.Context, tenantID, menuID string) (bool, error) {
	exists, err := s.tenantMenu.IsMenuInTenant(ctx, tenantID, menuID)
	if err := handleEntError(ctx, "TenantMenu", err); err != nil {
		return false, err
	}
	return exists, nil
}

// GetTenantMenus retrieves all menu IDs for a tenant.
func (s *tenantMenuService) GetTenantMenus(ctx context.Context, tenantID string) ([]string, error) {
	menuIDs, err := s.tenantMenu.GetTenantMenus(ctx, tenantID)
	if err := handleEntError(ctx, "TenantMenu", err); err != nil {
		return nil, err
	}
	return menuIDs, nil
}

// GetMenuTenants retrieves all tenant IDs for a menu.
func (s *tenantMenuService) GetMenuTenants(ctx context.Context, menuID string) ([]string, error) {
	tenantMenus, err := s.tenantMenu.GetByMenuID(ctx, menuID)
	if err := handleEntError(ctx, "TenantMenu", err); err != nil {
		return nil, err
	}

	var tenantIDs []string
	for _, tm := range tenantMenus {
		tenantIDs = append(tenantIDs, tm.TenantID)
	}

	return tenantIDs, nil
}

// RemoveAllMenusFromTenant removes all menus from a tenant.
func (s *tenantMenuService) RemoveAllMenusFromTenant(ctx context.Context, tenantID string) error {
	err := s.tenantMenu.DeleteAllByTenantID(ctx, tenantID)
	if err := handleEntError(ctx, "TenantMenu", err); err != nil {
		return err
	}
	return nil
}

// RemoveMenuFromAllTenants removes a menu from all tenants.
func (s *tenantMenuService) RemoveMenuFromAllTenants(ctx context.Context, menuID string) error {
	err := s.tenantMenu.DeleteAllByMenuID(ctx, menuID)
	if err := handleEntError(ctx, "TenantMenu", err); err != nil {
		return err
	}
	return nil
}

// SerializeTenantMenu serializes a tenant menu relationship.
func (s *tenantMenuService) SerializeTenantMenu(row *ent.TenantMenu) *structs.TenantMenu {
	return &structs.TenantMenu{
		TenantID: row.TenantID,
		MenuID:   row.MenuID,
	}
}
