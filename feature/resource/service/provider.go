package service

import (
	"ncobase/feature/resource/data"
)

// Service is the struct for the resource service.
type Service struct {
	Asset AssetServiceInterface
}

// New creates a new resource service.
func New(d *data.Data) *Service {
	return &Service{
		Asset: NewAssetService(d),
	}
}
