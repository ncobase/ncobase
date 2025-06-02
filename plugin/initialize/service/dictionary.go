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

// initDictionaries initializes the default dictionaries and creates tenant relationships
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

	var createdCount, relationshipCount int

	// Create dictionaries and establish tenant relationships
	for _, dict := range dictionaries {
		// Step 1: Create dictionary (without tenant_id)
		dict.CreatedBy = &adminUser.ID

		created, err := s.sys.Dictionary.Create(ctx, &dict)
		if err != nil {
			logger.Errorf(ctx, "Error creating dictionary %s: %v", dict.Name, err)
			return err
		}
		logger.Debugf(ctx, "Created dictionary: %s", dict.Name)
		createdCount++

		// Step 2: Create tenant-dictionary relationship
		_, err = s.ts.TenantDictionary.AddDictionaryToTenant(ctx, tenant.ID, created.ID)
		if err != nil {
			logger.Errorf(ctx, "Error linking dictionary %s to tenant %s: %v", created.ID, tenant.ID, err)
			return err
		}
		logger.Debugf(ctx, "Linked dictionary %s to tenant %s", created.ID, tenant.ID)
		relationshipCount++
	}

	logger.Infof(ctx, "Dictionary initialization completed in %s mode, created %d dictionaries and %d relationships using admin user '%s'",
		s.state.DataMode, createdCount, relationshipCount, adminUser.Username)
	return nil
}
