package proxy

import (
	"context"
	"fmt"
	"ncobase/proxy/data"
	"ncobase/proxy/event"
	"ncobase/proxy/handler"
	"ncobase/proxy/service"
	"sync"

	accessService "ncobase/access/service"
	spaceService "ncobase/space/service"
	tenantService "ncobase/tenant/service"
	userService "ncobase/user/service"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	extp "github.com/ncobase/ncore/extension/plugin"
	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

var (
	name         = "proxy"
	desc         = "Proxy plugin for third-party APIs"
	version      = "1.0.0"
	dependencies []string
	typeStr      = "plugin"
	group        = "tbp"
)

// Plugin represents the proxy plugin.
type Plugin struct {
	ext.OptionalImpl

	initialized bool
	mu          sync.RWMutex
	em          ext.ManagerInterface
	conf        *config.Config
	h           *handler.Handler
	s           *service.Service
	d           *data.Data
	cleanup     func(name ...string)

	// Internal services
	userService   *userService.Service
	tenantService *tenantService.Service
	spaceService  *spaceService.Service
	accessService *accessService.Service

	// Event components
	eventPublisher  *event.Publisher
	eventSubscriber *event.Subscriber
	eventRegistrar  *event.Registrar
	eventHandlers   event.HandlerProvider

	discovery
}

// discovery represents the service discovery
type discovery struct {
	address string
	tags    []string
	meta    map[string]string
}

// init registers the plugin
func init() {
	extp.RegisterPlugin(&Plugin{}, ext.Metadata{
		Name:         name,
		Version:      version,
		Dependencies: dependencies,
		Description:  desc,
		Type:         typeStr,
		Group:        group,
	})
}

// Init initializes the proxy plugin with the given config object
func (m *Plugin) Init(conf *config.Config, em ext.ManagerInterface) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("proxy plugin already initialized")
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
func (m *Plugin) PostInit() error {
	// Get internal services
	var err error

	// Get user service
	m.userService, err = m.getUserService()
	if err != nil {
		return err
	}

	// Get tenant service
	m.tenantService, err = m.getTenantService()
	if err != nil {
		return err
	}

	// Get space service
	m.spaceService, err = m.getSpaceService()
	if err != nil {
		return err
	}

	// Get access service
	m.accessService, err = m.getAccessService()
	if err != nil {
		return err
	}

	// Setup event system
	m.setupEventSystem()

	// Initialize services
	m.s = service.New(m.d)

	// Set event publisher in processor service
	if processorSvc, ok := m.s.Processor.(service.ProcessorServiceInterface); ok {
		processorSvc.SetEventPublisher(m.eventPublisher)
	}

	m.h = handler.New(m.s)

	// Register event handlers
	m.registerEventHandlers()

	// Initialize event subscribers
	if m.eventSubscriber != nil {
		m.eventSubscriber.Initialize(m.em)
	}

	return nil
}

// setupEventSystem initializes event system components
func (m *Plugin) setupEventSystem() {
	// Create event components
	m.eventPublisher = event.NewPublisher(m.em)
	m.eventRegistrar = event.NewRegistrar(m.em)
	m.eventSubscriber = event.NewSubscriber(
		m.userService,
		m.tenantService,
		m.spaceService,
		m.accessService,
		m.eventPublisher,
	)

	// Initialize event handlers
	m.eventHandlers = handler.NewEventProvider()

	logger.Info(context.Background(), "Event system initialized for proxy plugin")
}

// registerEventHandlers registers event handlers with the event system
func (m *Plugin) registerEventHandlers() {
	if m.eventRegistrar != nil && m.eventHandlers != nil {
		m.eventRegistrar.RegisterHandlers(m.eventHandlers)
	}
}

// Name returns the name of the plugin
func (m *Plugin) Name() string {
	return name
}

// RegisterRoutes registers routes for the plugin
func (m *Plugin) RegisterRoutes(r *gin.RouterGroup) {
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

// GetHandlers returns the handlers for the plugin
func (m *Plugin) GetHandlers() ext.Handler {
	return m.h
}

// GetServices returns the services for the plugin
func (m *Plugin) GetServices() ext.Service {
	return m.s
}

// Cleanup cleans up the plugin
func (m *Plugin) Cleanup() error {
	if m.cleanup != nil {
		m.cleanup(m.Name())
	}
	return nil
}

// GetMetadata returns the metadata of the plugin
func (m *Plugin) GetMetadata() ext.Metadata {
	return ext.Metadata{
		Name:         m.Name(),
		Version:      m.Version(),
		Dependencies: m.Dependencies(),
		Description:  m.Description(),
		Type:         m.Type(),
		Group:        m.Group(),
	}
}

// Version returns the version of the plugin
func (m *Plugin) Version() string {
	return version
}

// Dependencies returns the dependencies of the plugin
func (m *Plugin) Dependencies() []string {
	return dependencies
}

// Description returns the description of the plugin
func (m *Plugin) Description() string {
	return desc
}

// Type returns the type of the plugin
func (m *Plugin) Type() string {
	return typeStr
}

// Group returns the domain group of the plugin belongs
func (m *Plugin) Group() string {
	return group
}

// GetServiceInfo returns service registration info if NeedServiceDiscovery returns true
func (m *Plugin) GetServiceInfo() *ext.ServiceInfo {
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
