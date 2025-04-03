package service

import (
	"ncobase/core/system/data"
	"ncobase/ncore/extension"
)

// Service represents the system service.
type Service struct {
	Menu       MenuServiceInterface
	Dictionary DictionaryServiceInterface
	Options    OptionsServiceInterface
	em         *extension.Manager
}

// New creates a new service.
func New(d *data.Data, em *extension.Manager) *Service {
	return &Service{
		Menu:       NewMenuService(d, em),
		Dictionary: NewDictionaryService(d),
		Options:    NewOptionsService(d),
		em:         em,
	}
}
