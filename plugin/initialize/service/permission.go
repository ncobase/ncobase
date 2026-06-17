package service

import (
	"context"
	"fmt"
	accessStructs "ncobase/core/access/structs"

	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
)

// checkPermissionsInitialized checks if permissions are already initialized
func (s *Service) checkPermissionsInitialized(ctx context.Context) error {
	count := s.acs.Permission.CountX(ctx, &accessStructs.ListPermissionParams{})
	if count > 0 {
		logger.Infof(ctx, "Permissions already exist, synchronizing default definitions and role assignments")
		if err := s.syncDefaultPermissionDefinitions(ctx); err != nil {
			return err
		}
		return s.verifyExistingPermissionAssignments(ctx)
	}

	return s.initPermissions(ctx)
}

// initPermissions initializes permissions using current data mode
func (s *Service) initPermissions(ctx context.Context) error {
	logger.Infof(ctx, "Initializing system permissions in %s mode...", s.state.DataMode)

	if err := s.syncDefaultPermissionDefinitions(ctx); err != nil {
		return fmt.Errorf("failed to synchronize permissions: %w", err)
	}

	if err := s.assignPermissionsToRoles(ctx); err != nil {
		return fmt.Errorf("failed to assign permissions to roles: %w", err)
	}

	count := s.acs.Permission.CountX(ctx, &accessStructs.ListPermissionParams{})
	logger.Infof(ctx, "Permission initialization completed, total %d permissions", count)
	return nil
}

// syncDefaultPermissionDefinitions creates missing default permissions and repairs
// route-critical fields that are owned by the initialization data contract.
func (s *Service) syncDefaultPermissionDefinitions(ctx context.Context) error {
	var createdCount, updatedCount int

	dataLoader := s.getDataLoader()
	permissions := dataLoader.GetPermissions()

	for _, permission := range permissions {
		existing, err := s.acs.Permission.GetByName(ctx, permission.Name)
		if err == nil && existing != nil {
			updates := defaultPermissionUpdates(existing, permission)
			if len(updates) == 0 {
				logger.Debugf(ctx, "Permission '%s' already matches default definition", permission.Name)
				continue
			}
			if _, err := s.acs.Permission.Update(ctx, existing.ID, updates); err != nil {
				logger.Errorf(ctx, "Error updating permission '%s': %v", permission.Name, err)
				return fmt.Errorf("failed to update permission '%s': %w", permission.Name, err)
			}
			logger.Debugf(ctx, "Updated permission definition: %s", permission.Name)
			updatedCount++
			continue
		}

		if _, err := s.acs.Permission.Create(ctx, &permission); err != nil {
			logger.Errorf(ctx, "Error creating permission '%s': %v", permission.Name, err)
			return fmt.Errorf("failed to create permission '%s': %w", permission.Name, err)
		}
		logger.Debugf(ctx, "Created permission: %s", permission.Name)
		createdCount++
	}

	logger.Infof(ctx, "Permission definition synchronization completed, created %d and updated %d permissions", createdCount, updatedCount)
	return nil
}

func defaultPermissionUpdates(existing *accessStructs.ReadPermission, expected accessStructs.CreatePermissionBody) types.JSON {
	updates := types.JSON{}
	if expected.Name != "" && existing.Name != expected.Name {
		updates["name"] = expected.Name
	}
	if expected.Action != "" && existing.Action != expected.Action {
		updates["action"] = expected.Action
	}
	if expected.Subject != "" && existing.Subject != expected.Subject {
		updates["subject"] = expected.Subject
	}
	if expected.Description != "" && existing.Description != expected.Description {
		updates["description"] = expected.Description
	}
	if expected.Default != nil && (existing.Default == nil || *existing.Default != *expected.Default) {
		updates["default"] = *expected.Default
	}
	if expected.Disabled != nil && (existing.Disabled == nil || *existing.Disabled != *expected.Disabled) {
		updates["disabled"] = *expected.Disabled
	}
	return updates
}

