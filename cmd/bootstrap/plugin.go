package bootstrap

import (
	"context"
	"fmt"
	"ncobase/common/config"
	"ncobase/common/ecode"
	"ncobase/common/log"
	"ncobase/common/resp"
	"ncobase/helper"
	"ncobase/plugin"
	"path/filepath"
	"strings"

	_ "ncobase/plugin/asset"
	_ "ncobase/plugin/content"

	"github.com/gin-gonic/gin"
)

// PluginManager represents a plugin manager
type PluginManager struct {
	plugins map[string]*plugin.Wrapper
	conf    *config.Config
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(conf *config.Config) *PluginManager {
	return &PluginManager{
		plugins: make(map[string]*plugin.Wrapper),
		conf:    conf,
	}
}

// LoadPlugins loads all plugins based on the current configuration
func (pm *PluginManager) LoadPlugins() error {
	if helper.IsPluginDevMode(pm.conf) {
		return pm.loadPluginsInDevMode()
	}
	return pm.loadPluginsInProdMode()
}

// loadPluginsInProdMode loads plugins in production mode
func (pm *PluginManager) loadPluginsInProdMode() error {
	pluginConfig := pm.conf.Plugin
	pluginDir := pluginConfig.Path

	pluginPaths, err := filepath.Glob(filepath.Join(pluginDir, "*.so"))
	if err != nil {
		log.Errorf(context.Background(), "failed to list plugin files: %v", err)
		return err
	}

	for _, path := range pluginPaths {
		pluginName := strings.TrimSuffix(filepath.Base(path), ".so")
		if !pm.shouldLoadPlugin(pluginName) {
			log.Infof(context.Background(), "Skipping plugin %s based on configuration", pluginName)
			continue
		}
		if err := pm.loadPlugin(path); err != nil {
			log.Errorf(context.Background(), "Failed to load plugin %s: %v", pluginName, err)
		}
	}

	return nil
}

// loadPluginsInDevMode loads plugins in development mode
func (pm *PluginManager) loadPluginsInDevMode() error {
	devPlugins := plugin.GetRegisteredPlugins()

	for _, p := range devPlugins {
		if err := p.Instance.Init(pm.conf); err != nil {
			log.Errorf(context.Background(), "Failed to initialize plugin %s: %v", p.Metadata.Name, err)
			continue
		}
		pm.plugins[p.Metadata.Name] = p
		log.Infof(context.Background(), "Plugin %s loaded and initialized successfully", p.Metadata.Name)
	}

	return nil
}

// shouldLoadPlugin returns true if the plugin should be loaded
func (pm *PluginManager) shouldLoadPlugin(pluginName string) bool {
	pluginConfig := pm.conf.Plugin

	if len(pluginConfig.Includes) > 0 {
		for _, include := range pluginConfig.Includes {
			if include == pluginName {
				return true
			}
		}
		return false
	}

	if len(pluginConfig.Excludes) > 0 {
		for _, exclude := range pluginConfig.Excludes {
			if exclude == pluginName {
				return false
			}
		}
	}

	return true
}

// loadPlugin loads a single plugin
func (pm *PluginManager) loadPlugin(path string) error {
	pluginName := strings.TrimSuffix(filepath.Base(path), ".so")
	if _, exists := pm.plugins[pluginName]; exists {
		return nil // Plugin already loaded
	}

	if err := plugin.LoadPlugin(path, pm.conf); err != nil {
		log.Errorf(context.Background(), "failed to load plugin %s: %v", pluginName, err)
		return err
	}

	loadedPlugin := plugin.GetPlugin(pluginName)
	if loadedPlugin != nil {
		pm.plugins[pluginName] = loadedPlugin
		log.Infof(context.Background(), "Plugin %s loaded successfully", pluginName)
	}

	return nil
}

// UnloadPlugin unloads a single plugin
func (pm *PluginManager) UnloadPlugin(pluginName string) error {
	p, exists := pm.plugins[pluginName]
	if !exists {
		return nil // Plugin not loaded
	}

	if err := p.Instance.Cleanup(); err != nil {
		return err
	}

	delete(pm.plugins, pluginName)
	log.Infof(context.Background(), "Plugin %s unloaded successfully", pluginName)
	return nil
}

// ReloadPlugin reloads a single plugin
func (pm *PluginManager) ReloadPlugin(pluginName string) error {
	pluginConfig := pm.conf.Plugin
	pluginDir := pluginConfig.Path
	pluginPath := filepath.Join(pluginDir, pluginName+".so")

	if err := pm.UnloadPlugin(pluginName); err != nil {
		return err
	}

	return pm.loadPlugin(pluginPath)
}

// RegisterPluginRoutes registers routes for all plugins
func (pm *PluginManager) RegisterPluginRoutes(e *gin.Engine) {
	for name, p := range pm.plugins {
		p.Instance.RegisterRoutes(e)
		log.Infof(context.Background(), "Routes for plugin %s registered", name)
	}
}

// CleanupPlugins cleans up all plugins
func (pm *PluginManager) CleanupPlugins() {
	for name, p := range pm.plugins {
		if err := p.Instance.Cleanup(); err != nil {
			log.Errorf(context.Background(), "Error cleaning up plugin %s: %v", name, err)
		}
	}
}

// AddPluginRoutes adds new handler functions for dynamic plugin management
func (pm *PluginManager) AddPluginRoutes(e *gin.Engine) {
	e.GET("/plugins", func(c *gin.Context) {
		resp.Success(c.Writer, &resp.Exception{Data: pm.plugins})
	})

	e.POST("/plugins/load", func(c *gin.Context) {
		pluginName := c.Query("name")
		if pluginName == "" {
			resp.Fail(c.Writer, &resp.Exception{Code: ecode.RequestErr, Message: "Plugin name is required"})
			return
		}
		pluginConfig := pm.conf.Plugin
		pluginPath := filepath.Join(pluginConfig.Path, pluginName+".so")
		if err := pm.loadPlugin(pluginPath); err != nil {
			resp.Fail(c.Writer, &resp.Exception{Code: ecode.ServerErr, Message: fmt.Sprintf("Failed to load plugin %s: %v", pluginName, err)})
			return
		}
		resp.Success(c.Writer, &resp.Exception{Message: fmt.Sprintf("Plugin %s loaded successfully", pluginName)})
	})

	e.POST("/plugins/unload", func(c *gin.Context) {
		pluginName := c.Query("name")
		if pluginName == "" {
			resp.Fail(c.Writer, &resp.Exception{Code: ecode.RequestErr, Message: "Plugin name is required"})
			return
		}
		if err := pm.UnloadPlugin(pluginName); err != nil {
			resp.Fail(c.Writer, &resp.Exception{Code: ecode.ServerErr, Message: fmt.Sprintf("Failed to unload plugin %s: %v", pluginName, err)})
			return
		}
		resp.Success(c.Writer, &resp.Exception{Message: fmt.Sprintf("Plugin %s unloaded successfully", pluginName)})
	})

	e.POST("/plugins/reload", func(c *gin.Context) {
		pluginName := c.Query("name")
		if pluginName == "" {
			resp.Fail(c.Writer, &resp.Exception{Code: ecode.RequestErr, Message: "Plugin name is required"})
			return
		}
		if err := pm.ReloadPlugin(pluginName); err != nil {
			resp.Fail(c.Writer, &resp.Exception{Code: ecode.ServerErr, Message: fmt.Sprintf("Failed to reload plugin %s: %v", pluginName, err)})
			return
		}
		resp.Success(c.Writer, &resp.Exception{Message: fmt.Sprintf("Plugin %s reloaded successfully", pluginName)})
	})
}
