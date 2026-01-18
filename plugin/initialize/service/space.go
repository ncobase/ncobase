package service

import (
	"context"
	"fmt"
	spaceStructs "ncobase/core/space/structs"

	"github.com/ncobase/ncore/logging/logger"
)

// checkSpacesInitialized checks if spaces are already initialized
func (s *Service) checkSpacesInitialized(ctx context.Context) error {
	defaultSlug := s.getDefaultSpaceSlug()

	space, err := s.ts.Space.GetBySlug(ctx, defaultSlug)
	if err == nil && space != nil {
		logger.Infof(ctx, "Default space already exists, skipping initialization")
		return nil
	}

	params := &spaceStructs.ListSpaceParams{}
	count := s.ts.Space.CountX(ctx, params)
	if count > 0 {
		logger.Infof(ctx, "Spaces already exist, skipping initialization")
		return nil
	}

	return s.initSpaces(ctx)
}

// initSpaces initializes spaces using current data mode
func (s *Service) initSpaces(ctx context.Context) error {
	logger.Infof(ctx, "Initializing system spaces in %s mode...", s.state.DataMode)

	dataLoader := s.getDataLoader()
	spaces := dataLoader.GetSpaces()

	var createdCount int
	var spaceMap = make(map[string]string) // slug -> id mapping

	// Step 1: Create spaces
	for _, space := range spaces {
		existing, err := s.ts.Space.GetBySlug(ctx, space.Slug)
		if err == nil && existing != nil {
			logger.Infof(ctx, "Space %s already exists, skipping", space.Slug)
			spaceMap[space.Slug] = existing.ID
			continue
		}

		created, err := s.ts.Space.Create(ctx, &space)
		if err != nil {
			logger.Errorf(ctx, "Error creating space %s: %v", space.Name, err)
			return fmt.Errorf("failed to create space '%s': %w", space.Name, err)
		}
		logger.Debugf(ctx, "Created space: %s", space.Name)
		spaceMap[space.Slug] = created.ID
		createdCount++
	}

	// Step 2: Initialize space quotas
	if err := s.initSpaceQuotas(ctx, dataLoader, spaceMap); err != nil {
		return fmt.Errorf("failed to initialize space quotas: %w", err)
	}

	// Step 3: Initialize space settings
	if err := s.initSpaceSettings(ctx, dataLoader, spaceMap); err != nil {
		return fmt.Errorf("failed to initialize space settings: %w", err)
	}

	if createdCount == 0 {
		logger.Warnf(ctx, "No spaces were created during initialization")
	}

	// Verify default space exists
	defaultSlug := s.getDefaultSpaceSlug()
	defaultSpace, err := s.ts.Space.GetBySlug(ctx, defaultSlug)
	if err != nil || defaultSpace == nil {
		logger.Errorf(ctx, "Default space '%s' does not exist after initialization", defaultSlug)
		return fmt.Errorf("default space '%s' not found after initialization: %w", defaultSlug, err)
	}

	count := s.ts.Space.CountX(ctx, &spaceStructs.ListSpaceParams{})
	logger.Infof(ctx, "Space initialization completed in %s mode, total %d spaces", s.state.DataMode, count)
	return nil
}

// initSpaceQuotas initializes space quotas for all spaces
func (s *Service) initSpaceQuotas(ctx context.Context, dataLoader DataLoader, spaceMap map[string]string) error {
	logger.Infof(ctx, "Initializing space quotas...")

	quotaTemplates := dataLoader.GetSpaceQuotas()
	if len(quotaTemplates) == 0 {
		logger.Warnf(ctx, "No quota templates found, skipping quota initialization")
		return nil
	}

	var createdCount int
	for spaceSlug, spaceID := range spaceMap {
		for _, quotaTemplate := range quotaTemplates {
			// Check if quota already exists
			existing, err := s.ts.SpaceQuota.GetBySpaceAndType(ctx, spaceID, quotaTemplate.QuotaType)
			if err == nil && existing != nil {
				logger.Debugf(ctx, "Quota %s already exists for space %s, skipping", quotaTemplate.QuotaType, spaceSlug)
				continue
			}

			// Create quota for this space
			quota := quotaTemplate
			quota.SpaceID = spaceID

			if _, err := s.ts.SpaceQuota.Create(ctx, &quota); err != nil {
				logger.Errorf(ctx, "Error creating quota %s for space %s: %v", quota.QuotaType, spaceSlug, err)
				return fmt.Errorf("failed to create quota '%s' for space '%s': %w", quota.QuotaType, spaceSlug, err)
			}
			logger.Debugf(ctx, "Created quota %s for space %s", quota.QuotaType, spaceSlug)
			createdCount++
		}
	}

	logger.Infof(ctx, "Created %d space quotas", createdCount)
	return nil
}

// initSpaceSettings initializes space settings for all spaces
func (s *Service) initSpaceSettings(ctx context.Context, dataLoader DataLoader, spaceMap map[string]string) error {
	logger.Infof(ctx, "Initializing space settings...")

	settingTemplates := dataLoader.GetSpaceSettings()
	if len(settingTemplates) == 0 {
		logger.Warnf(ctx, "No setting templates found, skipping settings initialization")
		return nil
	}

	var createdCount int
	for spaceSlug, spaceID := range spaceMap {
		for _, settingTemplate := range settingTemplates {
			// Check if setting already exists
			existing, err := s.ts.SpaceSetting.GetByKey(ctx, spaceID, settingTemplate.SettingKey)
			if err == nil && existing != nil {
				logger.Debugf(ctx, "Setting %s already exists for space %s, skipping", settingTemplate.SettingKey, spaceSlug)
				continue
			}

			// Create setting for this space
			setting := settingTemplate
			setting.SpaceID = spaceID

			if _, err := s.ts.SpaceSetting.Create(ctx, &setting); err != nil {
				logger.Errorf(ctx, "Error creating setting %s for space %s: %v", setting.SettingKey, spaceSlug, err)
				return fmt.Errorf("failed to create setting '%s' for space '%s': %w", setting.SettingKey, spaceSlug, err)
			}
			logger.Debugf(ctx, "Created setting %s for space %s", setting.SettingKey, spaceSlug)
			createdCount++
		}
	}

	logger.Infof(ctx, "Created %d space settings", createdCount)
	return nil
}
