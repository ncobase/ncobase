package linker

import (
	"fmt"
	"ncobase/common/config"
	"ncobase/common/feature"
	"ncobase/feature/linker/service"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	name         = "linker"
	desc         = "Relationship Manager"
	version      = "1.0.0"
	dependencies []string
)

// Linker represents the linker
type Linker struct {
	initialized bool
	mu          sync.RWMutex
	fm          *feature.Manager
	s           *service.Service
	cleanup     func(name ...string)
}

// Name returns the name of the linker
func (l *Linker) Name() string {
	return name
}

// PreInit performs any necessary setup before initialization
func (l *Linker) PreInit() error {
	// Implement any pre-initialization logic here
	return nil
}

// Init initializes the linker
func (l *Linker) Init(conf *config.Config, fm *feature.Manager) (err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.initialized {
		return fmt.Errorf("linker already initialized")
	}

	l.fm = fm
	l.initialized = true

	return nil
}

// PostInit performs any necessary setup after initialization
func (l *Linker) PostInit() error {

	as, err := l.getAuthService()
	if err != nil {
		return err
	}
	us, err := l.getUserService()
	if err != nil {
		return err
	}
	ts, err := l.getTenantService()
	if err != nil {
		return err
	}
	acs, err := l.getAccessService()
	if err != nil {
		return err
	}
	l.s = service.New(as, us, ts, acs)
	// Subscribe to relevant events
	l.subscribeEvents(l.fm)
	// initialize data
	err = l.s.InitData()
	if err != nil {
		return err
	}
	return nil
}

// RegisterRoutes registers routes for the linker
func (l *Linker) RegisterRoutes(_ *gin.Engine) {}

// GetHandlers returns the handlers for the linker
func (l *Linker) GetHandlers() feature.Handler {
	return nil
}

// GetServices returns the services for the linker
func (l *Linker) GetServices() feature.Service {
	return nil
}

// PreCleanup performs any necessary cleanup before the main cleanup
func (l *Linker) PreCleanup() error {
	// Implement any pre-cleanup logic here
	return nil
}

// Cleanup cleans up the linker
func (l *Linker) Cleanup() error {
	if l.cleanup != nil {
		l.cleanup(l.Name())
	}
	return nil
}

// GetMetadata returns the metadata of the linker
func (l *Linker) GetMetadata() feature.Metadata {
	return feature.Metadata{
		Name:         l.Name(),
		Version:      l.Version(),
		Dependencies: l.Dependencies(),
		Description:  desc,
	}
}

// Status returns the status of the linker
func (l *Linker) Status() string {
	// Implement logic to check the linker status
	return "active"
}

// Version returns the version of the linker
func (l *Linker) Version() string {
	return version
}

// Dependencies returns the dependencies of the linker
func (l *Linker) Dependencies() []string {
	return dependencies
}

// SubscribeEvents subscribes to relevant events
func (l *Linker) subscribeEvents(_ *feature.Manager) {
	// Implement any event subscriptions here
}

// New returns a new linker
func New() *Linker {
	return &Linker{}
}
