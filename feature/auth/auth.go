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
	usi, err := m.getUserService(m.fm)
	if err != nil {
		return err
	}

	tsi, err := m.getTenantService(m.fm)
	if err != nil {
		return err
	}

	m.s = service.New(m.d, usi, tsi)
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
func (m *Module) GetHandlers() map[string]feature.Handler {
	return map[string]feature.Handler{
		"auth": m.h,
	}
}

// GetServices returns the services for the module
func (m *Module) GetServices() map[string]feature.Service {
	return map[string]feature.Service{
		"auth": m.s,
	}
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
func (m *Module) getUserService(fm *feature.Manager) (userService.UserServiceInterface, error) {
	usi, err := fm.GetService("user", "user")
	if err != nil {
		return nil, fmt.Errorf("failed to get user service: %v", err)
	}

	userServiceImpl, ok := usi.(userService.UserServiceInterface)
	if !ok {
		return nil, fmt.Errorf("user service does not implement UserServiceInterface")
	}

	return userServiceImpl, nil
}

// GetTenantService returns the tenant service
func (m *Module) getTenantService(fm *feature.Manager) (tenantService.TenantServiceInterface, error) {
	tsi, err := fm.GetService("tenant", "tenant")
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant service: %v", err)
	}

	// type assertion to ensure we have the correct interface
	tenantServiceImpl, ok := tsi.(tenantService.TenantServiceInterface)
	if !ok {
		return nil, fmt.Errorf("tenant service does not implement TenantServiceInterface")
	}

	return tenantServiceImpl, nil

}
