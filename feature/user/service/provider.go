package service

import (
	accessService "ncobase/feature/access/service"
	"ncobase/feature/user/data"
)

// Service represents the user service.
type Service struct {
	User        UserServiceInterface
	UserProfile UserProfileServiceInterface
}

// New creates a new service.
func New(d *data.Data, as *accessService.Service) *Service {
	return &Service{
		User:        NewUserService(d, as),
		UserProfile: NewUserProfileService(d),
	}
}
