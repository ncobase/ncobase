package service

import (
	"context"
	"ncobase/core/space/data"
	"ncobase/core/space/data/ent"
	"ncobase/core/space/data/repository"
	"ncobase/core/space/structs"
)

// SpaceOptionServiceInterface is the interface for the service.
type SpaceOptionServiceInterface interface {
	AddOptionsToSpace(ctx context.Context, spaceID, optionsID string) (*structs.SpaceOption, error)
	RemoveOptionsFromSpace(ctx context.Context, spaceID, optionsID string) error
	IsOptionsInSpace(ctx context.Context, spaceID, optionsID string) (bool, error)
	GetSpaceOption(ctx context.Context, spaceID string) ([]string, error)
	GetOptionsSpaces(ctx context.Context, optionsID string) ([]string, error)
	RemoveAllOptionsFromSpace(ctx context.Context, spaceID string) error
	RemoveOptionsFromAllSpaces(ctx context.Context, optionsID string) error
}

// spaceOptionService is the struct for the service.
type spaceOptionService struct {
	spaceOption repository.SpaceOptionRepositoryInterface
}

// NewSpaceOptionService creates a new service.
func NewSpaceOptionService(d *data.Data) SpaceOptionServiceInterface {
	return &spaceOptionService{
		spaceOption: repository.NewSpaceOptionRepository(d),
	}
}

// AddOptionsToSpace adds options to a space.
func (s *spaceOptionService) AddOptionsToSpace(ctx context.Context, spaceID, optionsID string) (*structs.SpaceOption, error) {
	row, err := s.spaceOption.Create(ctx, &structs.SpaceOption{
		SpaceID:  spaceID,
		OptionID: optionsID,
	})
	if err := handleEntError(ctx, "SpaceOption", err); err != nil {
		return nil, err
	}
	return s.SerializeSpaceOption(row), nil
}

// RemoveOptionsFromSpace removes options from a space.
func (s *spaceOptionService) RemoveOptionsFromSpace(ctx context.Context, spaceID, optionsID string) error {
	err := s.spaceOption.DeleteBySpaceIDAndOptionID(ctx, spaceID, optionsID)
	if err := handleEntError(ctx, "SpaceOption", err); err != nil {
		return err
	}
	return nil
}

// IsOptionsInSpace checks if options belong to a space.
func (s *spaceOptionService) IsOptionsInSpace(ctx context.Context, spaceID, optionsID string) (bool, error) {
	exists, err := s.spaceOption.IsOptionsInSpace(ctx, spaceID, optionsID)
	if err := handleEntError(ctx, "SpaceOption", err); err != nil {
		return false, err
	}
	return exists, nil
}

// GetSpaceOption retrieves all options IDs for a space.
func (s *spaceOptionService) GetSpaceOption(ctx context.Context, spaceID string) ([]string, error) {
	optionsIDs, err := s.spaceOption.GetSpaceOption(ctx, spaceID)
	if err := handleEntError(ctx, "SpaceOption", err); err != nil {
		return nil, err
	}
	return optionsIDs, nil
}

// GetOptionsSpaces retrieves all space IDs for options.
func (s *spaceOptionService) GetOptionsSpaces(ctx context.Context, optionsID string) ([]string, error) {
	spaceOption, err := s.spaceOption.GetByOptionID(ctx, optionsID)
	if err := handleEntError(ctx, "SpaceOption", err); err != nil {
		return nil, err
	}

	var spaceIDs []string
	for _, to := range spaceOption {
		spaceIDs = append(spaceIDs, to.SpaceID)
	}

	return spaceIDs, nil
}

// RemoveAllOptionsFromSpace removes all options from a space.
func (s *spaceOptionService) RemoveAllOptionsFromSpace(ctx context.Context, spaceID string) error {
	err := s.spaceOption.DeleteAllBySpaceID(ctx, spaceID)
	if err := handleEntError(ctx, "SpaceOption", err); err != nil {
		return err
	}
	return nil
}

// RemoveOptionsFromAllSpaces removes options from all spaces.
func (s *spaceOptionService) RemoveOptionsFromAllSpaces(ctx context.Context, optionsID string) error {
	err := s.spaceOption.DeleteAllByOptionID(ctx, optionsID)
	if err := handleEntError(ctx, "SpaceOption", err); err != nil {
		return err
	}
	return nil
}

// SerializeSpaceOption serializes a space option relationship.
func (s *spaceOptionService) SerializeSpaceOption(row *ent.SpaceOption) *structs.SpaceOption {
	return &structs.SpaceOption{
		SpaceID:  row.SpaceID,
		OptionID: row.OptionID,
	}
}
