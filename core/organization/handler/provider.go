package handler

import "ncobase/organization/service"

// Handler represents the organization handler.
type Handler struct {
	Organization OrganizationHandlerInterface
}

// New creates new organization handler.
func New(svc *service.Service) *Handler {
	return &Handler{
		Organization: NewOrganizationHandler(svc),
	}
}
