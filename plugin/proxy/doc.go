// Package proxy provides API gateway and proxy functionality for the Ncobase platform.
//
// This package implements a flexible API gateway that can proxy requests to external
// services, transform requests and responses, and manage API routes dynamically.
// It supports both HTTP and WebSocket proxying with middleware capabilities.
//
// Key Features:
//   - Dynamic route management
//   - HTTP and WebSocket proxying
//   - Request/response transformation
//   - Header manipulation
//   - Load balancing and failover
//   - Rate limiting and throttling
//   - Authentication and authorization
//   - Request logging and monitoring
//
// Main Components:
//   - Handler: HTTP endpoints for proxy operations
//   - Service: Business logic for proxy management
//   - Repository: Data access layer for proxy configuration
//   - Transformer: Request/response transformation logic
//
// Proxy Capabilities:
//   - Route Management: Dynamic route creation and updates
//   - Transformers: Modify requests and responses in flight
//   - WebSocket: Bidirectional WebSocket proxying
//   - Middleware: Custom processing pipeline
//   - Caching: Response caching for performance
//
// Usage Example:
//
//	// Create a proxy route
//	route, err := proxyService.CreateRoute(ctx, &structs.CreateRouteInput{
//	    Path:       "/api/external",
//	    Target:     "https://api.example.com",
//	    Method:     "GET",
//	    StripPath:  true,
//	    Timeout:    30,
//	})
//	if err != nil {
//	    return err
//	}
//
//	// Add request transformer
//	transformer, err := proxyService.AddTransformer(ctx, route.ID, &structs.TransformerInput{
//	    Type:   "request",
//	    Action: "add_header",
//	    Config: map[string]interface{}{
//	        "name":  "X-API-Key",
//	        "value": "secret-key",
//	    },
//	})
//
//	// Create WebSocket proxy
//	wsRoute, err := proxyService.CreateWebSocketRoute(ctx, &structs.CreateWSRouteInput{
//	    Path:   "/ws/chat",
//	    Target: "wss://chat.example.com",
//	})
//
// Transformer Types:
//   - Request: Modify incoming requests (headers, body, query params)
//   - Response: Modify outgoing responses (headers, body, status)
//   - Authentication: Add authentication to proxied requests
//   - Rate Limiting: Control request rates
//
// The package provides a powerful API gateway solution suitable for microservices
// architectures, API aggregation, and external service integration.
package proxy
