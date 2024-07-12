package group

import (
	"fmt"
	"ncobase/cmd/ncobase/middleware"
	"ncobase/common/config"
	"ncobase/feature"
	"ncobase/feature/group/data"
	"ncobase/feature/group/service"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	name         = "group"
	desc         = "group module"
	version      = "1.0.0"
	dependencies []string
)

// Module represents the group module.
type Module struct {
	initialized bool
	mu          sync.RWMutex
	fm          *feature.Manager
	// h       *handler.Handler
	s       *service.Service
	d       *data.Data
	cleanup func(name ...string)
}

// New creates a new instance of the group module.
func New() feature.Interface {
	return &Module{}
}

// PreInit performs any necessary setup before initialization
func (m *Module) PreInit() error {
	// Implement any pre-initialization logic here
	return nil
}

// Init initializes the group module with the given config object
func (m *Module) Init(conf *config.Config, fm *feature.Manager) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("group module already initialized")
	}

	m.d, m.cleanup, err = data.New(conf.Data)
	if err != nil {
		return err
	}

	m.fm = fm
	m.initialized = true

	return nil
}

// PostInit performs any necessary setup after initialization
func (m *Module) PostInit() error {
	m.s = service.New(m.d)
	// m.h = handler.New(m.s)
	return nil
}

// Name returns the name of the module
func (m *Module) Name() string {
	return name
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(e *gin.Engine) {
	// API v1 endpoints
	v1 := e.Group("/v1")
	// Group endpoints
	groups := v1.Group("/groups", middleware.AuthenticatedUser)
	{
		groups.GET("", nil)
		groups.POST("", nil)
		groups.GET("/:slug", nil)
		groups.PUT("/:slug", nil)
		groups.DELETE("/:slug", nil)
	}
}

// GetHandlers returns the handlers for the module
func (m *Module) GetHandlers() feature.Handler {
	return nil
}

// GetServices returns the services for the module
func (m *Module) GetServices() feature.Service {
	return m.s
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

// Status returns the status of the module
func (m *Module) Status() string {
	return "active"
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

// Version returns the version of the plugin
func (m *Module) Version() string {
	return version
}

// Dependencies returns the dependencies of the plugin
func (m *Module) Dependencies() []string {
	return dependencies
}
