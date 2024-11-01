package templates

import "fmt"

func PluginTemplate(name string) string {
	return fmt.Sprintf(`package %s

import (
	"fmt"
	"ncobase/common/config"
	"ncobase/common/feature"
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
)

// Plugin represents the %s plugin.
type Plugin struct {
	initialized bool
	mu          sync.RWMutex
	fm          *feature.Manager
	conf        *config.Config
	h           *handler.Handler
	s           *service.Service
	d           *data.Data
	cleanup     func(name ...string)
}

// New creates a new instance of the %s plugin.
func New() feature.Interface {
	return &Plugin{}
}

// PreInit performs any necessary setup before initialization
func (p *Plugin) PreInit() error {
	// Implement any pre-initialization logic here
	return nil
}

// Init initializes the %s plugin with the given config object
func (p *Plugin) Init(conf *config.Config, fm *feature.Manager) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.initialized {
		return fmt.Errorf("%s plugin already initialized")
	}

	p.d, p.cleanup, err = data.New(conf.Data)
	if err != nil {
		return err
	}

	p.fm = fm
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
func (p *Plugin) GetHandlers() feature.Handler {
	return p.h
}

// GetServices returns the services for the plugin
func (p *Plugin) GetServices() feature.Service {
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
func (p *Plugin) GetMetadata() feature.Metadata {
	return feature.Metadata{
		Name:         p.Name(),
		Version:      p.Version(),
		Dependencies: p.Dependencies(),
		Description:  desc,
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
`, name, name, name, name, name, name, name, name, name, name)
}
