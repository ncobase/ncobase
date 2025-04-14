package config

import (
	"fmt"
	"ncobase/core/workflow/engine/batcher"
	"ncobase/core/workflow/engine/coordinator"
	"ncobase/core/workflow/engine/metrics"
	"ncobase/core/workflow/engine/scheduler"

	"github.com/ncobase/ncore/concurrency/worker"
	"github.com/ncobase/ncore/messaging/queue"
	"github.com/ncobase/ncore/validation/expression"
)

// Config represents the complete engine configuration
type Config struct {
	// Basic engine settings
	Engine *EngineConfig `json:"engine,omitempty"`

	// Executor configs
	Executors *ExecutorConfig `json:"executors,omitempty"`

	// Handler configs
	Handlers *HandlerConfig `json:"handlers,omitempty"`

	// Component configs
	Components *ComponentConfig `json:"components,omitempty"`
}

// EngineConfig represents basic engine configuration
type EngineConfig struct {
	// Maximum concurrent executions
	MaxConcurrency int32 `json:"max_concurrency"`

	// Global timeout settings (in seconds)
	DefaultTimeout int `json:"default_timeout"`

	// Global retry settings
	MaxRetries    int `json:"max_retries"`
	RetryInterval int `json:"retry_interval"` // seconds

	// Default worker settings
	DefaultWorkers int `json:"default_workers"`
	DefaultQueue   int `json:"default_queue_size"`

	// Resource limits
	MaxHandlers   int `json:"max_handlers"`
	MaxExecutors  int `json:"max_executors"`
	MaxWorkers    int `json:"max_workers"`
	MaxMemory     int `json:"max_memory"`
	MaxGoroutines int `json:"max_goroutines"`
}

// ComponentConfig represents shared component configurations
type ComponentConfig struct {
	// Queue component config
	Queue *queue.Config `json:"queue,omitempty"`

	// Scheduler component config
	Scheduler *scheduler.Config `json:"scheduler,omitempty"`

	// Worker component config
	Worker *worker.Config `json:"worker,omitempty"`

	// Metrics component config
	Metrics *metrics.Config `json:"metrics,omitempty"`

	// Batcher component config
	Batcher *batcher.Config `json:"batcher,omitempty"`

	// Coordinator component config
	Coordinator *coordinator.Config `json:"coordinator,omitempty"`

	// Expression component config
	Expression *expression.Config `json:"expression,omitempty"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Engine: &EngineConfig{
			MaxConcurrency: 100,
			DefaultTimeout: 3600, // 1 hour
			MaxRetries:     3,
			RetryInterval:  60, // 1 minute
			DefaultWorkers: 10,
			DefaultQueue:   1000,
			MaxMemory:      1024 * 1024 * 1024, // 1GB
			MaxGoroutines:  10000,
		},
		Executors: &ExecutorConfig{
			Base:    DefaultBaseExecutorConfig(),
			Process: DefaultProcessExecutorConfig(),
			Node:    DefaultNodeExecutorConfig(),
			Task:    DefaultTaskExecutorConfig(),
			Service: DefaultServiceExecutorConfig(),
			Retry:   DefaultRetryExecutorConfig(),
		},
		Handlers: &HandlerConfig{
			Base:         DefaultBaseHandlerConfig(),
			Approval:     DefaultApprovalHandlerConfig(),
			Exclusive:    DefaultExclusiveHandlerConfig(),
			Notification: DefaultNotificationHandlerConfig(),
			Parallel:     DefaultParallelHandlerConfig(),
			SubProcess:   DefaultSubProcessHandlerConfig(),
			Service:      DefaultServiceHandlerConfig(),
			Timer:        DefaultTimerHandlerConfig(),
			Script:       DefaultScriptHandlerConfig(),
		},
		Components: &ComponentConfig{
			Queue:       queue.DefaultConfig(),
			Scheduler:   scheduler.DefaultConfig(),
			Worker:      worker.DefaultConfig(),
			Metrics:     metrics.DefaultConfig(),
			Batcher:     batcher.DefaultConfig(),
			Coordinator: coordinator.DefaultConfig(),
			Expression:  expression.DefaultConfig(),
		},
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Engine == nil {
		return fmt.Errorf("engine configuration is required")
	}

	if err := c.validateEngineConfig(); err != nil {
		return err
	}

	if err := c.validateExecutorConfig(); err != nil {
		return err
	}

	if err := c.validateComponentConfig(); err != nil {
		return err
	}

	return nil
}

// Configuration validation helpers
func (c *Config) validateEngineConfig() error {
	if c.Engine.MaxConcurrency <= 0 {
		return fmt.Errorf("max_concurrency must be greater than 0")
	}

	if c.Engine.DefaultTimeout <= 0 {
		return fmt.Errorf("default_timeout must be greater than 0")
	}

	if c.Engine.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative")
	}

	return nil
}

func (c *Config) validateExecutorConfig() error {
	if c.Executors == nil {
		return fmt.Errorf("executor configuration is required")
	}

	// Validate process executor config
	if c.Executors.Process != nil {
		if c.Executors.Process.Workers <= 0 {
			return fmt.Errorf("process workers must be greater than 0")
		}
	}

	// Validate other executors...
	return nil
}

func (c *Config) validateComponentConfig() error {
	if c.Components == nil {
		return fmt.Errorf("component configuration is required")
	}

	// Validate individual components
	if c.Components.Queue != nil {
		if err := c.Components.Queue.Validate(); err != nil {
			return fmt.Errorf("invalid queue config: %w", err)
		}
	}

	// Validate other components...
	return nil
}
