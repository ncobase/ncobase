package service

import (
	nec "github.com/ncobase/ncore/ext/core"
	"ncobase/core/system/data"
)

// Service represents the system service.
type Service struct {
	Menu       MenuServiceInterface
	Dictionary DictionaryServiceInterface
	Options    OptionsServiceInterface
	em         nec.ManagerInterface
}

// New creates a new service.
func New(d *data.Data, em nec.ManagerInterface) *Service {
	return &Service{
		Menu:       NewMenuService(d, em),
		Dictionary: NewDictionaryService(d),
		Options:    NewOptionsService(d),
		em:         em,
	}
}
