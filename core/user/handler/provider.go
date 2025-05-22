package handler

import "ncobase/user/service"

// Handler represents the user handler.
type Handler struct {
	User        UserHandlerInterface
	UserProfile UserProfileHandlerInterface
	Employee    EmployeeHandlerInterface
}

// New creates a new handler.
func New(svc *service.Service) *Handler {
	return &Handler{
		User:        NewUserHandler(svc),
		UserProfile: NewUserProfileHandler(svc),
		Employee:    NewEmployeeHandler(svc),
	}
}
