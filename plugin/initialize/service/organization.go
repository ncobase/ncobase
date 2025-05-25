package service

import (
	"context"
	"fmt"
	spaceStructs "ncobase/space/structs"
	"time"

	"github.com/ncobase/ncore/logging/logger"
)

// checkOrganizationsInitialized checks if organizations are already initialized
func (s *Service) checkOrganizationsInitialized(ctx context.Context) error {
	params := &spaceStructs.ListGroupParams{}
	count := s.ss.Group.CountX(ctx, params)
	if count > 0 {
		logger.Infof(ctx, "Organizations already exist, skipping initialization")
		return nil
	}

	return s.initOrganizations(ctx)
}

// initOrganizations initializes the organizational structure using current data mode
func (s *Service) initOrganizations(ctx context.Context) error {
	logger.Infof(ctx, "Initializing organizational structure in %s mode...", s.state.DataMode)

	tenant, err := s.ts.Tenant.GetBySlug(ctx, "digital-enterprise")
	if err != nil {
		return fmt.Errorf("failed to get default tenant: %v", err)
	}

	ea, err := s.getAdminUser(ctx, "creating organizational structure")
	if err != nil {
		return fmt.Errorf("failed to get admin user: %v", err)
	}

	dataLoader := s.getDataLoader()
	orgStructure := dataLoader.GetOrganizationStructure()

	var groupCount int

	// Handle organization structure based on data mode
	if s.state.DataMode == "enterprise" {
		// Enterprise mode - full organizational structure
		groupCount, err = s.initEnterpriseOrganization(ctx, orgStructure, tenant.ID, ea.ID)
	} else {
		// Company mode - simplified structure
		groupCount, err = s.initCompanyOrganization(ctx, orgStructure, tenant.ID, ea.ID)
	}

	if err != nil {
		return fmt.Errorf("failed to initialize organization structure: %v", err)
	}

	logger.Infof(ctx, "Organization structure initialization completed in %s mode, created %d groups",
		s.state.DataMode, groupCount)

	return nil
}

// initEnterpriseOrganization initializes full enterprise organizational structure
func (s *Service) initEnterpriseOrganization(ctx context.Context, orgStructure interface{}, tenantID, adminID string) (int, error) {
	// This would handle the full enterprise structure
	// Implementation would be similar to the existing enterprise logic
	logger.Infof(ctx, "Initializing enterprise organizational structure...")

	// Basic implementation for now - can be expanded
	mainGroup := spaceStructs.GroupBody{
		Name:        "Digital Enterprise Group",
		Slug:        "digital-enterprise",
		Description: "Main enterprise organization",
		TenantID:    &tenantID,
		CreatedBy:   &adminID,
		UpdatedBy:   &adminID,
	}

	_, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
		GroupBody: mainGroup,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create main enterprise group: %v", err)
	}

	return 1, nil
}

// initCompanyOrganization initializes simplified company organizational structure
func (s *Service) initCompanyOrganization(ctx context.Context, orgStructure interface{}, tenantID, adminID string) (int, error) {
	logger.Infof(ctx, "Initializing company organizational structure...")

	// Basic company structure
	groups := []spaceStructs.GroupBody{
		{
			Name:        "Digital Enterprise",
			Slug:        "digital-enterprise",
			Description: "Main company organization",
			TenantID:    &tenantID,
			CreatedBy:   &adminID,
			UpdatedBy:   &adminID,
		},
		{
			Name:        "Technology Department",
			Slug:        "technology",
			Description: "Technology and development",
			TenantID:    &tenantID,
			CreatedBy:   &adminID,
			UpdatedBy:   &adminID,
		},
		{
			Name:        "Business Operations",
			Slug:        "business-ops",
			Description: "Business operations and support",
			TenantID:    &tenantID,
			CreatedBy:   &adminID,
			UpdatedBy:   &adminID,
		},
	}

	var createdCount int
	for _, group := range groups {
		_, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
			GroupBody: group,
		})
		if err != nil {
			return createdCount, fmt.Errorf("failed to create group %s: %v", group.Name, err)
		}
		logger.Debugf(ctx, "Created group: %s", group.Name)
		createdCount++
	}

	return createdCount, nil
}

// InitializeOrganizations initializes only the organizations if the system is already initialized
func (s *Service) InitializeOrganizations(ctx context.Context) (*InitState, error) {
	logger.Infof(ctx, "Starting organization initialization in %s mode...", s.state.DataMode)

	if !s.IsInitialized(ctx) {
		logger.Infof(ctx, "System is not yet initialized")
		return s.state, fmt.Errorf("system is not initialized, please initialize the system first")
	}

	status := InitStatus{
		Component: "organizations",
		Status:    "initialized",
	}

	logger.Infof(ctx, "Initializing organizations...")
	if err := s.checkOrganizationsInitialized(ctx); err != nil {
		status.Status = "failed"
		status.Error = err.Error()
		s.state.Statuses = append(s.state.Statuses, status)
		logger.Errorf(ctx, "Failed to initialize organizations: %v", err)
		return s.state, fmt.Errorf("initialization step organizations failed: %v", err)
	}

	s.state.Statuses = append(s.state.Statuses, status)
	logger.Infof(ctx, "Successfully initialized organizations")

	s.state.LastRunTime = time.Now().UnixMilli()

	if s.c.Initialization.PersistState {
		if err := s.SaveState(ctx); err != nil {
			logger.Warnf(ctx, "Failed to save initialization state: %v", err)
		}
	}

	logger.Infof(ctx, "Organization initialization completed successfully in %s mode", s.state.DataMode)
	return s.state, nil
}
