package repository

import (
	"ncobase/user/data"
)

// Repository represents the user repository.
type Repository struct {
	User        UserRepositoryInterface
	UserProfile UserProfileRepositoryInterface
}

// New creates a new repository.
func New(d *data.Data) *Repository {
	return &Repository{
		User:        NewUserRepository(d),
		UserProfile: NewUserProfileRepository(d),
	}
}
