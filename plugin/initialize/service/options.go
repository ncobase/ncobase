package service

import (
	"context"
	"ncobase/initialize/data"
	systemStructs "ncobase/system/structs"

	"github.com/ncobase/ncore/logging/logger"
)

// checkOptionsInitialized checks if system options are already initialized.
func (s *Service) checkOptionsInitialized(ctx context.Context) error {
	// Check if system options data already exists
	count := s.sys.Options.CountX(ctx, &systemStructs.ListOptionsParams{})
	if count > 0 {
		logger.Infof(ctx, "System options already exist, skipping initialization")
		return nil
	}

	return s.initOptions(ctx)
}

// initOptions initializes the default system options.
func (s *Service) initOptions(ctx context.Context) error {
	logger.Infof(ctx, "Initializing system options...")

	// Get default tenant
	tenant, err := s.ts.Tenant.GetBySlug(ctx, "ncobase")
	if err != nil {
		logger.Errorf(ctx, "Error getting default tenant: %v", err)
		return err
	}

	var createdCount int
	for _, option := range data.SystemDefaultOptions {
		// Set tenant ID
		option.TenantID = tenant.ID

		// Check if already exists
		existing, err := s.sys.Options.GetByName(ctx, option.Name)
		if err == nil && existing != nil {
			logger.Infof(ctx, "Option %s already exists, skipping", option.Name)
			continue
		}

		// Create system option
		_, err = s.sys.Options.Create(ctx, &option)
		if err != nil {
			logger.Errorf(ctx, "Error creating option %s: %v", option.Name, err)
			return err
		}
		logger.Debugf(ctx, "Created option: %s", option.Name)
		createdCount++
	}

	logger.Infof(ctx, "System options initialization completed, created %d options", createdCount)
	return nil
}
