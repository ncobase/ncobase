package handler

import (
	"context"
	"ncobase/common/log"
	"ncobase/feature/socket/service"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins
	},
}

// WebSocketHandlerInterface represents the websocket handler interface.
type WebSocketHandlerInterface interface {
	Connect(w http.ResponseWriter, r *http.Request)
}

// websocketHandler represents the websocket handler.
type websocketHandler struct {
	s *service.Service
}

// NewWebSocketHandler creates a new websocket handler.
func NewWebSocketHandler(s *service.Service) WebSocketHandlerInterface {
	return &websocketHandler{
		s: s,
	}
}

// Connect handles WebSocket connections.
func (h *websocketHandler) Connect(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf(context.Background(), "Failed to set websocket upgrade: %+v", err)
		return
	}
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			log.Errorf(context.Background(), "Failed to close websocket: %+v", err)
		}
	}(conn)

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Errorf(context.Background(), "Read message error: %+v", err)
			return
		}
		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Errorf(context.Background(), "Write message error: %+v", err)
			return
		}
	}
}
