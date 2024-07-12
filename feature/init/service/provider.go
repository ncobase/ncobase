package service

import (
	accessService "ncobase/feature/access/service"
	tenantService "ncobase/feature/tenant/service"
	userService "ncobase/feature/user/service"
)

type Service struct {
	Tenant InitTenantServiceInterface
}

func NewService(ts *tenantService.Service, us *userService.Service, as *accessService.Service) *Service {
	return &Service{
		Tenant: NewInitTenantService(ts, us, as),
	}
}
