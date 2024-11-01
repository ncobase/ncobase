package realtime

import (
	"fmt"
	"ncobase/common/config"
	"ncobase/common/feature"
	"ncobase/common/resp"
	"ncobase/domain/realtime/handler"
	"ncobase/domain/realtime/service"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	name         = "realtime"
	desc         = "Realtime module, provides realtime communication"
	version      = "1.0.0"
	dependencies = []string{"access", "auth", "tenant", "user", "space"}
	typeStr      = "module"
	group        = "rt" // belongs to core module
)

// Module represents the socket
type Module struct {
	initialized bool
	mu          sync.RWMutex
	fm          *feature.Manager
	conf        *config.Config
	s           *service.Service
	h           *handler.Handler
	cleanup     func(name ...string)
}

// Name returns the name of the socket
func (m *Module) Name() string {
	return name
}

// PreInit performs any necessary setup before initialization
func (m *Module) PreInit() error {
	// Implement any pre-initialization logic here
	return nil
}

// Init initializes the socket
func (m *Module) Init(conf *config.Config, fm *feature.Manager) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("socket already initialized")
	}

	m.fm = fm
	m.conf = conf
	m.initialized = true

	return nil
}

// PostInit performs any necessary setup after initialization
func (m *Module) PostInit() error {
	m.s = service.New()
	m.h = handler.New(m.s)
	// Subscribe to relevant events
	m.subscribeEvents(m.fm)
	return nil
}

// RegisterRoutes registers routes for the socket
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	// Belong domain group
	r = r.Group("/" + m.Group())
	// WebSocket endpoints
	r.GET("/socket", func(c *gin.Context) {
		m.h.WebSocket.Connect(c.Writer, c.Request)
	})

	// Notification endpoints
	r.GET("/notification", func(c *gin.Context) {
		resp.Success(c.Writer, nil)
	})
}

// GetHandlers returns the handlers for the socket
func (m *Module) GetHandlers() feature.Handler {
	return m.h
}

// GetServices returns the services for the socket
func (m *Module) GetServices() feature.Service {
	return m.s
}

// PreCleanup performs any necessary cleanup before the main cleanup
func (m *Module) PreCleanup() error {
	// Implement any pre-cleanup logic here
	return nil
}

// Cleanup cleans up the socket
func (m *Module) Cleanup() error {
	if m.cleanup != nil {
		m.cleanup(m.Name())
	}
	return nil
}

// GetMetadata returns the metadata of the socket
func (m *Module) GetMetadata() feature.Metadata {
	return feature.Metadata{
		Name:         m.Name(),
		Version:      m.Version(),
		Dependencies: m.Dependencies(),
		Description:  m.Description(),
		Type:         m.Type(),
		Group:        m.Group(),
	}
}

// Status returns the status of the socket
func (m *Module) Status() string {
	// Implement logic to check the socket status
	return "active"
}

// Version returns the version of the socket
func (m *Module) Version() string {
	return version
}

// Dependencies returns the dependencies of the socket
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
func (m *Module) subscribeEvents(_ *feature.Manager) {
	// Implement any event subscriptions here
}

// New returns a new socket
func New() *Module {
	return &Module{}
}
