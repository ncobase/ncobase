package init

import (
	"context"
	"ncobase/common/config"
	"ncobase/common/log"
	"ncobase/feature"
	"ncobase/feature/init/service"

	"github.com/gin-gonic/gin"
)

const (
	name    = "init"
	desc    = "init plugin"
	version = "1.0.0"
)

// Plugin represents the group plugin.
type Plugin struct {
	conf    *config.Config
	s       service.InitService
	cleanup func(name ...string)
}

// PreInit performs any necessary setup before initialization
func (m *Plugin) PreInit() error {
	return nil
}

// Init initializes the group plugin with the given config object
func (m *Plugin) Init(conf *config.Config, fm *feature.Manager) (err error) {
	m.s = service.NewInitService()
	return nil
}

// PostInit performs any necessary setup after initialization
func (m *Plugin) PostInit() error {
	err := m.s.InitData()
	if err != nil {
		log.Errorf(context.Background(), "initialization failed: %v", err)
		return err
	}
	return nil
}

// Name returns the name of the plugin
func (m *Plugin) Name() string {
	return name
}

// RegisterRoutes registers routes for the plugin
func (m *Plugin) RegisterRoutes(_ *gin.Engine) {}

// GetHandlers returns the handlers for the plugin
func (m *Plugin) GetHandlers() feature.Handler {
	return nil
}

// GetServices returns the services for the plugin
func (m *Plugin) GetServices() feature.Service {
	return nil
}

// PreCleanup performs any necessary cleanup before the main cleanup
func (m *Plugin) PreCleanup() error {
	return nil
}

// Cleanup cleans up the plugin
func (m *Plugin) Cleanup() error {
	if m.cleanup != nil {
		m.cleanup(m.Name())
	}
	return nil
}

// Status returns the status of the plugin
func (m *Plugin) Status() string {
	return "active"
}

// GetMetadata returns the metadata of the plugin
func (m *Plugin) GetMetadata() feature.Metadata {
	return feature.Metadata{
		Name:         m.Name(),
		Version:      m.Version(),
		Dependencies: m.Dependencies(),
		Description:  desc,
	}
}

// Version returns the version of the plugin
func (m *Plugin) Version() string {
	return version
}

// Dependencies returns the dependencies of the plugin
func (m *Plugin) Dependencies() []string {
	return dependencies
}
