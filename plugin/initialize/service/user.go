package service

import (
	"context"
	"fmt"
	"ncobase/initialize/data"
	spaceStructs "ncobase/space/structs"
	userStructs "ncobase/user/structs"
	"strings"
	"time"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/logging/logger"
)

// checkUsersInitialized checks if users are already initialized
func (s *Service) checkUsersInitialized(ctx context.Context) error {
	params := &userStructs.ListUserParams{}
	count := s.us.User.CountX(ctx, params)
	if count > 0 {
		logger.Infof(ctx, "Users already exist, skipping initialization")
		keyUsers := []string{"super", "admin"}
		for _, username := range keyUsers {
			if err := s.VerifyUserRoleAssignment(ctx, username); err != nil {
				logger.Warnf(ctx, "Verification failed for user %s: %v", username, err)
			}
		}
		return nil
	}

	err := s.initUsers(ctx)
	if err != nil {
		return err
	}

	keyUsers := []string{"super", "admin"}
	for _, username := range keyUsers {
		if err := s.VerifyUserRoleAssignment(ctx, username); err != nil {
			logger.Warnf(ctx, "Post-initialization verification failed for user %s: %v", username, err)
		}
	}

	return nil
}

// validatePassword validates that a password meets the password policy
func (s *Service) validatePassword(ctx context.Context, username, password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	policy := s.c.Security.DefaultPasswordPolicy
	if len(password) < policy.MinLength {
		return fmt.Errorf("password must be at least %d characters long", policy.MinLength)
	}

	if policy.RequireUppercase && !strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if policy.RequireLowercase && !strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if policy.RequireDigits && !strings.ContainsAny(password, "0123456789") {
		return fmt.Errorf("password must contain at least one digit")
	}

	if policy.RequireSpecial && !strings.ContainsAny(password, "!@#$%^&*()_+-=[]{}|;:,.<>?") {
		return fmt.Errorf("password must contain at least one special character")
	}

	// Check if password contains the username
	if strings.Contains(strings.ToLower(password), strings.ToLower(username)) {
		return fmt.Errorf("password cannot contain the username")
	}

	return nil
}

// initUsers initializes enterprise users, employees, roles, and relationships
func (s *Service) initUsers(ctx context.Context) error {
	logger.Infof(ctx, "Initializing enterprise users and employees...")

	// Get default tenant
	defaultTenant, err := s.ts.Tenant.GetBySlug(ctx, "digital-enterprise")
	if err != nil {
		logger.Errorf(ctx, "Error getting default tenant: %v", err)
		return fmt.Errorf("default tenant 'digital-enterprise' not found: %w", err)
	}

	var createdCount int
	for _, userInfo := range data.SystemDefaultUsers {
		// Validate password
		if err := s.validatePassword(ctx, userInfo.User.Username, userInfo.Password); err != nil {
			logger.Warnf(ctx, "Password for user %s does not meet policy: %v", userInfo.User.Username, err)
		}

		// Check if user already exists
		existingUser, _ := s.us.User.Get(ctx, userInfo.User.Username)
		if existingUser != nil {
			logger.Infof(ctx, "User %s already exists, skipping", userInfo.User.Username)
			continue
		}

		// Create user
		createdUser, err := s.us.User.CreateUser(ctx, &userInfo.User)
		if err != nil {
			logger.Errorf(ctx, "Error creating user %s: %v", userInfo.User.Username, err)
			return fmt.Errorf("failed to create user '%s': %w", userInfo.User.Username, err)
		}
		logger.Debugf(ctx, "Created user: %s", userInfo.User.Username)
		createdCount++

		// Set password
		if err = s.us.User.UpdatePassword(ctx, &userStructs.UserPassword{
			User:        createdUser.Username,
			NewPassword: userInfo.Password,
			Confirm:     userInfo.Password,
		}); err != nil {
			logger.Errorf(ctx, "Error setting password for user %s: %v", createdUser.Username, err)
			return fmt.Errorf("failed to set password for user '%s': %w", createdUser.Username, err)
		}

		// Create user profile
		profileData := userInfo.Profile
		profileData.UserID = createdUser.ID
		if _, err := s.us.UserProfile.Create(ctx, &profileData); err != nil {
			logger.Errorf(ctx, "Error creating profile for user %s: %v", createdUser.Username, err)
			return fmt.Errorf("failed to create profile for user '%s': %w", createdUser.Username, err)
		}

		// Create employee record if provided
		if userInfo.Employee != nil {
			employeeData := *userInfo.Employee
			employeeData.UserID = createdUser.ID
			employeeData.TenantID = defaultTenant.ID

			// Resolve manager username to ID if specified
			if employeeData.ManagerID != "" {
				manager, err := s.us.User.Get(ctx, employeeData.ManagerID)
				if err == nil {
					employeeData.ManagerID = manager.ID
				} else {
					logger.Warnf(ctx, "Manager %s not found for employee %s", employeeData.ManagerID, createdUser.Username)
					employeeData.ManagerID = ""
				}
			}

			if _, err := s.us.Employee.Create(ctx, &userStructs.CreateEmployeeBody{
				EmployeeBody: employeeData,
			}); err != nil {
				logger.Errorf(ctx, "Error creating employee record for user %s: %v", createdUser.Username, err)
				return fmt.Errorf("failed to create employee record for user '%s': %w", createdUser.Username, err)
			}
			logger.Debugf(ctx, "Created employee record for user: %s", createdUser.Username)
		}

		// Assing user role
		if err := s.assignUserRole(ctx, createdUser, userInfo.Role, defaultTenant.ID); err != nil {
			logger.Errorf(ctx, "Error assigning role to user %s: %v", createdUser.Username, err)
			return fmt.Errorf("failed to assign role to user '%s': %w", createdUser.Username, err)
		}
	}

	// Initialize cross-company assignments after all users are created
	if err := s.initializeUserGroupAssignments(ctx); err != nil {
		logger.Errorf(ctx, "Failed to initialize user-group assignments: %v", err)
		return fmt.Errorf("user-group assignment failed: %w", err)
	}

	// Verify required users were created
	reqUsers := []string{"super", "admin", "chief.executive", "hr.manager", "finance.manager", "tech.lead"}
	for _, username := range reqUsers {
		user, err := s.us.User.Get(ctx, username)
		if err != nil || user == nil {
			logger.Errorf(ctx, "Required user '%s' was not created", username)
			return fmt.Errorf("required user '%s' was not created: %w", username, err)
		}
	}

	count := s.us.User.CountX(ctx, &userStructs.ListUserParams{})
	logger.Infof(ctx, "User and employee initialization completed, created %d users", count)

	return nil
}

