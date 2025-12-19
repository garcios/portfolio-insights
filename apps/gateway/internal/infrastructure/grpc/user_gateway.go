package grpc

import (
	"context"
	"fmt"

	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/entity"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/gateway"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/infrastructure/mapper"
	"github.com/garcios/portfolio-insights/pkg/resourcenames"
	userpb "github.com/garcios/portfolio-insights/services/user-service/user"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
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

	fmt.Printf("GetUser resp: %+v\n", resp)

	return mapper.ProtoToUserEntity(resp), nil
}

// CreateUser creates a new user using AIP-compliant User object
func (g *UserGRPCGateway) CreateUser(ctx context.Context, email, username, password, firstName, lastName string) (*entity.User, error) {
	req := &userpb.CreateUserRequest{
		User: &userpb.User{
			Email:     email,
			Username:  username,
			Password:  password,
			FirstName: firstName,
			LastName:  lastName,
		},
	}

	resp, err := g.client.CreateUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return mapper.ProtoToUserEntity(resp), nil
}

// UpdateUser updates an existing user using AIP-compliant Patch method
func (g *UserGRPCGateway) UpdateUser(ctx context.Context, id string, updates *entity.UserUpdate) (*entity.User, error) {
	userPb := &userpb.User{
		Name: resourcenames.UserName(id),
	}
	var paths []string

	if updates.Email != nil {
		userPb.Email = *updates.Email
		paths = append(paths, "email")
	}
	if updates.Username != nil {
		userPb.Username = *updates.Username
		paths = append(paths, "username")
	}
	if updates.FirstName != nil {
		userPb.FirstName = *updates.FirstName
		paths = append(paths, "first_name")
	}
	if updates.LastName != nil {
		userPb.LastName = *updates.LastName
		paths = append(paths, "last_name")
	}
	if updates.Preferences != nil {
		prefs, err := structpb.NewStruct(updates.Preferences)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal preferences: %w", err)
		}
		userPb.Preferences = prefs
		paths = append(paths, "preferences")
	}

	req := &userpb.UpdateUserRequest{
		User:       userPb,
		UpdateMask: &fieldmaskpb.FieldMask{Paths: paths},
	}

	resp, err := g.client.UpdateUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return mapper.ProtoToUserEntity(resp), nil
}
