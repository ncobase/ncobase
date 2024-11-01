package service

import (
	accessService "ncobase/core/access/service"
	authService "ncobase/core/auth/service"
	groupService "ncobase/core/group/service"
	"ncobase/core/linker/service/initialize"
	tenantService "ncobase/core/tenant/service"
	userService "ncobase/core/user/service"
)

// Service is the struct for the relationship service.
type Service struct {
	Initialize initialize.InitializeService
}

// New creates a new relationship service.
func New(as *authService.Service, us *userService.Service, ts *tenantService.Service, gs *groupService.Service, acs *accessService.Service) *Service {
	return &Service{
		Initialize: *initialize.New(as, us, ts, gs, acs),
	}
}
