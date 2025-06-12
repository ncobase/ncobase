package service

import (
	"context"
	systemStructs "ncobase/system/structs"

	"github.com/ncobase/ncore/logging/logger"
)

// checkDictionariesInitialized checks if dictionaries exist
func (s *Service) checkDictionariesInitialized(ctx context.Context) error {
	count := s.sys.Dictionary.CountX(ctx, &systemStructs.ListDictionaryParams{})
	if count > 0 {
		logger.Infof(ctx, "Dictionaries already exist, skipping initialization")
		return nil
	}

	return s.initDictionaries(ctx)
}

// initDictionaries initializes the default dictionaries and creates space relationships
func (s *Service) initDictionaries(ctx context.Context) error {
	logger.Infof(ctx, "Initializing system dictionaries in %s mode...", s.state.DataMode)

	space, err := s.getDefaultSpace(ctx)
	if err != nil {
		return err
	}

	adminUser, err := s.getAdminUser(ctx, "dictionary creation")
	if err != nil {
		return err
	}

	dataLoader := s.getDataLoader()
	dictionaries := dataLoader.GetDictionaries()

	var createdCount, relationshipCount int

	for _, dict := range dictionaries {
		dict.CreatedBy = &adminUser.ID

		created, err := s.sys.Dictionary.Create(ctx, &dict)
		if err != nil {
			logger.Errorf(ctx, "Error creating dictionary %s: %v", dict.Name, err)
			return err
		}
		logger.Debugf(ctx, "Created dictionary: %s", dict.Name)
		createdCount++

		// Create space-dictionary relationship
		_, err = s.ts.SpaceDictionary.AddDictionaryToSpace(ctx, space.ID, created.ID)
		if err != nil {
			logger.Errorf(ctx, "Error linking dictionary %s to space %s: %v", created.ID, space.ID, err)
			return err
		}
		logger.Debugf(ctx, "Linked dictionary %s to space %s", created.ID, space.ID)
		relationshipCount++
	}

	logger.Infof(ctx, "Dictionary initialization completed in %s mode, created %d dictionaries and %d relationships",
		s.state.DataMode, createdCount, relationshipCount)
	return nil
}
