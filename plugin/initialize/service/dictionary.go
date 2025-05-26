package service

import (
	"context"
	"fmt"
	systemStructs "ncobase/system/structs"

	"github.com/ncobase/ncore/logging/logger"
)

// checkDictionariesInitialized checks if dictionaries are already initialized
func (s *Service) checkDictionariesInitialized(ctx context.Context) error {
	count := s.sys.Dictionary.CountX(ctx, &systemStructs.ListDictionaryParams{})
	if count > 0 {
		logger.Infof(ctx, "Dictionaries already exist, skipping initialization")
		return nil
	}

	return s.initDictionaries(ctx)
}

// initDictionaries initializes the default dictionaries using current data mode
func (s *Service) initDictionaries(ctx context.Context) error {
	logger.Infof(ctx, "Initializing system dictionaries in %s mode...", s.state.DataMode)

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

	adminUser, err := s.getAdminUser(ctx, "dictionary creation")
	if err != nil {
		return err
	}

	if adminUser == nil {
		logger.Errorf(ctx, "initDictionaries error: no admin user found")
		return fmt.Errorf("no suitable admin user found for dictionary creation")
	}

	dataLoader := s.getDataLoader()
	dictionaries := dataLoader.GetDictionaries()

	var createdCount int
	for _, dict := range dictionaries {
		dict.TenantID = tenant.ID
		dict.CreatedBy = &adminUser.ID

		existing, err := s.sys.Dictionary.Get(ctx, &systemStructs.FindDictionary{Dictionary: dict.Slug})
		if err == nil && existing != nil {
			logger.Infof(ctx, "Dictionary %s already exists, skipping", dict.Slug)
			continue
		}

		_, err = s.sys.Dictionary.Create(ctx, &dict)
		if err != nil {
			logger.Errorf(ctx, "Error creating dictionary %s: %v", dict.Name, err)
			return err
		}
		logger.Debugf(ctx, "Created dictionary: %s", dict.Name)
		createdCount++
	}

	logger.Infof(ctx, "Dictionary initialization completed in %s mode, created %d dictionaries using admin user '%s'",
		s.state.DataMode, createdCount, adminUser.Username)
	return nil
}
