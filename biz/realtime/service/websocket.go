package service

import (
	"context"
	"encoding/json"
	"fmt"
	"ncobase/biz/realtime/data"
	"sync"
	"time"

	"github.com/ncobase/ncore/logging/logger"

	"github.com/gorilla/websocket"
)

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type    string `json:"type"`              // Message type (event, notification, etc)
	Channel string `json:"channel,omitempty"` // Target channel
	Data    any    `json:"data,omitempty"`    // Message payload
}

// Client represents a WebSocket client connection
type Client struct {
	ID            string
	UserID        string
	Conn          *websocket.Conn
	Send          chan []byte
	Subscriptions map[string]bool // Subscribed channels
	mu            sync.RWMutex
	lastPing      time.Time
}

// WebSocketService manages WebSocket connections and message broadcasting
type WebSocketService interface {
	RegisterClient(client *Client)
	UnregisterClient(client *Client)
	GetClient(clientID string) (*Client, bool)
	GetClientsByUser(userID string) []*Client
	GetClientsByChannel(channel string) []*Client

	BroadcastToChannel(channel string, message *WebSocketMessage) error
	BroadcastToUser(userID string, message *WebSocketMessage) error
	BroadcastToAll(message *WebSocketMessage) error

	SubscribeToChannel(clientID, channel string) error
	UnsubscribeFromChannel(clientID, channel string) error

	Cleanup()
	StartMaintenanceRoutine()
}

type webSocketService struct {
	clients    map[string]*Client            // All connected clients
	channels   map[string]map[string]*Client // Clients by channel
	users      map[string]map[string]*Client // Clients by user
	register   chan *Client
	unregister chan *Client
	broadcast  chan *WebSocketMessage
	mu         sync.RWMutex
	data       *data.Data
}

// NewWebSocketService creates a new WebSocket service
func NewWebSocketService(d *data.Data) WebSocketService {
	ws := &webSocketService{
		clients:    make(map[string]*Client),
		channels:   make(map[string]map[string]*Client),
		users:      make(map[string]map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *WebSocketMessage),
		data:       d,
	}

	go ws.run()
	go ws.StartMaintenanceRoutine()

	return ws
}

// run handles WebSocket events in a separate goroutine
func (ws *webSocketService) run() {
	for {
		select {
		case client := <-ws.register:
			ws.handleRegister(client)

		case client := <-ws.unregister:
			ws.handleUnregister(client)

		case message := <-ws.broadcast:
			ws.handleBroadcast(message)
		}
	}
}

// handleRegister registers a new client
func (ws *webSocketService) handleRegister(client *Client) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.clients[client.ID] = client

	// Register under user ID
	if client.UserID != "" {
		if _, exists := ws.users[client.UserID]; !exists {
			ws.users[client.UserID] = make(map[string]*Client)
		}
		ws.users[client.UserID][client.ID] = client
	}

	logger.Debugf(context.Background(), "Client %s registered (User: %s)", client.ID, client.UserID)
}

// handleUnregister unregisters a client
func (ws *webSocketService) handleUnregister(client *Client) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if _, ok := ws.clients[client.ID]; ok {
		// Remove from all channels
		for channel := range client.Subscriptions {
			delete(ws.channels[channel], client.ID)
			if len(ws.channels[channel]) == 0 {
				delete(ws.channels, channel)
			}
		}

		// Remove from user mapping
		if client.UserID != "" {
			if userClients, exists := ws.users[client.UserID]; exists {
				delete(userClients, client.ID)
				if len(userClients) == 0 {
					delete(ws.users, client.UserID)
				}
			}
		}

		// Close send channel and remove client
		close(client.Send)
		delete(ws.clients, client.ID)

		logger.Infof(context.Background(), "Client %s unregistered", client.ID)
	}
}

// handleBroadcast broadcasts a message to relevant clients
func (ws *webSocketService) handleBroadcast(msg *WebSocketMessage) {
	d, err := json.Marshal(msg)
	if err != nil {
		logger.Errorf(context.Background(), "Failed to marshal broadcast message: %v", err)
		return
	}

	ws.mu.RLock()
	defer ws.mu.RUnlock()

	// If channel specified, broadcast to channel subscribers
	if msg.Channel != "" {
		if subscribers, ok := ws.channels[msg.Channel]; ok {
			for _, client := range subscribers {
				select {
				case client.Send <- d:
				default:
					close(client.Send)
					ws.unregister <- client
				}
			}
		}
		return
	}

	// Otherwise broadcast to all clients
	for _, client := range ws.clients {
		select {
		case client.Send <- d:
		default:
			close(client.Send)
			ws.unregister <- client
		}
	}
}

// RegisterClient registers a new client
func (ws *webSocketService) RegisterClient(client *Client) {
	ws.register <- client
}

// UnregisterClient unregisters a client
func (ws *webSocketService) UnregisterClient(client *Client) {
	ws.unregister <- client
}

