package gateway

import (
	"context"

	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/entity"
)

// UserGateway defines the interface for interacting with the user service
type UserGateway interface {
	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, id string) (*entity.User, error)

	// CreateUser creates a new user
	CreateUser(ctx context.Context, email, username, password, firstName, lastName string) (*entity.User, error)

	// UpdateUser updates an existing user
	UpdateUser(ctx context.Context, id string, updates *entity.UserUpdate) (*entity.User, error)
}
