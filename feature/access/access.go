package access

import (
	"fmt"
	"ncobase/common/config"
	"ncobase/feature"
	"ncobase/feature/access/data"
	"ncobase/feature/access/handler"
	"ncobase/feature/access/middleware"
	"ncobase/feature/access/service"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	name         = "access"
	desc         = "access module"
	version      = "1.0.0"
	dependencies []string
)

// Module represents the access module.
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

// New creates a new instance of the access module.
func New() feature.Interface {
	return &Module{}
}

// PreInit performs any necessary setup before initialization
func (m *Module) PreInit() error {
	// Implement any pre-initialization logic here
	return nil
}

// Init initializes the access module with the given config object
func (m *Module) Init(conf *config.Config, fm *feature.Manager) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("access module already initialized")
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

	m.s = service.New(m.conf, m.d)
	m.h = handler.New(m.s)

	return nil
}

// Name returns the name of the module
func (m *Module) Name() string {
	return name
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(e *gin.Engine) {

	// Setup middleware
	// enforcer, err := m.s.CasbinAdapter.InitEnforcer()
	// if err != nil {
	// 	panic(err)
	// }
	// e.Use(middleware.Authorized(enforcer, m.conf.Auth.Whitelist, m.s))

	v1 := e.Group("/v1")
	// Role endpoints
	roles := v1.Group("/roles", middleware.AuthenticatedUser)
	{
		roles.GET("", m.h.Role.List)
		roles.POST("", m.h.Role.Create)
		roles.GET("/:slug", m.h.Role.Get)
		roles.PUT("/:slug", m.h.Role.Update)
		roles.DELETE("/:slug", m.h.Role.Delete)
		roles.GET("/:slug/permissions", m.h.RolePermission.ListRolePermission)
		// roles.GET("/:slug/users", m.h.Role.ListUser)
	}
	// Permission endpoints
	permissions := v1.Group("/permissions", middleware.AuthenticatedUser)
	{
		permissions.GET("", m.h.Permission.List)
		permissions.POST("", m.h.Permission.Create)
		permissions.GET("/:slug", m.h.Permission.Get)
		permissions.PUT("/:slug", m.h.Permission.Update)
		permissions.DELETE("/:slug", m.h.Permission.Delete)
	}
	// Casbin Rule endpoints
	policies := v1.Group("/policies", middleware.AuthenticatedUser)
	{
		policies.GET("", m.h.Casbin.List)
		policies.POST("", m.h.Casbin.Create)
		policies.GET("/:id", m.h.Casbin.Get)
		policies.PUT("/:id", m.h.Casbin.Update)
		policies.DELETE("/:id", m.h.Casbin.Delete)
	}
}

// GetHandlers returns the handlers for the module
func (m *Module) GetHandlers() map[string]feature.Handler {
	return map[string]feature.Handler{
		"access":          m.h.Permission,
		"role":            m.h.Role,
		"casbin":          m.h.Casbin,
		"role_permission": m.h.RolePermission,
	}
}

// GetServices returns the services for the module
func (m *Module) GetServices() map[string]feature.Service {
	return map[string]feature.Service{
		"access":           m.s.Permission,
		"role":             m.s.Role,
		"casbin":           m.s.Casbin,
		"casbin_adapter":   m.s.CasbinAdapter,
		"role_permission":  m.s.RolePermission,
		"user_role":        m.s.UserRole,
		"user_tenant_role": m.s.UserTenantRole,
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
