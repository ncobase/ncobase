package content

import (
	"fmt"
	"ncobase/common/config"
	"ncobase/feature"
	"ncobase/feature/content/data"
	"ncobase/feature/content/handler"
	"ncobase/feature/content/middleware"
	"ncobase/feature/content/service"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	name         = "content"
	desc         = "Content management plugin"
	version      = "1.0.0"
	dependencies []string
)

// Plugin represents the content plugin
type Plugin struct {
	initialized bool
	mu          sync.RWMutex
	fm          *feature.Manager
	s           *service.Service
	h           *handler.Handler
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
		return fmt.Errorf("content plugin already initialized")
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
	p.h = handler.New(p.s)
	// Subscribe to relevant events
	p.subscribeEvents(p.fm)
	return nil
}

// HasRoutes returns true if the plugin has routes, false otherwise
func (p *Plugin) HasRoutes() bool {
	return true
}

// RegisterRoutes registers routes for the plugin
func (p *Plugin) RegisterRoutes(e *gin.Engine) {
	// API v1 endpoints
	v1 := e.Group("/v1")
	// Taxonomy endpoints
	taxonomies := v1.Group("/taxonomies", middleware.AuthenticatedUser)
	{
		taxonomies.GET("", p.h.Taxonomy.List)
		taxonomies.POST("", p.h.Taxonomy.Create)
		taxonomies.GET("/:slug", p.h.Taxonomy.Get)
		taxonomies.PUT("/:slug", p.h.Taxonomy.Update)
		taxonomies.DELETE("/:slug", p.h.Taxonomy.Delete)
	}
	// Topic endpoints
	topics := v1.Group("/topics", middleware.AuthenticatedUser)
	{
		topics.GET("", p.h.Topic.List)
		topics.POST("", p.h.Topic.Create)
		topics.GET("/:slug", p.h.Topic.Get)
		topics.PUT("/:slug", p.h.Topic.Update)
		topics.DELETE("/:slug", p.h.Topic.Delete)
	}
}

// GetHandlers returns the handlers for the plugin
func (p *Plugin) GetHandlers() map[string]feature.Handler {
	return map[string]feature.Handler{
		"taxonomy": p.h.Taxonomy,
		"topic":    p.h.Topic,
	}
}

// GetServices returns the services for the plugin
func (p *Plugin) GetServices() map[string]feature.Service {
	return map[string]feature.Service{
		"taxonomy": p.s.Taxonomy,
		"topic":    p.s.Topic,
	}
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
	return []string{}
}

// RegisterEvents registers events for the plugin
func (p *Plugin) subscribeEvents(fm *feature.Manager) {
	// Implement any event subscriptions here
}

func init() {
	feature.RegisterPlugin(&Plugin{}, feature.Metadata{
		Name:         name + "-development",
		Version:      version,
		Dependencies: []string{},
		Description:  desc,
	})
}
