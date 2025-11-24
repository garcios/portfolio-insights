package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/garcios/portfolio-insights/pkg/logger"
	marketdatahandler "github.com/garcios/portfolio-insights/services/marketdata-service/internal/handler/grpc"
	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/infrastructure"
	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/metrics"
	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/middleware"
	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/repository"
	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/usecase"
	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/worker"
	pb "github.com/garcios/portfolio-insights/services/marketdata-service/proto/marketdata"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	l := logger.New()
	l.Info("Market Data Service starting...")

	// Start Metrics Server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		l.Info("Metrics server listening on :9099")
		if err := http.ListenAndServe(":9099", nil); err != nil {
			l.Error("failed to start metrics server", "error", err)
		}
	}()

	// Connect to Database
	db, err := infrastructure.NewPostgresDB()
	if err != nil {
		l.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize Repository
	repo := repository.NewPostgresMarketDataRepository(db)

	// Start background metrics updater
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if count, err := repo.CountAssets(); err == nil {
				metrics.TotalAssets.Set(float64(count))
			}
			if count, err := repo.CountPrices(); err == nil {
				metrics.TotalPrices.Set(float64(count))
			}
		}
	}()

	// Initialize Usecase
	uc := usecase.NewMarketDataUsecase(repo)

	// Initialize Ingestion Worker (Assets)
	ingestionWorker, err := worker.NewIngestionWorker(repo)
	if err != nil {
		l.Error("failed to create ingestion worker", "error", err)
		os.Exit(1)
	}

	// Initialize Price Ingestion Worker
	priceWorker, err := worker.NewPriceIngestionWorker(repo)
	if err != nil {
		l.Error("failed to create price worker", "error", err)
		os.Exit(1)
	}

	// Start Workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ingestionWorker.Start(ctx)
	priceWorker.Start(ctx)
	l.Info("Ingestion workers started")

	// Start gRPC Server
	lis, err := net.Listen("tcp", ":50054")
	if err != nil {
		l.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	// Register Metrics Middleware
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.MetricsUnaryInterceptor()),
	)
	handler := marketdatahandler.NewMarketDataHandler(uc)
	pb.RegisterMarketDataServiceServer(grpcServer, handler)

	go func() {
		l.Info("Market Data Service listening on port 50054")
		if err := grpcServer.Serve(lis); err != nil {
			l.Error("failed to serve gRPC", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	l.Info("Shutting down Market Data Service...")
	grpcServer.GracefulStop()
}
