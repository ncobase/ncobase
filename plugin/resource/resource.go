package resource

import (
	"context"
	"fmt"
	rConfig "ncobase/resource/config"
	"ncobase/resource/data"
	"ncobase/resource/event"
	"ncobase/resource/handler"
	"ncobase/resource/router"
	"ncobase/resource/service"
	"sync"
	"time"

	"github.com/ncobase/ncore/config"
	extp "github.com/ncobase/ncore/extension/plugin"
	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"

	"github.com/gin-gonic/gin"
)

var (
	name         = "resource"
	desc         = "Resource plugin, built-in"
	version      = "1.0.0"
	dependencies []string
	typeStr      = "plugin"
	group        = "res"
)

// Plugin represents resource plugin
type Plugin struct {
	ext.OptionalImpl

	initialized     bool
	mu              sync.RWMutex
	em              ext.ManagerInterface
	cleanup         func(name ...string)
	eventSubscriber event.SubscriberInterface

	c *rConfig.Config
	d *data.Data
	s *service.Service
	h *handler.Handler
	r *router.Router

	discovery
}

// discovery represents service discovery
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

// New returns new plugin instance
func New() *Plugin {
	return &Plugin{}
}

// Name returns plugin name
func (p *Plugin) Name() string {
	return name
}

// PreInit performs setup before initialization
func (p *Plugin) PreInit() error {
	p.c = rConfig.New()
	return nil
}

// Init initializes the plugin
func (p *Plugin) Init(conf *config.Config, em ext.ManagerInterface) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.initialized {
		return fmt.Errorf("resource plugin already initialized")
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

	// Load config from file
	if conf.Viper != nil {
		p.c.LoadFromViper(conf.Viper)
	}

	p.em = em
	p.initialized = true

	return nil
}

// PostInit performs setup after initialization
func (p *Plugin) PostInit() error {
	// Create event publisher
	publisher := event.NewPublisher(p.em)

	// Create services
	p.s = service.New(p.em, p.d, publisher)

	// Create handlers
	p.h = handler.New(p.s)

	// Create event subscriber
	p.eventSubscriber = event.NewSubscriber()

	// Set quota updater for event handler
	p.eventSubscriber.SetQuotaUpdater(p.s.Quota)

	// Start quota monitor if enabled
	if p.c.QuotaManagement.EnableQuotas {
		go p.startQuotaMonitor(p.s.Quota, p.c.QuotaManagement.QuotaCheckInterval)
	}

	// Subscribe to events
	p.subscribeEvents()

	// Subscribe to dependency refresh events
	p.subscribeDependencyEvents()

	p.r = router.New(p.h)

	return nil
}

// startQuotaMonitor starts background quota monitoring
func (p *Plugin) startQuotaMonitor(quotaService service.QuotaServiceInterface, intervalStr string) {
	ctx := context.Background()

	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		logger.Warnf(ctx, "Invalid quota check interval, using default 24h: %v", err)
		interval = 24 * time.Hour
	}

	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := quotaService.MonitorQuota(ctx); err != nil {
					logger.Errorf(ctx, "Error in quota monitoring: %v", err)
				}
			}
		}
	}()
}

// RegisterRoutes registers plugin routes
func (p *Plugin) RegisterRoutes(r *gin.RouterGroup) {
	p.r.Register(r, p.Group())
}

// GetHandlers returns plugin handlers
func (p *Plugin) GetHandlers() ext.Handler {
	return p.h
}

// GetServices returns plugin services
func (p *Plugin) GetServices() ext.Service {
	return p.s
}

// Cleanup cleans up plugin resources
func (p *Plugin) Cleanup() error {
	// Unsubscribe from events
	if p.eventSubscriber != nil && p.em != nil {
		p.eventSubscriber.Unsubscribe(p.em)
	}

	if p.cleanup != nil {
		p.cleanup(p.Name())
	}
	return nil
}

// GetMetadata returns plugin metadata
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

// Version returns plugin version
func (p *Plugin) Version() string {
	return version
}

// Dependencies returns plugin dependencies
func (p *Plugin) Dependencies() []string {
	return dependencies
}

// GetAllDependencies returns all dependencies with types
func (p *Plugin) GetAllDependencies() []ext.DependencyEntry {
	return []ext.DependencyEntry{}
}

// Description returns plugin description
func (p *Plugin) Description() string {
	return desc
}

// Type returns plugin type
func (p *Plugin) Type() string {
	return typeStr
}

// Group returns plugin domain group
func (p *Plugin) Group() string {
	return group
}

// subscribeEvents subscribes to resource-specific events
func (p *Plugin) subscribeEvents() {
	if p.eventSubscriber != nil && p.em != nil {
		p.eventSubscriber.Subscribe(p.em)
	}
}

// subscribeDependencyEvents subscribes to dependency-related events
func (p *Plugin) subscribeDependencyEvents() {
	if p.em == nil {
		return
	}

	// Subscribe to dependency refresh events
	p.em.SubscribeEvent("exts.space.ready", func(data any) {
		if p.s != nil {
			p.s.RefreshDependencies()
		}
	})

	p.em.SubscribeEvent("exts.all.registered", func(data any) {
		if p.s != nil {
			p.s.RefreshDependencies()
		}
	})
}

// GetServiceInfo returns service registration info
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
