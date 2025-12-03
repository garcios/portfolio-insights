package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwt"
)

// Config holds JWT middleware configuration
type Config struct {
	JWKSFetcher *JWKSFetcher
	Issuer      string
	Audience    string
	SkipPaths   []string // Paths that don't require authentication
}

// Middleware creates an HTTP middleware for JWT validation
func Middleware(config *Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if path should skip authentication
			for _, path := range config.SkipPaths {
				if r.URL.Path == path {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			// Check Bearer prefix
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Validate token
			authCtx, err := validateToken(r.Context(), tokenString, config)
			if err != nil {
				fmt.Println("***Invalid token: ", err)
				http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
				return
			}

			// Add auth context to request context
			ctx := WithAuthContext(r.Context(), authCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// validateToken validates a JWT token and returns the auth context
func validateToken(ctx context.Context, tokenString string, config *Config) (*AuthContext, error) {
	// Get JWKS
	keySet, err := config.JWKSFetcher.GetKeySet(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWKS: %w", err)
	}

	// Parse and validate token
	token, err := jwt.Parse(
		[]byte(tokenString),
		jwt.WithKeySet(keySet),
		jwt.WithValidate(true),
		jwt.WithIssuer(config.Issuer),
		jwt.WithAudience(config.Audience),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Check expiration
	if token.Expiration().Before(time.Now()) {
		return nil, fmt.Errorf("token has expired")
	}

	// Extract claims
	claims := make(map[string]interface{})
	for key, value := range token.PrivateClaims() {
		claims[key] = value
	}

	// Extract standard claims
	subject := token.Subject()
	if subject == "" {
		return nil, fmt.Errorf("token missing subject claim")
	}

	// Extract email from claims
	email, _ := claims["email"].(string)

	// Extract scopes
	var scopes []string
	if scopeClaim, ok := claims["scp"].([]interface{}); ok {
		for _, s := range scopeClaim {
			if scope, ok := s.(string); ok {
				scopes = append(scopes, scope)
			}
		}
	} else if scopeStr, ok := claims["scope"].(string); ok {
		// Handle space-separated scopes
		scopes = strings.Split(scopeStr, " ")
	}

	// Create auth context
	authCtx := &AuthContext{
		UserID: subject,
		Email:  email,
		Scopes: scopes,
		Claims: claims,
	}

	return authCtx, nil
}

// OptionalMiddleware creates middleware that allows unauthenticated requests
// but populates auth context if a valid token is present
func OptionalMiddleware(config *Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// No token, continue without auth context
				next.ServeHTTP(w, r)
				return
			}

			// Check Bearer prefix
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				// Invalid format, continue without auth context
				next.ServeHTTP(w, r)
				return
			}

			tokenString := parts[1]

			// Validate token
			authCtx, err := validateToken(r.Context(), tokenString, config)
			if err != nil {
				// Invalid token, continue without auth context
				fmt.Println("***Invalid token in optional middleware: ", err)
				next.ServeHTTP(w, r)
				return
			}

			// Add auth context to request context
			ctx := WithAuthContext(r.Context(), authCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GraphQLMiddleware creates a GraphQL-specific middleware
func GraphQLMiddleware(config *Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip introspection queries
			if r.URL.Path == "/query" && r.Method == "POST" {
				// For now, use optional middleware for GraphQL
				// This allows introspection and public queries
				OptionalMiddleware(config)(next).ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
