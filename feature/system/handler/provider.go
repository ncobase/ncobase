package handler

import "ncobase/feature/system/service"

// Handler represents the system handler.
type Handler struct {
	Menu       MenuHandlerInterface
	Dictionary DictionaryHandlerInterface
}

// New creates new system handler.
func New(svc *service.Service) *Handler {
	return &Handler{
		Menu:       NewMenuHandler(svc),
		Dictionary: NewDictionaryHandler(svc),
	}
}
