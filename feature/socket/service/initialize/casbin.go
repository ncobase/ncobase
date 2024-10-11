package initialize

import (
	"context"
	"ncobase/common/log"
	"ncobase/common/types"
	accessStructs "ncobase/feature/access/structs"
)

// checkCasbinPoliciesInitialized checks if Casbin policies are already initialized.
func (s *InitializeService) checkCasbinPoliciesInitialized(ctx context.Context) error {
	params := &accessStructs.ListCasbinRuleParams{}
	count := s.acs.Casbin.CountX(ctx, params)
	if count == 0 {
		return s.initCasbinPolicies(ctx)
	}

	return nil
}

// initCasbinPolicies initializes Casbin policies.
func (s *InitializeService) initCasbinPolicies(ctx context.Context) error {
	defaultTenant, err := s.ts.Tenant.GetBySlug(ctx, "ncobase")
	if err != nil {
		log.Errorf(ctx, "initCasbinPolicies error on get default tenant: %v", err)
		return err
	}
	allRoles, err := s.acs.Role.List(ctx, &accessStructs.ListRoleParams{})
	if err != nil {
		log.Errorf(ctx, "initCasbinPolicies error on list roles: %v", err)
		return err
	}

	// Initialize policies based on role-permission relationship
	for _, role := range allRoles.Items {
		rolePermissions, err := s.acs.RolePermission.GetRolePermissions(ctx, role.ID)
		if err != nil {
			log.Errorf(ctx, "initCasbinPolicies error on list role permissions for role %s: %v", role.Slug, err)
			return err
		}

		for _, p := range rolePermissions {
			permission, err := s.acs.Permission.GetByID(ctx, p.ID)
			if err != nil {
				log.Errorf(ctx, "initCasbinPolicies error on get permission %s: %v", p.ID, err)
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
				log.Errorf(ctx, "initCasbinPolicies error on create casbin rule: %v", err)
				return err
			}
		}
	}

	count := s.acs.Casbin.CountX(ctx, &accessStructs.ListCasbinRuleParams{})
	log.Infof(ctx, "-------- initCasbinPolicies done, created %d policies", count)

	return nil
}
