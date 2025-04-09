package workflow

import (
	"fmt"
	"ncobase/cmd/ncobase/middleware"
	"ncobase/core/workflow/data"
	we "ncobase/core/workflow/engine/core"
	"ncobase/core/workflow/handler"
	"ncobase/core/workflow/service"
	nec "ncore/ext/core"
	"ncore/pkg/config"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	name         = "workflow"
	desc         = "Workflow module"
	version      = "1.0.0"
	dependencies []string
	typeStr      = "module"
	group        = "flow"
)

// Module represents the workflow module.
type Module struct {
	nec.OptionalImpl
	initialized bool
	mu          sync.RWMutex
	em          nec.ManagerInterface
	conf        *config.Config
	we          *we.Engine
	h           *handler.Handler
	s           *service.Service
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

// New creates a new instance of the workflow module.
func New() nec.Interface {
	return &Module{}
}

// Init initializes the workflow module with the given config object
func (m *Module) Init(conf *config.Config, em nec.ManagerInterface) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("workflow module already initialized")
	}

	m.d, m.cleanup, err = data.New(conf.Data)
	if err != nil {
		return err
	}

	// service discovery
	if conf.Consul == nil {
		m.discovery.address = conf.Consul.Address
		m.discovery.tags = conf.Consul.Discovery.DefaultTags
		m.discovery.meta = conf.Consul.Discovery.DefaultMeta
	}

	m.em = em
	m.conf = conf
	m.initialized = true

	// Initialize workflow engine
	// TODO: Global config support
	m.we, err = we.NewEngine(nil, m.s, m.em)
	if err != nil {
		return fmt.Errorf("failed to create workflow engine: %w", err)
	}

	return nil
}

// PostInit performs any necessary setup after initialization
func (m *Module) PostInit() error {
	m.s = service.New(m.d, m.em)
	m.h = handler.New(m.s)

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
	m.h.RegisterRoutes(r)
}

// GetHandlers returns the handlers for the module
func (m *Module) GetHandlers() nec.Handler {
	return m.h
}

// GetServices returns the services for the module
func (m *Module) GetServices() nec.Service {
	return m.s
}

// Cleanup cleans up the module
func (m *Module) Cleanup() error {
	if m.cleanup != nil {
		m.cleanup(m.Name())
	}
	return nil
}

// GetMetadata returns the metadata of the module
func (m *Module) GetMetadata() nec.Metadata {
	return nec.Metadata{
		Name:         m.Name(),
		Version:      m.Version(),
		Dependencies: m.Dependencies(),
		Description:  m.Description(),
		Type:         m.Type(),
		Group:        m.Group(),
	}
}

// Version returns the version of the module
func (m *Module) Version() string {
	return version
}

// Dependencies returns the dependencies of the module
func (m *Module) Dependencies() []string {
	return dependencies
}

// Description returns the description of the module
func (m *Module) Description() string {
	return desc
}

// Type returns the type of the module
func (m *Module) Type() string {
	return typeStr
}

// Group returns the domain group of the module belongs
func (m *Module) Group() string {
	return group
}

// GetServiceInfo returns service registration info if NeedServiceDiscovery returns true
func (m *Module) GetServiceInfo() *nec.ServiceInfo {
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

	return &nec.ServiceInfo{
		Address: m.discovery.address,
		Tags:    tags,
		Meta:    meta,
	}
}
