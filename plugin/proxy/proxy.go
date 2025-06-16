package proxy

import (
	"fmt"
	"ncobase/proxy/data"
	"ncobase/proxy/event"
	"ncobase/proxy/handler"
	"ncobase/proxy/service"
	"sync"

	accessService "ncobase/access/service"
	orgService "ncobase/organization/service"
	spaceService "ncobase/space/service"
	userService "ncobase/user/service"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	extp "github.com/ncobase/ncore/extension/plugin"
	ext "github.com/ncobase/ncore/extension/types"
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
	spaceService  *spaceService.Service
	orgService    *orgService.Service
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
	extp.RegisterPlugin(New(), ext.Metadata{
		Name:         name,
		Version:      version,
		Dependencies: dependencies,
		Description:  desc,
		Type:         typeStr,
		Group:        group,
	})
}

// New returns a new instance of the plugin
func New() *Plugin {
	return &Plugin{}
}

// Init initializes the proxy plugin with the given config object
func (p *Plugin) Init(conf *config.Config, em ext.ManagerInterface) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.initialized {
		return fmt.Errorf("proxy plugin already initialized")
	}

	p.d, p.cleanup, err = data.New(conf.Data, conf.Environment)
	if err != nil {
		return err
	}

	// service discovery
	if conf.Consul != nil {
		p.discovery.address = conf.Consul.Address
		p.discovery.tags = conf.Consul.Discovery.DefaultTags
		p.discovery.meta = conf.Consul.Discovery.DefaultMeta
	}

	p.em = em
	p.conf = conf
	p.initialized = true

	return nil
}

// PostInit performs any necessary setup after initialization
func (p *Plugin) PostInit() error {
	// Get internal services
	var err error

	// Get user service
	p.userService, err = p.getUserService()
	if err != nil {
		return err
	}

	// Get space service
	p.spaceService, err = p.getSpaceService()
	if err != nil {
		return err
	}

	// Get organization service
	p.orgService, err = p.getOrganizationService()
	if err != nil {
		return err
	}

	// Get access service
	p.accessService, err = p.getAccessService()
	if err != nil {
		return err
	}

	// Setup event system
	p.setupEventSystem()

	// Initialize services
	p.s = service.New(p.d)

	// Set event publisher in processor service
	if processorSvc, ok := p.s.Processor.(service.ProcessorServiceInterface); ok {
		processorSvc.SetEventPublisher(p.eventPublisher)
	}

	p.h = handler.New(p.s)

	// Register event handlers
	p.registerEventHandlers()

	// Initialize event subscribers
	if p.eventSubscriber != nil {
		p.eventSubscriber.Initialize(p.em)
	}

	return nil
}

// setupEventSystem initializes event system components
func (p *Plugin) setupEventSystem() {
	// Create event components
	p.eventPublisher = event.NewPublisher(p.em)
	p.eventRegistrar = event.NewRegistrar(p.em)
	p.eventSubscriber = event.NewSubscriber(p.em, p.eventPublisher)

	// Initialize event handlers
	p.eventHandlers = handler.NewEventProvider()
}

// registerEventHandlers registers event handlers with the event system
func (p *Plugin) registerEventHandlers() {
	if p.eventRegistrar != nil && p.eventHandlers != nil {
		p.eventRegistrar.RegisterHandlers(p.eventHandlers)
	}
}

// Name returns the name of the plugin
func (p *Plugin) Name() string {
	return name
}

// RegisterRoutes registers routes for the plugin
func (p *Plugin) RegisterRoutes(r *gin.RouterGroup) {
	// Proxy domain group
	proxyGroup := r.Group("/" + p.Group())

	// Proxy endpoints
	proxyGroup.GET("/endpoints", p.h.Endpoint.List)
	proxyGroup.POST("/endpoints", p.h.Endpoint.Create)
	proxyGroup.GET("/endpoints/:id", p.h.Endpoint.Get)
	proxyGroup.PUT("/endpoints/:id", p.h.Endpoint.Update)
	proxyGroup.DELETE("/endpoints/:id", p.h.Endpoint.Delete)

	// Proxy route management
	proxyGroup.GET("/routes", p.h.Route.List)
	proxyGroup.POST("/routes", p.h.Route.Create)
	proxyGroup.GET("/routes/:id", p.h.Route.Get)
	proxyGroup.PUT("/routes/:id", p.h.Route.Update)
	proxyGroup.DELETE("/routes/:id", p.h.Route.Delete)

	// Proxy transformers
	proxyGroup.GET("/transformers", p.h.Transformer.List)
	proxyGroup.POST("/transformers", p.h.Transformer.Create)
	proxyGroup.GET("/transformers/:id", p.h.Transformer.Get)
	proxyGroup.PUT("/transformers/:id", p.h.Transformer.Update)
	proxyGroup.DELETE("/transformers/:id", p.h.Transformer.Delete)

	// Dynamic proxy routes - these will be registered based on configured endpoints
	dynGroup := r.Group("/proxy")
	p.h.Dynamic.RegisterDynamicRoutes(dynGroup)

	// WebSocket proxy endpoints
	wsGroup := r.Group("/ws")
	p.h.WebSocket.RegisterWebSocketRoutes(wsGroup)
}

// GetHandlers returns the handlers for the plugin
func (p *Plugin) GetHandlers() ext.Handler {
	return p.h
}

// GetServices returns the services for the plugin
func (p *Plugin) GetServices() ext.Service {
	return p.s
}

// Cleanup cleans up the plugin
func (p *Plugin) Cleanup() error {
	if p.cleanup != nil {
		p.cleanup(p.Name())
	}
	return nil
}

// GetMetadata returns the metadata of the plugin
func (p *Plugin) GetMetadata() ext.Metadata {
	return ext.Metadata{
		Name:         p.Name(),
		Version:      p.Version(),
		Dependencies: p.Dependencies(),
		Description:  p.Description(),
		Type:         p.Type(),
		Group:        p.Group(),
	}
}

// Version returns the version of the plugin
func (p *Plugin) Version() string {
	return version
}

// Dependencies returns the dependencies of the plugin
func (p *Plugin) Dependencies() []string {
	return dependencies
}

// GetAllDependencies returns all dependencies of the plugin
func (p *Plugin) GetAllDependencies() []ext.DependencyEntry {
	return []ext.DependencyEntry{
		{Name: "user", Type: ext.WeakDependency},
		{Name: "space", Type: ext.WeakDependency},
		{Name: "organization", Type: ext.WeakDependency},
		{Name: "access", Type: ext.WeakDependency},
	}
}

// Description returns the description of the plugin
func (p *Plugin) Description() string {
	return desc
}

// Type returns the type of the plugin
func (p *Plugin) Type() string {
	return typeStr
}

// Group returns the domain group of the plugin belongs
func (p *Plugin) Group() string {
	return group
}

// GetServiceInfo returns service registration info if NeedServiceDiscovery returns true
func (p *Plugin) GetServiceInfo() *ext.ServiceInfo {
	if !p.NeedServiceDiscovery() {
		return nil
	}

	metadata := p.GetMetadata()

	tags := append(p.discovery.tags, metadata.Group, metadata.Type)

	meta := make(map[string]string)
	for k, v := range p.discovery.meta {
		meta[k] = v
	}
	meta["name"] = metadata.Name
	meta["version"] = metadata.Version
	meta["group"] = metadata.Group
	meta["type"] = metadata.Type
	meta["description"] = metadata.Description

	return &ext.ServiceInfo{
		Address: p.discovery.address,
		Tags:    tags,
		Meta:    meta,
	}
}
