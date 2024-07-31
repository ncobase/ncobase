package initialize

import (
	"context"
	"fmt"
	accessService "ncobase/feature/access/service"
	authService "ncobase/feature/auth/service"
	groupService "ncobase/feature/group/service"
	tenantService "ncobase/feature/tenant/service"
	userService "ncobase/feature/user/service"
)

// InitializeService is the struct for the initialization service.
type InitializeService struct {
	as  *authService.Service
	us  *userService.Service
	ts  *tenantService.Service
	gs  *groupService.Service
	acs *accessService.Service
}

// New creates a new initDataService.
func New(as *authService.Service, us *userService.Service, ts *tenantService.Service, gs *groupService.Service, acs *accessService.Service) *InitializeService {
	return &InitializeService{
		as:  as,
		us:  us,
		ts:  ts,
		gs:  gs,
		acs: acs,
	}
}

// Execute initializes roles, permissions, Casbin policies, and initial users if necessary.
func (s *InitializeService) Execute() error {
	ctx := context.Background()

	steps := []func(context.Context) error{
		s.checkRolesInitialized,
		s.checkPermissionsInitialized,
		s.checkUsersInitialized,
		s.checkTenantsInitialized,
		s.checkCasbinPoliciesInitialized,
		s.checkGroupsInitialized,
	}

	for _, step := range steps {
		if err := step(ctx); err != nil {
			return fmt.Errorf("initialization step failed: %v", err)
		}
	}

	return nil
}
