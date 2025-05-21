package service

import (
	"context"
	"fmt"
	accessStructs "ncobase/access/structs"
	"ncobase/initialize/data"
	spaceStructs "ncobase/space/structs"
	userStructs "ncobase/user/structs"
	"time"

	"github.com/ncobase/ncore/logging/logger"
)

// checkOrganizationsInitialized checks if organizations are already initialized.
func (s *Service) checkOrganizationsInitialized(ctx context.Context) error {
	params := &spaceStructs.ListGroupParams{}
	count := s.ss.Group.CountX(ctx, params)
	if count > 0 {
		logger.Infof(ctx, "Organizations already exist, skipping initialization")
		return nil
	}

	return s.initOrganizations(ctx)
}

// initOrganizations initializes the main organizations and associates them with users.
func (s *Service) initOrganizations(ctx context.Context) error {
	logger.Infof(ctx, "Initializing organizational structure...")

	// Get default tenant
	tenant, err := s.ts.Tenant.GetBySlug(ctx, "ncobase")
	if err != nil {
		return fmt.Errorf("failed to get default tenant: %v", err)
	}

	// Get admin user
	admin, err := s.us.User.Get(ctx, "admin")
	if err != nil {
		return fmt.Errorf("failed to get admin user: %v", err)
	}

	// Step 1: Create the main group
	mainGroup := data.OrganizationStructure.MainGroup
	mainGroup.TenantID = &tenant.ID
	mainGroup.CreatedBy = &admin.ID
	mainGroup.UpdatedBy = &admin.ID

	createdMainGroup, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
		GroupBody: mainGroup,
	})
	if err != nil {
		return fmt.Errorf("failed to create main group: %v", err)
	}
	logger.Debugf(ctx, "Created main group: %s", mainGroup.Name)

	// Step 2: Create group-level departments
	for _, dept := range data.OrganizationStructure.GroupDepartments {
		dept.TenantID = &tenant.ID
		dept.CreatedBy = &admin.ID
		dept.UpdatedBy = &admin.ID
		dept.ParentID = &createdMainGroup.ID

		_, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
			GroupBody: dept,
		})
		if err != nil {
			return fmt.Errorf("failed to create group department %s: %v", dept.Name, err)
		}
		logger.Debugf(ctx, "Created group department: %s", dept.Name)
	}

	// Step 3: Create companies
	companies := make([]*spaceStructs.ReadGroup, 0, len(data.OrganizationStructure.Companies))
	for _, company := range data.OrganizationStructure.Companies {
		company.TenantID = &tenant.ID
		company.CreatedBy = &admin.ID
		company.UpdatedBy = &admin.ID
		company.ParentID = &createdMainGroup.ID

		createdCompany, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
			GroupBody: company,
		})
		if err != nil {
			return fmt.Errorf("failed to create company %s: %v", company.Name, err)
		}
		companies = append(companies, createdCompany)
		logger.Debugf(ctx, "Created company: %s", company.Name)
	}

	// Step 4: Create departments and teams for each company
	for _, company := range companies {
		// Get company structure from config
		companyStructure, exists := data.OrganizationStructure.CompanyStructures[company.Slug]
		if !exists {
			logger.Warnf(ctx, "No structure defined for company %s, skipping", company.Slug)
			continue
		}

		// Create departments and teams from company structure
		for _, dept := range companyStructure.Departments {
			deptInfo := dept.Info
			deptInfo.TenantID = &tenant.ID
			deptInfo.CreatedBy = &admin.ID
			deptInfo.UpdatedBy = &admin.ID
			deptInfo.ParentID = &company.ID

			createdDept, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
				GroupBody: deptInfo,
			})
			if err != nil {
				return fmt.Errorf("failed to create department %s: %v", deptInfo.Name, err)
			}
			logger.Debugf(ctx, "Created department: %s", deptInfo.Name)

			// Create teams for this department
			for _, team := range dept.Teams {
				team.TenantID = &tenant.ID
				team.CreatedBy = &admin.ID
				team.UpdatedBy = &admin.ID
				team.ParentID = &createdDept.ID

				_, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
					GroupBody: team,
				})
				if err != nil {
					return fmt.Errorf("failed to create team %s: %v", team.Name, err)
				}
				logger.Debugf(ctx, "Created team: %s", team.Name)
			}
		}

		// Create common departments for each company
		for _, commonDept := range data.CommonDepartments {
			// Format department slug with company slug
			deptInfo := commonDept.Info
			deptInfo.Slug = fmt.Sprintf(deptInfo.Slug, company.Slug)
			deptInfo.TenantID = &tenant.ID
			deptInfo.CreatedBy = &admin.ID
			deptInfo.UpdatedBy = &admin.ID
			deptInfo.ParentID = &company.ID

			createdDept, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
				GroupBody: deptInfo,
			})
			if err != nil {
				return fmt.Errorf("failed to create common department %s: %v", deptInfo.Name, err)
			}
			logger.Debugf(ctx, "Created common department: %s", deptInfo.Name)

			// Create teams for this department
			for _, team := range commonDept.Teams {
				// Format team slug with company slug
				team.Slug = fmt.Sprintf(team.Slug, company.Slug)
				team.TenantID = &tenant.ID
				team.CreatedBy = &admin.ID
				team.UpdatedBy = &admin.ID
				team.ParentID = &createdDept.ID

				_, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
					GroupBody: team,
				})
				if err != nil {
					return fmt.Errorf("failed to create team %s: %v", team.Name, err)
				}
				logger.Debugf(ctx, "Created team: %s", team.Name)
			}
		}
	}

	// Step 5: Create temporary organizations
	for _, tempOrg := range data.OrganizationStructure.TemporaryGroups {
		tempOrg.TenantID = &tenant.ID
		tempOrg.CreatedBy = &admin.ID
		tempOrg.UpdatedBy = &admin.ID
		tempOrg.ParentID = &createdMainGroup.ID

		_, err := s.ss.Group.Create(ctx, &spaceStructs.CreateGroupBody{
			GroupBody: tempOrg,
		})
		if err != nil {
			return fmt.Errorf("failed to create temporary organization %s: %v", tempOrg.Name, err)
		}
		logger.Debugf(ctx, "Created temporary organization: %s", tempOrg.Name)
	}

	// Step 6: Create organization-specific roles and permissions
	for _, orgRole := range data.OrganizationStructure.OrganizationRoles {
		// Create role
		_, err := s.acs.Role.Create(ctx, &accessStructs.CreateRoleBody{
			RoleBody: orgRole.Role,
		})
		if err != nil {
			return fmt.Errorf("failed to create organization role %s: %v", orgRole.Role.Name, err)
		}
		logger.Debugf(ctx, "Created organization role: %s", orgRole.Role.Name)
	}

	// Step 7: Associate users with groups and roles
	users, err := s.us.User.List(ctx, &userStructs.ListUserParams{})
	if err != nil {
		return fmt.Errorf("failed to list users: %v", err)
	}

	for _, user := range users.Items {
		// Associate user with main group
		if _, err := s.ss.UserGroup.AddUserToGroup(ctx, user.ID, createdMainGroup.ID); err != nil {
			return fmt.Errorf("failed to add user %s to main group: %v", user.Username, err)
		}

		// Assign organization role based on user type
		var roleSlug string
		switch user.Username {
		case "super":
			roleSlug = "group-admin"
		case "admin":
			roleSlug = "department-manager"
		default:
			roleSlug = "employee"
		}

		role, err := s.acs.Role.GetBySlug(ctx, roleSlug)
		if err != nil {
			return fmt.Errorf("failed to get role %s: %v", roleSlug, err)
		}

		if err := s.acs.UserRole.AddRoleToUser(ctx, user.ID, role.ID); err != nil {
			return fmt.Errorf("failed to add organization role %s to user %s: %v", roleSlug, user.Username, err)
		}
		logger.Debugf(ctx, "Associated user %s with organization role: %s", user.Username, roleSlug)
	}

	count := s.ss.Group.CountX(ctx, &spaceStructs.ListGroupParams{})
	logger.Infof(ctx, "Organization structure initialization completed, created %d groups", count)

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
	logger.Infof(ctx, "Organization initialization completed successfully")
	return s.state, nil
}
