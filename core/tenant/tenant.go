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
	tenants := tenantGroup.Group("/tenants", middleware.AuthenticatedTenant)
	{
		tenants.GET("", m.h.Tenant.List)
		tenants.POST("", m.h.Tenant.Create)
		tenants.GET("/:slug", m.h.Tenant.Get)
		tenants.PUT("/:slug", m.h.Tenant.Update)
		tenants.DELETE("/:slug", m.h.Tenant.Delete)

		// Tenant quota management
		tenants.GET("/quotas", m.h.TenantQuota.List)
		tenants.POST("/quotas", m.h.TenantQuota.Create)
		tenants.GET("/quotas/:id", m.h.TenantQuota.Get)
		tenants.PUT("/quotas/:id", m.h.TenantQuota.Update)
		tenants.DELETE("/quotas/:id", m.h.TenantQuota.Delete)
		tenants.POST("/quotas/usage", m.h.TenantQuota.UpdateUsage)
		tenants.GET("/quotas/check", m.h.TenantQuota.CheckLimit)
		tenants.GET("/:slug/quotas", m.h.TenantQuota.GetSummary)

		// Tenant settings management
		tenants.GET("/settings", m.h.TenantSetting.List)
		tenants.POST("/settings", m.h.TenantSetting.Create)
		tenants.GET("/settings/:id", m.h.TenantSetting.Get)
		tenants.PUT("/settings/:id", m.h.TenantSetting.Update)
		tenants.DELETE("/settings/:id", m.h.TenantSetting.Delete)
		tenants.POST("/settings/bulk", m.h.TenantSetting.BulkUpdate)
		tenants.GET("/:slug/settings", m.h.TenantSetting.GetTenantSettings)
		tenants.GET("/:slug/settings/public", m.h.TenantSetting.GetPublicSettings)
		tenants.PUT("/:slug/settings/:key", m.h.TenantSetting.SetSetting)
		tenants.GET("/:slug/settings/:key", m.h.TenantSetting.GetSetting)

		// Tenant billing management
		tenants.GET("/billing", m.h.TenantBilling.List)
		tenants.POST("/billing", m.h.TenantBilling.Create)
		tenants.GET("/billing/:id", m.h.TenantBilling.Get)
		tenants.PUT("/billing/:id", m.h.TenantBilling.Update)
		tenants.DELETE("/billing/:id", m.h.TenantBilling.Delete)
		tenants.POST("/billing/payment", m.h.TenantBilling.ProcessPayment)
		tenants.GET("/:slug/billing/summary", m.h.TenantBilling.GetSummary)
		tenants.GET("/:slug/billing/overdue", m.h.TenantBilling.GetOverdue)
		tenants.POST("/:slug/billing/invoice", m.h.TenantBilling.GenerateInvoice)
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
