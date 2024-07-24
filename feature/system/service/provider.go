package service

import (
	"ncobase/common/feature"
	"ncobase/feature/system/data"
)

// Service represents the system service.
type Service struct {
	Menu MenuServiceInterface
	fm   *feature.Manager
}

// New creates a new service.
func New(d *data.Data, fm *feature.Manager) *Service {
	return &Service{
		Menu: NewMenuService(d, fm),
		fm:   fm,
	}
}
