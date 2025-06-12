package wrapper

import (
	"context"
	orgStructs "ncobase/organization/structs"

	ext "github.com/ncobase/ncore/extension/types"
)

// UserOrganizationServiceInterface is the interface for the user organization service interface for auth module
type UserOrganizationServiceInterface interface {
	GetUserGroups(ctx context.Context, u string) ([]*orgStructs.ReadOrganization, error)
}

type OrganizationServiceWrapper struct {
	em                      ext.ManagerInterface
	userOrganizationService UserOrganizationServiceInterface
}

// NewOrganizationServiceWrapper creates a new organization service wrapper
func NewOrganizationServiceWrapper(em ext.ManagerInterface) *OrganizationServiceWrapper {
	wrapper := &OrganizationServiceWrapper{em: em}
	wrapper.loadServices()
	return wrapper
}

// loadServices loads organization services
func (w *OrganizationServiceWrapper) loadServices() {
	if userGroupSvc, err := w.em.GetCrossService("organization", "UserOrganization"); err == nil {
		if service, ok := userGroupSvc.(UserOrganizationServiceInterface); ok {
			w.userOrganizationService = service
		}
	}
}

// RefreshServices refreshes service references
func (w *OrganizationServiceWrapper) RefreshServices() {
	w.loadServices()
}

// GetUserGroups gets user orgs
func (w *OrganizationServiceWrapper) GetUserGroups(ctx context.Context, u string) ([]*orgStructs.ReadOrganization, error) {
	if w.userOrganizationService != nil {
		return w.userOrganizationService.GetUserGroups(ctx, u)
	}
	return nil, nil
}
