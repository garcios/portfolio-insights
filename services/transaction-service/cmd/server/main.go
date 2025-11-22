package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/garcios/portfolio-insights/pkg/logger"
	transactionhandler "github.com/garcios/portfolio-insights/services/transaction-service/internal/handler/grpc"
	"github.com/garcios/portfolio-insights/services/transaction-service/internal/infrastructure"
	"github.com/garcios/portfolio-insights/services/transaction-service/internal/repository"
	"github.com/garcios/portfolio-insights/services/transaction-service/internal/usecase"
	pb "github.com/garcios/portfolio-insights/services/transaction-service/proto/transaction"
	"google.golang.org/grpc"
)

func main() {
	l := logger.New()
	l.Info("Transaction Service starting...")

	// Connect to Database
	db, err := infrastructure.NewPostgresDB()
	if err != nil {
		l.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize Repository
	repo := repository.NewPostgresTransactionRepository(db)

	// Initialize User Gateway
	userGateway, err := infrastructure.NewUserGateway()
	if err != nil {
		l.Error("failed to create user gateway", "error", err)
		os.Exit(1)
	}

	// Initialize MarketData Gateway
	marketDataGateway, err := infrastructure.NewMarketDataGateway()
	if err != nil {
		l.Error("failed to create marketdata gateway", "error", err)
		os.Exit(1)
	}

	// Initialize Usecase
	uc := usecase.NewTransactionUsecase(repo, userGateway, marketDataGateway)

	port := os.Getenv("PORT")
	if port == "" {
		port = "50053"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		l.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	s := grpc.NewServer()
	handler := transactionhandler.NewTransactionHandler(uc)
	pb.RegisterTransactionServiceServer(s, handler)

	go func() {
		l.Info("Transaction Service listening on port " + port)
		if err := s.Serve(lis); err != nil {
			l.Error("failed to serve", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	l.Info("Shutting down Transaction Service...")
	s.GracefulStop()
}
