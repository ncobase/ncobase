package service

import (
	"context"
	"errors"
	"ncobase/common/ecode"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/core/user/data"
	"ncobase/core/user/data/ent"
	"ncobase/core/user/data/repository"
	"ncobase/core/user/structs"
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
}

// NewUserProfileService creates a new service.
func NewUserProfileService(d *data.Data) UserProfileServiceInterface {
	return &userProfileService{
		userProfile: repository.NewUserProfileRepository(d),
	}
}

// Create creates a new service.
func (s *userProfileService) Create(ctx context.Context, body *structs.UserProfileBody) (*structs.ReadUserProfile, error) {
	row, err := s.userProfile.Create(ctx, body)
	if err != nil {
		return nil, err
	}
	return s.Serialize(row), nil
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
	return s.Serialize(row), nil
}

// Get creates a new service.
func (s *userProfileService) Get(ctx context.Context, id string) (*structs.ReadUserProfile, error) {
	row, err := s.userProfile.Get(ctx, id)
	if err := handleEntError(ctx, "UserProfile", err); err != nil {
		return nil, err
	}
	return s.Serialize(row), nil

}

// Delete creates a new service.
func (s *userProfileService) Delete(ctx context.Context, id string) error {
	return s.userProfile.Delete(ctx, id)
}

// Serialize serialize user profile
func (s *userProfileService) Serialize(row *ent.UserProfile) *structs.ReadUserProfile {
	return &structs.ReadUserProfile{
		DisplayName: row.DisplayName,
		ShortBio:    row.ShortBio,
		About:       &row.About,
		Thumbnail:   &row.Thumbnail,
		Links:       &row.Links,
		Extras:      &row.Extras,
	}
}
