package service

import (
	accessService "ncobase/feature/access/service"
	tenantService "ncobase/feature/tenant/service"
	userService "ncobase/feature/user/service"
)

// InitTenantServiceInterface is the interface for the service.
type InitTenantServiceInterface any

// initTenantService is the struct for the service.
type initTenantService struct {
	ts *tenantService.Service
	us *userService.Service
	as *accessService.Service
}

// NewInitTenantService creates a new service.
func NewInitTenantService(ts *tenantService.Service, us *userService.Service, as *accessService.Service) InitTenantServiceInterface {
	return &initTenantService{
		ts: ts,
		us: us,
		as: as,
	}
}
