package sample

import (
	"fmt"
	"ncobase/common/config"
	"ncobase/common/feature"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	name         = "sample"
	desc         = "Sample plugin"
	version      = "1.0.0"
	dependencies []string
)

// Plugin represents the sample plugin
type Plugin struct {
	initialized bool
	mu          sync.RWMutex
	fm          *feature.Manager
	conf        *config.Config
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
		return fmt.Errorf("sample plugin already initialized")
	}

	p.fm = fm
	p.conf = conf
	p.initialized = true

	return nil
}

// PostInit performs any necessary setup after initialization
func (p *Plugin) PostInit() error {
	// Subscribe to relevant events
	p.subscribeEvents(p.fm)
	return nil
}

// RegisterRoutes registers routes for the plugin
func (p *Plugin) RegisterRoutes(e *gin.Engine) {
	// API v1 endpoints
	v1 := e.Group("/v1")
	// Sample endpoints
	samples := v1.Group("/samples")
	{
		samples.GET("", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Sample endpoint",
			})
		})
	}
}

// GetHandlers returns the handlers for the plugin
func (p *Plugin) GetHandlers() feature.Handler {
	return nil
}

// GetServices returns the services for the plugin
func (p *Plugin) GetServices() feature.Service {
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
func (p *Plugin) subscribeEvents(_ *feature.Manager) {
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
