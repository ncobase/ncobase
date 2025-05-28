package handler

import (
	"ncobase/auth/service"
)

// Handler represents the auth handler.
type Handler struct {
	Account  AccountHandlerInterface
	CodeAuth CodeAuthHandlerInterface
	Captcha  CaptchaHandlerInterface
	Session  SessionHandlerInterface
	// OAuth    OAuthHandlerInterface
}

// New creates a new auth handler.
func New(svc *service.Service) *Handler {
	return &Handler{
		Account:  NewAccountHandler(svc),
		CodeAuth: NewCodeAuthHandler(svc),
		Captcha:  NewCaptchaHandler(svc),
		Session:  NewSessionHandler(svc),
		// OAuth:    NewOAuthHandler(svc),
	}
}
