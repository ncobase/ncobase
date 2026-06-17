package handler

import "ncobase/plugin/ai/service"

type Handler struct {
	AI     AIHandlerInterface
	Run    RunHandlerInterface
	Action ActionHandlerInterface
}

func New(svc *service.Service) *Handler {
	return &Handler{
		AI:     NewAIHandler(svc.AI),
		Run:    NewRunHandler(svc.Run),
		Action: NewActionHandler(svc.AI),
	}
}
