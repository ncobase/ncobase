package service

import (
	"ncobase/core/system/data"

	ext "github.com/ncobase/ncore/extension/types"
)

// Service represents the system service.
type Service struct {
	Menu       MenuServiceInterface
	Dictionary DictionaryServiceInterface
	Options    OptionsServiceInterface
	em         ext.ManagerInterface
}

// New creates a new service.
func New(d *data.Data, em ext.ManagerInterface) *Service {
	return &Service{
		Menu:       NewMenuService(d, em),
		Dictionary: NewDictionaryService(d),
		Options:    NewOptionsService(d),
		em:         em,
	}
}
