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
		CreatedBy:   &adminID,
		UpdatedBy:   &adminID,
	}

	group, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
		GroupBody: mainGroup,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create main website group: %v", err)
	}

	// Add group to tenant using TenantGroup service
	if err := s.addGroupToTenant(ctx, tenantID, group.ID); err != nil {
		return 0, fmt.Errorf("failed to add group to tenant: %v", err)
	}

	return 1, nil
}

// initCompanyOrganization initializes company organizational structure
func (s *Service) initCompanyOrganization(ctx context.Context, orgStructure any, tenantID, adminID string) (int, error) {
	logger.Infof(ctx, "Initializing company organizational structure...")

	// Company structure without tenant_id
	groups := []spaceStructs.GroupBody{
		{
			Name:        "Digital Company",
			Slug:        "digital-company",
			Description: "Main company organization",
			CreatedBy:   &adminID,
			UpdatedBy:   &adminID,
		},
		{
			Name:        "Technology Department",
			Slug:        "technology",
			Description: "Technology and development",
			CreatedBy:   &adminID,
			UpdatedBy:   &adminID,
		},
		{
			Name:        "Business Operations",
			Slug:        "business-ops",
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

		// Add group to tenant using TenantGroup service
		if err := s.addGroupToTenant(ctx, tenantID, group.ID); err != nil {
			return createdCount, fmt.Errorf("failed to add group %s to tenant: %v", groupBody.Name, err)
		}

		// Set parent relationships for hierarchical structure
		if i == 0 {
			companyGroupID = group.ID
		} else if companyGroupID != "" {
			// Set parent for departments
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
func (s *Service) initEnterpriseOrganization(ctx context.Context, orgStructure any, tenantID, adminID string) (int, error) {
	logger.Infof(ctx, "Initializing enterprise organizational structure...")

	// Basic enterprise structure without tenant_id
	groups := []spaceStructs.GroupBody{
		{
			Name:        "Digital Enterprise Group",
			Slug:        "digital-enterprise",
			Description: "Main enterprise organization",
			CreatedBy:   &adminID,
			UpdatedBy:   &adminID,
		},
		{
			Name:        "Executive Office",
			Slug:        "executive",
			Description: "Executive leadership",
			CreatedBy:   &adminID,
			UpdatedBy:   &adminID,
		},
		{
			Name:        "Technology Division",
			Slug:        "technology",
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

		// Add group to tenant using TenantGroup service
		if err := s.addGroupToTenant(ctx, tenantID, group.ID); err != nil {
			return createdCount, fmt.Errorf("failed to add group %s to tenant: %v", groupBody.Name, err)
		}

		// Set parent relationships for hierarchical structure
		if i == 0 {
			enterpriseGroupID = group.ID
		} else if enterpriseGroupID != "" {
			// Set parent for divisions
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

// initComplexOrganization initializes complex organizational structure with full hierarchy
func (s *Service) initComplexOrganization(ctx context.Context, orgStructure any, tenantID, adminID string) (int, error) {
	logger.Infof(ctx, "Initializing complex organizational structure...")

	var createdCount int
	var rootGroupID string

	// Create root organization group
	rootGroup := spaceStructs.GroupBody{
		Name:        "Organization",
		Slug:        "organization",
		Description: "Root organizational structure",
		CreatedBy:   &adminID,
		UpdatedBy:   &adminID,
	}

	root, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
		GroupBody: rootGroup,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create root organization: %v", err)
	}

	if err := s.addGroupToTenant(ctx, tenantID, root.ID); err != nil {
		return 0, fmt.Errorf("failed to add root group to tenant: %v", err)
	}

	rootGroupID = root.ID
	createdCount++

	// Create departments
	departments := []spaceStructs.GroupBody{
		{Name: "Human Resources", Slug: "hr", Description: "Human resources management"},
		{Name: "Engineering", Slug: "engineering", Description: "Software engineering"},
		{Name: "Marketing", Slug: "marketing", Description: "Marketing and sales"},
		{Name: "Operations", Slug: "operations", Description: "Business operations"},
	}

	for _, dept := range departments {
		dept.CreatedBy = &adminID
		dept.UpdatedBy = &adminID

		deptGroup, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
			GroupBody: dept,
		})
		if err != nil {
			logger.Errorf(ctx, "Failed to create department %s: %v", dept.Name, err)
			continue
		}

		if err := s.addGroupToTenant(ctx, tenantID, deptGroup.ID); err != nil {
			logger.Warnf(ctx, "Failed to add department %s to tenant: %v", dept.Name, err)
		}

		if err := s.updateGroupParent(ctx, deptGroup.ID, rootGroupID); err != nil {
			logger.Warnf(ctx, "Failed to set parent for department %s: %v", dept.Name, err)
		}

		logger.Debugf(ctx, "Created department: %s", dept.Name)
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
