package wrapper

import (
	"context"
	"fmt"
	spaceStructs "ncobase/space/structs"

	ext "github.com/ncobase/ncore/extension/types"
)

// GroupServiceInterface defines group service interface for proxy plugin
type GroupServiceInterface interface {
	Get(ctx context.Context, params *spaceStructs.FindGroup) (*spaceStructs.ReadGroup, error)
}

type SpaceServiceWrapper struct {
	em           ext.ManagerInterface
	groupService GroupServiceInterface
}

// NewSpaceServiceWrapper creates a new space service wrapper
func NewSpaceServiceWrapper(em ext.ManagerInterface) *SpaceServiceWrapper {
	wrapper := &SpaceServiceWrapper{em: em}
	wrapper.loadServices()
	return wrapper
}

// loadServices loads space services
func (w *SpaceServiceWrapper) loadServices() {
	if groupSvc, err := w.em.GetCrossService("space", "Group"); err == nil {
		if service, ok := groupSvc.(GroupServiceInterface); ok {
			w.groupService = service
		}
	}
}

// RefreshServices refreshes service references
func (w *SpaceServiceWrapper) RefreshServices() {
	w.loadServices()
}

// GetGroup gets group
func (w *SpaceServiceWrapper) GetGroup(ctx context.Context, params *spaceStructs.FindGroup) (*spaceStructs.ReadGroup, error) {
	if w.groupService != nil {
		return w.groupService.Get(ctx, params)
	}
	return nil, fmt.Errorf("group service not available")
}

// HasGroupService returns true if group service is available
func (w *SpaceServiceWrapper) HasGroupService() bool {
	return w.groupService != nil
}
