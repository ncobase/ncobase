package resource

import (
	"fmt"
	"ncobase/cmd/ncobase/middleware"
	"ncobase/common/config"
	"ncobase/common/feature"
	"ncobase/feature/resource/data"
	"ncobase/feature/resource/handler"
	"ncobase/feature/resource/service"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	name         = "resource"
	desc         = "Resource module"
	version      = "1.0.0"
	dependencies []string
)

// Module represents the resource module
type Module struct {
	initialized bool
	mu          sync.RWMutex
	fm          *feature.Manager
	s           *service.Service
	h           *handler.Handler
	d           *data.Data
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
		return fmt.Errorf("resource module already initialized")
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
	m.h = handler.New(m.s)
	// Subscribe to relevant events
	m.subscribeEvents(m.fm)
	return nil
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(e *gin.Engine) {
	// API v1 endpoints
	v1 := e.Group("/v1")
	// Asset endpoints
	assets := v1.Group("/assets", middleware.AuthenticatedUser)
	{
		assets.GET("", m.h.Asset.List)
		assets.POST("", m.h.Asset.Create)
		assets.GET("/:slug", m.h.Asset.Get)
		assets.PUT("/:slug", m.h.Asset.Update)
		assets.DELETE("/:slug", m.h.Asset.Delete)
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
	return dependencies
}

// SubscribeEvents subscribes to relevant events
func (m *Module) subscribeEvents(_ *feature.Manager) {
	// Implement any event subscriptions here
}

// func init() {
// 	feature.RegisterModule(&Module{}, feature.Metadata{
// 		Name:         name + "-development",
// 		Version:      version,
// 		Dependencies: dependencies,
// 		Description:  desc,
// 	})
// }

// New creates a new instance of the auth module.
func New() feature.Interface {
	return &Module{}
}
