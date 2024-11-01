package service

import (
	"ncobase/core/user/data"
)

// Service represents the user service.
type Service struct {
	User        UserServiceInterface
	UserProfile UserProfileServiceInterface
}

// New creates a new service.
func New(d *data.Data) *Service {
	return &Service{
		User:        NewUserService(d),
		UserProfile: NewUserProfileService(d),
	}
}
