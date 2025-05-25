package service

import (
	"context"
	"fmt"
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

// initTenants initializes the default tenants using current data mode.
func (s *Service) initTenants(ctx context.Context) error {
	logger.Infof(ctx, "Initializing system tenants in %s mode...", s.state.DataMode)

	dataLoader := s.getDataLoader()
	tenants := dataLoader.GetTenants()

	var createdCount int
	for _, tenant := range tenants {
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

	if createdCount == 0 {
		logger.Warnf(ctx, "No tenants were created during initialization")
	}

	defaultTenant, err := s.ts.Tenant.GetBySlug(ctx, "digital-enterprise")
	if err != nil || defaultTenant == nil {
		logger.Errorf(ctx, "Default tenant 'digital-enterprise' does not exist after initialization")
		return fmt.Errorf("default tenant 'digital-enterprise' not found after initialization: %w", err)
	}

	count := s.ts.Tenant.CountX(ctx, &tenantStructs.ListTenantParams{})
	logger.Infof(ctx, "Tenant initialization completed in %s mode, created %d tenants", s.state.DataMode, count)
	return nil
}
