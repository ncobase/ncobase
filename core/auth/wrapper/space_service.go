package wrapper

import (
	"context"
	"fmt"
	spaceStructs "ncobase/space/structs"

	"github.com/ncobase/ncore/data/paging"
	ext "github.com/ncobase/ncore/extension/types"
)

// SpaceServiceInterface defines space service interface for auth module
type SpaceServiceInterface interface {
	Create(ctx context.Context, body *spaceStructs.CreateSpaceBody) (*spaceStructs.ReadSpace, error)
	Get(ctx context.Context, id string) (*spaceStructs.ReadSpace, error)
	GetByUser(ctx context.Context, uid string) (*spaceStructs.ReadSpace, error)
	List(ctx context.Context, params *spaceStructs.ListSpaceParams) (paging.Result[*spaceStructs.ReadSpace], error)
}

type UserSpaceServiceInterface interface {
	AddUserToSpace(ctx context.Context, u string, t string) (*spaceStructs.UserSpace, error)
	UserBelongSpace(ctx context.Context, userID string) (*spaceStructs.ReadSpace, error)
	UserBelongSpaces(ctx context.Context, uid string) ([]*spaceStructs.ReadSpace, error)
}

// UserSpaceRoleServiceInterface defines user space role service interface for auth module
type UserSpaceRoleServiceInterface interface {
	AddRoleToUserInSpace(ctx context.Context, u, t, r string) (*spaceStructs.UserSpaceRole, error)
	GetUserRolesInSpace(ctx context.Context, u, t string) ([]string, error)
}

// SpaceServiceWrapper wraps space service access
type SpaceServiceWrapper struct {
	em                   ext.ManagerInterface
	spaceService         SpaceServiceInterface
	userSpaceService     UserSpaceServiceInterface
	userSpaceRoleService UserSpaceRoleServiceInterface
}

// NewSpaceServiceWrapper creates a new space service wrapper
func NewSpaceServiceWrapper(em ext.ManagerInterface) *SpaceServiceWrapper {
	wrapper := &SpaceServiceWrapper{em: em}
	wrapper.loadServices()
	return wrapper
}

// loadServices loads space services
func (w *SpaceServiceWrapper) loadServices() {
	if spaceSvc, err := w.em.GetCrossService("space", "Space"); err == nil {
		if service, ok := spaceSvc.(SpaceServiceInterface); ok {
			w.spaceService = service
		}
	}

	if userSpaceSvc, err := w.em.GetCrossService("space", "UserSpace"); err == nil {
		if service, ok := userSpaceSvc.(UserSpaceServiceInterface); ok {
			w.userSpaceService = service
		}
	}

	if userSpaceRoleSvc, err := w.em.GetCrossService("space", "UserSpaceRole"); err == nil {
		if service, ok := userSpaceRoleSvc.(UserSpaceRoleServiceInterface); ok {
			w.userSpaceRoleService = service
		}
	}
}

// RefreshServices refreshes service references
func (w *SpaceServiceWrapper) RefreshServices() {
	w.loadServices()
}

// CreateSpace creates space with fallback
func (w *SpaceServiceWrapper) CreateSpace(ctx context.Context, body *spaceStructs.CreateSpaceBody) (*spaceStructs.ReadSpace, error) {
	if w.spaceService != nil {
		return w.spaceService.Create(ctx, body)
	}
	return nil, fmt.Errorf("space service not available")
}

// GetSpace gets space by ID with fallback
func (w *SpaceServiceWrapper) GetSpace(ctx context.Context, id string) (*spaceStructs.ReadSpace, error) {
	if w.spaceService != nil {
		return w.spaceService.Get(ctx, id)
	}
	return nil, fmt.Errorf("space service not available")
}

// GetSpaceByUser gets space by user ID with fallback
func (w *SpaceServiceWrapper) GetSpaceByUser(ctx context.Context, userID string) (*spaceStructs.ReadSpace, error) {
	if w.spaceService != nil {
		return w.spaceService.GetByUser(ctx, userID)
	}
	return nil, fmt.Errorf("space service not available")
}

// ListSpaces lists spaces with fallback
func (w *SpaceServiceWrapper) ListSpaces(ctx context.Context, params *spaceStructs.ListSpaceParams) (paging.Result[*spaceStructs.ReadSpace], error) {
	if w.spaceService != nil {
		return w.spaceService.List(ctx, params)
	}
	return paging.Result[*spaceStructs.ReadSpace]{}, fmt.Errorf("space service not available")
}

// AddUserToSpace adds user to space with fallback
func (w *SpaceServiceWrapper) AddUserToSpace(ctx context.Context, u string, t string) (*spaceStructs.UserSpace, error) {
	if w.userSpaceService != nil {
		return w.userSpaceService.AddUserToSpace(ctx, u, t)
	}
	return nil, fmt.Errorf("user space service not available")
}

// GetUserSpace gets user's space with fallback
func (w *SpaceServiceWrapper) GetUserSpace(ctx context.Context, userID string) (*spaceStructs.ReadSpace, error) {
	if w.userSpaceService != nil {
		return w.userSpaceService.UserBelongSpace(ctx, userID)
	}
	return nil, fmt.Errorf("space service not available")
}

// GetUserSpaces gets user's spaces with fallback
func (w *SpaceServiceWrapper) GetUserSpaces(ctx context.Context, userID string) ([]*spaceStructs.ReadSpace, error) {
	if w.userSpaceService != nil {
		return w.userSpaceService.UserBelongSpaces(ctx, userID)
	}
	return nil, fmt.Errorf("space service not available")
}

// AddRoleToUserInSpace adds role to user in space
func (w *SpaceServiceWrapper) AddRoleToUserInSpace(ctx context.Context, u, t, r string) (*spaceStructs.UserSpaceRole, error) {
	if w.userSpaceRoleService != nil {
		return w.userSpaceRoleService.AddRoleToUserInSpace(ctx, u, t, r)
	}
	return nil, fmt.Errorf("user space role service is not available")
}

// GetUserRolesInSpace gets user roles in space
func (w *SpaceServiceWrapper) GetUserRolesInSpace(ctx context.Context, u, t string) ([]string, error) {
	if w.userSpaceRoleService != nil {
		return w.userSpaceRoleService.GetUserRolesInSpace(ctx, u, t)
	}
	return nil, fmt.Errorf("user space role service is not available")
}

// HasSpaceService checks if space service is available
func (w *SpaceServiceWrapper) HasSpaceService() bool {
	return w.spaceService != nil
}

// HasUserSpaceService checks if user space service is available
func (w *SpaceServiceWrapper) HasUserSpaceService() bool {
	return w.userSpaceService != nil
}

// HasUserSpaceRoleService checks if user space role service is available
func (w *SpaceServiceWrapper) HasUserSpaceRoleService() bool {
	return w.userSpaceRoleService != nil
}
