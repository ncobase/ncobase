package repository

import (
	"ncobase/auth/data"
)

// Repository represents the auth repository.
type Repository struct {
	Captcha CaptchaRepositoryInterface
	Session SessionRepositoryInterface
}

// New creates a new repository.
func New(d *data.Data) *Repository {
	return &Repository{
		Captcha: NewCaptchaRepository(d),
		Session: NewSessionRepository(d),
	}
}
