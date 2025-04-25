# Proxy Module

A flexible and powerful proxy service for third-party API integration with
advanced internal service interaction capabilities.

## Overview

The Proxy Module provides a robust API gateway that enables seamless integration
between your internal services and external third-party APIs. It supports
bidirectional data flow, transformations, pre/post-processing, and event-driven
communication.

## Key Features

### API Gateway Capabilities

- Dynamic route configuration and management
- Support for HTTP/HTTPS, WebSocket, TCP, and UDP protocols
- Path parameter extraction and forwarding
- Configurable request/response caching

### Data Transformation

- Transform payloads with templates, JavaScript functions, or JSON mappings
- Content type transformation
- Header manipulation

### Internal Service Integration

- Pre/post-processing hooks for request and response data
- Bidirectional data synchronization with internal services
- Custom hook registration for specific endpoints and routes

### Event System

- Event-driven communication between proxied services and internal systems
- Event publishing for all proxy operations (requests, responses, errors)
- Subscription to system-wide events for cross-module integration

### Security & Reliability

- Circuit breaker pattern implementation for fault tolerance
- Rate limiting capabilities
- Authentication forwarding control
- SSL validation options

### WebSocket Support

- Bidirectional WebSocket proxying
- Real-time message transformation

### Monitoring

- Request/response logging
- Error tracking
- Performance metrics

## Architecture

```text
┌─────────────────┐      ┌───────────────────┐      ┌─────────────────┐
│                 │      │                   │      │                 │
│  Client         │◄────►│  Proxy Module     │◄────►│  Third-Party    │
│  Applications   │      │                   │      │  APIs           │
│                 │      │                   │      │                 │
└─────────────────┘      └───────┬───────────┘      └─────────────────┘
                                 │
                                 ▼
                         ┌───────────────────┐
                         │                   │
                         │  Internal         │
                         │  Services         │
                         │                   │
                         └───────────────────┘
```

## Getting Started

### Configuration

The module supports the following configuration:

```json
{
  "proxy": {
    "cache_enabled": true,
    "cache_ttl": 300,
    "default_timeout": 30,
    "max_retry_count": 3,
    "circuit_breaker": {
      "enabled": true,
      "failure_threshold": 0.5,
      "reset_timeout": 30
    }
  }
}
```

### API Endpoints

#### Management API

- `GET /tbp/endpoints` - List all registered endpoints
- `POST /tbp/endpoints` - Register a new endpoint
- `GET /tbp/endpoints/:id` - Get endpoint details
- `PUT /tbp/endpoints/:id` - Update endpoint
- `DELETE /tbp/endpoints/:id` - Delete endpoint

- `GET /tbp/routes` - List all routes
- `POST /tbp/routes` - Create a new route
- `GET /tbp/routes/:id` - Get route details
- `PUT /tbp/routes/:id` - Update route
- `DELETE /tbp/routes/:id` - Delete route

- `GET /tbp/transformers` - List all transformers
- `POST /tbp/transformers` - Create a new transformer
- `GET /tbp/transformers/:id` - Get transformer details
- `PUT /tbp/transformers/:id` - Update transformer
- `DELETE /tbp/transformers/:id` - Delete transformer

#### Proxy Endpoints

- `/proxy/*` - Dynamic proxy routes for HTTP/HTTPS
- `/ws/*` - WebSocket proxy endpoints

## Integration Examples

### Creating a CRM Integration

```go
// Register a hook for CRM contacts endpoint
err := processor.RegisterHook(
salesforceEndpointID,
contactsRouteID,
preProcessContactHook,
postProcessContactHook,
)
```

### Payment Processing Integration

```go
// Subscribe to payment success events
manager.SubscribeEvent(EventResponseReceived, handlePaymentResponse)
```

### Message Synchronization with Collaboration Tools

```go
// Pre-process outgoing message to add context
func preProcessMessageHook(ctx context.Context, data []byte) ([]byte, error) {
// Enrich message with internal user and space data
return enrichedData, nil
}
```

## Event System

The module leverages the extension event system for seamless integration:

- **Request Events**: `proxy.request.received`, `proxy.request.preprocessed`,
  etc.
- **Response Events**: `proxy.response.received`, `proxy.response.transformed`,
  etc.
- **Error Events**: `proxy.request.error`, `proxy.circuit_breaker.tripped`, etc.

## Advanced Features

### Custom Transformers

Create powerful data transformers in three formats:

1. **Template-based** - Using Go templates
2. **Script-based** - Using JavaScript
3. **Mapping-based** - Using JSON mapping configs

### Circuit Breaking

The module includes circuit breaker functionality to handle failures gracefully:

```go
if endpoint.UseCircuitBreaker {
// Execute with circuit breaker
result, err := circuitBreaker.Execute(func () (any, error) {
return httpClient.Do(request)
})
}
```

## Best Practices

- Use transformers for data structure changes
- Use pre/post-processing hooks for integration with internal services
- Leverage the event system for cross-module communication
- Implement circuit breakers for critical endpoints
- Monitor request/response logs for troubleshooting
