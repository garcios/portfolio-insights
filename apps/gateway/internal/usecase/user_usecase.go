package usecase

import (
	"context"

	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/entity"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/gateway"
)

// UserUseCase handles user-related business logic
type UserUseCase struct {
	userGateway gateway.UserGateway
}

// NewUserUseCase creates a new UserUseCase
func NewUserUseCase(userGateway gateway.UserGateway) *UserUseCase {
	return &UserUseCase{
		userGateway: userGateway,
	}
}

// GetUser retrieves a user by ID
func (uc *UserUseCase) GetUser(ctx context.Context, id string) (*entity.User, error) {
	return uc.userGateway.GetUser(ctx, id)
}

// CreateUser creates a new user
func (uc *UserUseCase) CreateUser(ctx context.Context, email, username, password, firstName, lastName string) (*entity.User, error) {
	// Here you could add domain validation, business rules, etc.
	// For now, we delegate directly to the gateway
	return uc.userGateway.CreateUser(ctx, email, username, password, firstName, lastName)
}

// GetCurrentUser retrieves the currently authenticated user
// In production, this would extract user ID from auth context
func (uc *UserUseCase) GetCurrentUser(ctx context.Context) (*entity.User, error) {
	// Hardcoded for now - should come from auth context
	userID := "02b28ee7-9ba2-427a-b918-a3d8e2cc00dc"
	return uc.GetUser(ctx, userID)
}