// assignUserRole assigns both global and tenant roles to user
func (s *Service) assignUserRole(ctx context.Context, user *userStructs.ReadUser, roleSlug string, tenantID string) error {
	// Get role by slug
	role, err := s.acs.Role.GetBySlug(ctx, roleSlug)
	if err != nil {
		logger.Errorf(ctx, "Role '%s' not found: %v", roleSlug, err)
		return fmt.Errorf("role '%s' not found: %w", roleSlug, err)
	}

	// Add user to tenant first
	if _, err := s.ts.UserTenant.AddUserToTenant(ctx, user.ID, tenantID); err != nil {
		logger.Errorf(ctx, "Failed to add user to tenant: %v", err)
		return fmt.Errorf("failed to add user '%s' to tenant: %w", user.Username, err)
	}

	// Assign global role (primary role assignment)
	if err := s.acs.UserRole.AddRoleToUser(ctx, user.ID, role.ID); err != nil {
		logger.Errorf(ctx, "Failed to assign global role: %v", err)
		return fmt.Errorf("failed to assign global role '%s' to user '%s': %w", roleSlug, user.Username, err)
	}

	// Assign tenant-specific role (for tenant-level permissions)
	if _, err := s.acs.UserTenantRole.AddRoleToUserInTenant(ctx, user.ID, tenantID, role.ID); err != nil {
		// Log warning but don't fail - global role assignment is sufficient
		logger.Warnf(ctx, "Failed to assign tenant role to user '%s': %v", user.Username, err)
	}

	logger.Infof(ctx, "Successfully assigned role '%s' to user '%s'", roleSlug, user.Username)
	return nil
}

