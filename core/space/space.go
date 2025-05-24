package space

import (
	"fmt"
	"ncobase/cmd/ncobase/middleware"
	"ncobase/space/data"
	"ncobase/space/handler"
	"ncobase/space/service"
	"sync"

	"github.com/ncobase/ncore/config"
	exr "github.com/ncobase/ncore/extension/registry"
	ext "github.com/ncobase/ncore/extension/types"

	"github.com/gin-gonic/gin"
)

var (
	name         = "space"
	desc         = "Space module, provides organization management"
	version      = "1.0.0"
	dependencies []string
	typeStr      = "module"
	group        = "org"
)

// Module represents the group module.
type Module struct {
	ext.OptionalImpl

	initialized bool
	mu          sync.RWMutex
	em          ext.ManagerInterface
	h           *handler.Handler
	s           *service.Service
	d           *data.Data
	cleanup     func(name ...string)

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
	exr.RegisterToGroupWithWeakDeps(New(), group, []string{"user"})
}

// New creates a new instance of the group module.
func New() ext.Interface {
	return &Module{}
}

// Init initializes the group module with the given config object
func (m *Module) Init(conf *config.Config, em ext.ManagerInterface) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("space module already initialized")
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

	// Subscribe to extension events for dependency refresh
	m.em.SubscribeEvent("exts.user.ready", func(data any) {
		m.s.RefreshDependencies()
	})

	// Subscribe to all extensions registration event
	m.em.SubscribeEvent("exts.all.registered", func(data any) {
		m.s.RefreshDependencies()
	})

	// Publish own service ready event
	m.em.PublishEvent("exts.space.ready", map[string]string{
		"name":   m.Name(),
		"status": "ready",
	})

	return nil
}

// Name returns the name of the module
func (m *Module) Name() string {
	return name
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	// Group endpoints - organization management permissions
	orgGroup := r.Group("/"+m.Group(), middleware.AuthenticatedUser)
	{
		orgGroup.GET("", middleware.HasPermission("read:group"), m.h.Group.List)
		orgGroup.POST("", middleware.HasPermission("manage:group"), m.h.Group.Create)
		orgGroup.GET("/:slug", middleware.HasPermission("read:group"), m.h.Group.Get)
		orgGroup.PUT("/:slug", middleware.HasPermission("manage:group"), m.h.Group.Update)
		orgGroup.DELETE("/:slug", middleware.HasPermission("manage:group"), m.h.Group.Delete)

		orgGroup.GET("/:slug/members", middleware.HasPermission("read:group"), m.h.Group.GetMembers)
		orgGroup.POST("/:slug/members", middleware.HasPermission("manage:group"), m.h.Group.AddMember)
		orgGroup.PUT("/:slug/members/:userId", middleware.HasPermission("manage:group"), m.h.Group.UpdateMember)
		orgGroup.DELETE("/:slug/members/:userId", middleware.HasPermission("manage:group"), m.h.Group.RemoveMember)
		orgGroup.GET("/:slug/members/:userId/check", middleware.HasPermission("read:group"), m.h.Group.IsUserMember)
		orgGroup.GET("/:slug/members/:userId/is-owner", middleware.HasPermission("read:group"), m.h.Group.IsUserOwner)
		orgGroup.GET("/:slug/members/:userId/role", middleware.HasPermission("read:group"), m.h.Group.GetUserRole)
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

// Version returns the version of the module
func (m *Module) Version() string {
	return version
}

// Dependencies returns the dependencies of the module
func (m *Module) Dependencies() []string {
	return dependencies
}

// GetAllDependencies returns all dependencies with their types
func (m *Module) GetAllDependencies() []ext.DependencyEntry {
	return []ext.DependencyEntry{
		{Name: "user", Type: ext.WeakDependency},
	}
}

// Description returns the description of the module
func (m *Module) Description() string {
	return desc
}

// Type returns the type of the module
func (m *Module) Type() string {
	return typeStr
}

// Group returns the domain group of the module belongs
func (m *Module) Group() string {
	return group
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
