package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"notification_service/internal/config"
	"notification_service/internal/grpc/notificationpb"
	"notification_service/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()

	lis, err := net.Listen("tcp", cfg.GRPCAddress())
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", cfg.GRPCAddress(), err)
	}

	grpcServer := grpc.NewServer()
	notificationpb.RegisterNotificationServiceServer(grpcServer, service.NewNotificationServer())

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	reflection.Register(grpcServer)

	go func() {
		log.Printf("notification grpc server started on %s", cfg.GRPCAddress())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("grpc server failed: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("shutting down notification grpc server...")
	grpcServer.GracefulStop()
}
