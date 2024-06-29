package plugin

import (
	"fmt"
	"log"
	"plugin"
	"sync"

	"ncobase/common/config"

	"github.com/gin-gonic/gin"
)

type Plugin interface {
	Name() string
	Init(conf *config.Config) error
	RegisterRoutes(router *gin.Engine)
	Cleanup() error
}

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

	symPlugin, err := p.Lookup("Plugin")
	if err != nil {
		return fmt.Errorf("plugin %s does not export 'Plugin' symbol: %v", path, err)
	}

	sp, ok := symPlugin.(Plugin)
	if !ok {
		return fmt.Errorf("plugin %s does not implement Plugin interface, got %T", path, err)
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
