package main

import (
	"net"
	"os"

	"github.com/garcios/portfolio-insights/pkg/logger"
	"github.com/garcios/portfolio-insights/services/user-service/internal/handler/grpc"
	"github.com/garcios/portfolio-insights/services/user-service/internal/infrastructure"
	"github.com/garcios/portfolio-insights/services/user-service/internal/repository"
	"github.com/garcios/portfolio-insights/services/user-service/internal/usecase"
	pb "github.com/garcios/portfolio-insights/services/user-service/proto/user"
	googleGrpc "google.golang.org/grpc"
)

func main() {
	l := logger.New()

	// Infrastructure
	db := infrastructure.NewPostgresDB()

	// Repository
	repo := repository.NewUserRepository(db)

	// Usecase
	uc := usecase.NewUserUsecase(repo)

	// Handler
	h := grpc.NewUserHandler(uc)

	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		l.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	l.Info("User Service listening on port " + port)

	s := googleGrpc.NewServer()
	pb.RegisterUserServiceServer(s, h)
	if err := s.Serve(lis); err != nil {
		l.Error("failed to serve", "error", err)
	}
}
