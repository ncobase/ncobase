package service

import (
	"context"
	"encoding/json"
	"fmt"
	accessService "ncobase/core/access/service"
	authService "ncobase/core/auth/service"
	initConfig "ncobase/plugin/initialize/config"
	orgService "ncobase/core/organization/service"
	spaceService "ncobase/core/space/service"
	systemService "ncobase/core/system/service"
	systemStructs "ncobase/core/system/structs"
	userService "ncobase/core/user/service"
	userStructs "ncobase/core/user/structs"
	"time"

	"ncobase/internal/version"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// State option key for persistence
const stateOptionKey = "system.initialization.state"

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
	Version       string       `json:"version,omitempty"`
	DataMode      string       `json:"data_mode,omitempty"`
}

// Service is the struct for the initialization service.
type Service struct {
	em ext.ManagerInterface

	sys *systemService.Service
	as  *authService.Service
	us  *userService.Service
	ts  *spaceService.Service
	ss  *orgService.Service
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
			Version:       version.GetVersionInfo().Version,
			DataMode:      "website", // Default mode
		},
		visualizationNameToID: make(map[string]string),
		dashboardNameToID:     make(map[string]string),
		analysisNameToID:      make(map[string]string),
	}
}

// SetDependencies sets the dependencies
func (s *Service) SetDependencies(c *initConfig.Config, sys *systemService.Service, as *authService.Service, us *userService.Service, ts *spaceService.Service, ss *orgService.Service, acs *accessService.Service) {
	s.c = c
	s.sys = sys
	s.as = as
	s.us = us
	s.ts = ts
	s.ss = ss
	s.acs = acs

	// Set data mode from config
	if c.Initialization != nil && c.Initialization.DataMode != "" {
		s.state.DataMode = c.Initialization.DataMode
	}
}

// SetDataMode sets the data mode for initialization
func (s *Service) SetDataMode(mode string) error {
	validModes := map[string]bool{
		"website":    true,
		"company":    true,
		"enterprise": true,
	}

	if !validModes[mode] {
		return fmt.Errorf("invalid data mode: %s, must be one of: website, company, enterprise", mode)
	}

	s.state.DataMode = mode
	return nil
}

// GetDataMode returns the current data mode
func (s *Service) GetDataMode() string {
	return s.state.DataMode
}

// LoadState loads initialization state from the database if available
func (s *Service) LoadState(ctx context.Context) error {
	if s.sys == nil || !s.c.Initialization.PersistState {
		return nil
	}

	option, err := s.sys.Option.GetByName(ctx, stateOptionKey)
	if err != nil {
		return nil
	}

	if option != nil && option.Value != "" {
		var state InitState
		if err := json.Unmarshal([]byte(option.Value), &state); err != nil {
			return fmt.Errorf("failed to parse initialization state: %w", err)
		}

		s.state = &state
		logger.Infof(ctx, "Loaded initialization state from database: initialized=%v, mode=%s, last run=%v",
			state.IsInitialized, state.DataMode, time.Unix(0, state.LastRunTime*int64(time.Millisecond)))
	}

	return nil
}

// SaveState persists the initialization state to the database
func (s *Service) SaveState(ctx context.Context) error {
	if s.sys == nil || !s.c.Initialization.PersistState {
		return nil
	}

	stateJSON, err := json.Marshal(s.state)
	if err != nil {
		return fmt.Errorf("failed to marshal initialization state: %w", err)
	}

	existingOption, err := s.sys.Option.GetByName(ctx, stateOptionKey)
	if err == nil && existingOption != nil {
		updateBody := &systemStructs.UpdateOptionBody{
			ID: existingOption.ID,
			OptionBody: systemStructs.OptionBody{
				Value:    string(stateJSON),
				Autoload: true,
			},
		}
		_, err = s.sys.Option.Update(ctx, updateBody)
		if err != nil {
			return fmt.Errorf("failed to update initialization state option: %w", err)
		}
	} else {
		createBody := &systemStructs.OptionBody{
			Name:     stateOptionKey,
			Type:     "object",
			Value:    string(stateJSON),
			Autoload: true,
		}
		_, err = s.sys.Option.Create(ctx, createBody)
		if err != nil {
			return fmt.Errorf("failed to create initialization state option: %w", err)
		}
	}

	logger.Debugf(ctx, "Saved initialization state to database")
	return nil
}

// IsInitialized checks if the system has been initialized
func (s *Service) IsInitialized(ctx context.Context) bool {
	if s.state.IsInitialized {
		return true
	}

	userCount := s.us.User.CountX(ctx, &userStructs.ListUserParams{})
	if userCount > 0 {
		s.state.IsInitialized = true
		if err := s.SaveState(ctx); err != nil {
			logger.Warnf(ctx, "Failed to save initialization state: %v", err)
		}
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

// Execute initializes all system components
func (s *Service) Execute(ctx context.Context, allowReinitialization bool) (*InitState, error) {
	logger.Infof(ctx, "Starting system initialization in %s mode...", s.state.DataMode)

	if s.IsInitialized(ctx) && !allowReinitialization {
		logger.Infof(ctx, "System is already initialized")
		return s.state, fmt.Errorf("system is already initialized")
	}

	// Pre-initialization validation
	if err := s.validateInitializationConsistency(ctx); err != nil {
		logger.Warnf(ctx, "Pre-initialization validation warnings: %v", err)
		// Continue with initialization despite warnings
	}

	steps := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"roles", s.checkRolesInitialized},
		{"permissions", s.checkPermissionsInitialized},
		{"spaces", s.checkSpacesInitialized},
		{"users", s.checkUsersInitialized}, // This now handles employees too
		{"casbin_policies", s.checkCasbinPoliciesInitialized},
		{"menus", s.checkMenusInitialized},
		{"options", s.checkOptionsInitialized},
		{"dictionaries", s.checkDictionariesInitialized},
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

			s.state.LastRunTime = time.Now().UnixMilli()
			if s.c.Initialization.PersistState {
				if err := s.SaveState(ctx); err != nil {
					logger.Warnf(ctx, "Failed to save initialization state: %v", err)
				}
			}

			return s.state, fmt.Errorf("initialization step %s failed: %v", step.name, err)
		}

		s.state.Statuses = append(s.state.Statuses, status)
		logger.Infof(ctx, "Successfully initialized %s", step.name)
	}

	s.state.IsInitialized = true
	s.state.LastRunTime = time.Now().UnixMilli()

	if s.c.Initialization.PersistState {
		if err := s.SaveState(ctx); err != nil {
			logger.Warnf(ctx, "Failed to save initialization state: %v", err)
		}
	}

	logger.Infof(ctx, "System initialization completed successfully in %s mode", s.state.DataMode)
	return s.state, nil
}

// ResetInitialization resets the initialization state if allowed
func (s *Service) ResetInitialization(ctx context.Context) (*InitState, error) {
	if !s.c.Initialization.AllowReinitialization {
		return s.state, fmt.Errorf("reinitialization is not allowed in configuration")
	}

	logger.Warnf(ctx, "Resetting system initialization state")

	// Reset in-memory state
	s.state = &InitState{
		IsInitialized: false,
		Statuses:      make([]InitStatus, 0),
		LastRunTime:   time.Now().UnixMilli(),
		Version:       s.state.Version,
	}

	// Persist reset state
	if s.c.Initialization.PersistState {
		if err := s.SaveState(ctx); err != nil {
			logger.Warnf(ctx, "Failed to save initialization state: %v", err)
		}
	}

	return s.state, nil
}

// GetState returns the current initialization state
func (s *Service) GetState() *InitState {
	return s.state
}