// GetClient gets a client by ID
func (ws *webSocketService) GetClient(clientID string) (*Client, bool) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	client, exists := ws.clients[clientID]
	return client, exists
}

// GetClientsByUser gets all clients for a user
func (ws *webSocketService) GetClientsByUser(userID string) []*Client {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	var clients []*Client
	if userClients, exists := ws.users[userID]; exists {
		for _, client := range userClients {
			clients = append(clients, client)
		}
	}
	return clients
}

// GetClientsByChannel gets all clients subscribed to a channel
func (ws *webSocketService) GetClientsByChannel(channel string) []*Client {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	var clients []*Client
	if channelClients, exists := ws.channels[channel]; exists {
		for _, client := range channelClients {
			clients = append(clients, client)
		}
	}
	return clients
}

// BroadcastToChannel broadcasts a message to a specific channel
func (ws *webSocketService) BroadcastToChannel(channel string, message *WebSocketMessage) error {
	if channel == "" {
		return fmt.Errorf("channel is required")
	}
	message.Channel = channel
	ws.broadcast <- message
	return nil
}

// BroadcastToUser broadcasts a message to a specific user
func (ws *webSocketService) BroadcastToUser(userID string, message *WebSocketMessage) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	ws.mu.RLock()
	userClients, exists := ws.users[userID]
	ws.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no clients found for user %s", userID)
	}

	d, err := json.Marshal(message)
	if err != nil {
		return err
	}

	for _, client := range userClients {
		select {
		case client.Send <- d:
		default:
			close(client.Send)
			ws.unregister <- client
		}
	}

	return nil
}

// BroadcastToAll broadcasts a message to all connected clients
func (ws *webSocketService) BroadcastToAll(message *WebSocketMessage) error {
	ws.broadcast <- message
	return nil
}

// SubscribeToChannel subscribes a client to a channel
func (ws *webSocketService) SubscribeToChannel(clientID, channel string) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	client, exists := ws.clients[clientID]
	if !exists {
		return fmt.Errorf("client %s not found", clientID)
	}

	// Add to channel map
	if _, exists := ws.channels[channel]; !exists {
		ws.channels[channel] = make(map[string]*Client)
	}
	ws.channels[channel][clientID] = client

	// Update client subscriptions
	client.mu.Lock()
	client.Subscriptions[channel] = true
	client.mu.Unlock()

	logger.Infof(context.Background(), "Client %s subscribed to channel %s", clientID, channel)
	return nil
}

// UnsubscribeFromChannel unsubscribes a client from a channel
func (ws *webSocketService) UnsubscribeFromChannel(clientID, channel string) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Remove from channel map
	if channelClients, exists := ws.channels[channel]; exists {
		delete(channelClients, clientID)
		if len(channelClients) == 0 {
			delete(ws.channels, channel)
		}
	}

	// Update client subscriptions
	if client, exists := ws.clients[clientID]; exists {
		client.mu.Lock()
		delete(client.Subscriptions, channel)
		client.mu.Unlock()
	}

	logger.Infof(context.Background(), "Client %s unsubscribed from channel %s", clientID, channel)
	return nil
}

// Cleanup performs cleanup of resources
func (ws *webSocketService) Cleanup() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	for _, client := range ws.clients {
		close(client.Send)
		client.Conn.Close()
	}

	ws.clients = make(map[string]*Client)
	ws.channels = make(map[string]map[string]*Client)
	ws.users = make(map[string]map[string]*Client)
}

// StartMaintenanceRoutine starts the maintenance routine
func (ws *webSocketService) StartMaintenanceRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			ws.performMaintenance()
		}
	}()
}

// performMaintenance performs periodic maintenance tasks
func (ws *webSocketService) performMaintenance() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	now := time.Now()
	for id, client := range ws.clients {
		// Check for stale connections
		if now.Sub(client.lastPing) > 2*time.Minute {
			logger.Infof(context.Background(), "Removing stale client %s", id)
			ws.unregister <- client
		}
	}
}

// Client methods

// Write writes a message to the client
func (c *Client) Write(messageType int, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn.WriteMessage(messageType, data)
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump(ws WebSocketService) {
	defer func() {
		ws.UnregisterClient(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512 * 1024) // 512KB max message size
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.mu.Lock()
		c.lastPing = time.Now()
		c.mu.Unlock()
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Errorf(context.Background(), "WebSocket read error: %v", err)
			}
			break
		}

		// Handle incoming message
		var msg WebSocketMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			logger.Errorf(context.Background(), "Failed to parse WebSocket message: %v", err)
			continue
		}

		// Process message based on type
		switch msg.Type {
		case "ping":
			c.mu.Lock()
			c.lastPing = time.Now()
			c.mu.Unlock()

		case "subscribe":
			if msg.Channel != "" {
				ws.SubscribeToChannel(c.ID, msg.Channel)
			}

		case "unsubscribe":
			if msg.Channel != "" {
				ws.UnsubscribeFromChannel(c.ID, msg.Channel)
			}
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
