package service

import (
	"ncobase/feature/menu/data"
	"ncobase/feature/menu/service/menu"
)

type Service struct {
	Menu menu.ServiceInterface
}

func New(d *data.Data) *Service {
	return &Service{
		Menu: menu.New(d),
	}
}
