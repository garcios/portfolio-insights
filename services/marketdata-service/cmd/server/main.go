package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/garcios/portfolio-insights/pkg/logger"
	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/worker"
)

func main() {
	l := logger.New()
	l.Info("Market Data Service starting...")

	// Connect to Database
	db, err := worker.ConnectDB()
	if err != nil {
		l.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize Ingestion Worker (Assets)
	ingestionWorker, err := worker.NewIngestionWorker(db)
	if err != nil {
		l.Error("failed to create ingestion worker", "error", err)
		os.Exit(1)
	}

	// Initialize Price Ingestion Worker
	priceWorker, err := worker.NewPriceIngestionWorker(db)
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

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	l.Info("Shutting down Market Data Service...")
}
