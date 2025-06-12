package wrapper

import (
	"context"
	"fmt"

	ext "github.com/ncobase/ncore/extension/types"
)

// SpaceMenuServiceInterface defines space menu service interface for system module
type SpaceMenuServiceInterface interface {
	GetSpaceMenus(ctx context.Context, spaceID string) ([]string, error)
	IsMenuInSpace(ctx context.Context, spaceID, menuID string) (bool, error)
}

// SpaceServiceWrapper wraps space service access
type SpaceServiceWrapper struct {
	em               ext.ManagerInterface
	spaceMenuService SpaceMenuServiceInterface
}

// NewSpaceServiceWrapper creates a new space service wrapper
func NewSpaceServiceWrapper(em ext.ManagerInterface) *SpaceServiceWrapper {
	wrapper := &SpaceServiceWrapper{em: em}
	wrapper.loadServices()
	return wrapper
}

// loadServices loads space services
func (w *SpaceServiceWrapper) loadServices() {
	if spaceSvc, err := w.em.GetCrossService("space", "SpaceMenu"); err == nil {
		if service, ok := spaceSvc.(SpaceMenuServiceInterface); ok {
			w.spaceMenuService = service
		}
	}
}

// RefreshServices refreshes service references
func (w *SpaceServiceWrapper) RefreshServices() {
	w.loadServices()
}

// GetSpaceMenus gets space menus
func (w *SpaceServiceWrapper) GetSpaceMenus(ctx context.Context, spaceID string) ([]string, error) {
	if w.spaceMenuService != nil {
		return w.spaceMenuService.GetSpaceMenus(ctx, spaceID)
	}
	return nil, fmt.Errorf("space menu service not available")
}

// IsMenuInSpace checks if menu is in space
func (w *SpaceServiceWrapper) IsMenuInSpace(ctx context.Context, spaceID, menuID string) (bool, error) {
	if w.spaceMenuService != nil {
		return w.spaceMenuService.IsMenuInSpace(ctx, spaceID, menuID)
	}
	return false, fmt.Errorf("space menu service not available")
}

// HasSpaceMenuService checks if space menu service is available
func (w *SpaceServiceWrapper) HasSpaceMenuService() bool {
	return w.spaceMenuService != nil
}
