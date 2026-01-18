package middleware

import (
	"context"
	"fmt"
	accessStructs "ncobase/core/access/structs"
	authStructs "ncobase/core/auth/structs"
	spaceStructs "ncobase/core/space/structs"
	userStructs "ncobase/core/user/structs"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/ncobase/ncore/data/paging"
	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/security/jwt"
)

// ServiceManager manages all service wrappers
type ServiceManager struct {
	em        ext.ManagerInterface
	authSvc   *AuthServiceWrapper
	userSvc   *UserServiceWrapper
	accessSvc *AccessServiceWrapper
	spaceSvc  *SpaceServiceWrapper
	once      sync.Once
}

var (
	serviceManager *ServiceManager
	managerOnce    sync.Once
)

// GetServiceManager returns singleton service manager instance
func GetServiceManager(em ext.ManagerInterface) *ServiceManager {
	managerOnce.Do(func() {
		serviceManager = &ServiceManager{em: em}
	})
	return serviceManager
}

// AuthServiceWrapper returns auth service wrapper
func (sm *ServiceManager) AuthServiceWrapper() *AuthServiceWrapper {
	sm.once.Do(sm.initServices)
	return sm.authSvc
}

// UserServiceWrapper returns user service wrapper
func (sm *ServiceManager) UserServiceWrapper() *UserServiceWrapper {
	sm.once.Do(sm.initServices)
	return sm.userSvc
}

// AccessServiceWrapper returns access service wrapper
func (sm *ServiceManager) AccessServiceWrapper() *AccessServiceWrapper {
	sm.once.Do(sm.initServices)
	return sm.accessSvc
}

// SpaceServiceWrapper returns space service wrapper
func (sm *ServiceManager) SpaceServiceWrapper() *SpaceServiceWrapper {
	sm.once.Do(sm.initServices)
	return sm.spaceSvc
}

// initServices initializes all service wrappers
func (sm *ServiceManager) initServices() {
	sm.authSvc = &AuthServiceWrapper{em: sm.em}
	sm.userSvc = &UserServiceWrapper{em: sm.em}
	sm.accessSvc = &AccessServiceWrapper{em: sm.em}
	sm.spaceSvc = &SpaceServiceWrapper{em: sm.em}
}

// AuthServiceWrapper wraps auth service calls
type AuthServiceWrapper struct {
	em ext.ManagerInterface
}

// GetTokenManager gets token manager
func (w *AuthServiceWrapper) GetTokenManager() *jwt.TokenManager {
	if authExt, err := w.em.GetExtensionByName("auth"); err == nil {
		if provider, ok := authExt.(interface {
			GetTokenManager() *jwt.TokenManager
		}); ok {
			return provider.GetTokenManager()
		}
	}
	return nil
}

// GetSessionByID gets session by ID
func (w *AuthServiceWrapper) GetSessionByID(ctx context.Context, id string) (*authStructs.ReadSession, error) {
	if svc, err := w.em.GetCrossService("auth", "Session"); err == nil {
		if service, ok := svc.(interface {
			GetByID(context.Context, string) (*authStructs.ReadSession, error)
		}); ok {
			return service.GetByID(ctx, id)
		}
	}
	return nil, fmt.Errorf("session service not available")
}

// DeleteSession deletes session
func (w *AuthServiceWrapper) DeleteSession(ctx context.Context, id string) error {
	if svc, err := w.em.GetCrossService("auth", "Session"); err == nil {
		if service, ok := svc.(interface {
			Delete(context.Context, string) error
		}); ok {
			return service.Delete(ctx, id)
		}
	}
	return fmt.Errorf("session service not available")
}

// UpdateSessionLastAccess updates session last access
func (w *AuthServiceWrapper) UpdateSessionLastAccess(ctx context.Context, tokenID string) error {
	if svc, err := w.em.GetCrossService("auth", "Session"); err == nil {
		if service, ok := svc.(interface {
			UpdateLastAccess(context.Context, string) error
		}); ok {
			return service.UpdateLastAccess(ctx, tokenID)
		}
	}
	return fmt.Errorf("session service not available")
}

