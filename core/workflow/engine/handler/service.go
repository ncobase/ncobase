package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"ncobase/workflow/engine/config"
	"ncobase/workflow/engine/types"
	"ncobase/workflow/engine/utils"
	"ncobase/workflow/service"
	"ncobase/workflow/structs"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	ext "github.com/ncobase/ncore/extension/types"
)

// ServiceHandler handles service node execution
type ServiceHandler struct {
	*BaseHandler

	// Service registry and discovery
	registry  *ServiceRegistry
	discovery *ServiceDiscovery

	// Configuration
	config *config.ServiceHandlerConfig

	// Cache management
	cache *ServiceCache

	// Circuit breaker
	breaker *CircuitBreaker

	// Monitoring
	metrics *ServiceMetrics

	// Runtime state
	activeServices sync.Map // serviceID -> *ServiceInfo
	mu             sync.RWMutex
	ctx            context.Context
}

// ServiceInfo represents runtime service info
type ServiceInfo struct {
	ID         string
	Name       string
	Type       string
	Status     string
	StartTime  time.Time
	EndTime    *time.Time
	Duration   time.Duration
	RetryCount int
	Error      error
	mu         sync.RWMutex
}

// ServiceRegistry manages service providers
type ServiceRegistry struct {
	providers map[string]ServiceProvider
	mu        sync.RWMutex
}

// ServiceProvider defines service provider interface
type ServiceProvider interface {
	// Basic info

	Name() string
	Type() string
	Capabilities() *ServiceCapabilities

	// Core operations

	Execute(ctx context.Context, req *ServiceRequest) (*ServiceResponse, error)
	Validate(req *ServiceRequest) error
	Healthcheck() error

	// Optional operations

	Compensate(ctx context.Context, req *ServiceRequest) (*ServiceResponse, error)
	Schema() *ServiceSchema
}

// RetrySettings defines retry settings
type RetrySettings struct {
	MaxRetries    int
	RetryInterval time.Duration
}

// ServiceCapabilities defines provider capabilities
type ServiceCapabilities struct {
	SupportsAsync    bool
	SupportsCallback bool
	SupportsBatch    bool
	SupportsRollback bool
	MaxConcurrent    int
	MaxBatchSize     int
	Timeout          time.Duration
	RetrySettings    *RetrySettings
}

// ServiceRequest represents a service request
type ServiceRequest struct {
	RequestID   string
	Method      string
	Endpoint    string
	Headers     map[string]string
	Parameters  map[string]any
	Body        any
	Timeout     time.Duration
	RetryPolicy *RetryPolicy
}

// ServiceResponse represents a service response
type ServiceResponse struct {
	Status   int
	Headers  map[string]string
	Body     any
	Duration time.Duration
	Error    error
}

// ServiceSchema represents service schema
type ServiceSchema struct {
	Input  map[string]FieldSchema
	Output map[string]FieldSchema
}

// FieldSchema represents field schema
type FieldSchema struct {
	Type       string
	Required   bool
	Default    any
	Validation string
}

// ServiceCache implements service response caching
type ServiceCache struct {
	items   sync.Map // key -> *CacheItem
	maxSize int
	ttl     time.Duration
}

type CacheItem struct {
	Response  *ServiceResponse
	ExpiresAt time.Time
}

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	failures    atomic.Int32
	lastFailure atomic.Int64
	threshold   int32
	timeout     time.Duration
	mu          sync.RWMutex
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxAttempts     int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	MaxElapsedTime  time.Duration
}

// ServiceMetrics tracks service metrics
type ServiceMetrics struct {
	RequestCount atomic.Int64
	SuccessCount atomic.Int64
	FailureCount atomic.Int64
	TimeoutCount atomic.Int64
	RetryCount   atomic.Int64
	ResponseTime atomic.Int64
	CacheHits    atomic.Int64
	CacheMisses  atomic.Int64
}

// NewServiceHandler creates a new service handler
func NewServiceHandler(svc *service.Service, em ext.ManagerInterface, cfg *config.Config) (*ServiceHandler, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	handler := &ServiceHandler{
		BaseHandler: NewBaseHandler("service", "Service Handler", svc, em, cfg.Handlers.Base),
		config:      cfg.Handlers.Service,
		registry:    NewServiceRegistry(),
		discovery:   NewServiceDiscovery(),
		cache:       NewServiceCache(),
		breaker:     NewCircuitBreaker(),
		metrics:     &ServiceMetrics{},
	}

	// Register built-in providers
	handler.registerBuiltinProviders()

	return handler, nil
}

// Type returns handler type
func (h *ServiceHandler) Type() types.HandlerType { return h.handlerType }

// Name returns handler name
func (h *ServiceHandler) Name() string { return h.name }

// Priority returns handler priority
func (h *ServiceHandler) Priority() int { return h.priority }

