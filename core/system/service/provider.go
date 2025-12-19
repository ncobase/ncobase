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
	Option     OptionServiceInterface
	Admin      AdminServiceInterface
	d          *data.Data
	em         ext.ManagerInterface
}

// New creates a new service.
func New(d *data.Data, em ext.ManagerInterface) *Service {
	tsw := wrapper.NewSpaceServiceWrapper(em)

	s := &Service{
		Menu:       NewMenuService(d, em, tsw),
		Dictionary: NewDictionaryService(d),
		Option:     NewOptionService(d),
		d:          d,
		em:         em,
	}

	// Initialize admin service with reference to the main service
	s.Admin = newAdminService(s)

	return s
}
