package handlers

import (
	"log"
	"net/http"
	"strconv"

	ws "notification_service/internal/websocket"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketHandler struct {
	hub *ws.Hub
}

func NewWebSocketHandler(hub *ws.Hub) *WebSocketHandler {
	return &WebSocketHandler{hub: hub}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-Auth-User-ID")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "X-Auth-User-ID header is required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}

	client := &ws.Client{
		Hub:    h.hub,
		Socket: conn,
		Send:   make(chan []byte, 256),
		UserID: userID,
	}

	h.hub.Register <- client

	go client.Write()
	go client.Read()

	log.Printf("notification websocket connected: user=%d", userID)
}
