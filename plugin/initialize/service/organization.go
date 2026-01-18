package service

import (
	"context"
	"fmt"
	orgStructs "ncobase/core/organization/structs"
	"time"

	"github.com/ncobase/ncore/logging/logger"
)

// checkOrganizationsInitialized checks if organizations exist
func (s *Service) checkOrganizationsInitialized(ctx context.Context) error {
	params := &orgStructs.ListOrganizationParams{}
	count := s.ss.Organization.CountX(ctx, params)
	if count > 0 {
		logger.Infof(ctx, "Organizations already exist, skipping initialization")
		return nil
	}

	return s.initOrganizations(ctx)
}

// initOrganizations initializes organizational structure based on current data mode
func (s *Service) initOrganizations(ctx context.Context) error {
	logger.Infof(ctx, "Initializing organizational structure in %s mode...", s.state.DataMode)

	space, err := s.getDefaultSpace(ctx)
	if err != nil {
		return fmt.Errorf("failed to get default space: %v", err)
	}

	adminUser, err := s.getAdminUser(ctx, "creating organizational structure")
	if err != nil {
		return fmt.Errorf("failed to get admin user: %v", err)
	}

	var organizationCount int

	switch s.state.DataMode {
	case "website":
		organizationCount, err = s.initWebsiteOrganization(ctx, space.ID, adminUser.ID)
	case "company":
		organizationCount, err = s.initCompanyOrganization(ctx, space.ID, adminUser.ID)
	case "enterprise":
		organizationCount, err = s.initEnterpriseOrganization(ctx, space.ID, adminUser.ID)
	default:
		organizationCount, err = s.initWebsiteOrganization(ctx, space.ID, adminUser.ID)
	}

	if err != nil {
		return fmt.Errorf("failed to initialize organization structure: %v", err)
	}

	logger.Infof(ctx, "Organization structure initialization completed in %s mode, created %d organizations",
		s.state.DataMode, organizationCount)

	return nil
}

