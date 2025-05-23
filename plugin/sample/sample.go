package sample

import (
	"fmt"
	"sync"

	"github.com/ncobase/ncore/config"
	extp "github.com/ncobase/ncore/extension/plugin"
	ext "github.com/ncobase/ncore/extension/types"

	"github.com/gin-gonic/gin"
)

var (
	name         = "sample"
	desc         = "Sample plugin"
	version      = "1.0.0"
	dependencies []string
	typeStr      = "plugin"
	group        = "plug"
)

// Plugin represents the sample plugin
type Plugin struct {
	ext.OptionalImpl

	initialized bool
	mu          sync.RWMutex
	em          ext.ManagerInterface
	conf        *config.Config
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
func New() ext.Interface {
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
		return fmt.Errorf("sample plugin already initialized")
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
	// Subscribe to relevant events
	p.subscribeEvents(p.em)
	return nil
}

// RegisterRoutes registers routes for the plugin
func (p *Plugin) RegisterRoutes(r *gin.RouterGroup) {
	// Sample endpoints
	samples := r.Group("/samples")
	{
		samples.GET("", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Sample endpoint",
			})
		})
	}
}

// GetHandlers returns the handlers for the plugin
func (p *Plugin) GetHandlers() ext.Handler {
	return nil
}

// GetServices returns the services for the plugin
func (p *Plugin) GetServices() ext.Service {
	return nil
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

func init() {
	extp.RegisterPlugin(&Plugin{}, ext.Metadata{
		Name:         name + "-development",
		Version:      version,
		Dependencies: dependencies,
		Description:  desc,
		Type:         typeStr,
		Group:        group,
	})
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
