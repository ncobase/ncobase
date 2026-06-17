package access

import (
	"fmt"
	"ncobase/core/access/data"
	"ncobase/core/access/event"
	"ncobase/core/access/handler"
	"ncobase/core/access/service"
	"ncobase/internal/middleware"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	exr "github.com/ncobase/ncore/extension/registry"
	ext "github.com/ncobase/ncore/extension/types"
)

var (
	name         = "access"
	desc         = "Access module"
	version      = "1.0.0"
	dependencies []string
	typeStr      = "module"
	group        = "sys"
)

// Module represents the access module.
type Module struct {
	ext.OptionalImpl

	initialized bool
	mu          sync.RWMutex
	em          ext.ManagerInterface
	conf        *config.Config
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
	exr.RegisterToGroupWithWeakDeps(New(), group, []string{})
}

// New creates a new instance of the access module.
func New() ext.Interface {
	return &Module{}
}

// Init initializes the access module with the given config object
func (m *Module) Init(conf *config.Config, em ext.ManagerInterface) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("access module already initialized")
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
	m.conf = conf
	m.initialized = true

	return nil
}

// PostInit performs any necessary setup after initialization
func (m *Module) PostInit() error {
	m.s = service.New(m.conf, m.d)
	m.h = handler.New(m.s)

	// Register event handlers
	m.registerEventHandlers()

	return nil
}

// Name returns the name of the module
func (m *Module) Name() string {
	return name
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	// Belong domain group
	accessGroup := r.Group("/"+m.Group(), middleware.AuthenticatedUser)

	// Role endpoints - graduated permissions
	roles := accessGroup.Group("/roles")
	{
		roles.GET("", middleware.HasPermission("read:roles"), m.h.Role.List)
		roles.POST("", middleware.HasPermission("manage:roles"), m.h.Role.Create)
		roles.PUT("/bulk", middleware.HasPermission("manage:roles"), m.h.Role.BulkUpdate)
		roles.DELETE("/bulk", middleware.HasPermission("manage:roles"), m.h.Role.BulkDelete)
		roles.GET("/slug/:slug", middleware.HasPermission("read:roles"), m.h.Role.GetBySlug)
		roles.GET("/:slug", middleware.HasPermission("read:roles"), m.h.Role.Get)
		roles.PUT("/:slug", middleware.HasPermission("manage:roles"), m.h.Role.Update)
		roles.DELETE("/:slug", middleware.HasPermission("manage:roles"), m.h.Role.Delete)
		roles.GET("/:slug/permissions", middleware.HasPermission("read:roles"), m.h.RolePermission.ListRolePermission)
		roles.POST("/:slug/permissions", middleware.HasPermission("manage:roles"), m.h.RolePermission.AddPermissionsToRole)
		roles.DELETE("/:slug/permissions", middleware.HasPermission("manage:roles"), m.h.RolePermission.RemovePermissionsFromRole)
		roles.GET("/:slug/users", middleware.HasPermission("read:roles"), m.h.Role.ListUsers)
		roles.POST("/:slug/users", middleware.HasPermission("manage:roles"), m.h.Role.AssignUsers)
		roles.DELETE("/:slug/users", middleware.HasPermission("manage:roles"), m.h.Role.RemoveUsers)
	}

	// Permission endpoints - permission-gated so the menu, token payload, and
	// backend boundary use the same contract.
	permissions := accessGroup.Group("/permissions")
	{
		permissions.GET("", middleware.HasAnyPermission("read:permissions", "manage:permissions"), m.h.Permission.List)
		permissions.POST("", middleware.HasPermission("manage:permissions"), m.h.Permission.Create)
		permissions.PUT("/bulk", middleware.HasPermission("manage:permissions"), m.h.Permission.BulkUpdate)
		permissions.DELETE("/bulk", middleware.HasPermission("manage:permissions"), m.h.Permission.BulkDelete)
		permissions.GET("/:slug", middleware.HasAnyPermission("read:permissions", "manage:permissions"), m.h.Permission.Get)
		permissions.PUT("/:slug", middleware.HasPermission("manage:permissions"), m.h.Permission.Update)
		permissions.DELETE("/:slug", middleware.HasPermission("manage:permissions"), m.h.Permission.Delete)
	}

	// Policy endpoints - super admin only
	policies := accessGroup.Group("/policies", middleware.HasRole("super-admin"))
	{
		policies.GET("", m.h.Casbin.List)
		policies.POST("", m.h.Casbin.Create)
		policies.GET("/by-rule", m.h.Casbin.GetByRule)
		policies.POST("/bulk", m.h.Casbin.BulkCreate)
		policies.POST("/import", m.h.Casbin.Import)
		policies.GET("/export", m.h.Casbin.Export)
		policies.POST("/validate", m.h.Casbin.Validate)
		policies.GET("/:id", m.h.Casbin.Get)
		policies.PUT("/:id", m.h.Casbin.Update)
		policies.DELETE("/:id", m.h.Casbin.Delete)
	}

	// Activity
	activities := accessGroup.Group("/activities")
	{
		activities.POST("", middleware.HasPermission("manage:system"), m.h.Activity.CreateActivity)
		activities.GET("", middleware.HasAnyPermission("read:system", "manage:system"), m.h.Activity.ListActivities)
		activities.GET("/search", middleware.HasAnyPermission("read:system", "manage:system"), m.h.Activity.SearchActivities)
		activities.GET("/users/:username", middleware.HasAnyPermission("read:system", "manage:system"), m.h.Activity.GetUserActivities)
		activities.GET("/analytics", middleware.HasAnyPermission("read:system", "manage:system"), m.h.Activity.GetAnalytics)
		activities.GET("/types", middleware.HasAnyPermission("read:system", "manage:system"), m.h.Activity.GetTypes)
		activities.DELETE("/bulk", middleware.HasPermission("manage:system"), m.h.Activity.BulkDelete)
		activities.GET("/:id", middleware.HasAnyPermission("read:system", "manage:system"), m.h.Activity.GetActivity)
	}

	// User-role relationship endpoints used by the user management console.
	users := accessGroup.Group("/users")
	{
		users.GET("/:username/roles", middleware.HasPermission("read:users"), m.h.UserRole.GetUserRoles)
		users.POST("/:username/roles", middleware.HasPermission("manage:roles"), m.h.UserRole.AssignRoles)
		users.DELETE("/:username/roles", middleware.HasPermission("manage:roles"), m.h.UserRole.RemoveRoles)
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

// registerEventHandlers registers the event handlers
func (m *Module) registerEventHandlers() {
	eventProvider := handler.NewEventProvider(m.s)
	eventRegistrar := event.NewRegistrar(m.em)
	eventRegistrar.RegisterHandlers(eventProvider)
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
