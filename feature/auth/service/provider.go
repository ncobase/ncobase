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
func New(d *data.Data, usi userService.UserServiceInterface, tsi tenantService.TenantServiceInterface) *Service {
	cas := NewCodeAuthService(d, usi)
	return &Service{
		Auth:     NewAuthService(d, cas, usi, tsi),
		CodeAuth: cas,
		Captcha:  NewCaptchaService(d, usi),
		// OAuth:    NewOAuthService(d),
	}
}
