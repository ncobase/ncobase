package menu

import (
	"ncobase/common/config"
	"ncobase/feature"
	"ncobase/feature/menu/data"
	"ncobase/feature/menu/handler"
	"ncobase/feature/menu/service"
	"ncobase/middleware"

	"github.com/gin-gonic/gin"
)

const (
	name    = "menu"
	desc    = "menu module"
	version = "1.0.0"
)

// Module represents the menu module.
type Module struct {
	h       *handler.Handler
	s       *service.Service
	d       *data.Data
	cleanup func()
}

// New creates a new instance of the menu module.
func New() feature.Interface {
	return &Module{}
}

// PreInit performs any necessary setup before initialization
func (m *Module) PreInit() error {
	// Implement any pre-initialization logic here
	return nil
}

// Init initializes the menu module with the given config object
func (m *Module) Init(conf *config.Config) (err error) {
	m.d, m.cleanup, err = data.New(conf.Data)
	if err != nil {
		return err
	}
	svc := service.New(m.d)
	m.s = svc
	m.h = handler.New(svc)
	return nil
}

// PostInit performs any necessary setup after initialization
func (m *Module) PostInit() error {
	// Implement any post-initialization logic here
	return nil
}

// Name returns the name of the module
func (m *Module) Name() string {
	return name
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(e *gin.Engine) {
	menus := e.Group("/v1/menus", middleware.Authenticated)
	{
		menus.GET("", m.h.Menu.List)
		menus.POST("", m.h.Menu.Create)
		menus.GET("/:slug", m.h.Menu.Get)
		menus.PUT("/:slug", m.h.Menu.Update)
		menus.DELETE("/:slug", m.h.Menu.Delete)
	}
}

// GetHandlers returns the handlers for the module
func (m *Module) GetHandlers() map[string]feature.Handler {
	return map[string]feature.Handler{
		"menu": m.h,
	}
}

// GetServices returns the services for the module
func (m *Module) GetServices() map[string]feature.Service {
	return map[string]feature.Service{
		"menu": m.s,
	}
}

// PreCleanup performs any necessary cleanup before the main cleanup
func (m *Module) PreCleanup() error {
	// Implement any pre-cleanup logic here
	return nil
}

// Cleanup cleans up the module
func (m *Module) Cleanup() error {
	if m.cleanup != nil {
		m.cleanup()
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
	return []string{}
}
