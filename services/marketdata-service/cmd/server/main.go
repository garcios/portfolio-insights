package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/garcios/portfolio-insights/pkg/logger"
	marketdatahandler "github.com/garcios/portfolio-insights/services/marketdata-service/internal/handler/grpc"
	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/infrastructure"
	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/repository"
	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/usecase"
	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/worker"
	pb "github.com/garcios/portfolio-insights/services/marketdata-service/proto/marketdata"
	"google.golang.org/grpc"
)

func main() {
	l := logger.New()
	l.Info("Market Data Service starting...")

	// Connect to Database
	db, err := infrastructure.NewPostgresDB()
	if err != nil {
		l.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize Repository
	repo := repository.NewPostgresMarketDataRepository(db)

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

	grpcServer := grpc.NewServer()
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
