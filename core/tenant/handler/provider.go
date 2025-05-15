package handler

import "ncobase/tenant/service"

// Handler represents the tenant handler.
type Handler struct {
	Tenant TenantHandlerInterface
}

// New creates a new handler.
func New(svc *service.Service) *Handler {
	return &Handler{
		Tenant: NewTenantHandler(svc),
	}
}
