package service

import (
	"context"
	"fmt"
	orgStructs "ncobase/organization/structs"
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

// initUsers initializes users based on current data mode
func (s *Service) initUsers(ctx context.Context) error {
	logger.Infof(ctx, "Initializing users in %s mode...", s.state.DataMode)

	// Get default space
	space, err := s.getDefaultSpace(ctx)
	if err != nil {
		logger.Errorf(ctx, "Error getting default space: %v", err)
		return err
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

		// Create employee record only for company/enterprise modes
		if userInfo.Employee != nil && (s.state.DataMode == "company" || s.state.DataMode == "enterprise") {
			if err := s.createEmployeeRecord(ctx, createdUser, userInfo.Employee, space.ID, false); err != nil {
				logger.Errorf(ctx, "Error creating employee record for user %s: %v", createdUser.Username, err)
				return fmt.Errorf("failed to create employee record for user '%s': %w", createdUser.Username, err)
			}
			employeeCount++
		}

		// Assign user role
		if err := s.assignUserRole(ctx, createdUser, userInfo.Role, space.ID); err != nil {
			logger.Errorf(ctx, "Error assigning role to user %s: %v", createdUser.Username, err)
			return fmt.Errorf("failed to assign role to user '%s': %w", createdUser.Username, err)
		}
	}

	// Phase 2: Resolve manager references for company/enterprise modes
	if s.state.DataMode == "company" || s.state.DataMode == "enterprise" {
		if err := s.resolveManagerReferences(ctx); err != nil {
			logger.Warnf(ctx, "Failed to resolve some manager references: %v", err)
		}
	}

	// Phase 3: Initialize user-organization assignments
	if err := s.initializeUserOrganizationAssignments(ctx); err != nil {
		logger.Errorf(ctx, "Failed to initialize user-organization assignments: %v", err)
		return fmt.Errorf("user-organization assignment failed: %w", err)
	}

	// Phase 4: Assign employees to departments (only for company/enterprise)
	if s.state.DataMode == "company" || s.state.DataMode == "enterprise" {
		if err := s.assignEmployeesToDepartments(ctx); err != nil {
			logger.Warnf(ctx, "Failed to assign some employees to departments: %v", err)
		}
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
	if s.state.DataMode == "website" {
		logger.Infof(ctx, "User initialization completed in %s mode, created %d users", s.state.DataMode, userCount)
	} else {
		logger.Infof(ctx, "User and employee initialization completed in %s mode, created %d users and %d employee records",
			s.state.DataMode, userCount, employeeCount)
	}

	return nil
}

// createEmployeeRecord creates employee record with optional manager resolution
func (s *Service) createEmployeeRecord(ctx context.Context, user *userStructs.ReadUser, employeeData *userStructs.EmployeeBody, spaceID string, resolveManager bool) error {
	employeeBody := *employeeData
	employeeBody.UserID = user.ID
	employeeBody.SpaceID = spaceID

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
				SpaceID:        employee.SpaceID,
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

// assignEmployeesToDepartments assigns employees to their department organizations
func (s *Service) assignEmployeesToDepartments(ctx context.Context) error {
	logger.Infof(ctx, "Assigning employees to department organizations...")

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

		// Find department organization
		organization, err := s.findGroupBySlug(ctx, employee.Department)
		if err != nil {
			logger.Warnf(ctx, "Department organization '%s' not found for employee %s", employee.Department, user.Username)
			continue
		}

		// Check if user is already in the organization
		isMember, err := s.ss.UserOrganization.IsUserMember(ctx, user.ID, organization.ID)
		if err != nil {
			logger.Warnf(ctx, "Failed to check organization membership for user %s: %v", user.Username, err)
			continue
		}

		if !isMember {
			// Add user to department organization
			if _, err := s.ss.UserOrganization.AddUserToOrganization(ctx, user.ID, organization.ID, orgStructs.RoleMember); err != nil {
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

// assignUserRole assigns both global and space roles to user
func (s *Service) assignUserRole(ctx context.Context, user *userStructs.ReadUser, roleSlug string, spaceID string) error {
	role, err := s.acs.Role.GetBySlug(ctx, roleSlug)
	if err != nil {
		logger.Errorf(ctx, "Role '%s' not found: %v", roleSlug, err)
		return fmt.Errorf("role '%s' not found: %w", roleSlug, err)
	}

	if _, err := s.ts.UserSpace.AddUserToSpace(ctx, user.ID, spaceID); err != nil {
		logger.Errorf(ctx, "Failed to add user to space: %v", err)
		return fmt.Errorf("failed to add user '%s' to space: %w", user.Username, err)
	}

	if err := s.acs.UserRole.AddRoleToUser(ctx, user.ID, role.ID); err != nil {
		logger.Errorf(ctx, "Failed to assign global role: %v", err)
		return fmt.Errorf("failed to assign global role '%s' to user '%s': %w", roleSlug, user.Username, err)
	}

	if _, err := s.ts.UserSpaceRole.AddRoleToUserInSpace(ctx, user.ID, spaceID, role.ID); err != nil {
		logger.Warnf(ctx, "Failed to assign space role to user '%s': %v", user.Username, err)
	}

	logger.Infof(ctx, "Successfully assigned role '%s' to user '%s'", roleSlug, user.Username)
	return nil
}

// initializeUserOrganizationAssignments assigns users to organizations based on mode
func (s *Service) initializeUserOrganizationAssignments(ctx context.Context) error {
	logger.Infof(ctx, "Initializing user-organization assignments...")

	var assignments map[string][]string

	switch s.state.DataMode {
	case "website":
		assignments = map[string][]string{
			"admin":   {"website-platform"},
			"manager": {"website-platform"},
			"member":  {"website-platform"},
		}
	case "company":
		assignments = map[string][]string{
			"admin":         {"digital-company"},
			"company.admin": {"digital-company"},
			"manager":       {"digital-company"},
			"employee":      {"digital-company"},
		}
	case "enterprise":
		assignments = map[string][]string{
			"admin":            {"digital-enterprise"},
			"enterprise.admin": {"digital-enterprise"},
			"dept.manager":     {"digital-enterprise"},
			"team.leader":      {"digital-enterprise"},
			"employee":         {"digital-enterprise"},
		}
	default:
		assignments = make(map[string][]string)
	}

	var assignmentCount int
	for username, organizationSlugs := range assignments {
		user, err := s.us.User.Get(ctx, username)
		if err != nil {
			logger.Warnf(ctx, "User '%s' not found, skipping organization assignments", username)
			continue
		}

		for _, organizationSlug := range organizationSlugs {
			organization, err := s.findGroupBySlug(ctx, organizationSlug)
			if err != nil {
				logger.Warnf(ctx, "Group '%s' not found for user '%s'", organizationSlug, username)
				continue
			}

			if _, err := s.ss.UserOrganization.AddUserToOrganization(ctx, user.ID, organization.ID, orgStructs.RoleMember); err != nil {
				logger.Warnf(ctx, "Failed to assign user '%s' to organization '%s': %v", username, organizationSlug, err)
				continue
			}

			logger.Debugf(ctx, "Assigned user '%s' to organization '%s'", username, organizationSlug)
			assignmentCount++
		}
	}

	logger.Infof(ctx, "User-organization assignment completed, created %d assignments", assignmentCount)
	return nil
}

// findGroupBySlug finds organization by slug
func (s *Service) findGroupBySlug(ctx context.Context, slug string) (*orgStructs.ReadOrganization, error) {
	organizations, err := s.ss.Organization.List(ctx, &orgStructs.ListOrganizationParams{})
	if err != nil {
		return nil, err
	}

	for _, organization := range organizations.Items {
		if organization.Slug == slug {
			return organization, nil
		}
	}

	return nil, fmt.Errorf("organization with slug '%s' not found", slug)
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
			{"spaces", s.checkSpacesInitialized},
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

	userSpaces, err := s.ts.UserSpace.UserBelongSpaces(ctx, user.ID)
	if err != nil {
		logger.Warnf(ctx, "Failed to get spaces for user '%s': %v", username, err)
		return nil
	}

	logger.Infof(ctx, "User '%s' belongs to %d spaces:", username, len(userSpaces))
	for _, space := range userSpaces {
		logger.Infof(ctx, "  - Space: %s (%s)", space.Name, space.Slug)

		ctx = ctxutil.SetSpaceID(ctx, space.ID)
		roleIDs, roleErr := s.ts.UserSpaceRole.GetUserRolesInSpace(ctx, user.ID, space.ID)
		if roleErr == nil && len(roleIDs) > 0 {
			spaceRoles, _ := s.acs.Role.GetByIDs(ctx, roleIDs)
			logger.Infof(ctx, "    User has %d roles in space '%s':", len(spaceRoles), space.Name)
			for _, role := range spaceRoles {
				logger.Infof(ctx, "      - Space role: %s (%s)", role.Name, role.Slug)
			}
		}
	}

	return nil
}
