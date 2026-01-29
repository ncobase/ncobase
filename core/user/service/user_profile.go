package service

import (
	"context"
	"errors"
	"ncobase/core/user/data/repository"
	"ncobase/core/user/event"
	"ncobase/core/user/structs"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation/validator"
)

// UserProfileServiceInterface is the interface for the service.
type UserProfileServiceInterface interface {
	Create(ctx context.Context, body *structs.UserProfileBody) (*structs.ReadUserProfile, error)
	Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadUserProfile, error)
	Get(ctx context.Context, id string) (*structs.ReadUserProfile, error)
	Delete(ctx context.Context, id string) error
}

// userProfileService is the struct for the service.
type userProfileService struct {
	userProfile repository.UserProfileRepositoryInterface
	ep          event.PublisherInterface
}

// NewUserProfileService creates a new service.
func NewUserProfileService(repo *repository.Repository, ep event.PublisherInterface) UserProfileServiceInterface {
	return &userProfileService{
		userProfile: repo.UserProfile,
		ep:          ep,
	}
}

// Create creates a new service.
func (s *userProfileService) Create(ctx context.Context, body *structs.UserProfileBody) (*structs.ReadUserProfile, error) {
	row, err := s.userProfile.Create(ctx, body)
	if err != nil {
		return nil, err
	}
	return repository.SerializeUserProfile(row), nil
}

// Update creates a new service.
func (s *userProfileService) Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadUserProfile, error) {
	if validator.IsEmpty(id) {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	// Validate the updates map
	if len(updates) == 0 {
		return nil, errors.New(ecode.FieldIsEmpty("updates fields"))
	}

	row, err := s.userProfile.Update(ctx, id, updates)
	if err := handleEntError(ctx, "UserProfile", err); err != nil {
		return nil, err
	}
	return repository.SerializeUserProfile(row), nil
}

// Get creates a new service.
func (s *userProfileService) Get(ctx context.Context, id string) (*structs.ReadUserProfile, error) {
	row, err := s.userProfile.Get(ctx, id)
	if err := handleEntError(ctx, "UserProfile", err); err != nil {
		return nil, err
	}
	return repository.SerializeUserProfile(row), nil

}

// Delete creates a new service.
func (s *userProfileService) Delete(ctx context.Context, id string) error {
	return s.userProfile.Delete(ctx, id)
}