// Start starts the service handler
func (h *ServiceHandler) Start() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if err := h.BaseHandler.Start(); err != nil {
		return err
	}

	// Initialize service registry
	h.registry = NewServiceRegistry()

	// Initialize service discovery
	if err := h.discovery.Start(); err != nil {
		return err
	}

	// Initialize circuit breaker
	h.breaker = NewCircuitBreaker()

	// Start metric collection
	go h.collectMetrics()

	h.status = types.HandlerRunning

	return nil
}

// Stop stops the service handler
func (h *ServiceHandler) Stop() error {
	if err := h.BaseHandler.Stop(); err != nil {
		return err
	}

	if h.status != types.HandlerRunning {
		return nil
	}

	// Stop service discovery
	h.discovery.Stop()

	// Clean up active services
	h.activeServices.Range(func(key, _ any) bool {
		h.activeServices.Delete(key)
		return true
	})

	// Cleanup resources
	h.cleanup()

	h.status = types.HandlerStopped
	return nil
}

// Reset resets the service handler
func (h *ServiceHandler) Reset() error {
	return h.BaseHandler.Reset()
}

// GetState returns handler state
func (h *ServiceHandler) GetState() *types.HandlerState {
	baseState := h.BaseHandler.GetState()

	// Add service-specific info
	activeCount := 0
	h.activeServices.Range(func(_, _ any) bool {
		activeCount++
		return true
	})

	if extra, ok := (*baseState)["extra"].(map[string]any); ok {
		extra["active_services"] = activeCount
		// extra["circuit_breaker"] = h.breaker.GetState()
	} else {
		(*baseState)["extra"] = map[string]any{
			"active_services": activeCount,
			// "circuit_breaker": h.breaker.GetState(),
		}
	}

	return baseState
}

// Execute executes a service node
func (h *ServiceHandler) Execute(ctx context.Context, node *structs.ReadNode) (*types.Response, error) {
	startTime := time.Now()

	// Parse config
	cfg, err := h.parseServiceConfig(node)
	if err != nil {
		return nil, err
	}

	// Check circuit breaker
	if !h.breaker.Allow() {
		return nil, types.NewError(types.ErrServiceUnavailable, "circuit breaker is open", nil)
	}

	var response *types.Response

	// Execute with retry if enabled
	err = utils.RetryWithBackoff(ctx, func(ctx context.Context) error {
		// Track service execution
		info := &ServiceInfo{
			ID:        node.ID,
			Name:      cfg.Name,
			Type:      cfg.Type,
			StartTime: time.Now(),
			Status:    string(types.ExecutionActive),
		}
		h.activeServices.Store(node.ID, info)
		defer h.activeServices.Delete(node.ID)

		// Check cache if enabled
		// if cfg.EnableCache {
		// 	if cached := h.cache.Get(node.ID); cached != nil {
		// 		response = cached
		// 		return nil
		// 	}
		// }

		// Get service provider
		provider, err := h.registry.GetProvider(cfg.Name)
		if err != nil {
			return err
		}

		// Build request
		req, err := h.buildRequest(ctx, node, cfg)
		if err != nil {
			return err
		}

		// Execute service
		resp, err := provider.Execute(ctx, req)
		if err != nil {
			info.Status = string(types.ExecutionError)
			info.Error = err
			h.breaker.RecordFailure()
			return err
		}

		// Update service info
		info.Status = string(types.ExecutionCompleted)
		now := time.Now()
		info.EndTime = &now

		// Cache response if enabled
		if cfg.EnableCache {
			h.cache.Set(node.ID, resp, cfg.CacheTTL)
		}

		// Build response
		response = &types.Response{
			ID:        node.ID,
			Status:    types.ExecutionCompleted,
			Data:      resp.Body,
			StartTime: startTime,
			EndTime:   &now,
			Duration:  now.Sub(startTime),
		}

		return nil
	}, cfg.RetryPolicy)

	if err != nil {
		h.recordError(err)
		return nil, err
	}

	return response, nil
}

// Complete completes a service node
func (h *ServiceHandler) Complete(ctx context.Context, node *structs.ReadNode, req *structs.CompleteTaskRequest) error {
	info, ok := h.activeServices.Load(node.ID)
	if !ok {
		return types.NewError(types.ErrNotFound, "service not found", nil)
	}
	svcInfo := info.(*ServiceInfo)

	// Update service status
	svcInfo.Status = string(types.ExecutionCompleted)
	now := time.Now()
	svcInfo.EndTime = &now

	// Process response variables if any
	return h.processResponseVariables(ctx, node, req.Variables)
}

