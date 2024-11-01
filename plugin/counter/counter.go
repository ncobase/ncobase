package counter

import (
	"fmt"
	"ncobase/cmd/ncobase/middleware"
	"ncobase/common/config"
	"ncobase/common/feature"
	"ncobase/plugin/counter/data"
	"ncobase/plugin/counter/data/repository"
	"ncobase/plugin/counter/handler"
	"ncobase/plugin/counter/service"
	"sync"

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
	initialized bool
	mu          sync.RWMutex
	fm          *feature.Manager
	conf        *config.Config
	d           *data.Data
	r           *repository.Repository
	s           *service.Service
	h           *handler.Handler
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
		return fmt.Errorf("counter plugin already initialized")
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
	p.r = repository.New(p.d)
	p.s = service.New(p.d)
	p.h = handler.New(p.s)
	// Subscribe to relevant events
	p.subscribeEvents(p.fm)
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

// GetMetadata returns the metadata of the plugin
func (p *Plugin) GetMetadata() feature.Metadata {
	return feature.Metadata{
		Name:         p.Name(),
		Version:      p.Version(),
		Dependencies: p.Dependencies(),
		Description:  p.Description(),
		Type:         p.Type(),
		Group:        p.Group(),
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
func (p *Plugin) subscribeEvents(_ *feature.Manager) {
	// Implement any event subscriptions here
}

func init() {
	feature.RegisterPlugin(&Plugin{}, feature.Metadata{
		Name:         name,
		Version:      version,
		Dependencies: dependencies,
		Description:  desc,
		Type:         typeStr,
		Group:        group,
	})
}
