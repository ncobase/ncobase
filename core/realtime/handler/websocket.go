package handler

import (
	"ncobase/realtime/service"
	"net/http"

	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils/uuid"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins
	},
}

type WebSocketHandler interface {
	HandleConnection(c *gin.Context)
}

type webSocketHandler struct {
	ws service.WebSocketService
}

func NewWebSocketHandler(ws service.WebSocketService) WebSocketHandler {
	return &webSocketHandler{ws: ws}
}

// HandleConnection handles WebSocket connections
//
// @Summary Handle WebSocket connection
// @Description Handles WebSocket connection
// @Tags rt
// @Router /rt/ws [get]
// @Security Bearer
func (h *webSocketHandler) HandleConnection(c *gin.Context) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Errorf(c, "Failed to upgrade connection: %v", err)
		return
	}

	// Create new client
	client := &service.Client{
		ID:            uuid.New().String(),
		UserID:        c.GetString("user_id"), // From auth middleware
		Conn:          conn,
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
	}

	// Register client
	h.ws.RegisterClient(client)

	// Start read/write pumps
	go client.ReadPump(h.ws)
	go client.WritePump()
}
