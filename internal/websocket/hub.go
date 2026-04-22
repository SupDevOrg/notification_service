package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"notification_service/internal/grpc/notificationpb"
	"notification_service/internal/grpc/usernotificationpb"
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

type UserNotification struct {
	Type             string            `json:"type"`
	NotificationType string            `json:"notification_type"`
	RecipientID      uint64            `json:"recipient_id"`
	SenderID         uint64            `json:"sender_id"`
	Payload          map[string]string `json:"payload,omitempty"`
	CreatedAtUnixMs  int64             `json:"created_at_unix_ms"`
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

	return h.enqueueBroadcast(ctx, &Message{
		RecipientIDs: append([]uint64(nil), req.GetRecipientIds()...),
		Content:      payload,
	})
}

func (h *Hub) BroadcastUserNotification(ctx context.Context, req *usernotificationpb.SendNotificationRequest) error {
	payload, err := json.Marshal(UserNotification{
		Type:             "user_notification",
		NotificationType: userNotificationTypeName(req.GetType()),
		RecipientID:      uint64(req.GetRecipientId()),
		SenderID:         uint64(req.GetSenderId()),
		Payload:          copyPayload(req.GetPayload()),
		CreatedAtUnixMs:  req.GetCreatedAtUnixMs(),
	})
	if err != nil {
		return err
	}
	
	return h.enqueueBroadcast(ctx, &Message{
		RecipientIDs: []uint64{uint64(req.GetRecipientId())},
		Content:      payload,
	})
}

func (h *Hub) enqueueBroadcast(ctx context.Context, msg *Message) error {
	select {
	case h.Broadcast <- msg:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("broadcast timed out: %w", ctx.Err())
	}
}

func userNotificationTypeName(notificationType usernotificationpb.NotificationType) string {
	switch notificationType {
	case usernotificationpb.NotificationType_FRIEND_REQUEST_RECEIVED:
		return "friend_request_received"
	case usernotificationpb.NotificationType_FRIEND_REQUEST_ACCEPTED:
		return "friend_request_accepted"
	case usernotificationpb.NotificationType_FRIEND_REQUEST_REJECTED:
		return "friend_request_rejected"
	default:
		return "notification_type_unspecified"
	}
}

func copyPayload(payload map[string]string) map[string]string {
	if len(payload) == 0 {
		return nil
	}

	cloned := make(map[string]string, len(payload))
	for key, value := range payload {
		cloned[key] = value
	}

	return cloned
}
