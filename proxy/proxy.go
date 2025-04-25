package proxy

import (
	"fmt"
	"ncobase/proxy/data"
	"ncobase/proxy/handler"
	"ncobase/proxy/service"
	"sync"

	"github.com/ncobase/ncore/config"
	ext "github.com/ncobase/ncore/extension/types"

	"github.com/gin-gonic/gin"
)

var (
	name             = "proxy"
	desc             = "Proxy module for third-party APIs"
	version          = "1.0.0"
	dependencies     []string
	typeStr          = "module"
	group            = "tbp"
	enabledDiscovery = false
)

// Module represents the proxy module.
type Module struct {
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

// New creates a new instance of the proxy module.
func New() ext.Interface {
	return &Module{}
}

// PreInit performs any necessary setup before initialization
func (m *Module) PreInit() error {
	// Implement any pre-initialization logic here
	return nil
}

// Init initializes the proxy module with the given config object
func (m *Module) Init(conf *config.Config, em ext.ManagerInterface) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("proxy module already initialized")
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
	// Service interaction
	_, err := m.getUserService()
	if err != nil {
		return err
	}
	_, err = m.getTenantService()
	if err != nil {
		return err
	}
	_, err = m.getSpaceService()
	if err != nil {
		return err
	}
	_, err = m.getAccessService()
	if err != nil {
		return err
	}
	m.s = service.New(m.conf, m.d)
	m.h = handler.New(m.s)
	return nil
}

// Name returns the name of the module
func (m *Module) Name() string {
	return name
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	// Proxy domain group
	proxyGroup := r.Group("/" + m.Group())

	// Proxy endpoints
	proxyGroup.GET("/endpoints", m.h.Endpoint.List)
	proxyGroup.POST("/endpoints", m.h.Endpoint.Create)
	proxyGroup.GET("/endpoints/:id", m.h.Endpoint.Get)
	proxyGroup.PUT("/endpoints/:id", m.h.Endpoint.Update)
	proxyGroup.DELETE("/endpoints/:id", m.h.Endpoint.Delete)

	// Proxy route management
	proxyGroup.GET("/routes", m.h.Route.List)
	proxyGroup.POST("/routes", m.h.Route.Create)
	proxyGroup.GET("/routes/:id", m.h.Route.Get)
	proxyGroup.PUT("/routes/:id", m.h.Route.Update)
	proxyGroup.DELETE("/routes/:id", m.h.Route.Delete)

	// Proxy transformers
	proxyGroup.GET("/transformers", m.h.Transformer.List)
	proxyGroup.POST("/transformers", m.h.Transformer.Create)
	proxyGroup.GET("/transformers/:id", m.h.Transformer.Get)
	proxyGroup.PUT("/transformers/:id", m.h.Transformer.Update)
	proxyGroup.DELETE("/transformers/:id", m.h.Transformer.Delete)

	// Dynamic proxy routes - these will be registered based on configured endpoints
	dynGroup := r.Group("/proxy")
	m.h.Dynamic.RegisterDynamicRoutes(dynGroup)

	// WebSocket proxy endpoints
	wsGroup := r.Group("/ws")
	m.h.WebSocket.RegisterWebSocketRoutes(wsGroup)
}

// GetHandlers returns the handlers for the module
func (m *Module) GetHandlers() ext.Handler {
	return m.h
}

// GetServices returns the services for the module
func (m *Module) GetServices() ext.Service {
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

// NeedServiceDiscovery returns if the module needs to be registered as a service
func (m *Module) NeedServiceDiscovery() bool {
	return enabledDiscovery
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
