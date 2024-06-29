//go:build !plugin

package cmd

import (
	"context"
	"ncobase/common/config"
	"ncobase/common/log"
	"ncobase/internal/server/middleware"
	"ncobase/plugin"
	"ncobase/plugin/asset/data"
	"ncobase/plugin/asset/handler"
	"ncobase/plugin/asset/service"

	"github.com/gin-gonic/gin"
)

type Plugin struct {
	s       *service.Service
	h       *handler.Handler
	d       *data.Data
	cleanup func()
}

func (p *Plugin) Name() string {
	return "asset"
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

	// Asset endpoints
	assets := e.Group("/assets", middleware.Authenticated)
	{
		assets.GET("", p.h.Asset.ListAssetsHandler)
		assets.POST("", p.h.Asset.CreateAssetsHandler)
		assets.GET("/:slug", p.h.Asset.GetAssetHandler)
		assets.PUT("/:slug", p.h.Asset.UpdateAssetHandler)
		assets.DELETE("/:slug", p.h.Asset.DeleteAssetHandler)
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
