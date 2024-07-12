package feature

import (
	"context"
	"fmt"
	"ncobase/common/config"
	"ncobase/common/ecode"
	"ncobase/common/log"
	"ncobase/common/resp"
	"ncobase/helper"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// Manager represents a feature / plugin manager
type Manager struct {
	features    map[string]*Wrapper
	conf        *config.Config
	mu          sync.RWMutex
	initialized bool
	eventBus    *EventBus
}

// NewManager creates a new feature / plugin manager
func NewManager(conf *config.Config) *Manager {
	return &Manager{
		features: make(map[string]*Wrapper),
		conf:     conf,
		eventBus: NewEventBus(),
	}
}

// Register registers a feature
func (m *Manager) Register(f Interface) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("cannot register feature after initialization")
	}

	name := f.Name()
	if _, exists := m.features[name]; exists {
		return fmt.Errorf("feature %s already registered", name)
	}

	m.features[name] = &Wrapper{
		Metadata: f.GetMetadata(),
		Instance: f,
	}

	log.Infof(context.Background(), "Feature %s registered successfully", name)
	return nil
}

// LoadPlugins loads all plugins based on the current configuration
func (m *Manager) LoadPlugins() error {
	if helper.IsPluginDevMode(m.conf) {
		return m.loadPluginsInDev()
	}
	return m.loadPluginsInProd()
}

// loadPluginsInProd loads plugins in production mode
func (m *Manager) loadPluginsInProd() error {
	fc := m.conf.Feature
	fd := fc.Path

	pds, err := filepath.Glob(filepath.Join(fd, "*.so"))
	if err != nil {
		log.Errorf(context.Background(), "failed to list plugin files: %v", err)
		return err
	}

	for _, pp := range pds {
		pluginName := strings.TrimSuffix(filepath.Base(pp), ".so")
		if !m.shouldLoadPlugin(pluginName) {
			log.Infof(context.Background(), "Skipping plugin %s based on configuration", pluginName)
			continue
		}
		if err := m.loadPlugin(pp); err != nil {
			log.Errorf(context.Background(), "Failed to load plugin %s: %v", pluginName, err)
			return err
		}
	}

	return nil
}

// loadPluginsInDev loads plugins in development mode
func (m *Manager) loadPluginsInDev() error {
	plugins := GetRegisteredPlugins()

	for _, c := range plugins {
		if err := c.Instance.PreInit(); err != nil {
			log.Errorf(context.Background(), "Failed pre-initialization of plugin %s: %v", c.Metadata.Name, err)
			continue
		}
		if err := c.Instance.Init(m.conf, m); err != nil {
			log.Errorf(context.Background(), "Failed to initialize plugin %s: %v", c.Metadata.Name, err)
			continue
		}
		if err := c.Instance.PostInit(); err != nil {
			log.Errorf(context.Background(), "Failed post-initialization of plugin %s: %v", c.Metadata.Name, err)
			continue
		}
		m.features[c.Metadata.Name] = c
		log.Infof(context.Background(), "Plugin %s loaded and initialized successfully", c.Metadata.Name)
	}

	return nil
}

// shouldLoadPlugin returns true if the plugin should be loaded
func (m *Manager) shouldLoadPlugin(name string) bool {
	fc := m.conf.Feature

	if len(fc.Includes) > 0 {
		for _, include := range fc.Includes {
			if include == name {
				return true
			}
		}
		return false
	}

	if len(fc.Excludes) > 0 {
		for _, exclude := range fc.Excludes {
			if exclude == name {
				return false
			}
		}
	}

	return true
}

// loadPlugin loads a single plugin
func (m *Manager) loadPlugin(path string) error {
	name := strings.TrimSuffix(filepath.Base(path), ".so")
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.features[name]; exists {
		return nil // plugin already loaded
	}

	if err := LoadPlugin(path, m); err != nil {
		log.Errorf(context.Background(), "failed to load plugin %s: %v", name, err)
		return err
	}

	loadedPlugin := GetPlugin(name)
	if loadedPlugin != nil {
		m.features[name] = loadedPlugin
		log.Infof(context.Background(), "Plugin %s loaded successfully", name)
	}

	return nil
}

// UnloadPlugin unloads a single feature
func (m *Manager) UnloadPlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	feature, exists := m.features[name]
	if !exists {
		return fmt.Errorf("feature %s not found", name)
	}

	if err := feature.Instance.PreCleanup(); err != nil {
		log.Errorf(context.Background(), "failed pre-cleanup of feature %s: %v", name, err)
	}

	if err := feature.Instance.Cleanup(); err != nil {
		log.Errorf(context.Background(), "failed to cleanup feature %s: %v", name, err)
		return err
	}

	delete(m.features, name)
	return nil
}

