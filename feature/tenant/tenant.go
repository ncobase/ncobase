package tenant

import (
	"fmt"
	"ncobase/common/config"
	"ncobase/feature"
	"ncobase/feature/tenant/data"
	"ncobase/feature/tenant/handler"
	"ncobase/feature/tenant/middleware"
	"ncobase/feature/tenant/service"
	userService "ncobase/feature/user/service"
	"sync"

	accessService "ncobase/feature/access/service"

	"github.com/gin-gonic/gin"
)

var (
	name         = "tenant"
	desc         = "tenant module"
	version      = "1.0.0"
	dependencies = []string{"user", "access"}
)

// Module represents the tenant module.
type Module struct {
	initialized bool
	mu          sync.RWMutex
	fm          *feature.Manager
	h           *handler.Handler
	s           *service.Service
	d           *data.Data
	cleanup     func(n ...string)
}

// New creates a new instance of the tenant module.
func New() feature.Interface {
	return &Module{}
}

// PreInit performs any necessary setup before initialization
func (m *Module) PreInit() error {
	// Implement any pre-initialization logic here
	return nil
}

// Init initializes the tenant module with the given config object
func (m *Module) Init(conf *config.Config, fm *feature.Manager) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("tenant module already initialized")
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
	// get user service
	usi, err := m.getUserService(m.fm)
	if err != nil {
		return err
	}
	// get role service
	rsi, err := m.getRoleService(m.fm)
	if err != nil {
		return err
	}
	// get user role service
	ursi, err := m.getUserRoleService(m.fm)
	if err != nil {
		return err
	}

	// get user tenant role service
	utrsi, err := m.getUserTenantRoleService(m.fm)
	if err != nil {
		return err
	}

	m.s = service.New(m.d, usi, rsi, ursi, utrsi)
	m.h = handler.New(m.s)

	return nil
}

