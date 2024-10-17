package user

import (
	"fmt"
	"ncobase/common/config"
	"ncobase/common/feature"
	"ncobase/feature/user/data"
	"ncobase/feature/user/handler"
	"ncobase/feature/user/service"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	name         = "user"
	desc         = "user module"
	version      = "1.0.0"
	dependencies []string
)

// Module represents the user module.
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

// New creates a new instance of the user module.
func New() feature.Interface {
	return &Module{}
}

// PreInit performs any necessary setup before initialization
func (m *Module) PreInit() error {
	// Implement any pre-initialization logic here
	return nil
}

// Init initializes the user module with the given config object
func (m *Module) Init(conf *config.Config, fm *feature.Manager) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("user module already initialized")
	}

	m.d, m.cleanup, err = data.New(conf.Data)
	if err != nil {
		return err
	}

	m.fm = fm
	m.conf = conf
	m.initialized = true

	return nil
}

// PostInit performs any necessary setup after initialization
func (m *Module) PostInit() error {
	m.s = service.New(m.d)
	m.h = handler.New(m.s)
	return nil
}

// Name returns the name of the module
func (m *Module) Name() string {
	return name
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {

	// User endpoints
	users := r.Group("/users")
	{
		// users.GET("", m.h.List)
		// users.POST("", m.h.Create)
		users.GET("/:username", m.h.User.Get)
		// users.PUT("/:username", m.h.Update)
		// users.DELETE("/:username", m.h.Delete)
		// users.GET("/:username/roles", m.h.ListUserRoles)
		// users.GET("/:username/groups", m.h.ListUserGroups)
		// users.GET("/:username/tenants", m.h.UserTenants)
		// users.GET("/:username/tenants/:slug", m.h.UserTenant)
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
