package resource

import (
	"context"
	"fmt"
	"ncobase/cmd/ncobase/middleware"
	"ncobase/resource/data"
	"ncobase/resource/event"
	"ncobase/resource/handler"
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

// Plugin represents the resource plugin
type Plugin struct {
	ext.OptionalImpl

	initialized bool
	mu          sync.RWMutex
	em          ext.ManagerInterface
	s           *service.Service
	h           *handler.Handler
	d           *data.Data
	cleanup     func(name ...string)
	config      *Config

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

// Name returns the name of the plugin
func (p *Plugin) Name() string {
	return name
}

// PreInit performs any necessary setup before initialization
func (p *Plugin) PreInit() error {
	p.config = p.GetDefaultConfig()
	return nil
}

// Init initializes the plugin
func (p *Plugin) Init(conf *config.Config, em ext.ManagerInterface) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.initialized {
		return fmt.Errorf("resource plugin already initialized")
	}

	p.d, p.cleanup, err = data.New(conf.Data)
	if err != nil {
		return err
	}

	// service discovery
	if conf.Consul != nil {
		p.discovery.address = conf.Consul.Address
		p.discovery.tags = conf.Consul.Discovery.DefaultTags
		p.discovery.meta = conf.Consul.Discovery.DefaultMeta
	}

	// Load config from configuration file if available
	if conf.Viper != nil {
		p.GetConfigFromFile(conf.Viper)
	}

	p.em = em
	p.initialized = true

	return nil
}

// PostInit performs any necessary setup after initialization
func (p *Plugin) PostInit() error {
	// Create event publisher
	publisher := event.NewPublisher(p.em)

	// Create services
	p.s = service.New(p.em, p.d, publisher)

	// Create handlers
	p.h = handler.New(p.s)

	// Start quota monitor if enabled
	if p.config.QuotaManagement.EnableQuotas {
		go p.startQuotaMonitor(p.s.Quota, p.config.QuotaManagement.QuotaCheckInterval)
	}

	// Subscribe to events
	p.subscribeEvents()

	return nil
}

// startQuotaMonitor starts a background goroutine to monitor storage quotas
func (p *Plugin) startQuotaMonitor(quotaService service.QuotaServiceInterface, intervalStr string) {
	ctx := context.Background()

	// Parse check interval
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		logger.Warnf(ctx, "Invalid quota check interval, using default 24h: %v", err)
		interval = 24 * time.Hour
	}

	ticker := time.NewTicker(interval)

	go func() {
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

// RegisterRoutes registers routes for the plugin
func (p *Plugin) RegisterRoutes(r *gin.RouterGroup) {
	// Base domain group
	resGroup := r.Group("/"+p.Group(), middleware.AuthenticatedUser)

	resGroup.GET("", p.h.File.List)
	resGroup.POST("", p.h.File.Create)
	resGroup.GET("/:slug", p.h.File.Get)
	resGroup.PUT("/:slug", p.h.File.Update)
	resGroup.DELETE("/:slug", p.h.File.Delete)

	// Advanced endpoints
	resGroup.GET("/search", p.h.File.Search)
	resGroup.GET("/categories", p.h.File.ListCategories)
	resGroup.GET("/tags", p.h.File.ListTags)
	resGroup.GET("/stats", p.h.File.GetStorageStats)

	// Version and thumbnail operations
	resGroup.GET("/:slug/versions", p.h.File.GetVersions)
	resGroup.POST("/:slug/versions", p.h.File.CreateVersion)
	resGroup.POST("/:slug/thumbnail", p.h.File.CreateThumbnail)
	resGroup.PUT("/:slug/access", p.h.File.SetAccessLevel)
	resGroup.POST("/:slug/share", p.h.File.GenerateShareURL)

	// Batch operations
	batch := resGroup.Group("/batch")
	{
		batch.POST("/upload", p.h.Batch.BatchUpload)
		batch.POST("/process", p.h.Batch.BatchProcess)
	}

	// Quota management
	quotas := resGroup.Group("/quotas")
	{
		quotas.GET("", p.h.Quota.GetQuota)
		quotas.PUT("", p.h.Quota.SetQuota)
		quotas.GET("/usage", p.h.Quota.GetUsage)
	}
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

// GetAllDependencies returns all dependencies with their types
func (p *Plugin) GetAllDependencies() []ext.DependencyEntry {
	return []ext.DependencyEntry{
		{Name: "tenant", Type: ext.WeakDependency},
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

// SubscribeEvents subscribes to events for the plugin
func (p *Plugin) subscribeEvents() {
	if p.em == nil {
		return
	}

	// Subscribe to relevant events
	p.em.SubscribeEvent(event.FileCreated, p.handleFileCreated)
	p.em.SubscribeEvent(event.FileDeleted, p.handleFileDeleted)
	p.em.SubscribeEvent(event.StorageQuotaWarning, p.handleQuotaWarning)
	p.em.SubscribeEvent(event.StorageQuotaExceeded, p.handleQuotaExceeded)

	// Subscribe to tenant module events for dependency refresh
	p.em.SubscribeEvent("exts.tenant.ready", func(data any) {
		p.s.RefreshDependencies()
	})

	// Subscribe to all extensions registration event
	p.em.SubscribeEvent("exts.all.registered", func(data any) {
		p.s.RefreshDependencies()
	})
}

// Event handlers
func (p *Plugin) handleFileCreated(data any) {
	eventData, ok := data.(*event.FileEventData)
	if !ok {
		return
	}

	logger.Infof(context.Background(), "File created: %s, size: %d bytes, tenant: %s",
		eventData.Name, eventData.Size, eventData.TenantID)
}

func (p *Plugin) handleFileDeleted(data any) {
	eventData, ok := data.(*event.FileEventData)
	if !ok {
		return
	}

	logger.Infof(context.Background(), "File deleted: %s, tenant: %s",
		eventData.Name, eventData.TenantID)

	// Update usage in quota service when file is deleted
	if p.s != nil && p.s.Quota != nil && eventData.Size > 0 {
		ctx := context.Background()
		// Use the interface method to update usage (negative delta for deletion)
		err := p.s.Quota.UpdateUsage(ctx, eventData.TenantID, "storage", -int64(eventData.Size))
		if err != nil {
			logger.Warnf(ctx, "Failed to update quota usage after file deletion: %v", err)
		}
	}
}

func (p *Plugin) handleQuotaWarning(data any) {
	eventData, ok := data.(*event.StorageQuotaEventData)
	if !ok {
		return
	}

	logger.Warnf(context.Background(), "Storage quota warning: tenant %s at %.2f%% (%d/%d bytes)",
		eventData.TenantID, eventData.UsagePercent, eventData.CurrentUsage, eventData.Quota)

	// In a real implementation, you might send notifications to admins or users
}

func (p *Plugin) handleQuotaExceeded(data any) {
	eventData, ok := data.(*event.StorageQuotaEventData)
	if !ok {
		return
	}

	logger.Errorf(context.Background(), "Storage quota exceeded: tenant %s at %.2f%% (%d/%d bytes)",
		eventData.TenantID, eventData.UsagePercent, eventData.CurrentUsage, eventData.Quota)

	// In a real implementation, you might send urgent notifications or implement cleanup procedures
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
