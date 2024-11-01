package service

import (
	"ncobase/common/feature"
	"ncobase/core/system/data"
)

// Service represents the system service.
type Service struct {
	Menu       MenuServiceInterface
	Dictionary DictionaryServiceInterface
	Options    OptionsServiceInterface
	fm         *feature.Manager
}

// New creates a new service.
func New(d *data.Data, fm *feature.Manager) *Service {
	return &Service{
		Menu:       NewMenuService(d, fm),
		Dictionary: NewDictionaryService(d),
		Options:    NewOptionsService(d),
		fm:         fm,
	}
}
