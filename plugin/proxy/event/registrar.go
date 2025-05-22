package event

import (
	"context"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
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

	logger.Info(context.Background(), "Proxy event handlers registered")
}
