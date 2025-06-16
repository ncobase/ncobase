package auth

import (
	"fmt"
	"ncobase/auth/data"
	"ncobase/auth/handler"
	"ncobase/auth/service"
	"ncobase/internal/middleware"
	"sync"

	"github.com/ncobase/ncore/config"
	exr "github.com/ncobase/ncore/extension/registry"
	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/security/jwt"

	"github.com/gin-gonic/gin"
)

var (
	name         = "auth"
	desc         = "Auth module"
	version      = "1.0.0"
	dependencies []string
	typeStr      = "module"
	group        = ""
)

// Module represents the auth module.
type Module struct {
	ext.OptionalImpl

	initialized bool
	mu          sync.RWMutex
	em          ext.ManagerInterface
	h           *handler.Handler
	s           *service.Service
	d           *data.Data
	jtm         *jwt.TokenManager
	cleanup     func(name ...string)

	discovery
}

// discovery represents the service discovery
type discovery struct {
	address string
	tags    []string
	meta    map[string]string
}

// init registers the module
func init() {
	exr.RegisterToGroupWithWeakDeps(New(), group, []string{"user", "space", "access"})
}

// New creates a new instance of the auth module.
func New() ext.Interface {
	return &Module{}
}

// Init initializes the auth module with the given config object
func (m *Module) Init(conf *config.Config, em ext.ManagerInterface) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("auth module already initialized")
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

	// token manager
	m.jtm = jwt.NewTokenManager(conf.Auth.JWT.Secret)

	m.em = em
	m.initialized = true

	return nil
}

// PostInit performs any necessary setup after initialization
func (m *Module) PostInit() error {
	m.s = service.New(m.d, m.jtm, m.em)
	m.h = handler.New(m.s)

	// Subscribe to extension events for dependency refresh
	m.em.SubscribeEvent("exts.user.ready", func(data any) {
		m.s.RefreshDependencies()
	})
	m.em.SubscribeEvent("exts.space.ready", func(data any) {
		m.s.RefreshDependencies()
	})
	m.em.SubscribeEvent("exts.access.ready", func(data any) {
		m.s.RefreshDependencies()
	})

	// Subscribe to all extensions registration event
	m.em.SubscribeEvent("exts.all.registered", func(data any) {
		m.s.RefreshDependencies()
	})

	return nil
}

// Name returns the name of the module
func (m *Module) Name() string {
	return name
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	// Belong domain group
	authGroup := r.Group("/" + m.Group())

	// Authentication endpoints
	authGroup.POST("/login", m.h.Account.Login)
	authGroup.POST("/register", m.h.Account.Register)
	authGroup.POST("/logout", m.h.Account.Logout)

	// Captcha endpoints
	captcha := authGroup.Group("/captcha")
	{
		captcha.GET("/generate", m.h.Captcha.GenerateCaptcha)
		captcha.GET("/:captcha", m.h.Captcha.CaptchaStream)
		captcha.POST("/validate", m.h.Captcha.ValidateCaptcha)
	}

	// Authorization endpoints
	authorize := authGroup.Group("/authorize")
	{
		authorize.POST("/send", m.h.CodeAuth.SendCode)
		authorize.GET("/:code", m.h.CodeAuth.CodeAuth)
	}

	// Account endpoints
	account := authGroup.Group("/account", middleware.AuthenticatedUser)
	{
		account.GET("", m.h.Account.GetMe)
		account.PUT("/password", m.h.Account.UpdatePassword)
		account.GET("/space", m.h.Account.Space)
		account.GET("/spaces", m.h.Account.Spaces)
	}

	// Token endpoints
	r.POST("/refresh-token", m.h.Account.RefreshToken)
	r.GET("/token-status", m.h.Account.TokenStatus)

	// Session endpoints
	sessions := authGroup.Group("/sessions", middleware.AuthenticatedUser)
	{
		sessions.GET("", m.h.Session.List)
		sessions.GET("/:session_id", m.h.Session.Get)
		sessions.DELETE("/:session_id", m.h.Session.Delete)
		sessions.POST("/deactivate-all", m.h.Session.DeactivateAll)
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

// GetTokenManager returns the JWT token manager
func (m *Module) GetTokenManager() *jwt.TokenManager {
	return m.jtm
}

// GetSessionService returns the session service
func (m *Module) GetSessionService() service.SessionServiceInterface {
	if m.s == nil {
		return nil
	}
	return m.s.Session
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

// GetAllDependencies returns all dependencies with their types
func (m *Module) GetAllDependencies() []ext.DependencyEntry {
	return []ext.DependencyEntry{
		{Name: "user", Type: ext.WeakDependency},
		{Name: "organization", Type: ext.WeakDependency},
		{Name: "space", Type: ext.WeakDependency},
		{Name: "access", Type: ext.WeakDependency},
	}
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

// GetServiceInfo returns service registration info if NeedServiceDiscovery returns true
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
