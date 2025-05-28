package initialize

import (
	"fmt"
	initConfig "ncobase/initialize/config"
	"ncobase/initialize/handler"
	"ncobase/initialize/service"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	extp "github.com/ncobase/ncore/extension/plugin"
	ext "github.com/ncobase/ncore/extension/types"
)

var (
	name         = "initialize"
	desc         = "Initialize plugin"
	version      = "1.0.0"
	dependencies []string
	typeStr      = "plugin"
	group        = "sys"
)

// Plugin represents the initialize plugin.
type Plugin struct {
	ext.OptionalImpl

	initialized bool
	mu          sync.RWMutex
	em          ext.ManagerInterface
	cleanup     func(name ...string)

	c *initConfig.Config
	s *service.Service
	h *handler.Handler
}

// init registers the plugin
func init() {
	extp.RegisterPlugin(New(), ext.Metadata{
		Name:         name,
		Version:      version,
		Dependencies: dependencies,
		Description:  desc,
		Type:         typeStr,
		Group:        group,
	})
}

// New returns a new instance of the plugin
func New() *Plugin {
	return &Plugin{}
}

// Init initializes the initialize plugin with the given config object
func (p *Plugin) Init(conf *config.Config, em ext.ManagerInterface) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.initialized {
		return fmt.Errorf("initialize plugin already initialized")
	}

	p.c = initConfig.GetDefaultConfig()
	if conf.Viper != nil {
		p.c = initConfig.GetConfigFromFile(p.c, conf.Viper)
	}

	p.em = em
	p.initialized = true

	return nil
}

// PostInit performs any necessary setup after initialization
func (p *Plugin) PostInit() error {
	sys, err := p.getSystemService()
	if err != nil {
		return err
	}
	as, err := p.getAuthService()
	if err != nil {
		return err
	}
	us, err := p.getUserService()
	if err != nil {
		return err
	}
	ts, err := p.getTenantService()
	if err != nil {
		return err
	}
	ss, err := p.getSpaceService()
	if err != nil {
		return err
	}
	acs, err := p.getAccessService()
	if err != nil {
		return err
	}

	// Create service
	p.s = service.New(p.em)
	// Set dependencies
	p.s.SetDependencies(p.c, sys, as, us, ts, ss, acs)
	// Create handler
	p.h = handler.New(p.s)

	return nil
}

// Name returns the name of the plugin
func (p *Plugin) Name() string {
	return name
}

// RegisterRoutes registers routes for the plugin
func (p *Plugin) RegisterRoutes(r *gin.RouterGroup) {
	// Initialization-related endpoints
	initGroup := r.Group("/" + p.Group() + "/initialize")
	{
		// Initialize all
		initGroup.POST("", p.h.Initialize.Execute)
		// Initialize organizations only
		initGroup.POST("/organizations", p.h.Initialize.InitializeOrganizations)
		// Initialize users only
		initGroup.POST("/users", p.h.Initialize.InitializeUsers)
		// Check status
		initGroup.GET("/status", p.h.Initialize.GetStatus)
	}
}

// GetHandlers returns the handlers for the plugin
func (p *Plugin) GetHandlers() ext.Handler {
	return p.h
}

// GetServices returns the services for the plugin
func (p *Plugin) GetServices() ext.Service {
	return p.s
}

// Cleanup cleans up the plugin
func (p *Plugin) Cleanup() error {
	if p.cleanup != nil {
		p.cleanup(p.Name())
	}
	return nil
}

// GetMetadata returns the metadata of the plugin
func (p *Plugin) GetMetadata() ext.Metadata {
	return ext.Metadata{
		Name:         p.Name(),
		Version:      p.Version(),
		Dependencies: p.Dependencies(),
		Description:  p.Description(),
		Type:         p.Type(),
		Group:        p.Group(),
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

// GetAllDependencies returns all dependencies of the plugin
func (p *Plugin) GetAllDependencies() []ext.DependencyEntry {
	return []ext.DependencyEntry{
		{Name: "system", Type: ext.WeakDependency},
		{Name: "auth", Type: ext.WeakDependency},
		{Name: "user", Type: ext.WeakDependency},
		{Name: "tenant", Type: ext.WeakDependency},
		{Name: "space", Type: ext.WeakDependency},
		{Name: "access", Type: ext.WeakDependency},
	}
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

// GetServiceInfo returns service registration info if NeedServiceDiscovery returns true
func (p *Plugin) GetServiceInfo() *ext.ServiceInfo {
	if !p.NeedServiceDiscovery() {
		return nil
	}

	metadata := p.GetMetadata()
	tags := append([]string{}, metadata.Group, metadata.Type)

	meta := make(map[string]string)
	meta["name"] = metadata.Name
	meta["version"] = metadata.Version
	meta["group"] = metadata.Group
	meta["type"] = metadata.Type
	meta["description"] = metadata.Description

	return &ext.ServiceInfo{
		Tags: tags,
		Meta: meta,
	}
}
