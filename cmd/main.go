package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notification_service/internal/config"
	"notification_service/internal/grpc/notificationpb"
	"notification_service/internal/grpc/usernotificationpb"
	"notification_service/internal/handlers"
	"notification_service/internal/service"
	"notification_service/internal/websocket"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()
	hub := websocket.NewHub()
	go hub.Run()

	lis, err := net.Listen("tcp", cfg.GRPCAddress())
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", cfg.GRPCAddress(), err)
	}

	grpcServer := grpc.NewServer()
	notificationpb.RegisterNotificationServiceServer(grpcServer, service.NewNotificationServer(hub))
	usernotificationpb.RegisterNotificationServiceServer(grpcServer, service.NewUserNotificationServer(hub))

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	reflection.Register(grpcServer)

	wsHandler := handlers.NewWebSocketHandler(hub)
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"OK"}`))
	})
	httpMux.HandleFunc("/api/v1/notification/ws", wsHandler.HandleWebSocket)
	httpMux.HandleFunc("/api/v1/notification/ws/", wsHandler.HandleWebSocket)

	httpServer := &http.Server{
		Addr:    cfg.HTTPAddress(),
		Handler: httpMux,
	}

	go func() {
		log.Printf("notification grpc server started on %s", cfg.GRPCAddress())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("grpc server failed: %v", err)
		}
	}()

	go func() {
		log.Printf("notification http server started on %s", cfg.HTTPAddress())
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server failed: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("shutting down notification servers...")
	grpcServer.GracefulStop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("http server shutdown error: %v", err)
	}
}
