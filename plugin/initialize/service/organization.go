package service

import (
	"context"
	"fmt"
	accessStructs "ncobase/access/structs"
	"ncobase/initialize/data"
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

// initOrganizations initializes the enterprise organizational structure
func (s *Service) initOrganizations(ctx context.Context) error {
	logger.Infof(ctx, "Initializing enterprise organizational structure...")

	// Get default tenant
	tenant, err := s.ts.Tenant.GetBySlug(ctx, "digital-enterprise")
	if err != nil {
		return fmt.Errorf("failed to get default tenant: %v", err)
	}

	// Get CEO user for creation attribution
	ceo, err := s.us.User.Get(ctx, "chief.executive")
	if err != nil {
		return fmt.Errorf("failed to get CEO user: %v", err)
	}

	var groupCount int

	// Step 1: Create the main enterprise group
	mainGroup := data.EnterpriseOrganizationStructure.Enterprise
	mainGroup.TenantID = &tenant.ID
	mainGroup.CreatedBy = &ceo.ID
	mainGroup.UpdatedBy = &ceo.ID

	createdMainGroup, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
		GroupBody: mainGroup,
	})
	if err != nil {
		return fmt.Errorf("failed to create main enterprise group: %v", err)
	}
	logger.Debugf(ctx, "Created main enterprise group: %s", mainGroup.Name)
	groupCount++

	// Step 2: Create headquarters departments
	for _, hq := range data.EnterpriseOrganizationStructure.Headquarters {
		hq.TenantID = &tenant.ID
		hq.CreatedBy = &ceo.ID
		hq.UpdatedBy = &ceo.ID
		hq.ParentID = &createdMainGroup.ID

		_, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
			GroupBody: hq,
		})
		if err != nil {
			return fmt.Errorf("failed to create headquarters department %s: %v", hq.Name, err)
		}
		logger.Debugf(ctx, "Created headquarters department: %s", hq.Name)
		groupCount++
	}

	// Step 3: Create companies
	companies := make([]*spaceStructs.ReadGroup, 0, len(data.EnterpriseOrganizationStructure.Companies))
	for _, company := range data.EnterpriseOrganizationStructure.Companies {
		company.TenantID = &tenant.ID
		company.CreatedBy = &ceo.ID
		company.UpdatedBy = &ceo.ID
		company.ParentID = &createdMainGroup.ID

		createdCompany, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
			GroupBody: company,
		})
		if err != nil {
			return fmt.Errorf("failed to create company %s: %v", company.Name, err)
		}
		companies = append(companies, createdCompany)
		logger.Debugf(ctx, "Created company: %s", company.Name)
		groupCount++
	}

	// Step 4: Create company-specific departments and teams
	for _, company := range companies {
		companyStructure, exists := data.EnterpriseOrganizationStructure.CompanyStructures[company.Slug]
		if !exists {
			logger.Warnf(ctx, "No structure defined for company %s, skipping", company.Slug)
			continue
		}

		// Create departments for this company
		for _, dept := range companyStructure.Departments {
			deptInfo := dept.Info
			deptInfo.TenantID = &tenant.ID
			deptInfo.CreatedBy = &ceo.ID
			deptInfo.UpdatedBy = &ceo.ID
			deptInfo.ParentID = &company.ID

			createdDept, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
				GroupBody: deptInfo,
			})
			if err != nil {
				return fmt.Errorf("failed to create department %s: %v", deptInfo.Name, err)
			}
			logger.Debugf(ctx, "Created department: %s", deptInfo.Name)
			groupCount++

			// Create teams for this department
			for _, team := range dept.Teams {
				team.TenantID = &tenant.ID
				team.CreatedBy = &ceo.ID
				team.UpdatedBy = &ceo.ID
				team.ParentID = &createdDept.ID

				_, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
					GroupBody: team,
				})
				if err != nil {
					return fmt.Errorf("failed to create team %s: %v", team.Name, err)
				}
				logger.Debugf(ctx, "Created team: %s", team.Name)
				groupCount++
			}
		}

		// Create shared departments for each company
		for _, sharedDept := range data.EnterpriseOrganizationStructure.SharedDepartments {
			deptInfo := sharedDept.Info
			deptInfo.Slug = fmt.Sprintf(deptInfo.Slug, company.Slug)
			deptInfo.TenantID = &tenant.ID
			deptInfo.CreatedBy = &ceo.ID
			deptInfo.UpdatedBy = &ceo.ID
			deptInfo.ParentID = &company.ID

			createdDept, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
				GroupBody: deptInfo,
			})
			if err != nil {
				return fmt.Errorf("failed to create shared department %s: %v", deptInfo.Name, err)
			}
			logger.Debugf(ctx, "Created shared department: %s", deptInfo.Name)
			groupCount++

			// Create teams for this shared department
			for _, team := range sharedDept.Teams {
				team.Slug = fmt.Sprintf(team.Slug, company.Slug)
				team.TenantID = &tenant.ID
				team.CreatedBy = &ceo.ID
				team.UpdatedBy = &ceo.ID
				team.ParentID = &createdDept.ID

				_, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
					GroupBody: team,
				})
				if err != nil {
					return fmt.Errorf("failed to create team %s: %v", team.Name, err)
				}
				logger.Debugf(ctx, "Created team: %s", team.Name)
				groupCount++
			}
		}
	}

	// Step 5: Create organization-specific roles
	var roleCount int
	for _, orgRole := range data.EnterpriseOrganizationStructure.OrganizationRoles {
		existingRole, err := s.acs.Role.GetBySlug(ctx, orgRole.Role.Slug)
		if err == nil && existingRole != nil {
			logger.Debugf(ctx, "Organization role %s already exists, skipping", orgRole.Role.Slug)
			continue
		}

		_, err = s.acs.Role.Create(ctx, &accessStructs.CreateRoleBody{
			RoleBody: orgRole.Role,
		})
		if err != nil {
			return fmt.Errorf("failed to create organization role %s: %v", orgRole.Role.Name, err)
		}
		logger.Debugf(ctx, "Created organization role: %s", orgRole.Role.Name)
		roleCount++
	}

	logger.Infof(ctx, "Enterprise organization structure initialization completed, created %d groups and %d roles",
		groupCount, roleCount)

	return nil
}

// InitializeOrganizations initializes only the organizations if the system is already initialized
func (s *Service) InitializeOrganizations(ctx context.Context) (*InitState, error) {
	logger.Infof(ctx, "Starting organization initialization...")

	// Check if the system is initialized
	if !s.IsInitialized(ctx) {
		logger.Infof(ctx, "System is not yet initialized")
		return s.state, fmt.Errorf("system is not initialized, please initialize the system first")
	}

	// Initialize just organizations
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

	// Persist state if configured
	if s.c.Initialization.PersistState {
		if err := s.SaveState(ctx); err != nil {
			logger.Warnf(ctx, "Failed to save initialization state: %v", err)
		}
	}

	logger.Infof(ctx, "Organization initialization completed successfully")
	return s.state, nil
}
