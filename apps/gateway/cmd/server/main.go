package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/garcios/portfolio-insights/apps/gateway/graph"
	"github.com/garcios/portfolio-insights/apps/gateway/graph/generated"
	"github.com/garcios/portfolio-insights/pkg/logger"
	pb "github.com/garcios/portfolio-insights/services/portfolio-service/proto/portfolio"
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

	portfolioServiceAddr := os.Getenv("PORTFOLIO_SERVICE_ADDR")
	if portfolioServiceAddr == "" {
		portfolioServiceAddr = "localhost:50052"
	}

	conn, err := grpc.NewClient(portfolioServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.Error("Failed to connect to portfolio service", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	portfolioClient := pb.NewPortfolioServiceClient(conn)

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
		PortfolioClient: portfolioClient,
	}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	l.Info("connect to http://localhost:" + port + "/ for GraphQL playground")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
