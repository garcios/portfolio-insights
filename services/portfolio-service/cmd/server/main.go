// Package main is the entry point for the portfolio-service.
package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/garcios/portfolio-insights/pkg/database"
	"github.com/garcios/portfolio-insights/pkg/logger"
	"github.com/garcios/portfolio-insights/pkg/middleware"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/config"
	portfoliohandler "github.com/garcios/portfolio-insights/services/portfolio-service/internal/handler/grpc"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/infrastructure"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/metrics"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/repository"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/snapshotter"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/usecase"
	pb "github.com/garcios/portfolio-insights/services/portfolio-service/portfolio"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	transactionpb "github.com/garcios/portfolio-insights/services/transaction-service/transaction"
)

func main() {
	cfg := config.LoadConfig()

	l := logger.New()
	l.Info("Portfolio Service starting...")

	// Start Metrics Server
	metricsPort := cfg.MetricsPort
	if metricsPort == "" {
		metricsPort = "9098"
	}
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		l.Info("Metrics server listening on :" + metricsPort)
		if err := http.ListenAndServe(":"+metricsPort, nil); err != nil {
			l.Error("failed to start metrics server", "error", err)
		}
	}()

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

	// Connect to Redis
	redisClient, err := infrastructure.NewRedisClient(cfg)
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
		assetCache = infrastructure.NewAssetCache(redisClient, cfg)
		l.Info("Asset caching enabled")
	}

	// Initialize Historical Cache
	var historicalCache *infrastructure.HistoricalCache
	if redisClient != nil {
		historicalCache = infrastructure.NewHistoricalCache(redisClient)
		l.Info("Historical data caching enabled")
	}

	// Initialize Repositories
	holdingRepo := repository.NewPostgresHoldingRepository(db)
	historyRepo := repository.NewPostgresHistoryRepository(db)
	cashBalanceRepo := repository.NewPostgresCashBalanceRepository(db)

	// Start background metrics updater
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			count, err := holdingRepo.Count()
			if err == nil {
				metrics.HoldingsTotal.Set(float64(count))
			}
		}
	}()

	// Initialize Transaction Service Client
	transactionConn, err := grpc.NewClient(cfg.TransactionServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.Error("failed to connect to transaction service", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := transactionConn.Close(); err != nil {
			l.Error("failed to close transaction connection", "error", err)
		}
	}()
	transactionClient := transactionpb.NewTransactionServiceClient(transactionConn)

	// Initialize MarketData Gateway with cache
	marketDataGateway, err := infrastructure.NewMarketDataGateway(priceCache, assetCache, historicalCache, cfg)
	if err != nil {
		l.Error("failed to create marketdata gateway", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := marketDataGateway.Close(); err != nil {
			l.Error("failed to close market data gateway", "error", err)
		}
	}()

	// Initialize Cache Warmer
	if assetCache != nil && marketDataGateway != nil {
		cacheWarmer := infrastructure.NewCacheWarmer(marketDataGateway, assetCache, l)
		warmingInterval, err := time.ParseDuration(cfg.CacheWarmingInterval)
		if err != nil {
			l.Warn("Invalid cache warming interval, using default 6h", "error", err)
			warmingInterval = 6 * time.Hour
		}
		cacheWarmer.SchedulePeriodicWarming(warmingInterval)
		l.Info("Cache warmer initialized", "interval", warmingInterval)
	}

	// Initialize Detailed Snapshot Repository
	snapshotRepo := repository.NewPostgresSnapshotRepository(db)

	// Initialize NATS Subscriber
	natsSubscriber, err := infrastructure.NewNATSSubscriber(
		holdingRepo,
		cashBalanceRepo,
		snapshotRepo, // Added
		marketDataGateway,
		transactionClient,
		assetCache,
		l,
		cfg,
	)
	if err != nil {
		l.Error("failed to create NATS subscriber", "error", err)
		os.Exit(1)
	}

	if err := natsSubscriber.Start(); err != nil {
		l.Error("failed to start NATS subscriber", "error", err)
		os.Exit(1)
	}
	defer natsSubscriber.Stop()

	// Initialize Snapshot Manager
	snapshotConfig := snapshotter.WorkerConfig{
		PoolSize:   5,
		MaxRetries: 3,
		RateLimit:  10.0,
		Burst:      5,
	}
	snapshotManager := snapshotter.NewManager(snapshotConfig)

	// Create Usecase
	var portfolioUsecase usecase.PortfolioUsecase
	portfolioUsecase = usecase.NewPortfolioUsecase(
		holdingRepo,
		historyRepo,
		snapshotRepo,
		cashBalanceRepo,
		marketDataGateway,
		transactionClient,
		snapshotManager,
	)

	if redisClient != nil {
		wrappedRedis := database.NewRedisClientFromRaw(redisClient)
		portfolioUsecase = usecase.NewCachingPortfolioUsecase(portfolioUsecase, wrappedRedis, cfg.Caching)
		l.Info("Portfolio caching enabled")
	}

	// Initialize gRPC Handler
	portfolioHandler := portfoliohandler.NewPortfolioHandler(portfolioUsecase, historyRepo)

	// Start Snapshot Manager
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // This will cancel on main exit, which is fine

	snapshotManager.Start(ctx, func(userID string) error {
		// Logic to repair/refresh snapshot
		return portfolioUsecase.RefreshSnapshot(context.Background(), userID)
	})

	// Initialize Snapshot Worker
	historyWorker := infrastructure.NewPortfolioHistoryWorker(
		portfolioUsecase,
		historyRepo,
		transactionClient,
		l,
	)
	go historyWorker.Start()
	defer historyWorker.Stop()

	// Start gRPC Server
	port := cfg.Port
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
		grpc.UnaryInterceptor(middleware.MetricsUnaryInterceptor(metrics.RecordGrpcRequest)),
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
