package service

import (
	"context"
	"fmt"
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
		logger.Infof(ctx, "Users already exist, verifying setup")

		// Verify key users and their roles
		keyUsers := []string{"super", "admin"}
		for _, username := range keyUsers {
			if err := s.VerifyUserRoleAssignment(ctx, username); err != nil {
				logger.Warnf(ctx, "Verification failed for user %s: %v", username, err)
			}
		}

		// Verify employee relationships and department assignments
		if err := s.verifyEmployeeSetup(ctx); err != nil {
			logger.Warnf(ctx, "Employee setup verification issues: %v", err)
		}

		return nil
	}

	return s.initUsers(ctx)
}

// initUsers initializes users and employees using current data mode
func (s *Service) initUsers(ctx context.Context) error {
	logger.Infof(ctx, "Initializing users and employees in %s mode...", s.state.DataMode)

	defaultTenant, err := s.ts.Tenant.GetBySlug(ctx, "digital-enterprise")
	if err != nil {
		logger.Errorf(ctx, "Error getting default tenant: %v", err)
		return fmt.Errorf("default tenant 'digital-enterprise' not found: %w", err)
	}

	dataLoader := s.getDataLoader()
	users := dataLoader.GetUsers()

	var createdCount int
	var employeeCount int

	// Phase 1: Create all users first
	for _, userInfo := range users {
		if err := s.validatePassword(ctx, userInfo.User.Username, userInfo.Password); err != nil {
			logger.Warnf(ctx, "Password for user %s does not meet policy: %v", userInfo.User.Username, err)
		}

		existingUser, _ := s.us.User.Get(ctx, userInfo.User.Username)
		if existingUser != nil {
			logger.Infof(ctx, "User %s already exists, skipping", userInfo.User.Username)
			continue
		}

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

		// Create employee record (without manager reference for now)
		if userInfo.Employee != nil {
			if err := s.createEmployeeRecord(ctx, createdUser, userInfo.Employee, defaultTenant.ID, false); err != nil {
				logger.Errorf(ctx, "Error creating employee record for user %s: %v", createdUser.Username, err)
				return fmt.Errorf("failed to create employee record for user '%s': %w", createdUser.Username, err)
			}
			employeeCount++
		}

		// Assign user role
		if err := s.assignUserRole(ctx, createdUser, userInfo.Role, defaultTenant.ID); err != nil {
			logger.Errorf(ctx, "Error assigning role to user %s: %v", createdUser.Username, err)
			return fmt.Errorf("failed to assign role to user '%s': %w", createdUser.Username, err)
		}
	}

	// Phase 2: Resolve manager references now that all users exist
	if err := s.resolveManagerReferences(ctx); err != nil {
		logger.Warnf(ctx, "Failed to resolve some manager references: %v", err)
		// Don't fail initialization for this
	}

	// Phase 3: Initialize user-group assignments and department assignments
	if err := s.initializeUserGroupAssignments(ctx); err != nil {
		logger.Errorf(ctx, "Failed to initialize user-group assignments: %v", err)
		return fmt.Errorf("user-group assignment failed: %w", err)
	}

	// Phase 4: Assign employees to department groups
	if err := s.assignEmployeesToDepartments(ctx); err != nil {
		logger.Warnf(ctx, "Failed to assign some employees to departments: %v", err)
		// Don't fail initialization for this
	}

	// Validate required users were created
	reqUsers := []string{"super", "admin"}
	for _, username := range reqUsers {
		user, err := s.us.User.Get(ctx, username)
		if err != nil || user == nil {
			logger.Errorf(ctx, "Required user '%s' was not created", username)
			return fmt.Errorf("required user '%s' was not created: %w", username, err)
		}
	}

	userCount := s.us.User.CountX(ctx, &userStructs.ListUserParams{})
	logger.Infof(ctx, "User and employee initialization completed in %s mode, created %d users and %d employee records",
		s.state.DataMode, userCount, employeeCount)

	return nil
}

