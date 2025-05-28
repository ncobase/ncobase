package wrapper

import (
	"context"
	spaceStructs "ncobase/space/structs"

	ext "github.com/ncobase/ncore/extension/types"
)

// UserGroupServiceInterface is the interface for the user group service interface for auth module
type UserGroupServiceInterface interface {
	GetUserGroups(ctx context.Context, u string) ([]*spaceStructs.ReadGroup, error)
}

type SpaceServiceWrapper struct {
	em               ext.ManagerInterface
	userGroupService UserGroupServiceInterface
}

// NewSpaceServiceWrapper creates a new space service wrapper
func NewSpaceServiceWrapper(em ext.ManagerInterface) *SpaceServiceWrapper {
	wrapper := &SpaceServiceWrapper{em: em}
	wrapper.loadServices()
	return wrapper
}

// loadServices loads space services
func (w *SpaceServiceWrapper) loadServices() {
	if userGroupSvc, err := w.em.GetCrossService("space", "UserGroup"); err == nil {
		if service, ok := userGroupSvc.(UserGroupServiceInterface); ok {
			w.userGroupService = service
		}
	}
}

// RefreshServices refreshes service references
func (w *SpaceServiceWrapper) RefreshServices() {
	w.loadServices()
}

// GetUserGroups gets user groups
func (w *SpaceServiceWrapper) GetUserGroups(ctx context.Context, u string) ([]*spaceStructs.ReadGroup, error) {
	if w.userGroupService != nil {
		return w.userGroupService.GetUserGroups(ctx, u)
	}
	return nil, nil
}
