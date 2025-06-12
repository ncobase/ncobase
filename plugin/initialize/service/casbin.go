package service

import (
	"context"
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

// initCasbinPolicies initializes Casbin policies using current data mode
func (s *Service) initCasbinPolicies(ctx context.Context) error {
	logger.Infof(ctx, "Initializing Casbin policies in %s mode...", s.state.DataMode)

	staticCount, err := s.initStaticPolicies(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to initialize static policies: %v", err)
	}

	inheritanceCount, err := s.initRoleInheritance(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to initialize role inheritance: %v", err)
	}

	finalCount := s.acs.Casbin.CountX(ctx, &accessStructs.ListCasbinRuleParams{})
	logger.Infof(ctx, "Casbin policy initialization completed: %d static + %d inheritance = %d total policies",
		staticCount, inheritanceCount, finalCount)

	return nil
}

// initStaticPolicies creates static route-based policies using data loader
func (s *Service) initStaticPolicies(ctx context.Context) (int, error) {
	logger.Infof(ctx, "Initializing static Casbin policies from predefined rules...")

	var createdPolicies int

	dataLoader := s.getDataLoader()
	policyRules := dataLoader.GetCasbinPolicyRules()

	for _, rule := range policyRules {
		if len(rule) < 4 {
			logger.Warnf(ctx, "Invalid policy rule (insufficient parameters): %v", rule)
			continue
		}

		policy := &accessStructs.CasbinRuleBody{
			PType: "p",
			V0:    rule[0],                    // subject (role)
			V1:    rule[1],                    // domain (space)
			V2:    rule[2],                    // object (resource)
			V3:    convert.ToPointer(rule[3]), // action
		}

		if len(rule) > 4 && rule[4] != "" {
			policy.V4 = convert.ToPointer(rule[4])
		}
		if len(rule) > 5 && rule[5] != "" {
			policy.V5 = convert.ToPointer(rule[5])
		}

		if exists, err := s.casbinPolicyExists(ctx, policy); err != nil {
			logger.Warnf(ctx, "Failed to check if policy exists: %v", err)
		} else if exists {
			logger.Debugf(ctx, "Static policy already exists: %v", rule)
			continue
		}

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

// initRoleInheritance creates role inheritance rules using data loader
func (s *Service) initRoleInheritance(ctx context.Context) (int, error) {
	logger.Infof(ctx, "Initializing role inheritance rules...")

	var createdRules int

	dataLoader := s.getDataLoader()
	inheritanceRules := dataLoader.GetRoleInheritanceRules()

	for _, rule := range inheritanceRules {
		if len(rule) < 3 {
			logger.Warnf(ctx, "Invalid inheritance rule (insufficient parameters): %v", rule)
			continue
		}

		organizationingPolicy := &accessStructs.CasbinRuleBody{
			PType: "g",
			V0:    rule[0], // child role
			V1:    rule[1], // parent role
			V2:    rule[2], // domain
		}

		if exists, err := s.casbinGroupingPolicyExists(ctx, organizationingPolicy); err != nil {
			logger.Warnf(ctx, "Failed to check if organizationing policy exists: %v", err)
		} else if exists {
			logger.Debugf(ctx, "Role inheritance already exists: %v", rule)
			continue
		}

		if _, err := s.acs.Casbin.Create(ctx, organizationingPolicy); err != nil {
			logger.Errorf(ctx, "Failed to create role inheritance %v: %v", rule, err)
			continue
		}

		logger.Debugf(ctx, "Created role inheritance: %v", rule)
		createdRules++
	}

	logger.Infof(ctx, "Created %d role inheritance rules", createdRules)
	return createdRules, nil
}

// casbinPolicyExists checks if a Casbin policy already exists
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

// casbinGroupingPolicyExists checks if a Casbin organizationing policy already exists
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
