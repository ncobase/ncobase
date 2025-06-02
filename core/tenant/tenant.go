package tenant

import (
	"fmt"
	"ncobase/cmd/ncobase/middleware"
	"ncobase/tenant/data"
	"ncobase/tenant/handler"
	"ncobase/tenant/service"
	"sync"

	"github.com/ncobase/ncore/config"
	exr "github.com/ncobase/ncore/extension/registry"
	ext "github.com/ncobase/ncore/extension/types"

	"github.com/gin-gonic/gin"
)

var (
	name         = "tenant"
	desc         = "Tenant module with user role management and group relationships"
	version      = "1.0.0"
	dependencies []string
	typeStr      = "module"
	group        = "sys"
)

// Module represents the tenant module.
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
	exr.RegisterToGroupWithWeakDeps(New(), group, []string{"space"})
}

// New creates a new instance of the tenant module.
func New() ext.Interface {
	return &Module{}
}

// Init initializes the tenant module with the given config object
func (m *Module) Init(conf *config.Config, em ext.ManagerInterface) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("tenant module already initialized")
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
	tenantGroup := r.Group("/"+m.Group(), middleware.AuthenticatedTenant)

	// Tenant endpoints
	tenants := tenantGroup.Group("/tenants", middleware.HasPermission("manage:tenants"), middleware.AuthenticatedTenant)
	{
		// Basic tenant management
		tenants.GET("", m.h.Tenant.List)
		tenants.POST("", m.h.Tenant.Create)
		tenants.GET("/:tenantId", m.h.Tenant.Get)
		tenants.PUT("/:tenantId", m.h.Tenant.Update)
		tenants.DELETE("/:tenantId", m.h.Tenant.Delete)

		// User-Tenant-Role management
		tenants.GET("/:tenantId/users", middleware.HasPermission("read:tenants"), m.h.UserTenantRole.ListTenantUsers)
		tenants.POST("/:tenantId/users/roles", middleware.HasPermission("manage:tenants"), m.h.UserTenantRole.AddUserToTenantRole)
		tenants.PUT("/:tenantId/users/roles/bulk", middleware.HasPermission("manage:tenants"), m.h.UserTenantRole.BulkUpdateUserTenantRoles)

		// User role management in tenant
		tenants.GET("/:tenantId/users/:userId/roles", middleware.HasPermission("read:tenants"), m.h.UserTenantRole.GetUserTenantRoles)
		tenants.PUT("/:tenantId/users/:userId/roles", middleware.HasPermission("manage:tenants"), m.h.UserTenantRole.UpdateUserTenantRole)
		tenants.DELETE("/:tenantId/users/:userId/roles/:roleId", middleware.HasPermission("manage:tenants"), m.h.UserTenantRole.RemoveUserFromTenantRole)
		tenants.GET("/:tenantId/users/:userId/roles/:roleId/check", middleware.HasPermission("read:tenants"), m.h.UserTenantRole.CheckUserTenantRole)

		// Role-based user queries
		tenants.GET("/:tenantId/roles/:roleId/users", middleware.HasPermission("read:tenants"), m.h.UserTenantRole.GetTenantUsersByRole)

		// Tenant-Group management
		tenants.GET("/:tenantId/groups", middleware.HasPermission("read:tenants"), m.h.TenantGroup.GetTenantGroups)
		tenants.POST("/:tenantId/groups", middleware.HasPermission("manage:tenants"), m.h.TenantGroup.AddGroupToTenant)
		tenants.DELETE("/:tenantId/groups/:groupId", middleware.HasPermission("manage:tenants"), m.h.TenantGroup.RemoveGroupFromTenant)
		tenants.GET("/:tenantId/groups/:groupId/check", middleware.HasPermission("read:tenants"), m.h.TenantGroup.IsGroupInTenant)

		// Tenant quota management
		tenants.GET("/quotas", m.h.TenantQuota.List)
		tenants.POST("/quotas", m.h.TenantQuota.Create)
		tenants.GET("/quotas/:id", m.h.TenantQuota.Get)
		tenants.PUT("/quotas/:id", m.h.TenantQuota.Update)
		tenants.DELETE("/quotas/:id", m.h.TenantQuota.Delete)
		tenants.POST("/quotas/usage", m.h.TenantQuota.UpdateUsage)
		tenants.GET("/quotas/check", m.h.TenantQuota.CheckLimit)
		tenants.GET("/:tenantId/quotas", m.h.TenantQuota.GetSummary)

		// Tenant settings management
		tenants.GET("/settings", m.h.TenantSetting.List)
		tenants.POST("/settings", m.h.TenantSetting.Create)
		tenants.GET("/settings/:id", m.h.TenantSetting.Get)
		tenants.PUT("/settings/:id", m.h.TenantSetting.Update)
		tenants.DELETE("/settings/:id", m.h.TenantSetting.Delete)
		tenants.POST("/settings/bulk", m.h.TenantSetting.BulkUpdate)
		tenants.GET("/:tenantId/settings", m.h.TenantSetting.GetTenantSettings)
		tenants.GET("/:tenantId/settings/public", m.h.TenantSetting.GetPublicSettings)
		tenants.PUT("/:tenantId/settings/:key", m.h.TenantSetting.SetSetting)
		tenants.GET("/:tenantId/settings/:key", m.h.TenantSetting.GetSetting)

		// Tenant billing management
		tenants.GET("/billing", m.h.TenantBilling.List)
		tenants.POST("/billing", m.h.TenantBilling.Create)
		tenants.GET("/billing/:id", m.h.TenantBilling.Get)
		tenants.PUT("/billing/:id", m.h.TenantBilling.Update)
		tenants.DELETE("/billing/:id", m.h.TenantBilling.Delete)
		tenants.POST("/billing/payment", m.h.TenantBilling.ProcessPayment)
		tenants.GET("/:tenantId/billing/summary", m.h.TenantBilling.GetSummary)
		tenants.GET("/:tenantId/billing/overdue", m.h.TenantBilling.GetOverdue)
		tenants.POST("/:tenantId/billing/invoice", m.h.TenantBilling.GenerateInvoice)

		// Tenant Menu relations
		tenants.GET("/:tenantId/menus", m.h.TenantMenu.GetTenantMenus)
		tenants.POST("/:tenantId/menus", m.h.TenantMenu.AddMenuToTenant)
		tenants.DELETE("/:tenantId/menus/:menuId", m.h.TenantMenu.RemoveMenuFromTenant)
		tenants.GET("/:tenantId/menus/:menuId/check", m.h.TenantMenu.CheckMenuInTenant)

		// Tenant Dictionary relations
		tenants.GET("/:tenantId/dictionaries", m.h.TenantDictionary.GetTenantDictionaries)
		tenants.POST("/:tenantId/dictionaries", m.h.TenantDictionary.AddDictionaryToTenant)
		tenants.DELETE("/:tenantId/dictionaries/:dictionaryId", m.h.TenantDictionary.RemoveDictionaryFromTenant)
		tenants.GET("/:tenantId/dictionaries/:dictionaryId/check", m.h.TenantDictionary.CheckDictionaryInTenant)

		// Tenant Options relations
		tenants.GET("/:tenantId/options", m.h.TenantOption.GetTenantOption)
		tenants.POST("/:tenantId/options", m.h.TenantOption.AddOptionsToTenant)
		tenants.DELETE("/:tenantId/options/:optionsId", m.h.TenantOption.RemoveOptionsFromTenant)
		tenants.GET("/:tenantId/options/:optionsId/check", m.h.TenantOption.CheckOptionsInTenant)
	}

	// User endpoints with tenant context
	users := tenantGroup.Group("/users", middleware.AuthenticatedUser)
	{
		// User's tenant ownership
		users.GET("/:username/tenant", m.h.Tenant.UserOwn)

		// User's roles across tenants
		users.GET("/:username/tenants/:tenantId/roles", middleware.HasPermission("read:users"), m.h.UserTenantRole.GetUserTenantRoles)
		users.GET("/:username/tenants/:tenantId/roles/:roleId/check", middleware.HasPermission("read:users"), m.h.UserTenantRole.CheckUserTenantRole)
	}

	// Group endpoints (cross-module)
	groups := tenantGroup.Group("/groups", middleware.AuthenticatedUser)
	{
		// Group tenant relationships (accessible from both space and tenant modules)
		groups.GET("/:groupId/tenants", middleware.HasPermission("read:groups"), m.h.TenantGroup.GetGroupTenants)
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
		{Name: "space", Type: ext.WeakDependency},
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
