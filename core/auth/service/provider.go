package service

import (
	"ncobase/auth/data"
	"ncobase/auth/event"
	"ncobase/auth/wrapper"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/security/jwt"
)

// Service represents the auth service.
type Service struct {
	Account  AccountServiceInterface
	CodeAuth CodeAuthServiceInterface
	// OAuth    OAuthServiceInterface
	Captcha    CaptchaServiceInterface
	AuthTenant AuthTenantServiceInterface

	usw *wrapper.UserServiceWrapper
	tsw *wrapper.TenantServiceWrapper
	asw *wrapper.AccessServiceWrapper
}

// New creates a new service.
func New(d *data.Data, jtm *jwt.TokenManager, em ext.ManagerInterface) *Service {
	ep := event.NewPublisher(em)

	usw := wrapper.NewUserServiceWrapper(em)
	tsw := wrapper.NewTenantServiceWrapper(em)
	asw := wrapper.NewAccessServiceWrapper(em)

	cas := NewCodeAuthService(d, jtm, ep, usw, tsw, asw)
	ats := NewAuthTenantService(d, usw, tsw, asw)
	return &Service{
		Account:    NewAccountService(d, jtm, ep, cas, ats, usw, tsw, asw),
		AuthTenant: ats,
		CodeAuth:   cas,
		Captcha:    NewCaptchaService(d),
		// OAuth:    NewOAuthService(d),
		usw: usw,
		tsw: tsw,
		asw: asw,
	}
}

// RefreshDependencies refreshes external service dependencies
func (s *Service) RefreshDependencies() {
	s.usw.RefreshServices()
	s.tsw.RefreshServices()
	s.asw.RefreshServices()
}
