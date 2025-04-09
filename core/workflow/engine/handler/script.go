package handler

import (
	"context"
	"fmt"
	"math"
	"ncobase/core/workflow/engine/config"
	"ncobase/core/workflow/engine/types"
	"ncobase/core/workflow/service"
	"ncobase/core/workflow/structs"
	nec "ncore/ext/core"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
)

// ScriptHandler implements script node handling
type ScriptHandler struct {
	*BaseHandler

	// Configuration
	config *config.ScriptHandlerConfig

	// Script engine
	se *ScriptEngine
}

// NewScriptHandler creates a new script handler
func NewScriptHandler(svc *service.Service, em nec.ManagerInterface, cfg *config.Config) *ScriptHandler {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	handler := &ScriptHandler{
		BaseHandler: NewBaseHandler("script", "Script Handler", svc, em, cfg.Handlers.Base),
		config:      cfg.Handlers.Script,
	}

	return handler
}

// Type returns handler type
func (h *ScriptHandler) Type() types.HandlerType { return h.handlerType }

// Name returns handler name
func (h *ScriptHandler) Name() string { return h.name }

// Priority returns handler priority
func (h *ScriptHandler) Priority() int { return h.priority }

// Start starts the script handler
func (h *ScriptHandler) Start() error {
	if err := h.BaseHandler.Start(); err != nil {
		return err
	}

	// Initialize script engine
	h.se = NewScriptEngine(h.config.Sandbox)

	return nil
}

// Stop stops the script handler
func (h *ScriptHandler) Stop() error {
	if err := h.BaseHandler.Stop(); err != nil {
		return err
	}

	return nil
}

// executeInternal executes the script node
func (h *ScriptHandler) executeInternal(ctx context.Context, node *structs.ReadNode) error {
	// Get script config
	c, err := h.parseScriptConfig(node)
	if err != nil {
		return err
	}

	// Prepare parameters
	params := make(map[string]any)
	if c.InputVars != nil {
		params = c.InputVars
	}

	// Add process variables
	process, err := h.services.Process.Get(ctx, &structs.FindProcessParams{
		ProcessKey: node.ProcessID,
	})
	if err != nil {
		return fmt.Errorf("failed to get process: %w", err)
	}

	if process.Variables != nil {
		for k, v := range process.Variables {
			params[k] = v
		}
	}

	// Execute script
	result, err := h.se.Execute(ctx, c.Script, params)
	if err != nil {
		if c.FailureMode == "continue" {
			// Log error but continue
			h.logger.Errorf(ctx, "script execution failed: %v", err)
		} else {
			return fmt.Errorf("script execution failed: %w", err)
		}
	}

	// Handle output
	if len(c.OutputVars) > 0 && result != nil {
		if resultMap, ok := result.(map[string]any); ok {
			// Update process variables with output
			variables := process.Variables
			if variables == nil {
				variables = make(map[string]any)
			}

			for _, key := range c.OutputVars {
				if value, exists := resultMap[key]; exists {
					variables[key] = value
				}
			}

			// Update process
			_, err = h.services.Process.Update(ctx, &structs.UpdateProcessBody{
				ID: process.ID,
				ProcessBody: structs.ProcessBody{
					Variables: variables,
				},
			})
			if err != nil {
				return fmt.Errorf("failed to update process variables: %w", err)
			}
		}
	}

	return nil
}

// Script engine methods

// ScriptEngine represents the script execution engine
type ScriptEngine struct {
	// VM Pool for reuse
	pool sync.Pool

	// Sandbox configuration
	sandbox *config.Sandbox

	// Metrics and monitoring
	metrics *ScriptEngineMetrics
}

// ScriptEngineMetrics tracks script execution metrics
type ScriptEngineMetrics struct {
	ExecutionCount   int64
	SuccessCount     int64
	FailureCount     int64
	TimeoutCount     int64
	ExecutionTime    int64 // nanoseconds
	MaxExecutionTime int64
	MemoryUsage      int64
	InstructionCount int64
	mu               sync.RWMutex
}

