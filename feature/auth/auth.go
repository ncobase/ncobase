package auth

import (
	"fmt"
	"ncobase/common/config"
	"ncobase/feature"
	"ncobase/feature/auth/data"
	"ncobase/feature/auth/handler"
	"ncobase/feature/auth/service"
	tenantService "ncobase/feature/tenant/service"
	userService "ncobase/feature/user/service"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	name         = "auth"
	desc         = "auth module"
	version      = "1.0.0"
	dependencies = []string{"user", "tenant"}
)

// Module represents the auth module.
type Module struct {
	initialized bool
	mu          sync.RWMutex
	fm          *feature.Manager
	h           *handler.Handler
	s           *service.Service
	d           *data.Data
	cleanup     func(name ...string)
}

// New creates a new instance of the auth module.
func New() feature.Interface {
	return &Module{}
}

// PreInit performs any necessary setup before initialization
func (m *Module) PreInit() error {
	// Implement any pre-initialization logic here
	return nil
}

// Init initializes the auth module with the given config object
func (m *Module) Init(conf *config.Config, fm *feature.Manager) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("auth module already initialized")
	}

	m.d, m.cleanup, err = data.New(conf.Data)
	if err != nil {
		return err
	}

	m.fm = fm
	m.initialized = true

	return nil
}

// PostInit performs any necessary setup after initialization
func (m *Module) PostInit() error {
	us, err := m.getUserService()
	if err != nil {
		return err
	}

	ts, err := m.getTenantService()
	if err != nil {
		return err
	}

	m.s = service.New(m.d, us, ts)
	m.h = handler.New(m.s)
	return nil
}

// Name returns the name of the module
func (m *Module) Name() string {
	return name
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(e *gin.Engine) {
	// API v1 endpoints
	v1 := e.Group("/v1")
	// Authentication endpoints
	v1.POST("/login", m.h.Auth.Login)
	v1.POST("/register", m.h.Auth.Register)
	v1.POST("/logout", m.h.Auth.Logout)
	// Captcha endpoints
	captcha := v1.Group("/captcha")
	{
		captcha.GET("/generate", m.h.Captcha.GenerateCaptcha)
		captcha.GET("/:captcha", m.h.Captcha.CaptchaStream)
		captcha.POST("/validate", m.h.Captcha.ValidateCaptcha)
	}
	// Authorization endpoints
	authorize := v1.Group("/authorize")
	{
		authorize.POST("/send", m.h.CodeAuth.SendCode)
		authorize.GET("/:code", m.h.CodeAuth.CodeAuth)
	}
}

// GetHandlers returns the handlers for the module
func (m *Module) GetHandlers() feature.Handler {
	return m.h
}

// GetServices returns the services for the module
func (m *Module) GetServices() feature.Service {
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
func (m *Module) GetMetadata() feature.Metadata {
	return feature.Metadata{
		Name:         m.Name(),
		Version:      m.Version(),
		Dependencies: m.Dependencies(),
		Description:  desc,
	}
}

// Version returns the version of the plugin
func (m *Module) Version() string {
	return version
}

// Dependencies returns the dependencies of the plugin
func (m *Module) Dependencies() []string {
	return dependencies
}

// GetUserService returns the user service
func (m *Module) getUserService() (*userService.Service, error) {
	f, err := m.fm.GetService("user")
	if err != nil {
		return nil, fmt.Errorf("failed to get user service: %v", err)
	}

	us, ok := f.(*userService.Service)
	if !ok {
		return nil, fmt.Errorf("user service does not implement UserServiceInterface")
	}

	return us, nil
}

// GetInitService returns the init service
func (m *Module) getTenantService() (*tenantService.Service, error) {
	f, err := m.fm.GetService("tenant")
	if err != nil {
		return nil, fmt.Errorf("failed to get init service: %v", err)
	}
	is, ok := f.(*tenantService.Service)
	if !ok {
		return nil, fmt.Errorf("init service does not implement")
	}
	return is, nil
}
