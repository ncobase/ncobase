package service

import (
	"ncobase/user/data"
)

// Service represents the user service.
type Service struct {
	User        UserServiceInterface
	UserProfile UserProfileServiceInterface
	Employee    EmployeeServiceInterface
}

// New creates a new service.
func New(d *data.Data) *Service {
	return &Service{
		User:        NewUserService(d),
		UserProfile: NewUserProfileService(d),
		Employee:    NewEmployeeService(d),
	}
}
