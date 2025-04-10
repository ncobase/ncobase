package resource

import (
	"fmt"
	nec "github.com/ncobase/ncore/ext/core"
	"github.com/ncobase/ncore/pkg/config"
	"ncobase/cmd/ncobase/middleware"
	"ncobase/domain/resource/data"
	"ncobase/domain/resource/handler"
	"ncobase/domain/resource/service"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	name             = "resource"
	desc             = "Resource module"
	version          = "1.0.0"
	dependencies     []string
	typeStr          = "module"
	group            = "res"
	enabledDiscovery = false
)

// Module represents the resource module
type Module struct {
	initialized bool
	mu          sync.RWMutex
	em          nec.ManagerInterface
	s           *service.Service
	h           *handler.Handler
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

// Name returns the name of the module
func (m *Module) Name() string {
	return name
}

// PreInit performs any necessary setup before initialization
func (m *Module) PreInit() error {
	// Implement any pre-initialization logic here
	return nil
}

// Init initializes the module
func (m *Module) Init(conf *config.Config, em nec.ManagerInterface) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("resource module already initialized")
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
	m.initialized = true

	return nil
}

// PostInit performs any necessary setup after initialization
func (m *Module) PostInit() error {
	m.s = service.New(m.d)
	m.h = handler.New(m.s)
	// Subscribe to relevant events
	m.subscribeEvents(m.em)
	return nil
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	// Belong domain group
	r = r.Group("/"+m.Group(), middleware.AuthenticatedUser)
	// Attachment endpoints
	attachments := r.Group("/attachments")
	{
		attachments.GET("", m.h.Attachment.List)
		attachments.POST("", m.h.Attachment.Create)
		attachments.GET("/:slug", m.h.Attachment.Get)
		attachments.PUT("/:slug", m.h.Attachment.Update)
		attachments.DELETE("/:slug", m.h.Attachment.Delete)
	}
}

// GetHandlers returns the handlers for the module
func (m *Module) GetHandlers() nec.Handler {
	return m.h
}

// GetServices returns the services for the module
func (m *Module) GetServices() nec.Service {
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

// Status returns the status of the module
func (m *Module) Status() string {
	// Implement logic to check the module status
	return "active"
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

// SubscribeEvents subscribes to relevant events
func (m *Module) subscribeEvents(_ nec.ManagerInterface) {
	// Implement any event subscriptions here
}

// func init() {
// 	extension.RegisterModule(&Module{}, extension.Metadata{
// 		Name:         name + "-development",
// 		Version:      version,
// 		Dependencies: dependencies,
// 		Description:  m.Description(),
// 	})
// }

// New creates a new instance of the auth module.
func New() nec.Interface {
	return &Module{}
}

// NeedServiceDiscovery returns if the module needs to be registered as a service
func (m *Module) NeedServiceDiscovery() bool {
	return enabledDiscovery
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
