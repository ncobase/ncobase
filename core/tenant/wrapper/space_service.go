package wrapper

import (
	"context"
	"fmt"
	"ncobase/tenant/structs"

	ext "github.com/ncobase/ncore/extension/types"
)

// GroupServiceInterface defines group service interface for tenant module
type GroupServiceInterface interface {
	Get(ctx context.Context, groupID string) (*structs.ReadGroup, error)
	GetByIDs(ctx context.Context, groupIDs []string) ([]*structs.ReadGroup, error)
}

// SpaceServiceWrapper wraps space service access with fallback behavior
type SpaceServiceWrapper struct {
	em           ext.ManagerInterface
	groupService GroupServiceInterface
}

// NewGroupServiceWrapper creates a new space service wrapper
func NewGroupServiceWrapper(em ext.ManagerInterface) *SpaceServiceWrapper {
	wrapper := &SpaceServiceWrapper{em: em}
	wrapper.loadServices()
	return wrapper
}

// loadServices loads group services using existing extension manager methods
func (w *SpaceServiceWrapper) loadServices() {
	// Try to get space service using existing GetCrossService
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

// GetGroupByIDs gets groups by IDs with graceful fallback
func (w *SpaceServiceWrapper) GetGroupByIDs(ctx context.Context, groupIDs []string) ([]*structs.ReadGroup, error) {
	if w.groupService != nil {
		return w.groupService.GetByIDs(ctx, groupIDs)
	}

	return nil, fmt.Errorf("group service is not available")
}

// GetGroup gets a single group with graceful fallback
func (w *SpaceServiceWrapper) GetGroup(ctx context.Context, groupID string) (*structs.ReadGroup, error) {
	if w.groupService != nil {
		return w.groupService.Get(ctx, groupID)
	}

	return nil, fmt.Errorf("group service is not available")
}

// HasGroupService checks if space service is available
func (w *SpaceServiceWrapper) HasGroupService() bool {
	return w.groupService != nil
}
