package payment

import (
	"fmt"
	"ncobase/plugin/payment/data"
	"ncobase/plugin/payment/event"
	"ncobase/plugin/payment/handler"
	"ncobase/plugin/payment/service"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	extp "github.com/ncobase/ncore/extension/plugin"
	ext "github.com/ncobase/ncore/extension/types"
)

var (
	name         = "payment"
	desc         = "Payment plugin for supporting multiple payment channels and subscriptions"
	version      = "1.0.0"
	dependencies []string
	typeStr      = "plugin"
	group        = "pay"
)

// Plugin represents the payment plugin.
type Plugin struct {
	ext.OptionalImpl

	initialized bool
	mu          sync.RWMutex
	em          ext.ManagerInterface
	conf        *config.Config
	cleanup     func(name ...string)

	h *handler.Handler
	s *service.Service
	d *data.Data

	discovery
}

// discovery represents the service discovery
type discovery struct {
	address string
	tags    []string
	meta    map[string]string
}

// init registers the plugin
func init() {
	extp.RegisterPlugin(New(), ext.Metadata{
		Name:         name,
		Version:      version,
		Dependencies: dependencies,
		Description:  desc,
		Type:         typeStr,
		Group:        group,
	})
}

// New returns a new instance of the plugin
func New() *Plugin {
	return &Plugin{}
}

// PreInit performs any necessary setup before initialization
func (p *Plugin) PreInit() error {
	// Register payment providers
	// These will be automatically registered through init() functions
	// in their respective files

	// You could add additional pre-initialization logic here
	return nil
}

// Init initializes the payment plugin with the given config object
func (p *Plugin) Init(conf *config.Config, em ext.ManagerInterface) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.initialized {
		return fmt.Errorf("payment plugin already initialized")
	}

	p.d, p.cleanup, err = data.New(conf.Data, conf.Environment)
	if err != nil {
		return err
	}

	// Service discovery
	if conf.Consul != nil {
		p.discovery.address = conf.Consul.Address
		p.discovery.tags = conf.Consul.Discovery.DefaultTags
		p.discovery.meta = conf.Consul.Discovery.DefaultMeta
	}

	p.em = em
	p.conf = conf
	p.initialized = true

	return nil
}

// PostInit performs any necessary setup after initialization
func (p *Plugin) PostInit() error {

	// Create event publisher
	publisher := event.NewPublisher(p.em)

	// Initialize services
	p.s = service.New(p.d, publisher)

	// Initialize handlers
	p.h = handler.New(p.em, p.s)

	// Subscribe to extension events for dependency refresh
	p.em.SubscribeEvent("exts.user.ready", func(data any) {
		p.h.Event.RefreshDependencies()
	})
	p.em.SubscribeEvent("exts.space.ready", func(data any) {
		p.h.Event.RefreshDependencies()
	})

	return nil
}

// Name returns the name of the plugin
func (p *Plugin) Name() string {
	return name
}

