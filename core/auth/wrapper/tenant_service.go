package wrapper

import (
	"context"
	"fmt"
	tenantStructs "ncobase/tenant/structs"

	"github.com/ncobase/ncore/data/paging"
	ext "github.com/ncobase/ncore/extension/types"
)

// TenantServiceInterface defines tenant service interface for auth module
type TenantServiceInterface interface {
	Create(ctx context.Context, body *tenantStructs.CreateTenantBody) (*tenantStructs.ReadTenant, error)
	Get(ctx context.Context, id string) (*tenantStructs.ReadTenant, error)
	GetByUser(ctx context.Context, uid string) (*tenantStructs.ReadTenant, error)
	List(ctx context.Context, params *tenantStructs.ListTenantParams) (paging.Result[*tenantStructs.ReadTenant], error)
}

type UserTenantServiceInterface interface {
	AddUserToTenant(ctx context.Context, u string, t string) (*tenantStructs.UserTenant, error)
	UserBelongTenant(ctx context.Context, userID string) (*tenantStructs.ReadTenant, error)
	UserBelongTenants(ctx context.Context, uid string) ([]*tenantStructs.ReadTenant, error)
}

// TenantServiceWrapper wraps tenant service access
type TenantServiceWrapper struct {
	em                ext.ManagerInterface
	tenantService     TenantServiceInterface
	userTenantService UserTenantServiceInterface
}

// NewTenantServiceWrapper creates a new tenant service wrapper
func NewTenantServiceWrapper(em ext.ManagerInterface) *TenantServiceWrapper {
	wrapper := &TenantServiceWrapper{em: em}
	wrapper.loadServices()
	return wrapper
}

// loadServices loads tenant services
func (w *TenantServiceWrapper) loadServices() {
	if tenantSvc, err := w.em.GetCrossService("tenant", "Tenant"); err == nil {
		if service, ok := tenantSvc.(TenantServiceInterface); ok {
			w.tenantService = service
		}
	}

	if userTenantSvc, err := w.em.GetCrossService("tenant", "UserTenant"); err == nil {
		if service, ok := userTenantSvc.(UserTenantServiceInterface); ok {
			w.userTenantService = service
		}
	}
}

// RefreshServices refreshes service references
func (w *TenantServiceWrapper) RefreshServices() {
	w.loadServices()
}

// CreateTenant creates tenant with fallback
func (w *TenantServiceWrapper) CreateTenant(ctx context.Context, body *tenantStructs.CreateTenantBody) (*tenantStructs.ReadTenant, error) {
	if w.tenantService != nil {
		return w.tenantService.Create(ctx, body)
	}
	return nil, fmt.Errorf("tenant service not available")
}

// GetTenant gets tenant by ID with fallback
func (w *TenantServiceWrapper) GetTenant(ctx context.Context, id string) (*tenantStructs.ReadTenant, error) {
	if w.tenantService != nil {
		return w.tenantService.Get(ctx, id)
	}
	return nil, fmt.Errorf("tenant service not available")
}

// GetTenantByUser gets tenant by user ID with fallback
func (w *TenantServiceWrapper) GetTenantByUser(ctx context.Context, userID string) (*tenantStructs.ReadTenant, error) {
	if w.tenantService != nil {
		return w.tenantService.GetByUser(ctx, userID)
	}
	return nil, fmt.Errorf("tenant service not available")
}

// ListTenants lists tenants with fallback
func (w *TenantServiceWrapper) ListTenants(ctx context.Context, params *tenantStructs.ListTenantParams) (paging.Result[*tenantStructs.ReadTenant], error) {
	if w.tenantService != nil {
		return w.tenantService.List(ctx, params)
	}
	return paging.Result[*tenantStructs.ReadTenant]{}, fmt.Errorf("tenant service not available")
}

// AddUserToTenant adds user to tenant with fallback
func (w *TenantServiceWrapper) AddUserToTenant(ctx context.Context, u string, t string) (*tenantStructs.UserTenant, error) {
	if w.userTenantService != nil {
		return w.userTenantService.AddUserToTenant(ctx, u, t)
	}
	return nil, fmt.Errorf("user tenant service not available")
}

// GetUserTenant gets user's tenant with fallback
func (w *TenantServiceWrapper) GetUserTenant(ctx context.Context, userID string) (*tenantStructs.ReadTenant, error) {
	if w.userTenantService != nil {
		return w.userTenantService.UserBelongTenant(ctx, userID)
	}
	return nil, fmt.Errorf("tenant service not available")
}

// GetUserTenants gets user's tenants with fallback
func (w *TenantServiceWrapper) GetUserTenants(ctx context.Context, userID string) ([]*tenantStructs.ReadTenant, error) {
	if w.userTenantService != nil {
		return w.userTenantService.UserBelongTenants(ctx, userID)
	}
	return nil, fmt.Errorf("tenant service not available")
}

// HasTenantService checks if tenant service is available
func (w *TenantServiceWrapper) HasTenantService() bool {
	return w.tenantService != nil
}

// HasUserTenantService checks if user tenant service is available
func (w *TenantServiceWrapper) HasUserTenantService() bool {
	return w.userTenantService != nil
}
