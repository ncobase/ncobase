package handler

import "ncobase/core/system/service"

// Handler represents the system handler.
type Handler struct {
	Menu       MenuHandlerInterface
	Dictionary DictionaryHandlerInterface
	Option     OptionHandlerInterface
	Admin      AdminHandlerInterface
}

// New creates new system handler.
func New(svc *service.Service) *Handler {
	return &Handler{
		Menu:       NewMenuHandler(svc),
		Dictionary: NewDictionaryHandler(svc),
		Option:     NewOptionHandler(svc),
		Admin:      NewAdminHandler(svc),
	}
}