// createEmployeeRecord creates employee record with optional manager resolution
func (s *Service) createEmployeeRecord(ctx context.Context, user *userStructs.ReadUser, employeeData *userStructs.EmployeeBody, tenantID string, resolveManager bool) error {
	employeeBody := *employeeData
	employeeBody.UserID = user.ID
	employeeBody.TenantID = tenantID

	// Resolve manager username to ID if requested and manager is specified
	if resolveManager && employeeBody.ManagerID != "" {
		manager, err := s.us.User.Get(ctx, employeeBody.ManagerID)
		if err == nil && manager != nil {
			employeeBody.ManagerID = manager.ID
			logger.Debugf(ctx, "Resolved manager %s for employee %s", manager.Username, user.Username)
		} else {
			logger.Warnf(ctx, "Manager %s not found for employee %s, will be resolved later", employeeBody.ManagerID, user.Username)
			employeeBody.ManagerID = "" // Clear invalid reference for now
		}
	} else if !resolveManager && employeeBody.ManagerID != "" {
		// Store the username temporarily, will be resolved later
		logger.Debugf(ctx, "Manager reference %s for employee %s will be resolved later", employeeBody.ManagerID, user.Username)
		employeeBody.ManagerID = "" // Clear for now
	}

	// Create employee record
	if _, err := s.us.Employee.Create(ctx, &userStructs.CreateEmployeeBody{
		EmployeeBody: employeeBody,
	}); err != nil {
		return fmt.Errorf("failed to create employee record: %w", err)
	}

	logger.Debugf(ctx, "Created employee record for user: %s", user.Username)
	return nil
}

// resolveManagerReferences resolves manager references after all users are created
func (s *Service) resolveManagerReferences(ctx context.Context) error {
	logger.Infof(ctx, "Resolving manager references...")

	dataLoader := s.getDataLoader()
	users := dataLoader.GetUsers()

	var updatedCount int
	for _, userInfo := range users {
		if userInfo.Employee == nil || userInfo.Employee.ManagerID == "" {
			continue
		}

		user, err := s.us.User.Get(ctx, userInfo.User.Username)
		if err != nil {
			logger.Warnf(ctx, "User %s not found during manager resolution", userInfo.User.Username)
			continue
		}

		manager, err := s.us.User.Get(ctx, userInfo.Employee.ManagerID)
		if err != nil {
			logger.Warnf(ctx, "Manager %s not found for user %s", userInfo.Employee.ManagerID, user.Username)
			continue
		}

		// Get employee record
		employees, err := s.us.Employee.GetByManager(ctx, user.ID)
		if err != nil || len(employees) == 0 {
			logger.Warnf(ctx, "Employee record not found for user %s", user.Username)
			continue
		}

		employee := employees[0]
		// Update manager reference
		updateBody := &userStructs.UpdateEmployeeBody{
			EmployeeBody: userStructs.EmployeeBody{
				UserID:         employee.UserID,
				TenantID:       employee.TenantID,
				EmployeeID:     employee.EmployeeID,
				Department:     employee.Department,
				Position:       employee.Position,
				ManagerID:      manager.ID,
				EmploymentType: employee.EmploymentType,
				Status:         employee.Status,
				HireDate:       employee.HireDate,
				Skills:         employee.Skills,
			},
		}

		if _, err := s.us.Employee.Update(ctx, employee.UserID, updateBody); err != nil {
			logger.Warnf(ctx, "Failed to update manager for employee %s: %v", user.Username, err)
			continue
		}

		logger.Debugf(ctx, "Updated manager for employee %s -> %s", user.Username, manager.Username)
		updatedCount++
	}

	logger.Infof(ctx, "Manager reference resolution completed, updated %d employees", updatedCount)
	return nil
}

