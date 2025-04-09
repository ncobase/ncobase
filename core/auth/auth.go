package auth

import (
	"fmt"
	"ncobase/cmd/ncobase/middleware"
	"ncobase/core/auth/data"
	"ncobase/core/auth/handler"
	"ncobase/core/auth/service"
	nec "ncore/ext/core"
	"ncore/pkg/config"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	name             = "auth"
	desc             = "Auth module"
	version          = "1.0.0"
	dependencies     = []string{"access", "tenant", "user"}
	typeStr          = "module"
	group            = "iam"
	enabledDiscovery = false
)

// Module represents the auth module.
type Module struct {
	initialized bool
	mu          sync.RWMutex
	em          nec.ManagerInterface
	h           *handler.Handler
	s           *service.Service
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

// New creates a new instance of the auth module.
func New() nec.Interface {
	return &Module{}
}

// PreInit performs any necessary setup before initialization
func (m *Module) PreInit() error {
	// Implement any pre-initialization logic here
	return nil
}

// Init initializes the auth module with the given config object
func (m *Module) Init(conf *config.Config, em nec.ManagerInterface) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("auth module already initialized")
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
	us, err := m.getUserService()
	if err != nil {
		return err
	}
	ts, err := m.getTenantService()
	if err != nil {
		return err
	}
	as, err := m.getAccessService()
	if err != nil {
		return err
	}
	m.s = service.New(m.d, us, as, ts)
	m.h = handler.New(m.s)
	return nil
}

// Name returns the name of the module
func (m *Module) Name() string {
	return name
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	// Belong domain group
	r = r.Group("/" + m.Group())
	// Authentication endpoints
	r.POST("/login", m.h.Account.Login)
	r.POST("/register", m.h.Account.Register)
	r.POST("/logout", m.h.Account.Logout)
	// Captcha endpoints
	captcha := r.Group("/captcha")
	{
		captcha.GET("/generate", m.h.Captcha.GenerateCaptcha)
		captcha.GET("/:captcha", m.h.Captcha.CaptchaStream)
		captcha.POST("/validate", m.h.Captcha.ValidateCaptcha)
	}
	// Authorization endpoints
	authorize := r.Group("/authorize")
	{
		authorize.POST("/send", m.h.CodeAuth.SendCode)
		authorize.GET("/:code", m.h.CodeAuth.CodeAuth)
	}

	// Account endpoints
	account := r.Group("/account", middleware.AuthenticatedUser)
	{
		account.GET("", m.h.Account.GetMe)
		account.PUT("/password", m.h.Account.UpdatePassword)
		account.GET("/tenant", m.h.Account.Tenant)
		account.GET("/tenants", m.h.Account.Tenants)
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

// Status returns the status of the module
func (m *Module) Status() string {
	return "active"
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
