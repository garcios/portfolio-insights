// Package main is the entry point for the transaction-service.
package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/garcios/portfolio-insights/pkg/database"
	"github.com/garcios/portfolio-insights/pkg/logger"
	"github.com/garcios/portfolio-insights/pkg/middleware"
	"github.com/garcios/portfolio-insights/services/transaction-service/internal/config"
	"github.com/garcios/portfolio-insights/services/transaction-service/internal/handler/grpc"
	httpHandler "github.com/garcios/portfolio-insights/services/transaction-service/internal/handler/http"
	"github.com/garcios/portfolio-insights/services/transaction-service/internal/infrastructure"
	"github.com/garcios/portfolio-insights/services/transaction-service/internal/metrics"
	"github.com/garcios/portfolio-insights/services/transaction-service/internal/repository"
	"github.com/garcios/portfolio-insights/services/transaction-service/internal/usecase"
	pb "github.com/garcios/portfolio-insights/services/transaction-service/transaction"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	googleGrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.LoadConfig()

	l := logger.New()
	l.Info("Transaction Service starting...")

	// Connect to Database
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
			l.Error("failed to close database connection", "error", err)
		}
	}()

	l.Info("Successfully connected to PostgreSQL database")

	// Initialize Repository
	repo := repository.NewPostgresTransactionRepository(db)

	// Start background metrics updater
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if count, err := repo.Count(); err == nil {
				metrics.TotalTransactions.Set(float64(count))
			}
		}
	}()

	// Initialize Gateways
	userGateway, err := infrastructure.NewUserGateway(cfg)
	if err != nil {
		l.Error("failed to create user gateway", "error", err)
		os.Exit(1)
	}

	marketDataGateway, err := infrastructure.NewMarketDataGateway(cfg)
	if err != nil {
		l.Error("failed to create marketdata gateway", "error", err)
		os.Exit(1)
	}

	// Initialize Event Publisher
	eventPublisher, err := infrastructure.NewNATSEventPublisher(cfg)
	if err != nil {
		l.Error("failed to create NATS event publisher", "error", err)
		os.Exit(1)
	}

	// Initialize Usecase
	uc := usecase.NewTransactionUsecase(repo, userGateway, marketDataGateway, eventPublisher)
	csvUc := usecase.NewCSVUploadUsecase(repo, userGateway, marketDataGateway, eventPublisher)

	// Initialize gRPC Handler
	handler := grpc.NewTransactionHandler(uc)

	// Initialize HTTP Handler
	csvHandler := httpHandler.NewCSVUploadHandler(csvUc)

	// Start Metrics Server
	metricsPort := cfg.MetricsPort
	if metricsPort == "" {
		metricsPort = "9097"
	}
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		l.Info("Metrics server listening on :" + metricsPort)
		if err := http.ListenAndServe(":"+metricsPort, nil); err != nil {
			l.Error("failed to start metrics server", "error", err)
		}
	}()

	// Start HTTP Server for CSV Upload
	go func() {
		http.HandleFunc("/upload-csv", csvHandler.UploadCSV)
		l.Info("HTTP server listening on :" + cfg.HTTPPort)
		if err := http.ListenAndServe(":"+cfg.HTTPPort, nil); err != nil {
			l.Error("failed to start HTTP server", "error", err)
		}
	}()

	// Start gRPC Server
	port := cfg.Port
	if port == "" {
		port = "50053"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		l.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	// Register Metrics Middleware
	s := googleGrpc.NewServer(
		googleGrpc.UnaryInterceptor(middleware.MetricsUnaryInterceptor(metrics.RecordGrpcRequest)),
	)
	pb.RegisterTransactionServiceServer(s, handler)
	reflection.Register(s)

	l.Info("Transaction Service listening on port " + port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
