package asset

import (
	"ncobase/common/config"
	"ncobase/middleware"
	"ncobase/plugin"
	"ncobase/plugin/asset/data"
	"ncobase/plugin/asset/handler"
	"ncobase/plugin/asset/service"

	"github.com/gin-gonic/gin"
)

// Plugin represents the asset plugin
type Plugin struct {
	s       *service.Service
	h       *handler.Handler
	d       *data.Data
	cleanup func()
}

// Name returns the name of the plugin
func (p *Plugin) Name() string {
	return "asset"
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

// GetHandlers returns the handlers for the plugin
func (p *Plugin) GetHandlers() map[string]plugin.Handler {
	// Implement your own logic to get the handlers
	return map[string]plugin.Handler{}
}

// GetServices returns the services for the plugin
func (p *Plugin) GetServices() map[string]plugin.Service {
	// Implement your own logic to get the services
	return map[string]plugin.Service{}
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
		Name:         "asset",
		Version:      "1.0.0",
		Dependencies: []string{},
		Description:  "Asset management plugin",
	}
}

// Status returns the status of the plugin
func (p *Plugin) Status() string {
	// Implement your own logic to check the plugin status
	return "active"
}

func init() {
	metadata := plugin.Metadata{
		Name:         "asset-development",
		Version:      "1.0.0",
		Dependencies: []string{},
		Description:  "Asset management plugin",
	}
	plugin.RegisterPlugin(&Plugin{}, metadata)
}