// NewScriptEngine creates a new script engine
func NewScriptEngine(cfg *config.Sandbox) *ScriptEngine {
	if cfg == nil {
		cfg = config.DefaultSandbox()
	}

	se := &ScriptEngine{
		sandbox: cfg,
		metrics: &ScriptEngineMetrics{},
	}

	se.pool = sync.Pool{
		New: func() any {
			vm := goja.New()
			vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

			// Initialize default built-ins
			_ = vm.Set("console", se.sandbox.Console)
			modules := se.sandbox.Modules
			if modules == nil {
				modules = make(map[string]map[string]any)
			}
			modules["math"] = builtMath
			modules["date"] = builtDate
			modules["string"] = builtStr

			_ = vm.Set("require", func(name string) map[string]any {
				return modules[name]
			})

			// Initialize global variables
			for k, v := range se.sandbox.Globals {
				_ = vm.Set(k, v)
			}

			return vm
		},
	}

	return se
}

// Execute executes a script with parameters
func (e *ScriptEngine) Execute(ctx context.Context, script string, input map[string]any) (any, error) {
	start := time.Now()
	e.metrics.ExecutionCount++

	// Get VM from pool
	vm := e.pool.Get().(*goja.Runtime)
	defer e.pool.Put(vm)

	// Create execution context with timeout
	execCtx, cancel := context.WithTimeout(ctx, e.sandbox.Timeout)
	defer cancel()

	// Set input parameters
	for k, v := range input {
		switch expr := v.(type) {
		case string:
			value, err := vm.RunString(expr)
			if err != nil {
				return nil, fmt.Errorf("failed to evaluate input '%s': %v", k, err)
			}
			_ = vm.Set(k, value)
		default:
			_ = vm.Set(k, v)
		}
	}

	// Monitor resources
	done := make(chan error, 1)
	go e.monitorResources(execCtx, vm, done)

	// Execute script
	var result any
	var err error

	go func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("script panic: %v", r)
			}
			done <- err
		}()

		if e.sandbox.IsStrict {
			// Validate imports
			if err := e.validateImports(script); err != nil {
				done <- err
				return
			}

			// Validate API usage
			if err := e.validateAPIs(script); err != nil {
				done <- err
				return
			}
		}

		// Run script
		value, rerr := vm.RunString(script)
		if rerr != nil {
			done <- fmt.Errorf("script execution failed: %w", rerr)
			return
		}

		// Convert result
		result = value.Export()
		done <- nil
	}()

	// Wait for completion or timeout
	select {
	case err := <-done:
		if err != nil {
			e.metrics.FailureCount++
			return nil, err
		}
	case <-execCtx.Done():
		e.metrics.TimeoutCount++
		return nil, fmt.Errorf("script execution timeout")
	}

	// Update metrics
	duration := time.Since(start)
	e.metrics.ExecutionTime += duration.Nanoseconds()
	if duration.Nanoseconds() > e.metrics.MaxExecutionTime {
		e.metrics.MaxExecutionTime = duration.Nanoseconds()
	}
	e.metrics.SuccessCount++

	return result, nil
}

// Validate validates a script
func (e *ScriptEngine) Validate(script string) error {
	if err := e.validateImports(script); err != nil {
		return err
	}

	if err := e.validateAPIs(script); err != nil {
		return err
	}

	_, err := goja.Compile("", script, true)
	if err != nil {
		return fmt.Errorf("script compilation failed: %w", err)
	}

	return nil
}

// Helper functions

// parseScriptConfig parses script configuration
func (h *ScriptHandler) parseScriptConfig(node *structs.ReadNode) (*config.ScriptHandlerConfig, error) {
	c, ok := node.Properties["scriptConfig"].(map[string]any)
	if !ok {
		return nil, types.NewError(types.ErrValidation, "missing script configuration", nil)
	}

	result := config.DefaultScriptHandlerConfig()

	// Parse fields
	if script, ok := c["script"].(string); ok {
		result.Script = script
	}
	if language, ok := c["language"].(string); ok {
		result.Language = language
	}
	if timeout, ok := c["timeout"].(float64); ok {
		result.Timeout = time.Duration(int(timeout))
	}
	if maxMemory, ok := c["max_memory"].(float64); ok {
		result.MaxMemory = int64(maxMemory)
	}
	if apis, ok := c["allowed_apis"].([]any); ok {
		result.AllowedAPIs = make([]string, len(apis))
		for i, api := range apis {
			if apiStr, ok := api.(string); ok {
				result.AllowedAPIs[i] = apiStr
			}
		}
	}
	if input, ok := c["input"].(map[string]any); ok {
		result.InputVars = input
	}
	if output, ok := c["output"].([]any); ok {
		result.OutputVars = make([]string, len(output))
		for i, out := range output {
			if outStr, ok := out.(string); ok {
				result.OutputVars[i] = outStr
			}
		}
	}
	if mode, ok := c["failure_mode"].(string); ok {
		result.FailureMode = mode
	}

	return result, nil
}

