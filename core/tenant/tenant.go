package tenant

import (
	"fmt"
	"ncobase/cmd/ncobase/middleware"
	"ncobase/common/config"
	"ncobase/common/feature"
	"ncobase/core/tenant/data"
	"ncobase/core/tenant/handler"
	"ncobase/core/tenant/service"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	name         = "tenant"
	desc         = "tenant module"
	version      = "1.0.0"
	dependencies []string
	typeStr      = "module"
	group        = "iam"
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
	m.s = service.New(m.d)
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
	r = r.Group("/"+m.Group(), middleware.AuthenticatedTenant)
	// Tenant endpoints
	tenants := r.Group("/tenants", middleware.AuthenticatedTenant)
	{
		tenants.GET("", m.h.Tenant.List)
		tenants.POST("", m.h.Tenant.Create)
		tenants.GET("/:slug", m.h.Tenant.Get)
		tenants.PUT("/:slug", m.h.Tenant.Update)
		tenants.DELETE("/:slug", m.h.Tenant.Delete)

		// // Tenant attachment endpoints
		// tenants.GET("/:tenant/attachments", m.h.ListTenantAttachmentHandler)
		// tenants.POST("/:tenant/attachments", m.h.CreateTenantAttachmentsHandler)
		// tenants.GET("/:tenant/attachments/:attachment", m.h.GetTenantAttachmentHandler)
		// tenants.PUT("/:tenant/attachments/:attachment", m.h.UpdateTenantAttachmentHandler)
		// tenants.DELETE("/:tenant/attachments/:attachment", m.h.DeleteTenantAttachmentHandler)
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
func (m *Module) GetHandlers() feature.Handler {
	return m.h
}

// GetServices returns the services for the module
func (m *Module) GetServices() feature.Service {
	return m.s
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