// InitFeatures initializes all registered features
func (m *Manager) InitFeatures() error {
	m.mu.Lock()
	if m.initialized {
		m.mu.Unlock()
		return fmt.Errorf("features already initialized")
	}
	// Check dependencies before determining initialization order
	if err := m.checkDependencies(); err != nil {
		m.mu.Unlock()
		return err
	}
	initOrder, err := getInitOrder(m.features)
	if err != nil {
		log.Errorf(context.Background(), "failed to determine initialization order: %v", err)
		m.mu.Unlock()
		return err
	}
	m.mu.Unlock() // Unlock after dependencies check and order determination

	// Pre-initialization
	for _, name := range initOrder {
		feature := m.features[name]
		if err := feature.Instance.PreInit(); err != nil {
			log.Errorf(context.Background(), "failed pre-initialization of feature %s: %v", name, err)
			continue // Skip current feature and move to the next one
		}
	}

	// Initialization
	for _, name := range initOrder {
		feature := m.features[name]
		if err := feature.Instance.Init(m.conf, m); err != nil {
			log.Errorf(context.Background(), "failed to initialize feature %s: %v", name, err)
			continue // Skip current feature and move to the next one
		}
	}

	// Post-initialization
	for _, name := range initOrder {
		feature := m.features[name]
		if err := feature.Instance.PostInit(); err != nil {
			log.Errorf(context.Background(), "failed post-initialization of feature %s: %v", name, err)
			continue // Skip current feature and move to the next one
		}
	}

	// Ensure all services are initialized
	for _, feature := range m.features {
		_ = feature.Instance.GetServices()
	}

	// Lock again to safely update the initialized flag
	m.mu.Lock()
	m.initialized = true
	m.mu.Unlock()

	log.Infof(context.Background(), "All features initialized successfully")
	return nil
}

// GetFeature returns a specific feature
func (m *Manager) GetFeature(name string) (Interface, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	feature, exists := m.features[name]
	if !exists {
		return nil, fmt.Errorf("feature %s not found", name)
	}

	return feature.Instance, nil
}

// GetFeatures returns the loaded features
func (m *Manager) GetFeatures() map[string]*Wrapper {
	m.mu.RLock()
	defer m.mu.RUnlock()

	features := make(map[string]*Wrapper)
	for name, feature := range m.features {
		features[name] = feature
	}
	return features
}

// Cleanup cleans up all loaded features
func (m *Manager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, feature := range m.features {
		if err := feature.Instance.PreCleanup(); err != nil {
			log.Errorf(context.Background(), "failed pre-cleanup of feature %s: %v", feature.Metadata.Name, err)
		}
		if err := feature.Instance.Cleanup(); err != nil {
			log.Errorf(context.Background(), "failed to cleanup feature %s: %v", feature.Metadata.Name, err)
		}
	}

	m.features = make(map[string]*Wrapper)
	m.initialized = false
}

// RegisterRoutes registers all feature routes with the provided router
func (m *Manager) RegisterRoutes(router *gin.Engine) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, f := range m.features {
		f.Instance.RegisterRoutes(router)
		log.Infof(context.Background(), "Registered routes for %s", f.Metadata.Name)
	}
}

// getInitOrder returns the initialization order based on dependencies
func getInitOrder(features map[string]*Wrapper) ([]string, error) {
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	for name, feature := range features {
		for _, dep := range feature.Metadata.Dependencies {
			graph[dep] = append(graph[dep], name)
			inDegree[name]++
		}
	}

	var order []string
	var queue []string
	for name := range features {
		if inDegree[name] == 0 {
			queue = append(queue, name)
		}
	}

	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		order = append(order, name)

		for _, dep := range graph[name] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	if len(order) != len(features) {
		return nil, fmt.Errorf("cyclic dependency detected")
	}

	return order, nil
}

// GetHandler returns a specific handler from a feature
func (m *Manager) GetHandler(featureName string, handlerName string) (Handler, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	feature, exists := m.features[featureName]
	if !exists {
		return nil, fmt.Errorf("feature %s not found", featureName)
	}

	handlers := feature.Instance.GetHandlers()
	handler, exists := handlers[handlerName]
	if !exists {
		return nil, fmt.Errorf("handler %s not found in feature %s", handlerName, featureName)
	}

	return handler, nil
}

// GetHandlers returns all registered feature handlers
func (m *Manager) GetHandlers() map[string]map[string]Handler {
	m.mu.RLock()
	defer m.mu.RUnlock()
	handlers := make(map[string]map[string]Handler)
	for name, feature := range m.features {
		handlers[name] = feature.Instance.GetHandlers()
	}
	return handlers
}

