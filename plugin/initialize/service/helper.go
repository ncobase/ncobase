package service

import (
	"context"
	"fmt"
	tenantStructs "ncobase/tenant/structs"
	userStructs "ncobase/user/structs"

	"github.com/ncobase/ncore/logging/logger"
)

// getDefaultTenantSlug returns default tenant slug based on mode
func (s *Service) getDefaultTenantSlug() string {
	switch s.state.DataMode {
	case "website":
		return "website-platform"
	case "company":
		return "digital-company"
	case "enterprise":
		return "digital-enterprise"
	default:
		return "website-platform"
	}
}

// getDefaultTenant retrieves default tenant based on mode
func (s *Service) getDefaultTenant(ctx context.Context) (*tenantStructs.ReadTenant, error) {
	tenantSlug := s.getDefaultTenantSlug()

	tenant, err := s.ts.Tenant.GetBySlug(ctx, tenantSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to get default tenant '%s': %v", tenantSlug, err)
	}

	return tenant, nil
}

// getAdminUser retrieves admin user based on mode
func (s *Service) getAdminUser(ctx context.Context, operation string) (*userStructs.ReadUser, error) {
	var adminCandidates []string

	switch s.state.DataMode {
	case "website":
		adminCandidates = []string{"admin", "super", "manager"}
	case "company":
		adminCandidates = []string{"company.admin", "super", "admin", "manager"}
	case "enterprise":
		adminCandidates = []string{"enterprise.admin", "super", "admin", "dept.manager"}
	default:
		adminCandidates = []string{"admin", "super"}
	}

	for _, username := range adminCandidates {
		user, err := s.us.User.Get(ctx, username)
		if err == nil && user != nil {
			logger.Debugf(ctx, "Using user '%s' for %s", username, operation)
			return user, nil
		}
	}

	return nil, fmt.Errorf("no suitable admin user found for %s", operation)
}
