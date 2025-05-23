package service

import (
	"context"
	"fmt"
	accessStructs "ncobase/access/structs"
	data "ncobase/initialize/data/company"

	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils/convert"
)

// checkCasbinPoliciesInitialized checks if Casbin policies are already initialized
func (s *Service) checkCasbinPoliciesInitialized(ctx context.Context) error {
	count := s.acs.Casbin.CountX(ctx, &accessStructs.ListCasbinRuleParams{})
	if count > 0 {
		logger.Infof(ctx, "Casbin policies already exist (%d rules), skipping initialization", count)
		return nil
	}

	return s.initCasbinPolicies(ctx)
}

// initCasbinPolicies initializes Casbin policies using both dynamic and static approaches
func (s *Service) initCasbinPolicies(ctx context.Context) error {
	logger.Infof(ctx, "Initializing Casbin policies...")

	// Method 1: Create policies from role-permission relationships (existing approach)
	dynamicCount, err := s.initDynamicPolicies(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to initialize dynamic policies: %v", err)
	}

	// Method 2: Create static policies for route-based access control
	staticCount, err := s.initStaticPolicies(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to initialize static policies: %v", err)
	}

	// Method 3: Create role inheritance rules
	inheritanceCount, err := s.initRoleInheritance(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to initialize role inheritance: %v", err)
	}

	finalCount := s.acs.Casbin.CountX(ctx, &accessStructs.ListCasbinRuleParams{})
	logger.Infof(ctx, "Casbin policy initialization completed: %d dynamic + %d static + %d inheritance = %d total policies",
		dynamicCount, staticCount, inheritanceCount, finalCount)

	return nil
}

// initDynamicPolicies creates policies from role-permission relationships (existing logic)
func (s *Service) initDynamicPolicies(ctx context.Context) (int, error) {
	logger.Infof(ctx, "Initializing dynamic Casbin policies from role-permission relationships...")

	// Get default tenant for domain-based policies
	defaultTenant, err := s.ts.Tenant.GetBySlug(ctx, "digital-enterprise")
	if err != nil {
		logger.Warnf(ctx, "Default tenant 'digital-enterprise' not found, using wildcard domain")
		return s.createDynamicPoliciesWithDomain(ctx, "*")
	}

	return s.createDynamicPoliciesWithDomain(ctx, defaultTenant.ID)
}

// createDynamicPoliciesWithDomain creates dynamic policies for a specific domain
func (s *Service) createDynamicPoliciesWithDomain(ctx context.Context, domain string) (int, error) {
	// Get all roles
	allRoles, err := s.acs.Role.List(ctx, &accessStructs.ListRoleParams{})
	if err != nil {
		return 0, fmt.Errorf("failed to list roles: %w", err)
	}

	if len(allRoles.Items) == 0 {
		logger.Warnf(ctx, "No roles found for dynamic policy creation")
		return 0, nil
	}

	var createdPolicies int
	for _, role := range allRoles.Items {
		policyCount, err := s.createPoliciesForRole(ctx, role, domain)
		if err != nil {
			logger.Errorf(ctx, "Failed to create policies for role '%s': %v", role.Slug, err)
			continue
		}
		createdPolicies += policyCount
	}

	logger.Infof(ctx, "Created %d dynamic policies", createdPolicies)
	return createdPolicies, nil
}