// Cancel cancels a service execution
func (h *ServiceHandler) Cancel(ctx context.Context, id string) error {
	info, ok := h.activeServices.Load(id)
	if !ok {
		return types.NewError(types.ErrNotFound, "service not found", nil)
	}
	svcInfo := info.(*ServiceInfo)

	// Get provider
	provider, err := h.registry.GetProvider(svcInfo.Type)
	if err != nil {
		return err
	}

	// Cancel through provider if supported
	if canceler, ok := provider.(interface {
		Cancel(context.Context, string) error
	}); ok {
		if err := canceler.Cancel(ctx, id); err != nil {
			return err
		}
	}

	// Update service status
	svcInfo.Status = string(types.ExecutionCancelled)
	now := time.Now()
	svcInfo.EndTime = &now

	h.activeServices.Delete(id)

	return nil
}

// cancelService cancels a service execution
func (h *ServiceHandler) cancelService(ctx context.Context, info *ServiceInfo) error {
	// Get provider
	provider, err := h.registry.GetProvider(info.Type)
	if err != nil {
		return err
	}

	// Cancel through provider if supported
	if canceler, ok := provider.(interface {
		Cancel(context.Context, string) error
	}); ok {
		return canceler.Cancel(ctx, info.ID)
	}

	// Default cancellation
	info.Status = string(types.ExecutionCancelled)
	now := time.Now()
	info.EndTime = &now

	// Publish event
	h.em.PublishEvent("service.cancelled", map[string]any{
		"service_id": info.ID,
		"time":       now,
	})

	return nil
}

// Rollback rolls back a service execution
func (h *ServiceHandler) Rollback(ctx context.Context, id string) error {
	info, ok := h.activeServices.Load(id)
	if !ok {
		return types.NewError(types.ErrNotFound, "service not found", nil)
	}
	svcInfo := info.(*ServiceInfo)

	// Get provider
	provider, err := h.registry.GetProvider(svcInfo.Type)
	if err != nil {
		return err
	}

	// Execute rollback if supported
	if rollbacker, ok := provider.(interface {
		Rollback(context.Context, *ServiceRequest) error
	}); ok {
		return rollbacker.Rollback(ctx, &ServiceRequest{
			RequestID: id,
			Body: map[string]any{
				"type": svcInfo.Type,
			},
		})
	}

	return types.NewError(types.ErrNotSupported, "rollback not supported", nil)
}

// Validate validates a service node
func (h *ServiceHandler) Validate(node *structs.ReadNode) error {
	// Parse config
	cfg, err := h.parseServiceConfig(node)
	if err != nil {
		return err
	}

	// Get provider
	provider, err := h.registry.GetProvider(cfg.Name)
	if err != nil {
		return err
	}

	// Build request for validation
	req, err := h.buildRequest(context.Background(), node, cfg)
	if err != nil {
		return err
	}

	// Validate request
	return provider.Validate(req)
}

// Event handling

// HandleTimeout handles service timeout
func (h *ServiceHandler) HandleTimeout(ctx context.Context, nodeID string) error {
	info, ok := h.activeServices.Load(nodeID)
	if !ok {
		return types.NewError(types.ErrNotFound, "service not found", nil)
	}
	svcInfo := info.(*ServiceInfo)

	// Update service status
	svcInfo.mu.Lock()
	svcInfo.Status = string(types.ExecutionTimeout)
	svcInfo.Error = types.NewError(types.ErrTimeout, "service timeout", nil)
	svcInfo.mu.Unlock()

	h.activeServices.Store(nodeID, svcInfo)

	// Publish timeout event
	h.em.PublishEvent("service.timeout", map[string]any{
		"node_id":  nodeID,
		"type":     svcInfo.Type,
		"duration": time.Since(svcInfo.StartTime),
	})

	return nil
}

// HandleError handles service error
func (h *ServiceHandler) HandleError(ctx context.Context, nodeID string, err error) error {
	info, ok := h.activeServices.Load(nodeID)
	if !ok {
		return types.NewError(types.ErrNotFound, "service not found", nil)
	}
	svcInfo := info.(*ServiceInfo)

	// Update service status
	svcInfo.mu.Lock()
	svcInfo.Status = string(types.ExecutionError)
	svcInfo.Error = err
	svcInfo.mu.Unlock()

	h.activeServices.Store(nodeID, svcInfo)

	// Publish error event
	h.em.PublishEvent("service.error", map[string]any{
		"node_id": nodeID,
		"type":    svcInfo.Type,
		"error":   err.Error(),
	})

	return nil
}

