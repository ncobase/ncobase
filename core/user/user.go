package user

import (
	"fmt"
	"ncobase/cmd/ncobase/middleware"
	"ncobase/user/data"
	"ncobase/user/handler"
	"ncobase/user/service"
	"sync"

	"github.com/ncobase/ncore/config"
	exr "github.com/ncobase/ncore/extension/registry"
	ext "github.com/ncobase/ncore/extension/types"

	"github.com/gin-gonic/gin"
)

var (
	name         = "user"
	desc         = "User module"
	version      = "1.0.0"
	dependencies []string
	typeStr      = "module"
	group        = "sys"
)

// Module represents the user module.
type Module struct {
	ext.OptionalImpl

	initialized bool
	mu          sync.RWMutex
	em          ext.ManagerInterface
	conf        *config.Config
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
	exr.RegisterToGroupWithWeakDeps(New(), group, []string{})
}

// New creates a new instance of the user module.
func New() ext.Interface {
	return &Module{}
}

// Init initializes the user module with the given config object
func (m *Module) Init(conf *config.Config, em ext.ManagerInterface) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("user module already initialized")
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
	m.conf = conf
	m.initialized = true

	return nil
}

// PostInit performs any necessary setup after initialization
func (m *Module) PostInit() error {
	m.s = service.New(m.em, m.d)
	m.h = handler.New(m.s)

	return nil
}

// Name returns the name of the module
func (m *Module) Name() string {
	return name
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	// Belong domain group
	userGroup := r.Group("/"+m.Group(), middleware.AuthenticatedUser)
	// User endpoints
	users := userGroup.Group("/users")
	{
		users.GET("", middleware.HasPermission("read:users"), m.h.User.List)
		users.POST("", middleware.HasPermission("create:users"), m.h.User.Create)
		users.GET("/filter", middleware.HasPermission("read:users"), m.h.User.GetFiltered)
		users.GET("/me", middleware.HasPermission("read:users"), m.h.User.GetCurrentUser)
		users.GET("/by-email/:email", middleware.HasPermission("read:users"), m.h.User.GetByEmail)
		users.GET("/by-username/:username", middleware.HasPermission("read:users"), m.h.User.GetByUsername)
		users.GET("/:username", middleware.HasPermission("read:users"), m.h.User.Get)
		users.PUT("/:username", middleware.HasPermission("update:users"), m.h.User.Update)
		users.DELETE("/:username", middleware.HasPermission("delete:users"), m.h.User.Delete)
		users.PUT("/:username/password", middleware.HasPermission("update:users"), m.h.User.UpdatePassword)
		users.PATCH("/:username/status", middleware.HasPermission("update:users"), m.h.User.UpdateStatus)
		users.GET("/:username/profile", middleware.HasAnyPermission("read:users", "manage:profile"), m.h.UserProfile.Get)
		users.PUT("/:username/profile", middleware.HasAnyPermission("update:users", "manage:profile"), m.h.UserProfile.Update)
		users.POST("/reset-password", middleware.HasPermission("update:users"), m.h.User.ResetPassword)
		users.GET("/:username/api-keys", middleware.HasPermission("read:users"), m.h.ApiKey.GetUserApiKeys)
		users.GET("/me/api-keys", middleware.HasPermission("read:users"), m.h.ApiKey.GetMyApiKeys)
		users.POST("/api-keys", middleware.HasPermission("create:users"), m.h.ApiKey.GenerateApiKey)
		users.GET("/api-keys/:id", middleware.HasPermission("read:users"), m.h.ApiKey.GetApiKey)
		users.DELETE("/api-keys/:id", middleware.HasPermission("delete:users"), m.h.ApiKey.DeleteApiKey)
		users.GET("/:username/meshes", middleware.HasAnyPermission("read:users", "manage:profile"), m.h.UserMeshes.GetUserMeshes)
		users.PUT("/:username/meshes", middleware.HasAnyPermission("update:users", "manage:profile"), m.h.UserMeshes.UpdateUserMeshes)
	}

	// Employee endpoints
	employees := userGroup.Group("/employees")
	{
		employees.GET("", middleware.HasPermission("read:employees"), m.h.Employee.List)
		employees.POST("", middleware.HasAnyPermission("create:employees", "manage:hr"), m.h.Employee.Create)
		employees.GET("/:user_id", middleware.HasPermission("read:employees"), m.h.Employee.Get)
		employees.PUT("/:user_id", middleware.HasAnyPermission("update:employees", "manage:hr"), m.h.Employee.Update)
		employees.DELETE("/:user_id", middleware.HasPermission("manage:employees"), m.h.Employee.Delete)
		employees.GET("/department/:department", middleware.HasPermission("read:employees"), m.h.Employee.GetByDepartment)
		employees.GET("/manager/:manager_id", middleware.HasPermission("read:employees"), m.h.Employee.GetByManager)
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
