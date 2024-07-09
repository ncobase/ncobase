package handler

import (
	"ncobase/feature/system/handler/menu"
	"ncobase/feature/system/service"
)

type Handler struct {
	Menu menu.HandlerInterface
}

func New(s *service.Service) *Handler {
	return &Handler{
		Menu: menu.New(s),
	}
}
