package plugin

import (
	"fmt"
	"log"
	"plugin"
	"sync"

	"ncobase/common/config"

	"github.com/gin-gonic/gin"
)

// Plugin interface defines the structure for a plugin
type Plugin interface {
	Name() string
	Init(conf *config.Config) error
	RegisterRoutes(router *gin.Engine)
	Cleanup() error
}

// Registry manages the loaded plugins
type Registry struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
}

var registry = &Registry{
	plugins: make(map[string]Plugin),
}

// devPlugins is a slice of plugins that are loaded in development mode
var devPlugins []Plugin

// RegisterPlugin registers a plugin in development mode
func RegisterPlugin(p Plugin) {
	devPlugins = append(devPlugins, p)
}

// GetRegisteredPlugins returns a slice of plugins that are registered in development mode
func GetRegisteredPlugins() []Plugin {
	return devPlugins
}

// LoadPlugin loads a plugin and initializes it
func LoadPlugin(path string, conf *config.Config) error {
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin %s: %v", path, err)
	}

	symPlugin, err := p.Lookup("PluginInstance")
	if err != nil {
		return fmt.Errorf("plugin %s does not export 'PluginInstance' symbol: %v", path, err)
	}

	sp, ok := symPlugin.(Plugin)
	if !ok {
		return fmt.Errorf("plugin %s does not implement Plugin interface, got %T", path, sp)
	}

	if err := sp.Init(conf); err != nil {
		return fmt.Errorf("failed to initialize plugin %s: %v", path, err)
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	name := sp.Name()
	if _, exists := registry.plugins[name]; exists {
		log.Printf("Warning: Plugin %s is being overwritten", name)
	}
	registry.plugins[name] = sp
	log.Printf("Plugin %s loaded and initialized successfully", name)

	return nil
}

// UnloadPlugin unloads a plugin and cleans it up
func UnloadPlugin(pluginName string) error {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	p, exists := registry.plugins[pluginName]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginName)
	}

	if err := p.Cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup plugin %s: %v", pluginName, err)
	}

	delete(registry.plugins, pluginName)
	log.Printf("Plugin %s unloaded successfully", pluginName)
	return nil
}

// GetPlugin returns a specific plugin by name
func GetPlugin(name string) Plugin {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	return registry.plugins[name]
}

// GetPlugins returns a map of all plugins
func GetPlugins() map[string]Plugin {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	plugins := make(map[string]Plugin)
	for name, p := range registry.plugins {
		plugins[name] = p
	}
	return plugins
}
