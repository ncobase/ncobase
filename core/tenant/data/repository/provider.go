package repository

import (
	"ncobase/tenant/data"
)

// Repository represents the tenant repository
type Repository struct {
	tenant        TenantRepositoryInterface
	userTenant    UserTenantRepositoryInterface
	tenantQuota   TenantQuotaRepositoryInterface
	tenantSetting TenantSettingRepositoryInterface
	tenantBilling TenantBillingRepositoryInterface
}

// New creates a new repository
func New(d *data.Data) *Repository {
	return &Repository{
		tenant:        NewTenantRepository(d),
		userTenant:    NewUserTenantRepository(d),
		tenantQuota:   NewTenantQuotaRepository(d),
		tenantSetting: NewTenantSettingRepository(d),
		tenantBilling: NewTenantBillingRepository(d),
	}
}
