package service

import (
	"ncobase/tenant/data"
	"ncobase/tenant/wrapper"

	ext "github.com/ncobase/ncore/extension/types"
)

// Service represents the tenant service
type Service struct {
	Tenant         TenantServiceInterface
	UserTenant     UserTenantServiceInterface
	UserTenantRole UserTenantRoleServiceInterface
	TenantQuota    TenantQuotaServiceInterface
	TenantSetting  TenantSettingServiceInterface
	TenantBilling  TenantBillingServiceInterface
	TenantGroup    TenantGroupServiceInterface
	gsw            *wrapper.SpaceServiceWrapper
}

// New creates a new service
func New(d *data.Data, em ext.ManagerInterface) *Service {
	ts := NewTenantService(d)
	uts := NewUserTenantService(d, ts)
	utrs := NewUserTenantRoleService(d)
	tqs := NewTenantQuotaService(d)
	tss := NewTenantSettingService(d)
	tbs := NewTenantBillingService(d)

	// Create space service wrapper
	gsw := wrapper.NewGroupServiceWrapper(em)
	tgs := NewTenantGroupService(d, gsw)

	return &Service{
		Tenant:         ts,
		UserTenant:     uts,
		UserTenantRole: utrs,
		TenantQuota:    tqs,
		TenantSetting:  tss,
		TenantBilling:  tbs,
		TenantGroup:    tgs,
		gsw:            gsw,
	}
}

// RefreshDependencies refreshes external service dependencies
func (s *Service) RefreshDependencies() {
	s.gsw.RefreshServices()
}
