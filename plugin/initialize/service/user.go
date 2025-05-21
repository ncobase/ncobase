package service

import (
	"context"
	"fmt"
	"ncobase/initialize/data"
	tenantStructs "ncobase/tenant/structs"
	userStructs "ncobase/user/structs"
	"strings"
	"time"

	"github.com/ncobase/ncore/logging/logger"
)

// checkUsersInitialized checks if users are already initialized.
func (s *Service) checkUsersInitialized(ctx context.Context) error {
	params := &userStructs.ListUserParams{}
	count := s.us.User.CountX(ctx, params)
	if count > 0 {
		logger.Infof(ctx, "Users already exist, skipping initialization")
		return nil
	}

	return s.initUsers(ctx)
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

// initUsers initializes users, their tenants, roles, and tenant roles.
func (s *Service) initUsers(ctx context.Context) error {
	logger.Infof(ctx, "Initializing system users...")

	// Get default tenant
	defaultTenant, err := s.ts.Tenant.GetBySlug(ctx, "ncobase")
	if err != nil {
		logger.Errorf(ctx, "Error getting default tenant: %v", err)
		return fmt.Errorf("default tenant 'ncobase' not found: %w", err)
	}

	var createdCount int
	for _, userInfo := range data.SystemDefaultUsers {
		// Validate password
		if err := s.validatePassword(ctx, userInfo.User.Username, userInfo.Password); err != nil {
			logger.Warnf(ctx, "Password for user %s does not meet policy: %v", userInfo.User.Username, err)
			// Continue anyway as these are initial users
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
		profileData.ID = createdUser.ID // Set the user ID
		if _, err := s.us.UserProfile.Create(ctx, &profileData); err != nil {
			logger.Errorf(ctx, "Error creating profile for user %s: %v", createdUser.Username, err)
			return fmt.Errorf("failed to create profile for user '%s': %w", createdUser.Username, err)
		}

		// Handle the "super" user differently
		if userInfo.User.Username == "super" {
			// Check if the tenant already exists
			existingTenant, err := s.ts.Tenant.GetBySlug(ctx, defaultTenant.Slug)
			if err == nil && existingTenant != nil {
				logger.Infof(ctx, "Initial tenant already exists, ensuring super user is associated with it")

				// Ensure super user is added to the existing tenant
				_, err = s.ts.UserTenant.AddUserToTenant(ctx, createdUser.ID, existingTenant.ID)
				if err != nil {
					logger.Errorf(ctx, "Error adding super user to tenant: %v", err)
					return fmt.Errorf("failed to add super user to tenant: %w", err)
				}

				// Get super-admin role
				superAdminRole, err := s.acs.Role.GetBySlug(ctx, "super-admin")
				if err != nil {
					logger.Errorf(ctx, "Error getting super-admin role: %v", err)
					return fmt.Errorf("super-admin role not found: %w", err)
				}

				// Assign super-admin role to user
				if err := s.acs.UserRole.AddRoleToUser(ctx, createdUser.ID, superAdminRole.ID); err != nil {
					logger.Errorf(ctx, "Error adding super-admin role to super user: %v", err)
					return fmt.Errorf("failed to add super-admin role to super user: %w", err)
				}

				// Assign the tenant role to the super user
				if _, err := s.acs.UserTenantRole.AddRoleToUserInTenant(ctx, createdUser.ID, existingTenant.ID, superAdminRole.ID); err != nil {
					logger.Errorf(ctx, "Error adding tenant role to super user: %v", err)
					return fmt.Errorf("failed to add tenant role to super user: %w", err)
				}
			} else {
				// Create initial tenant with super user as creator
				tenantBody := &tenantStructs.CreateTenantBody{
					TenantBody: tenantStructs.TenantBody{
						Name:      defaultTenant.Name,
						Slug:      defaultTenant.Slug,
						CreatedBy: &createdUser.ID,
					},
				}

				if _, err := s.as.AuthTenant.CreateInitialTenant(ctx, tenantBody); err != nil {
					logger.Errorf(ctx, "Error creating initial tenant: %v", err)
					return fmt.Errorf("failed to create initial tenant: %w", err)
				}
			}
		} else {
			// For non-super users, add to default tenant
			if _, err := s.ts.UserTenant.AddUserToTenant(ctx, createdUser.ID, defaultTenant.ID); err != nil {
				logger.Errorf(ctx, "Error adding user %s to tenant: %v", createdUser.Username, err)
				return fmt.Errorf("failed to add user '%s' to tenant: %w", createdUser.Username, err)
			}

			// Assign role based on configuration
			role, err := s.acs.Role.GetBySlug(ctx, userInfo.Role)
			if err != nil {
				logger.Errorf(ctx, "Error getting role %s: %v", userInfo.Role, err)
				return fmt.Errorf("role '%s' not found for user '%s': %w", userInfo.Role, createdUser.Username, err)
			}

			if err := s.acs.UserRole.AddRoleToUser(ctx, createdUser.ID, role.ID); err != nil {
				logger.Errorf(ctx, "Error adding role to user %s: %v", createdUser.Username, err)
				return fmt.Errorf("failed to add role '%s' to user '%s': %w", userInfo.Role, createdUser.Username, err)
			}

			if _, err := s.acs.UserTenantRole.AddRoleToUserInTenant(ctx, createdUser.ID, defaultTenant.ID, role.ID); err != nil {
				logger.Errorf(ctx, "Error adding tenant role to user %s: %v", createdUser.Username, err)
				return fmt.Errorf("failed to add tenant role to user '%s': %w", createdUser.Username, err)
			}
		}
	}

	// Verify required users were created
	reqUsers := []string{"super", "admin", "user"}
	for _, username := range reqUsers {
		user, err := s.us.User.Get(ctx, username)
		if err != nil || user == nil {
			logger.Errorf(ctx, "Required user '%s' was not created", username)
			return fmt.Errorf("required user '%s' was not created: %w", username, err)
		}
	}

	count := s.us.User.CountX(ctx, &userStructs.ListUserParams{})
	logger.Infof(ctx, "User initialization completed, created %d users", count)

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
