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
	usw               *wrapper.UserServiceWrapper
	rfw               *wrapper.ResourceFileWrapper
}

// New creates a new service
func New(d *data.Data, em ext.ManagerInterface) *Service {
	ts := NewSpaceService(d)

	// Create organization service wrapper
	gsw := wrapper.NewOrganizationServiceWrapper(em)
	usw := wrapper.NewUserServiceWrapper(em)
	rfw := wrapper.NewResourceFileWrapper(em)

	return &Service{
		Space:             ts,
		UserSpace:         NewUserSpaceService(d, ts),
		UserSpaceRole:     NewUserSpaceRoleService(d, usw),
		SpaceQuota:        NewSpaceQuotaService(d),
		SpaceSetting:      NewSpaceSettingService(d),
		SpaceBilling:      NewSpaceBillingService(d),
		SpaceOrganization: NewSpaceOrganizationService(d, gsw),
		SpaceMenu:         NewSpaceMenuService(d),
		SpaceDictionary:   NewSpaceDictionaryService(d),
		SpaceOption:       NewSpaceOptionService(d),
		gsw:               gsw,
		usw:               usw,
		rfw:               rfw,
	}
}

// RefreshDependencies refreshes external service dependencies
func (s *Service) RefreshDependencies() {
	s.gsw.RefreshServices()
	s.usw.RefreshServices()
	s.rfw.RefreshServices()
}

// ResourceFileWrapper exposes resource file wrapper
func (s *Service) ResourceFileWrapper() *wrapper.ResourceFileWrapper {
	return s.rfw
}
