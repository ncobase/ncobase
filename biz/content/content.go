package content

import (
	"context"
	"fmt"
	"ncobase/biz/content/data"
	"ncobase/biz/content/handler"
	"ncobase/biz/content/service"
	"ncobase/internal/middleware"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	exr "github.com/ncobase/ncore/extension/registry"
	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"go.opentelemetry.io/otel/trace"
)

var (
	name         = "content"
	desc         = "Content module"
	version      = "1.0.0"
	dependencies []string
	typeStr      = "module"
	group        = "cms"
)

// Module represents the content module
type Module struct {
	ext.OptionalImpl

	initialized bool
	mu          sync.RWMutex
	em          ext.ManagerInterface
	conf        *config.Config
	h           *handler.Handler
	s           *service.Service
	d           *data.Data
	tracer      trace.Tracer
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

// Init initializes the module
func (m *Module) Init(conf *config.Config, em ext.ManagerInterface) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("content module already initialized")
	}

	m.d, m.cleanup, err = data.New(conf.Data, conf.Environment)
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
	m.s = service.New(m.em, m.d)
	m.h = handler.New(m.s)

	m.subscribeEvents(m.em)
	// Subscribe to extension events for dependency refresh
	m.em.SubscribeEvent("exts.resource.ready", func(data any) {
		m.s.RefreshDependencies()
	})

	m.em.SubscribeEvent("exts.all.registered", func(data any) {
		m.s.RefreshDependencies()
	})

	return nil
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	// Belong domain group
	r = r.Group("/"+m.Group(), middleware.AuthenticatedUser)

	// Taxonomy endpoints
	taxonomies := r.Group("/taxonomies")
	{
		taxonomies.GET("", m.h.Taxonomy.List)
		taxonomies.POST("", m.h.Taxonomy.Create)
		taxonomies.GET("/:slug", m.h.Taxonomy.Get)
		taxonomies.PUT("/:slug", m.h.Taxonomy.Update)
		taxonomies.DELETE("/:slug", m.h.Taxonomy.Delete)
	}

	// Topic endpoints
	topics := r.Group("/topics")
	{
		topics.GET("", m.h.Topic.List)
		topics.POST("", m.h.Topic.Create)
		topics.GET("/:slug", m.h.Topic.Get)
		topics.PUT("/:slug", m.h.Topic.Update)
		topics.DELETE("/:slug", m.h.Topic.Delete)
	}

	// Channel endpoints
	channels := r.Group("/channels")
	{
		channels.GET("", m.h.Channel.List)
		channels.POST("", m.h.Channel.Create)
		channels.GET("/:slug", m.h.Channel.Get)
		channels.PUT("/:slug", m.h.Channel.Update)
		channels.DELETE("/:slug", m.h.Channel.Delete)
	}

	// Distribution endpoints
	distributions := r.Group("/distributions")
	{
		distributions.GET("", m.h.Distribution.List)
		distributions.POST("", m.h.Distribution.Create)
		distributions.GET("/:id", m.h.Distribution.Get)
		distributions.PUT("/:id", m.h.Distribution.Update)
		distributions.DELETE("/:id", m.h.Distribution.Delete)
		distributions.POST("/:id/publish", m.h.Distribution.Publish)
		distributions.POST("/:id/cancel", m.h.Distribution.Cancel)
	}

	// Media endpoints
	media := r.Group("/media")
	{
		media.GET("", m.h.Media.List)
		media.POST("", m.h.Media.Create)
		media.GET("/:id", m.h.Media.Get)
		media.PUT("/:id", m.h.Media.Update)
		media.DELETE("/:id", m.h.Media.Delete)
	}

	// Topic Media endpoints
	topicMedia := r.Group("/topic-media")
	{
		topicMedia.GET("", m.h.TopicMedia.List)
		topicMedia.POST("", m.h.TopicMedia.Create)
		topicMedia.GET("/:id", m.h.TopicMedia.Get)
		topicMedia.PUT("/:id", m.h.TopicMedia.Update)
		topicMedia.DELETE("/:id", m.h.TopicMedia.Delete)
		topicMedia.GET("/by-topic-and-media", m.h.TopicMedia.GetByTopicAndMedia)
		topicMedia.GET("/by-topic/:topicId", m.h.TopicMedia.ListByTopic)
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

// Cleanup cleans up the module
func (m *Module) Cleanup() error {
	if m.cleanup != nil {
		m.cleanup(m.Name())
	}
	return nil
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

// RegisterEvents registers events for the module
func (m *Module) subscribeEvents(_ ext.ManagerInterface) {
	// Implement any event subscriptions here
}

// init registers the module
func init() {
	exr.RegisterToGroupWithWeakDeps(New(), group, []string{})
}

// New creates a new instance of the auth module.
func New() ext.Interface {
	return &Module{}
}

// GetServiceInfo returns service registration info if NeedServiceDiscovery returns true
func (m *Module) GetServiceInfo() *ext.ServiceInfo {
	if !m.NeedServiceDiscovery() {
		return nil
	}

	logger.Infof(context.Background(), "Getting service info for %s with address: %s",
		m.Name(), m.discovery.address)

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
