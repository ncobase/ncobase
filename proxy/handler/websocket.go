package handler

import (
	"context"
	"ncobase/proxy/service"
	"ncobase/proxy/structs"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation/validator"
)

// WebSocketHandlerInterface is the interface for the WebSocket handler.
type WebSocketHandlerInterface interface {
	RegisterWebSocketRoutes(r *gin.RouterGroup)
	HandleWebSocket(c *gin.Context)
}

// webSocketHandler represents the WebSocket handler.
type webSocketHandler struct {
	s                 *service.Service
	upgrader          websocket.Upgrader
	activeConnections sync.Map // map[string][]*websocket.Conn
	transformerCache  map[string]service.TransformerFunc
}

// NewWebSocketHandler creates a new WebSocket handler.
func NewWebSocketHandler(svc *service.Service) WebSocketHandlerInterface {
	return &webSocketHandler{
		s: svc,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin check
				return true
			},
		},
		transformerCache: make(map[string]service.TransformerFunc),
	}
}

// RegisterWebSocketRoutes registers WebSocket routes based on configured endpoints.
func (h *webSocketHandler) RegisterWebSocketRoutes(r *gin.RouterGroup) {
	ctx := context.Background()

	// Get all active endpoints with WebSocket protocol
	params := &structs.ListEndpointParams{
		Protocol: "WS,WSS", // Filter to only websocket protocols
	}
	endpoints, err := h.s.Endpoint.List(ctx, params)
	if err != nil {
		logger.Errorf(ctx, "Failed to load WebSocket endpoints: %v", err)
		return
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

	// Register routes for WebSocket endpoints
	for _, endpoint := range endpoints.Items {
		if endpoint.Disabled {
			continue
		}

		// Get routes for this endpoint
		routeParams := &structs.ListRouteParams{
			EndpointID: endpoint.ID,
		}
		routes, err := h.s.Route.List(ctx, routeParams)
		if err != nil {
			logger.Errorf(ctx, "Failed to load routes for endpoint %s: %v", endpoint.ID, err)
			continue
		}

		// Register each WebSocket route
		for _, route := range routes.Items {
			if route.Disabled {
				continue
			}

			path := route.PathPattern
			r.GET(path, h.HandleWebSocket)
			logger.Infof(ctx, "Registered WebSocket route: %s", path)
		}
	}
}

// HandleWebSocket handles WebSocket connections.
func (h *webSocketHandler) HandleWebSocket(c *gin.Context) {
	ctx := c.Request.Context()

	// Find the matching route
	path := c.Request.URL.Path

	// Extract route path from the full path (removing "/ws" prefix)
	routePath := path
	if path[:3] == "/ws" {
		routePath = path[3:]
	}

	// Find the route configuration
	route, err := h.s.Route.FindByPathAndMethod(ctx, routePath, "GET") // WebSockets use GET
	if err != nil {
		logger.Errorf(ctx, "Failed to find WebSocket route for path %s: %v", routePath, err)
		c.String(http.StatusNotFound, "WebSocket route not found")
		return
	}

	// Get the associated endpoint
	endpoint, err := h.s.Endpoint.GetByID(ctx, route.EndpointID)
	if err != nil {
		logger.Errorf(ctx, "Failed to find endpoint for route %s: %v", route.ID, err)
		c.String(http.StatusInternalServerError, "Endpoint not found")
		return
	}

	// Verify endpoint is WebSocket
	if endpoint.Protocol != "WS" && endpoint.Protocol != "WSS" {
		logger.Errorf(ctx, "Endpoint %s is not a WebSocket endpoint", endpoint.ID)
		c.String(http.StatusBadRequest, "Not a WebSocket endpoint")
		return
	}

	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Errorf(ctx, "Failed to upgrade to WebSocket: %v", err)
		return
	}

	// Store the connection
	// connID := fmt.Sprintf("%s_%s", endpoint.ID, time.Now().Format("20060102150405"))
	connections, _ := h.activeConnections.LoadOrStore(endpoint.ID, make([]*websocket.Conn, 0))
	connectionsList := connections.([]*websocket.Conn)
	connectionsList = append(connectionsList, conn)
	h.activeConnections.Store(endpoint.ID, connectionsList)

	// Ensure connection is closed and removed when done
	defer func() {
		conn.Close()
		h.removeConnection(endpoint.ID, conn)
	}()

	// Connect to target WebSocket
	targetURL, err := url.Parse(endpoint.BaseURL)
	if err != nil {
		logger.Errorf(ctx, "Invalid endpoint URL %s: %v", endpoint.BaseURL, err)
		conn.WriteMessage(websocket.TextMessage, []byte("Error: Invalid endpoint configuration"))
		return
	}

	// Apply path parameters from route
	targetPath := route.TargetPath
	for _, param := range c.Params {
		targetPath = strings.Replace(targetPath, ":"+param.Key, param.Value, -1)
	}

	targetURL.Path = targetPath

	// Determine the protocol
	var wsProtocol string
	if endpoint.Protocol == "WSS" {
		wsProtocol = "wss"
	} else {
		wsProtocol = "ws"
	}

	// Set the scheme to ws or wss
	targetURL.Scheme = wsProtocol

	// Create header for target connection
	header := http.Header{}
	for k, v := range c.Request.Header {
		if k != "Upgrade" && k != "Connection" && k != "Sec-Websocket-Key" &&
			k != "Sec-Websocket-Version" && k != "Sec-Websocket-Extensions" && k != "Sec-Websocket-Protocol" {
			header[k] = v
		}
	}

	// Add authentication if configured
	if endpoint.AuthType != "None" {
		// Handle authentication based on endpoint configuration
		switch endpoint.AuthType {
		case "Basic", "Bearer", "ApiKey", "OAuth":
			// Implement as needed
		}
	}

	// Connect to target WebSocket
	targetConn, _, err := websocket.DefaultDialer.Dial(targetURL.String(), header)
	if err != nil {
		logger.Errorf(ctx, "Failed to connect to target WebSocket %s: %v", targetURL.String(), err)
		conn.WriteMessage(websocket.TextMessage, []byte("Error: Failed to connect to target WebSocket"))
		return
	}
	defer targetConn.Close()

	// Create a context with cancel for goroutines
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	// Bidirectional communication
	go h.proxyClientToServer(ctxWithCancel, conn, targetConn, route)
	h.proxyServerToClient(ctxWithCancel, targetConn, conn, route)
}

