package service

import (
	"ncobase/tenant/data"
)

// Service represents the tenant service.
type Service struct {
	Tenant     TenantServiceInterface
	UserTenant UserTenantServiceInterface
}

// New creates a new service.
func New(d *data.Data) *Service {
	ts := NewTenantService(d)
	uts := NewUserTenantService(d, ts)

	return &Service{
		Tenant:     ts,
		UserTenant: uts,
	}
}
