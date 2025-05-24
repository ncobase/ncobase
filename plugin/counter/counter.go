package counter

import (
	"fmt"
	"ncobase/cmd/ncobase/middleware"
	"ncobase/plugin/counter/data"
	"ncobase/plugin/counter/data/repository"
	"ncobase/plugin/counter/handler"
	"ncobase/plugin/counter/service"
	"sync"

	"github.com/ncobase/ncore/config"
	extp "github.com/ncobase/ncore/extension/plugin"
	ext "github.com/ncobase/ncore/extension/types"

	"github.com/gin-gonic/gin"
)

var (
	name         = "counter"
	desc         = "Counter plugin, built-in"
	version      = "1.0.0"
	dependencies []string
	typeStr      = "plugin"
	group        = "plug"
)

// Plugin represents the counter plugin
type Plugin struct {
	ext.OptionalImpl

	initialized bool
	mu          sync.RWMutex
	em          ext.ManagerInterface
	conf        *config.Config
	d           *data.Data
	r           *repository.Repository
	s           *service.Service
	h           *handler.Handler
	cleanup     func(name ...string)

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

// Name returns the name of the plugin
func (p *Plugin) Name() string {
	return name
}

// Init initializes the plugin
func (p *Plugin) Init(conf *config.Config, em ext.ManagerInterface) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.initialized {
		return fmt.Errorf("counter plugin already initialized")
	}

	p.d, p.cleanup, err = data.New(conf.Data)
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
	p.r = repository.New(p.d)
	p.s = service.New(p.d)
	p.h = handler.New(p.s)
	// Subscribe to relevant events
	p.subscribeEvents(p.em)

	// Publish own plugin ready event
	p.em.PublishEvent("exts.counter.ready", map[string]string{
		"name":   p.Name(),
		"status": "ready",
	})

	return nil
}

// RegisterRoutes registers routes for the plugin
func (p *Plugin) RegisterRoutes(r *gin.RouterGroup) {
	// Belong domain group
	r = r.Group("/"+p.Group(), middleware.AuthenticatedUser)
	// Counter endpoints
	counters := r.Group("/counters")
	{
		counters.GET("", p.h.Counter.List)
		counters.POST("", p.h.Counter.Create)
		counters.GET("/:id", p.h.Counter.Get)
		counters.PUT("/:id", p.h.Counter.Update)
		counters.DELETE("/:id", p.h.Counter.Delete)
	}
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

// SubscribeEvents subscribes to relevant events
func (p *Plugin) subscribeEvents(_ ext.ManagerInterface) {
	// Implement any event subscriptions here
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
