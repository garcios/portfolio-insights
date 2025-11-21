package main

import (
	"net"
	"os"

	"github.com/garcios/portfolio-insights/pkg/logger"
	pb "github.com/garcios/portfolio-insights/services/transaction-service/proto/transaction"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedTransactionServiceServer
}

func main() {
	l := logger.New()
	l.Info("Transaction Service starting...")

	port := os.Getenv("PORT")
	if port == "" {
		port = "50052"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		l.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	s := grpc.NewServer()
	pb.RegisterTransactionServiceServer(s, &server{})
	l.Info("Transaction Service listening on port " + port)
	if err := s.Serve(lis); err != nil {
		l.Error("failed to serve", "error", err)
	}
}
