package wrapper

import (
	"context"
	"fmt"

	ext "github.com/ncobase/ncore/extension/types"
)

// TenantMenuServiceInterface defines tenant menu service interface for system module
type TenantMenuServiceInterface interface {
	GetTenantMenus(ctx context.Context, tenantID string) ([]string, error)
	IsMenuInTenant(ctx context.Context, tenantID, menuID string) (bool, error)
}

// TenantServiceWrapper wraps tenant service access
type TenantServiceWrapper struct {
	em                ext.ManagerInterface
	tenantMenuService TenantMenuServiceInterface
}

// NewTenantServiceWrapper creates a new tenant service wrapper
func NewTenantServiceWrapper(em ext.ManagerInterface) *TenantServiceWrapper {
	wrapper := &TenantServiceWrapper{em: em}
	wrapper.loadServices()
	return wrapper
}

// loadServices loads tenant services
func (w *TenantServiceWrapper) loadServices() {
	if tenantSvc, err := w.em.GetCrossService("tenant", "TenantMenu"); err == nil {
		if service, ok := tenantSvc.(TenantMenuServiceInterface); ok {
			w.tenantMenuService = service
		}
	}
}

// RefreshServices refreshes service references
func (w *TenantServiceWrapper) RefreshServices() {
	w.loadServices()
}

// GetTenantMenus gets tenant menus
func (w *TenantServiceWrapper) GetTenantMenus(ctx context.Context, tenantID string) ([]string, error) {
	if w.tenantMenuService != nil {
		return w.tenantMenuService.GetTenantMenus(ctx, tenantID)
	}
	return nil, fmt.Errorf("tenant menu service not available")
}

// IsMenuInTenant checks if menu is in tenant
func (w *TenantServiceWrapper) IsMenuInTenant(ctx context.Context, tenantID, menuID string) (bool, error) {
	if w.tenantMenuService != nil {
		return w.tenantMenuService.IsMenuInTenant(ctx, tenantID, menuID)
	}
	return false, fmt.Errorf("tenant menu service not available")
}

// HasTenantMenuService checks if tenant menu service is available
func (w *TenantServiceWrapper) HasTenantMenuService() bool {
	return w.tenantMenuService != nil
}
