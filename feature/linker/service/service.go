package service

import (
	accessService "ncobase/feature/access/service"
	authService "ncobase/feature/auth/service"
	tenantService "ncobase/feature/tenant/service"
	userService "ncobase/feature/user/service"
)

// Service is the struct for the relationship service.
type Service struct {
	as  *authService.Service
	us  *userService.Service
	ts  *tenantService.Service
	acs *accessService.Service
}

// New creates a new relationship service.
func New(as *authService.Service, us *userService.Service, ts *tenantService.Service, acs *accessService.Service) *Service {
	return &Service{
		as:  as,
		us:  us,
		ts:  ts,
		acs: acs,
	}
}
