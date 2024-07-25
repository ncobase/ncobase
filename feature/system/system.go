package system

import (
	"fmt"
	"ncobase/cmd/ncobase/middleware"
	"ncobase/common/config"
	"ncobase/common/feature"
	"ncobase/feature/system/data"
	"ncobase/feature/system/handler"
	"ncobase/feature/system/service"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	name         = "system"
	desc         = "system module"
	version      = "1.0.0"
	dependencies []string
)

// Module represents the system module.
type Module struct {
	initialized bool
	mu          sync.RWMutex
	fm          *feature.Manager
	conf        *config.Config
	h           *handler.Handler
	s           *service.Service
	d           *data.Data
	cleanup     func(name ...string)
}

// New creates a new instance of the system module.
func New() feature.Interface {
	return &Module{}
}

// PreInit performs any necessary setup before initialization
func (m *Module) PreInit() error {
	// Implement any pre-initialization logic here
	return nil
}

// Init initializes the system module with the given config object
func (m *Module) Init(conf *config.Config, fm *feature.Manager) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("system module already initialized")
	}

	m.d, m.cleanup, err = data.New(conf.Data)
	if err != nil {
		return err
	}

	m.fm = fm
	m.initialized = true
	m.conf = conf

	return nil
}

// PostInit performs any necessary setup after initialization
func (m *Module) PostInit() error {
	m.s = service.New(m.d, m.fm)
	m.h = handler.New(m.s)
	// Subscribe to relevant events
	m.subscribeEvents(m.fm)
	return nil
}

// Name returns the name of the module
func (m *Module) Name() string {
	return name
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(e *gin.Engine) {
	// Setup middleware
	// API v1 endpoints
	v1 := e.Group("/v1")
	// Menu endpoints
	menus := v1.Group("/menus", middleware.AuthenticatedUser)
	{
		menus.GET("", m.h.Menu.List)
		menus.POST("", m.h.Menu.Create)
		menus.GET("/:slug", m.h.Menu.Get)
		menus.PUT("/:slug", m.h.Menu.Update)
		menus.DELETE("/:slug", m.h.Menu.Delete)
	}
	// Dictionary endpoints
	dictionaries := v1.Group("/dictionaries", middleware.AuthenticatedUser)
	{
		dictionaries.GET("", m.h.Dictionary.List)
		dictionaries.POST("", m.h.Dictionary.Create)
		dictionaries.GET("/:slug", m.h.Dictionary.Get)
		dictionaries.PUT("/:slug", m.h.Dictionary.Update)
		dictionaries.DELETE("/:slug", m.h.Dictionary.Delete)
	}
}

// GetHandlers returns the handlers for the module
func (m *Module) GetHandlers() feature.Handler {
	return m.h
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

// Version returns the version of the module
func (m *Module) Version() string {
	return version
}

// Dependencies returns the dependencies of the module
func (m *Module) Dependencies() []string {
	return dependencies
}

// SubscribeEvents subscribes to relevant events
func (m *Module) subscribeEvents(_ *feature.Manager) {
	// Implement any event subscriptions here
}
