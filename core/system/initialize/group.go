package initialize

import (
	"context"
	"fmt"
	accessStructs "ncobase/core/access/structs"
	spaceStructs "ncobase/core/space/structs"
	userStructs "ncobase/core/user/structs"
	"ncobase/ncore/logger"
)

// checkGroupsInitialized checks if groups are already initialized.
func (s *Service) checkGroupsInitialized(ctx context.Context) error {
	params := &spaceStructs.ListGroupParams{}
	count := s.ss.Group.CountX(ctx, params)
	if count == 0 {
		return s.initOrganizationStructure(ctx)
	}

	return nil
}

// InitOrganizationStructure initializes the organizational structure, permissions, and associates them with users and tenants.
func (s *Service) initOrganizationStructure(ctx context.Context) error {
	// Step 1: Create the main group (Yedu Group)
	yeduGroup, err := s.createMainGroup(ctx)
	if err != nil {
		return err
	}

	// Step 2: Create group-level departments
	if err := s.createGroupLevelDepartments(ctx, yeduGroup.ID); err != nil {
		return err
	}

	// Step 3: Create companies
	companies, err := s.createCompanies(ctx, yeduGroup.ID)
	if err != nil {
		return err
	}

	// Step 4: Create departments and teams for each company
	for _, company := range companies {
		if err := s.createCompanyDepartmentsAndTeams(ctx, company.ID, company.Slug); err != nil {
			return err
		}
	}

	// Step 5: Create temporary organizations
	if err := s.createTemporaryOrganizations(ctx, yeduGroup.ID); err != nil {
		return err
	}

	// Step 6: Create organization-specific roles and permissions
	if err := s.createOrganizationRolesAndPermissions(ctx); err != nil {
		return err
	}

	// Step 7: Associate users with groups and roles
	if err := s.associateUsersWithGroupsAndRoles(ctx); err != nil {
		return err
	}

	logger.Infof(ctx, "Organization structure initialization completed successfully")
	return nil
}

func (s *Service) createGroupLevelDepartments(ctx context.Context, groupID string) error {
	tenant, err := s.ts.Tenant.GetBySlug(ctx, "ncobase")
	if err != nil {
		return fmt.Errorf("failed to get default tenant: %v", err)
	}

	admin, err := s.us.User.Get(ctx, "admin")
	if err != nil {
		return fmt.Errorf("failed to get admin user: %v", err)
	}

	departments := []spaceStructs.CreateGroupBody{
		{
			GroupBody: spaceStructs.GroupBody{
				Name:      "总经办",
				Slug:      "executive-office",
				ParentID:  &groupID,
				TenantID:  &tenant.ID,
				CreatedBy: &admin.ID,
				UpdatedBy: &admin.ID,
			},
		},
		{
			GroupBody: spaceStructs.GroupBody{
				Name:      "集团人力资源部",
				Slug:      "group-hr-department",
				ParentID:  &groupID,
				TenantID:  &tenant.ID,
				CreatedBy: &admin.ID,
				UpdatedBy: &admin.ID,
			},
		},
		{
			GroupBody: spaceStructs.GroupBody{
				Name:      "集团财务部",
				Slug:      "group-finance-department",
				ParentID:  &groupID,
				TenantID:  &tenant.ID,
				CreatedBy: &admin.ID,
				UpdatedBy: &admin.ID,
			},
		},
	}

	for _, dept := range departments {
		if _, err := s.ss.Group.Create(ctx, &dept); err != nil {
			return fmt.Errorf("failed to create group-level department %s: %v", dept.Name, err)
		}
	}

	return nil
}

func (s *Service) createMainGroup(ctx context.Context) (*spaceStructs.ReadGroup, error) {
	tenant, err := s.ts.Tenant.GetBySlug(ctx, "ncobase")
	if err != nil {
		return nil, fmt.Errorf("failed to get default tenant: %v", err)
	}

	admin, err := s.us.User.Get(ctx, "admin")
	if err != nil {
		return nil, fmt.Errorf("failed to get admin user: %v", err)
	}

	mainGroup := &spaceStructs.CreateGroupBody{
		GroupBody: spaceStructs.GroupBody{
			Name:      "邺都集团",
			Slug:      "yedu-group",
			TenantID:  &tenant.ID,
			CreatedBy: &admin.ID,
			UpdatedBy: &admin.ID,
		},
	}

	createdGroup, err := s.ss.Group.Create(ctx, mainGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to create main group: %v", err)
	}

	return createdGroup, nil
}