// initStaticPolicies creates static route-based policies
func (s *Service) initStaticPolicies(ctx context.Context) (int, error) {
	logger.Infof(ctx, "Initializing static Casbin policies from predefined rules...")

	var createdPolicies int

	// Create policies from static rules
	for _, rule := range data.CasbinPolicyRules {
		if len(rule) < 4 {
			logger.Warnf(ctx, "Invalid policy rule (insufficient parameters): %v", rule)
			continue
		}

		// Build policy structure for 6-parameter model
		policy := &accessStructs.CasbinRuleBody{
			PType: "p",
			V0:    rule[0],                    // subject (role)
			V1:    rule[1],                    // domain (tenant)
			V2:    rule[2],                    // object (resource)
			V3:    convert.ToPointer(rule[3]), // action
		}

		// Add V4 and V5 if provided
		if len(rule) > 4 && rule[4] != "" {
			policy.V4 = convert.ToPointer(rule[4])
		}
		if len(rule) > 5 && rule[5] != "" {
			policy.V5 = convert.ToPointer(rule[5])
		}

		// Check if policy already exists
		if exists, err := s.casbinPolicyExists(ctx, policy); err != nil {
			logger.Warnf(ctx, "Failed to check if policy exists: %v", err)
		} else if exists {
			logger.Debugf(ctx, "Static policy already exists: %v", rule)
			continue
		}

		// Create the policy
		if _, err := s.acs.Casbin.Create(ctx, policy); err != nil {
			logger.Errorf(ctx, "Failed to create static policy %v: %v", rule, err)
			continue
		}

		logger.Debugf(ctx, "Created static Casbin policy: %v", rule)
		createdPolicies++
	}

	logger.Infof(ctx, "Created %d static policies", createdPolicies)
	return createdPolicies, nil
}

// initRoleInheritance creates role inheritance rules
func (s *Service) initRoleInheritance(ctx context.Context) (int, error) {
	logger.Infof(ctx, "Initializing role inheritance rules...")

	var createdRules int

	for _, rule := range data.RoleInheritanceRules {
		if len(rule) < 3 {
			logger.Warnf(ctx, "Invalid inheritance rule (insufficient parameters): %v", rule)
			continue
		}

		// Build grouping policy structure
		groupingPolicy := &accessStructs.CasbinRuleBody{
			PType: "g",
			V0:    rule[0], // child role
			V1:    rule[1], // parent role
			V2:    rule[2], // domain
		}

		// Check if grouping policy already exists
		if exists, err := s.casbinGroupingPolicyExists(ctx, groupingPolicy); err != nil {
			logger.Warnf(ctx, "Failed to check if grouping policy exists: %v", err)
		} else if exists {
			logger.Debugf(ctx, "Role inheritance already exists: %v", rule)
			continue
		}

		// Create the grouping policy
		if _, err := s.acs.Casbin.Create(ctx, groupingPolicy); err != nil {
			logger.Errorf(ctx, "Failed to create role inheritance %v: %v", rule, err)
			continue
		}

		logger.Debugf(ctx, "Created role inheritance: %v", rule)
		createdRules++
	}

	logger.Infof(ctx, "Created %d role inheritance rules", createdRules)
	return createdRules, nil
}

// createPoliciesForRole creates Casbin policies for a specific role (existing logic)
func (s *Service) createPoliciesForRole(ctx context.Context, role *accessStructs.ReadRole, tenantID string) (int, error) {
	// Get permissions for this role
	rolePermissions, err := s.acs.RolePermission.GetRolePermissions(ctx, role.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to get permissions for role '%s': %w", role.Slug, err)
	}

	if len(rolePermissions) == 0 {
		logger.Debugf(ctx, "Role '%s' has no permissions, skipping dynamic policy creation", role.Slug)
		return 0, nil
	}

	var policyCount int
	for _, permission := range rolePermissions {
		// Skip if permission is disabled
		if convert.ToValue(permission.Disabled) {
			continue
		}

		// Map permission to resource path and action
		resource, action := s.mapPermissionToResourceAction(permission.Subject, permission.Action)
		if resource == "" || action == "" {
			logger.Debugf(ctx, "Could not map permission %s:%s to resource/action", permission.Action, permission.Subject)
			continue
		}

		// Create Casbin rule: p, role, tenant, resource, action
		policy := &accessStructs.CasbinRuleBody{
			PType: "p",
			V0:    role.Slug,                 // subject (role)
			V1:    tenantID,                  // domain (tenant)
			V2:    resource,                  // object (resource path)
			V3:    convert.ToPointer(action), // action (HTTP method)
		}

		// Check if policy already exists
		if exists, err := s.casbinPolicyExists(ctx, policy); err != nil {
			logger.Warnf(ctx, "Failed to check if policy exists: %v", err)
		} else if exists {
			logger.Debugf(ctx, "Dynamic policy already exists for role '%s', permission '%s'", role.Slug, permission.Name)
			continue
		}

		// Create the policy
		if _, err := s.acs.Casbin.Create(ctx, policy); err != nil {
			logger.Errorf(ctx, "Failed to create dynamic policy for role '%s', permission '%s': %v",
				role.Slug, permission.Name, err)
			continue
		}

		logger.Debugf(ctx, "Created dynamic Casbin policy: role=%s, resource=%s, action=%s",
			role.Slug, resource, action)
		policyCount++
	}

	if policyCount > 0 {
		logger.Infof(ctx, "Created %d dynamic policies for role '%s'", policyCount, role.Slug)
	}

	return policyCount, nil
}

