package service

import (
	accessService "ncobase/access/service"
	"ncobase/auth/data"
	tenantService "ncobase/tenant/service"
	userService "ncobase/user/service"

	"github.com/ncobase/ncore/security/jwt"
)

// Service represents the auth service.
type Service struct {
	Account  AccountServiceInterface
	CodeAuth CodeAuthServiceInterface
	// OAuth    OAuthServiceInterface
	Captcha    CaptchaServiceInterface
	AuthTenant AuthTenantServiceInterface
}

// New creates a new service.
func New(d *data.Data, jtm *jwt.TokenManager, us *userService.Service, as *accessService.Service, ts *tenantService.Service) *Service {
	cas := NewCodeAuthService(d, jtm, as, us, ts)
	ats := NewAuthTenantService(d, us, as, ts)
	return &Service{
		Account:    NewAccountService(d, jtm, cas, ats, us, as, ts),
		AuthTenant: ats,
		CodeAuth:   cas,
		Captcha:    NewCaptchaService(d),
		// OAuth:    NewOAuthService(d),
	}
}
