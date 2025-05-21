package service

import (
	"context"
	"ncobase/initialize/data"
	systemStructs "ncobase/system/structs"

	"github.com/ncobase/ncore/logging/logger"
)

// checkDictionariesInitialized checks if dictionaries are already initialized.
func (s *Service) checkDictionariesInitialized(ctx context.Context) error {
	// Check if dictionary data already exists
	count := s.sys.Dictionary.CountX(ctx, &systemStructs.ListDictionaryParams{})
	if count > 0 {
		logger.Infof(ctx, "Dictionaries already exist, skipping initialization")
		return nil
	}

	return s.initDictionaries(ctx)
}

// initDictionaries initializes the default dictionaries.
func (s *Service) initDictionaries(ctx context.Context) error {
	logger.Infof(ctx, "Initializing system dictionaries...")

	// Get default tenant
	tenant, err := s.ts.Tenant.GetBySlug(ctx, "ncobase")
	if err != nil {
		logger.Errorf(ctx, "Error getting default tenant: %v", err)
		return err
	}

	// Get admin user
	admin, err := s.us.User.Get(ctx, "admin")
	if err != nil {
		logger.Errorf(ctx, "Error getting admin user: %v", err)
		return err
	}

	var createdCount int
	for _, dict := range data.SystemDefaultDictionaries {
		// Set tenant ID and creator
		dict.TenantID = tenant.ID
		dict.CreatedBy = &admin.ID

		// Check if already exists
		existing, err := s.sys.Dictionary.Get(ctx, &systemStructs.FindDictionary{Dictionary: dict.Slug})
		if err == nil && existing != nil {
			logger.Infof(ctx, "Dictionary %s already exists, skipping", dict.Slug)
			continue
		}

		// Create dictionary data
		_, err = s.sys.Dictionary.Create(ctx, &dict)
		if err != nil {
			logger.Errorf(ctx, "Error creating dictionary %s: %v", dict.Name, err)
			return err
		}
		logger.Debugf(ctx, "Created dictionary: %s", dict.Name)
		createdCount++
	}

	logger.Infof(ctx, "Dictionary initialization completed, created %d dictionaries", createdCount)
	return nil
}
