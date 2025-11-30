package auth

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// Directive implements the @auth directive logic
func Directive(ctx context.Context, obj interface{}, next graphql.Resolver, scopes []string) (interface{}, error) {
	authCtx, err := AuthContextFromContext(ctx)
	if err != nil {
		return nil, &gqlerror.Error{
			Message: "Access denied: authentication required",
			Extensions: map[string]interface{}{
				"code": "UNAUTHENTICATED",
			},
		}
	}

	// If no specific scopes are required, just authentication is enough
	if len(scopes) == 0 {
		return next(ctx)
	}

	// Check if user has required scopes
	if !authCtx.HasAllScopes(scopes) {
		return nil, &gqlerror.Error{
			Message: fmt.Sprintf("Access denied: missing required scopes: %v", scopes),
			Extensions: map[string]interface{}{
				"code": "FORBIDDEN",
			},
		}
	}

	return next(ctx)
}
