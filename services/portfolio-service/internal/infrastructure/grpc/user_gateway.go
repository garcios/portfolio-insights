// Package grpc provides the gRPC implementation of the infrastructure layer.
package grpc

import (
	"context"
	"fmt"

	"github.com/garcios/portfolio-insights/pkg/resourcenames"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
	userpb "github.com/garcios/portfolio-insights/services/user-service/user"
)

// UserGateway implements domain.UserGateway using gRPC.
type UserGateway struct {
	client userpb.UserServiceClient
}

// NewUserGateway creates a new UserGateway.
func NewUserGateway(client userpb.UserServiceClient) *UserGateway {
	return &UserGateway{
		client: client,
	}
}

// GetUserPreferences fetches the user's preferences from the user service.
func (g *UserGateway) GetUserPreferences(ctx context.Context, userID string) (*domain.UserPreferences, error) {
	// Construct resource name for user
	userName := resourcenames.UserName(userID)

	req := &userpb.GetUserRequest{
		Name: userName,
	}

	user, err := g.client.GetUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	// Extract preferences from user object (Structpb to Map to Struct)
	// Assuming preferences are stored in the Preferences field as a structpb.Struct
	prefs := &domain.UserPreferences{}

	if user.Preferences == nil {
		// Return empty preferences (or defaults) if nil
		return prefs, nil
	}

	prefMap := user.Preferences.AsMap()

	if val, ok := prefMap["default_currency"]; ok {
		if strVal, ok := val.(string); ok {
			prefs.DefaultCurrency = strVal
		}
	}

	if val, ok := prefMap["date_format"]; ok {
		if strVal, ok := val.(string); ok {
			prefs.DateFormat = strVal
		}
	}

	return prefs, nil
}