func (s *Service) createCompanyDepartmentsAndTeams(ctx context.Context, companyID string, companySlug string) error {
	tenant, err := s.ts.Tenant.GetBySlug(ctx, "ncobase")
	if err != nil {
		return fmt.Errorf("failed to get tenant: %v", err)
	}

	admin, err := s.us.User.Get(ctx, "admin")
	if err != nil {
		return fmt.Errorf("failed to get admin user: %v", err)
	}

	departments := map[string][]struct {
		department spaceStructs.CreateGroupBody
		teams      []spaceStructs.CreateGroupBody
	}{
		"tech-company": {
			{
				department: spaceStructs.CreateGroupBody{
					GroupBody: spaceStructs.GroupBody{
						Name:      "技术部",
						Slug:      "tech-department",
						ParentID:  &companyID,
						TenantID:  &tenant.ID,
						CreatedBy: &admin.ID,
						UpdatedBy: &admin.ID,
					},
				},
				teams: []spaceStructs.CreateGroupBody{
					{GroupBody: spaceStructs.GroupBody{Name: "研发组", Slug: "rd-team"}},
					{GroupBody: spaceStructs.GroupBody{Name: "运维组", Slug: "operations-team"}},
					{GroupBody: spaceStructs.GroupBody{Name: "测试组", Slug: "qa-team"}},
				},
			},
			{
				department: spaceStructs.CreateGroupBody{
					GroupBody: spaceStructs.GroupBody{
						Name:      "产品部",
						Slug:      "product-department",
						ParentID:  &companyID,
						TenantID:  &tenant.ID,
						CreatedBy: &admin.ID,
						UpdatedBy: &admin.ID,
					},
				},
				teams: []spaceStructs.CreateGroupBody{
					{GroupBody: spaceStructs.GroupBody{Name: "产品规划组", Slug: "product-planning-team"}},
					{GroupBody: spaceStructs.GroupBody{Name: "用户体验组", Slug: "ux-team"}},
				},
			},
		},
		"media-company": {
			{
				department: spaceStructs.CreateGroupBody{
					GroupBody: spaceStructs.GroupBody{
						Name:      "内容制作部",
						Slug:      "content-production-department",
						ParentID:  &companyID,
						TenantID:  &tenant.ID,
						CreatedBy: &admin.ID,
						UpdatedBy: &admin.ID,
					},
				},
				teams: []spaceStructs.CreateGroupBody{
					{GroupBody: spaceStructs.GroupBody{Name: "视频制作组", Slug: "video-production-team"}},
					{GroupBody: spaceStructs.GroupBody{Name: "文字编辑组", Slug: "text-editing-team"}},
				},
			},
			{
				department: spaceStructs.CreateGroupBody{
					GroupBody: spaceStructs.GroupBody{
						Name:      "媒体运营部",
						Slug:      "media-operations-department",
						ParentID:  &companyID,
						TenantID:  &tenant.ID,
						CreatedBy: &admin.ID,
						UpdatedBy: &admin.ID,
					},
				},
				teams: []spaceStructs.CreateGroupBody{
					{GroupBody: spaceStructs.GroupBody{Name: "社交媒体组", Slug: "social-media-team"}},
					{GroupBody: spaceStructs.GroupBody{Name: "数据分析组", Slug: "data-analysis-team"}},
				},
			},
		},
	}

	// Common departments for both companies
	commonDepartments := []struct {
		department spaceStructs.CreateGroupBody
		teams      []spaceStructs.CreateGroupBody
	}{
		{
			department: spaceStructs.CreateGroupBody{
				GroupBody: spaceStructs.GroupBody{
					Name:      "市场部",
					Slug:      fmt.Sprintf("%s-marketing-department", companySlug),
					ParentID:  &companyID,
					TenantID:  &tenant.ID,
					CreatedBy: &admin.ID,
					UpdatedBy: &admin.ID,
				},
			},
			teams: []spaceStructs.CreateGroupBody{
				{GroupBody: spaceStructs.GroupBody{Name: "品牌组", Slug: fmt.Sprintf("%s-brand-team", companySlug)}},
				{GroupBody: spaceStructs.GroupBody{Name: "市场调研组", Slug: fmt.Sprintf("%s-marketing-research-team", companySlug)}},
			},
		},
		{
			department: spaceStructs.CreateGroupBody{
				GroupBody: spaceStructs.GroupBody{
					Name:      "人力资源部",
					Slug:      fmt.Sprintf("%s-hr-department", companySlug),
					ParentID:  &companyID,
					TenantID:  &tenant.ID,
					CreatedBy: &admin.ID,
					UpdatedBy: &admin.ID,
				},
			},
			teams: []spaceStructs.CreateGroupBody{
				{GroupBody: spaceStructs.GroupBody{Name: "招聘组", Slug: fmt.Sprintf("%s-recruitment-team", companySlug)}},
				{GroupBody: spaceStructs.GroupBody{Name: "培训发展组", Slug: fmt.Sprintf("%s-train-development-team", companySlug)}},
			},
		},
	}

	departments[companySlug] = append(departments[companySlug], commonDepartments...)

	for _, dept := range departments[companySlug] {
		createdDepartment, err := s.ss.Group.Create(ctx, &dept.department)
		if err != nil {
			return fmt.Errorf("failed to create department %s: %v", dept.department.Name, err)
		}

		for _, team := range dept.teams {
			team.ParentID = &createdDepartment.ID
			team.TenantID = &tenant.ID
			team.CreatedBy = &admin.ID
			team.UpdatedBy = &admin.ID
			if _, err = s.ss.Group.Create(ctx, &team); err != nil {
				return fmt.Errorf("failed to create team %s: %v", team.Name, err)
			}
		}
	}

	return nil
}

