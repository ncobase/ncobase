package organization

import (
	"fmt"
	"ncobase/internal/middleware"
	"ncobase/core/organization/data"
	"ncobase/core/organization/handler"
	"ncobase/core/organization/service"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	exr "github.com/ncobase/ncore/extension/registry"
	ext "github.com/ncobase/ncore/extension/types"
)

var (
	name         = "organization"
	desc         = "Organization module, provides organization management"
	version      = "1.0.0"
	dependencies []string
	typeStr      = "module"
	group        = "sys"
)

// Module represents the organization module.
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

// New creates a new instance of the organization module.
func New() ext.Interface {
	return &Module{}
}

// Init initializes the organization module with the given config object
func (m *Module) Init(conf *config.Config, em ext.ManagerInterface) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("organization module already initialized")
	}

	m.d, m.cleanup, err = data.New(conf.Data, conf.Environment)
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

	return nil
}

// Name returns the name of the module
func (m *Module) Name() string {
	return name
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	// Organization endpoints
	spaceGroup := r.Group("/"+m.Group(), middleware.AuthenticatedUser)

	// Organization management endpoints
	spaceGroup = spaceGroup.Group("/orgs", middleware.HasAnyRole("super-admin", "system-admin"))
	{
		spaceGroup.GET("", middleware.HasPermission("read:organizations"), m.h.Organization.List)
		spaceGroup.POST("", middleware.HasPermission("manage:organizations"), m.h.Organization.Create)
		spaceGroup.GET("/:orgId", middleware.HasPermission("read:organizations"), m.h.Organization.Get)
		spaceGroup.PUT("/:orgId", middleware.HasPermission("manage:organizations"), m.h.Organization.Update)
		spaceGroup.DELETE("/:orgId", middleware.HasPermission("manage:organizations"), m.h.Organization.Delete)

		// Organization member management
		spaceGroup.GET("/:orgId/members", middleware.HasPermission("read:organizations"), m.h.Organization.GetMembers)
		spaceGroup.POST("/:orgId/members", middleware.HasPermission("manage:organizations"), m.h.Organization.AddMember)
		spaceGroup.PUT("/:orgId/members/:userId", middleware.HasPermission("manage:organizations"), m.h.Organization.UpdateMember)
		spaceGroup.DELETE("/:orgId/members/:userId", middleware.HasPermission("manage:organizations"), m.h.Organization.RemoveMember)
		spaceGroup.GET("/:orgId/members/:userId/check", middleware.HasPermission("read:organizations"), m.h.Organization.IsUserMember)
		spaceGroup.GET("/:orgId/members/:userId/is-owner", middleware.HasPermission("read:organizations"), m.h.Organization.IsUserOwner)
		spaceGroup.GET("/:orgId/members/:userId/role", middleware.HasPermission("read:organizations"), m.h.Organization.GetUserRole)
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
