package event

import (
	ext "github.com/ncobase/ncore/extension/types"
)

// Registrar handles registration of event handlers
type Registrar struct {
	em ext.ManagerInterface
}

// NewRegistrar creates new event registrar
func NewRegistrar(manager ext.ManagerInterface) *Registrar {
	return &Registrar{em: manager}
}

// RegisterHandlers registers event handlers with manager
func (r *Registrar) RegisterHandlers(provider HandlerProvider) {
	handlers := provider.GetHandlers()

	for eventName, handler := range handlers {
		r.em.SubscribeEvent(eventName, handler)
	}
}