// proxyClientToServer forwards messages from client to server
func (h *webSocketHandler) proxyClientToServer(ctx context.Context, clientConn, serverConn *websocket.Conn, route *structs.ReadRoute) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Read message from client
			messageType, message, err := clientConn.ReadMessage()
			if err != nil {
				logger.Errorf(ctx, "Error reading from client: %v", err)
				return
			}

			// Apply input transformer if configured
			if validator.IsNotEmpty(route.InputTransformerID) {
				transformer, exists := h.transformerCache[types.ToString(route.InputTransformerID)]
				if exists {
					transformedMessage, err := transformer(message)
					if err != nil {
						logger.Errorf(ctx, "Failed to transform client message: %v", err)
					} else {
						message = transformedMessage
					}
				}
			}

			// Send message to server
			if err := serverConn.WriteMessage(messageType, message); err != nil {
				logger.Errorf(ctx, "Error writing to server: %v", err)
				return
			}
		}
	}
}

// proxyServerToClient forwards messages from server to client
func (h *webSocketHandler) proxyServerToClient(ctx context.Context, serverConn, clientConn *websocket.Conn, route *structs.ReadRoute) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Read message from server
			messageType, message, err := serverConn.ReadMessage()
			if err != nil {
				logger.Errorf(ctx, "Error reading from server: %v", err)
				return
			}

			// Apply output transformer if configured
			if validator.IsNotEmpty(route.OutputTransformerID) {
				transformer, exists := h.transformerCache[types.ToString(route.OutputTransformerID)]
				if exists {
					transformedMessage, err := transformer(message)
					if err != nil {
						logger.Errorf(ctx, "Failed to transform server message: %v", err)
					} else {
						message = transformedMessage
					}
				}
			}

			// Send message to client
			if err := clientConn.WriteMessage(messageType, message); err != nil {
				logger.Errorf(ctx, "Error writing to client: %v", err)
				return
			}
		}
	}
}

// removeConnection removes a connection from the active connections list
func (h *webSocketHandler) removeConnection(endpointID string, conn *websocket.Conn) {
	connectionsObj, exists := h.activeConnections.Load(endpointID)
	if !exists {
		return
	}

	connections := connectionsObj.([]*websocket.Conn)
	for i, c := range connections {
		if c == conn {
			// Remove the connection from the slice
			connections = append(connections[:i], connections[i+1:]...)
			h.activeConnections.Store(endpointID, connections)
			return
		}
	}
}