// parseServiceConfig parses service configuration from node properties
func (h *ServiceHandler) parseServiceConfig(node *structs.ReadNode) (*config.ServiceHandlerConfig, error) {
	c, ok := node.Properties["serviceConfig"].(map[string]any)
	if !ok {
		return nil, types.NewError(types.ErrValidation, "missing service configuration", nil)
	}

	result := config.DefaultServiceHandlerConfig()

	// Parse required fields
	if name, ok := c["name"].(string); ok {
		result.Name = name
	} else {
		return nil, types.NewError(types.ErrValidation, "service name is required", nil)
	}

	if svcType, ok := c["type"].(string); ok {
		result.Type = svcType
	}

	// Parse optional fields
	if endpoint, ok := c["endpoint"].(string); ok {
		result.Endpoint = endpoint
	}
	if method, ok := c["method"].(string); ok {
		result.Method = method
	}
	if headers, ok := c["headers"].(map[string]string); ok {
		result.Headers = headers
	}
	if input, ok := c["input_vars"].(map[string]any); ok {
		result.InputVars = input
	}
	if output, ok := c["output_vars"].(map[string]any); ok {
		result.OutputVars = output
	}
	if timeout, ok := c["timeout"].(float64); ok {
		result.Timeout = time.Duration(timeout) * time.Second
	}
	if retryPolicy, ok := c["retry_policy"].(config.RetryConfig); ok {
		result.RetryPolicy = &retryPolicy
	}
	if mode, ok := c["failure_mode"].(string); ok {
		result.FailureMode = mode
	}
	if async, ok := c["async"].(bool); ok {
		result.Async = async
	}
	if enableCache, ok := c["enable_cache"].(bool); ok {
		result.EnableCache = enableCache
	}
	if cacheTTL, ok := c["cache_ttl"].(float64); ok {
		result.CacheTTL = time.Duration(cacheTTL) * time.Second
	}
	return result, nil
}

// buildRequest builds service request from node and config
func (h *ServiceHandler) buildRequest(ctx context.Context, node *structs.ReadNode, cfg *config.ServiceHandlerConfig) (*ServiceRequest, error) {
	// Get process variables
	process, err := h.services.Process.Get(ctx, &structs.FindProcessParams{
		ProcessKey: node.ProcessID,
	})
	if err != nil {
		return nil, err
	}

	// Build parameters
	params := make(map[string]any)
	if cfg.InputVars != nil {
		for k, v := range cfg.InputVars {
			if expr, ok := v.(string); ok && len(expr) > 0 && expr[0] == '$' {
				// Replace with process variable
				varName := expr[1:]
				if val, exists := process.Variables[varName]; exists {
					params[k] = val
					continue
				}
			}
			params[k] = v
		}
	}

	return &ServiceRequest{
		RequestID:  node.ID,
		Method:     cfg.Method,
		Endpoint:   cfg.Endpoint,
		Headers:    cfg.Headers,
		Parameters: params,
		Timeout:    cfg.Timeout,
	}, nil
}

// processResponseVariables processes service response variables
func (h *ServiceHandler) processResponseVariables(ctx context.Context, node *structs.ReadNode, response map[string]any) error {
	// Get process
	process, err := h.services.Process.Get(ctx, &structs.FindProcessParams{
		ProcessKey: node.ProcessID,
	})
	if err != nil {
		return err
	}

	// Update process variables with output
	if process.Variables == nil {
		process.Variables = make(map[string]any)
	}

	cfg, err := h.parseServiceConfig(node)
	if err != nil {
		return err
	}

	for k, v := range response {
		if cfg.OutputVars != nil {
			if mapping, ok := cfg.OutputVars[k]; ok {
				process.Variables[mapping.(string)] = v
			}
		}
	}

	// Update process
	_, err = h.services.Process.Update(ctx, &structs.UpdateProcessBody{
		ID: process.ID,
		ProcessBody: structs.ProcessBody{
			Variables: process.Variables,
		},
	})

	return err
}

// executeSync executes service synchronously
func (h *ServiceHandler) executeSync(ctx context.Context, provider ServiceProvider, req *ServiceRequest, cfg *config.ServiceHandlerConfig) (*ServiceResponse, error) {
	// Check circuit breaker
	if !h.breaker.Allow() {
		return nil, types.NewError(types.ErrServiceUnavailable, "circuit breaker is open", nil)
	}

	// Check cache
	if cfg.EnableCache {
		if resp := h.cache.Get(h.getCacheKey(req)); resp != nil {
			h.metrics.CacheHits.Add(1)
			return resp, nil
		}
		h.metrics.CacheMisses.Add(1)
	}

	startTime := time.Now()
	h.metrics.RequestCount.Add(1)

	var resp *ServiceResponse
	// var err error

	// // Execute with retry
	// err = h.retryPolicy.Execute(ctx, func() error {
	// 	resp, err = provider.Execute(ctx, req)
	// 	return err
	// })

	// if err != nil {
	// 	h.metrics.FailureCount.Add(1)
	// 	h.breaker.RecordFailure()
	// 	return nil, err
	// }

	// Update metrics
	h.metrics.SuccessCount.Add(1)
	h.metrics.ResponseTime.Add(time.Since(startTime).Nanoseconds())

	// Cache response
	if cfg.EnableCache {
		h.cache.Set(h.getCacheKey(req), resp, cfg.CacheTTL)
	}

	return resp, nil
}

