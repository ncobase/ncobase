package service

import (
	"context"
	"ncobase/core/space/data"
	"ncobase/core/space/data/repository"
	"ncobase/core/space/structs"
)

// SpaceDictionaryServiceInterface is the interface for the service.
type SpaceDictionaryServiceInterface interface {
	AddDictionaryToSpace(ctx context.Context, spaceID, dictionaryID string) (*structs.SpaceDictionary, error)
	RemoveDictionaryFromSpace(ctx context.Context, spaceID, dictionaryID string) error
	IsDictionaryInSpace(ctx context.Context, spaceID, dictionaryID string) (bool, error)
	GetSpaceDictionaries(ctx context.Context, spaceID string) ([]string, error)
	GetDictionarySpaces(ctx context.Context, dictionaryID string) ([]string, error)
	RemoveAllDictionariesFromSpace(ctx context.Context, spaceID string) error
	RemoveDictionaryFromAllSpaces(ctx context.Context, dictionaryID string) error
}

// spaceDictionaryService is the struct for the service.
type spaceDictionaryService struct {
	spaceDictionary repository.SpaceDictionaryRepositoryInterface
}

// NewSpaceDictionaryService creates a new service.
func NewSpaceDictionaryService(d *data.Data) SpaceDictionaryServiceInterface {
	return &spaceDictionaryService{
		spaceDictionary: repository.NewSpaceDictionaryRepository(d),
	}
}

// AddDictionaryToSpace adds a dictionary to a space.
func (s *spaceDictionaryService) AddDictionaryToSpace(ctx context.Context, spaceID, dictionaryID string) (*structs.SpaceDictionary, error) {
	row, err := s.spaceDictionary.Create(ctx, &structs.SpaceDictionary{
		SpaceID:      spaceID,
		DictionaryID: dictionaryID,
	})
	if err := handleEntError(ctx, "SpaceDictionary", err); err != nil {
		return nil, err
	}
	return repository.SerializeSpaceDictionary(row), nil
}

// RemoveDictionaryFromSpace removes a dictionary from a space.
func (s *spaceDictionaryService) RemoveDictionaryFromSpace(ctx context.Context, spaceID, dictionaryID string) error {
	err := s.spaceDictionary.DeleteBySpaceIDAndDictionaryID(ctx, spaceID, dictionaryID)
	if err := handleEntError(ctx, "SpaceDictionary", err); err != nil {
		return err
	}
	return nil
}

// IsDictionaryInSpace checks if a dictionary belongs to a space.
func (s *spaceDictionaryService) IsDictionaryInSpace(ctx context.Context, spaceID, dictionaryID string) (bool, error) {
	exists, err := s.spaceDictionary.IsDictionaryInSpace(ctx, spaceID, dictionaryID)
	if err := handleEntError(ctx, "SpaceDictionary", err); err != nil {
		return false, err
	}
	return exists, nil
}

// GetSpaceDictionaries retrieves all dictionary IDs for a space.
func (s *spaceDictionaryService) GetSpaceDictionaries(ctx context.Context, spaceID string) ([]string, error) {
	dictionaryIDs, err := s.spaceDictionary.GetSpaceDictionaries(ctx, spaceID)
	if err := handleEntError(ctx, "SpaceDictionary", err); err != nil {
		return nil, err
	}
	return dictionaryIDs, nil
}

// GetDictionarySpaces retrieves all space IDs for a dictionary.
func (s *spaceDictionaryService) GetDictionarySpaces(ctx context.Context, dictionaryID string) ([]string, error) {
	spaceDictionaries, err := s.spaceDictionary.GetByDictionaryID(ctx, dictionaryID)
	if err := handleEntError(ctx, "SpaceDictionary", err); err != nil {
		return nil, err
	}

	var spaceIDs []string
	for _, td := range spaceDictionaries {
		spaceIDs = append(spaceIDs, td.SpaceID)
	}

	return spaceIDs, nil
}

// RemoveAllDictionariesFromSpace removes all dictionaries from a space.
func (s *spaceDictionaryService) RemoveAllDictionariesFromSpace(ctx context.Context, spaceID string) error {
	err := s.spaceDictionary.DeleteAllBySpaceID(ctx, spaceID)
	if err := handleEntError(ctx, "SpaceDictionary", err); err != nil {
		return err
	}
	return nil
}

// RemoveDictionaryFromAllSpaces removes a dictionary from all spaces.
func (s *spaceDictionaryService) RemoveDictionaryFromAllSpaces(ctx context.Context, dictionaryID string) error {
	err := s.spaceDictionary.DeleteAllByDictionaryID(ctx, dictionaryID)
	if err := handleEntError(ctx, "SpaceDictionary", err); err != nil {
		return err
	}
	return nil
}
