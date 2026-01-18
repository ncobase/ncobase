package wrapper

import (
	"context"
	"fmt"
	"ncobase/core/space/structs"

	ext "github.com/ncobase/ncore/extension/types"
)

// OrganizationServiceInterface defines organization service interface for space module
type OrganizationServiceInterface interface {
	Get(ctx context.Context, orgID string) (*structs.ReadOrganization, error)
	GetByIDs(ctx context.Context, orgIDs []string) ([]*structs.ReadOrganization, error)
}

// OrganizationServiceWrapper wraps organization service access with fallback behavior
type OrganizationServiceWrapper struct {
	em                  ext.ManagerInterface
	organizationService OrganizationServiceInterface
}

// NewOrganizationServiceWrapper creates a new organization service wrapper
func NewOrganizationServiceWrapper(em ext.ManagerInterface) *OrganizationServiceWrapper {
	wrapper := &OrganizationServiceWrapper{em: em}
	wrapper.loadServices()
	return wrapper
}

// loadServices loads organization services using existing extension manager methods
func (w *OrganizationServiceWrapper) loadServices() {
	// Try to get organization service using existing GetCrossService
	if groupSvc, err := w.em.GetCrossService("organization", "Organization"); err == nil {
		if service, ok := groupSvc.(OrganizationServiceInterface); ok {
			w.organizationService = service
		}
	}
}

// RefreshServices refreshes service references
func (w *OrganizationServiceWrapper) RefreshServices() {
	w.loadServices()
}

// GetOrganizationByIDs get organizations by IDs
func (w *OrganizationServiceWrapper) GetOrganizationByIDs(ctx context.Context, orgIDs []string) ([]*structs.ReadOrganization, error) {
	if w.organizationService != nil {
		return w.organizationService.GetByIDs(ctx, orgIDs)
	}

	return nil, fmt.Errorf("organization service is not available")
}

// GetOrganization gets a single group
func (w *OrganizationServiceWrapper) GetOrganization(ctx context.Context, orgID string) (*structs.ReadOrganization, error) {
	if w.organizationService != nil {
		return w.organizationService.Get(ctx, orgID)
	}

	return nil, fmt.Errorf("organization service is not available")
}

// HasOrganizationService checks if organization service is available
func (w *OrganizationServiceWrapper) HasOrganizationService() bool {
	return w.organizationService != nil
}
