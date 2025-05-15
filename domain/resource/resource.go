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
	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"

	"github.com/gin-gonic/gin"
)

var (
	name             = "resource"
	desc             = "Resource module"
	version          = "1.0.0"
	dependencies     []string
	typeStr          = "module"
	group            = "res"
	enabledDiscovery = false
)

// Module represents the resource module
type Module struct {
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

// Name returns the name of the module
func (m *Module) Name() string {
	return name
}

// PreInit performs any necessary setup before initialization
func (m *Module) PreInit() error {
	m.config = m.GetDefaultConfig()
	return nil
}

// Init initializes the module
func (m *Module) Init(conf *config.Config, em ext.ManagerInterface) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("resource module already initialized")
	}

	m.d, m.cleanup, err = data.New(conf.Data)
	if err != nil {
		return err
	}

	// service discovery
	if conf.Consul != nil {
		m.discovery.address = conf.Consul.Address
		m.discovery.tags = conf.Consul.Discovery.DefaultTags
		m.discovery.meta = conf.Consul.Discovery.DefaultMeta
	}

	// Load config from configuration file if available
	if conf.Viper != nil {
		m.GetConfigFromFile(conf.Viper)
	}

	m.em = em
	m.initialized = true

	return nil
}

// PostInit performs any necessary setup after initialization
func (m *Module) PostInit() error {
	// Create event publisher
	publisher := event.NewPublisher(m.em)

	// Create services
	m.s = service.New(m.d, publisher)

	// Create handlers
	m.h = handler.New(m.s)

	// Start quota monitor if enabled
	if m.config.QuotaManagement.EnableQuotas {
		go m.startQuotaMonitor(m.s.Quota, m.config.QuotaManagement.QuotaCheckInterval)
	}

	// Subscribe to events
	m.subscribeEvents()

	return nil
}

// startQuotaMonitor starts a background goroutine to monitor storage quotas
func (m *Module) startQuotaMonitor(quotaService service.QuotaServiceInterface, intervalStr string) {
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

	logger.Info(ctx, "Started resource quota monitor with interval %s", interval.String())
}

// RegisterRoutes registers routes for the module
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	// Base domain group
	resGroup := r.Group("/"+m.Group(), middleware.AuthenticatedUser)

	resGroup.GET("", m.h.File.List)
	resGroup.POST("", m.h.File.Create)
	resGroup.GET("/:slug", m.h.File.Get)
	resGroup.PUT("/:slug", m.h.File.Update)
	resGroup.DELETE("/:slug", m.h.File.Delete)

	// Advanced endpoints
	resGroup.GET("/search", m.h.File.Search)
	resGroup.GET("/categories", m.h.File.ListCategories)
	resGroup.GET("/tags", m.h.File.ListTags)
	resGroup.GET("/stats", m.h.File.GetStorageStats)

	// Version and thumbnail operations
	resGroup.GET("/:slug/versions", m.h.File.GetVersions)
	resGroup.POST("/:slug/versions", m.h.File.CreateVersion)
	resGroup.POST("/:slug/thumbnail", m.h.File.CreateThumbnail)
	resGroup.PUT("/:slug/access", m.h.File.SetAccessLevel)
	resGroup.POST("/:slug/share", m.h.File.GenerateShareURL)

	// Batch operations
	batch := resGroup.Group("/batch")
	{
		batch.POST("/upload", m.h.Batch.BatchUpload)
		batch.POST("/process", m.h.Batch.BatchProcess)
	}

	// Quota management
	quotas := resGroup.Group("/quotas")
	{
		quotas.GET("", m.h.Quota.GetQuota)
		quotas.PUT("", m.h.Quota.SetQuota)
		quotas.GET("/usage", m.h.Quota.GetUsage)
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

// Status returns the status of the module
func (m *Module) Status() string {
	// Implement logic to check the module status
	return "active"
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
	return []ext.DependencyEntry{}
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

// SubscribeEvents subscribes to events for the module
func (m *Module) subscribeEvents() {
	if m.em == nil {
		return
	}

	// Subscribe to relevant events
	m.em.SubscribeEvent(event.FileCreated, m.handleFileCreated)
	m.em.SubscribeEvent(event.FileDeleted, m.handleFileDeleted)
	m.em.SubscribeEvent(event.StorageQuotaWarning, m.handleQuotaWarning)
	m.em.SubscribeEvent(event.StorageQuotaExceeded, m.handleQuotaExceeded)
}

// Event handlers
func (m *Module) handleFileCreated(data any) {
	eventData, ok := data.(*event.FileEventData)
	if !ok {
		return
	}

	logger.Infof(context.Background(), "File created: %s, size: %d bytes, tenant: %s",
		eventData.Name, eventData.Size, eventData.TenantID)
}

func (m *Module) handleFileDeleted(data any) {
	eventData, ok := data.(*event.FileEventData)
	if !ok {
		return
	}

	logger.Infof(context.Background(), "File deleted: %s, tenant: %s",
		eventData.Name, eventData.TenantID)
}

func (m *Module) handleQuotaWarning(data any) {
	eventData, ok := data.(*event.StorageQuotaEventData)
	if !ok {
		return
	}

	logger.Warnf(context.Background(), "Storage quota warning: tenant %s at %.2f%% (%d/%d bytes)",
		eventData.TenantID, eventData.UsagePercent, eventData.CurrentUsage, eventData.Quota)

	// In a real implementation, you might send notifications to admins or users
}

func (m *Module) handleQuotaExceeded(data any) {
	eventData, ok := data.(*event.StorageQuotaEventData)
	if !ok {
		return
	}

	logger.Errorf(context.Background(), "Storage quota exceeded: tenant %s at %.2f%% (%d/%d bytes)",
		eventData.TenantID, eventData.UsagePercent, eventData.CurrentUsage, eventData.Quota)

	// In a real implementation, you might send urgent notifications or implement cleanup procedures
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

// New creates a new instance of the resource module.
func New() ext.Interface {
	return &Module{}
}