// initializeUserGroupAssignments assigns users to groups based on enterprise structure
func (s *Service) initializeUserGroupAssignments(ctx context.Context) error {
	logger.Infof(ctx, "Initializing user-group assignments...")

	// Cross-company assignments for enterprise roles
	assignments := map[string][]string{
		"chief.executive": {
			"digital-enterprise",
			"executive-office",
			"techcorp",
			"mediacorp",
			"consultcorp",
		},
		"hr.manager": {
			"digital-enterprise",
			"corporate-hr",
			"techcorp-hr",
			"mediacorp-hr",
			"consultcorp-hr",
		},
		"finance.manager": {
			"digital-enterprise",
			"corporate-finance",
			"techcorp-finance",
			"mediacorp-finance",
			"consultcorp-finance",
		},
		"tech.lead": {
			"techcorp",
			"technology",
			"backend-dev",
			"frontend-dev",
		},
		"senior.developer": {
			"techcorp",
			"technology",
			"backend-dev",
		},
		"marketing.manager": {
			"mediacorp",
			"digital-marketing",
			"social-media",
		},
		"content.creator": {
			"mediacorp",
			"content-production",
			"video-production",
		},
	}

	var assignmentCount int
	for username, groupSlugs := range assignments {
		user, err := s.us.User.Get(ctx, username)
		if err != nil {
			logger.Warnf(ctx, "User '%s' not found, skipping group assignments", username)
			continue
		}

		for _, groupSlug := range groupSlugs {
			group, err := s.findGroupBySlug(ctx, groupSlug)
			if err != nil {
				logger.Warnf(ctx, "Group '%s' not found for user '%s'", groupSlug, username)
				continue
			}

			// Assign user to group
			if _, err := s.ss.UserGroup.AddUserToGroup(ctx, user.ID, group.ID); err != nil {
				logger.Warnf(ctx, "Failed to assign user '%s' to group '%s': %v", username, groupSlug, err)
				continue
			}

			logger.Debugf(ctx, "Assigned user '%s' to group '%s'", username, groupSlug)
			assignmentCount++
		}

		// Assign organizational roles based on position
		if err := s.assignOrganizationalRoles(ctx, user, username); err != nil {
			logger.Warnf(ctx, "Failed to assign organizational roles to user '%s': %v", username, err)
		}
	}

	logger.Infof(ctx, "User-group assignment completed, created %d assignments", assignmentCount)
	return nil
}

// findGroupBySlug finds group by slug
func (s *Service) findGroupBySlug(ctx context.Context, slug string) (*spaceStructs.ReadGroup, error) {
	groups, err := s.ss.Group.List(ctx, &spaceStructs.ListGroupParams{})
	if err != nil {
		return nil, err
	}

	for _, group := range groups.Items {
		if group.Slug == slug {
			return group, nil
		}
	}

	return nil, fmt.Errorf("group with slug '%s' not found", slug)
}

// assignOrganizationalRoles assigns roles based on user position
func (s *Service) assignOrganizationalRoles(ctx context.Context, user *userStructs.ReadUser, username string) error {
	// Get user info from initialization data
	var userInfo *data.UserCreationInfo
	for _, info := range data.SystemDefaultUsers {
		if info.User.Username == username {
			userInfo = &info
			break
		}
	}

	if userInfo == nil || userInfo.Employee == nil {
		return nil
	}

	// Determine organizational role based on position
	var orgRoleSlug string
	switch userInfo.Employee.Position {
	case "Chief Executive Officer":
		orgRoleSlug = "enterprise-executive"
	case "Technical Lead", "Marketing Manager", "HR Manager", "Finance Manager":
		orgRoleSlug = "department-head"
	default:
		// Regular employees don't get additional organizational roles
		return nil
	}

	// Get and assign organizational role
	orgRole, err := s.acs.Role.GetBySlug(ctx, orgRoleSlug)
	if err != nil {
		return fmt.Errorf("organizational role '%s' not found: %w", orgRoleSlug, err)
	}

	if err := s.acs.UserRole.AddRoleToUser(ctx, user.ID, orgRole.ID); err != nil {
		return fmt.Errorf("failed to assign organizational role '%s': %w", orgRoleSlug, err)
	}

	logger.Debugf(ctx, "Assigned organizational role '%s' to user '%s'", orgRoleSlug, user.Username)
	return nil
}

