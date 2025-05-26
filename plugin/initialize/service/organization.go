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

// initOrganizations initializes organizational structure based on current data mode
func (s *Service) initOrganizations(ctx context.Context) error {
	logger.Infof(ctx, "Initializing organizational structure in %s mode...", s.state.DataMode)

	// Get default tenant based on data mode
	var defaultSlug string
	switch s.state.DataMode {
	case "website":
		defaultSlug = "website-platform"
	case "company":
		defaultSlug = "digital-company"
	case "enterprise":
		defaultSlug = "digital-enterprise"
	default:
		defaultSlug = "website-platform"
	}

	tenant, err := s.ts.Tenant.GetBySlug(ctx, defaultSlug)
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

	switch s.state.DataMode {
	case "website":
		groupCount, err = s.initWebsiteOrganization(ctx, orgStructure, tenant.ID, ea.ID)
	case "company":
		groupCount, err = s.initCompanyOrganization(ctx, orgStructure, tenant.ID, ea.ID)
	case "enterprise":
		groupCount, err = s.initEnterpriseOrganization(ctx, orgStructure, tenant.ID, ea.ID)
	default:
		groupCount, err = s.initCompanyOrganization(ctx, orgStructure, tenant.ID, ea.ID)
	}

	if err != nil {
		return fmt.Errorf("failed to initialize organization structure: %v", err)
	}

	logger.Infof(ctx, "Organization structure initialization completed in %s mode, created %d groups",
		s.state.DataMode, groupCount)

	return nil
}

// initWebsiteOrganization initializes simple website organization
func (s *Service) initWebsiteOrganization(ctx context.Context, orgStructure any, tenantID, adminID string) (int, error) {
	logger.Infof(ctx, "Initializing website organizational structure...")

	// Simple website structure - just main group
	mainGroup := spaceStructs.GroupBody{
		Name:        "Website",
		Slug:        "website-platform",
		Description: "Main website organization",
		TenantID:    &tenantID,
		CreatedBy:   &adminID,
		UpdatedBy:   &adminID,
	}

	_, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
		GroupBody: mainGroup,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create main website group: %v", err)
	}

	return 1, nil
}

// initCompanyOrganization initializes company organizational structure
func (s *Service) initCompanyOrganization(ctx context.Context, orgStructure any, tenantID, adminID string) (int, error) {
	logger.Infof(ctx, "Initializing company organizational structure...")

	// Updated company structure
	groups := []spaceStructs.GroupBody{
		{
			Name:        "Digital Company",
			Slug:        "digital-company",
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

// initEnterpriseOrganization initializes enterprise organizational structure
func (s *Service) initEnterpriseOrganization(ctx context.Context, orgStructure any, tenantID, adminID string) (int, error) {
	logger.Infof(ctx, "Initializing enterprise organizational structure...")

	// Basic enterprise structure
	groups := []spaceStructs.GroupBody{
		{
			Name:        "Digital Enterprise Group",
			Slug:        "digital-enterprise",
			Description: "Main enterprise organization",
			TenantID:    &tenantID,
			CreatedBy:   &adminID,
			UpdatedBy:   &adminID,
		},
		{
			Name:        "Executive Office",
			Slug:        "executive",
			Description: "Executive leadership",
			TenantID:    &tenantID,
			CreatedBy:   &adminID,
			UpdatedBy:   &adminID,
		},
		{
			Name:        "Technology Division",
			Slug:        "technology",
			Description: "Technology and innovation",
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

// InitializeOrganizations initializes only organizations
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
