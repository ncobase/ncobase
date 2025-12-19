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
	Account   AccountServiceInterface
	CodeAuth  CodeAuthServiceInterface
	Captcha   CaptchaServiceInterface
	AuthSpace AuthSpaceServiceInterface
	Session   SessionServiceInterface
	MFA       MFAServiceInterface

	usw  *wrapper.UserServiceWrapper
	tsw  *wrapper.SpaceServiceWrapper
	asw  *wrapper.AccessServiceWrapper
	ugsw *wrapper.OrganizationServiceWrapper
}

// New creates a new service.
func New(d *data.Data, jtm *jwt.TokenManager, em ext.ManagerInterface) *Service {
	ep := event.NewPublisher(em)

	usw := wrapper.NewUserServiceWrapper(em)
	tsw := wrapper.NewSpaceServiceWrapper(em)
	asw := wrapper.NewAccessServiceWrapper(em)
	ugsw := wrapper.NewOrganizationServiceWrapper(em)

	cas := NewCodeAuthService(d, jtm, ep, usw, tsw, asw)
	ats := NewAuthSpaceService(d, usw, tsw, asw)
	ss := NewSessionService(d)
	mfa := NewMFAService(d, jtm, usw, asw, tsw, ss)

	return &Service{
		Account:   NewAccountService(d, jtm, ep, cas, ats, ss, mfa, usw, tsw, asw, ugsw),
		AuthSpace: ats,
		CodeAuth:  cas,
		Captcha:   NewCaptchaService(d),
		Session:   ss,
		MFA:       mfa,
		usw:       usw,
		tsw:       tsw,
		asw:       asw,
		ugsw:      ugsw,
	}
}

// RefreshDependencies refreshes external service dependencies
func (s *Service) RefreshDependencies() {
	s.usw.RefreshServices()
	s.tsw.RefreshServices()
	s.asw.RefreshServices()
	s.ugsw.RefreshServices()
}