// GetSessionByTokenID gets session by token ID
func (w *AuthServiceWrapper) GetSessionByTokenID(ctx context.Context, tokenID string) (*authStructs.ReadSession, error) {
	if svc, err := w.em.GetCrossService("auth", "Session"); err == nil {
		if service, ok := svc.(interface {
			GetByTokenID(context.Context, string) (*authStructs.ReadSession, error)
		}); ok {
			return service.GetByTokenID(ctx, tokenID)
		}
	}
	return nil, fmt.Errorf("session service not available")
}

// DeactivateSessionByTokenID deactivates session by token ID
func (w *AuthServiceWrapper) DeactivateSessionByTokenID(ctx context.Context, tokenID string) error {
	if svc, err := w.em.GetCrossService("auth", "Session"); err == nil {
		if service, ok := svc.(interface {
			DeactivateByTokenID(context.Context, string) error
		}); ok {
			return service.DeactivateByTokenID(ctx, tokenID)
		}
	}
	return fmt.Errorf("session service not available")
}

// CleanupExpiredSessions cleans up expired sessions
func (w *AuthServiceWrapper) CleanupExpiredSessions(ctx context.Context) error {
	if svc, err := w.em.GetCrossService("auth", "Session"); err == nil {
		if service, ok := svc.(interface {
			CleanupExpiredSessions(context.Context) error
		}); ok {
			return service.CleanupExpiredSessions(ctx)
		}
	}
	return fmt.Errorf("session service not available")
}

// GetActiveSessionsCount gets active sessions count
func (w *AuthServiceWrapper) GetActiveSessionsCount(ctx context.Context, userID string) int {
	if svc, err := w.em.GetCrossService("auth", "Session"); err == nil {
		if service, ok := svc.(interface {
			GetActiveSessionsCount(context.Context, string) int
		}); ok {
			return service.GetActiveSessionsCount(ctx, userID)
		}
	}
	return 0
}

// UserServiceWrapper wraps user service calls
type UserServiceWrapper struct {
	em ext.ManagerInterface
}

// GetUserByID gets user by ID
func (w *UserServiceWrapper) GetUserByID(ctx context.Context, id string) (*userStructs.ReadUser, error) {
	if userSvc, err := w.em.GetCrossService("user", "User"); err == nil {
		if service, ok := userSvc.(interface {
			GetByID(context.Context, string) (*userStructs.ReadUser, error)
		}); ok {
			return service.GetByID(ctx, id)
		}
	}
	return nil, fmt.Errorf("user service not available")
}

// ValidateApiKey validates API key
func (w *UserServiceWrapper) ValidateApiKey(ctx context.Context, key string) (*userStructs.ApiKey, error) {
	if userSvc, err := w.em.GetCrossService("user", "ApiKey"); err == nil {
		if service, ok := userSvc.(interface {
			ValidateApiKey(context.Context, string) (*userStructs.ApiKey, error)
		}); ok {
			return service.ValidateApiKey(ctx, key)
		}
	}
	return nil, fmt.Errorf("user service not available")
}

// AccessServiceWrapper wraps access service calls
type AccessServiceWrapper struct {
	em ext.ManagerInterface
}

// GetUserRoles gets user roles
func (w *AccessServiceWrapper) GetUserRoles(ctx context.Context, userID string) ([]*accessStructs.ReadRole, error) {
	if svc, err := w.em.GetCrossService("access", "UserRole"); err == nil {
		if service, ok := svc.(interface {
			GetUserRoles(context.Context, string) ([]*accessStructs.ReadRole, error)
		}); ok {
			return service.GetUserRoles(ctx, userID)
		}
	}
	return nil, fmt.Errorf("user role service not available")
}

// GetUserRolesInSpace gets user roles in space
func (w *AccessServiceWrapper) GetUserRolesInSpace(ctx context.Context, userID, spaceID string) ([]string, error) {
	if svc, err := w.em.GetCrossService("access", "UserSpaceRole"); err == nil {
		if service, ok := svc.(interface {
			GetUserRolesInSpace(context.Context, string, string) ([]string, error)
		}); ok {
			return service.GetUserRolesInSpace(ctx, userID, spaceID)
		}
	}
	return nil, fmt.Errorf("user space role service not available")
}