// executeAsync executes service asynchronously
func (h *ServiceHandler) executeAsync(ctx context.Context, node *structs.ReadNode, provider ServiceProvider, req *ServiceRequest, cfg *config.ServiceHandlerConfig) {
	resp, err := h.executeSync(ctx, provider, req, cfg)

	// Store result
	info := &ServiceInfo{
		ID:        node.ID,
		Name:      cfg.Name,
		Type:      cfg.Type,
		StartTime: time.Now(),
	}

	if err != nil {
		info.Error = err
		info.Status = string(types.ExecutionError)
	} else {
		info.Status = string(types.ExecutionCompleted)
		endTime := info.StartTime.Add(resp.Duration)
		info.EndTime = &endTime
	}

	h.activeServices.Store(node.ID, info)

	// Update node status
	h.updateNodeStatus(ctx, node, info.Status)
}

// handleResponse handles service response
func (h *ServiceHandler) handleResponse(ctx context.Context, node *structs.ReadNode, resp *ServiceResponse) error {
	// Get process
	process, err := h.services.Process.Get(ctx, &structs.FindProcessParams{
		ProcessKey: node.ProcessID,
	})
	if err != nil {
		return err
	}

	// Update process variables with output
	if output, ok := resp.Body.(map[string]any); ok {
		if process.Variables == nil {
			process.Variables = make(map[string]any)
		}

		if outputMapping, ok := node.Properties["outputMapping"].(map[string]any); ok {
			for varName, outputKey := range outputMapping {
				if key, ok := outputKey.(string); ok {
					if value, exists := output[key]; exists {
						process.Variables[varName] = value
					}
				}
			}
		}
	}

	// Update process
	_, err = h.services.Process.Update(ctx, &structs.UpdateProcessBody{
		ID: process.ID,
		ProcessBody: structs.ProcessBody{
			Variables: process.Variables,
		},
	})

	return err
}

// Registry implementation

func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		providers: make(map[string]ServiceProvider),
	}
}

func (r *ServiceRegistry) RegisterProvider(name string, provider ServiceProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if provider == nil {
		return types.NewError(types.ErrInvalidParam, "provider is nil", nil)
	}

	if _, exists := r.providers[name]; exists {
		return types.NewError(types.ErrConflict, "provider already exists", nil)
	}

	r.providers[name] = provider
	return nil
}

func (r *ServiceRegistry) GetProvider(name string) (ServiceProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[name]
	if !exists {
		return nil, types.NewError(types.ErrNotFound, "provider not found", nil)
	}

	return provider, nil
}

// Service discovery implementation

type ServiceDiscovery struct {
	services   map[string]*ServiceEndpoint
	mu         sync.RWMutex
	done       chan struct{}
	updateChan chan *ServiceEndpoint
}

type ServiceEndpoint struct {
	Name      string
	URL       string
	Status    string
	LastCheck time.Time
	Metadata  map[string]string
}

func NewServiceDiscovery() *ServiceDiscovery {
	return &ServiceDiscovery{
		services:   make(map[string]*ServiceEndpoint),
		done:       make(chan struct{}),
		updateChan: make(chan *ServiceEndpoint, 100),
	}
}

func (d *ServiceDiscovery) Start() error {
	go d.watchUpdates()
	go d.healthCheck()
	return nil
}

func (d *ServiceDiscovery) Stop() {
	close(d.done)
}

func (d *ServiceDiscovery) watchUpdates() {
	for {
		select {
		case <-d.done:
			return
		case endpoint := <-d.updateChan:
			d.mu.Lock()
			d.services[endpoint.Name] = endpoint
			d.mu.Unlock()
		}
	}
}

func (d *ServiceDiscovery) healthCheck() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-d.done:
			return
		case <-ticker.C:
			d.mu.Lock()
			for name, endpoint := range d.services {
				if time.Since(endpoint.LastCheck) > time.Minute*5 {
					endpoint.Status = "unknown"
				}
				d.services[name] = endpoint
			}
			d.mu.Unlock()
		}
	}
}

// Cache implementation

func NewServiceCache() *ServiceCache {
	return &ServiceCache{
		maxSize: 1000,
		ttl:     time.Minute,
	}
}

func (c *ServiceCache) Get(key string) *ServiceResponse {
	value, ok := c.items.Load(key)
	if !ok {
		return nil
	}

	item := value.(*CacheItem)
	if time.Now().After(item.ExpiresAt) {
		c.items.Delete(key)
		return nil
	}

	return item.Response
}

func (c *ServiceCache) Set(key string, resp *ServiceResponse, ttl time.Duration) {
	c.items.Store(key, &CacheItem{
		Response:  resp,
		ExpiresAt: time.Now().Add(ttl),
	})
}

