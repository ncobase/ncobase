package service

import (
	"context"
	"fmt"
	systemStructs "ncobase/system/structs"

	"github.com/ncobase/ncore/logging/logger"
)

// checkOptionsInitialized checks if system options are already initialized
func (s *Service) checkOptionsInitialized(ctx context.Context) error {
	count := s.sys.Options.CountX(ctx, &systemStructs.ListOptionParams{})
	if count > 0 {
		logger.Infof(ctx, "System options already exist, skipping initialization")
		return nil
	}

	return s.initOptions(ctx)
}

// initOptions initializes the default system options and creates tenant relationships
func (s *Service) initOptions(ctx context.Context) error {
	logger.Infof(ctx, "Initializing system options in %s mode...", s.state.DataMode)

	// Get default tenant based on data mode
	var tenantSlug string
	switch s.state.DataMode {
	case "website":
		tenantSlug = "website-platform"
	case "company":
		tenantSlug = "digital-company"
	case "enterprise":
		tenantSlug = "digital-enterprise"
	default:
		tenantSlug = "website-platform"
	}

	tenant, err := s.ts.Tenant.GetBySlug(ctx, tenantSlug)
	if err != nil {
		logger.Errorf(ctx, "Error getting default tenant: %v", err)
		return err
	}

	adminUser, err := s.getAdminUser(ctx, "options creation")
	if err != nil {
		return err
	}

	if adminUser == nil {
		logger.Errorf(ctx, "initOptions error: no admin user found")
		return fmt.Errorf("no suitable admin user found for options creation")
	}

	dataLoader := s.getDataLoader()
	options := dataLoader.GetOptions()

	var createdCount, relationshipCount int

	// Create options and establish tenant relationships
	for _, option := range options {
		// Step 1: Create option (without tenant_id)
		option.CreatedBy = &adminUser.ID

		created, err := s.sys.Options.Create(ctx, &option)
		if err != nil {
			logger.Errorf(ctx, "Error creating option %s: %v", option.Name, err)
			return err
		}
		logger.Debugf(ctx, "Created option: %s", option.Name)
		createdCount++

		// Step 2: Create tenant-options relationship
		_, err = s.ts.TenantOption.AddOptionsToTenant(ctx, tenant.ID, created.ID)
		if err != nil {
			logger.Errorf(ctx, "Error linking options %s to tenant %s: %v", created.ID, tenant.ID, err)
			return err
		}
		logger.Debugf(ctx, "Linked options %s to tenant %s", created.ID, tenant.ID)
		relationshipCount++
	}

	logger.Infof(ctx, "System options initialization completed in %s mode, created %d options and %d relationships using admin user '%s'",
		s.state.DataMode, createdCount, relationshipCount, adminUser.Username)
	return nil
}
