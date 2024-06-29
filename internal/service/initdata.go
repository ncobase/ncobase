package service

import (
	"context"
	"fmt"
	"ncobase/common/log"
	"ncobase/common/types"
	"ncobase/internal/data/structs"
)

// initData initializes roles, permissions, Casbin policies, and initial users if necessary.
func (svc *Service) initData() error {
	ctx := context.Background()

	steps := []func(context.Context) error{
		svc.checkRolesInitialized,
		svc.checkPermissionsInitialized,
		svc.checkUsersInitialized,
		svc.checkDomainsInitialized,
		svc.checkCasbinPoliciesInitialized,
	}

	for _, step := range steps {
		if err := step(ctx); err != nil {
			return fmt.Errorf("initialization step failed: %v", err)
		}
	}

	return nil
}

// checkUsersInitialized checks if users are already initialized.
func (svc *Service) checkUsersInitialized(ctx context.Context) error {
	params := &structs.ListUserParams{}
	count := svc.user.CountX(ctx, params)
	if count == 0 {
		return svc.initUsers(ctx)
	}

	return nil
}

// initUsers initializes users, their tenants, roles, and tenant roles.
func (svc *Service) initUsers(ctx context.Context) error {
	users := []structs.UserBody{
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
		createdUser, err := svc.user.Create(ctx, &user)
		if err != nil {
			log.Errorf(ctx, "initUsers error on create user: %v\n", err)
			return err
		}

		// update user password
		if err := svc.updateUserPassword(ctx, &structs.UserPassword{
			User:        createdUser.Username,
			NewPassword: "Ac123456",
		}); err != nil {
			log.Errorf(ctx, "initUsers error on update user password: %v\n", err)
			return err
		}

		// create user profile
		if _, err := svc.userProfile.Create(ctx, &structs.UserProfileBody{
			ID:          createdUser.ID,
			DisplayName: user.Username,
		}); err != nil {
			log.Errorf(ctx, "initUsers error on create user profile: %v\n", err)
			return err
		}

		if user.Username == "super" {
			tenantBody := &structs.CreateTenantBody{
				TenantBody: structs.TenantBody{
					Name: "Ncobase Co, Ltd.",
					Slug: "ncobase",
					OperatorBy: structs.OperatorBy{
						CreatedBy: &createdUser.ID,
					},
				},
			}

			if _, err := svc.createInitialTenant(ctx, tenantBody); err != nil {
				log.Errorf(ctx, "initUsers error on create initial tenant: %v\n", err)
				return err
			}
		}

		// Skip role assignment for "super" as it's handled by createInitialTenant
		if user.Username != "super" {
			// related to tenant
			existedTenant, err := svc.tenant.GetBySlug(ctx, "ncobase")
			if err != nil {
				log.Errorf(ctx, "initUsers error on get tenant: %v\n", err)
				return err
			}
			if _, err := svc.userTenant.Create(ctx, &structs.UserTenant{
				UserID:   createdUser.ID,
				TenantID: existedTenant.ID,
			}); err != nil {
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

			role, err := svc.role.GetBySlug(ctx, roleSlug)
			if err != nil {
				log.Errorf(ctx, "initUsers error on get role (%s): %v\n", roleSlug, err)
				return err
			}
			if _, err := svc.userRole.Create(ctx, &structs.UserRole{
				UserID: createdUser.ID,
				RoleID: role.ID,
			}); err != nil {
				log.Errorf(ctx, "initUsers error on create role (%s): %v\n", roleSlug, err)
				return err
			}
			if _, err := svc.userTenantRole.Create(ctx, &structs.UserTenantRole{
				UserID:   createdUser.ID,
				TenantID: existedTenant.ID,
				RoleID:   role.ID,
			}); err != nil {
				log.Errorf(ctx, "initUsers error on create tenant role (%s): %v\n", roleSlug, err)
				return err
			}
		}
	}

	log.Infof(ctx, "-------- initUsers done, created %d users\n", len(users))

	return nil
}

// checkDomainsInitialized checks if domains are already initialized.
func (svc *Service) checkDomainsInitialized(ctx context.Context) error {
	params := &structs.ListTenantParams{}
	count := svc.tenant.CountX(ctx, params)
	if count == 0 {
		return svc.initDomains(ctx)
	}
	return nil
}

// initDomains initializes the domains (tenants).
func (svc *Service) initDomains(ctx context.Context) error {
	domains := []structs.CreateTenantBody{
		{
			TenantBody: structs.TenantBody{
				Name: "Ncobase Co, Ltd.",
				Slug: "ncobase",
				OperatorBy: structs.OperatorBy{
					CreatedBy: nil,
				},
			},
		},
	}

	for _, domain := range domains {
		if _, err := svc.tenant.Create(ctx, &domain); err != nil {
			log.Errorf(ctx, "initDomains error on create domain: %v\n", err)
			return err
		}
	}

	log.Infof(ctx, "-------- initDomains done, created %d domains\n", len(domains))

	return nil
}

// checkRolesInitialized checks if roles are already initialized.
func (svc *Service) checkRolesInitialized(ctx context.Context) error {
	params := &structs.ListRoleParams{}
	count := svc.role.CountX(ctx, params)
	if count == 0 {
		return svc.initRoles(ctx)
	}

	return nil
}

// initRoles initializes roles.
func (svc *Service) initRoles(ctx context.Context) error {
	roles := []*structs.CreateRoleBody{
		{
			RoleBody: structs.RoleBody{
				Name:        "Super Admin",
				Slug:        "super-admin",
				Disabled:    false,
				Description: "Super Administrator role with all permissions",
				Extras:      nil,
			},
		},
		{
			RoleBody: structs.RoleBody{
				Name:        "Admin",
				Slug:        "admin",
				Disabled:    false,
				Description: "Administrator role with some permissions",
				Extras:      nil,
			},
		},
		{
			RoleBody: structs.RoleBody{
				Name:        "User",
				Slug:        "user",
				Disabled:    false,
				Description: "User role with some permissions",
				Extras:      nil,
			},
		},
	}

	for _, role := range roles {
		if _, err := svc.role.Create(ctx, role); err != nil {
			log.Errorf(ctx, "initRoles error on create role: %v\n", err)
			return err
		}
	}

	count := svc.role.CountX(ctx, &structs.ListRoleParams{})
	log.Infof(ctx, "-------- initRoles done, created %d roles\n", count)

	return nil
}

// checkPermissionsInitialized checks if permissions are already initialized.
func (svc *Service) checkPermissionsInitialized(ctx context.Context) error {
	params := &structs.ListPermissionParams{}
	count := svc.permission.CountX(ctx, params)
	if count == 0 {
		return svc.initPermissions(ctx)
	}

	return nil
}

// initPermissions initializes permissions and their relationships.
func (svc *Service) initPermissions(ctx context.Context) error {
	permissions := []structs.CreatePermissionBody{
		{
			PermissionBody: structs.PermissionBody{
				Name:        "all_permissions",
				Action:      "*",
				Subject:     "*",
				Description: "All permissions",
			},
		},
		{
			PermissionBody: structs.PermissionBody{
				Name:        "read",
				Action:      "GET",
				Subject:     "*",
				Description: "Read permissions",
			},
		},
		{
			PermissionBody: structs.PermissionBody{
				Name:        "write",
				Action:      "POST",
				Subject:     "*",
				Description: "Write permissions",
			},
		},
	}

	for _, permission := range permissions {
		if _, err := svc.permission.Create(ctx, &permission); err != nil {
			log.Errorf(ctx, "initPermissions error on create permission: %v\n", err)
			return err
		}
	}

	allPermissions, err := svc.permission.List(ctx, &structs.ListPermissionParams{})
	if err != nil {
		log.Errorf(ctx, "initPermissions error on list permissions: %v\n", err)
		return err
	}

	roles, err := svc.role.List(ctx, &structs.ListRoleParams{})
	if err != nil {
		log.Errorf(ctx, "initPermissions error on list roles: %v\n", err)
		return err
	}

	for _, role := range roles {
		for _, perm := range allPermissions {
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
				rolePerm := structs.RolePermission{
					RoleID:       role.ID,
					PermissionID: perm.ID,
				}
				if _, err := svc.rolePermission.Create(ctx, &rolePerm); err != nil {
					log.Errorf(ctx, "initPermissions error on create role-permission: %v\n", err)
					return err
				}
			}
		}
	}

	count := svc.permission.CountX(ctx, &structs.ListPermissionParams{})
	log.Infof(ctx, "-------- initPermissions done, created %d permissions\n", count)

	return nil
}

