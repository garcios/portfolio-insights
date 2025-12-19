package domain

import "context"

// UserPreferences represents the user's preferences.
type UserPreferences struct {
	DefaultCurrency string
	DateFormat      string
}

// UserGateway defines the interface for interacting with the user service.
type UserGateway interface {
	GetUserPreferences(ctx context.Context, userID string) (*UserPreferences, error)
}
