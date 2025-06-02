package repository

import (
	"ncobase/tenant/data"
)

// Repository represents the tenant repository
type Repository struct {
	Tenant         TenantRepositoryInterface
	UserTenant     UserTenantRepositoryInterface
	UserTenantRole UserTenantRoleRepositoryInterface
	TenantQuota    TenantQuotaRepositoryInterface
	TenantSetting  TenantSettingRepositoryInterface
	TenantBilling  TenantBillingRepositoryInterface
	TenantGroup    TenantGroupRepositoryInterface
}

// New creates a new repository
func New(d *data.Data) *Repository {
	return &Repository{
		Tenant:         NewTenantRepository(d),
		UserTenant:     NewUserTenantRepository(d),
		UserTenantRole: NewUserTenantRoleRepository(d),
		TenantQuota:    NewTenantQuotaRepository(d),
		TenantSetting:  NewTenantSettingRepository(d),
		TenantBilling:  NewTenantBillingRepository(d),
		TenantGroup:    NewTenantGroupRepository(d),
	}
}
