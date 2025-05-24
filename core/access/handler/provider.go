package handler

import "ncobase/access/service"

// Handler represents the access handler.
type Handler struct {
	Activity       ActivityHandlerInterface
	Casbin         CasbinHandlerInterface
	Role           RoleHandlerInterface
	Permission     PermissionHandlerInterface
	RolePermission RolePermissionHandlerInterface
}

// New creates a new handler.
func New(s *service.Service) *Handler {
	return &Handler{
		Activity:       NewActivityHandler(s),
		Casbin:         NewCasbinHandler(s),
		Role:           NewRoleHandler(s),
		Permission:     NewPermissionHandler(s),
		RolePermission: NewRolePermissionHandler(s),
	}
}
