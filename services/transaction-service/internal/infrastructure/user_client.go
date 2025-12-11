package infrastructure

import (
	"context"
	"fmt"

	"github.com/garcios/portfolio-insights/services/transaction-service/internal/config"
	"github.com/garcios/portfolio-insights/services/transaction-service/internal/domain"
	userpb "github.com/garcios/portfolio-insights/services/user-service/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type userGateway struct {
	client userpb.UserServiceClient
}

// NewUserGateway creates a new user gateway.
func NewUserGateway(cfg config.Config) (domain.UserGateway, error) {
	target := cfg.UserServiceAddr
	if target == "" {
		target = "localhost:50051"
	}

	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	client := userpb.NewUserServiceClient(conn)
	return &userGateway{client: client}, nil
}

func (g *userGateway) Exists(ctx context.Context, userID string) (bool, error) {
	// Use resource name format: users/{user}
	resourceName := fmt.Sprintf("users/%s", userID)
	_, err := g.client.GetUser(ctx, &userpb.GetUserRequest{Name: resourceName})
	if err != nil {
		// TODO: Check specific gRPC error code for NotFound
		return false, nil
	}
	return true, nil
}
