package bootstrap

import (
	"fmt"
	"ncobase/common/config"
	"ncobase/feature"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	name         = "bootstrap"
	desc         = "Bootstrap module"
	version      = "1.0.0"
	dependencies []string
)

// Module represents the bootstrap module
type Module struct {
	initialized bool
	mu          sync.RWMutex
	fm          *feature.Manager
	cleanup     func(name ...string)
}

// Name returns the name of the module
func (m *Module) Name() string {
	return name
}

// PreInit performs any necessary setup before initialization
func (m *Module) PreInit() error {
	// Implement any pre-initialization logic here
	return nil
}

// Init initializes the module
func (m *Module) Init(conf *config.Config, fm *feature.Manager) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("bootstrap module already initialized")
	}

	m.fm = fm
	m.initialized = true

	return nil
}

// PostInit performs any necessary setup after initialization
func (m *Module) PostInit() error {
	return nil
}

// HasRoutes returns true if the module has routes, false otherwise
func (m *Module) HasRoutes() bool {
	return false
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(e *gin.Engine) {}

// GetHandlers returns the handlers for the module
func (m *Module) GetHandlers() map[string]feature.Handler {
	return nil
}

// GetServices returns the services for the module
func (m *Module) GetServices() map[string]feature.Service {
	return nil
}

// PreCleanup performs any necessary cleanup before the main cleanup
func (m *Module) PreCleanup() error {
	// Implement any pre-cleanup logic here
	return nil
}

// Cleanup cleans up the module
func (m *Module) Cleanup() error {
	if m.cleanup != nil {
		m.cleanup(m.Name())
	}
	return nil
}

// GetMetadata returns the metadata of the module
func (m *Module) GetMetadata() feature.Metadata {
	return feature.Metadata{
		Name:         m.Name(),
		Version:      m.Version(),
		Dependencies: m.Dependencies(),
		Description:  desc,
	}
}

// Status returns the status of the module
func (m *Module) Status() string {
	// Implement logic to check the module status
	return "active"
}

// Version returns the version of the module
func (m *Module) Version() string {
	return version
}

// Dependencies returns the dependencies of the module
func (m *Module) Dependencies() []string {
	return []string{}
}

// SubscribeEvents subscribes to relevant events
func (m *Module) subscribeEvents(fm *feature.Manager) {
	// Implement any event subscriptions here
}
