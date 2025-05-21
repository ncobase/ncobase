package service

import (
	"context"
	"fmt"
	accessStructs "ncobase/access/structs"
	"ncobase/initialize/data"

	"github.com/ncobase/ncore/logging/logger"
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

// initPermissions initializes permissions from configuration.
func (s *Service) initPermissions(ctx context.Context) error {
	logger.Infof(ctx, "Initializing system permissions...")

	// Create standard system permissions
	for _, permission := range data.SystemDefaultPermissions {
		if _, err := s.acs.Permission.Create(ctx, &permission); err != nil {
			logger.Errorf(ctx, "Error creating permission %s: %v", permission.Name, err)
			return fmt.Errorf("failed to create permission '%s': %w", permission.Name, err)
		}
		logger.Debugf(ctx, "Created permission: %s", permission.Name)
	}

	// Create organization-specific permissions
	for _, permission := range data.OrganizationPermissions {
		createPermission := accessStructs.CreatePermissionBody{
			PermissionBody: permission,
		}
		if _, err := s.acs.Permission.Create(ctx, &createPermission); err != nil {
			logger.Errorf(ctx, "Error creating organization permission %s: %v", permission.Name, err)
			return fmt.Errorf("failed to create organization permission '%s': %w", permission.Name, err)
		}
		logger.Debugf(ctx, "Created organization permission: %s", permission.Name)
	}

	// Assign permissions to roles based on mapping
	allPermissions, err := s.acs.Permission.List(ctx, &accessStructs.ListPermissionParams{})
	if err != nil {
		logger.Errorf(ctx, "Error listing permissions: %v", err)
		return fmt.Errorf("failed to list permissions for role assignment: %w", err)
	}

	roles, err := s.acs.Role.List(ctx, &accessStructs.ListRoleParams{})
	if err != nil {
		logger.Errorf(ctx, "Error listing roles: %v", err)
		return fmt.Errorf("failed to list roles for permission assignment: %w", err)
	}

	// Map permissions to roles
	for _, role := range roles.Items {
		// Get permissions for this role from config
		permissionNames, ok := data.RolePermissionMapping[role.Slug]
		if !ok {
			logger.Debugf(ctx, "No permission mapping configured for role: %s", role.Slug)
			continue
		}

		// Add each configured permission to role
		for _, permName := range permissionNames {
			// Find permission by name
			var permToAdd *accessStructs.ReadPermission
			for _, perm := range allPermissions.Items {
				if perm.Name == permName {
					permToAdd = perm
					break
				}
			}

			if permToAdd == nil {
				logger.Warnf(ctx, "Permission %s not found for role %s", permName, role.Slug)
				continue
			}

			if _, err := s.acs.RolePermission.AddPermissionToRole(ctx, role.ID, permToAdd.ID); err != nil {
				logger.Errorf(ctx, "Error adding permission %s to role %s: %v", permName, role.Slug, err)
				return fmt.Errorf("failed to add permission '%s' to role '%s': %w", permName, role.Slug, err)
			}
			logger.Debugf(ctx, "Added permission %s to role %s", permName, role.Slug)
		}
	}

	count := s.acs.Permission.CountX(ctx, &accessStructs.ListPermissionParams{})
	logger.Infof(ctx, "Permission initialization completed, created %d permissions", count)

	return nil
}
