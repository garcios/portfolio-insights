package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/garcios/portfolio-insights/apps/gateway/graph"
	"github.com/garcios/portfolio-insights/apps/gateway/graph/generated"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/container"
	"github.com/garcios/portfolio-insights/pkg/logger"
	portfoliopb "github.com/garcios/portfolio-insights/services/portfolio-service/proto/portfolio"
	transactionpb "github.com/garcios/portfolio-insights/services/transaction-service/proto/transaction"
	userpb "github.com/garcios/portfolio-insights/services/user-service/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultPort = "8080"

func main() {
	l := logger.New()
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Connect to portfolio service
	portfolioServiceAddr := os.Getenv("PORTFOLIO_SERVICE_ADDR")
	if portfolioServiceAddr == "" {
		portfolioServiceAddr = "localhost:50052"
	}

	portfolioConn, err := grpc.NewClient(portfolioServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.Error("Failed to connect to portfolio service", "error", err)
		os.Exit(1)
	}
	defer portfolioConn.Close()

	portfolioClient := portfoliopb.NewPortfolioServiceClient(portfolioConn)

	// Connect to user service
	userServiceAddr := os.Getenv("USER_SERVICE_ADDR")
	if userServiceAddr == "" {
		userServiceAddr = "localhost:50051"
	}

	userConn, err := grpc.NewClient(userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.Error("Failed to connect to user service", "error", err)
		os.Exit(1)
	}
	defer userConn.Close()

	userClient := userpb.NewUserServiceClient(userConn)

	// Connect to transaction service
	transactionServiceAddr := os.Getenv("TRANSACTION_SERVICE_ADDR")
	if transactionServiceAddr == "" {
		transactionServiceAddr = "localhost:50053"
	}

	transactionConn, err := grpc.NewClient(transactionServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.Error("Failed to connect to transaction service", "error", err)
		os.Exit(1)
	}
	defer transactionConn.Close()

	transactionClient := transactionpb.NewTransactionServiceClient(transactionConn)

	// Transaction service HTTP URL
	transactionServiceHTTPAddr := os.Getenv("TRANSACTION_SERVICE_HTTP_ADDR")
	if transactionServiceHTTPAddr == "" {
		transactionServiceHTTPAddr = "http://localhost:8081"
	}

	// Initialize dependency injection container
	c := container.NewContainer(userClient, portfolioClient, transactionClient, transactionServiceHTTPAddr)

	// Create GraphQL server with clean architecture
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: &graph.Resolver{
			Container: c,
		},
	}))

	// CORS middleware
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow all origins for development
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

			if r.Method == "OPTIONS" {
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	http.Handle("/", corsMiddleware(playground.Handler("GraphQL playground", "/query")))
	http.Handle("/query", corsMiddleware(srv))

	l.Info("connect to http://localhost:" + port + "/ for GraphQL playground")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
