package service

import (
	"ncobase/feature/user/data"
	"ncobase/feature/user/data/repository"
)

// UserProfileServiceInterface is the interface for the service.
type UserProfileServiceInterface interface{}

// userProfileService is the struct for the service.
type userProfileService struct {
	user        UserProfileServiceInterface
	userProfile repository.UserProfileRepositoryInterface
}

// NewUserProfileService creates a new service.
func NewUserProfileService(d *data.Data) UserProfileServiceInterface {
	return &userProfileService{}
}
