package service

import (
	"context"
	"fmt"
	"ncobase/initialize/data"
	tenantStructs "ncobase/tenant/structs"
	userStructs "ncobase/user/structs"
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

// initUsers initializes users, their tenants, roles, and tenant roles.
func (s *Service) initUsers(ctx context.Context) error {
	logger.Infof(ctx, "Initializing system users...")

	// Get default tenant
	defaultTenant, err := s.ts.Tenant.GetBySlug(ctx, "ncobase")
	if err != nil {
		logger.Errorf(ctx, "Error getting default tenant: %v", err)
		return err
	}

	for _, userInfo := range data.SystemDefaultUsers {
		// Create user
		createdUser, err := s.us.User.CreateUser(ctx, &userInfo.User)
		if err != nil {
			logger.Errorf(ctx, "Error creating user %s: %v", userInfo.User.Username, err)
			return err
		}
		logger.Debugf(ctx, "Created user: %s", userInfo.User.Username)

		// Set password
		if err = s.us.User.UpdatePassword(ctx, &userStructs.UserPassword{
			User:        createdUser.Username,
			NewPassword: userInfo.Password,
			Confirm:     userInfo.Password,
		}); err != nil {
			logger.Errorf(ctx, "Error setting password for user %s: %v", createdUser.Username, err)
			return err
		}

		// Create user profile
		profileData := userInfo.Profile
		profileData.ID = createdUser.ID // Set the user ID
		if _, err := s.us.UserProfile.Create(ctx, &profileData); err != nil {
			logger.Errorf(ctx, "Error creating profile for user %s: %v", createdUser.Username, err)
			return err
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
					return err
				}

				// Get super-admin role
				superAdminRole, err := s.acs.Role.GetBySlug(ctx, "super-admin")
				if err != nil {
					logger.Errorf(ctx, "Error getting super-admin role: %v", err)
					return err
				}

				// Assign super-admin role to user
				if err := s.acs.UserRole.AddRoleToUser(ctx, createdUser.ID, superAdminRole.ID); err != nil {
					logger.Errorf(ctx, "Error adding super-admin role to super user: %v", err)
					return err
				}

				// Assign the tenant role to the super user
				if _, err := s.acs.UserTenantRole.AddRoleToUserInTenant(ctx, createdUser.ID, existingTenant.ID, superAdminRole.ID); err != nil {
					logger.Errorf(ctx, "Error adding tenant role to super user: %v", err)
					return err
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
					return err
				}
			}
		} else {
			// For non-super users, add to default tenant
			if _, err := s.ts.UserTenant.AddUserToTenant(ctx, createdUser.ID, defaultTenant.ID); err != nil {
				logger.Errorf(ctx, "Error adding user %s to tenant: %v", createdUser.Username, err)
				return err
			}

			// Assign role based on configuration
			role, err := s.acs.Role.GetBySlug(ctx, userInfo.Role)
			if err != nil {
				logger.Errorf(ctx, "Error getting role %s: %v", userInfo.Role, err)
				return err
			}

			if err := s.acs.UserRole.AddRoleToUser(ctx, createdUser.ID, role.ID); err != nil {
				logger.Errorf(ctx, "Error adding role to user %s: %v", createdUser.Username, err)
				return err
			}

			if _, err := s.acs.UserTenantRole.AddRoleToUserInTenant(ctx, createdUser.ID, defaultTenant.ID, role.ID); err != nil {
				logger.Errorf(ctx, "Error adding tenant role to user %s: %v", createdUser.Username, err)
				return err
			}
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
	logger.Infof(ctx, "User initialization completed successfully")
	return s.state, nil
}
