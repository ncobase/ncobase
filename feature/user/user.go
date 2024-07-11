package user

import (
	"fmt"
	"ncobase/common/config"
	"ncobase/feature"
	"ncobase/feature/user/data"
	"ncobase/feature/user/handler"
	"ncobase/feature/user/service"
	"ncobase/middleware"
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

// HasRoutes returns true if the plugin has routes, false otherwise
func (m *Module) HasRoutes() bool {
	return true
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(e *gin.Engine) {
	// API v1 endpoints
	v1 := e.Group("/v1")

	// Account endpoints
	account := v1.Group("/account", middleware.Authenticated)
	{
		account.GET("", m.h.User.GetMeHandler)
		account.PUT("/password", m.h.User.UpdatePasswordHandler)
		// account.GET("/tenant", m.h.User.AccountTenantHandler)
		// account.GET("/tenants", m.h.User.AccountTenantsHandler)
	}

	// User endpoints
	users := v1.Group("/users")
	{
		// users.GET("", m.h.ListUserHandler)
		// users.POST("", m.h.CreateUserHandler)
		users.GET("/:username", m.h.User.GetUserHandler)
		// users.PUT("/:username", m.h.UpdateUserHandler)
		// users.DELETE("/:username", m.h.DeleteUserHandler)
		// users.GET("/:username/roles", m.h.ListUserRoleHandler)
		// users.GET("/:username/groups", m.h.ListUserGroupHandler)
		// users.GET("/:username/tenants", m.h.UserTenantHandler)
		// users.GET("/:username/tenants/:slug", m.h.UserTenantHandler)
		// users.GET("/:username/tenant/belongs", middleware.Authenticated, m.h.ListUserBelongHandler)
	}
}

// GetHandlers returns the handlers for the module
func (m *Module) GetHandlers() map[string]feature.Handler {
	return map[string]feature.Handler{
		"user": m.h.User,
	}
}

// GetServices returns the services for the module
func (m *Module) GetServices() map[string]feature.Service {
	return map[string]feature.Service{
		"user": m.s.User,
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
