package repository

import (
	"ncobase/feature/auth/data"
)

// Repository represents the auth repository.
type Repository struct {
	Captcha CaptchaRepositoryInterface
}

// New creates a new repository.
func New(d *data.Data) *Repository {
	return &Repository{
		Captcha: NewCaptchaRepository(d),
	}
}