func (s *Service) createCompanies(ctx context.Context, parentID string) ([]*spaceStructs.ReadGroup, error) {
	tenant, err := s.ts.Tenant.GetBySlug(ctx, "ncobase")
	if err != nil {
		return nil, fmt.Errorf("failed to get default tenant: %v", err)
	}

	admin, err := s.us.User.Get(ctx, "admin")
	if err != nil {
		return nil, fmt.Errorf("failed to get admin user: %v", err)
	}

	companies := []spaceStructs.CreateGroupBody{
		{
			GroupBody: spaceStructs.GroupBody{
				Name:      "信息科技公司",
				Slug:      "tech-company",
				ParentID:  &parentID,
				TenantID:  &tenant.ID,
				CreatedBy: &admin.ID,
				UpdatedBy: &admin.ID,
			},
		},
		{
			GroupBody: spaceStructs.GroupBody{
				Name:      "文化传媒公司",
				Slug:      "media-company",
				ParentID:  &parentID,
				TenantID:  &tenant.ID,
				CreatedBy: &admin.ID,
				UpdatedBy: &admin.ID,
			},
		},
	}

	var createdCompanies []*spaceStructs.ReadGroup
	for _, company := range companies {
		createdCompany, err := s.ss.Group.Create(ctx, &company)
		if err != nil {
			return nil, fmt.Errorf("failed to create company %s: %v", company.Name, err)
		}
		createdCompanies = append(createdCompanies, createdCompany)
	}

	return createdCompanies, nil
}

func (s *Service) createTemporaryOrganizations(ctx context.Context, groupID string) error {
	tenant, err := s.ts.Tenant.GetBySlug(ctx, "ncobase")
	if err != nil {
		return fmt.Errorf("failed to get default tenant: %v", err)
	}

	admin, err := s.us.User.Get(ctx, "admin")
	if err != nil {
		return fmt.Errorf("failed to get admin user: %v", err)
	}

	tempOrgs := []spaceStructs.CreateGroupBody{
		{
			GroupBody: spaceStructs.GroupBody{
				Name:      "战略委员会",
				Slug:      "strategy-committee",
				ParentID:  &groupID,
				TenantID:  &tenant.ID,
				CreatedBy: &admin.ID,
				UpdatedBy: &admin.ID,
			},
		},
		{
			GroupBody: spaceStructs.GroupBody{
				Name:      "数字化转型项目组",
				Slug:      "digital-transformation-team",
				ParentID:  &groupID,
				TenantID:  &tenant.ID,
				CreatedBy: &admin.ID,
				UpdatedBy: &admin.ID,
			},
		},
	}

	for _, org := range tempOrgs {
		if _, err := s.ss.Group.Create(ctx, &org); err != nil {
			return fmt.Errorf("failed to create temporary organization %s: %v", org.Name, err)
		}
	}

	return nil
}

