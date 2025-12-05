// Package main is the entry point for the gateway application.
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/garcios/portfolio-insights/apps/gateway/graph"
	"github.com/garcios/portfolio-insights/apps/gateway/graph/generated"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/auth"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/container"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/middleware"
	"github.com/garcios/portfolio-insights/pkg/logger"
	portfoliopb "github.com/garcios/portfolio-insights/services/portfolio-service/proto/portfolio"
	transactionpb "github.com/garcios/portfolio-insights/services/transaction-service/proto/transaction"
	userpb "github.com/garcios/portfolio-insights/services/user-service/proto/user"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	defer func() {
		if err := portfolioConn.Close(); err != nil {
			l.Error("Failed to close portfolio connection", "error", err)
		}
	}()

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
	defer func() {
		if err := userConn.Close(); err != nil {
			l.Error("Failed to close user connection", "error", err)
		}
	}()

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
	defer func() {
		if err := transactionConn.Close(); err != nil {
			l.Error("Failed to close transaction connection", "error", err)
		}
	}()

	transactionClient := transactionpb.NewTransactionServiceClient(transactionConn)

	// Transaction service HTTP URL
	transactionServiceHTTPAddr := os.Getenv("TRANSACTION_SERVICE_HTTP_ADDR")
	if transactionServiceHTTPAddr == "" {
		transactionServiceHTTPAddr = "http://localhost:8081"
	}

	// Initialize dependency injection container
	c := container.NewContainer(userClient, portfolioClient, transactionClient, transactionServiceHTTPAddr)

	// Initialize JWT authentication (optional for development)
	var authMiddleware func(http.Handler) http.Handler

	hydraPublicURL := os.Getenv("HYDRA_PUBLIC_URL")
	if hydraPublicURL != "" {
		// JWT authentication enabled
		jwksURL := os.Getenv("JWKS_URL")
		if jwksURL == "" {
			jwksURL = hydraPublicURL + "/.well-known/jwks.json"
		}

		jwtIssuer := os.Getenv("JWT_ISSUER")
		if jwtIssuer == "" {
			jwtIssuer = hydraPublicURL
		}

		jwtAudience := os.Getenv("JWT_AUDIENCE")
		if jwtAudience == "" {
			jwtAudience = "portfolio-insights-spa"
		}

		// Create JWKS fetcher
		jwksFetcher := auth.NewJWKSFetcher(jwksURL, 1*time.Hour)

		// Create auth config
		authConfig := &auth.Config{
			JWKSFetcher: jwksFetcher,
			Issuer:      jwtIssuer,
			Audience:    jwtAudience,
			SkipPaths:   []string{"/", "/health"},
		}

		// Use optional middleware for GraphQL (allows introspection)
		authMiddleware = auth.OptionalMiddleware(authConfig)
		l.Info("JWT authentication enabled", "issuer", jwtIssuer, "jwks_url", jwksURL)
	} else {
		// No authentication (development mode)
		authMiddleware = func(next http.Handler) http.Handler {
			return next
		}
		l.Info("JWT authentication disabled (development mode)")
	}

	// Create GraphQL server with clean architecture
	// Create GraphQL server with clean architecture
	config := generated.Config{
		Resolvers: &graph.Resolver{
			Container: c,
		},
	}
	config.Directives.Auth = auth.Directive

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(config))

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
	http.Handle("/query", corsMiddleware(authMiddleware(middleware.MetricsMiddleware(srv))))
	http.Handle("/metrics", promhttp.Handler())

	l.Info("connect to http://localhost:" + port + "/ for GraphQL playground")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
