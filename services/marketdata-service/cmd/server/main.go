// Package main is the entry point for the marketdata service.
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
	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/config"
	marketdatahandler "github.com/garcios/portfolio-insights/services/marketdata-service/internal/handler/grpc"
	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/metrics"
	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/repository"
	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/usecase"
	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/worker"
	pb "github.com/garcios/portfolio-insights/services/marketdata-service/proto/marketdata"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.LoadConfig()

	l := logger.New()
	l.Info("Market Data Service starting...")

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
	ingestionWorker, err := worker.NewIngestionWorker(repo, cfg)
	if err != nil {
		l.Error("failed to create ingestion worker", "error", err)
		os.Exit(1)
	}

	// Initialize Price Ingestion Worker
	priceWorker, err := worker.NewPriceIngestionWorker(repo, cfg)
	if err != nil {
		l.Error("failed to create price worker", "error", err)
		os.Exit(1)
	}

	// Initialize Currency Ingestion Worker
	currencyWorker, err := worker.NewCurrencyIngestionWorker(repo, cfg)
	if err != nil {
		l.Error("failed to create currency worker", "error", err)
		os.Exit(1)
	}

	// Start Workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ingestionWorker.Start(ctx)
	priceWorker.Start(ctx)
	currencyWorker.Start(ctx)
	l.Info("Ingestion workers started (assets, prices, currency rates)")

	// Initialize Price Sync Worker
	priceSyncWorker, err := worker.NewEODHDPriceSyncWorker(repo, cfg)
	if err != nil {
		l.Error("failed to create price sync worker", "error", err)
		os.Exit(1)
	}

	// Start Price Sync Worker
	priceSyncWorker.Start(ctx)
	l.Info("Price sync worker started")

	// Initialize Currency Sync Worker
	currencySyncWorker, err := worker.NewEODHDCurrencySyncWorker(repo, cfg)
	if err != nil {
		l.Error("failed to create currency sync worker", "error", err)
		os.Exit(1)
	}
	currencySyncWorker.Start(ctx)
	l.Info("Currency sync worker started")

	// Start HTTP Server (Metrics + Admin)
	metricsPort := cfg.MetricsPort
	if metricsPort == "" {
		metricsPort = "9099"
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())

		http.HandleFunc("/sync/currencies", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			// Trigger in background
			go func() {
				l.Info("Manual currency sync triggered via HTTP")
				if err := currencySyncWorker.TriggerSync(context.Background()); err != nil {
					l.Error("Manual currency sync failed", "error", err)
				}
			}()

			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte("Currency sync triggered"))
		})

		l.Info("HTTP server listening on :" + metricsPort)
		if err := http.ListenAndServe(":"+metricsPort, nil); err != nil {
			l.Error("failed to start HTTP server", "error", err)
		}
	}()

	// Start gRPC Server
	port := cfg.Port
	if port == "" {
		port = "50054"
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
	handler := marketdatahandler.NewMarketDataHandler(uc)
	pb.RegisterMarketDataServiceServer(grpcServer, handler)

	go func() {
		l.Info("Market Data Service listening on port " + port)
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
