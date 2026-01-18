package service

import (
	"context"
	"fmt"
	menuData "ncobase/plugin/initialize/data"

	"github.com/ncobase/ncore/logging/logger"
)

// validateInitializationConsistency validates that permissions, roles, and menus are properly aligned
func (s *Service) validateInitializationConsistency(ctx context.Context) error {
	logger.Infof(ctx, "Validating initialization consistency...")

	// Get current data loader
	dataLoader := s.getDataLoader()

	// Validate permission-menu alignment
	if err := s.validatePermissionMenuAlignment(ctx, dataLoader); err != nil {
		logger.Warnf(ctx, "Permission-menu alignment issues found: %v", err)
	}

	// Validate role-permission mapping
	if err := s.validateRolePermissionMapping(ctx, dataLoader); err != nil {
		logger.Warnf(ctx, "Role-permission mapping issues found: %v", err)
	}

	// Validate Casbin policy alignment
	if err := s.validateCasbinPolicyAlignment(ctx, dataLoader); err != nil {
		logger.Warnf(ctx, "Casbin policy alignment issues found: %v", err)
	}

	logger.Infof(ctx, "Initialization consistency validation completed")
	return nil
}

// validatePermissionMenuAlignment validates that menu permissions exist in permission definitions
func (s *Service) validatePermissionMenuAlignment(ctx context.Context, dataLoader DataLoader) error {
	permissions := dataLoader.GetPermissions()

	// Extract permission names
	permissionNames := make([]string, len(permissions))
	for i, perm := range permissions {
		permissionNames[i] = perm.Name
	}

	// Validate menu permissions
	issues := menuData.ValidateMenuPermissions(permissionNames)

	if len(issues) > 0 {
		logger.Warnf(ctx, "Found %d menu-permission alignment issues:", len(issues))
		for menuSlug, menuIssues := range issues {
			for _, issue := range menuIssues {
				logger.Warnf(ctx, "Menu '%s': %s", menuSlug, issue)
			}
		}
		return fmt.Errorf("found %d menu-permission alignment issues", len(issues))
	}

	logger.Infof(ctx, "Menu-permission alignment validation passed")
	return nil
}

// validateRolePermissionMapping validates that all permissions in role mappings exist
func (s *Service) validateRolePermissionMapping(ctx context.Context, dataLoader DataLoader) error {
	permissions := dataLoader.GetPermissions()
	rolePermissionMapping := dataLoader.GetRolePermissionMapping()

	// Create permission name set
	permissionSet := make(map[string]bool)
	for _, perm := range permissions {
		permissionSet[perm.Name] = true
	}

	var issues []string
	for roleSlug, permissionNames := range rolePermissionMapping {
		for _, permName := range permissionNames {
			if !permissionSet[permName] {
				issues = append(issues, fmt.Sprintf("Role '%s' references undefined permission '%s'", roleSlug, permName))
			}
		}
	}

	if len(issues) > 0 {
		logger.Warnf(ctx, "Found %d role-permission mapping issues:", len(issues))
		for _, issue := range issues {
			logger.Warnf(ctx, "%s", issue)
		}
		return fmt.Errorf("found %d role-permission mapping issues", len(issues))
	}

	logger.Infof(ctx, "Role-permission mapping validation passed")
	return nil
}

// validateCasbinPolicyAlignment validates Casbin policies align with actual routes
func (s *Service) validateCasbinPolicyAlignment(ctx context.Context, dataLoader DataLoader) error {
	roles := dataLoader.GetRoles()
	policyRules := dataLoader.GetCasbinPolicyRules()

	// Create role set
	roleSet := make(map[string]bool)
	for _, role := range roles {
		roleSet[role.Slug] = true
	}

	var issues []string
	for _, rule := range policyRules {
		if len(rule) >= 1 {
			roleSlug := rule[0]
			if roleSlug != "*" && !roleSet[roleSlug] {
				issues = append(issues, fmt.Sprintf("Casbin policy references undefined role '%s'", roleSlug))
			}
		}
	}

	if len(issues) > 0 {
		logger.Warnf(ctx, "Found %d Casbin policy alignment issues:", len(issues))
		for _, issue := range issues {
			logger.Warnf(ctx, "%s", issue)
		}
		return fmt.Errorf("found %d Casbin policy alignment issues", len(issues))
	}

	logger.Infof(ctx, "Casbin policy alignment validation passed")
	return nil
}
