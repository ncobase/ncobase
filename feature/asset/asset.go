package asset

import (
	"ncobase/common/config"
	"ncobase/feature"
	"ncobase/feature/asset/data"
	"ncobase/feature/asset/handler"
	"ncobase/feature/asset/service"
	"ncobase/middleware"

	"github.com/gin-gonic/gin"
)

const (
	name    = "asset"
	desc    = "Asset management plugin"
	version = "1.0.0"
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
func (p *Plugin) GetHandlers() map[string]feature.Handler {
	return map[string]feature.Handler{
		"asset": p.h,
	}
}

// GetServices returns the services for the plugin
func (p *Plugin) GetServices() map[string]feature.Service {
	return map[string]feature.Service{
		"asset": p.s,
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
