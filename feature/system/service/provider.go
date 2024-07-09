package service

import (
	"ncobase/feature"
	"ncobase/feature/system/data"
	"ncobase/feature/system/service/menu"
)

type Service struct {
	Menu menu.ServiceInterface
	fm   *feature.Manager
}

func New(d *data.Data, fm *feature.Manager) *Service {
	return &Service{
		Menu: menu.New(d, fm),
		fm:   fm,
	}
}
