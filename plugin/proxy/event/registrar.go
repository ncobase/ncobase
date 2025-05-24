package event

import (
	ext "github.com/ncobase/ncore/extension/types"
)

// Registrar handles registration of event handlers
type Registrar struct {
	em ext.ManagerInterface
}

// NewRegistrar creates a new event registrar
func NewRegistrar(manager ext.ManagerInterface) *Registrar {
	return &Registrar{
		em: manager,
	}
}

// RegisterHandlers registers event handlers with the manager
func (r *Registrar) RegisterHandlers(provider HandlerProvider) {
	handlers := provider.GetHandlers()

	// Register all handlers directly using the event name from the map key
	for eventName, handler := range handlers {
		r.em.SubscribeEvent(eventName, handler)
	}
}
