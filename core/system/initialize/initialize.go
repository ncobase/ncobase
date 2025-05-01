package initialize

import (
	"context"
	"fmt"
	accessService "ncobase/core/access/service"
	authService "ncobase/core/auth/service"
	spaceService "ncobase/core/space/service"
	"ncobase/core/system/service"
	tenantService "ncobase/core/tenant/service"
	userService "ncobase/core/user/service"
	"time"

	"github.com/ncobase/ncore/logging/logger"
)

// InitStatus represents the initialization status
type InitStatus struct {
	Component string `json:"component"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
}

// InitState tracks the overall system initialization state
type InitState struct {
	IsInitialized bool         `json:"is_initialized"`
	Statuses      []InitStatus `json:"statuses,omitempty"`
	LastRunTime   int64        `json:"last_run_time,omitempty"`
}

// Service is the struct for the initialization service.
type Service struct {
	menu  service.MenuServiceInterface
	as    *authService.Service
	us    *userService.Service
	ts    *tenantService.Service
	ss    *spaceService.Service
	acs   *accessService.Service
	state *InitState
}

// New creates a new initDataService.
func New(menu service.MenuServiceInterface, as *authService.Service, us *userService.Service, ts *tenantService.Service, ss *spaceService.Service, acs *accessService.Service) *Service {
	return &Service{
		menu: menu,
		as:   as,
		us:   us,
		ts:   ts,
		ss:   ss,
		acs:  acs,
		state: &InitState{
			IsInitialized: false,
			Statuses:      make([]InitStatus, 0),
		},
	}
}

// IsInitialized checks if the system has been initialized
func (s *Service) IsInitialized(ctx context.Context) bool {
	// Check if we already know it's initialized
	if s.state.IsInitialized {
		return true
	}

	// Perform a check to see if required data exists
	userCount := s.us.User.CountX(ctx, nil)
	if userCount > 0 {
		s.state.IsInitialized = true
		return true
	}

	return false
}

// Execute initializes roles, permissions, Casbin policies, and initial users if necessary.
func (s *Service) Execute(ctx context.Context, allowReinitialization bool) (*InitState, error) {
	logger.Infof(ctx, "Starting system initialization...")

	// Check if already initialized
	if s.IsInitialized(ctx) && !allowReinitialization {
		logger.Infof(ctx, "System is already initialized")
		return s.state, fmt.Errorf("system is already initialized")
	}

	steps := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"roles", s.checkRolesInitialized},
		{"permissions", s.checkPermissionsInitialized},
		{"tenants", s.checkTenantsInitialized},
		{"users", s.checkUsersInitialized},
		{"menus", s.checkMenusInitialized},
		{"casbin_policies", s.checkCasbinPoliciesInitialized},
		{"organizational_structure", s.checkGroupsInitialized},
	}

	s.state.Statuses = make([]InitStatus, 0)
	for _, step := range steps {
		status := InitStatus{
			Component: step.name,
			Status:    "initialized",
		}

		logger.Infof(ctx, "Initializing %s...", step.name)
		if err := step.fn(ctx); err != nil {
			status.Status = "failed"
			status.Error = err.Error()
			s.state.Statuses = append(s.state.Statuses, status)
			logger.Errorf(ctx, "Failed to initialize %s: %v", step.name, err)
			return s.state, fmt.Errorf("initialization step %s failed: %v", step.name, err)
		}

		s.state.Statuses = append(s.state.Statuses, status)
		logger.Infof(ctx, "Successfully initialized %s", step.name)
	}

	s.state.IsInitialized = true
	s.state.LastRunTime = time.Now().Unix()
	logger.Infof(ctx, "System initialization completed successfully")
	return s.state, nil
}

// GetState returns the current initialization state
func (s *Service) GetState() *InitState {
	return s.state
}
