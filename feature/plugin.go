// feature/plugin.go
package feature

import (
	"context"
	"fmt"
	"ncobase/common/config"
	"ncobase/common/log"
	"path/filepath"
	"plugin"
	"sync"
)

// PluginRegistry manages the loaded plugins
type PluginRegistry struct {
	mu      sync.RWMutex
	plugins map[string]*Wrapper
}

var registry = &PluginRegistry{
	plugins: make(map[string]*Wrapper),
}

var plugins []*Wrapper

// RegisterPlugin registers a new plugin
func RegisterPlugin(c Interface, metadata Metadata) {
	plugins = append(plugins, &Wrapper{
		Metadata: metadata,
		Instance: c,
	})
}

// GetRegisteredPlugins returns the registered plugins
func GetRegisteredPlugins() []*Wrapper {
	return plugins
}

// LoadPlugin loads a single plugin
func LoadPlugin(path string, conf *config.Config) error {
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin %s: %v", path, err)
	}

	symPlugin, err := p.Lookup("Instance")
	if err != nil {
		return fmt.Errorf("plugin %s does not export 'Instance' symbol: %v", path, err)
	}

	sc, ok := symPlugin.(Interface)
	if !ok {
		return fmt.Errorf("plugin %s does not implement interface, got %T", path, sc)
	}

	if err := sc.Init(conf); err != nil {
		return fmt.Errorf("failed to initialize plugin %s: %v", path, err)
	}

	metadata := sc.GetMetadata()

	registry.mu.Lock()
	defer registry.mu.Unlock()

	name := sc.Name()
	if _, exists := registry.plugins[name]; exists {
		log.Printf(context.Background(), "Warning: Plugin %s is being overwritten", name)
	}
	registry.plugins[name] = &Wrapper{
		Metadata: metadata,
		Instance: sc,
	}
	log.Printf(context.Background(), "Plugin %s loaded and initialized successfully", name)

	return nil
}

// UnloadPlugin unloads a single plugin
func UnloadPlugin(name string) error {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	c, exists := registry.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	if err := c.Instance.Cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup plugin %s: %v", name, err)
	}

	delete(registry.plugins, name)
	log.Printf(context.Background(), "plugin %s unloaded successfully", name)
	return nil
}

// GetPlugin returns a single plugin
func GetPlugin(name string) *Wrapper {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	return registry.plugins[name]
}

// GetPlugins returns all loaded plugins
func GetPlugins() map[string]*Wrapper {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	plugins := make(map[string]*Wrapper)
	for name, c := range registry.plugins {
		plugins[name] = c
	}
	return plugins
}

// ReloadPlugin reloads a single feature
func (m *Manager) ReloadPlugin(name string) error {
	fc := m.conf.Feature
	fd := fc.Path
	fp := filepath.Join(fd, name+".so")

	if err := m.UnloadPlugin(name); err != nil {
		return err
	}

	return m.loadPlugin(fp)
}

// ReloadPlugins reloads all features
func (m *Manager) ReloadPlugins() error {
	fc := m.conf.Feature
	fd := fc.Path
	pds, err := filepath.Glob(filepath.Join(fd, "*.so"))
	if err != nil {
		log.Errorf(context.Background(), "failed to list plugin files: %v", err)
		return err
	}
	for _, fp := range pds {
		if err := m.ReloadPlugin(filepath.Base(fp)); err != nil {
			return err
		}
	}
	return nil
}
