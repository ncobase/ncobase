package initialize

import (
	"context"
	tenantStructs "ncobase/core/tenant/structs"
	"ncore/pkg/logger"
)

// checkTenantsInitialized checks if domains are already initialized.
func (s *Service) checkTenantsInitialized(ctx context.Context) error {
	params := &tenantStructs.ListTenantParams{}
	count := s.ts.Tenant.CountX(ctx, params)
	if count == 0 {
		return s.initTenants(ctx)
	}
	return nil
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
		if _, err := s.ts.Tenant.Create(ctx, &tenant); err != nil {
			logger.Errorf(ctx, "initTenants error on create domain: %v", err)
			return err
		}
	}

	logger.Debugf(ctx, "-------- initTenants done, created %d domains", len(tenants))

	return nil
}
