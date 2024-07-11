package repository

import (
	"ncobase/feature/user/data"
)

// Repository represents the user repository.
type Repository struct {
	User        UserRepositoryInterface
	UserProfile UserProfileRepositoryInterface
}

// NewRepository creates a new repository.
func NewRepository(d *data.Data) *Repository {
	return &Repository{
		User:        NewUserRepository(d),
		UserProfile: NewUserProfileRepository(d),
	}
}
