package service

import (
	"ncobase/core/space/data"
	"ncobase/core/space/wrapper"

	ext "github.com/ncobase/ncore/extension/types"
)

// Service represents the space service
type Service struct {
	Space             SpaceServiceInterface
	UserSpace         UserSpaceServiceInterface
	UserSpaceRole     UserSpaceRoleServiceInterface
	SpaceQuota        SpaceQuotaServiceInterface
	SpaceSetting      SpaceSettingServiceInterface
	SpaceBilling      SpaceBillingServiceInterface
	SpaceOrganization SpaceOrganizationServiceInterface
	SpaceMenu         SpaceMenuServiceInterface
	SpaceDictionary   SpaceDictionaryServiceInterface
	SpaceOption       SpaceOptionServiceInterface
	gsw               *wrapper.OrganizationServiceWrapper
}

// New creates a new service
func New(d *data.Data, em ext.ManagerInterface) *Service {
	ts := NewSpaceService(d)

	// Create organization service wrapper
	gsw := wrapper.NewOrganizationServiceWrapper(em)

	return &Service{
		Space:             ts,
		UserSpace:         NewUserSpaceService(d, ts),
		UserSpaceRole:     NewUserSpaceRoleService(d),
		SpaceQuota:        NewSpaceQuotaService(d),
		SpaceSetting:      NewSpaceSettingService(d),
		SpaceBilling:      NewSpaceBillingService(d),
		SpaceOrganization: NewSpaceOrganizationService(d, gsw),
		SpaceMenu:         NewSpaceMenuService(d),
		SpaceDictionary:   NewSpaceDictionaryService(d),
		SpaceOption:       NewSpaceOptionService(d),
		gsw:               gsw,
	}
}

// RefreshDependencies refreshes external service dependencies
func (s *Service) RefreshDependencies() {
	s.gsw.RefreshServices()
}
