package templates

import "fmt"

func PluginTemplate(name string) string {
	return fmt.Sprintf(`package %s

import (
	"fmt"
	"ncobase/common/config"
	"ncobase/common/extension"
	"ncobase/plugin/%s/data"
	"ncobase/plugin/%s/handler"
	"ncobase/plugin/%s/service"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	name         = "%s"
	desc         = "%s plugin"
	version      = "1.0.0"
	dependencies []string
	typeStr      = "plugin"
	group        = "plug"
	enabledDiscovery = false
)

// Plugin represents the %s plugin.
type Plugin struct {
	initialized bool
	mu          sync.RWMutex
	em          *extension.Manager
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

// New creates a new instance of the %s plugin.
func New() extension.Interface {
	return &Plugin{}
}

// PreInit performs any necessary setup before initialization
func (p *Plugin) PreInit() error {
	// Implement any pre-initialization logic here
	return nil
}

// Init initializes the %s plugin with the given config object
func (p *Plugin) Init(conf *config.Config, em *extension.Manager) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.initialized {
		return fmt.Errorf("%s plugin already initialized")
	}

	p.d, p.cleanup, err = data.New(conf.Data)
	if err != nil {
		return err
	}

	// service discovery
	if conf.Consul == nil {
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
	p.s = service.New(p.conf, p.d)
	p.h = handler.New(p.s)

	return nil
}

// Name returns the name of the plugin
func (p *Plugin) Name() string {
	return name
}

// RegisterRoutes registers routes for the plugin
func (p *Plugin) RegisterRoutes(r *gin.RouterGroup) {
	// Implement your route registration logic here
}

// GetHandlers returns the handlers for the plugin
func (p *Plugin) GetHandlers() extension.Handler {
	return p.h
}

// GetServices returns the services for the plugin
func (p *Plugin) GetServices() extension.Service {
	return p.s
}

// PreCleanup performs any necessary cleanup before the main cleanup
func (p *Plugin) PreCleanup() error {
	// Implement any pre-cleanup logic here
	return nil
}

// Cleanup cleans up the plugin
func (p *Plugin) Cleanup() error {
	if p.cleanup != nil {
		p.cleanup(p.Name())
	}
	return nil
}

// Status returns the status of the plugin
func (p *Plugin) Status() string {
	return "active"
}

// GetMetadata returns the metadata of the plugin
func (p *Plugin) GetMetadata() extension.Metadata {
	return extension.Metadata{
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

// NeedServiceDiscovery returns if the module needs to be registered as a service
func (p *Plugin) NeedServiceDiscovery() bool {
	return enabledDiscovery
}

// GetServiceInfo returns service registration info if NeedServiceDiscovery returns true
func (p *Plugin) GetServiceInfo() *extension.ServiceInfo {
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

	return &extension.ServiceInfo{
		Address: p.discovery.address,
		Tags:    tags,
		Meta:    meta,
	}
}
`, name, name, name, name, name, name, name, name, name, name)
}
