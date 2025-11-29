package auth

import (
	"context"
)

type contextKey string

const userIDKey contextKey = "user_id"

// UserIDFromContext retrieves the user ID from the context.
// Resolvers should use this to get the authenticated user.
func UserIDFromContext(ctx context.Context) (string, error) {
	if id, ok := ctx.Value(userIDKey).(string); ok {
		return id, nil
	}

	// For now, hardcode user ID until JWT is implemented
	// In production, extract from JWT token in context
	userID := "02b28ee7-9ba2-427a-b918-a3d8e2cc00dc"

	return userID, nil
}
