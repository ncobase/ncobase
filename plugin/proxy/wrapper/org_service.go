package wrapper

import (
	"context"
	"fmt"
	orgStructs "ncobase/organization/structs"

	ext "github.com/ncobase/ncore/extension/types"
)

// OrganizationServiceInterface defines organization service interface for proxy plugin
type OrganizationServiceInterface interface {
	Get(ctx context.Context, params *orgStructs.FindOrganization) (*orgStructs.ReadOrganization, error)
}

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

// loadServices loads organization services
func (w *OrganizationServiceWrapper) loadServices() {
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

// GetOrganization get organization
func (w *OrganizationServiceWrapper) GetOrganization(ctx context.Context, params *orgStructs.FindOrganization) (*orgStructs.ReadOrganization, error) {
	if w.organizationService != nil {
		return w.organizationService.Get(ctx, params)
	}
	return nil, fmt.Errorf("organization service not available")
}

// HasOrganizationService returns true if organization service is available
func (w *OrganizationServiceWrapper) HasOrganizationService() bool {
	return w.organizationService != nil
}
