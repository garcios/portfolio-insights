package auth

import (
	"context"
	"errors"
)

// AuthContext holds authenticated user information
type AuthContext struct {
	UserID string
	Email  string
	Scopes []string
	Claims map[string]interface{}
}

type contextKey string

const authContextKey contextKey = "auth_context"

// WithAuthContext adds auth context to the context
func WithAuthContext(ctx context.Context, authCtx *AuthContext) context.Context {
	return context.WithValue(ctx, authContextKey, authCtx)
}

// AuthContextFromContext retrieves auth context from the context
func AuthContextFromContext(ctx context.Context) (*AuthContext, error) {
	authCtx, ok := ctx.Value(authContextKey).(*AuthContext)
	if !ok || authCtx == nil {
		return nil, errors.New("no auth context found")
	}
	return authCtx, nil
}

// UserIDFromContext retrieves user ID from the context
func UserIDFromContext(ctx context.Context) (string, error) {
	authCtx, err := AuthContextFromContext(ctx)
	if err != nil {
		return "", err
	}
	if authCtx.UserID == "" {
		return "", errors.New("user ID not found in auth context")
	}
	return authCtx.UserID, nil
}

// HasScope checks if the auth context has a specific scope
func (a *AuthContext) HasScope(scope string) bool {
	for _, s := range a.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// HasAnyScope checks if the auth context has any of the specified scopes
func (a *AuthContext) HasAnyScope(scopes []string) bool {
	for _, required := range scopes {
		if a.HasScope(required) {
			return true
		}
	}
	return false
}

// HasAllScopes checks if the auth context has all of the specified scopes
func (a *AuthContext) HasAllScopes(scopes []string) bool {
	for _, required := range scopes {
		if !a.HasScope(required) {
			return false
		}
	}
	return true
}
