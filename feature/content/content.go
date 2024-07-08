package content

import (
	"ncobase/common/config"
	"ncobase/feature"
	"ncobase/feature/content/data"
	"ncobase/feature/content/handler"
	"ncobase/feature/content/service"
	"ncobase/middleware"

	"github.com/gin-gonic/gin"
)

const (
	name    = "content"
	desc    = "Content management plugin"
	version = "1.0.0"
)

// Plugin represents the content plugin
type Plugin struct {
	s       *service.Service
	h       *handler.Handler
	d       *data.Data
	cleanup func()
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
func (p *Plugin) Init(conf *config.Config) (err error) {
	p.d, p.cleanup, err = data.New(conf.Data)
	if err != nil {
		return err
	}
	svc := service.New(p.d)
	p.s = svc
	p.h = handler.New(svc)
	return nil
}

// PostInit performs any necessary setup after initialization
func (p *Plugin) PostInit() error {
	// Implement any post-initialization logic here
	return nil
}

// RegisterRoutes registers routes for the plugin
func (p *Plugin) RegisterRoutes(e *gin.Engine) {
	// Taxonomy endpoints
	taxonomies := e.Group("/taxonomies", middleware.Authenticated)
	{
		taxonomies.GET("", p.h.Taxonomy.List)
		taxonomies.POST("", p.h.Taxonomy.Create)
		taxonomies.GET("/:slug", p.h.Taxonomy.Get)
		taxonomies.PUT("/:slug", p.h.Taxonomy.Update)
		taxonomies.DELETE("/:slug", p.h.Taxonomy.Delete)
	}

	// Topic endpoints
	topics := e.Group("/topics", middleware.Authenticated)
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
		"content": p.h,
	}
}

// GetServices returns the services for the plugin
func (p *Plugin) GetServices() map[string]feature.Service {
	return map[string]feature.Service{
		"content": p.s,
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
		p.cleanup()
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
	// Implement your own logic to check the plugin status
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

func init() {
	feature.RegisterPlugin(&Plugin{}, feature.Metadata{
		Name:         name + "-development",
		Version:      version,
		Dependencies: []string{},
		Description:  desc,
	})
}