// validateImports validates imports
func (e *ScriptEngine) validateImports(script string) error {
	re := regexp.MustCompile(`require\(['"]([^'"]+)['"]\)`)
	matches := re.FindAllStringSubmatch(script, -1)
	for _, match := range matches {
		if len(match) > 1 {
			pkg := match[1]
			if e.contains(e.sandbox.BlockedPackages, pkg) {
				return fmt.Errorf("blocked import: %s", pkg)
			}
		}
	}
	return nil
}

// validateAPIs validates API usage
func (e *ScriptEngine) validateAPIs(script string) error {
	for _, api := range e.sandbox.AllowedAPIs {
		re := regexp.MustCompile(fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(api)))
		if !re.MatchString(script) {
			return fmt.Errorf("blocked API usage: %s", api)
		}
	}
	return nil
}

// contains checks if a string is in an array
func (e *ScriptEngine) contains(arr []string, s string) bool {
	for _, item := range arr {
		if item == s {
			return true
		}
	}
	return false
}

// monitorResources monitors resource usage
func (e *ScriptEngine) monitorResources(ctx context.Context, vm *goja.Runtime, done chan<- error) {
	ticker := time.NewTicker(time.Millisecond * 100)
	defer ticker.Stop()

	var memStats runtime.MemStats
	var instructionCount int64
	var eventLoopLag time.Duration

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Check memory usage
			runtime.ReadMemStats(&memStats)
			if memStats.Alloc > uint64(e.sandbox.MaxMemory) {
				done <- fmt.Errorf("memory limit exceeded")
				return
			}

			// Check CPU usage
			cpuUsage := float64(memStats.NumGC) / float64(e.sandbox.MaxCPU) * 100
			if cpuUsage > 100 {
				done <- fmt.Errorf("cpu limit exceeded")
				return
			}

			// Check instruction count
			instructionCount = vm.GlobalObject().Get("$instCount").ToInteger()
			if instructionCount > e.sandbox.MaxInstructions {
				done <- fmt.Errorf("instruction limit exceeded")
				return
			}

			// Check event loop lag
			start := time.Now()
			vm.Interrupt("check event loop lag")
			eventLoopLag = time.Since(start)
			if eventLoopLag > e.sandbox.MaxEventLoopLag {
				done <- fmt.Errorf("event loop blocked")
				return
			}
		}
	}
}

// Built-in objects exposed to scripts
var builtMath = map[string]any{
	"abs":   math.Abs,
	"floor": math.Floor,
	"ceil":  math.Ceil,
	"round": math.Round,
	"max":   math.Max,
	"min":   math.Min,
}

var builtDate = map[string]any{
	"now": time.Now,
	"parse": func(value string, layout ...string) (time.Time, error) {
		l := "2006-01-02 15:04:05"
		if len(layout) > 0 {
			l = layout[0]
		}
		return time.Parse(l, value)
	},
	"format": func(t time.Time, layout ...string) string {
		l := "2006-01-02 15:04:05"
		if len(layout) > 0 {
			l = layout[0]
		}
		return t.Format(l)
	},
}
var builtStr = map[string]any{
	"trim":      strings.TrimSpace,
	"toLower":   strings.ToLower,
	"toUpper":   strings.ToUpper,
	"split":     strings.Split,
	"join":      strings.Join,
	"replace":   strings.Replace,
	"hasPrefix": strings.HasPrefix,
	"hasSuffix": strings.HasSuffix,
}
