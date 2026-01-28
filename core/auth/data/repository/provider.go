package repository

import "ncobase/core/auth/data"

// Repository represents all repositories
type Repository struct {
	Captcha  CaptchaRepositoryInterface
	Session  SessionRepositoryInterface
	CodeAuth CodeAuthRepositoryInterface
	UserMFA  UserMFARepositoryInterface
}

// New creates a new repository
func New(d *data.Data) *Repository {
	return &Repository{
		Captcha:  NewCaptchaRepository(d),
		Session:  NewSessionRepository(d),
		CodeAuth: NewCodeAuthRepository(d),
		UserMFA:  NewUserMFARepository(d),
	}
}
