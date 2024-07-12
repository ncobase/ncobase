package relationship

import (
	"fmt"
	"ncobase/common/config"
	"ncobase/feature"
	"ncobase/feature/relationship/data"
	"ncobase/feature/relationship/service"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	name         = "relationship"
	desc         = "Relationship plugin"
	version      = "1.0.0"
	dependencies []string
)

// Plugin represents the relationship plugin
type Plugin struct {
	initialized bool
	mu          sync.RWMutex
	fm          *feature.Manager
	s           *service.Service
	d           *data.Data
	cleanup     func(name ...string)
}

// Name returns the name of the plugin
func (p *Plugin) Name() string {
	return name
}

// PreInit performs any necessary setup before initialization
func (p *Plugin) PreInit() error {
	// Implement any pre-initialization logic here
	return nil
}

// Init initializes the plugin
func (p *Plugin) Init(conf *config.Config, fm *feature.Manager) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.initialized {
		return fmt.Errorf("relationship plugin already initialized")
	}

	p.d, p.cleanup, err = data.New(conf.Data)
	if err != nil {
		return err
	}

	p.fm = fm
	p.initialized = true

	return nil
}

// PostInit performs any necessary setup after initialization
func (p *Plugin) PostInit() error {
	p.s = service.New(p.d)
	// Subscribe to relevant events
	p.subscribeEvents(p.fm)
	return nil
}

// RegisterRoutes registers routes for the plugin
func (p *Plugin) RegisterRoutes(e *gin.Engine) {}

// GetHandlers returns the handlers for the plugin
func (p *Plugin) GetHandlers() map[string]feature.Handler {
	return nil
}

// GetServices returns the services for the plugin
func (p *Plugin) GetServices() map[string]feature.Service {
	return nil
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

// GetMetadata returns the metadata of the plugin
func (p *Plugin) GetMetadata() feature.Metadata {
	return feature.Metadata{
		Name:         p.Name(),
		Version:      p.Version(),
		Dependencies: p.Dependencies(),
		Description:  desc,
	}
}

// Status returns the status of the plugin
func (p *Plugin) Status() string {
	// Implement logic to check the plugin status
	return "active"
}

// Version returns the version of the plugin
func (p *Plugin) Version() string {
	return version
}

// Dependencies returns the dependencies of the plugin
func (p *Plugin) Dependencies() []string {
	return dependencies
}

// SubscribeEvents subscribes to relevant events
func (p *Plugin) subscribeEvents(fm *feature.Manager) {
	// Implement any event subscriptions here
}

func init() {
	feature.RegisterPlugin(&Plugin{}, feature.Metadata{
		Name:         name + "-development",
		Version:      version,
		Dependencies: dependencies,
		Description:  desc,
	})
}
