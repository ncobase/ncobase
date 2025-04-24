package handler

import (
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

	// Construct target URL
	targetURL, err := url.Parse(endpoint.BaseURL)
	if err != nil {
		logger.Errorf(ctx, "Invalid endpoint URL %s: %v", endpoint.BaseURL, err)
		resp.Fail(c.Writer, resp.InternalServer("Invalid endpoint configuration"))
		return
	}

	// Apply path parameters from route
	targetPath := route.TargetPath
	for _, param := range c.Params {
		targetPath = strings.Replace(targetPath, ":"+param.Key, param.Value, -1)
	}

	targetURL.Path = targetPath

	// Clone the request
	proxyReq, err := http.NewRequestWithContext(ctx, c.Request.Method, targetURL.String(), c.Request.Body)
	if err != nil {
		logger.Errorf(ctx, "Failed to create request: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to create request"))
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

	// Apply input transformer if configured
	var requestBody []byte
	if validator.IsNotEmpty(route.InputTransformerID) {
		transformer, exists := h.transformerCache[types.ToString(route.InputTransformerID)]
		if exists {
			// Read request body
			if c.Request.Body != nil {
				requestBody, err = io.ReadAll(c.Request.Body)
				if err != nil {
					logger.Errorf(ctx, "Failed to read request body: %v", err)
					resp.Fail(c.Writer, resp.InternalServer("Failed to read request body"))
					return
				}

				// Close and reset the body
				c.Request.Body.Close()
				c.Request.Body = io.NopCloser(strings.NewReader(string(requestBody)))

				// Apply transformer
				transformedBody, err := transformer(requestBody)
				if err != nil {
					logger.Errorf(ctx, "Failed to transform request: %v", err)
					resp.Fail(c.Writer, resp.InternalServer("Failed to transform request"))
					return
				}

				// Replace request body with transformed version
				proxyReq.Body = io.NopCloser(strings.NewReader(string(transformedBody)))
				proxyReq.ContentLength = int64(len(transformedBody))
				proxyReq.Header.Set("Content-Length", fmt.Sprintf("%d", len(transformedBody)))
			}
		} else {
			logger.Warnf(ctx, "Input transformer %s not found in cache", route.InputTransformerID)
		}
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

	// Read the response body
	responseBody, err := io.ReadAll(stdResp.Body)
	if err != nil {
		logger.Errorf(ctx, "Failed to read response body: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to read response from third-party API"))
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
				return
			}
			responseBody = transformedBody
		} else {
			logger.Warnf(ctx, "Output transformer %s not found in cache", route.OutputTransformerID)
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

// dynamicHandler represents the dynamic handler.
type dynamicHandler struct {
	s                *service.Service
	circuitBreakers  map[string]*gobreaker.CircuitBreaker
	httpClient       *http.Client
	transformerCache map[string]service.TransformerFunc
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
