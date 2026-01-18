package service

import (
	"ncobase/core/user/data"
	"ncobase/core/user/data/repository"
	"ncobase/core/user/event"

	ext "github.com/ncobase/ncore/extension/types"
)

// Service represents the user service.
type Service struct {
	User        UserServiceInterface
	UserProfile UserProfileServiceInterface
	Employee    EmployeeServiceInterface
	ApiKey      ApiKeyServiceInterface
	UserMeshes  UserMeshesServiceInterface
	Events      event.PublisherInterface
}

// New creates a new service.
func New(em ext.ManagerInterface, d *data.Data) *Service {
	ep := event.NewPublisher(em)
	repo := repository.New(d)

	userService := NewUserService(repo, ep)
	userProfileService := NewUserProfileService(repo, ep)
	employeeService := NewEmployeeService(repo, ep)
	apiKeyService := NewApiKeyService(repo, ep)

	userMeshesService := NewUserMeshesService(userService, userProfileService, employeeService, apiKeyService)

	return &Service{
		User:        userService,
		UserProfile: userProfileService,
		Employee:    employeeService,
		ApiKey:      apiKeyService,
		UserMeshes:  userMeshesService,
		Events:      ep,
	}
}
