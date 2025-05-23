package service

import (
	"context"
	"fmt"
	"ncobase/initialize/data"
	tenantStructs "ncobase/tenant/structs"

	"github.com/ncobase/ncore/logging/logger"
)

// checkTenantsInitialized checks if tenants are already initialized.
func (s *Service) checkTenantsInitialized(ctx context.Context) error {
	tenant, err := s.ts.Tenant.GetBySlug(ctx, "digital-enterprise")
	if err == nil && tenant != nil {
		logger.Infof(ctx, "Default tenant already exists, skipping initialization")
		return nil
	}

	params := &tenantStructs.ListTenantParams{}
	count := s.ts.Tenant.CountX(ctx, params)
	if count > 0 {
		logger.Infof(ctx, "Tenants already exist, skipping initialization")
		return nil
	}

	return s.initTenants(ctx)
}

// initTenants initializes the default tenants.
func (s *Service) initTenants(ctx context.Context) error {
	logger.Infof(ctx, "Initializing system tenants...")

	var createdCount int
	for _, tenant := range data.SystemDefaultTenants {
		existing, err := s.ts.Tenant.GetBySlug(ctx, tenant.Slug)
		if err == nil && existing != nil {
			logger.Infof(ctx, "Tenant %s already exists, skipping", tenant.Slug)
			continue
		}

		if _, err := s.ts.Tenant.Create(ctx, &tenant); err != nil {
			logger.Errorf(ctx, "Error creating tenant %s: %v", tenant.Name, err)
			return fmt.Errorf("failed to create tenant '%s': %w", tenant.Name, err)
		}
		logger.Debugf(ctx, "Created tenant: %s", tenant.Name)
		createdCount++
	}

	// Verify at least one tenant was created
	if createdCount == 0 {
		logger.Warnf(ctx, "No tenants were created during initialization")
	}

	// Verify the default tenant exists
	defaultTenant, err := s.ts.Tenant.GetBySlug(ctx, "digital-enterprise")
	if err != nil || defaultTenant == nil {
		logger.Errorf(ctx, "Default tenant 'ncobase' does not exist after initialization")
		return fmt.Errorf("default tenant 'ncobase' not found after initialization: %w", err)
	}

	count := s.ts.Tenant.CountX(ctx, &tenantStructs.ListTenantParams{})
	logger.Infof(ctx, "Tenant initialization completed, created %d tenants", count)
	return nil
}
