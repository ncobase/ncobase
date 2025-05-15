package handler

import "ncobase/system/service"

// Handler represents the system handler.
type Handler struct {
	Menu       MenuHandlerInterface
	Dictionary DictionaryHandlerInterface
	Options    OptionsHandlerInterface
}

// New creates new system handler.
func New(svc *service.Service) *Handler {
	return &Handler{
		Menu:       NewMenuHandler(svc),
		Dictionary: NewDictionaryHandler(svc),
		Options:    NewOptionsHandler(svc),
	}
}
