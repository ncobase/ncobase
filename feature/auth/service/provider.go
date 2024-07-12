package service

import (
	"ncobase/feature/auth/data"
	tenantService "ncobase/feature/tenant/service"
	userService "ncobase/feature/user/service"
)

// Service represents the auth service.
type Service struct {
	Auth     AuthServiceInterface
	CodeAuth CodeAuthServiceInterface
	// OAuth    OAuthServiceInterface
	Captcha CaptchaServiceInterface
}

// New creates a new service.
func New(d *data.Data, us *userService.Service, ts *tenantService.Service) *Service {
	cas := NewCodeAuthService(d, us)
	return &Service{
		Auth:     NewAuthService(d, cas, us, ts),
		CodeAuth: cas,
		Captcha:  NewCaptchaService(d),
		// OAuth:    NewOAuthService(d),
	}
}
