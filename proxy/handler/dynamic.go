package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"ncobase/proxy/service"
	"ncobase/proxy/structs"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation/validator"
	"github.com/sony/gobreaker"
)

// DynamicHandlerInterface is the interface for the dynamic handler.
type DynamicHandlerInterface interface {
	RegisterDynamicRoutes(r *gin.RouterGroup)
	ProxyRequest(c *gin.Context)
	SetExtensionManager(manager ext.ManagerInterface)
}

// dynamicHandler represents the dynamic handler.
type dynamicHandler struct {
	s                *service.Service
	circuitBreakers  map[string]*gobreaker.CircuitBreaker
	httpClient       *http.Client
	transformerCache map[string]service.TransformerFunc
	manager          ext.ManagerInterface
}

// NewDynamicHandler creates a new dynamic handler.
func NewDynamicHandler(svc *service.Service) DynamicHandlerInterface {
	return &dynamicHandler{
		s:               svc,
		circuitBreakers: make(map[string]*gobreaker.CircuitBreaker),
		httpClient: &http.Client{
			Timeout: time.Second * 30,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxConnsPerHost:     100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		transformerCache: make(map[string]service.TransformerFunc),
		manager:          nil, // Will be set later via SetManager
	}
}

// SetExtensionManager sets the extension manager for event publishing
func (h *dynamicHandler) SetExtensionManager(manager ext.ManagerInterface) {
	h.manager = manager
}

// ProxyRequest handles the proxying of requests to third-party APIs.
func (h *dynamicHandler) ProxyRequest(c *gin.Context) {
	ctx := c.Request.Context()
	startTime := time.Now()

	// Find the matching route
	path := c.Request.URL.Path
	method := c.Request.Method

	// Extract route path from the full path (removing "/proxy" prefix)
	routePath := strings.TrimPrefix(path, "/proxy")

	// Find the route configuration
	route, err := h.s.Route.FindByPathAndMethod(ctx, routePath, method)
	if err != nil {
		logger.Errorf(ctx, "Failed to find route for path %s and method %s: %v", routePath, method, err)
		resp.Fail(c.Writer, resp.NotFound("Route not found"))
		return
	}

	// Get the associated endpoint
	endpoint, err := h.s.Endpoint.GetByID(ctx, route.EndpointID)
	if err != nil {
		logger.Errorf(ctx, "Failed to find endpoint for route %s: %v", route.ID, err)
		resp.Fail(c.Writer, resp.InternalServer("Endpoint not found"))
		return
	}

	// Create event data for tracking this request
	eventData := &service.ProxyEventData{
		Timestamp:   time.Now(),
		EndpointID:  endpoint.ID,
		EndpointURL: endpoint.BaseURL,
		RouteID:     route.ID,
		RoutePath:   routePath,
		Method:      method,
		Metadata:    make(map[string]any),
	}

	// Publish request received event if manager is available
	if h.manager != nil {
		h.s.Processor.PublishEvent(h.manager, service.EventRequestReceived, eventData)
	}

	// Construct target URL
	targetURL, err := url.Parse(endpoint.BaseURL)
	if err != nil {
		logger.Errorf(ctx, "Invalid endpoint URL %s: %v", endpoint.BaseURL, err)
		resp.Fail(c.Writer, resp.InternalServer("Invalid endpoint configuration"))
		h.handleRequestError(ctx, eventData, err)
		return
	}

	// Apply path parameters from route
	targetPath := route.TargetPath
	for _, param := range c.Params {
		targetPath = strings.Replace(targetPath, ":"+param.Key, param.Value, -1)
		// Add parameters to event metadata for easier tracking
		eventData.Metadata[param.Key] = param.Value
	}

	targetURL.Path = targetPath

	// Clone the request
	proxyReq, err := http.NewRequestWithContext(ctx, c.Request.Method, targetURL.String(), c.Request.Body)
	if err != nil {
		logger.Errorf(ctx, "Failed to create request: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to create request"))
		h.handleRequestError(ctx, eventData, err)
		return
	}

	// Copy headers from original request
	for name, values := range c.Request.Header {
		// Skip some headers
		if route.StripAuthHeader && (name == "Authorization" || name == "Cookie") {
			continue
		}
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// Add custom headers based on endpoint configuration
	if endpoint.AuthType != "None" {
		// Handle authentication based on endpoint configuration
		switch endpoint.AuthType {
		case "Basic":
			// Parse auth config and add Basic auth header
			// Implement as needed
		case "Bearer":
			// Parse auth config and add Bearer token
			// Implement as needed
		case "ApiKey":
			// Parse auth config and add API key
			// Implement as needed
		case "OAuth":
			// Handle OAuth auth
			// Implement as needed
		}
	}

	// Read and store the original request body for potential pre-processing
	var requestBody []byte
	if c.Request.Body != nil {
		requestBody, err = io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Errorf(ctx, "Failed to read request body: %v", err)
			resp.Fail(c.Writer, resp.InternalServer("Failed to read request body"))
			h.handleRequestError(ctx, eventData, err)
			return
		}

		// Close and reset the body
		c.Request.Body.Close()
		c.Request.Body = io.NopCloser(strings.NewReader(string(requestBody)))
	}

	// Apply input transformer if configured
	if validator.IsNotEmpty(route.InputTransformerID) {
		transformer, exists := h.transformerCache[types.ToString(route.InputTransformerID)]
		if exists && requestBody != nil {
			// Apply transformer
			transformedBody, err := transformer(requestBody)
			if err != nil {
				logger.Errorf(ctx, "Failed to transform request: %v", err)
				resp.Fail(c.Writer, resp.InternalServer("Failed to transform request"))
				h.handleRequestError(ctx, eventData, err)
				return
			}

			requestBody = transformedBody

			// Publish event for request transformation
			if h.manager != nil {
				h.s.Processor.PublishEvent(h.manager, service.EventRequestTransformed, eventData)
			}
		} else if !exists {
			logger.Warnf(ctx, "Input transformer %s not found in cache", route.InputTransformerID)
		}
	}

	// Pre-process the request body with the processor service
	if requestBody != nil {
		processedBody, err := h.s.Processor.PreProcess(ctx, endpoint, route, requestBody)
		if err != nil {
			logger.Errorf(ctx, "Failed to pre-process request: %v", err)
			resp.Fail(c.Writer, resp.InternalServer("Failed to pre-process request"))
			h.handleRequestError(ctx, eventData, err)
			return
		}

		// Replace request body with processed version if it changed
		if !bytes.Equal(requestBody, processedBody) {
			requestBody = processedBody
			proxyReq.Body = io.NopCloser(strings.NewReader(string(requestBody)))
			proxyReq.ContentLength = int64(len(requestBody))
			proxyReq.Header.Set("Content-Length", fmt.Sprintf("%d", len(requestBody)))

			// Publish event for request pre-processing
			if h.manager != nil {
				h.s.Processor.PublishEvent(h.manager, service.EventRequestPreProcessed, eventData)
			}
		}
	}

	// Publish event for request being sent
	if h.manager != nil {
		h.s.Processor.PublishEvent(h.manager, service.EventRequestSent, eventData)
	}

	// Execute the request, potentially through a circuit breaker
	var stdResp *http.Response
	var circuitErr error

	if endpoint.UseCircuitBreaker {
		if cb, exists := h.circuitBreakers[endpoint.ID]; exists {
			result, err := cb.Execute(func() (any, error) {
				return h.httpClient.Do(proxyReq)
			})

			if err != nil {
				circuitErr = err

				// If circuit breaker was tripped, publish an event
				if h.manager != nil && strings.Contains(err.Error(), "circuit breaker is open") {
					h.s.Processor.PublishEvent(h.manager, service.EventCircuitBreakerTripped, eventData)
				}
			} else {
				stdResp = result.(*http.Response)
			}
		} else {
			// Circuit breaker not found, execute directly
			stdResp, err = h.httpClient.Do(proxyReq)
		}
	} else {
		// Execute without circuit breaker
		stdResp, err = h.httpClient.Do(proxyReq)
	}

	// Handle error cases
	if circuitErr != nil {
		logger.Errorf(ctx, "Circuit breaker error: %v", circuitErr)
		resp.Fail(c.Writer, resp.InternalServer("Service unavailable: circuit breaker open"))

		eventData.Error = circuitErr.Error()
		eventData.StatusCode = http.StatusServiceUnavailable
		h.handleRequestError(ctx, eventData, circuitErr)

		// Log the failed request
		h.s.Log.Create(ctx, &structs.CreateLogBody{
			LogBody: structs.LogBody{
				EndpointID:     endpoint.ID,
				RouteID:        route.ID,
				RequestMethod:  c.Request.Method,
				RequestPath:    c.Request.URL.Path,
				RequestHeaders: c.Request.Header,
				RequestBody:    string(requestBody),
				StatusCode:     http.StatusServiceUnavailable,
				Error:          circuitErr.Error(),
				Duration:       int(time.Since(startTime).Milliseconds()),
				ClientIP:       c.ClientIP(),
				UserID:         "", // Fill from authenticated user if available
			},
		})
		return
	}

	if err != nil {
		logger.Errorf(ctx, "Request failed: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Request to third-party API failed"))

		eventData.Error = err.Error()
		eventData.StatusCode = http.StatusInternalServerError
		h.handleRequestError(ctx, eventData, err)

		// Log the failed request
		h.s.Log.Create(ctx, &structs.CreateLogBody{
			LogBody: structs.LogBody{
				EndpointID:     endpoint.ID,
				RouteID:        route.ID,
				RequestMethod:  c.Request.Method,
				RequestPath:    c.Request.URL.Path,
				RequestHeaders: c.Request.Header,
				RequestBody:    string(requestBody),
				StatusCode:     http.StatusInternalServerError,
				Error:          err.Error(),
				Duration:       int(time.Since(startTime).Milliseconds()),
				ClientIP:       c.ClientIP(),
				UserID:         "", // Fill from authenticated user if available
			},
		})
		return
	}

	defer stdResp.Body.Close()

	// Update event data with response info
	eventData.StatusCode = stdResp.StatusCode

	// Publish response received event
	if h.manager != nil {
		h.s.Processor.PublishEvent(h.manager, service.EventResponseReceived, eventData)
	}

	// Read the response body
	responseBody, err := io.ReadAll(stdResp.Body)
	if err != nil {
		logger.Errorf(ctx, "Failed to read response body: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to read response from third-party API"))

		eventData.Error = err.Error()
		h.handleRequestError(ctx, eventData, err)
		return
	}

	// Apply output transformer if configured
	if validator.IsNotEmpty(route.OutputTransformerID) {
		transformer, exists := h.transformerCache[types.ToString(route.OutputTransformerID)]
		if exists {
			transformedBody, err := transformer(responseBody)
			if err != nil {
				logger.Errorf(ctx, "Failed to transform response: %v", err)
				resp.Fail(c.Writer, resp.InternalServer("Failed to transform response"))

				eventData.Error = err.Error()
				h.handleRequestError(ctx, eventData, err)
				return
			}
			responseBody = transformedBody

			// Publish event for response transformation
			if h.manager != nil {
				h.s.Processor.PublishEvent(h.manager, service.EventResponseTransformed, eventData)
			}
		} else {
			logger.Warnf(ctx, "Output transformer %s not found in cache", route.OutputTransformerID)
		}
	}

	// Post-process the response body with the processor service
	processedResponseBody, err := h.s.Processor.PostProcess(ctx, endpoint, route, responseBody)
	if err != nil {
		logger.Errorf(ctx, "Failed to post-process response: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to post-process response"))

		eventData.Error = err.Error()
		h.handleRequestError(ctx, eventData, err)
		return
	}

	// Only publish event if the response was actually changed by post-processing
	if !bytes.Equal(responseBody, processedResponseBody) {
		responseBody = processedResponseBody

		// Publish event for response post-processing
		if h.manager != nil {
			h.s.Processor.PublishEvent(h.manager, service.EventResponsePostProcessed, eventData)
		}
	}

	// Copy response headers
	for name, values := range stdResp.Header {
		for _, value := range values {
			c.Writer.Header().Add(name, value)
		}
	}

	// Set status code and write response body
	c.Writer.WriteHeader(stdResp.StatusCode)
	c.Writer.Write(responseBody)

	// Update event data with final timing
	eventData.Duration = int(time.Since(startTime).Milliseconds())

	// Publish response sent event
	if h.manager != nil {
		h.s.Processor.PublishEvent(h.manager, service.EventResponseSent, eventData)
	}

	// Log the successful request if logging is enabled
	if endpoint.LogRequests || endpoint.LogResponses {
		var reqBody, respBody string

		if endpoint.LogRequests {
			reqBody = string(requestBody)
		}

		if endpoint.LogResponses {
			respBody = string(responseBody)
		}

		h.s.Log.Create(ctx, &structs.CreateLogBody{
			LogBody: structs.LogBody{
				EndpointID:      endpoint.ID,
				RouteID:         route.ID,
				RequestMethod:   c.Request.Method,
				RequestPath:     c.Request.URL.Path,
				RequestHeaders:  c.Request.Header,
				RequestBody:     reqBody,
				StatusCode:      stdResp.StatusCode,
				ResponseHeaders: stdResp.Header,
				ResponseBody:    respBody,
				Duration:        int(time.Since(startTime).Milliseconds()),
				ClientIP:        c.ClientIP(),
				UserID:          "", // Fill from authenticated user if available
			},
		})
	}
}

