package initialize

import (
	"context"
	"ncobase/common/logger"
	accessStructs "ncobase/core/access/structs"
)

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
				Name:        "All",
				Action:      "*",
				Subject:     "*",
				Description: "Full access to all resources",
			},
		},
		{
			PermissionBody: accessStructs.PermissionBody{
				Name:        "Read all",
				Action:      "GET",
				Subject:     "*",
				Description: "Read access to all resources",
			},
		},
		{
			PermissionBody: accessStructs.PermissionBody{
				Name:        "Write all",
				Action:      "POST",
				Subject:     "*",
				Description: "Write access to all resources",
			},
		},
		{
			PermissionBody: accessStructs.PermissionBody{
				Name:        "Update all",
				Action:      "PUT",
				Subject:     "*",
				Description: "Update access to all resources",
			},
		},
		{
			PermissionBody: accessStructs.PermissionBody{
				Name:        "Delete all",
				Action:      "DELETE",
				Subject:     "*",
				Description: "Delete access to all resources",
			},
		},
	}

	for _, permission := range permissions {
		if _, err := s.acs.Permission.Create(ctx, &permission); err != nil {
			logger.Errorf(ctx, "initPermissions error on create permission: %v", err)
			return err
		}
	}

	allPermissions, err := s.acs.Permission.List(ctx, &accessStructs.ListPermissionParams{})
	if err != nil {
		logger.Errorf(ctx, "initPermissions error on list permissions: %v", err)
		return err
	}

	roles, err := s.acs.Role.List(ctx, &accessStructs.ListRoleParams{})
	if err != nil {
		logger.Errorf(ctx, "initPermissions error on list roles: %v", err)
		return err
	}

	for _, role := range roles.Items {
		for _, perm := range allPermissions.Items {
			var assignPermission bool
			switch role.Slug {
			case "super-admin":
				// Super Admin gets all permissions
				if perm.Action == "*" && perm.Subject == "*" {
					assignPermission = true
				}
			case "admin":
				// Admin gets read and write permissions
				if perm.Action == "GET" || perm.Action == "POST" || perm.Action == "PUT" || perm.Action == "DELETE" {
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
					logger.Errorf(ctx, "initPermissions error on create role-permission: %v", err)
					return err
				}
			}
		}
	}

	count := s.acs.Permission.CountX(ctx, &accessStructs.ListPermissionParams{})
	logger.Debugf(ctx, "-------- initPermissions done, created %d permissions", count)

	return nil
}
