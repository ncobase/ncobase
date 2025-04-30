package payment

import (
	"context"
	"fmt"
	"ncobase/core/payment/data"
	"ncobase/core/payment/event"
	"ncobase/core/payment/handler"
	"ncobase/core/payment/service"
	tenantService "ncobase/core/tenant/service"
	userService "ncobase/core/user/service"
	"sync"

	"github.com/ncobase/ncore/config"
	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"

	"github.com/gin-gonic/gin"
)

var (
	name             = "payment"
	desc             = "Payment module for supporting multiple payment channels and subscriptions"
	version          = "1.0.0"
	dependencies     = []string{"user", "tenant"}
	typeStr          = "module"
	group            = "pay"
	enabledDiscovery = false
)

// Module represents the payment module.
type Module struct {
	initialized bool
	mu          sync.RWMutex
	em          ext.ManagerInterface
	conf        *config.Config
	h           *handler.Handler
	s           *service.Service
	d           *data.Data
	cleanup     func(name ...string)

	// Internal services
	userService   *userService.Service
	tenantService *tenantService.Service

	// Event system
	eventFactory *event.Factory

	discovery
}

// discovery represents the service discovery
type discovery struct {
	address string
	tags    []string
	meta    map[string]string
}

// New creates a new instance of the payment module.
func New() ext.Interface {
	return &Module{
		eventFactory: event.NewFactory(),
	}
}

// PreInit performs any necessary setup before initialization
func (m *Module) PreInit() error {
	// Register payment providers
	// These will be automatically registered through init() functions
	// in their respective files

	// You could add additional pre-initialization logic here
	return nil
}

// Init initializes the payment module with the given config object
func (m *Module) Init(conf *config.Config, em ext.ManagerInterface) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("payment module already initialized")
	}

	m.d, m.cleanup, err = data.New(conf.Data)
	if err != nil {
		return err
	}

	// Service discovery
	if conf.Consul != nil {
		m.discovery.address = conf.Consul.Address
		m.discovery.tags = conf.Consul.Discovery.DefaultTags
		m.discovery.meta = conf.Consul.Discovery.DefaultMeta
	}

	m.em = em
	m.conf = conf
	m.initialized = true

	return nil
}

// PostInit performs any necessary setup after initialization
func (m *Module) PostInit() error {
	var err error

	// Get internal services
	m.userService, err = m.getUserService()
	if err != nil {
		logger.Warnf(context.Background(), "Failed to get user service: %v", err)
		// Continue without user service
	}

	m.tenantService, err = m.getTenantService()
	if err != nil {
		logger.Warnf(context.Background(), "Failed to get tenant service: %v", err)
		// Continue without tenant service
	}

	// Create event publisher
	publisher := m.eventFactory.CreatePublisher(m.em)

	// Initialize services
	m.s = service.New(m.d, publisher)

	// Register event handlers
	if m.em != nil {
		handlerProvider := service.NewEventProvider(m.s, m.userService, m.tenantService)
		registrar := m.eventFactory.CreateRegistrar(m.em)
		registrar.RegisterHandlers(handlerProvider)
	}

	// Initialize handlers
	m.h = handler.New(m.s)

	return nil
}

// Name returns the name of the module
func (m *Module) Name() string {
	return name
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	// Payment domain group
	payGroup := r.Group("/" + m.Group())

	// Channel routes
	channelGroup := payGroup.Group("/channels")
	channelGroup.GET("", m.h.Channel.List)
	channelGroup.POST("", m.h.Channel.Create)
	channelGroup.GET("/:id", m.h.Channel.Get)
	channelGroup.PUT("/:id", m.h.Channel.Update)
	channelGroup.DELETE("/:id", m.h.Channel.Delete)
	channelGroup.PUT("/:id/status", m.h.Channel.ChangeStatus)

	// Order routes
	orderGroup := payGroup.Group("/orders")
	orderGroup.GET("", m.h.Order.List)
	orderGroup.POST("", m.h.Order.Create)
	orderGroup.GET("/:id", m.h.Order.Get)
	orderGroup.GET("/number/:orderNumber", m.h.Order.GetByOrderNumber)
	orderGroup.POST("/:id/payment-url", m.h.Order.GeneratePaymentURL)
	orderGroup.POST("/:id/verify", m.h.Order.VerifyPayment)
	orderGroup.POST("/:id/refund", m.h.Order.RefundPayment)

	// Product routes
	productGroup := payGroup.Group("/products")
	productGroup.GET("", m.h.Product.List)
	productGroup.POST("", m.h.Product.Create)
	productGroup.GET("/:id", m.h.Product.Get)
	productGroup.PUT("/:id", m.h.Product.Update)
	productGroup.DELETE("/:id", m.h.Product.Delete)

	// Subscription routes
	subscriptionGroup := payGroup.Group("/subscriptions")
	subscriptionGroup.GET("", m.h.Subscription.List)
	subscriptionGroup.POST("", m.h.Subscription.Create)
	subscriptionGroup.GET("/:id", m.h.Subscription.Get)
	subscriptionGroup.PUT("/:id", m.h.Subscription.Update)
	subscriptionGroup.POST("/:id/cancel", m.h.Subscription.Cancel)
	subscriptionGroup.GET("/user/:userId", m.h.Subscription.GetByUser)

	// Webhook routes
	webhookGroup := payGroup.Group("/webhooks")
	webhookGroup.POST("/:channel", m.h.Webhook.ProcessWebhook)

	// Log routes
	logGroup := payGroup.Group("/logs")
	logGroup.GET("", m.h.Log.List)
	logGroup.GET("/:id", m.h.Log.Get)
	logGroup.GET("/order/:orderId", m.h.Log.GetByOrder)

	// Utility routes
	payGroup.GET("/providers", m.h.Utility.ListProviders)
	payGroup.GET("/stats", m.h.Utility.GetStats)
}

// GetHandlers returns the handlers for the module
func (m *Module) GetHandlers() ext.Handler {
	return m.h
}

// GetServices returns the services for the module
func (m *Module) GetServices() ext.Service {
	return m.s
}

// PreCleanup performs any necessary cleanup before the main cleanup
func (m *Module) PreCleanup() error {
	// Cleanup any resources here
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
	return ext.StatusActive
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

// NeedServiceDiscovery returns if the module needs to be registered as a service
func (m *Module) NeedServiceDiscovery() bool {
	return enabledDiscovery
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
