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
	desc         = "Tenant module"
	version      = "1.0.0"
	dependencies []string
	typeStr      = "module"
	group        = "iam"
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
	exr.RegisterToGroupWithWeakDeps(New(), group, []string{})
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
	m.s = service.New(m.d)
	m.h = handler.New(m.s)

	// Publish service ready event
	m.em.PublishEvent("exts.tenant.ready", map[string]string{
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
