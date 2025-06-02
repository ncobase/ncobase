package service

import (
	"ncobase/tenant/data"
)

// Service represents the tenant service
type Service struct {
	Tenant         TenantServiceInterface
	UserTenant     UserTenantServiceInterface
	UserTenantRole UserTenantRoleServiceInterface
	TenantQuota    TenantQuotaServiceInterface
	TenantSetting  TenantSettingServiceInterface
	TenantBilling  TenantBillingServiceInterface
}

// New creates a new service
func New(d *data.Data) *Service {
	ts := NewTenantService(d)
	uts := NewUserTenantService(d, ts)
	utrs := NewUserTenantRoleService(d)
	tqs := NewTenantQuotaService(d)
	tss := NewTenantSettingService(d)
	tbs := NewTenantBillingService(d)

	return &Service{
		Tenant:         ts,
		UserTenant:     uts,
		UserTenantRole: utrs,
		TenantQuota:    tqs,
		TenantSetting:  tss,
		TenantBilling:  tbs,
	}
}
