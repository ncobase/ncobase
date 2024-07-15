package service

import (
	"context"
	"fmt"
	"ncobase/common/log"
	"ncobase/common/types"
	accessStructs "ncobase/feature/access/structs"
	tenantStructs "ncobase/feature/tenant/structs"
	userStructs "ncobase/feature/user/structs"
)

// InitData initializes roles, permissions, Casbin policies, and initial users if necessary.
func (s *Service) InitData() error {
	ctx := context.Background()

	steps := []func(context.Context) error{
		s.checkRolesInitialized,
		s.checkPermissionsInitialized,
		s.checkUsersInitialized,
		s.checkDomainsInitialized,
		s.checkCasbinPoliciesInitialized,
	}

	for _, step := range steps {
		if err := step(ctx); err != nil {
			return fmt.Errorf("initialization step failed: %v", err)
		}
	}

	return nil
}

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
		if err := s.us.User.UpdatePassword(ctx, &userStructs.UserPassword{
			User:        createdUser.Username,
			NewPassword: "Ac123456",
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

// checkDomainsInitialized checks if domains are already initialized.
func (s *Service) checkDomainsInitialized(ctx context.Context) error {
	params := &tenantStructs.ListTenantParams{}
	count := s.ts.Tenant.CountX(ctx, params)
	if count == 0 {
		return s.initTenants(ctx)
	}
	return nil
}

// initTenants initializes the domains (tenants).
func (s *Service) initTenants(ctx context.Context) error {
	tenants := []tenantStructs.CreateTenantBody{
		{
			TenantBody: tenantStructs.TenantBody{
				Name:      "Ncobase Co, Ltd.",
				Slug:      "ncobase",
				CreatedBy: nil,
			},
		},
	}

	for _, tenant := range tenants {
		if _, err := s.ts.Tenant.Create(ctx, &tenant); err != nil {
			log.Errorf(ctx, "initTenants error on create domain: %v\n", err)
			return err
		}
	}

	log.Infof(ctx, "-------- initTenants done, created %d domains\n", len(tenants))

	return nil
}

// checkRolesInitialized checks if roles are already initialized.
func (s *Service) checkRolesInitialized(ctx context.Context) error {
	params := &accessStructs.ListRoleParams{}
	count := s.acs.Role.CountX(ctx, params)
	if count == 0 {
		return s.initRoles(ctx)
	}

	return nil
}

// initRoles initializes roles.
func (s *Service) initRoles(ctx context.Context) error {
	roles := []*accessStructs.CreateRoleBody{
		{
			RoleBody: accessStructs.RoleBody{
				Name:        "Super Admin",
				Slug:        "super-admin",
				Disabled:    false,
				Description: "Super Administrator role with all permissions",
				Extras:      nil,
			},
		},
		{
			RoleBody: accessStructs.RoleBody{
				Name:        "Admin",
				Slug:        "admin",
				Disabled:    false,
				Description: "Administrator role with some permissions",
				Extras:      nil,
			},
		},
		{
			RoleBody: accessStructs.RoleBody{
				Name:        "User",
				Slug:        "user",
				Disabled:    false,
				Description: "User role with some permissions",
				Extras:      nil,
			},
		},
	}

	for _, role := range roles {
		if _, err := s.acs.Role.Create(ctx, role); err != nil {
			log.Errorf(ctx, "initRoles error on create role: %v\n", err)
			return err
		}
	}

	count := s.acs.Role.CountX(ctx, &accessStructs.ListRoleParams{})
	log.Infof(ctx, "-------- initRoles done, created %d roles\n", count)

	return nil
}

// checkPermissionsInitialized checks if permissions are already initialized.
func (s *Service) checkPermissionsInitialized(ctx context.Context) error {
	params := &accessStructs.ListPermissionParams{}
	count := s.acs.Permission.CountX(ctx, params)
	if count == 0 {
		return s.initPermissions(ctx)
	}

	return nil
}

// initPermissions initializes permissions and their relationships.
func (s *Service) initPermissions(ctx context.Context) error {
	permissions := []accessStructs.CreatePermissionBody{
		{
			PermissionBody: accessStructs.PermissionBody{
				Name:        "all",
				Action:      "*",
				Subject:     "*",
				Description: "Full access to all resources",
			},
		},
		{
			PermissionBody: accessStructs.PermissionBody{
				Name:        "read_all",
				Action:      "GET",
				Subject:     "*",
				Description: "Read access to all resources",
			},
		},
		{
			PermissionBody: accessStructs.PermissionBody{
				Name:        "write_all",
				Action:      "POST",
				Subject:     "*",
				Description: "Write access to all resources",
			},
		},
		{
			PermissionBody: accessStructs.PermissionBody{
				Name:        "update_all",
				Action:      "POST",
				Subject:     "*",
				Description: "Update access to all resources",
			},
		},
		{
			PermissionBody: accessStructs.PermissionBody{
				Name:        "delete_all",
				Action:      "POST",
				Subject:     "*",
				Description: "Delete access to all resources",
			},
		},
	}

	for _, permission := range permissions {
		if _, err := s.acs.Permission.Create(ctx, &permission); err != nil {
			log.Errorf(ctx, "initPermissions error on create permission: %v\n", err)
			return err
		}
	}

	allPermissions, err := s.acs.Permission.List(ctx, &accessStructs.ListPermissionParams{})
	if err != nil {
		log.Errorf(ctx, "initPermissions error on list permissions: %v\n", err)
		return err
	}

	aps := allPermissions["all_permissions"].([]*accessStructs.ReadPermission)

	roles, err := s.acs.Role.List(ctx, &accessStructs.ListRoleParams{})
	if err != nil {
		log.Errorf(ctx, "initPermissions error on list roles: %v\n", err)
		return err
	}

	rs := roles["content"].([]*accessStructs.ReadRole)

	for _, role := range rs {
		for _, perm := range aps {
			var assignPermission bool
			switch role.Slug {
			case "super-admin":
				// Super Admin gets all permissions
				assignPermission = true
			case "admin":
				// Admin gets read and write permissions
				if perm.Action == "GET" || perm.Action == "POST" {
					assignPermission = true
				}
			case "user":
				// User gets read permissions only
				if perm.Action == "GET" {
					assignPermission = true
				}
			}

			if assignPermission {
				if _, err := s.acs.RolePermission.AddPermissionToRole(ctx, role.ID, perm.ID); err != nil {
					log.Errorf(ctx, "initPermissions error on create role-permission: %v\n", err)
					return err
				}
			}
		}
	}

	count := s.acs.Permission.CountX(ctx, &accessStructs.ListPermissionParams{})
	log.Infof(ctx, "-------- initPermissions done, created %d permissions\n", count)

	return nil
}

// checkCasbinPoliciesInitialized checks if Casbin policies are already initialized.
func (s *Service) checkCasbinPoliciesInitialized(ctx context.Context) error {
	params := &accessStructs.ListCasbinRuleParams{}
	count := s.acs.Casbin.CountX(ctx, params)
	if count == 0 {
		return s.initCasbinPolicies(ctx)
	}

	return nil
}

// initCasbinPolicies initializes Casbin policies.
func (s *Service) initCasbinPolicies(ctx context.Context) error {
	defaultTenant, err := s.ts.Tenant.GetBySlug(ctx, "ncobase")
	if err != nil {
		log.Errorf(ctx, "initCasbinPolicies error on get default tenant: %v\n", err)
		return err
	}
	allRoles, err := s.acs.Role.List(ctx, &accessStructs.ListRoleParams{})
	if err != nil {
		log.Errorf(ctx, "initCasbinPolicies error on list roles: %v\n", err)
		return err
	}

	roles := allRoles["content"].([]*accessStructs.ReadRole)

	// Initialize policies based on role-permission relationship
	for _, role := range roles {
		rolePermissions, err := s.acs.RolePermission.GetRolePermissions(ctx, role.ID)
		if err != nil {
			log.Errorf(ctx, "initCasbinPolicies error on list role permissions for role %s: %v\n", role.Slug, err)
			return err
		}

		for _, p := range rolePermissions {
			permission, err := s.acs.Permission.GetByID(ctx, p.ID)
			if err != nil {
				log.Errorf(ctx, "initCasbinPolicies error on get permission %s: %v\n", p.ID, err)
				return err
			}

			policy := accessStructs.CasbinRuleBody{
				PType: "p",
				V0:    role.Slug,                          // sub
				V1:    defaultTenant.ID,                   // dom
				V2:    permission.Subject,                 // obj
				V3:    types.ToPointer(permission.Action), // act
				// V4, V5 are not used
			}

			if _, err := s.acs.Casbin.Create(ctx, &policy); err != nil {
				log.Errorf(ctx, "initCasbinPolicies error on create casbin rule: %v\n", err)
				return err
			}
		}
	}

	count := s.acs.Casbin.CountX(ctx, &accessStructs.ListCasbinRuleParams{})
	log.Infof(ctx, "-------- initCasbinPolicies done, created %d policies\n", count)

	return nil
}
