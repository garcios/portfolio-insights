package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/garcios/portfolio-insights/pkg/logger"
	portfoliohandler "github.com/garcios/portfolio-insights/services/portfolio-service/internal/handler/grpc"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/infrastructure"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/repository"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/usecase"
	pb "github.com/garcios/portfolio-insights/services/portfolio-service/proto/portfolio"
	"google.golang.org/grpc"
)

func main() {
	l := logger.New()
	l.Info("Portfolio Service starting...")

	// Connect to Database
	db, err := infrastructure.NewPostgresDB()
	if err != nil {
		l.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	l.Info("Successfully connected to PostgreSQL database")

	// Connect to Redis
	redisClient, err := infrastructure.NewRedisClient()
	if err != nil {
		l.Warn("failed to connect to Redis, caching will be disabled", "error", err)
		redisClient = nil
	} else {
		l.Info("Successfully connected to Redis")
	}

	// Initialize Price Cache
	var priceCache *infrastructure.PriceCache
	if redisClient != nil {
		priceCache = infrastructure.NewPriceCache(redisClient)
		l.Info("Price caching enabled")
	}

	// Initialize Repository
	repo := repository.NewPostgresHoldingRepository(db)

	// Initialize MarketData Gateway with cache
	marketDataGateway, err := infrastructure.NewMarketDataGateway(priceCache)
	if err != nil {
		l.Error("failed to create marketdata gateway", "error", err)
		os.Exit(1)
	}
	defer marketDataGateway.Close()

	// Initialize Usecase
	portfolioUsecase := usecase.NewPortfolioUsecase(repo, marketDataGateway)

	// Initialize gRPC Handler
	portfolioHandler := portfoliohandler.NewPortfolioHandler(portfolioUsecase)

	// Initialize NATS Subscriber
	subscriber, err := infrastructure.NewNATSSubscriber(repo, l)
	if err != nil {
		l.Error("failed to create NATS subscriber", "error", err)
		os.Exit(1)
	}
	defer subscriber.Stop()

	// Start subscribing to events
	if err := subscriber.Start(); err != nil {
		l.Error("failed to start NATS subscriber", "error", err)
		os.Exit(1)
	}

	// Start gRPC Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "50052"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		l.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterPortfolioServiceServer(grpcServer, portfolioHandler)

	go func() {
		l.Info("gRPC server listening on port " + port)
		if err := grpcServer.Serve(lis); err != nil {
			l.Error("failed to serve gRPC", "error", err)
			os.Exit(1)
		}
	}()

	l.Info("Portfolio Service is running and listening for events...")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	l.Info("Shutting down Portfolio Service...")
	grpcServer.GracefulStop()
}