// assignEmployeesToDepartments assigns employees to their department groups
func (s *Service) assignEmployeesToDepartments(ctx context.Context) error {
	logger.Infof(ctx, "Assigning employees to department groups...")

	employees, err := s.us.Employee.List(ctx, &userStructs.ListEmployeeParams{})
	if err != nil {
		return fmt.Errorf("failed to list employees: %w", err)
	}

	var assignedCount int
	for _, employee := range employees.Items {
		if employee.Department == "" {
			continue // Skip employees without department
		}

		// Get user information
		user, err := s.us.User.GetByID(ctx, employee.UserID)
		if err != nil {
			logger.Warnf(ctx, "User not found for employee %s", employee.EmployeeID)
			continue
		}

		// Find department group
		group, err := s.findGroupBySlug(ctx, employee.Department)
		if err != nil {
			logger.Warnf(ctx, "Department group '%s' not found for employee %s", employee.Department, user.Username)
			continue
		}

		// Check if user is already in the group
		isMember, err := s.ss.UserGroup.IsUserMember(ctx, user.ID, group.ID)
		if err != nil {
			logger.Warnf(ctx, "Failed to check group membership for user %s: %v", user.Username, err)
			continue
		}

		if !isMember {
			// Add user to department group
			if _, err := s.ss.UserGroup.AddUserToGroup(ctx, user.ID, group.ID, spaceStructs.RoleMember); err != nil {
				logger.Warnf(ctx, "Failed to add user %s to department %s: %v", user.Username, employee.Department, err)
				continue
			}

			logger.Debugf(ctx, "Added employee %s to department %s", user.Username, employee.Department)
			assignedCount++
		}
	}

	logger.Infof(ctx, "Employee department assignment completed, added %d assignments", assignedCount)
	return nil
}

// verifyEmployeeSetup verifies existing employee setup
func (s *Service) verifyEmployeeSetup(ctx context.Context) error {
	logger.Infof(ctx, "Verifying employee setup...")

	// Check if we need to resolve any missing manager references
	if err := s.resolveManagerReferences(ctx); err != nil {
		logger.Warnf(ctx, "Manager reference resolution issues: %v", err)
	}

	// Check if we need to assign employees to departments
	if err := s.assignEmployeesToDepartments(ctx); err != nil {
		logger.Warnf(ctx, "Employee department assignment issues: %v", err)
	}

	logger.Infof(ctx, "Employee setup verification completed")
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

	if strings.Contains(strings.ToLower(password), strings.ToLower(username)) {
		return fmt.Errorf("password cannot contain the username")
	}

	return nil
}

// assignUserRole assigns both global and tenant roles to user
func (s *Service) assignUserRole(ctx context.Context, user *userStructs.ReadUser, roleSlug string, tenantID string) error {
	role, err := s.acs.Role.GetBySlug(ctx, roleSlug)
	if err != nil {
		logger.Errorf(ctx, "Role '%s' not found: %v", roleSlug, err)
		return fmt.Errorf("role '%s' not found: %w", roleSlug, err)
	}

	if _, err := s.ts.UserTenant.AddUserToTenant(ctx, user.ID, tenantID); err != nil {
		logger.Errorf(ctx, "Failed to add user to tenant: %v", err)
		return fmt.Errorf("failed to add user '%s' to tenant: %w", user.Username, err)
	}

	if err := s.acs.UserRole.AddRoleToUser(ctx, user.ID, role.ID); err != nil {
		logger.Errorf(ctx, "Failed to assign global role: %v", err)
		return fmt.Errorf("failed to assign global role '%s' to user '%s': %w", roleSlug, user.Username, err)
	}

	if _, err := s.acs.UserTenantRole.AddRoleToUserInTenant(ctx, user.ID, tenantID, role.ID); err != nil {
		logger.Warnf(ctx, "Failed to assign tenant role to user '%s': %v", user.Username, err)
	}

	logger.Infof(ctx, "Successfully assigned role '%s' to user '%s'", roleSlug, user.Username)
	return nil
}

