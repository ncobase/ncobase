package server

import (
	"context"
	"ncobase/common/config"
	"ncobase/common/log"
	"ncobase/plugin"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	_ "ncobase/plugin/content/cmd"
)

// PluginManager represents a plugin manager
type PluginManager struct {
	plugins map[string]plugin.Plugin
	conf    *config.Config
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(conf *config.Config) *PluginManager {
	return &PluginManager{
		plugins: make(map[string]plugin.Plugin),
		conf:    conf,
	}
}

// LoadPlugins loads plugins
func (pm *PluginManager) LoadPlugins() error {
	if pm.conf.RunMode == "c2hlbgo" {
		return pm.loadPluginsInDevMode()
	}
	return pm.loadPluginsInProdMode()
}

// loadPluginsInProdMode loads plugins in production mode
func (pm *PluginManager) loadPluginsInProdMode() error {
	pluginConfig := pm.conf.Plugin
	pluginDir := pluginConfig.Path

	// Get all .so files in the plugin directory
	pluginPaths, err := filepath.Glob(filepath.Join(pluginDir, "*.so"))
	if err != nil {
		return err
	}

	for _, path := range pluginPaths {
		pluginName := strings.TrimSuffix(filepath.Base(path), ".so")

		// Check if the plugin should be included or excluded
		if !pm.shouldLoadPlugin(pluginName) {
			log.Infof(context.Background(), "Skipping plugin %s based on configuration", pluginName)
			continue
		}

		if err := plugin.LoadPlugin(path, pm.conf); err != nil {
			log.Errorf(context.Background(), "Failed to load plugin %s: %v", path, err)
			continue
		}
	}

	pm.plugins = plugin.GetPlugins()
	return nil
}

// LoadPluginsInDevMode loads plugins in development mode
func (pm *PluginManager) loadPluginsInDevMode() error {
	devPlugins := plugin.GetRegisteredPlugins()

	for _, p := range devPlugins {
		if err := p.Init(pm.conf); err != nil {
			log.Errorf(context.Background(), "Failed to initialize plugin %s: %v", p.Name(), err)
			continue
		}
		pm.plugins[p.Name()] = p
		log.Infof(context.Background(), "Plugin %s loaded and initialized successfully", p.Name())
	}

	return nil
}

// shouldLoadPlugin returns true if the plugin should be loaded
func (pm *PluginManager) shouldLoadPlugin(pluginName string) bool {
	pluginConfig := pm.conf.Plugin

	// If includes is not empty, only load plugins in the includes list
	if len(pluginConfig.Includes) > 0 {
		for _, include := range pluginConfig.Includes {
			if include == pluginName {
				return true
			}
		}
		return false
	}

	// If excludes is not empty, do not load plugins in the excludes list
	if len(pluginConfig.Excludes) > 0 {
		for _, exclude := range pluginConfig.Excludes {
			if exclude == pluginName {
				return false
			}
		}
	}

	// If neither includes nor excludes are specified, load all plugins
	return true
}

// RegisterPluginRoutes registers routes for all plugins
func (pm *PluginManager) RegisterPluginRoutes(e *gin.Engine) {
	for name, p := range pm.plugins {
		p.RegisterRoutes(e)
		log.Infof(context.Background(), "Routes for plugin %s registered", name)
	}
}

// CleanupPlugins cleans up all plugins
func (pm *PluginManager) CleanupPlugins() {
	for name, p := range pm.plugins {
		if err := p.Cleanup(); err != nil {
			log.Errorf(context.Background(), "Error cleaning up plugin %s: %v", name, err)
		}
	}
}
