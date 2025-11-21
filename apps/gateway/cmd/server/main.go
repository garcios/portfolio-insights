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
)

const defaultPort = "8080"

func main() {
	l := logger.New()
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	l.Info("connect to http://localhost:" + port + "/ for GraphQL playground")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
