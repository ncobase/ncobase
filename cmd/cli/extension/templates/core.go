package templates

import "fmt"

func CoreTemplate(name string) string {
	return fmt.Sprintf(`package %s

import (
	"fmt"
	"ncobase/common/config"
	"ncobase/common/extension"
	"ncobase/core/%s/data"
	"ncobase/core/%s/handler"
	"ncobase/core/%s/service"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	name         = "%s"
	desc         = "%s module"
	version      = "1.0.0"
	dependencies []string
	typeStr      = "module"
	group        = ""
)

// Module represents the %s module.
type Module struct {
	initialized bool
	mu          sync.RWMutex
	em          *extension.Manager
	conf        *config.Config
	h           *handler.Handler
	s           *service.Service
	d           *data.Data
	cleanup     func(name ...string)
}

// New creates a new instance of the %s module.
func New() extension.Interface {
	return &Module{}
}

// PreInit performs any necessary setup before initialization
func (m *Module) PreInit() error {
	// Implement any pre-initialization logic here
	return nil
}

// Init initializes the %s module with the given config object
func (m *Module) Init(conf *config.Config, em *extension.Manager) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("%s module already initialized")
	}

	m.d, m.cleanup, err = data.New(conf.Data)
	if err != nil {
		return err
	}

	m.em = em
	m.conf = conf
	m.initialized = true

	return nil
}

// PostInit performs any necessary setup after initialization
func (m *Module) PostInit() error {
	m.s = service.New(m.conf, m.d)
	m.h = handler.New(m.s)

	return nil
}

// Name returns the name of the module
func (m *Module) Name() string {
	return name
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	// Implement your route registration logic here
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

// Status returns the status of the module
func (m *Module) Status() string {
	return "active"
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
`, name, name, name, name, name, name, name, name, name, name)
}
