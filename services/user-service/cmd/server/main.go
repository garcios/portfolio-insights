// Package main is the entry point for the user-service
package main

import (
	"net"
	"net/http"
	"os"
	"time"

	"github.com/garcios/portfolio-insights/pkg/database"
	"github.com/garcios/portfolio-insights/pkg/logger"
	"github.com/garcios/portfolio-insights/pkg/middleware"
	"github.com/garcios/portfolio-insights/services/user-service/internal/config"
	"github.com/garcios/portfolio-insights/services/user-service/internal/handler/grpc"
	"github.com/garcios/portfolio-insights/services/user-service/internal/metrics"
	"github.com/garcios/portfolio-insights/services/user-service/internal/repository"
	"github.com/garcios/portfolio-insights/services/user-service/internal/usecase"
	pb "github.com/garcios/portfolio-insights/services/user-service/proto/user"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	googleGrpc "google.golang.org/grpc"
)

func main() {
	cfg := config.LoadConfig()

	l := logger.New()
	l.Info("User Service starting...")

	// Start Metrics Server
	metricsPort := cfg.MetricsPort
	if metricsPort == "" {
		metricsPort = "9096"
	}
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		l.Info("Metrics server listening on :" + metricsPort)
		if err := http.ListenAndServe(":"+metricsPort, nil); err != nil {
			l.Error("failed to start metrics server", "error", err)
		}
	}()

	// Infrastructure
	// Infrastructure
	dbCfg := database.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	}
	db, err := database.NewPostgresDB(dbCfg)
	if err != nil {
		l.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			l.Error("failed to close database", "error", err)
		}
	}()

	// Repository
	repo := repository.NewUserRepository(db)

	// Start background metrics updater
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if count, err := repo.Count(); err == nil {
				metrics.TotalUsers.Set(float64(count))
			}
		}
	}()

	// Usecase
	uc := usecase.NewUserUsecase(repo)

	// Handler
	h := grpc.NewUserHandler(uc)

	port := cfg.Port
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		l.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	l.Info("User Service listening on port " + port)

	// Register Metrics Middleware
	s := googleGrpc.NewServer(
		googleGrpc.UnaryInterceptor(middleware.MetricsUnaryInterceptor(metrics.RecordGrpcRequest)),
	)
	pb.RegisterUserServiceServer(s, h)
	if err := s.Serve(lis); err != nil {
		l.Error("failed to serve", "error", err)
	}
}
