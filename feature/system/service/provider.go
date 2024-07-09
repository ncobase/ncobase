package service

import (
	"ncobase/feature/system/data"
	"ncobase/feature/system/service/menu"
)

type Service struct {
	Menu menu.ServiceInterface
}

func New(d *data.Data) *Service {
	return &Service{
		Menu: menu.New(d),
	}
}
