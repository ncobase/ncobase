//go:build !plugin

package cmd

import (
	"context"
	"ncobase/common/config"
	"ncobase/common/log"
	"ncobase/internal/server/middleware"
	"ncobase/plugin"
	"ncobase/plugin/content/data"
	"ncobase/plugin/content/handler"
	"ncobase/plugin/content/service"

	"github.com/gin-gonic/gin"
)

type Plugin struct {
	s       *service.Service
	h       *handler.Handler
	d       *data.Data
	cleanup func()
}

func (p *Plugin) Name() string {
	return "content"
}

func (p *Plugin) Init(conf *config.Config) (err error) {
	p.d, p.cleanup, err = data.New(&conf.Data)
	if err != nil {
		return err
	}
	svc := service.New(p.d)
	p.s = svc
	p.h = handler.New(svc)
	return nil
}

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

func (p *Plugin) Cleanup() error {
	if p.cleanup != nil {
		p.cleanup()
	}
	return nil
}

// PluginInstance is the exported symbol that will be looked up by the plugin loader
var PluginInstance Plugin

func init() {
	plugin.RegisterPlugin(&PluginInstance)
	log.Infof(context.Background(), "%s plugin initialized", PluginInstance.Name())
}
