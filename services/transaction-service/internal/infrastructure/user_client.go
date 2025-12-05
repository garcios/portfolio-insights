package infrastructure

import (
	"context"
	"fmt"
	"os"

	"github.com/garcios/portfolio-insights/services/transaction-service/internal/domain"
	userpb "github.com/garcios/portfolio-insights/services/user-service/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type userGateway struct {
	client userpb.UserServiceClient
}

// NewUserGateway creates a new user gateway.
func NewUserGateway() (domain.UserGateway, error) {
	host := os.Getenv("USER_SERVICE_HOST")
	port := os.Getenv("USER_SERVICE_PORT")

	if host == "" {
		host = "user-service"
	}
	if port == "" {
		port = "50051"
	}

	target := fmt.Sprintf("%s:%s", host, port)
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	client := userpb.NewUserServiceClient(conn)
	return &userGateway{client: client}, nil
}

func (g *userGateway) Exists(ctx context.Context, userID string) (bool, error) {
	_, err := g.client.GetUser(ctx, &userpb.GetUserRequest{Id: userID})
	if err != nil {
		// TODO: Check specific gRPC error code for NotFound
		return false, nil
	}
	return true, nil
}