// Circuit breaker implementation

func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		threshold: 5,
		timeout:   time.Second * 60,
	}
}

func (cb *CircuitBreaker) Allow() bool {
	failures := cb.failures.Load()
	if failures >= cb.threshold {
		lastFailure := time.Unix(0, cb.lastFailure.Load())
		if time.Since(lastFailure) > cb.timeout {
			cb.failures.Store(0)
			return true
		}
		return false
	}
	return true
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.failures.Add(1)
	cb.lastFailure.Store(time.Now().UnixNano())
}

// HTTPServiceProvider implements HTTP service calls
type HTTPServiceProvider struct {
	client  *http.Client
	baseURL string
	timeout time.Duration
	retries int
}

// NewHTTPServiceProvider creates a new HTTP service provider
func NewHTTPServiceProvider(baseURL string) *HTTPServiceProvider {
	return &HTTPServiceProvider{
		client: &http.Client{
			Timeout: time.Second * 30,
			Transport: &http.Transport{
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   10,
				MaxConnsPerHost:       100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
		baseURL: baseURL,
		timeout: time.Second * 30,
		retries: 3,
	}
}

func (p *HTTPServiceProvider) Name() string {
	return "http"
}

func (p *HTTPServiceProvider) Type() string {
	return "http"
}

func (p *HTTPServiceProvider) Capabilities() *ServiceCapabilities {
	return &ServiceCapabilities{
		SupportsAsync:    false,
		SupportsCallback: false,
		SupportsBatch:    false,
		SupportsRollback: false,
		MaxConcurrent:    100,
		Timeout:          p.timeout,
		RetrySettings: &RetrySettings{
			MaxRetries:    p.retries,
			RetryInterval: time.Second,
		},
	}
}

func (p *HTTPServiceProvider) Execute(ctx context.Context, req *ServiceRequest) (*ServiceResponse, error) {
	// Build HTTP request
	httpReq, err := p.buildHTTPRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	// Execute request with retry
	var resp *http.Response
	var startTime = time.Now()
	for i := 0; i <= p.retries; i++ {
		resp, err = p.client.Do(httpReq)
		if err == nil {
			break
		}

		if !p.shouldRetry(err) {
			return nil, err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(i+1) * time.Second):
			continue
		}
	}

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check response status
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(body))

	}

	// Parse response
	var result any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &ServiceResponse{
		Status:   resp.StatusCode,
		Headers:  flattenHeaders(resp.Header),
		Body:     result,
		Duration: time.Since(startTime),
	}, nil
}

func (p *HTTPServiceProvider) Validate(cfg *ServiceRequest) error {
	if cfg.Endpoint == "" {
		return types.NewError(types.ErrValidation, "endpoint is required", nil)
	}

	if !isValidHTTPMethod(cfg.Method) {
		return types.NewError(types.ErrValidation, "invalid HTTP method", nil)
	}

	if _, err := url.Parse(cfg.Endpoint); err != nil {
		return types.NewError(types.ErrValidation, "invalid endpoint URL", err)
	}

	return nil
}

func (p *HTTPServiceProvider) Healthcheck() error {
	req, err := http.NewRequest("GET", p.baseURL+"/health", nil)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

// Compensate handles the compensation request
func (p *HTTPServiceProvider) Compensate(ctx context.Context, req *ServiceRequest) (*ServiceResponse, error) {
	// Build compensate request
	compensateReq := &ServiceRequest{
		Method:   "POST",
		Endpoint: fmt.Sprintf("/compensate/%s", req.RequestID),
		Headers: map[string]string{
			"X-Original-Request-ID": req.RequestID,
		},
		Body: map[string]any{
			"original_request": req,
			"timestamp":        time.Now(),
		},
	}

	return p.Execute(ctx, compensateReq)
}

// Schema returns the service schema
func (p *HTTPServiceProvider) Schema() *ServiceSchema {
	// TODO implement me
	panic("implement me")
}

func (p *HTTPServiceProvider) buildHTTPRequest(ctx context.Context, req *ServiceRequest) (*http.Request, error) {
	u := p.baseURL + req.Endpoint

	var body io.Reader
	if req.Body != nil {
		data, err := json.Marshal(req.Body)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(data)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, u, body)
	if err != nil {
		return nil, err
	}

	// Add headers
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Add query parameters
	q := httpReq.URL.Query()
	for k, v := range req.Parameters {
		q.Add(k, fmt.Sprint(v))
	}
	httpReq.URL.RawQuery = q.Encode()

	return httpReq, nil
}

func (p *HTTPServiceProvider) shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// Check standard errors
	if errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}

	// Check network errors
	if netErr, ok := err.(net.Error); ok {
		return netErr.Temporary() || netErr.Timeout()
	}

	// Check syscall errors
	var sysErr *os.SyscallError
	if errors.As(err, &sysErr) {
		switch {
		case errors.Is(sysErr.Err, syscall.ECONNREFUSED), errors.Is(sysErr.Err, syscall.ECONNRESET), errors.Is(sysErr.Err, syscall.ETIMEDOUT):
			return true
		}
	}

	return false
}

