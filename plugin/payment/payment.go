package payment

import (
	"fmt"
	"ncobase/internal/middleware"
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

	// Webhook callbacks must remain unauthenticated at the route layer because
	// external providers cannot send user tokens. Provider signature and
	// idempotency validation are enforced in the payment service/provider layer.
	webhookGroup := payGroup.Group("/webhooks")
	webhookGroup.POST("/:channel", p.h.Webhook.ProcessWebhook)

	protected := payGroup.Group("", middleware.ValidateContentType(), middleware.RequireAuth())
	readPayments := protected.Group("", middleware.HasAnyPermission("read:payments", "manage:payments", "refund:payments", "admin:payments"))
	managePayments := protected.Group("", middleware.HasAnyPermission("manage:payments", "admin:payments"))
	refundPayments := protected.Group("", middleware.HasAnyPermission("refund:payments", "admin:payments"))
	adminPayments := protected.Group("", middleware.HasPermission("admin:payments"))

	// Channel routes expose provider configuration and require payment management.
	managePayments.GET("/channels", p.h.Channel.List)
	managePayments.POST("/channels", p.h.Channel.Create)
	managePayments.GET("/channels/:id", p.h.Channel.Get)
	managePayments.PUT("/channels/:id", p.h.Channel.Update)
	managePayments.DELETE("/channels/:id", p.h.Channel.Delete)
	managePayments.PUT("/channels/:id/status", p.h.Channel.ChangeStatus)

	// Order routes
	readPayments.GET("/orders", p.h.Order.List)
	managePayments.POST("/orders", p.h.Order.Create)
	readPayments.GET("/orders/number/:orderNumber", p.h.Order.GetByOrderNumber)
	readPayments.GET("/orders/:id", p.h.Order.Get)
	managePayments.POST("/orders/:id/payment-url", p.h.Order.GeneratePaymentURL)
	managePayments.POST("/orders/:id/verify", p.h.Order.VerifyPayment)
	refundPayments.POST("/orders/:id/refund", p.h.Order.RefundPayment)

	// Product routes
	managePayments.GET("/products", p.h.Product.List)
	managePayments.POST("/products", p.h.Product.Create)
	managePayments.GET("/products/:id", p.h.Product.Get)
	managePayments.PUT("/products/:id", p.h.Product.Update)
	managePayments.DELETE("/products/:id", p.h.Product.Delete)

	// Subscription routes
	managePayments.GET("/subscriptions", p.h.Subscription.List)
	managePayments.POST("/subscriptions", p.h.Subscription.Create)
	managePayments.GET("/subscriptions/user/:userId", p.h.Subscription.GetByUser)
	managePayments.GET("/subscriptions/:id", p.h.Subscription.Get)
	managePayments.PUT("/subscriptions/:id", p.h.Subscription.Update)
	managePayments.POST("/subscriptions/:id/cancel", p.h.Subscription.Cancel)

	// Log routes
	adminPayments.GET("/logs", p.h.Log.List)
	adminPayments.GET("/logs/order/:orderId", p.h.Log.GetByOrder)
	adminPayments.GET("/logs/:id", p.h.Log.Get)

	// Utility routes
	readPayments.GET("/providers", p.h.Utility.ListProviders)
	readPayments.GET("/stats", p.h.Utility.GetStats)
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