// Name returns the name of the module
func (m *Module) Name() string {
	return name
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(e *gin.Engine) {

	// middleware
	e.Use(middleware.ConsumeTenant(m.s))

	// API v1 endpoints
	v1 := e.Group("/v1")
	// Tenant endpoints
	tenants := v1.Group("/tenants", middleware.AuthenticatedTenant)
	{
		tenants.GET("", m.h.Tenant.ListTenantHandler)
		tenants.POST("", m.h.Tenant.CreateTenantHandler)
		tenants.GET("/:slug", m.h.Tenant.GetTenantHandler)
		tenants.PUT("/:slug", m.h.Tenant.UpdateTenantHandler)
		tenants.DELETE("/:slug", m.h.Tenant.DeleteTenantHandler)

		// // Tenant asset endpoints
		// tenants.GET("/:tenant/assets", m.h.ListTenantAssetHandler)
		// tenants.POST("/:tenant/assets", m.h.CreateTenantAssetsHandler)
		// tenants.GET("/:tenant/assets/:asset", m.h.GetTenantAssetHandler)
		// tenants.PUT("/:tenant/assets/:asset", m.h.UpdateTenantAssetHandler)
		// tenants.DELETE("/:tenant/assets/:asset", m.h.DeleteTenantAssetHandler)
		//
		// // // Tenant role endpoints
		// // tenants.GET("/:tenant/roles", m.h.ListTenantRoleHandler)
		// // tenants.POST("/:tenant/roles", m.h.CreateTenantRoleHandler)
		// // tenants.GET("/:tenant/roles/:role", m.h.GetTenantRoleHandler)
		// // tenants.PUT("/:tenant/roles/:role", m.h.UpdateTenantRoleHandler)
		// // tenants.DELETE("/:tenant/roles/:role", m.h.DeleteTenantRoleHandler)
		// // tenants.GET("/:tenant/roles/:roleSlug/permissions", m.h.ListTenantRolePermissionHandler)
		// // tenants.GET("/:tenant/roles/:roleSlug/users", m.h.ListTenantRoleUserHandler)
		// //
		// // // Tenant permission endpoints
		// // tenants.GET("/:tenant/permissions", m.h.ListTenantPermissionHandler)
		// // tenants.POST("/:tenant/permissions", m.h.CreateTenantPermissionHandler)
		// // tenants.GET("/:tenant/permissions/:permission", m.h.GetTenantPermissionHandler)
		// // tenants.PUT("/:tenant/permissions/:permission", m.h.UpdateTenantPermissionHandler)
		// // tenants.DELETE("/:tenant/permissions/:permission", m.h.DeleteTenantPermissionHandler)
		// //
		// // // Tenant module endpoints
		// // tenants.GET("/:tenant/modules", m.h.ListTenantModuleHandler)
		// // tenants.POST("/:tenant/modules", m.h.CreateTenantModuleHandler)
		// // tenants.GET("/:tenant/modules/:module", m.h.GetTenantModuleHandler)
		// // tenants.PUT("/:tenant/modules/:module", m.h.UpdateTenantModuleHandler)
		// // tenants.DELETE("/:tenant/modules/:module", m.h.DeleteTenantModuleHandler)
		// //
		// // Tenant menu endpoints
		// tenants.GET("/:tenant/menus", m.h.ListTenantMenusHandler)
		// tenants.POST("/:tenant/menus", m.h.CreateTenantMenuHandler)
		// tenants.GET("/:tenant/menus/:menu", m.h.GetTenantMenuHandler)
		// tenants.PUT("/:tenant/menus/:menu", m.h.UpdateTenantMenuHandler)
		// tenants.DELETE("/:tenant/menus/:menu", m.h.DeleteTenantMenuHandler)
		// //
		// // // Tenant policy endpoints
		// // tenants.GET("/:tenant/policies", m.h.ListTenantPolicyHandler)
		// // tenants.POST("/:tenant/policies", m.h.CreateTenantPolicyHandler)
		// // tenants.GET("/:tenant/policies/:policyId", m.h.GetTenantPolicyHandler)
		// // tenants.PUT("/:tenant/policies/:policyId", m.h.UpdateTenantPolicyHandler)
		// // tenants.DELETE("/:tenant/policies/:policyId", m.h.DeleteTenantPolicyHandler)
		// //
		// // // Tenant taxonomy endpoints
		// // tenants.GET("/:tenant/taxonomies", m.h.ListTenantTaxonomyHandler)
		// // tenants.POST("/:tenant/taxonomies", m.h.CreateTenantTaxonomyHandler)
		// // tenants.GET("/:tenant/taxonomies/:taxonomy", m.h.GetTenantTaxonomyHandler)
		// // tenants.PUT("/:tenant/taxonomies/:taxonomy", m.h.UpdateTenantTaxonomyHandler)
		// // tenants.DELETE("/:tenant/taxonomies/:taxonomy", m.h.DeleteTenantTaxonomyHandler)
		// //
		// // // Tenant topic endpoints
		// // tenants.GET("/:tenant/topics", m.h.ListTenantTopicHandler)
		// // tenants.POST("/:tenant/topics", m.h.CreateTenantTopicHandler)
		// // tenants.GET("/:tenant/topics/:topic", m.h.GetTenantTopicHandler)
		// // tenants.PUT("/:tenant/topics/:topic", m.h.UpdateTenantTopicHandler)
		// // tenants.DELETE("/:tenant/topics/:topic", m.h.DeleteTenantTopicHandler)
	}
}

// GetHandlers returns the handlers for the module
func (m *Module) GetHandlers() map[string]feature.Handler {
	return map[string]feature.Handler{
		"tenant": m.h.Tenant,
	}
}

// GetServices returns the services for the module
func (m *Module) GetServices() map[string]feature.Service {
	return map[string]feature.Service{
		"tenant":      m.s.Tenant,
		"user_tenant": m.s.UserTenant,
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

// GetUserService returns the user service
func (m *Module) getUserService(fm *feature.Manager) (userService.UserServiceInterface, error) {
	us, err := fm.GetService("user", "user")
	if err != nil {
		return nil, fmt.Errorf("failed to get user service: %v", err)
	}
	usi, ok := us.(userService.UserServiceInterface)
	if !ok {
		return nil, fmt.Errorf("user service does not implement UserServiceInterface")
	}
	return usi, nil
}

// GetRoleService returns the role service
func (m *Module) getRoleService(fm *feature.Manager) (accessService.RoleServiceInterface, error) {
	rs, err := fm.GetService("access", "role")
	if err != nil {
		return nil, fmt.Errorf("failed to get access service: %v", err)
	}
	rsi, ok := rs.(accessService.RoleServiceInterface)
	if !ok {
		return nil, fmt.Errorf("access service does not implement AccessServiceInterface")
	}
	return rsi, nil
}

// GetUserRoleService returns the user role service
func (m *Module) getUserRoleService(fm *feature.Manager) (accessService.UserRoleServiceInterface, error) {
	// get user role service
	ursi, err := fm.GetService("access", "user_role")
	if err != nil {
		return nil, fmt.Errorf("failed to get access service: %v", err)
	}

	// Type assertion to ensure we have the correct interface
	userRoleServiceImpl, ok := ursi.(accessService.UserRoleServiceInterface)
	if !ok {
		return nil, fmt.Errorf("access service does not implement AccessServiceInterface")
	}

	return userRoleServiceImpl, nil
}

// getUserTenantRoleService returns the user tenant role service
func (m *Module) getUserTenantRoleService(fm *feature.Manager) (accessService.UserTenantRoleServiceInterface, error) {
	utrsi, err := fm.GetService("access", "user_tenant_role")
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant service: %v", err)
	}

	// type assertion to ensure we have the correct interface
	userTenantRoleServiceImpl, ok := utrsi.(accessService.UserTenantRoleServiceInterface)
	if !ok {
		return nil, fmt.Errorf("tenant service does not implement TenantServiceInterface")
	}

	return userTenantRoleServiceImpl, nil

}
