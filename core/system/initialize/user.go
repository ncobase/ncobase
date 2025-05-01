package initialize

import (
	"context"
	"ncobase/core/system/initialize/data"
	tenantStructs "ncobase/core/tenant/structs"
	userStructs "ncobase/core/user/structs"

	"github.com/ncobase/ncore/logging/logger"
)

// checkUsersInitialized checks if users are already initialized.
func (s *Service) checkUsersInitialized(ctx context.Context) error {
	params := &userStructs.ListUserParams{}
	count := s.us.User.CountX(ctx, params)
	if count == 0 {
		return s.initUsers(ctx)
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

		// Handle the "super" user differently - create initial tenant
		if userInfo.User.Username == "super" {
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
