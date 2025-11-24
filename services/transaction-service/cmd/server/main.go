package main

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/garcios/portfolio-insights/pkg/logger"
	transactionhandler "github.com/garcios/portfolio-insights/services/transaction-service/internal/handler/grpc"
	httphandler "github.com/garcios/portfolio-insights/services/transaction-service/internal/handler/http"
	"github.com/garcios/portfolio-insights/services/transaction-service/internal/infrastructure"
	"github.com/garcios/portfolio-insights/services/transaction-service/internal/middleware"
	"github.com/garcios/portfolio-insights/services/transaction-service/internal/repository"
	"github.com/garcios/portfolio-insights/services/transaction-service/internal/usecase"
	pb "github.com/garcios/portfolio-insights/services/transaction-service/proto/transaction"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	// Initialize Event Publisher
	eventPublisher, err := infrastructure.NewNATSEventPublisher()
	if err != nil {
		l.Error("failed to create event publisher", "error", err)
		os.Exit(1)
	}

	// Initialize Usecases
	uc := usecase.NewTransactionUsecase(repo, userGateway, marketDataGateway, eventPublisher)
	csvUsecase := usecase.NewCSVUploadUsecase(repo, userGateway, marketDataGateway, eventPublisher)

	// Initialize HTTP handlers
	csvHandler := httphandler.NewCSVUploadHandler(csvUsecase)

	// Start HTTP server for CSV upload and metrics
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8081"
	}

	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "9097"
	}

	go func() {
		httpMux := http.NewServeMux()
		httpMux.HandleFunc("/upload-csv", csvHandler.UploadCSV)
		httpMux.Handle("/metrics", promhttp.Handler())
		httpMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		l.Info("HTTP server listening on port " + httpPort)
		if err := http.ListenAndServe(":"+httpPort, httpMux); err != nil {
			l.Error("HTTP server failed", "error", err)
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "50053"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		l.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	// Create gRPC server with metrics interceptor
	s := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.UnaryServerInterceptor("transaction")),
	)
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