func (s *Service) createOrganizationRolesAndPermissions(ctx context.Context) error {
	roles := []accessStructs.CreateRoleBody{
		{RoleBody: accessStructs.RoleBody{Name: "Group Admin", Slug: "group-admin"}},
		{RoleBody: accessStructs.RoleBody{Name: "Department Manager", Slug: "department-manager"}},
		{RoleBody: accessStructs.RoleBody{Name: "Team Leader", Slug: "team-leader"}},
		{RoleBody: accessStructs.RoleBody{Name: "Employee", Slug: "employee"}},
	}

	for _, role := range roles {
		if _, err := s.acs.Role.Create(ctx, &role); err != nil {
			return fmt.Errorf("failed to create role %s: %v", role.Name, err)
		}
	}

	permissions := []accessStructs.CreatePermissionBody{
		{PermissionBody: accessStructs.PermissionBody{Name: "Manage Group", Action: "*", Subject: "group"}},
		{PermissionBody: accessStructs.PermissionBody{Name: "Manage Department", Action: "*", Subject: "department"}},
		{PermissionBody: accessStructs.PermissionBody{Name: "Manage Team", Action: "*", Subject: "team"}},
		{PermissionBody: accessStructs.PermissionBody{Name: "View Group", Action: "GET", Subject: "group"}},
		{PermissionBody: accessStructs.PermissionBody{Name: "View Department", Action: "GET", Subject: "department"}},
		{PermissionBody: accessStructs.PermissionBody{Name: "View Team", Action: "GET", Subject: "team"}},
	}

	for _, permission := range permissions {
		if _, err := s.acs.Permission.Create(ctx, &permission); err != nil {
			return fmt.Errorf("failed to create permission %s: %v", permission.Name, err)
		}
	}

	// Associate permissions with roles
	rolePermissions := map[string][]string{
		"group-admin":        {"Manage Group", "Manage Department", "Manage Team"},
		"department-manager": {"Manage Department", "Manage Team", "View Group"},
		"team-leader":        {"Manage Team", "View Department", "View Group"},
		"employee":           {"View Team", "View Department", "View Group"},
	}

	for roleSlug, permissionNames := range rolePermissions {
		role, err := s.acs.Role.GetBySlug(ctx, roleSlug)
		if err != nil {
			return fmt.Errorf("failed to get role %s: %v", roleSlug, err)
		}

		for _, permissionName := range permissionNames {
			permission, err := s.acs.Permission.GetByName(ctx, permissionName)
			if err != nil {
				return fmt.Errorf("failed to get permission %s: %v", permissionName, err)
			}

			if _, err := s.acs.RolePermission.AddPermissionToRole(ctx, role.ID, permission.ID); err != nil {
				return fmt.Errorf("failed to add permission %s to role %s: %v", permissionName, roleSlug, err)
			}
		}
	}

	return nil
}

func (s *Service) associateUsersWithGroupsAndRoles(ctx context.Context) error {
	users, err := s.us.User.List(ctx, &userStructs.ListUserParams{})
	if err != nil {
		return fmt.Errorf("failed to list users: %v", err)
	}

	yeduGroup, err := s.ss.Group.Get(ctx, &spaceStructs.FindGroup{Group: "yedu-group"})
	if err != nil {
		return fmt.Errorf("failed to get Yedu Group: %v", err)
	}

	for _, user := range users.Items {
		// Associate user with Yedu Group
		if _, err := s.ss.UserGroup.AddUserToGroup(ctx, user.ID, yeduGroup.ID); err != nil {
			return fmt.Errorf("failed to add user %s to Yedu Group: %v", user.Username, err)
		}

		// Assign role based on user type
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
			return fmt.Errorf("failed to add role %s to user %s: %v", roleSlug, user.Username, err)
		}
	}

	return nil
}
