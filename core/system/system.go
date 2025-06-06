package system

import (
	"fmt"
	"ncobase/cmd/ncobase/middleware"
	"ncobase/system/data"
	"ncobase/system/handler"
	"ncobase/system/service"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	exr "github.com/ncobase/ncore/extension/registry"
	ext "github.com/ncobase/ncore/extension/types"
)

var (
	name         = "system"
	desc         = "System module"
	version      = "1.0.0"
	dependencies []string
	typeStr      = "module"
	group        = "sys"
)

// Module represents the system module.
type Module struct {
	ext.OptionalImpl

	initialized bool
	mu          sync.RWMutex
	em          ext.ManagerInterface
	cleanup     func(name ...string)

	h *handler.Handler
	s *service.Service
	d *data.Data

	discovery
}

// discovery represents the service discovery
type discovery struct {
	address string
	tags    []string
	meta    map[string]string
}

// init registers the module
func init() {
	exr.RegisterToGroupWithWeakDeps(New(), group, []string{"tenant"})
}

// New creates a new instance of the system module.
func New() ext.Interface {
	return &Module{}
}

// Init initializes the system module with the given config object
func (m *Module) Init(conf *config.Config, em ext.ManagerInterface) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("system module already initialized")
	}

	m.d, m.cleanup, err = data.New(conf.Data)
	if err != nil {
		return err
	}

	// service discovery
	if conf.Consul != nil {
		m.discovery.address = conf.Consul.Address
		m.discovery.tags = conf.Consul.Discovery.DefaultTags
		m.discovery.meta = conf.Consul.Discovery.DefaultMeta
	}

	m.em = em
	m.initialized = true

	return nil
}

// PostInit performs any necessary setup after initialization
func (m *Module) PostInit() error {
	m.s = service.New(m.d, m.em)
	m.h = handler.New(m.s)
	// Subscribe to relevant events
	m.subscribeEvents(m.em)

	return nil
}

// Name returns the name of the module
func (m *Module) Name() string {
	return name
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	// Authenticated routes
	sysGroup := r.Group("/"+m.Group(), middleware.AuthenticatedUser)

	// Menu endpoints
	menus := sysGroup.Group("/menus")
	{
		// Basic menu access - all authenticated users
		menus.GET("", m.h.Menu.List)
		menus.GET("/navigation", m.h.Menu.GetNavigationMenus)
		menus.GET("/tree", m.h.Menu.GetMenuTree)
		menus.GET("/authorized/:user_id", m.h.Menu.GetUserAuthorizedMenus)
		menus.GET("/:slug", m.h.Menu.Get)
		menus.GET("/slug/:slug", m.h.Menu.GetBySlug)

		// Menu management - requires specific permission
		menus.POST("", middleware.HasPermission("manage:menu"), m.h.Menu.Create)
		menus.PUT("", middleware.HasPermission("manage:menu"), m.h.Menu.Update)
		menus.DELETE("/:slug", middleware.HasPermission("manage:menu"), m.h.Menu.Delete)

		// Menu operations - requires specific permission
		menus.PUT("/:id/move", middleware.HasPermission("manage:menu"), m.h.Menu.MoveMenu)
		menus.POST("/reorder", middleware.HasPermission("manage:menu"), m.h.Menu.ReorderMenus)

		// Status toggle endpoint - requires specific permission
		menus.PUT("/:id/:action", middleware.HasPermission("manage:menu"), m.h.Menu.ToggleMenuStatus)
	}

	// Dictionary endpoints
	dictionaries := sysGroup.Group("/dictionaries")
	{
		// Basic dictionary access
		dictionaries.GET("", m.h.Dictionary.List)
		dictionaries.GET("/:slug", m.h.Dictionary.Get)
		dictionaries.GET("/slug/:slug", m.h.Dictionary.GetBySlug)
		dictionaries.GET("/options/:slug", m.h.Dictionary.GetEnumOptions)
		dictionaries.GET("/validate/:slug", m.h.Dictionary.ValidateEnumValue)
		dictionaries.POST("/batch", m.h.Dictionary.BatchGetBySlug)

		// Dictionary management
		dictionaries.POST("", middleware.HasPermission("manage:dictionary"), m.h.Dictionary.Create)
		dictionaries.PUT("", middleware.HasPermission("manage:dictionary"), m.h.Dictionary.Update)
		dictionaries.DELETE("/:slug", middleware.HasPermission("manage:dictionary"), m.h.Dictionary.Delete)
	}

	// Options endpoints
	options := sysGroup.Group("/options")
	{
		// Basic options access
		options.GET("", m.h.Option.List)
		options.GET("/:option", m.h.Option.Get)
		options.GET("/name/:name", m.h.Option.GetByName)
		options.GET("/type/:type", m.h.Option.GetByType)
		options.POST("/batch", m.h.Option.BatchGetByNames)

		// Options management - requires system management permission
		options.POST("", middleware.HasPermission("manage:system"), m.h.Option.Create)
		options.PUT("", middleware.HasPermission("manage:system"), m.h.Option.Update)
		options.DELETE("/:option", middleware.HasPermission("manage:system"), m.h.Option.Delete)
		options.DELETE("/prefix", middleware.HasPermission("manage:system"), m.h.Option.DeleteByPrefix)
	}
}

// GetHandlers returns the handlers for the module
func (m *Module) GetHandlers() ext.Handler {
	return m.h
}

// GetServices returns the services for the module
func (m *Module) GetServices() ext.Service {
	return m.s
}

// Cleanup cleans up the module
func (m *Module) Cleanup() error {
	if m.cleanup != nil {
		m.cleanup(m.Name())
	}
	return nil
}

// GetMetadata returns the metadata of the module
func (m *Module) GetMetadata() ext.Metadata {
	return ext.Metadata{
		Name:         m.Name(),
		Version:      m.Version(),
		Dependencies: m.Dependencies(),
		Description:  m.Description(),
		Type:         m.Type(),
		Group:        m.Group(),
	}
}

// Dependencies returns the dependencies of the module
func (m *Module) Dependencies() []string {
	return dependencies
}

// GetAllDependencies returns all dependencies of the module and its dependencies
func (m *Module) GetAllDependencies() []ext.DependencyEntry {
	return []ext.DependencyEntry{
		{Name: "tenant", Type: ext.WeakDependency},
	}
}

// Description returns the description of the module
func (m *Module) Description() string {
	return desc
}

// Version returns the version of the module
func (m *Module) Version() string {
	return version
}

// Type returns the type of the module
func (m *Module) Type() string {
	return typeStr
}

// Group returns the domain group of the module belongs
func (m *Module) Group() string {
	return group
}

// SubscribeEvents subscribes to relevant events
func (m *Module) subscribeEvents(_ ext.ManagerInterface) {
	// Implement any event subscriptions here
}

// GetServiceInfo returns service registration info if NeedServiceDiscovery returns true
func (m *Module) GetServiceInfo() *ext.ServiceInfo {
	if !m.NeedServiceDiscovery() {
		return nil
	}

	metadata := m.GetMetadata()

	tags := append(m.discovery.tags, metadata.Group, metadata.Type)

	meta := make(map[string]string)
	for k, v := range m.discovery.meta {
		meta[k] = v
	}
	meta["name"] = metadata.Name
	meta["version"] = metadata.Version
	meta["group"] = metadata.Group
	meta["type"] = metadata.Type
	meta["description"] = metadata.Description

	return &ext.ServiceInfo{
		Address: m.discovery.address,
		Tags:    tags,
		Meta:    meta,
	}
}
