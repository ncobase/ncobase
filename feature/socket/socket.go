package socket

import (
	"fmt"
	"ncobase/common/config"
	"ncobase/common/feature"
	"ncobase/common/resp"
	"ncobase/feature/socket/handler"
	"ncobase/feature/socket/service"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	name         = "socket"
	desc         = "Relationship Manager"
	version      = "1.0.0"
	dependencies = []string{"access", "auth", "tenant", "user", "group"}
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
	gs, err := m.getGroupService()
	if err != nil {
		return err
	}
	acs, err := m.getAccessService()
	if err != nil {
		return err
	}
	m.s = service.New(as, us, ts, gs, acs)
	m.h = handler.New(m.s)
	// Subscribe to relevant events
	m.subscribeEvents(m.fm)
	return nil
}

// RegisterRoutes registers routes for the socket
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	// initialize routes, TODO: move to a separate module
	r.GET("/socket/initialize", func(c *gin.Context) {
		err := m.s.Initialize.Execute()
		if err != nil {
			resp.Fail(c.Writer, resp.BadRequest(err.Error()))
			return
		}
		resp.Success(c.Writer)
	})
	// user meshes
	r.GET("/:username/meshes", func(c *gin.Context) {
		resp.Success(c.Writer)
	})

	// websocket routes
	r.GET("/websocket", func(c *gin.Context) {
		m.h.WebSocket.Connect(c.Writer, c.Request)
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
		Description:  desc,
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

// SubscribeEvents subscribes to relevant events
func (m *Module) subscribeEvents(_ *feature.Manager) {
	// Implement any event subscriptions here
}

// New returns a new socket
func New() *Module {
	return &Module{}
}