// initWebsiteOrganization initializes simple website organization
func (s *Service) initWebsiteOrganization(ctx context.Context, spaceID, adminID string) (int, error) {
	logger.Infof(ctx, "Initializing website organizational structure...")

	mainGroup := orgStructs.OrganizationBody{
		Name:        "Website",
		Slug:        "website-platform",
		Type:        "website",
		Description: "Main website organization",
		CreatedBy:   &adminID,
		UpdatedBy:   &adminID,
	}

	organization, err := s.ss.Organization.Create(ctx, &orgStructs.CreateOrganizationBody{
		OrganizationBody: mainGroup,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create main website organization: %v", err)
	}

	if err := s.addGroupToSpace(ctx, spaceID, organization.ID); err != nil {
		return 0, fmt.Errorf("failed to add organization to space: %v", err)
	}

	return 1, nil
}

// initCompanyOrganization initializes company organizational structure
func (s *Service) initCompanyOrganization(ctx context.Context, spaceID, adminID string) (int, error) {
	logger.Infof(ctx, "Initializing company organizational structure...")

	organizations := []orgStructs.OrganizationBody{
		{
			Name:        "Digital Company",
			Slug:        "digital-company",
			Type:        "company",
			Description: "Main company organization",
			CreatedBy:   &adminID,
			UpdatedBy:   &adminID,
		},
		{
			Name:        "Technology Department",
			Slug:        "technology",
			Type:        "department",
			Description: "Technology and development",
			CreatedBy:   &adminID,
			UpdatedBy:   &adminID,
		},
		{
			Name:        "Business Operations",
			Slug:        "business-ops",
			Type:        "department",
			Description: "Business operations and support",
			CreatedBy:   &adminID,
			UpdatedBy:   &adminID,
		},
	}

	var createdCount int
	var companyOrgID string

	for i, organizationBody := range organizations {
		organization, err := s.ss.Organization.Create(ctx, &orgStructs.CreateOrganizationBody{
			OrganizationBody: organizationBody,
		})
		if err != nil {
			return createdCount, fmt.Errorf("failed to create organization %s: %v", organizationBody.Name, err)
		}

		if err := s.addGroupToSpace(ctx, spaceID, organization.ID); err != nil {
			return createdCount, fmt.Errorf("failed to add organization %s to space: %v", organizationBody.Name, err)
		}

		// Set parent relationships for hierarchical structure
		if i == 0 {
			companyOrgID = organization.ID
		} else if companyOrgID != "" {
			if err := s.updateGroupParent(ctx, organization.ID, companyOrgID); err != nil {
				logger.Warnf(ctx, "Failed to set parent for organization %s: %v", organizationBody.Name, err)
			}
		}

		logger.Debugf(ctx, "Created organization: %s", organizationBody.Name)
		createdCount++
	}

	return createdCount, nil
}

// initEnterpriseOrganization initializes enterprise organizational structure
func (s *Service) initEnterpriseOrganization(ctx context.Context, spaceID, adminID string) (int, error) {
	logger.Infof(ctx, "Initializing enterprise organizational structure...")

	organizations := []orgStructs.OrganizationBody{
		{
			Name:        "Digital Enterprise Group",
			Slug:        "digital-enterprise",
			Type:        "enterprise",
			Description: "Main enterprise organization",
			CreatedBy:   &adminID,
			UpdatedBy:   &adminID,
		},
		{
			Name:        "Executive Office",
			Slug:        "executive",
			Type:        "department",
			Description: "Executive leadership",
			CreatedBy:   &adminID,
			UpdatedBy:   &adminID,
		},
		{
			Name:        "Technology Division",
			Slug:        "technology",
			Type:        "division",
			Description: "Technology and innovation",
			CreatedBy:   &adminID,
			UpdatedBy:   &adminID,
		},
	}

	var createdCount int
	var enterpriseOrgID string

	for i, organizationBody := range organizations {
		organization, err := s.ss.Organization.Create(ctx, &orgStructs.CreateOrganizationBody{
			OrganizationBody: organizationBody,
		})
		if err != nil {
			return createdCount, fmt.Errorf("failed to create organization %s: %v", organizationBody.Name, err)
		}

		if err := s.addGroupToSpace(ctx, spaceID, organization.ID); err != nil {
			return createdCount, fmt.Errorf("failed to add organization %s to space: %v", organizationBody.Name, err)
		}

		// Set parent relationships for hierarchical structure
		if i == 0 {
			enterpriseOrgID = organization.ID
		} else if enterpriseOrgID != "" {
			if err := s.updateGroupParent(ctx, organization.ID, enterpriseOrgID); err != nil {
				logger.Warnf(ctx, "Failed to set parent for organization %s: %v", organizationBody.Name, err)
			}
		}

		logger.Debugf(ctx, "Created organization: %s", organizationBody.Name)
		createdCount++
	}

	return createdCount, nil
}

// addGroupToSpace adds a organization to a space using the SpaceOrganization service
func (s *Service) addGroupToSpace(ctx context.Context, spaceID, organizationID string) error {
	if s.ts.SpaceOrganization == nil {
		logger.Warnf(ctx, "SpaceOrganization service not available, skipping space-organization relationship")
		return nil
	}

	_, err := s.ts.SpaceOrganization.AddGroupToSpace(ctx, spaceID, organizationID)
	if err != nil {
		return fmt.Errorf("failed to add organization %s to space %s: %v", organizationID, spaceID, err)
	}

	logger.Debugf(ctx, "Added organization %s to space %s", organizationID, spaceID)
	return nil
}

// updateGroupParent updates the parent of a organization
func (s *Service) updateGroupParent(ctx context.Context, organizationID, parentID string) error {
	updates := map[string]any{
		"parent_id": parentID,
	}

	_, err := s.ss.Organization.Update(ctx, organizationID, updates)
	if err != nil {
		return fmt.Errorf("failed to update parent for organization %s: %v", organizationID, err)
	}

	logger.Debugf(ctx, "Set parent %s for organization %s", parentID, organizationID)
	return nil
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