// RegisterRoutes registers routes for the plugin
func (p *Plugin) RegisterRoutes(r *gin.RouterGroup) {
	// Payment domain group
	payGroup := r.Group("/" + p.Group())

	// Channel routes
	channelGroup := payGroup.Group("/channels")
	channelGroup.GET("", p.h.Channel.List)
	channelGroup.POST("", p.h.Channel.Create)
	channelGroup.GET("/:id", p.h.Channel.Get)
	channelGroup.PUT("/:id", p.h.Channel.Update)
	channelGroup.DELETE("/:id", p.h.Channel.Delete)
	channelGroup.PUT("/:id/status", p.h.Channel.ChangeStatus)

	// Order routes
	orderGroup := payGroup.Group("/orders")
	orderGroup.GET("", p.h.Order.List)
	orderGroup.POST("", p.h.Order.Create)
	orderGroup.GET("/:id", p.h.Order.Get)
	orderGroup.GET("/number/:orderNumber", p.h.Order.GetByOrderNumber)
	orderGroup.POST("/:id/payment-url", p.h.Order.GeneratePaymentURL)
	orderGroup.POST("/:id/verify", p.h.Order.VerifyPayment)
	orderGroup.POST("/:id/refund", p.h.Order.RefundPayment)

	// Product routes
	productGroup := payGroup.Group("/products")
	productGroup.GET("", p.h.Product.List)
	productGroup.POST("", p.h.Product.Create)
	productGroup.GET("/:id", p.h.Product.Get)
	productGroup.PUT("/:id", p.h.Product.Update)
	productGroup.DELETE("/:id", p.h.Product.Delete)

	// Subscription routes
	subscriptionGroup := payGroup.Group("/subscriptions")
	subscriptionGroup.GET("", p.h.Subscription.List)
	subscriptionGroup.POST("", p.h.Subscription.Create)
	subscriptionGroup.GET("/:id", p.h.Subscription.Get)
	subscriptionGroup.PUT("/:id", p.h.Subscription.Update)
	subscriptionGroup.POST("/:id/cancel", p.h.Subscription.Cancel)
	subscriptionGroup.GET("/user/:userId", p.h.Subscription.GetByUser)

	// Webhook routes
	webhookGroup := payGroup.Group("/webhooks")
	webhookGroup.POST("/:channel", p.h.Webhook.ProcessWebhook)

	// Log routes
	logGroup := payGroup.Group("/logs")
	logGroup.GET("", p.h.Log.List)
	logGroup.GET("/:id", p.h.Log.Get)
	logGroup.GET("/order/:orderId", p.h.Log.GetByOrder)

	// Utility routes
	payGroup.GET("/providers", p.h.Utility.ListProviders)
	payGroup.GET("/stats", p.h.Utility.GetStats)
}

// GetHandlers returns the handlers for the plugin
func (p *Plugin) GetHandlers() ext.Handler {
	return p.h
}

// GetServices returns the services for the plugin
func (p *Plugin) GetServices() ext.Service {
	return p.s
}

// Cleanup cleans up the plugin
func (p *Plugin) Cleanup() error {
	if p.cleanup != nil {
		p.cleanup(p.Name())
	}
	return nil
}

// GetMetadata returns the metadata of the plugin
func (p *Plugin) GetMetadata() ext.Metadata {
	return ext.Metadata{
		Name:         p.Name(),
		Version:      p.Version(),
		Dependencies: p.Dependencies(),
		Description:  p.Description(),
		Type:         p.Type(),
		Group:        p.Group(),
	}
}

// Version returns the version of the plugin
func (p *Plugin) Version() string {
	return version
}

// Dependencies returns the dependencies of the plugin
func (p *Plugin) Dependencies() []string {
	return dependencies
}

// GetAllDependencies returns all dependencies of the plugin
func (p *Plugin) GetAllDependencies() []ext.DependencyEntry {
	return []ext.DependencyEntry{
		{Name: "user", Type: ext.WeakDependency},
		{Name: "space", Type: ext.WeakDependency},
	}
}

// Description returns the description of the plugin
func (p *Plugin) Description() string {
	return desc
}

// Type returns the type of the plugin
func (p *Plugin) Type() string {
	return typeStr
}

// Group returns the domain group of the plugin belongs
func (p *Plugin) Group() string {
	return group
}

// GetServiceInfo returns service registration info if NeedServiceDiscovery returns true
func (p *Plugin) GetServiceInfo() *ext.ServiceInfo {
	if !p.NeedServiceDiscovery() {
		return nil
	}

	metadata := p.GetMetadata()

	tags := append(p.discovery.tags, metadata.Group, metadata.Type)

	meta := make(map[string]string)
	for k, v := range p.discovery.meta {
		meta[k] = v
	}
	meta["name"] = metadata.Name
	meta["version"] = metadata.Version
	meta["group"] = metadata.Group
	meta["type"] = metadata.Type
	meta["description"] = metadata.Description

	return &ext.ServiceInfo{
		Address: p.discovery.address,
		Tags:    tags,
		Meta:    meta,
	}
}
