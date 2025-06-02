package service

import (
	"ncobase/system/data"
	"ncobase/system/wrapper"

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
	tsw := wrapper.NewTenantServiceWrapper(em)

	return &Service{
		Menu:       NewMenuService(d, em, tsw),
		Dictionary: NewDictionaryService(d),
		Options:    NewOptionsService(d),
		em:         em,
	}
}
