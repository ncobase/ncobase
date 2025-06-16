package space

import (
	"fmt"
	"ncobase/internal/middleware"
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
	desc         = "Space module, providing space (space) management, relationship processing"
	version      = "1.0.0"
	dependencies []string
	typeStr      = "module"
	group        = "sys"
)

// Module represents the space module.
type Module struct {
	ext.OptionalImpl

	initialized bool
	mu          sync.RWMutex
	em          ext.ManagerInterface
	h           *handler.Handler
	s           *service.Service
	d           *data.Data
	cleanup     func(n ...string)

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
	exr.RegisterToGroupWithWeakDeps(New(), group, []string{"organization"})
}

// New creates a new instance of the space module.
func New() ext.Interface {
	return &Module{}
}

// Init initializes the space module with the given config object
func (m *Module) Init(conf *config.Config, em ext.ManagerInterface) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("space module already initialized")
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
	m.s = service.New(m.d, m.em) // Pass extension manager
	m.h = handler.New(m.s)

	// Subscribe to extension events for dependency refresh
	m.em.SubscribeEvent("exts.space.ready", func(data any) {
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
	// Belong domain group
	spaceGroup := r.Group("/"+m.Group(), middleware.AuthenticatedSpace)

	// Space endpoints
	spaces := spaceGroup.Group("/spaces", middleware.HasPermission("manage:spaces"), middleware.AuthenticatedSpace)
	{
		// Basic space management
		spaces.GET("", m.h.Space.List)
		spaces.POST("", m.h.Space.Create)
		spaces.GET("/:spaceId", m.h.Space.Get)
		spaces.PUT("/:spaceId", m.h.Space.Update)
		spaces.DELETE("/:spaceId", m.h.Space.Delete)

		// User-Space-Role management
		spaces.GET("/:spaceId/users", middleware.HasPermission("read:spaces"), m.h.UserSpaceRole.ListSpaceUsers)
		spaces.POST("/:spaceId/users/roles", middleware.HasPermission("manage:spaces"), m.h.UserSpaceRole.AddUserToSpaceRole)
		spaces.PUT("/:spaceId/users/roles/bulk", middleware.HasPermission("manage:spaces"), m.h.UserSpaceRole.BulkUpdateUserSpaceRoles)

		// User role management in space
		spaces.GET("/:spaceId/users/:userId/roles", middleware.HasPermission("read:spaces"), m.h.UserSpaceRole.GetUserSpaceRoles)
		spaces.PUT("/:spaceId/users/:userId/roles", middleware.HasPermission("manage:spaces"), m.h.UserSpaceRole.UpdateUserSpaceRole)
		spaces.DELETE("/:spaceId/users/:userId/roles/:roleId", middleware.HasPermission("manage:spaces"), m.h.UserSpaceRole.RemoveUserFromSpaceRole)
		spaces.GET("/:spaceId/users/:userId/roles/:roleId/check", middleware.HasPermission("read:spaces"), m.h.UserSpaceRole.CheckUserSpaceRole)

		// Role-based user queries
		spaces.GET("/:spaceId/roles/:roleId/users", middleware.HasPermission("read:spaces"), m.h.UserSpaceRole.GetSpaceUsersByRole)

		// Space-Group management
		spaces.GET("/:spaceId/orgs", middleware.HasPermission("read:spaces"), m.h.SpaceOrganization.GetSpaceOrganizations)
		spaces.POST("/:spaceId/orgs", middleware.HasPermission("manage:spaces"), m.h.SpaceOrganization.AddGroupToSpace)
		spaces.DELETE("/:spaceId/orgs/:orgId", middleware.HasPermission("manage:spaces"), m.h.SpaceOrganization.RemoveGroupFromSpace)
		spaces.GET("/:spaceId/orgs/:orgId/check", middleware.HasPermission("read:spaces"), m.h.SpaceOrganization.IsGroupInSpace)

		// Space quota management
		spaces.GET("/quotas", m.h.SpaceQuota.List)
		spaces.POST("/quotas", m.h.SpaceQuota.Create)
		spaces.GET("/quotas/:id", m.h.SpaceQuota.Get)
		spaces.PUT("/quotas/:id", m.h.SpaceQuota.Update)
		spaces.DELETE("/quotas/:id", m.h.SpaceQuota.Delete)
		spaces.POST("/quotas/usage", m.h.SpaceQuota.UpdateUsage)
		spaces.GET("/quotas/check", m.h.SpaceQuota.CheckLimit)
		spaces.GET("/:spaceId/quotas", m.h.SpaceQuota.GetSummary)

		// Space settings management
		spaces.GET("/settings", m.h.SpaceSetting.List)
		spaces.POST("/settings", m.h.SpaceSetting.Create)
		spaces.GET("/settings/:id", m.h.SpaceSetting.Get)
		spaces.PUT("/settings/:id", m.h.SpaceSetting.Update)
		spaces.DELETE("/settings/:id", m.h.SpaceSetting.Delete)
		spaces.POST("/settings/bulk", m.h.SpaceSetting.BulkUpdate)
		spaces.GET("/:spaceId/settings", m.h.SpaceSetting.GetSpaceSettings)
		spaces.GET("/:spaceId/settings/public", m.h.SpaceSetting.GetPublicSettings)
		spaces.PUT("/:spaceId/settings/:key", m.h.SpaceSetting.SetSetting)
		spaces.GET("/:spaceId/settings/:key", m.h.SpaceSetting.GetSetting)

		// Space billing management
		spaces.GET("/billing", m.h.SpaceBilling.List)
		spaces.POST("/billing", m.h.SpaceBilling.Create)
		spaces.GET("/billing/:id", m.h.SpaceBilling.Get)
		spaces.PUT("/billing/:id", m.h.SpaceBilling.Update)
		spaces.DELETE("/billing/:id", m.h.SpaceBilling.Delete)
		spaces.POST("/billing/payment", m.h.SpaceBilling.ProcessPayment)
		spaces.GET("/:spaceId/billing/summary", m.h.SpaceBilling.GetSummary)
		spaces.GET("/:spaceId/billing/overdue", m.h.SpaceBilling.GetOverdue)
		spaces.POST("/:spaceId/billing/invoice", m.h.SpaceBilling.GenerateInvoice)

		// Space Menu relations
		spaces.GET("/:spaceId/menus", m.h.SpaceMenu.GetSpaceMenus)
		spaces.POST("/:spaceId/menus", m.h.SpaceMenu.AddMenuToSpace)
		spaces.DELETE("/:spaceId/menus/:menuId", m.h.SpaceMenu.RemoveMenuFromSpace)
		spaces.GET("/:spaceId/menus/:menuId/check", m.h.SpaceMenu.CheckMenuInSpace)

		// Space Dictionary relations
		spaces.GET("/:spaceId/dictionaries", m.h.SpaceDictionary.GetSpaceDictionaries)
		spaces.POST("/:spaceId/dictionaries", m.h.SpaceDictionary.AddDictionaryToSpace)
		spaces.DELETE("/:spaceId/dictionaries/:dictionaryId", m.h.SpaceDictionary.RemoveDictionaryFromSpace)
		spaces.GET("/:spaceId/dictionaries/:dictionaryId/check", m.h.SpaceDictionary.CheckDictionaryInSpace)

		// Space Options relations
		spaces.GET("/:spaceId/options", m.h.SpaceOption.GetSpaceOption)
		spaces.POST("/:spaceId/options", m.h.SpaceOption.AddOptionsToSpace)
		spaces.DELETE("/:spaceId/options/:optionsId", m.h.SpaceOption.RemoveOptionsFromSpace)
		spaces.GET("/:spaceId/options/:optionsId/check", m.h.SpaceOption.CheckOptionsInSpace)
	}

	// User endpoints with space context
	users := spaceGroup.Group("/users", middleware.AuthenticatedUser)
	{
		// User's space ownership
		users.GET("/:username/space", m.h.Space.UserOwn)

		// User's roles across spaces
		users.GET("/:username/spaces/:spaceId/roles", middleware.HasPermission("read:users"), m.h.UserSpaceRole.GetUserSpaceRoles)
		users.GET("/:username/spaces/:spaceId/roles/:roleId/check", middleware.HasPermission("read:users"), m.h.UserSpaceRole.CheckUserSpaceRole)
	}

	// Organization endpoints (cross-module)
	orgs := spaceGroup.Group("/orgs", middleware.AuthenticatedUser)
	{
		// organization space relationships (accessible from both space and space modules)
		orgs.GET("/:orgId/spaces", middleware.HasPermission("read:orgs"), m.h.SpaceOrganization.GetOrganizationSpaces)
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
		{Name: "organization", Type: ext.WeakDependency},
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
