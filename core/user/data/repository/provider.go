package repository

import "ncobase/core/user/data"

// Repository represents all repositories
type Repository struct {
	User        UserRepositoryInterface
	UserProfile UserProfileRepositoryInterface
	Employee    EmployeeRepositoryInterface
	ApiKey      ApiKeyRepositoryInterface
}

// New creates a new repository
func New(d *data.Data) *Repository {
	return &Repository{
		User:        NewUserRepository(d),
		UserProfile: NewUserProfileRepository(d),
		Employee:    NewEmployeeRepository(d),
		ApiKey:      NewApiKeyRepository(d),
	}
}