// flattenHeaders flattens HTTP headers into a map[string]string
func flattenHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)
	for k, v := range headers {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}

func isValidHTTPMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut,
		http.MethodDelete, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

// Event publishing helper
func (h *ServiceHandler) publishEvent(eventType string, details map[string]any) {
	if details == nil {
		details = make(map[string]any)
	}
	details["time"] = time.Now()
	h.em.PublishEvent(eventType, details)
}

// Resource cleanup
func (h *ServiceHandler) cleanup() {
	// Clear caches
	h.cache.items.Range(func(key, _ any) bool {
		h.cache.items.Delete(key)
		return true
	})

	// Clear active services
	h.activeServices.Range(func(key, _ any) bool {
		h.activeServices.Delete(key)
		return true
	})

	// Reset metrics
	h.metrics = &ServiceMetrics{}
}

func (h *ServiceHandler) GetMetrics() map[string]any {
	return map[string]any{
		"request_count":    h.metrics.RequestCount.Load(),
		"success_count":    h.metrics.SuccessCount.Load(),
		"failure_count":    h.metrics.FailureCount.Load(),
		"timeout_count":    h.metrics.TimeoutCount.Load(),
		"retry_count":      h.metrics.RetryCount.Load(),
		"response_time_ns": h.metrics.ResponseTime.Load(),
		"cache_hits":       h.metrics.CacheHits.Load(),
		"cache_misses":     h.metrics.CacheMisses.Load(),
		"circuit_breaker": map[string]any{
			"failures":     h.breaker.failures.Load(),
			"last_failure": time.Unix(0, h.breaker.lastFailure.Load()),
		},
	}
}

// Background metrics collection
func (h *ServiceHandler) collectMetrics() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			if h.metrics == nil {
				continue
			}

			// Update metrics
			var activeCount int32
			h.activeServices.Range(func(_, _ any) bool {
				activeCount++
				return true
			})

			// Report metrics
			h.publishEvent("service.metrics", map[string]any{
				"active_services": activeCount,
				"metrics":         h.GetMetrics(),
			})
		}
	}
}

// isAsyncService checks if service is async
func (h *ServiceHandler) isAsyncService(node *structs.ReadNode) bool {
	if cfg, ok := node.Properties["serviceConfig"].(map[string]any); ok {
		if async, ok := cfg["async"].(bool); ok {
			return async
		}
	}
	return false
}

// getAsyncResult gets async execution result
func (h *ServiceHandler) getAsyncResult(nodeID string) (*ServiceResponse, error) {
	info, ok := h.activeServices.Load(nodeID)
	if !ok {
		return nil, types.NewError(types.ErrNotFound, "async result not found", nil)
	}

	svcInfo := info.(*ServiceInfo)
	if svcInfo.Error != nil {
		return nil, svcInfo.Error
	}

	return &ServiceResponse{
		Status:   200,
		Duration: svcInfo.Duration,
	}, nil
}

// updateNodeStatus updates node execution status
func (h *ServiceHandler) updateNodeStatus(ctx context.Context, node *structs.ReadNode, status string) error {
	_, err := h.services.Node.Update(ctx, &structs.UpdateNodeBody{
		ID: node.ID,
		NodeBody: structs.NodeBody{
			Status: status,
		},
	})
	return err
}

// buildCompensateRequest builds compensation request
func (h *ServiceHandler) buildCompensateRequest(ctx context.Context, node *structs.ReadNode) (*ServiceRequest, error) {
	// Get original request details
	info, ok := h.activeServices.Load(node.ID)
	if !ok {
		return nil, types.NewError(types.ErrNotFound, "service info not found", nil)
	}
	svcInfo := info.(*ServiceInfo)

	// Build compensation request
	return &ServiceRequest{
		Method:   "POST",
		Endpoint: fmt.Sprintf("/compensate/%s", svcInfo.ID),
		Headers: map[string]string{
			"X-Original-ID": svcInfo.ID,
		},
		Parameters: map[string]any{
			"service": svcInfo.Name,
			"type":    svcInfo.Type,
			"time":    svcInfo.StartTime,
		},
	}, nil
}

// getCacheKey generates cache key for request
func (h *ServiceHandler) getCacheKey(req *ServiceRequest) string {
	// Generate cache key based on request properties
	key := fmt.Sprintf("%s:%s", req.Method, req.Endpoint)

	// Add query parameters
	var params []string
	for k, v := range req.Parameters {
		params = append(params, fmt.Sprintf("%s=%v", k, v))
	}
	if len(params) > 0 {
		sort.Strings(params)
		key += "?" + strings.Join(params, "&")
	}

	return key
}

