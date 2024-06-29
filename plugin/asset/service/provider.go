package service

import (
	"ncobase/plugin/asset/data"
	"ncobase/plugin/asset/service/asset"
)

type Service struct {
	Asset asset.ServiceInterface
}

func New(d *data.Data) *Service {
	return &Service{
		Asset: asset.New(d),
	}
}
