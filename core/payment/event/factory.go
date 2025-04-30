package event

import (
	ext "github.com/ncobase/ncore/extension/types"
)

// Factory creates event components
type Factory struct{}

// NewFactory creates a new event factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateManager creates a new event manager
func (f *Factory) CreateManager(em ext.ManagerInterface) *Manager {
	return NewManager(em)
}

// CreatePublisher creates a new publisher
func (f *Factory) CreatePublisher(em ext.ManagerInterface) Publisher {
	manager := f.CreateManager(em)
	return NewPublisher(manager)
}

// CreateRegistrar creates a new registrar
func (f *Factory) CreateRegistrar(em ext.ManagerInterface) *Registrar {
	return NewRegistrar(em)
}