// handleRequestError handles errors and publishes appropriate events
func (h *dynamicHandler) handleRequestError(ctx context.Context, eventData *service.ProxyEventData, err error) {
	eventData.Error = err.Error()

	// Publish error event if manager is available
	if h.manager != nil {
		h.s.Processor.PublishEvent(h.manager, service.EventRequestError, eventData)
	}
}

// registerCircuitBreaker creates and registers a circuit breaker for an endpoint
func (h *dynamicHandler) registerCircuitBreaker(endpointID, endpointName string) {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        endpointName,
		MaxRequests: 100,
		Interval:    5 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Infof(context.Background(), "Circuit breaker %s state changed from %s to %s", name, from, to)

			// If circuit breaker state changes, publish an event
			if h.manager != nil {
				eventData := &service.ProxyEventData{
					Timestamp:  time.Now(),
					EndpointID: endpointID,
					Method:     "STATE_CHANGE",
					Metadata: map[string]any{
						"from_state": from.String(),
						"to_state":   to.String(),
					},
				}

				if to == gobreaker.StateOpen {
					h.s.Processor.PublishEvent(h.manager, service.EventCircuitBreakerTripped, eventData)
				} else if from == gobreaker.StateOpen {
					h.s.Processor.PublishEvent(h.manager, service.EventCircuitBreakerReset, eventData)
				}
			}
		},
	})
	h.circuitBreakers[endpointID] = cb
}