// assignPermissionsToRoles assigns permissions to roles based on mapping
func (s *Service) assignPermissionsToRoles(ctx context.Context) error {
	logger.Infof(ctx, "Assigning permissions to roles...")

	allPermissions, err := s.acs.Permission.List(ctx, &accessStructs.ListPermissionParams{Limit: 10000})
	if err != nil {
		return fmt.Errorf("failed to list permissions: %w", err)
	}

	permissionMap := make(map[string]string)
	for _, perm := range allPermissions.Items {
		permissionMap[perm.Name] = perm.ID
	}

	roles, err := s.acs.Role.List(ctx, &accessStructs.ListRoleParams{Limit: 10000})
	if err != nil {
		return fmt.Errorf("failed to list roles: %w", err)
	}

	dataLoader := s.getDataLoader()
	rolePermissionMapping := dataLoader.GetRolePermissionMapping()

	var totalAssignments int
	for _, role := range roles.Items {
		assignmentCount, err := s.assignPermissionsToRole(ctx, role, permissionMap, rolePermissionMapping)
		if err != nil {
			logger.Errorf(ctx, "Failed to assign permissions to role '%s': %v", role.Slug, err)
			continue
		}
		totalAssignments += assignmentCount
	}

	logger.Infof(ctx, "Permission assignment completed, created %d role-permission assignments", totalAssignments)
	return nil
}

// assignPermissionsToRole assigns permissions to a specific role
func (s *Service) assignPermissionsToRole(ctx context.Context, role *accessStructs.ReadRole, permissionMap map[string]string, rolePermissionMapping map[string][]string) (int, error) {
	permissionNames, exists := rolePermissionMapping[role.Slug]
	if !exists {
		logger.Debugf(ctx, "No permission mapping found for role: %s", role.Slug)
		return 0, nil
	}

	existingPerms, err := s.acs.RolePermission.GetRolePermissions(ctx, role.ID)
	if err != nil {
		logger.Warnf(ctx, "Failed to get existing permissions for role '%s': %v", role.Slug, err)
		existingPerms = []*accessStructs.ReadPermission{}
	}

	existingPermIDs := make(map[string]bool)
	for _, perm := range existingPerms {
		existingPermIDs[perm.ID] = true
	}

	var assignmentCount int
	for _, permName := range permissionNames {
		permID, exists := permissionMap[permName]
		if !exists {
			logger.Warnf(ctx, "Permission '%s' not found for role '%s'", permName, role.Slug)
			continue
		}

		if existingPermIDs[permID] {
			logger.Debugf(ctx, "Permission '%s' already assigned to role '%s'", permName, role.Slug)
			continue
		}

		if _, err := s.acs.RolePermission.AddPermissionToRole(ctx, role.ID, permID); err != nil {
			logger.Errorf(ctx, "Failed to assign permission '%s' to role '%s': %v", permName, role.Slug, err)
			continue
		}

		logger.Debugf(ctx, "Assigned permission '%s' to role '%s'", permName, role.Slug)
		assignmentCount++
	}

	if assignmentCount > 0 {
		logger.Infof(ctx, "Assigned %d permissions to role '%s'", assignmentCount, role.Slug)
	}

	return assignmentCount, nil
}

// verifyExistingPermissionAssignments verifies existing permissions are properly assigned
func (s *Service) verifyExistingPermissionAssignments(ctx context.Context) error {
	logger.Infof(ctx, "Verifying existing permission assignments...")

	roles, err := s.acs.Role.List(ctx, &accessStructs.ListRoleParams{Limit: 10000})
	if err != nil {
		return fmt.Errorf("failed to list roles: %w", err)
	}

	dataLoader := s.getDataLoader()
	rolePermissionMapping := dataLoader.GetRolePermissionMapping()

	var missingAssignments int
	for _, role := range roles.Items {
		expectedPerms, exists := rolePermissionMapping[role.Slug]
		if !exists {
			continue
		}

		actualPerms, err := s.acs.RolePermission.GetRolePermissions(ctx, role.ID)
		if err != nil {
			logger.Warnf(ctx, "Failed to get permissions for role '%s': %v", role.Slug, err)
			continue
		}

		actualPermNames := make(map[string]bool)
		for _, perm := range actualPerms {
			actualPermNames[perm.Name] = true
		}

		for _, expectedPerm := range expectedPerms {
			if !actualPermNames[expectedPerm] {
				logger.Warnf(ctx, "Role '%s' is missing permission '%s'", role.Slug, expectedPerm)
				missingAssignments++
			}
		}
	}

	if missingAssignments > 0 {
		logger.Warnf(ctx, "Found %d missing permission assignments, running assignment process", missingAssignments)
		return s.assignPermissionsToRoles(ctx)
	}

	logger.Infof(ctx, "All permission assignments verified successfully")
	return nil
}
