package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/garcios/portfolio-insights/pkg/logger"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/infrastructure"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/repository"
)

func main() {
	l := logger.New()
	l.Info("Portfolio Service starting...")

	// Initialize Repository
	repo := repository.NewInMemoryHoldingRepository()

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

	l.Info("Portfolio Service is running and listening for events...")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	l.Info("Shutting down Portfolio Service...")
}