// initializeUserGroupAssignments assigns users to groups
func (s *Service) initializeUserGroupAssignments(ctx context.Context) error {
	logger.Infof(ctx, "Initializing user-group assignments...")

	// Basic assignments for company mode
	assignments := map[string][]string{
		"admin": {
			"digital-enterprise",
		},
		"manager": {
			"digital-enterprise",
		},
		"employee": {
			"digital-enterprise",
		},
	}

	// Extended assignments for enterprise mode
	if s.state.DataMode == "enterprise" {
		assignments = map[string][]string{
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
			},
			"finance.manager": {
				"digital-enterprise",
				"corporate-finance",
			},
			"tech.lead": {
				"techcorp",
				"technology",
			},
			"marketing.manager": {
				"mediacorp",
				"digital-marketing",
			},
		}
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

			if _, err := s.ss.UserGroup.AddUserToGroup(ctx, user.ID, group.ID, spaceStructs.RoleMember); err != nil {
				logger.Warnf(ctx, "Failed to assign user '%s' to group '%s': %v", username, groupSlug, err)
				continue
			}

			logger.Debugf(ctx, "Assigned user '%s' to group '%s'", username, groupSlug)
			assignmentCount++
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

// InitializeUsers initializes only the users if the system is already initialized
func (s *Service) InitializeUsers(ctx context.Context) (*InitState, error) {
	logger.Infof(ctx, "Starting user initialization in %s mode...", s.state.DataMode)

	if !s.IsInitialized(ctx) {
		logger.Infof(ctx, "System is not yet fully initialized")

		steps := []struct {
			name string
			fn   func(context.Context) error
		}{
			{"roles", s.checkRolesInitialized},
			{"permissions", s.checkPermissionsInitialized},
			{"tenants", s.checkTenantsInitialized},
		}

		for _, step := range steps {
			status := InitStatus{
				Component: step.name,
				Status:    "initialized",
			}

			if err := step.fn(ctx); err != nil {
				status.Status = "failed"
				status.Error = err.Error()
				s.state.Statuses = append(s.state.Statuses, status)
				logger.Errorf(ctx, "Failed to initialize %s: %v", step.name, err)
				return s.state, fmt.Errorf("initialization step %s failed: %v", step.name, err)
			}

			s.state.Statuses = append(s.state.Statuses, status)
		}
	}

	// Initialize users (includes employee records)
	status := InitStatus{
		Component: "users",
		Status:    "initialized",
	}

	logger.Infof(ctx, "Initializing users and employees...")
	if err := s.checkUsersInitialized(ctx); err != nil {
		status.Status = "failed"
		status.Error = err.Error()
		s.state.Statuses = append(s.state.Statuses, status)
		logger.Errorf(ctx, "Failed to initialize users: %v", err)
		return s.state, fmt.Errorf("initialization step users failed: %v", err)
	}

	s.state.Statuses = append(s.state.Statuses, status)
	logger.Infof(ctx, "Successfully initialized users and employees")

	// Also initialize Casbin policies
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

	if s.c.Initialization.PersistState {
		if err := s.SaveState(ctx); err != nil {
			logger.Warnf(ctx, "Failed to save initialization state: %v", err)
		}
	}

	logger.Infof(ctx, "User initialization completed successfully in %s mode", s.state.DataMode)
	return s.state, nil
}

// VerifyUserRoleAssignment verifies user role assignments with detailed logging
func (s *Service) VerifyUserRoleAssignment(ctx context.Context, username string) error {
	logger.Infof(ctx, "Verifying role assignment for user: %s", username)

	user, err := s.us.User.Get(ctx, username)
	if err != nil {
		return fmt.Errorf("user '%s' not found: %w", username, err)
	}

	globalRoles, err := s.acs.UserRole.GetUserRoles(ctx, user.ID)
	if err != nil {
		logger.Warnf(ctx, "Failed to get global roles for user '%s': %v", username, err)
	} else {
		logger.Infof(ctx, "User '%s' has %d global roles:", username, len(globalRoles))
		for _, role := range globalRoles {
			logger.Infof(ctx, "  - Global role: %s (%s)", role.Name, role.Slug)
		}
	}

	userTenants, err := s.ts.UserTenant.UserBelongTenants(ctx, user.ID)
	if err != nil {
		logger.Warnf(ctx, "Failed to get tenants for user '%s': %v", username, err)
		return nil
	}

	logger.Infof(ctx, "User '%s' belongs to %d tenants:", username, len(userTenants))
	for _, tenant := range userTenants {
		logger.Infof(ctx, "  - Tenant: %s (%s)", tenant.Name, tenant.Slug)

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
