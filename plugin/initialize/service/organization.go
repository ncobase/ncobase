package service

import (
	"context"
	"fmt"
	spaceStructs "ncobase/space/structs"
	"time"

	"github.com/ncobase/ncore/logging/logger"
)

// checkOrganizationsInitialized checks if organizations exist
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

	tenant, err := s.getDefaultTenant(ctx)
	if err != nil {
		return fmt.Errorf("failed to get default tenant: %v", err)
	}

	adminUser, err := s.getAdminUser(ctx, "creating organizational structure")
	if err != nil {
		return fmt.Errorf("failed to get admin user: %v", err)
	}

	var groupCount int

	switch s.state.DataMode {
	case "website":
		groupCount, err = s.initWebsiteOrganization(ctx, tenant.ID, adminUser.ID)
	case "company":
		groupCount, err = s.initCompanyOrganization(ctx, tenant.ID, adminUser.ID)
	case "enterprise":
		groupCount, err = s.initEnterpriseOrganization(ctx, tenant.ID, adminUser.ID)
	default:
		groupCount, err = s.initWebsiteOrganization(ctx, tenant.ID, adminUser.ID)
	}

	if err != nil {
		return fmt.Errorf("failed to initialize organization structure: %v", err)
	}

	logger.Infof(ctx, "Organization structure initialization completed in %s mode, created %d groups",
		s.state.DataMode, groupCount)

	return nil
}

// initWebsiteOrganization initializes simple website organization
func (s *Service) initWebsiteOrganization(ctx context.Context, tenantID, adminID string) (int, error) {
	logger.Infof(ctx, "Initializing website organizational structure...")

	mainGroup := spaceStructs.GroupBody{
		Name:        "Website",
		Slug:        "website-platform",
		Type:        "website",
		Description: "Main website organization",
		CreatedBy:   &adminID,
		UpdatedBy:   &adminID,
	}

	group, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
		GroupBody: mainGroup,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create main website group: %v", err)
	}

	if err := s.addGroupToTenant(ctx, tenantID, group.ID); err != nil {
		return 0, fmt.Errorf("failed to add group to tenant: %v", err)
	}

	return 1, nil
}

// initCompanyOrganization initializes company organizational structure
func (s *Service) initCompanyOrganization(ctx context.Context, tenantID, adminID string) (int, error) {
	logger.Infof(ctx, "Initializing company organizational structure...")

	groups := []spaceStructs.GroupBody{
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
	var companyGroupID string

	for i, groupBody := range groups {
		group, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
			GroupBody: groupBody,
		})
		if err != nil {
			return createdCount, fmt.Errorf("failed to create group %s: %v", groupBody.Name, err)
		}

		if err := s.addGroupToTenant(ctx, tenantID, group.ID); err != nil {
			return createdCount, fmt.Errorf("failed to add group %s to tenant: %v", groupBody.Name, err)
		}

		// Set parent relationships for hierarchical structure
		if i == 0 {
			companyGroupID = group.ID
		} else if companyGroupID != "" {
			if err := s.updateGroupParent(ctx, group.ID, companyGroupID); err != nil {
				logger.Warnf(ctx, "Failed to set parent for group %s: %v", groupBody.Name, err)
			}
		}

		logger.Debugf(ctx, "Created group: %s", groupBody.Name)
		createdCount++
	}

	return createdCount, nil
}

// initEnterpriseOrganization initializes enterprise organizational structure
func (s *Service) initEnterpriseOrganization(ctx context.Context, tenantID, adminID string) (int, error) {
	logger.Infof(ctx, "Initializing enterprise organizational structure...")

	groups := []spaceStructs.GroupBody{
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
	var enterpriseGroupID string

	for i, groupBody := range groups {
		group, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
			GroupBody: groupBody,
		})
		if err != nil {
			return createdCount, fmt.Errorf("failed to create group %s: %v", groupBody.Name, err)
		}

		if err := s.addGroupToTenant(ctx, tenantID, group.ID); err != nil {
			return createdCount, fmt.Errorf("failed to add group %s to tenant: %v", groupBody.Name, err)
		}

		// Set parent relationships for hierarchical structure
		if i == 0 {
			enterpriseGroupID = group.ID
		} else if enterpriseGroupID != "" {
			if err := s.updateGroupParent(ctx, group.ID, enterpriseGroupID); err != nil {
				logger.Warnf(ctx, "Failed to set parent for group %s: %v", groupBody.Name, err)
			}
		}

		logger.Debugf(ctx, "Created group: %s", groupBody.Name)
		createdCount++
	}

	return createdCount, nil
}

// addGroupToTenant adds a group to a tenant using the TenantGroup service
func (s *Service) addGroupToTenant(ctx context.Context, tenantID, groupID string) error {
	if s.ts.TenantGroup == nil {
		logger.Warnf(ctx, "TenantGroup service not available, skipping tenant-group relationship")
		return nil
	}

	_, err := s.ts.TenantGroup.AddGroupToTenant(ctx, tenantID, groupID)
	if err != nil {
		return fmt.Errorf("failed to add group %s to tenant %s: %v", groupID, tenantID, err)
	}

	logger.Debugf(ctx, "Added group %s to tenant %s", groupID, tenantID)
	return nil
}

// updateGroupParent updates the parent of a group
func (s *Service) updateGroupParent(ctx context.Context, groupID, parentID string) error {
	updates := map[string]any{
		"parent_id": parentID,
	}

	_, err := s.ss.Group.Update(ctx, groupID, updates)
	if err != nil {
		return fmt.Errorf("failed to update parent for group %s: %v", groupID, err)
	}

	logger.Debugf(ctx, "Set parent %s for group %s", parentID, groupID)
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
