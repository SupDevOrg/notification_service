package service

import (
	"context"
	"log"
	"time"
	
	"notification_service/internal/grpc/notificationpb"
	"notification_service/internal/websocket"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NotificationServer struct {
	notificationpb.UnimplementedNotificationServiceServer
	hub *websocket.Hub
}

func NewNotificationServer(hub *websocket.Hub) *NotificationServer {
	return &NotificationServer{hub: hub}
}

func (s *NotificationServer) SendMessageNotification(
	ctx context.Context,
	req *notificationpb.SendMessageNotificationRequest,
) (*notificationpb.SendMessageNotificationResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if req.GetMessageId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "message_id is required")
	}
	if req.GetChatId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "chat_id is required")
	}
	if req.GetSenderId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "sender_id is required")
	}
	if len(req.GetRecipientIds()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "recipient_ids must not be empty")
	}

	log.Printf(
		"notification received: message_id=%d chat_id=%d sender_id=%d sender_username=%q recipients=%v content=%q created_at_unix_ms=%d",
		req.GetMessageId(),
		req.GetChatId(),
		req.GetSenderId(),
		req.GetSenderUsername(),
		req.GetRecipientIds(),
		req.GetContent(),
		req.GetCreatedAtUnixMs(),
	)

	if s.hub != nil {
		if err := s.hub.BroadcastMessageNotification(ctx, req); err != nil {
			log.Printf("failed to broadcast notification: %v", err)
			return nil, status.Error(codes.Internal, "failed to broadcast notification")
		}
	}

	return &notificationpb.SendMessageNotificationResponse{
		Accepted: true,
	}, nil
}
