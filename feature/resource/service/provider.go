package service

import (
	"ncobase/feature/resource/data"
	"ncobase/feature/resource/service/asset"
)

type Service struct {
	Asset asset.ServiceInterface
}

func New(d *data.Data) *Service {
	return &Service{
		Asset: asset.New(d),
	}
}