// RegisterDynamicRoutes registers dynamic routes based on configured proxy routes.
func (h *dynamicHandler) RegisterDynamicRoutes(r *gin.RouterGroup) {
	ctx := context.Background()

	// Get all active routes
	params := &structs.ListRouteParams{}
	routes, err := h.s.Route.List(ctx, params)
	if err != nil {
		logger.Errorf(ctx, "Failed to load proxy routes: %v", err)
		return
	}

	// Register circuit breakers for each endpoint
	endpoints, err := h.s.Endpoint.List(ctx, &structs.ListEndpointParams{})
	if err != nil {
		logger.Errorf(ctx, "Failed to load endpoints: %v", err)
		return
	}

	for _, endpoint := range endpoints.Items {
		if endpoint.UseCircuitBreaker {
			h.registerCircuitBreaker(endpoint.ID, endpoint.Name)
		}
	}

	// Load all transformers into cache
	transformers, err := h.s.Transformer.List(ctx, &structs.ListTransformerParams{})
	if err != nil {
		logger.Errorf(ctx, "Failed to load transformers: %v", err)
	} else {
		for _, transformer := range transformers.Items {
			if !transformer.Disabled {
				tf, err := h.s.Transformer.CompileTransformer(ctx, transformer.ID)
				if err != nil {
					logger.Errorf(ctx, "Failed to compile transformer %s: %v", transformer.ID, err)
					continue
				}
				h.transformerCache[transformer.ID] = tf
			}
		}
	}

	// Register dynamic routes
	for _, route := range routes.Items {
		if route.Disabled {
			continue
		}

		// Determine HTTP method and register appropriate handler
		method := strings.ToUpper(route.Method)
		path := route.PathPattern

		switch method {
		case "GET":
			r.GET(path, h.ProxyRequest)
		case "POST":
			r.POST(path, h.ProxyRequest)
		case "PUT":
			r.PUT(path, h.ProxyRequest)
		case "DELETE":
			r.DELETE(path, h.ProxyRequest)
		case "PATCH":
			r.PATCH(path, h.ProxyRequest)
		case "HEAD":
			r.HEAD(path, h.ProxyRequest)
		case "OPTIONS":
			r.OPTIONS(path, h.ProxyRequest)
		case "ANY", "*":
			r.Any(path, h.ProxyRequest)
		default:
			r.GET(path, h.ProxyRequest) // Default to GET if method is unrecognized
		}

		logger.Infof(ctx, "Registered dynamic proxy route: %s %s", method, path)
	}
}
