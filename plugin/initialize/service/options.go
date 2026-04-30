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
		logger.Infof(ctx, "System options already exist, ensuring missing default options")
	}

	return s.initOptions(ctx)
}

// initOptions initializes missing default system options and creates space relationships.
func (s *Service) initOptions(ctx context.Context) error {
	logger.Infof(ctx, "Ensuring system options in %s mode...", s.state.DataMode)

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
		var optionID string
		existing, err := s.sys.Option.GetByName(ctx, option.Name)
		if err == nil && existing != nil {
			optionID = existing.ID
			logger.Debugf(ctx, "Option %s already exists, preserving current value", option.Name)
		} else {
			option.CreatedBy = &adminUser.ID
			created, createErr := s.sys.Option.Create(ctx, &option)
			if createErr != nil {
				logger.Errorf(ctx, "Error creating option %s: %v", option.Name, createErr)
				return createErr
			}
			optionID = created.ID
			logger.Debugf(ctx, "Created option: %s", option.Name)
			createdCount++
		}

		linked, err := s.ts.SpaceOption.IsOptionsInSpace(ctx, space.ID, optionID)
		if err != nil {
			return err
		}
		if linked {
			continue
		}

		_, err = s.ts.SpaceOption.AddOptionsToSpace(ctx, space.ID, optionID)
		if err != nil {
			logger.Errorf(ctx, "Error linking options %s to space %s: %v", optionID, space.ID, err)
			return err
		}
		logger.Debugf(ctx, "Linked options %s to space %s", optionID, space.ID)
		relationshipCount++
	}

	logger.Infof(ctx, "System options ensured in %s mode, created %d options and %d relationships",
		s.state.DataMode, createdCount, relationshipCount)
	return nil
}