// mapPermissionToResourceAction maps permission subject/action to HTTP resource path and method
func (s *Service) mapPermissionToResourceAction(subject, action string) (string, string) {
	// Map permission subjects to resource paths
	subjectToPath := map[string]string{
		"account":    "/iam/account",
		"user":       "/user/users",
		"employee":   "/user/employees",
		"menu":       "/sys/menus",
		"dictionary": "/sys/dictionaries",
		"system":     "/sys/options",
		"role":       "/access/roles",
		"permission": "/access/permissions",
		"tenant":     "/tenant/tenants",
		"group":      "/space/groups",
		"content":    "/content/topics",
		"taxonomy":   "/content/taxonomies",
		"resource":   "/resources",
		"workflow":   "/workflow/processes",
		"task":       "/workflow/tasks",
		"payment":    "/payment/orders",
		"realtime":   "/realtime/notifications",
		"proxy":      "/proxy/*",
		"counter":    "/counter/*",
	}

	// Map permission actions to HTTP methods
	actionToMethod := map[string]string{
		"read":   "GET",
		"create": "POST",
		"update": "PUT",
		"delete": "DELETE",
		"manage": "*",
		"*":      "*",
	}

	resourcePath := subjectToPath[subject]
	httpMethod := actionToMethod[action]

	// Handle wildcard subjects and actions
	if resourcePath == "" {
		resourcePath = "/" + subject + "/*"
	}
	if httpMethod == "" {
		httpMethod = "*"
	}

	return resourcePath, httpMethod
}

// casbinPolicyExists checks if a Casbin policy already exists (existing logic)
func (s *Service) casbinPolicyExists(ctx context.Context, policy *accessStructs.CasbinRuleBody) (bool, error) {
	params := &accessStructs.ListCasbinRuleParams{
		PType: &policy.PType,
		V0:    &policy.V0,
		V1:    &policy.V1,
		V2:    &policy.V2,
		V3:    policy.V3,
	}

	if policy.V4 != nil {
		params.V4 = policy.V4
	}
	if policy.V5 != nil {
		params.V5 = policy.V5
	}

	existing, err := s.acs.Casbin.List(ctx, params)
	if err != nil {
		return false, err
	}

	return len(existing.Items) > 0, nil
}

// casbinGroupingPolicyExists checks if a Casbin grouping policy already exists
func (s *Service) casbinGroupingPolicyExists(ctx context.Context, policy *accessStructs.CasbinRuleBody) (bool, error) {
	params := &accessStructs.ListCasbinRuleParams{
		PType: &policy.PType,
		V0:    &policy.V0,
		V1:    &policy.V1,
		V2:    &policy.V2,
	}

	if policy.V4 != nil {
		params.V4 = policy.V4
	}
	if policy.V5 != nil {
		params.V5 = policy.V5
	}

	existing, err := s.acs.Casbin.List(ctx, params)
	if err != nil {
		return false, err
	}

	return len(existing.Items) > 0, nil
}
