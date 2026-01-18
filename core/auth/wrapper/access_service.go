package wrapper

import (
	"context"
	"fmt"
	accessStructs "ncobase/core/access/structs"

	ext "github.com/ncobase/ncore/extension/types"
)

// RoleServiceInterface defines role service interface for auth module
type RoleServiceInterface interface {
	GetByIDs(ctx context.Context, roleIDs []string) ([]*accessStructs.ReadRole, error)
	Find(ctx context.Context, p *accessStructs.FindRole) (*accessStructs.ReadRole, error)
	CreateSuperAdminRole(ctx context.Context) (*accessStructs.ReadRole, error)
}

// RolePermissionServiceInterface defines role permission service interface for auth module
type RolePermissionServiceInterface interface {
	GetRolePermissions(ctx context.Context, r string) ([]*accessStructs.ReadPermission, error)
}

// UserRoleServiceInterface defines user role service interface for auth module
type UserRoleServiceInterface interface {
	AddRoleToUser(ctx context.Context, u string, r string) error
	GetUserRoles(ctx context.Context, userID string) ([]*accessStructs.ReadRole, error)
}

// AccessServiceWrapper wraps access service access
type AccessServiceWrapper struct {
	em                    ext.ManagerInterface
	roleService           RoleServiceInterface
	rolePermissionService RolePermissionServiceInterface
	userRoleService       UserRoleServiceInterface
}

// NewAccessServiceWrapper creates a new access service wrapper
func NewAccessServiceWrapper(em ext.ManagerInterface) *AccessServiceWrapper {
	wrapper := &AccessServiceWrapper{em: em}
	wrapper.loadServices()
	return wrapper
}

// loadServices loads access services
func (w *AccessServiceWrapper) loadServices() {
	if roleSvc, err := w.em.GetCrossService("access", "Role"); err == nil {
		if service, ok := roleSvc.(RoleServiceInterface); ok {
			w.roleService = service
		}
	}
	if rolePermissionSvc, err := w.em.GetCrossService("access", "RolePermission"); err == nil {
		if service, ok := rolePermissionSvc.(RolePermissionServiceInterface); ok {
			w.rolePermissionService = service
		}
	}
	if userRoleSvc, err := w.em.GetCrossService("access", "UserRole"); err == nil {
		if service, ok := userRoleSvc.(UserRoleServiceInterface); ok {
			w.userRoleService = service
		}
	}

}

// RefreshServices refreshes service references
func (w *AccessServiceWrapper) RefreshServices() {
	w.loadServices()
}

// GetByIDs gets roles by ids
func (w *AccessServiceWrapper) GetByIDs(ctx context.Context, roleIDs []string) ([]*accessStructs.ReadRole, error) {
	if w.roleService != nil {
		return w.roleService.GetByIDs(ctx, roleIDs)
	}
	return nil, fmt.Errorf("role service is not available")
}

// FindRole finds role
func (w *AccessServiceWrapper) FindRole(ctx context.Context, p *accessStructs.FindRole) (*accessStructs.ReadRole, error) {
	if w.roleService != nil {
		return w.roleService.Find(ctx, p)
	}
	return nil, fmt.Errorf("role service is not available")
}

// CreateSuperAdminRole creates super admin role
func (w *AccessServiceWrapper) CreateSuperAdminRole(ctx context.Context) (*accessStructs.ReadRole, error) {
	if w.roleService != nil {
		return w.roleService.CreateSuperAdminRole(ctx)
	}
	return nil, fmt.Errorf("role service is not available")
}

// GetRolePermissions gets role permissions
func (w *AccessServiceWrapper) GetRolePermissions(ctx context.Context, r string) ([]*accessStructs.ReadPermission, error) {
	if w.rolePermissionService != nil {
		return w.rolePermissionService.GetRolePermissions(ctx, r)
	}
	return nil, fmt.Errorf("role permission service is not available")
}

// AddRoleToUser adds role to user
func (w *AccessServiceWrapper) AddRoleToUser(ctx context.Context, u, r string) error {
	if w.userRoleService != nil {
		return w.userRoleService.AddRoleToUser(ctx, u, r)
	}
	return fmt.Errorf("user role service is not available")
}

// GetUserRoles gets user roles
func (w *AccessServiceWrapper) GetUserRoles(ctx context.Context, u string) ([]*accessStructs.ReadRole, error) {
	if w.userRoleService != nil {
		return w.userRoleService.GetUserRoles(ctx, u)
	}
	return nil, fmt.Errorf("user role service is not available")
}

// HasRoleService checks if role service is available
func (w *AccessServiceWrapper) HasRoleService() bool {
	return w.roleService != nil
}

// HasRolePermissionService checks if role permission service is available
func (w *AccessServiceWrapper) HasRolePermissionService() bool {
	return w.rolePermissionService != nil
}

// HasUserRoleService checks if user role service is available
func (w *AccessServiceWrapper) HasUserRoleService() bool {
	return w.userRoleService != nil
}
