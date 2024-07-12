package init

import (
	"context"
	"fmt"
	"ncobase/common/config"
	"ncobase/common/log"
	"ncobase/feature"
	accessService "ncobase/feature/access/service"
	"ncobase/feature/init/service"
	tenantService "ncobase/feature/tenant/service"
	userService "ncobase/feature/user/service"

	"github.com/gin-gonic/gin"
)

var (
	name         = "init"
	desc         = "init plugin"
	version      = "1.0.0"
	dependencies []string
)

// Plugin represents the group plugin.
type Plugin struct {
	conf    *config.Config
	s       service.InitService
	fm      *feature.Manager
	cleanup func(name ...string)
}

// PreInit performs any necessary setup before initialization
func (p *Plugin) PreInit() error {
	return nil
}

// Init initializes the group plugin with the given config object
func (p *Plugin) Init(conf *config.Config, fm *feature.Manager) (err error) {
	p.s = service.NewInitService()
	p.fm = fm
	return nil
}

// PostInit performs any necessary setup after initialization
func (p *Plugin) PostInit() error {
	err := p.s.InitData()
	if err != nil {
		log.Errorf(context.Background(), "initialization failed: %v", err)
		return err
	}
	return nil
}

// Name returns the name of the plugin
func (p *Plugin) Name() string {
	return name
}

// RegisterRoutes registers routes for the plugin
func (p *Plugin) RegisterRoutes(_ *gin.Engine) {}

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
	return nil
}

// Cleanup cleans up the plugin
func (p *Plugin) Cleanup() error {
	if p.cleanup != nil {
		p.cleanup(p.Name())
	}
	return nil
}

// Status returns the status of the plugin
func (p *Plugin) Status() string {
	return "active"
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

// Version returns the version of the plugin
func (p *Plugin) Version() string {
	return version
}

// Dependencies returns the dependencies of the plugin
func (p *Plugin) Dependencies() []string {
	return dependencies
}

// GetUserService returns the user service
func (p *Plugin) getUserService() (*userService.Service, error) {
	f, err := p.fm.GetService("user")
	if err != nil {
		return nil, fmt.Errorf("failed to get user service: %v", err)
	}

	us, ok := f.(*userService.Service)
	if !ok {
		return nil, fmt.Errorf("user service does not implement UserServiceInterface")
	}

	return us, nil
}

// GetTenantService returns the tenant service
func (p *Plugin) getTenantService() (*tenantService.Service, error) {
	f, err := p.fm.GetService("tenant")
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant service: %v", err)
	}

	ts, ok := f.(*tenantService.Service)
	if !ok {
		return nil, fmt.Errorf("tenant service does not implement TenantServiceInterface")
	}

	return ts, nil
}

// GetAccessService returns the access service
func (p *Plugin) getAccessService() (*accessService.Service, error) {
	f, err := p.fm.GetService("access")
	if err != nil {
		return nil, fmt.Errorf("failed to get access service: %v", err)
	}
	as, ok := f.(*accessService.Service)
	if !ok {
		return nil, fmt.Errorf("access service does not implement")
	}
	return as, nil
}

func init() {
	feature.RegisterPlugin(&Plugin{}, feature.Metadata{
		Name:         name + "-development",
		Version:      version,
		Dependencies: dependencies,
		Description:  desc,
	})
}
