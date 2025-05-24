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

	logger.Debugf(context.Background(), "Access event handlers registered: %d handlers", len(handlers))
}
