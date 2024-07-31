package initialize

import (
	"context"
	"ncobase/common/log"
	tenantStructs "ncobase/feature/tenant/structs"
	userStructs "ncobase/feature/user/structs"
)

// checkUsersInitialized checks if users are already initialized.
func (s *InitializeService) checkUsersInitialized(ctx context.Context) error {
	params := &userStructs.ListUserParams{}
	count := s.us.User.CountX(ctx, params)
	if count == 0 {
		return s.initUsers(ctx)
	}

	return nil
}

// initUsers initializes users, their tenants, roles, and tenant roles.
func (s *InitializeService) initUsers(ctx context.Context) error {
	users := []userStructs.UserBody{
		{
			Username:    "super",
			Email:       "super@example.com",
			Phone:       "13800000100",
			IsCertified: true,
			IsAdmin:     true,
		},
		{
			Username:    "admin",
			Email:       "admin@example.com",
			Phone:       "13800000101",
			IsCertified: true,
			IsAdmin:     true,
		},
		{
			Username:    "user",
			Email:       "user@example.com",
			Phone:       "13800000102",
			IsCertified: true,
			IsAdmin:     false,
		},
	}

	for _, user := range users {
		createdUser, err := s.us.User.CreateUser(ctx, &user)
		if err != nil {
			log.Errorf(ctx, "initUsers error on create user: %v\n", err)
			return err
		}

		// update user password
		if err = s.us.User.UpdatePassword(ctx, &userStructs.UserPassword{
			User:        createdUser.Username,
			NewPassword: "Ac123456",
			Confirm:     "Ac123456",
		}); err != nil {
			log.Errorf(ctx, "initUsers error on update user password: %v\n", err)
			return err
		}

		// create user profile
		if _, err := s.us.UserProfile.Create(ctx, &userStructs.UserProfileBody{
			ID:          createdUser.ID,
			DisplayName: user.Username,
		}); err != nil {
			log.Errorf(ctx, "initUsers error on create user profile: %v\n", err)
			return err
		}

		if user.Username == "super" {
			tenantBody := &tenantStructs.CreateTenantBody{
				TenantBody: tenantStructs.TenantBody{
					Name:      "Ncobase Co, Ltd.",
					Slug:      "ncobase",
					CreatedBy: &createdUser.ID,
				},
			}

			if _, err := s.as.AuthTenant.CreateInitialTenant(ctx, tenantBody); err != nil {
				log.Errorf(ctx, "initUsers error on create initial tenant: %v\n", err)
				return err
			}
		}

		// Skip role assignment for "super" as it's handled by createInitialTenant
		if user.Username != "super" {
			// related to tenant
			existedTenant, err := s.ts.Tenant.GetBySlug(ctx, "ncobase")
			if err != nil {
				log.Errorf(ctx, "initUsers error on get tenant: %v\n", err)
				return err
			}
			if _, err := s.ts.UserTenant.AddUserToTenant(ctx, createdUser.ID, existedTenant.ID); err != nil {
				log.Errorf(ctx, "initUsers error on create user tenant: %v\n", err)
				return err
			}
			// Assign roles based on user type
			var roleSlug string
			if user.Username == "admin" {
				roleSlug = "admin"
			} else {
				roleSlug = "user"
			}

			role, err := s.acs.Role.GetBySlug(ctx, roleSlug)
			if err != nil {
				log.Errorf(ctx, "initUsers error on get role (%s): %v\n", roleSlug, err)
				return err
			}
			if err := s.acs.UserRole.AddRoleToUser(ctx, createdUser.ID, role.ID); err != nil {
				log.Errorf(ctx, "initUsers error on create role (%s): %v\n", roleSlug, err)
				return err
			}
			if _, err := s.acs.UserTenantRole.AddRoleToUserInTenant(ctx, createdUser.ID, existedTenant.ID, role.ID); err != nil {
				log.Errorf(ctx, "initUsers error on create tenant role (%s): %v\n", roleSlug, err)
				return err
			}
		}
	}

	log.Infof(ctx, "-------- initUsers done, created %d users\n", len(users))

	return nil
}
