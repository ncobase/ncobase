package handler

import (
	"ncobase/feature/menu/handler/menu"
	"ncobase/feature/menu/service"
)

type Handler struct {
	Menu menu.HandlerInterface
}

func New(s *service.Service) *Handler {
	return &Handler{
		Menu: menu.New(s),
	}
}
