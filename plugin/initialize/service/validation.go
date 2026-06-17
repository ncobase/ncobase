package service

import (
	"context"
	"fmt"
	accessStructs "ncobase/core/access/structs"
	menuData "ncobase/plugin/initialize/data"
	"strings"

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

	// Menus store permission codes while role mappings store permission names.
	// Include both forms here so validation reflects the actual runtime contract.
	permissionKeys := permissionKeysForDefinitions(permissions)

	// Validate menu permissions
	issues := menuData.ValidateMenuPermissions(permissionKeys)

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

func permissionKeysForDefinitions(permissions []accessStructs.CreatePermissionBody) []string {
	keys := make([]string, 0, len(permissions)*4)
	seen := make(map[string]struct{}, len(permissions)*4)
	add := func(key string) {
		key = strings.TrimSpace(key)
		if key == "" {
			return
		}
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}

	for _, perm := range permissions {
		add(perm.Name)
		if perm.Action == "" || perm.Subject == "" {
			continue
		}

		action := strings.TrimSpace(perm.Action)
		subject := strings.TrimSpace(perm.Subject)
		add(fmt.Sprintf("%s:%s", action, subject))

		if singular := singularizePermissionSubject(subject); singular != subject {
			add(fmt.Sprintf("%s:%s", action, singular))
		}
		if plural := pluralizePermissionSubject(subject); plural != subject {
			add(fmt.Sprintf("%s:%s", action, plural))
		}
	}

	return keys
}

func singularizePermissionSubject(subject string) string {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return subject
	}
	if strings.HasSuffix(subject, "ies") && len(subject) > 3 {
		return subject[:len(subject)-3] + "y"
	}
	if strings.HasSuffix(subject, "s") && len(subject) > 1 {
		return subject[:len(subject)-1]
	}
	return subject
}

func pluralizePermissionSubject(subject string) string {
	subject = strings.TrimSpace(subject)
	if subject == "" || strings.HasSuffix(subject, "s") {
		return subject
	}
	if strings.HasSuffix(subject, "y") && len(subject) > 1 {
		prev := subject[len(subject)-2]
		if !strings.ContainsRune("aeiou", rune(prev)) {
			return subject[:len(subject)-1] + "ies"
		}
	}
	return subject + "s"
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
