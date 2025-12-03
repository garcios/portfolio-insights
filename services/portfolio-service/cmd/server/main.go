package main

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/garcios/portfolio-insights/pkg/logger"
	portfoliohandler "github.com/garcios/portfolio-insights/services/portfolio-service/internal/handler/grpc"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/infrastructure"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/metrics"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/middleware"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/repository"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/usecase"
	pb "github.com/garcios/portfolio-insights/services/portfolio-service/proto/portfolio"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	l := logger.New()
	l.Info("Portfolio Service starting...")

	// Start Metrics Server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		l.Info("Metrics server listening on :9098")
		if err := http.ListenAndServe(":9098", nil); err != nil {
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

	// Initialize Asset Cache
	var assetCache *infrastructure.AssetCache
	if redisClient != nil {
		assetCache = infrastructure.NewAssetCache(redisClient)
		l.Info("Asset caching enabled")
	}

	// Initialize Repository
	repo := repository.NewPostgresHoldingRepository(db)
	historyRepo := repository.NewPostgresHistoryRepository(db)

	// Start background metrics updater
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			count, err := repo.Count()
			if err == nil {
				metrics.HoldingsTotal.Set(float64(count))
			}
		}
	}()

	// Initialize MarketData Gateway with cache
	marketDataGateway, err := infrastructure.NewMarketDataGateway(priceCache, assetCache)
	if err != nil {
		l.Error("failed to create marketdata gateway", "error", err)
		os.Exit(1)
	}
	defer marketDataGateway.Close()

	// Initialize Usecase
	portfolioUsecase := usecase.NewPortfolioUsecase(repo, marketDataGateway)

	// Initialize gRPC Handler
	portfolioHandler := portfoliohandler.NewPortfolioHandler(portfolioUsecase, historyRepo)

	// Initialize NATS Subscriber
	subscriber, err := infrastructure.NewNATSSubscriber(repo, marketDataGateway, assetCache, l)
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

	// Initialize and start cache warmer
	if assetCache != nil && marketDataGateway != nil {
		cacheWarmer := infrastructure.NewCacheWarmer(marketDataGateway, assetCache, l)

		// Schedule periodic cache warming every 6 hours
		// This keeps the cache fresh as assets don't change frequently
		warmingInterval := 6 * time.Hour
		cacheWarmer.SchedulePeriodicWarming(warmingInterval)

		l.Info("Asset cache warmer started", "interval", warmingInterval.String())
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

	// Register Metrics Middleware
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.MetricsUnaryInterceptor()),
	)
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
