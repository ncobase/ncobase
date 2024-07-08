package service

import (
	"ncobase/feature/asset/data"
	"ncobase/feature/asset/service/asset"
)

type Service struct {
	Asset asset.ServiceInterface
}

func New(d *data.Data) *Service {
	return &Service{
		Asset: asset.New(d),
	}
}
