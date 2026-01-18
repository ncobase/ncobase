package service

import (
	"context"
	systemStructs "ncobase/core/system/structs"

	"github.com/ncobase/ncore/logging/logger"
)

// checkOptionsInitialized checks if system options exist
func (s *Service) checkOptionsInitialized(ctx context.Context) error {
	count := s.sys.Option.CountX(ctx, &systemStructs.ListOptionParams{})
	if count > 0 {
		logger.Infof(ctx, "System options already exist, skipping initialization")
		return nil
	}

	return s.initOptions(ctx)
}

// initOptions initializes the default system options and creates space relationships
func (s *Service) initOptions(ctx context.Context) error {
	logger.Infof(ctx, "Initializing system options in %s mode...", s.state.DataMode)

	space, err := s.getDefaultSpace(ctx)
	if err != nil {
		return err
	}

	adminUser, err := s.getAdminUser(ctx, "options creation")
	if err != nil {
		return err
	}

	dataLoader := s.getDataLoader()
	options := dataLoader.GetOptions()

	var createdCount, relationshipCount int

	for _, option := range options {
		option.CreatedBy = &adminUser.ID

		created, err := s.sys.Option.Create(ctx, &option)
		if err != nil {
			logger.Errorf(ctx, "Error creating option %s: %v", option.Name, err)
			return err
		}
		logger.Debugf(ctx, "Created option: %s", option.Name)
		createdCount++

		// Create space-options relationship
		_, err = s.ts.SpaceOption.AddOptionsToSpace(ctx, space.ID, created.ID)
		if err != nil {
			logger.Errorf(ctx, "Error linking options %s to space %s: %v", created.ID, space.ID, err)
			return err
		}
		logger.Debugf(ctx, "Linked options %s to space %s", created.ID, space.ID)
		relationshipCount++
	}

	logger.Infof(ctx, "System options initialization completed in %s mode, created %d options and %d relationships",
		s.state.DataMode, createdCount, relationshipCount)
	return nil
}
