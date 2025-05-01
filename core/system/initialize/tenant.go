package initialize

import (
	"context"
	"ncobase/core/system/initialize/data"
	tenantStructs "ncobase/core/tenant/structs"

	"github.com/ncobase/ncore/logging/logger"
)

// checkTenantsInitialized checks if domains are already initialized.
func (s *Service) checkTenantsInitialized(ctx context.Context) error {
	tenant, err := s.ts.Tenant.GetBySlug(ctx, "ncobase")
	if err == nil && tenant != nil {
		logger.Infof(ctx, "Default tenant already exists, skipping tenant initialization")
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

// initTenants initializes the domains (tenants).
func (s *Service) initTenants(ctx context.Context) error {
	logger.Infof(ctx, "Initializing system tenants...")

	for _, tenant := range data.SystemDefaultTenants {
		existing, err := s.ts.Tenant.GetBySlug(ctx, tenant.Slug)
		if err == nil && existing != nil {
			logger.Infof(ctx, "Tenant %s already exists, skipping", tenant.Slug)
			continue
		}

		if _, err := s.ts.Tenant.Create(ctx, &tenant); err != nil {
			logger.Errorf(ctx, "Error creating tenant %s: %v", tenant.Name, err)
			return err
		}
		logger.Debugf(ctx, "Created tenant: %s", tenant.Name)
	}

	count := s.ts.Tenant.CountX(ctx, &tenantStructs.ListTenantParams{})
	logger.Infof(ctx, "Tenant initialization completed, created %d tenants", count)
	return nil
}
