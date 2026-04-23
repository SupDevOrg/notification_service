package service

import (
	"context"
	"log"
	"time" 
	"notification_service/internal/grpc/usernotificationpb"
	"notification_service/internal/websocket"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserNotificationServer struct {
	usernotificationpb.UnimplementedNotificationServiceServer
	hub *websocket.Hub
}

func NewUserNotificationServer(hub *websocket.Hub) *UserNotificationServer {
	return &UserNotificationServer{hub: hub}
}

func (s *UserNotificationServer) SendNotification(
	ctx context.Context,
	req *usernotificationpb.SendNotificationRequest,
) (*usernotificationpb.SendNotificationResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if req.GetRecipientId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "recipient_id must be greater than 0")
	}
	if req.GetSenderId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "sender_id must be greater than 0")
	}
	if req.GetType() == usernotificationpb.NotificationType_NOTIFICATION_TYPE_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "type is required")
	}

	log.Printf(
		"user notification received: type=%s recipient_id=%d sender_id=%d payload=%v created_at_unix_ms=%d",
		req.GetType().String(),
		req.GetRecipientId(),
		req.GetSenderId(),
		req.GetPayload(),
		req.GetCreatedAtUnixMs(),
)

	if s.hub != nil {
		if err := s.hub.BroadcastUserNotification(ctx, req); err != nil {
			log.Printf("failed to broadcast user notification: %v", err)
			return nil, status.Error(codes.Internal, "failed to broadcast notification")
		}
	}

	return &usernotificationpb.SendNotificationResponse{
		Success: true,
		Message: "notification accepted",
	}, nil
}
