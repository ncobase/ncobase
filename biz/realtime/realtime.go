package realtime

import (
	"fmt"
	"ncobase/internal/middleware"
	"ncobase/realtime/data"
	"ncobase/realtime/handler"
	"ncobase/realtime/service"
	"sync"

	"github.com/ncobase/ncore/config"
	exr "github.com/ncobase/ncore/extension/registry"
	ext "github.com/ncobase/ncore/extension/types"

	"github.com/gin-gonic/gin"
)

var (
	name         = "realtime"
	desc         = "Real-time event and notification module"
	version      = "1.0.0"
	dependencies []string
	typeStr      = "module"
	group        = "rt"
)

// Module represents the realtime module
type Module struct {
	ext.OptionalImpl

	initialized bool
	mu          sync.RWMutex
	em          ext.ManagerInterface
	conf        *config.Config
	h           *handler.Handler
	s           *service.Service
	d           *data.Data
	cleanup     func(name ...string)

	discovery
}

type discovery struct {
	address string
	tags    []string
	meta    map[string]string
}

// init registers the module
func init() {
	exr.RegisterToGroupWithWeakDeps(New(), group, []string{})
}

// New creates a new instance of the realtime module
func New() ext.Interface {
	return &Module{}
}

// Init initializes the realtime module
func (m *Module) Init(conf *config.Config, em ext.ManagerInterface) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("realtime module already initialized")
	}

	m.d, m.cleanup, err = data.New(conf.Data)
	if err != nil {
		return err
	}

	// Configure service discovery
	if conf.Consul != nil {
		m.discovery.address = conf.Consul.Address
		m.discovery.tags = conf.Consul.Discovery.DefaultTags
		m.discovery.meta = conf.Consul.Discovery.DefaultMeta
	}

	m.em = em
	m.conf = conf
	m.initialized = true

	return nil
}

// PostInit performs setup after initialization
func (m *Module) PostInit() error {
	m.s = service.New(m.d, m.em)
	m.h = handler.New(m.s)

	return nil
}

// RegisterRoutes registers HTTP and WebSocket routes
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	// Group routes under /realtime
	rg := r.Group("/"+m.Group(), middleware.AuthenticatedUser)

	// WebSocket endpoint
	rg.GET("/ws", m.h.WebSocket.HandleConnection)

	// Notification endpoints
	notifications := rg.Group("/notifications")
	{
		notifications.GET("", m.h.Notification.List)
		notifications.POST("", m.h.Notification.Create)
		notifications.GET("/:id", m.h.Notification.Get)
		notifications.PUT("/:id", m.h.Notification.Update)
		notifications.DELETE("/:id", m.h.Notification.Delete)
		notifications.PUT("/:id/read", m.h.Notification.MarkAsRead)
		notifications.PUT("/:id/unread", m.h.Notification.MarkAsUnread)
		notifications.PUT("/read-all", m.h.Notification.MarkAllAsRead)
		notifications.PUT("/unread-all", m.h.Notification.MarkAllAsUnread)
	}

	// Channel endpoints
	channels := rg.Group("/channels")
	{
		channels.GET("", m.h.Channel.List)
		channels.POST("", m.h.Channel.Create)
		channels.GET("/:id", m.h.Channel.Get)
		channels.PUT("/:id", m.h.Channel.Update)
		channels.DELETE("/:id", m.h.Channel.Delete)
		channels.POST("/:id/subscribe", m.h.Channel.Subscribe)
		channels.POST("/:id/unsubscribe", m.h.Channel.Unsubscribe)
		channels.GET("/:id/subscribers", m.h.Channel.GetSubscribers)
		channels.GET("/:id/user", m.h.Channel.GetUserChannels)
	}

	// Event endpoints
	events := rg.Group("/events")
	{
		events.GET("", m.h.Event.List)
		events.GET("/:id", m.h.Event.Get)
		events.DELETE("/:id", m.h.Event.Delete)
		events.POST("/publish", m.h.Event.Publish)
		events.GET("/history", m.h.Event.GetHistory)
	}
}

// Name returns the name of the module
func (m *Module) Name() string {
	return name
}

// GetHandlers returns the handlers
func (m *Module) GetHandlers() ext.Handler {
	return m.h
}

// GetServices returns the services
func (m *Module) GetServices() ext.Service {
	return m.s
}

// Cleanup cleans up resources
func (m *Module) Cleanup() error {
	if m.cleanup != nil {
		m.cleanup(m.Name())
	}
	return nil
}

// GetMetadata returns module metadata
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

// Version returns the module version
func (m *Module) Version() string {
	return version
}

// Dependencies returns module dependencies
func (m *Module) Dependencies() []string {
	return dependencies
}

// Description returns the module description
func (m *Module) Description() string {
	return desc
}

// Type returns the module type
func (m *Module) Type() string {
	return typeStr
}

// Group returns the module group
func (m *Module) Group() string {
	return group
}

// GetServiceInfo returns service discovery info
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
