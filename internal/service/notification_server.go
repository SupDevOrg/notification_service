package service

import (
	"context"
	"log"

	"notification_service/internal/grpc/notificationpb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NotificationServer struct {
	notificationpb.UnimplementedNotificationServiceServer
}

func NewNotificationServer() *NotificationServer {
	return &NotificationServer{}
}

func (s *NotificationServer) SendMessageNotification(
	ctx context.Context,
	req *notificationpb.SendMessageNotificationRequest,
) (*notificationpb.SendMessageNotificationResponse, error) {
	_ = ctx

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

	return &notificationpb.SendMessageNotificationResponse{
		Accepted: true,
	}, nil
}
