# Proxy Module

Proxy service for third-party APIs

## Features

### Endpoint Management

- Register and manage external API endpoints
- Configure authentication, timeout, and retry settings
- Support for HTTP/HTTPS, WebSocket, TCP, and UDP protocols

### Route Configuration

- Dynamic route management with path patterns and HTTP methods
- Path parameter extraction and forwarding
- Configurable request/response caching

### Request/Response Transformation

- Transform payloads with templates, JavaScript functions, or JSON mappings
- Content type transformation
- Header manipulation

### Security

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
- Performance metrics (request duration)

## API Documentation

API endpoints are documented with Swagger annotations. Key endpoints include:

- `/tbp/endpoints` - Manage API endpoints
- `/tbp/routes` - Configure proxy routes
- `/tbp/transformers` - Set up data transformers
- `/proxy/*` - Dynamic proxy routes for third-party APIs
- `/ws/*` - WebSocket proxy endpoints
