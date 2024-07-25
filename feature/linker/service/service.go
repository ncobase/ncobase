package service

import (
	accessService "ncobase/feature/access/service"
	authService "ncobase/feature/auth/service"
	"ncobase/feature/linker/service/initialize"
	tenantService "ncobase/feature/tenant/service"
	userService "ncobase/feature/user/service"
)

// Service is the struct for the relationship service.
type Service struct {
	Initialize initialize.InitializeService
}

// New creates a new relationship service.
func New(as *authService.Service, us *userService.Service, ts *tenantService.Service, acs *accessService.Service) *Service {
	return &Service{
		Initialize: *initialize.New(as, us, ts, acs),
	}
}
