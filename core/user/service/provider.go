package service

import (
	"ncobase/user/data"
	"ncobase/user/data/repository"
	"ncobase/user/event"

	ext "github.com/ncobase/ncore/extension/types"
)

// Service represents the user service.
type Service struct {
	User        UserServiceInterface
	UserProfile UserProfileServiceInterface
	Employee    EmployeeServiceInterface
	ApiKey      ApiKeyServiceInterface
	Events      event.PublisherInterface
}

// New creates a new service.
func New(em ext.ManagerInterface, d *data.Data) *Service {
	ep := event.NewPublisher(em)
	repo := repository.New(d)
	return &Service{
		User:        NewUserService(repo, ep),
		UserProfile: NewUserProfileService(repo, ep),
		Employee:    NewEmployeeService(repo, ep),
		ApiKey:      NewApiKeyService(repo, ep),
		Events:      ep,
	}
}