// GetRolesByIDs gets roles by IDs
func (w *AccessServiceWrapper) GetRolesByIDs(ctx context.Context, roleIDs []string) ([]*accessStructs.ReadRole, error) {
	if svc, err := w.em.GetCrossService("access", "Role"); err == nil {
		if service, ok := svc.(interface {
			GetByIDs(context.Context, []string) ([]*accessStructs.ReadRole, error)
		}); ok {
			return service.GetByIDs(ctx, roleIDs)
		}
	}
	return nil, fmt.Errorf("role service not available")
}

// GetRolePermissions gets role permissions
func (w *AccessServiceWrapper) GetRolePermissions(ctx context.Context, roleID string) ([]*accessStructs.ReadPermission, error) {
	if svc, err := w.em.GetCrossService("access", "RolePermission"); err == nil {
		if service, ok := svc.(interface {
			GetRolePermissions(context.Context, string) ([]*accessStructs.ReadPermission, error)
		}); ok {
			return service.GetRolePermissions(ctx, roleID)
		}
	}
	return nil, fmt.Errorf("role permission service not available")
}

// GetEnforcer gets casbin enforcer
func (w *AccessServiceWrapper) GetEnforcer() *casbin.Enforcer {
	if svc, err := w.em.GetCrossService("access", "CasbinAdapter"); err == nil {
		if service, ok := svc.(interface {
			GetEnforcer() *casbin.Enforcer
		}); ok {
			return service.GetEnforcer()
		}
	}
	return nil
}

// SpaceServiceWrapper wraps space service calls
type SpaceServiceWrapper struct {
	em ext.ManagerInterface
}

// GetUserSpaces gets user spaces
func (w *SpaceServiceWrapper) GetUserSpaces(ctx context.Context, userID string) ([]*spaceStructs.ReadSpace, error) {
	if svc, err := w.em.GetCrossService("space", "UserSpace"); err == nil {
		if service, ok := svc.(interface {
			UserBelongSpaces(context.Context, string) ([]*spaceStructs.ReadSpace, error)
		}); ok {
			return service.UserBelongSpaces(ctx, userID)
		}
	}
	return nil, fmt.Errorf("user space service not available")
}

// GetUserDefaultSpace gets user default space
func (w *SpaceServiceWrapper) GetUserDefaultSpace(ctx context.Context, userID string) (*spaceStructs.ReadSpace, error) {
	if svc, err := w.em.GetCrossService("space", "UserSpace"); err == nil {
		if service, ok := svc.(interface {
			UserBelongSpace(context.Context, string) (*spaceStructs.ReadSpace, error)
		}); ok {
			return service.UserBelongSpace(ctx, userID)
		}
	}
	return nil, fmt.Errorf("user space service not available")
}

// IsSpaceInUser checks if space belongs to user
func (w *SpaceServiceWrapper) IsSpaceInUser(ctx context.Context, spaceID, userID string) (bool, error) {
	if svc, err := w.em.GetCrossService("space", "UserSpace"); err == nil {
		if service, ok := svc.(interface {
			IsSpaceInUser(context.Context, string, string) (bool, error)
		}); ok {
			return service.IsSpaceInUser(ctx, spaceID, userID)
		}
	}
	return false, fmt.Errorf("user space service not available")
}

// ListSpaces lists spaces
func (w *SpaceServiceWrapper) ListSpaces(ctx context.Context, params *spaceStructs.ListSpaceParams) (paging.Result[*spaceStructs.ReadSpace], error) {
	if svc, err := w.em.GetCrossService("space", "Space"); err == nil {
		if service, ok := svc.(interface {
			List(context.Context, *spaceStructs.ListSpaceParams) (paging.Result[*spaceStructs.ReadSpace], error)
		}); ok {
			return service.List(ctx, params)
		}
	}
	return paging.Result[*spaceStructs.ReadSpace]{}, fmt.Errorf("space service not available")
}