// registerBuiltinProviders registers built-in service providers
func (h *ServiceHandler) registerBuiltinProviders() {
	// Register HTTP provider
	h.registry.RegisterProvider("http", NewHTTPServiceProvider(""))

	// Register REST provider with default config
	restCfg := &RESTProviderConfig{
		BaseURL: "",
		Timeout: time.Second * 30,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
	h.registry.RegisterProvider("rest", NewRESTServiceProvider(restCfg))

	// Register SOAP provider
	soapCfg := &SOAPProviderConfig{
		WSDL:    "",
		Timeout: time.Second * 60,
	}
	h.registry.RegisterProvider("soap", NewSOAPServiceProvider(soapCfg))

	// Register gRPC provider
	grpcCfg := &GRPCProviderConfig{
		Target:  "",
		Timeout: time.Second * 30,
	}
	h.registry.RegisterProvider("grpc", NewGRPCServiceProvider(grpcCfg))
}

// Additional provider implementations (stubs)

type RESTServiceProvider struct {
	config *RESTProviderConfig
}

func (r RESTServiceProvider) Name() string {
	// TODO implement me
	panic("implement me")
}

func (r RESTServiceProvider) Type() string {
	// TODO implement me
	panic("implement me")
}

func (r RESTServiceProvider) Capabilities() *ServiceCapabilities {
	// TODO implement me
	panic("implement me")
}

func (r RESTServiceProvider) Execute(ctx context.Context, req *ServiceRequest) (*ServiceResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (r RESTServiceProvider) Validate(req *ServiceRequest) error {
	// TODO implement me
	panic("implement me")
}

func (r RESTServiceProvider) Healthcheck() error {
	// TODO implement me
	panic("implement me")
}

func (r RESTServiceProvider) Compensate(ctx context.Context, req *ServiceRequest) (*ServiceResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (r RESTServiceProvider) Schema() *ServiceSchema {
	// TODO implement me
	panic("implement me")
}

type RESTProviderConfig struct {
	BaseURL string
	Timeout time.Duration
	Headers map[string]string
}

func NewRESTServiceProvider(cfg *RESTProviderConfig) *RESTServiceProvider {
	return &RESTServiceProvider{config: cfg}
}

type SOAPServiceProvider struct {
	config *SOAPProviderConfig
}

func (s SOAPServiceProvider) Name() string {
	// TODO implement me
	panic("implement me")
}

func (s SOAPServiceProvider) Type() string {
	// TODO implement me
	panic("implement me")
}

func (s SOAPServiceProvider) Capabilities() *ServiceCapabilities {
	// TODO implement me
	panic("implement me")
}

func (s SOAPServiceProvider) Execute(ctx context.Context, req *ServiceRequest) (*ServiceResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (s SOAPServiceProvider) Validate(req *ServiceRequest) error {
	// TODO implement me
	panic("implement me")
}

func (s SOAPServiceProvider) Healthcheck() error {
	// TODO implement me
	panic("implement me")
}

func (s SOAPServiceProvider) Compensate(ctx context.Context, req *ServiceRequest) (*ServiceResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (s SOAPServiceProvider) Schema() *ServiceSchema {
	// TODO implement me
	panic("implement me")
}

type SOAPProviderConfig struct {
	WSDL    string
	Timeout time.Duration
}

func NewSOAPServiceProvider(cfg *SOAPProviderConfig) *SOAPServiceProvider {
	return &SOAPServiceProvider{config: cfg}
}

type GRPCServiceProvider struct {
	config *GRPCProviderConfig
}

func (G GRPCServiceProvider) Name() string {
	// TODO implement me
	panic("implement me")
}

func (G GRPCServiceProvider) Type() string {
	// TODO implement me
	panic("implement me")
}

func (G GRPCServiceProvider) Capabilities() *ServiceCapabilities {
	// TODO implement me
	panic("implement me")
}

func (G GRPCServiceProvider) Execute(ctx context.Context, req *ServiceRequest) (*ServiceResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (G GRPCServiceProvider) Validate(req *ServiceRequest) error {
	// TODO implement me
	panic("implement me")
}

func (G GRPCServiceProvider) Healthcheck() error {
	// TODO implement me
	panic("implement me")
}

func (G GRPCServiceProvider) Compensate(ctx context.Context, req *ServiceRequest) (*ServiceResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (G GRPCServiceProvider) Schema() *ServiceSchema {
	// TODO implement me
	panic("implement me")
}

type GRPCProviderConfig struct {
	Target  string
	Timeout time.Duration
}

func NewGRPCServiceProvider(cfg *GRPCProviderConfig) *GRPCServiceProvider {
	return &GRPCServiceProvider{config: cfg}
}
