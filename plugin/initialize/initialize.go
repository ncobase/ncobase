package initialize

import (
	"fmt"
	initConfig "ncobase/initialize/config"
	"ncobase/initialize/handler"
	"ncobase/initialize/service"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	exr "github.com/ncobase/ncore/extension/registry"
	ext "github.com/ncobase/ncore/extension/types"
)

var (
	name         = "initialize"
	desc         = "Initialize plugin"
	version      = "1.0.0"
	dependencies = []string{"system", "auth", "user", "tenant", "space", "access"}
	typeStr      = "plugin"
	group        = "sys"
)

// Plugin represents the initialize module.
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

// init registers the module
func init() {
	exr.RegisterToGroupWithWeakDeps(New(), group, []string{})
}

// New creates a new instance of the initialize module.
func New() ext.Interface {
	return &Plugin{}
}

// Init initializes the initialize module with the given config object
func (p *Plugin) Init(conf *config.Config, em ext.ManagerInterface) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.initialized {
		return fmt.Errorf("initialize module already initialized")
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

// Name returns the name of the module
func (p *Plugin) Name() string {
	return name
}

// RegisterRoutes registers routes for the module
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

// GetHandlers returns the handlers for the module
func (p *Plugin) GetHandlers() ext.Handler {
	return p.h
}

// GetServices returns the services for the module
func (p *Plugin) GetServices() ext.Service {
	return p.s
}

// Cleanup cleans up the module
func (p *Plugin) Cleanup() error {
	if p.cleanup != nil {
		p.cleanup(p.Name())
	}
	return nil
}

// GetMetadata returns the metadata of the module
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

// Version returns the version of the module
func (p *Plugin) Version() string {
	return version
}

// Dependencies returns the dependencies of the module
func (p *Plugin) Dependencies() []string {
	return dependencies
}

// Description returns the description of the module
func (p *Plugin) Description() string {
	return desc
}

// Type returns the type of the module
func (p *Plugin) Type() string {
	return typeStr
}

// Group returns the domain group of the module belongs
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
