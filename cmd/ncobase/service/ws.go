package service

import (
	"context"
	"ncobase/common/log"
	"net/http"

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

// HandleWebSocket handles WebSocket connections.
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf(context.Background(), "Failed to set websocket upgrade: %+v", err)
		return
	}
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf(context.Background(), "Failed to close websocket: %+v", err)
		}
	}(conn)

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf(context.Background(), "Read message error: %+v", err)
			return
		}
		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Printf(context.Background(), "Write message error: %+v", err)
			return
		}
	}
}

// registerWebSocketRoutes registers WebSocket routes.
func registerWebSocketRoutes(e *gin.Engine) {
	e.GET("/ws", func(c *gin.Context) {
		HandleWebSocket(c.Writer, c.Request)
	})
}
