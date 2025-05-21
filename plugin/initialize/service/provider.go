package service

import (
	"context"
	"fmt"
	accessService "ncobase/access/service"
	authService "ncobase/auth/service"
	initConfig "ncobase/initialize/config"
	spaceService "ncobase/space/service"
	systemService "ncobase/system/service"
	tenantService "ncobase/tenant/service"
	userService "ncobase/user/service"
	"time"

	ext "github.com/ncobase/ncore/extension/types"
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
	em ext.ManagerInterface

	sys *systemService.Service
	as  *authService.Service
	us  *userService.Service
	ts  *tenantService.Service
	ss  *spaceService.Service
	acs *accessService.Service

	c *initConfig.Config

	state *InitState

	// Maps to store created entity IDs by name for cross-references
	visualizationNameToID map[string]string
	dashboardNameToID     map[string]string
	analysisNameToID      map[string]string
}

// New creates a new initDataService.
func New(
	em ext.ManagerInterface,

) *Service {
	return &Service{
		em: em,
		state: &InitState{
			IsInitialized: false,
			Statuses:      make([]InitStatus, 0),
		},
	}
}

// SetDependencies sets the dependencies
func (s *Service) SetDependencies(c *initConfig.Config, sys *systemService.Service, as *authService.Service, us *userService.Service, ts *tenantService.Service, ss *spaceService.Service, acs *accessService.Service) {
	s.c = c
	s.sys = sys
	s.as = as
	s.us = us
	s.ts = ts
	s.ss = ss
	s.acs = acs
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

// RequiresInitToken returns whether an init token is required
func (s *Service) RequiresInitToken() bool {
	return s.c.Initialization.InitToken != ""
}

// GetInitToken returns the initialization token
func (s *Service) GetInitToken() string {
	return s.c.Initialization.InitToken
}

// AllowReinitialization returns whether reinitialization is allowed
func (s *Service) AllowReinitialization() bool {
	return s.c.Initialization.AllowReinitialization
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
		{"casbin_policies", s.checkCasbinPoliciesInitialized},
		{"menus", s.checkMenusInitialized},
		{"organizations", s.checkOrganizationsInitialized},
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
	s.state.LastRunTime = time.Now().UnixMilli()
	logger.Infof(ctx, "System initialization completed successfully")
	return s.state, nil
}

// GetState returns the current initialization state
func (s *Service) GetState() *InitState {
	return s.state
}
