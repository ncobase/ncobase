package content

import (
	"ncobase/common/config"
	"ncobase/middleware"
	"ncobase/plugin"
	"ncobase/plugin/content/data"
	"ncobase/plugin/content/handler"
	"ncobase/plugin/content/service"

	"github.com/gin-gonic/gin"
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
	return "content"
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

// Cleanup cleans up the plugin
func (p *Plugin) Cleanup() error {
	if p.cleanup != nil {
		p.cleanup()
	}
	return nil
}

// GetMetadata returns the metadata of the plugin
func (p *Plugin) GetMetadata() plugin.Metadata {
	return plugin.Metadata{
		Name:         "content",
		Version:      "1.0.0",
		Dependencies: []string{},
		Description:  "Content management plugin",
	}
}

// Status returns the status of the plugin
func (p *Plugin) Status() string {
	// Implement your own logic to check the plugin status
	return "active"
}

// PluginInstance is the exported symbol that will be looked up by the plugin loader
var PluginInstance = &Plugin{}

func init() {
	metadata := plugin.Metadata{
		Name:         "content-development",
		Version:      "1.0.0",
		Dependencies: []string{},
		Description:  "Content management plugin",
	}
	plugin.RegisterPlugin(PluginInstance, metadata)
}