// checkCasbinPoliciesInitialized checks if Casbin policies are already initialized.
func (svc *Service) checkCasbinPoliciesInitialized(ctx context.Context) error {
	params := &structs.ListCasbinRuleParams{}
	count := svc.casbinRule.CountX(ctx, params)
	if count == 0 {
		return svc.initCasbinPolicies(ctx)
	}

	return nil
}

// initCasbinPolicies initializes Casbin policies.
func (svc *Service) initCasbinPolicies(ctx context.Context) error {
	defaultTenant, err := svc.tenant.GetBySlug(ctx, "ncobase")
	if err != nil {
		log.Errorf(ctx, "initCasbinPolicies error on get default tenant: %v\n", err)
		return err
	}
	roles, err := svc.role.List(ctx, &structs.ListRoleParams{})
	if err != nil {
		log.Errorf(ctx, "initCasbinPolicies error on list roles: %v\n", err)
		return err
	}

	// Initialize policies based on role-permission relationship
	for _, role := range roles {
		rolePermissions, err := svc.rolePermission.GetPermissionsByRoleID(ctx, role.ID)
		if err != nil {
			log.Errorf(ctx, "initCasbinPolicies error on list role permissions for role %s: %v\n", role.Slug, err)
			return err
		}

		for _, rp := range rolePermissions {
			permission, err := svc.permission.GetByID(ctx, rp.ID)
			if err != nil {
				log.Errorf(ctx, "initCasbinPolicies error on get permission %s: %v\n", rp.ID, err)
				return err
			}

			policy := structs.CasbinRuleBody{
				PType: "p",
				V0:    role.Slug,                          // sub
				V1:    defaultTenant.ID,                   // dom
				V2:    permission.Subject,                 // obj
				V3:    types.ToPointer(permission.Action), // act
				// V4, V5 are not used
			}

			if _, err := svc.casbinRule.Create(ctx, &policy); err != nil {
				log.Errorf(ctx, "initCasbinPolicies error on create casbin rule: %v\n", err)
				return err
			}
		}
	}

	count := svc.casbinRule.CountX(ctx, &structs.ListCasbinRuleParams{})
	log.Infof(ctx, "-------- initCasbinPolicies done, created %d policies\n", count)

	return nil
}
