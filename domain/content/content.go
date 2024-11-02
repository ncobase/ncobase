package content

import (
	"fmt"
	"ncobase/cmd/ncobase/middleware"
	"ncobase/common/config"
	"ncobase/common/extension"
	"ncobase/common/observes"
	"ncobase/domain/content/data"
	"ncobase/domain/content/handler"
	"ncobase/domain/content/service"
	"sync"

	"github.com/gin-gonic/gin"
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
	initialized bool
	mu          sync.RWMutex
	em          *extension.Manager
	conf        *config.Config
	h           *handler.Handler
	s           *service.Service
	d           *data.Data
	tracer      trace.Tracer
	cleanup     func(name ...string)
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
func (m *Module) Init(conf *config.Config, em *extension.Manager) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("content module already initialized")
	}

	m.d, m.cleanup, err = data.New(conf.Data)
	if err != nil {
		return err
	}

	m.em = em
	m.initialized = true
	m.conf = conf

	return nil
}

// PostInit performs any necessary setup after initialization
func (m *Module) PostInit() error {
	serviceOpt := observes.TracingDecoratorOption{
		Layer:                    observes.LayerService,
		CreateSpanForEachMethod:  true,
		RecordMethodParams:       true,
		RecordMethodReturnValues: true,
	}
	m.s = observes.DecorateStruct(service.New(m.d), serviceOpt)

	handlerOpt := observes.TracingDecoratorOption{
		Layer:                    observes.LayerHandler,
		CreateSpanForEachMethod:  true,
		RecordMethodParams:       true,
		RecordMethodReturnValues: false,
	}
	m.h = observes.DecorateStruct(handler.New(m.s), handlerOpt)

	m.subscribeEvents(m.em)
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
}

// GetHandlers returns the handlers for the module
func (m *Module) GetHandlers() extension.Handler {
	return m.h
}

// GetServices returns the services for the module
func (m *Module) GetServices() extension.Service {
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
func (m *Module) GetMetadata() extension.Metadata {
	return extension.Metadata{
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

// RegisterEvents registers events for the module
func (m *Module) subscribeEvents(_ *extension.Manager) {
	// Implement any event subscriptions here
}

// New creates a new instance of the auth module.
func New() extension.Interface {
	return &Module{}
}
