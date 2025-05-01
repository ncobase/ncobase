package initialize

import (
	"context"
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
	tenants := []tenantStructs.CreateTenantBody{
		{
			TenantBody: tenantStructs.TenantBody{
				Name:      "Ncobase Co, Ltd.",
				Slug:      "ncobase",
				CreatedBy: nil,
			},
		},
	}

	for _, tenant := range tenants {
		existing, err := s.ts.Tenant.GetBySlug(ctx, tenant.Slug)
		if err == nil && existing != nil {
			logger.Infof(ctx, "Tenant %s already exists, skipping", tenant.Slug)
			continue
		}

		if _, err := s.ts.Tenant.Create(ctx, &tenant); err != nil {
			logger.Errorf(ctx, "initTenants error on create domain: %v", err)
			return err
		}
	}

	logger.Debugf(ctx, "-------- initTenants done, created %d domains", len(tenants))
	return nil
}
