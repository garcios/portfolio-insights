package grpc

import (
	"context"
	"fmt"

	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/entity"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/gateway"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/infrastructure/mapper"
	"github.com/garcios/portfolio-insights/pkg/resourcenames"
	userpb "github.com/garcios/portfolio-insights/services/user-service/user"
)

// UserGRPCGateway implements the UserGateway interface using gRPC
type UserGRPCGateway struct {
	client userpb.UserServiceClient
}

// NewUserGRPCGateway creates a new UserGRPCGateway
func NewUserGRPCGateway(client userpb.UserServiceClient) gateway.UserGateway {
	return &UserGRPCGateway{
		client: client,
	}
}

// GetUser retrieves a user by ID using AIP-compliant resource name
func (g *UserGRPCGateway) GetUser(ctx context.Context, id string) (*entity.User, error) {
	req := &userpb.GetUserRequest{
		Name: resourcenames.UserName(id),
	}

	resp, err := g.client.GetUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from service: %w", err)
	}

	return mapper.ProtoToUserEntity(resp), nil
}

// CreateUser creates a new user using AIP-compliant User object
func (g *UserGRPCGateway) CreateUser(ctx context.Context, email, username, password string) (*entity.User, error) {
	req := &userpb.CreateUserRequest{
		User: &userpb.User{
			Email:    email,
			Username: username,
			Password: password,
		},
	}

	resp, err := g.client.CreateUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return entity.NewUser(resp.UserId, username, email), nil
}