// GetService returns a specific service from a feature
func (m *Manager) GetService(featureName string, serviceName string) (Service, error) {
	m.mu.RLock()
	feature, exists := m.features[featureName]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("feature %s not found", featureName)
	}

	services := feature.Instance.GetServices()
	service, exists := services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found in feature %s", serviceName, featureName)
	}

	return service, nil
}

// GetServices returns all registered feature services
func (m *Manager) GetServices() map[string]map[string]Service {
	m.mu.RLock()
	defer m.mu.RUnlock()

	services := make(map[string]map[string]Service)
	for name, feature := range m.features {
		services[name] = feature.Instance.GetServices()
	}
	return services
}

// GetMetadata returns the metadata of all registered features
func (m *Manager) GetMetadata() map[string]Metadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metadata := make(map[string]Metadata)
	for name, feature := range m.features {
		metadata[name] = feature.Metadata
	}
	return metadata
}

// GetStatus returns the status of all registered features
func (m *Manager) GetStatus() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]string)
	for name, feature := range m.features {
		status[name] = feature.Instance.Status()
	}
	return status
}

// ManageRoutes manages routes for all features / plugins
func (m *Manager) ManageRoutes(e *gin.Engine) {
	e.GET("/features", func(c *gin.Context) {
		resp.Success(c.Writer, &resp.Exception{Data: m.features})
	})

	e.POST("/features/load", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			resp.Fail(c.Writer, &resp.Exception{Code: ecode.RequestErr, Message: "name is required"})
			return
		}
		fc := m.conf.Feature
		fp := filepath.Join(fc.Path, name+".so")
		if err := m.loadPlugin(fp); err != nil {
			resp.Fail(c.Writer, &resp.Exception{Code: ecode.ServerErr, Message: fmt.Sprintf("Failed to load feature %s: %v", name, err)})
			return
		}
		resp.Success(c.Writer, &resp.Exception{Message: fmt.Sprintf("%s loaded successfully", name)})
	})

	e.POST("/features/unload", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			resp.Fail(c.Writer, &resp.Exception{Code: ecode.RequestErr, Message: "name is required"})
			return
		}
		if err := m.UnloadPlugin(name); err != nil {
			resp.Fail(c.Writer, &resp.Exception{Code: ecode.ServerErr, Message: fmt.Sprintf("Failed to unload feature %s: %v", name, err)})
			return
		}
		resp.Success(c.Writer, &resp.Exception{Message: fmt.Sprintf("%s unloaded successfully", name)})
	})

	e.POST("/features/reload", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			resp.Fail(c.Writer, &resp.Exception{Code: ecode.RequestErr, Message: "name is required"})
			return
		}
		if err := m.ReloadPlugin(name); err != nil {
			resp.Fail(c.Writer, &resp.Exception{Code: ecode.ServerErr, Message: fmt.Sprintf("Failed to reload feature %s: %v", name, err)})
			return
		}
		resp.Success(c.Writer, &resp.Exception{Message: fmt.Sprintf("%s reloaded successfully", name)})
	})
}

// ReloadPlugin reloads a single feature / plugin
func (m *Manager) ReloadPlugin(name string) error {
	fc := m.conf.Feature
	fd := fc.Path
	fp := filepath.Join(fd, name+".so")

	if err := m.UnloadPlugin(name); err != nil {
		return err
	}

	return m.loadPlugin(fp)
}

// ReloadPlugins reloads all features / plugins
func (m *Manager) ReloadPlugins() error {
	fc := m.conf.Feature
	fd := fc.Path
	pds, err := filepath.Glob(filepath.Join(fd, "*.so"))
	if err != nil {
		log.Errorf(context.Background(), "failed to list plugin files: %v", err)
		return err
	}
	for _, fp := range pds {
		if err := m.ReloadPlugin(strings.TrimSuffix(filepath.Base(fp), ".so")); err != nil {
			return err
		}
	}
	return nil
}

// PublishEvent publishes an event to all features
func (m *Manager) PublishEvent(eventName string, data interface{}) {
	m.eventBus.Publish(eventName, data)
}

// SubscribeEvent allows a feature to subscribe to an event
func (m *Manager) SubscribeEvent(eventName string, handler func(interface{})) {
	m.eventBus.Subscribe(eventName, handler)
}

// checkDependencies checks if all dependencies are loaded
func (m *Manager) checkDependencies() error {
	for name, feature := range m.features {
		for _, dep := range feature.Instance.Dependencies() {
			if _, ok := m.features[dep]; !ok {
				return fmt.Errorf("feature '%s' depends on '%s', which is not available", name, dep)
			}
		}
	}
	return nil
}
