package space

import (
	"fmt"
	"ncobase/core/space/data"
	"ncobase/core/space/handler"
	"ncobase/core/space/service"
	"ncobase/internal/middleware"
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

	// Space endpoints. Read and management permissions are applied per route so
	// read-only space operators are not blocked by the whole group.
	spaces := spaceGroup.Group("/spaces")
	{
		// Basic space collection management
		spaces.GET("", middleware.HasPermission("read:spaces"), m.h.Space.List)
		spaces.POST("", middleware.HasPermission("manage:spaces"), m.h.Space.Create)

		// Space quota collection management
		spaces.GET("/quotas", middleware.HasPermission("read:spaces"), m.h.SpaceQuota.List)
		spaces.POST("/quotas", middleware.HasPermission("manage:spaces"), m.h.SpaceQuota.Create)
		spaces.POST("/quotas/usage", middleware.HasPermission("manage:spaces"), m.h.SpaceQuota.UpdateUsage)
		spaces.GET("/quotas/check", middleware.HasPermission("read:spaces"), m.h.SpaceQuota.CheckLimit)
		spaces.GET("/quotas/:id", middleware.HasPermission("read:spaces"), m.h.SpaceQuota.Get)
		spaces.PUT("/quotas/:id", middleware.HasPermission("manage:spaces"), m.h.SpaceQuota.Update)
		spaces.DELETE("/quotas/:id", middleware.HasPermission("manage:spaces"), m.h.SpaceQuota.Delete)

		// Space settings collection management
		spaces.GET("/settings", middleware.HasPermission("read:spaces"), m.h.SpaceSetting.List)
		spaces.POST("/settings", middleware.HasPermission("manage:spaces"), m.h.SpaceSetting.Create)
		spaces.POST("/settings/bulk", middleware.HasPermission("manage:spaces"), m.h.SpaceSetting.BulkUpdate)
		spaces.GET("/settings/:id", middleware.HasPermission("read:spaces"), m.h.SpaceSetting.Get)
		spaces.PUT("/settings/:id", middleware.HasPermission("manage:spaces"), m.h.SpaceSetting.Update)
		spaces.DELETE("/settings/:id", middleware.HasPermission("manage:spaces"), m.h.SpaceSetting.Delete)

		// Space billing collection management
		spaces.GET("/billing", middleware.HasPermission("read:spaces"), m.h.SpaceBilling.List)
		spaces.POST("/billing", middleware.HasPermission("manage:spaces"), m.h.SpaceBilling.Create)
		spaces.POST("/billing/payment", middleware.HasPermission("manage:spaces"), m.h.SpaceBilling.ProcessPayment)
		spaces.GET("/billing/:id", middleware.HasPermission("read:spaces"), m.h.SpaceBilling.Get)
		spaces.PUT("/billing/:id", middleware.HasPermission("manage:spaces"), m.h.SpaceBilling.Update)
		spaces.DELETE("/billing/:id", middleware.HasPermission("manage:spaces"), m.h.SpaceBilling.Delete)

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
		spaces.GET("/:spaceId/roles", middleware.HasPermission("read:spaces"), m.h.Space.ListRoles)

		// Space attachments
		spaces.GET("/:spaceId/attachments", middleware.HasPermission("read:spaces"), m.h.Space.ListAttachments)

		// Space-Group management
		spaces.GET("/:spaceId/orgs", middleware.HasPermission("read:spaces"), m.h.SpaceOrganization.GetSpaceOrganizations)
		spaces.POST("/:spaceId/orgs", middleware.HasPermission("manage:spaces"), m.h.SpaceOrganization.AddGroupToSpace)
		spaces.DELETE("/:spaceId/orgs/:orgId", middleware.HasPermission("manage:spaces"), m.h.SpaceOrganization.RemoveGroupFromSpace)
		spaces.GET("/:spaceId/orgs/:orgId/check", middleware.HasPermission("read:spaces"), m.h.SpaceOrganization.IsGroupInSpace)

		spaces.GET("/:spaceId/quotas", middleware.HasPermission("read:spaces"), m.h.SpaceQuota.GetSummary)

		// Space settings management
		spaces.GET("/:spaceId/settings", middleware.HasPermission("read:spaces"), m.h.SpaceSetting.GetSpaceSettings)
		spaces.GET("/:spaceId/settings/public", middleware.HasPermission("read:spaces"), m.h.SpaceSetting.GetPublicSettings)
		spaces.PUT("/:spaceId/settings/:key", middleware.HasPermission("manage:spaces"), m.h.SpaceSetting.SetSetting)
		spaces.GET("/:spaceId/settings/:key", middleware.HasPermission("read:spaces"), m.h.SpaceSetting.GetSetting)

		// Space billing management
		spaces.GET("/:spaceId/billing/summary", middleware.HasPermission("read:spaces"), m.h.SpaceBilling.GetSummary)
		spaces.GET("/:spaceId/billing/overdue", middleware.HasPermission("read:spaces"), m.h.SpaceBilling.GetOverdue)
		spaces.POST("/:spaceId/billing/invoice", middleware.HasPermission("manage:spaces"), m.h.SpaceBilling.GenerateInvoice)

		// Space Menu relations
		spaces.GET("/:spaceId/menus", middleware.HasPermission("read:spaces"), m.h.SpaceMenu.GetSpaceMenus)
		spaces.POST("/:spaceId/menus", middleware.HasPermission("manage:spaces"), m.h.SpaceMenu.AddMenuToSpace)
		spaces.DELETE("/:spaceId/menus/:menuId", middleware.HasPermission("manage:spaces"), m.h.SpaceMenu.RemoveMenuFromSpace)
		spaces.GET("/:spaceId/menus/:menuId/check", middleware.HasPermission("read:spaces"), m.h.SpaceMenu.CheckMenuInSpace)

		// Space Dictionary relations
		spaces.GET("/:spaceId/dictionaries", middleware.HasPermission("read:spaces"), m.h.SpaceDictionary.GetSpaceDictionaries)
		spaces.POST("/:spaceId/dictionaries", middleware.HasPermission("manage:spaces"), m.h.SpaceDictionary.AddDictionaryToSpace)
		spaces.DELETE("/:spaceId/dictionaries/:dictionaryId", middleware.HasPermission("manage:spaces"), m.h.SpaceDictionary.RemoveDictionaryFromSpace)
		spaces.GET("/:spaceId/dictionaries/:dictionaryId/check", middleware.HasPermission("read:spaces"), m.h.SpaceDictionary.CheckDictionaryInSpace)

		// Space Options relations
		spaces.GET("/:spaceId/options", middleware.HasPermission("read:spaces"), m.h.SpaceOption.GetSpaceOption)
		spaces.POST("/:spaceId/options", middleware.HasPermission("manage:spaces"), m.h.SpaceOption.AddOptionsToSpace)
		spaces.DELETE("/:spaceId/options/:optionsId", middleware.HasPermission("manage:spaces"), m.h.SpaceOption.RemoveOptionsFromSpace)
		spaces.GET("/:spaceId/options/:optionsId/check", middleware.HasPermission("read:spaces"), m.h.SpaceOption.CheckOptionsInSpace)

		// Basic space item management must stay after static collection routes.
		spaces.GET("/:spaceId", middleware.HasPermission("read:spaces"), m.h.Space.Get)
		spaces.PUT("/:spaceId", middleware.HasPermission("manage:spaces"), m.h.Space.Update)
		spaces.DELETE("/:spaceId", middleware.HasPermission("manage:spaces"), m.h.Space.Delete)
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
		orgs.GET("/:orgId/spaces", middleware.HasPermission("read:organizations"), m.h.SpaceOrganization.GetOrganizationSpaces)
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
