package core

import (
	"fmt"
	"strings"
	"sync"
)

// EngineComponent represents a workflow engine component
type EngineComponent interface {
	// Basic lifecycle

	Start() error
	Stop() error
	Status() ComponentStatus
	Name() string

	// Health check

	IsHealthy() bool
	GetMetrics() map[string]any

	// Component specific

	Dependencies() []string
	GetConfig() any
}

// ComponentStatus represents engine component status
type ComponentStatus string

const (
	ComponentStatusReady   ComponentStatus = "ready"
	ComponentStatusRunning ComponentStatus = "running"
	ComponentStatusStopped ComponentStatus = "stopped"
	ComponentStatusFailure ComponentStatus = "failure"
)

// ComponentRegistry manages engine components
type ComponentRegistry struct {
	components map[string]EngineComponent
	mu         sync.RWMutex
}

// NewComponentRegistry creates a new component registry
func NewComponentRegistry() *ComponentRegistry {
	return &ComponentRegistry{
		components: make(map[string]EngineComponent),
	}
}

// Register registers a component
func (r *ComponentRegistry) Register(name string, component EngineComponent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if component == nil {
		return fmt.Errorf("component cannot be nil")
	}

	name = strings.ToLower(name)
	if _, exists := r.components[name]; exists {
		return fmt.Errorf("component %s already registered", name)
	}

	r.components[name] = component
	return nil
}

// Get gets a component by name
func (r *ComponentRegistry) Get(name string) (EngineComponent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	name = strings.ToLower(name)
	component, exists := r.components[name]
	if !exists {
		return nil, fmt.Errorf("component %s not found", name)
	}

	return component, nil
}

// GetAll returns all registered components
func (r *ComponentRegistry) GetAll() map[string]EngineComponent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	components := make(map[string]EngineComponent)
	for name, component := range r.components {
		components[name] = component
	}
	return components
}

// Remove removes a component
func (r *ComponentRegistry) Remove(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name = strings.ToLower(name)
	if _, exists := r.components[name]; !exists {
		return fmt.Errorf("component %s not found", name)
	}

	delete(r.components, name)
	return nil
}

// Start starts all registered components
func (r *ComponentRegistry) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Get initialization order
	order, err := r.getInitOrder()
	if err != nil {
		return fmt.Errorf("failed to determine init order: %w", err)
	}

	// Start components in order
	for _, name := range order {
		component := r.components[name]
		if err := component.Start(); err != nil {
			return fmt.Errorf("failed to start component %s: %w", name, err)
		}
	}

	return nil
}

// Stop stops all registered components
func (r *ComponentRegistry) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Stop in reverse order
	order, err := r.getInitOrder()
	if err != nil {
		return fmt.Errorf("failed to determine shutdown order: %w", err)
	}

	// Reverse the order
	for i := len(order)/2 - 1; i >= 0; i-- {
		opp := len(order) - 1 - i
		order[i], order[opp] = order[opp], order[i]
	}

	// Stop components
	for _, name := range order {
		component := r.components[name]
		if err := component.Stop(); err != nil {
			return fmt.Errorf("failed to stop component %s: %w", name, err)
		}
	}

	return nil
}

// getInitOrder returns component initialization order based on dependencies
func (r *ComponentRegistry) getInitOrder() ([]string, error) {
	// Build dependency graph
	graph := make(map[string][]string)
	for name, component := range r.components {
		graph[name] = component.Dependencies()
	}

	// Topological sort
	visited := make(map[string]bool)
	temp := make(map[string]bool)
	var order []string

	var visit func(name string) error
	visit = func(name string) error {
		if temp[name] {
			return fmt.Errorf("cyclic dependency detected")
		}
		if visited[name] {
			return nil
		}
		temp[name] = true
		for _, dep := range graph[name] {
			if err := visit(dep); err != nil {
				return err
			}
		}
		temp[name] = false
		visited[name] = true
		order = append(order, name)
		return nil
	}

	// Visit all nodes
	for name := range graph {
		if !visited[name] {
			if err := visit(name); err != nil {
				return nil, err
			}
		}
	}

	return order, nil
}

// GetStatus returns status of all components
func (r *ComponentRegistry) GetStatus() map[string]ComponentStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	status := make(map[string]ComponentStatus)
	for name, component := range r.components {
		status[name] = component.Status()
	}
	return status
}

// IsHealthy checks if all components are healthy
func (r *ComponentRegistry) IsHealthy() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, component := range r.components {
		if !component.IsHealthy() {
			return false
		}
	}
	return true
}

// GetMetrics returns metrics from all components
func (r *ComponentRegistry) GetMetrics() map[string]map[string]any {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metrics := make(map[string]map[string]any)
	for name, component := range r.components {
		metrics[name] = component.GetMetrics()
	}
	return metrics
}
