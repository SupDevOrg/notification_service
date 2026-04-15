package websocket

import (
	"encoding/json"
	"log"
	"context"
	"fmt"
	"notification_service/internal/grpc/notificationpb"
)

type Message struct {
	RecipientIDs []uint64
	Content      []byte
}

type MessageNotification struct {
	Type            string   `json:"type"`
	MessageID       uint64   `json:"message_id"`
	ChatID          uint64   `json:"chat_id"`
	SenderID        uint64   `json:"sender_id"`
	SenderUsername  string   `json:"sender_username"`
	Content         string   `json:"content"`
	RecipientIDs    []uint64 `json:"recipient_ids"`
	CreatedAtUnixMs int64    `json:"created_at_unix_ms"`
}

type Hub struct {
	Clients    map[uint64]map[*Client]struct{}
	Broadcast  chan *Message
	Register   chan *Client
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[uint64]map[*Client]struct{}),
		Broadcast:  make(chan *Message, 100),  
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			if _, ok := h.Clients[client.UserID]; !ok {
				h.Clients[client.UserID] = make(map[*Client]struct{})
			}
			h.Clients[client.UserID][client] = struct{}{}

		case client := <-h.Unregister:
			if clients, ok := h.Clients[client.UserID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.Send)
				}
				if len(clients) == 0 {
					delete(h.Clients, client.UserID)
				}
			}

		case message := <-h.Broadcast:
			for _, recipientID := range message.RecipientIDs {
				clients, ok := h.Clients[recipientID]
				if !ok {
					continue
				}

				for client := range clients {
					select {
					case client.Send <- message.Content:
					default:
						log.Printf("notification websocket buffer is full: user=%d", recipientID)
					}
				}
			}
		}
	}
}

func (h *Hub) BroadcastMessageNotification(ctx context.Context, req *notificationpb.SendMessageNotificationRequest) error {
	payload, err := json.Marshal(MessageNotification{
		Type:            "message_notification",
		MessageID:       req.GetMessageId(),
		ChatID:          req.GetChatId(),
		SenderID:        req.GetSenderId(),
		SenderUsername:  req.GetSenderUsername(),
		Content:         req.GetContent(),
		RecipientIDs:    append([]uint64(nil), req.GetRecipientIds()...),
		CreatedAtUnixMs: req.GetCreatedAtUnixMs(),
	})
	if err != nil {
		return err
	}
	msg := &Message{
			RecipientIDs: append([]uint64(nil), req.GetRecipientIds()...),
			Content: payload,
		}
		select {
		case h.Broadcast <- msg:
			return nil
		case <-ctx.Done():
			return fmt.Errorf("broadcast timed out: %w", ctx.Err())
		}
}