// InitializeUsers initializes only the users if the system is already initialized
func (s *Service) InitializeUsers(ctx context.Context) (*InitState, error) {
	logger.Infof(ctx, "Starting user initialization...")

	// Check if the system is initialized
	if !s.IsInitialized(ctx) {
		logger.Infof(ctx, "System is not yet fully initialized")
		// For users, we need to ensure roles and permissions are initialized
		// before attempting to initialize users
		rolesStatus := InitStatus{
			Component: "roles",
			Status:    "initialized",
		}

		if err := s.checkRolesInitialized(ctx); err != nil {
			rolesStatus.Status = "failed"
			rolesStatus.Error = err.Error()
			s.state.Statuses = append(s.state.Statuses, rolesStatus)
			logger.Errorf(ctx, "Failed to initialize roles: %v", err)
			return s.state, fmt.Errorf("initialization step roles failed: %v", err)
		}

		permissionsStatus := InitStatus{
			Component: "permissions",
			Status:    "initialized",
		}

		if err := s.checkPermissionsInitialized(ctx); err != nil {
			permissionsStatus.Status = "failed"
			permissionsStatus.Error = err.Error()
			s.state.Statuses = append(s.state.Statuses, permissionsStatus)
			logger.Errorf(ctx, "Failed to initialize permissions: %v", err)
			return s.state, fmt.Errorf("initialization step permissions failed: %v", err)
		}

		tenantsStatus := InitStatus{
			Component: "tenants",
			Status:    "initialized",
		}

		if err := s.checkTenantsInitialized(ctx); err != nil {
			tenantsStatus.Status = "failed"
			tenantsStatus.Error = err.Error()
			s.state.Statuses = append(s.state.Statuses, tenantsStatus)
			logger.Errorf(ctx, "Failed to initialize tenants: %v", err)
			return s.state, fmt.Errorf("initialization step tenants failed: %v", err)
		}

		s.state.Statuses = append(s.state.Statuses, rolesStatus, permissionsStatus, tenantsStatus)
	}

	// Initialize just users
	status := InitStatus{
		Component: "users",
		Status:    "initialized",
	}

	logger.Infof(ctx, "Initializing users...")
	if err := s.checkUsersInitialized(ctx); err != nil {
		status.Status = "failed"
		status.Error = err.Error()
		s.state.Statuses = append(s.state.Statuses, status)
		logger.Errorf(ctx, "Failed to initialize users: %v", err)
		return s.state, fmt.Errorf("initialization step users failed: %v", err)
	}

	s.state.Statuses = append(s.state.Statuses, status)
	logger.Infof(ctx, "Successfully initialized users")

	// Also initialize Casbin policies for the users
	policiesStatus := InitStatus{
		Component: "casbin_policies",
		Status:    "initialized",
	}

	if err := s.checkCasbinPoliciesInitialized(ctx); err != nil {
		policiesStatus.Status = "failed"
		policiesStatus.Error = err.Error()
		s.state.Statuses = append(s.state.Statuses, policiesStatus)
		logger.Errorf(ctx, "Failed to initialize Casbin policies: %v", err)
		return s.state, fmt.Errorf("initialization step casbin_policies failed: %v", err)
	}

	s.state.Statuses = append(s.state.Statuses, policiesStatus)
	logger.Infof(ctx, "Successfully initialized Casbin policies")

	s.state.LastRunTime = time.Now().UnixMilli()

	// Persist state if configured
	if s.c.Initialization.PersistState {
		if err := s.SaveState(ctx); err != nil {
			logger.Warnf(ctx, "Failed to save initialization state: %v", err)
		}
	}

	logger.Infof(ctx, "User initialization completed successfully")
	return s.state, nil
}

// VerifyUserRoleAssignment verifies user role assignments with detailed logging
func (s *Service) VerifyUserRoleAssignment(ctx context.Context, username string) error {
	logger.Infof(ctx, "Verifying role assignment for user: %s", username)

	user, err := s.us.User.Get(ctx, username)
	if err != nil {
		return fmt.Errorf("user '%s' not found: %w", username, err)
	}

	// Check global roles
	globalRoles, err := s.acs.UserRole.GetUserRoles(ctx, user.ID)
	if err != nil {
		logger.Warnf(ctx, "Failed to get global roles for user '%s': %v", username, err)
	} else {
		logger.Infof(ctx, "User '%s' has %d global roles:", username, len(globalRoles))
		for _, role := range globalRoles {
			logger.Infof(ctx, "  - Global role: %s (%s)", role.Name, role.Slug)
		}
	}

	// Check tenant assignments
	userTenants, err := s.ts.UserTenant.UserBelongTenants(ctx, user.ID)
	if err != nil {
		logger.Warnf(ctx, "Failed to get tenants for user '%s': %v", username, err)
		return nil
	}

	logger.Infof(ctx, "User '%s' belongs to %d tenants:", username, len(userTenants))
	for _, tenant := range userTenants {
		logger.Infof(ctx, "  - Tenant: %s (%s)", tenant.Name, tenant.Slug)

		// Get tenant-specific roles
		ctx = ctxutil.SetTenantID(ctx, tenant.ID)
		roleIDs, roleErr := s.acs.UserTenantRole.GetUserRolesInTenant(ctx, user.ID, tenant.ID)
		if roleErr == nil && len(roleIDs) > 0 {
			tenantRoles, _ := s.acs.Role.GetByIDs(ctx, roleIDs)
			logger.Infof(ctx, "    User has %d roles in tenant '%s':", len(tenantRoles), tenant.Name)
			for _, role := range tenantRoles {
				logger.Infof(ctx, "      - Tenant role: %s (%s)", role.Name, role.Slug)
			}
		}
	}

	return nil
}
