package service

import (
	"ncobase/tenant/data"
	"ncobase/tenant/wrapper"

	ext "github.com/ncobase/ncore/extension/types"
)

// Service represents the tenant service
type Service struct {
	Tenant           TenantServiceInterface
	UserTenant       UserTenantServiceInterface
	UserTenantRole   UserTenantRoleServiceInterface
	TenantQuota      TenantQuotaServiceInterface
	TenantSetting    TenantSettingServiceInterface
	TenantBilling    TenantBillingServiceInterface
	TenantGroup      TenantGroupServiceInterface
	TenantMenu       TenantMenuServiceInterface
	TenantDictionary TenantDictionaryServiceInterface
	TenantOptions    TenantOptionsServiceInterface
	gsw              *wrapper.SpaceServiceWrapper
}

// New creates a new service
func New(d *data.Data, em ext.ManagerInterface) *Service {
	ts := NewTenantService(d)

	// Create space service wrapper
	gsw := wrapper.NewGroupServiceWrapper(em)

	return &Service{
		Tenant:           ts,
		UserTenant:       NewUserTenantService(d, ts),
		UserTenantRole:   NewUserTenantRoleService(d),
		TenantQuota:      NewTenantQuotaService(d),
		TenantSetting:    NewTenantSettingService(d),
		TenantBilling:    NewTenantBillingService(d),
		TenantGroup:      NewTenantGroupService(d, gsw),
		TenantMenu:       NewTenantMenuService(d),
		TenantDictionary: NewTenantDictionaryService(d),
		TenantOptions:    NewTenantOptionsService(d),
		gsw:              gsw,
	}
}

// RefreshDependencies refreshes external service dependencies
func (s *Service) RefreshDependencies() {
	s.gsw.RefreshServices()
}
