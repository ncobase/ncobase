package service

import (
	"context"
	"fmt"
	accessStructs "ncobase/access/structs"

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

// initCasbinPolicies initializes Casbin policies based on role-permission relationships
func (s *Service) initCasbinPolicies(ctx context.Context) error {
	logger.Infof(ctx, "Initializing Casbin policies...")

	// Get default tenant for domain-based policies
	defaultTenant, err := s.ts.Tenant.GetBySlug(ctx, "digital-enterprise")
	if err != nil {
		return fmt.Errorf("default tenant 'digital-enterprise' not found: %w", err)
	}

	// Get all roles
	allRoles, err := s.acs.Role.List(ctx, &accessStructs.ListRoleParams{})
	if err != nil {
		return fmt.Errorf("failed to list roles: %w", err)
	}

	if len(allRoles.Items) == 0 {
		return fmt.Errorf("no roles found for Casbin policy initialization")
	}

	var createdPolicies int
	for _, role := range allRoles.Items {
		policyCount, err := s.createPoliciesForRole(ctx, role, defaultTenant.ID)
		if err != nil {
			logger.Errorf(ctx, "Failed to create policies for role '%s': %v", role.Slug, err)
			continue
		}
		createdPolicies += policyCount
	}

	if createdPolicies == 0 {
		logger.Warnf(ctx, "No Casbin policies were created - roles may not have permissions assigned")
	}

	finalCount := s.acs.Casbin.CountX(ctx, &accessStructs.ListCasbinRuleParams{})
	logger.Infof(ctx, "Casbin policy initialization completed, total %d policies", finalCount)
	return nil
}

// createPoliciesForRole creates Casbin policies for a specific role
func (s *Service) createPoliciesForRole(ctx context.Context, role *accessStructs.ReadRole, tenantID string) (int, error) {
	// Get permissions for this role
	rolePermissions, err := s.acs.RolePermission.GetRolePermissions(ctx, role.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to get permissions for role '%s': %w", role.Slug, err)
	}

	if len(rolePermissions) == 0 {
		logger.Debugf(ctx, "Role '%s' has no permissions, skipping policy creation", role.Slug)
		return 0, nil
	}

	var policyCount int
	for _, permission := range rolePermissions {
		// Create Casbin rule: p, role, tenant, resource, action
		policy := &accessStructs.CasbinRuleBody{
			PType: "p",
			V0:    role.Slug,                            // subject (role)
			V1:    tenantID,                             // domain (tenant)
			V2:    permission.Subject,                   // object (resource)
			V3:    convert.ToPointer(permission.Action), // action
			// V4, V5 unused for basic RBAC
		}

		// Check if policy already exists
		if exists, err := s.casbinPolicyExists(ctx, policy); err != nil {
			logger.Warnf(ctx, "Failed to check if policy exists: %v", err)
		} else if exists {
			logger.Debugf(ctx, "Policy already exists for role '%s', permission '%s'", role.Slug, permission.Name)
			continue
		}

		// Create the policy
		if _, err := s.acs.Casbin.Create(ctx, policy); err != nil {
			logger.Errorf(ctx, "Failed to create policy for role '%s', permission '%s': %v",
				role.Slug, permission.Name, err)
			continue
		}

		logger.Debugf(ctx, "Created Casbin policy: role=%s, resource=%s, action=%s",
			role.Slug, permission.Subject, permission.Action)
		policyCount++
	}

	if policyCount > 0 {
		logger.Infof(ctx, "Created %d policies for role '%s'", policyCount, role.Slug)
	}

	return policyCount, nil
}

// casbinPolicyExists checks if a Casbin policy already exists
func (s *Service) casbinPolicyExists(ctx context.Context, policy *accessStructs.CasbinRuleBody) (bool, error) {
	// Query for existing policies with same parameters
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
