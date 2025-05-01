package system

import (
	"fmt"
	"ncobase/cmd/ncobase/middleware"
	"ncobase/core/system/data"
	"ncobase/core/system/handler"
	"ncobase/core/system/initialize"
	"ncobase/core/system/service"
	"sync"

	"github.com/ncobase/ncore/config"
	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/net/resp"

	"github.com/gin-gonic/gin"
)

var (
	name             = "system"
	desc             = "System module"
	version          = "1.0.0"
	dependencies     = []string{"access", "auth", "tenant", "user"}
	typeStr          = "module"
	group            = "sys"
	enabledDiscovery = false
)

// Module represents the system module.
type Module struct {
	initialized bool
	mu          sync.RWMutex
	em          ext.ManagerInterface
	conf        *config.Config
	h           *handler.Handler
	s           *service.Service
	i           *initialize.Service
	d           *data.Data
	cleanup     func(name ...string)

	discovery
}

// discovery represents the service discovery
type discovery struct {
	address string
	tags    []string
	meta    map[string]string
}

// New creates a new instance of the system module.
func New() ext.Interface {
	return &Module{}
}

// PreInit performs any necessary setup before initialization
func (m *Module) PreInit() error {
	// Implement any pre-initialization logic here
	return nil
}

// Init initializes the system module with the given config object
func (m *Module) Init(conf *config.Config, em ext.ManagerInterface) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("system module already initialized")
	}

	m.d, m.cleanup, err = data.New(conf.Data)
	if err != nil {
		return err
	}

	// service discovery
	if conf.Consul != nil {
		m.discovery.address = conf.Consul.Address
		m.discovery.tags = conf.Consul.Discovery.DefaultTags
		m.discovery.meta = conf.Consul.Discovery.DefaultMeta
	}

	m.em = em
	m.initialized = true
	m.conf = conf

	return nil
}

// PostInit performs any necessary setup after initialization
func (m *Module) PostInit() error {
	m.s = service.New(m.d, m.em)
	m.h = handler.New(m.s)
	// Subscribe to relevant events
	m.subscribeEvents(m.em)
	// get dependencies services
	as, err := m.getAuthService()
	if err != nil {
		return err
	}
	us, err := m.getUserService()
	if err != nil {
		return err
	}
	ts, err := m.getTenantService()
	if err != nil {
		return err
	}
	ss, err := m.getSpaceService()
	if err != nil {
		return err
	}
	acs, err := m.getAccessService()
	if err != nil {
		return err
	}
	// initialize
	m.i = initialize.New(as, us, ts, ss, acs)
	return nil
}

// Name returns the name of the module
func (m *Module) Name() string {
	return name
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	// Belong domain group
	r = r.Group("/"+m.Group(), middleware.AuthenticatedUser)
	// System initialization, roles, permissions, Casbin policies...
	r.POST("/initialize", func(c *gin.Context) {
		err := m.i.Execute()
		if err != nil {
			resp.Fail(c.Writer, resp.BadRequest(err.Error()))
			return
		}
		resp.Success(c.Writer)
	})
	// Menu endpoints
	menus := r.Group("/menus")
	{
		menus.GET("", m.h.Menu.List)
		menus.POST("", m.h.Menu.Create)
		menus.GET("/:slug", m.h.Menu.Get)
		menus.PUT("/:slug", m.h.Menu.Update)
		menus.DELETE("/:slug", m.h.Menu.Delete)
	}
	// Dictionary endpoints
	dictionaries := r.Group("/dictionaries")
	{
		dictionaries.GET("", m.h.Dictionary.List)
		dictionaries.POST("", m.h.Dictionary.Create)
		dictionaries.GET("/:slug", m.h.Dictionary.Get)
		dictionaries.PUT("/:slug", m.h.Dictionary.Update)
		dictionaries.DELETE("/:slug", m.h.Dictionary.Delete)
	}
	// Options endpoints
	options := r.Group("/options")
	{
		options.POST("/initialize", m.h.Options.Initialize)
		options.GET("", m.h.Options.List)
		options.POST("", m.h.Options.Create)
		options.GET("/:slug", m.h.Options.Get)
		options.PUT("/:slug", m.h.Options.Update)
		options.DELETE("/:slug", m.h.Options.Delete)
	}
}

// GetHandlers returns the handlers for the module
func (m *Module) GetHandlers() ext.Handler {
	return m.h
}

// GetServices returns the services for the module
func (m *Module) GetServices() ext.Service {
	return m.s
}

// PreCleanup performs any necessary cleanup before the main cleanup
func (m *Module) PreCleanup() error {
	// Implement any pre-cleanup logic here
	return nil
}

// Cleanup cleans up the module
func (m *Module) Cleanup() error {
	if m.cleanup != nil {
		m.cleanup(m.Name())
	}
	return nil
}

// Status returns the status of the module
func (m *Module) Status() string {
	return "active"
}

// GetMetadata returns the metadata of the module
func (m *Module) GetMetadata() ext.Metadata {
	return ext.Metadata{
		Name:         m.Name(),
		Version:      m.Version(),
		Dependencies: m.Dependencies(),
		Description:  m.Description(),
		Type:         m.Type(),
		Group:        m.Group(),
	}
}

// Dependencies returns the dependencies of the module
func (m *Module) Dependencies() []string {
	return dependencies
}

// Description returns the description of the module
func (m *Module) Description() string {
	return desc
}

// Version returns the version of the module
func (m *Module) Version() string {
	return version
}

// Type returns the type of the module
func (m *Module) Type() string {
	return typeStr
}

// Group returns the domain group of the module belongs
func (m *Module) Group() string {
	return group
}

// SubscribeEvents subscribes to relevant events
func (m *Module) subscribeEvents(_ ext.ManagerInterface) {
	// Implement any event subscriptions here
}

// NeedServiceDiscovery returns if the module needs to be registered as a service
func (m *Module) NeedServiceDiscovery() bool {
	return enabledDiscovery
}

// GetServiceInfo returns service registration info if NeedServiceDiscovery returns true
func (m *Module) GetServiceInfo() *ext.ServiceInfo {
	if !m.NeedServiceDiscovery() {
		return nil
	}

	metadata := m.GetMetadata()

	tags := append(m.discovery.tags, metadata.Group, metadata.Type)

	meta := make(map[string]string)
	for k, v := range m.discovery.meta {
		meta[k] = v
	}
	meta["name"] = metadata.Name
	meta["version"] = metadata.Version
	meta["group"] = metadata.Group
	meta["type"] = metadata.Type
	meta["description"] = metadata.Description

	return &ext.ServiceInfo{
		Address: m.discovery.address,
		Tags:    tags,
		Meta:    meta,
	}
}
