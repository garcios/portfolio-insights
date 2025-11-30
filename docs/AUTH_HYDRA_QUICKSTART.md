# Ory Hydra OAuth2/OIDC Implementation - Quick Start Guide

## Overview

This guide provides step-by-step instructions to implement OAuth2/OIDC authentication for Portfolio Insights using Ory Hydra.

## Prerequisites

- Docker/Podman installed
- Go 1.21+ installed
- Node.js 18+ installed
- PostgreSQL running (for user database)

## Implementation Status

### ✅ Completed Components

1. **Docker Compose Configuration** (`deployments/docker-compose/docker-compose.hydra.yml`)
   - Hydra Public API (port 4444)
   - Hydra Admin API (port 4445)
   - PostgreSQL for Hydra
   - Login/Consent Provider (port 3001)

2. **Login/Consent Provider** (`apps/login-consent-provider/`)
   - Go application with Gin framework
   - Login flow handler
   - Consent flow handler
   - Logout handler
   - HTML templates with modern UI
   - Database integration for user authentication

3. **OAuth2 Client Registration Script** (`scripts/create-oauth-client.sh`)
   - Automated client creation
   - PKCE support
   - Proper scopes configuration

4. **Documentation** (`docs/AUTH_HYDRA_IMPLEMENTATION.md`)
   - Complete architecture overview
   - Flow diagrams
   - Security best practices

### 🚧 Remaining Implementation

The following components need to be implemented:

#### 1. Gateway Authentication Middleware

**File:** `apps/gateway/internal/auth/middleware.go`

**Required functionality:**
```go
// JWT validation middleware
// - Extract Bearer token from Authorization header
// - Fetch JWKS from Hydra
// - Validate JWT signature, issuer, audience, expiration
// - Extract claims (sub, email, scopes)
// - Inject AuthContext into Gin context
```

**File:** `apps/gateway/internal/auth/context.go`

**Required functionality:**
```go
type AuthContext struct {
    UserID string
    Email  string
    Scopes []string
    Claims map[string]interface{}
}

// Helper functions to get/set auth context
```

**File:** `apps/gateway/internal/auth/jwks.go`

**Required functionality:**
```go
// JWKS fetcher with caching
// - Fetch public keys from Hydra
// - Cache with TTL (1 hour)
// - Automatic refresh on cache miss
```

#### 2. GraphQL Authorization Directive

**File:** `apps/gateway/graph/schema.graphqls`

Add directive:
```graphql
directive @auth(requiredScopes: [String!]) on FIELD_DEFINITION
```

**File:** `apps/gateway/graph/directives.go`

Implement directive middleware:
```go
// Check if user has required scopes
// Return authentication error if not authorized
```

#### 3. Frontend OAuth2 Client

**Directory:** `apps/frontend/src/auth/`

**Required files:**

**`oauth.ts`** - OAuth2 flow implementation:
```typescript
// PKCE code generation
// Authorization URL builder
// Token exchange
// Token refresh
// Logout
```

**`AuthContext.tsx`** - React context for auth state:
```typescript
// Auth state management
// Token storage
// Auto-refresh logic
// Login/logout functions
```

**`ProtectedRoute.tsx`** - Route guard component:
```typescript
// Check if user is authenticated
// Redirect to login if not
```

**`useAuth.ts`** - Custom hook:
```typescript
// Access auth context
// Login/logout helpers
```

#### 4. Apollo Client Integration

**File:** `apps/frontend/src/utils/apolloClient.ts`

Update to include auth headers:
```typescript
// Add Authorization header with access token
// Handle 401 errors (token expired)
// Trigger token refresh
```

## Quick Start

### Step 1: Start Hydra Services

```bash
cd /Users/oscargarcia/Documents/workspace/portfolio-insights

# Start Hydra and Login/Consent Provider
podman-compose -f deployments/docker-compose/docker-compose.hydra.yml up -d

# Wait for services to be ready
sleep 10

# Create OAuth2 client
./scripts/create-oauth-client.sh
```

### Step 2: Implement Gateway Middleware

Create the following files in `apps/gateway/internal/auth/`:

1. `middleware.go` - JWT validation middleware
2. `context.go` - Auth context helpers
3. `jwks.go` - JWKS fetcher with caching

Update `apps/gateway/cmd/server/main.go`:
```go
// Add auth middleware to Gin router
router.Use(auth.JWTMiddleware(hydraPublicURL))
```

### Step 3: Implement Frontend OAuth2 Client

Create the following files in `apps/frontend/src/auth/`:

1. `oauth.ts` - OAuth2 flow functions
2. `AuthContext.tsx` - Auth state provider
3. `ProtectedRoute.tsx` - Route protection
4. `useAuth.ts` - Auth hook

Update `apps/frontend/src/App.tsx`:
```typescript
// Wrap app with AuthProvider
// Add login/logout buttons
// Handle OAuth callback route
```

### Step 4: Test the Flow

1. **Access the app:**
   ```
   http://localhost:5173
   ```

2. **Click "Login":**
   - Redirects to Hydra
   - Hydra redirects to Login Provider
   - Enter credentials
   - Grant consent
   - Redirects back with tokens

3. **Make GraphQL request:**
   - Token automatically included in Authorization header
   - Gateway validates token
   - Request succeeds if token is valid

## Environment Variables

### Hydra Services

```bash
# docker-compose.hydra.yml
URLS_SELF_ISSUER=http://localhost:4444
URLS_CONSENT=http://localhost:3001/consent
URLS_LOGIN=http://localhost:3001/login
```

### Login/Consent Provider

```bash
PORT=3001
HYDRA_ADMIN_URL=http://hydra-admin:4445
DATABASE_URL=postgres://postgres:postgres@postgres:5432/investments?sslmode=disable
SESSION_SECRET=changeThisToASecureRandomString
```

### Gateway

```bash
HYDRA_PUBLIC_URL=http://localhost:4444
JWKS_URL=http://localhost:4444/.well-known/jwks.json
JWT_ISSUER=http://localhost:4444
JWT_AUDIENCE=portfolio-insights-spa
```

### Frontend

```bash
VITE_HYDRA_AUTH_URL=http://localhost:4444/oauth2/auth
VITE_HYDRA_TOKEN_URL=http://localhost:4444/oauth2/token
VITE_HYDRA_LOGOUT_URL=http://localhost:4444/oauth2/sessions/logout
VITE_CLIENT_ID=portfolio-insights-spa
VITE_REDIRECT_URI=http://localhost:5173/auth/callback
```

## Testing

### Manual Testing

1. **Test Login Flow:**
   ```bash
   # Open browser
   open http://localhost:5173
   
   # Click login, should redirect to Hydra
   # Login with test credentials
   # Should redirect back with tokens
   ```

2. **Test Token Validation:**
   ```bash
   # Get access token from browser localStorage
   TOKEN="<access_token>"
   
   # Make GraphQL request
   curl -X POST http://localhost:8080/query \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"query": "{ me { id email } }"}'
   ```

3. **Test Token Refresh:**
   ```bash
   # Wait for token to expire (15 minutes)
   # App should automatically refresh token
   # GraphQL requests should continue to work
   ```

### Automated Testing

Create integration tests in `tests/auth/`:

```go
// Test login flow
// Test token validation
// Test scope enforcement
// Test token refresh
// Test logout
```

## Security Checklist

- [ ] HTTPS enabled in production
- [ ] Secure cookie flags set (HttpOnly, Secure, SameSite)
- [ ] PKCE required for all clients
- [ ] Short-lived access tokens (15 minutes)
- [ ] Refresh token rotation enabled
- [ ] JWKS caching with reasonable TTL
- [ ] CORS properly configured
- [ ] Rate limiting on auth endpoints
- [ ] Audit logging for auth events
- [ ] Secrets stored in environment variables (not in code)

## Troubleshooting

### "Invalid redirect URI"
- Check client configuration matches exactly
- Verify no trailing slashes

### "Token validation failed"
- Check JWKS endpoint is accessible
- Verify token hasn't expired
- Validate issuer and audience claims

### "CORS errors"
- Add Hydra URL to Gateway CORS whitelist
- Ensure credentials are included in requests

### "Login/Consent provider not responding"
- Check service is running: `podman ps`
- Check logs: `podman logs login-consent-provider`
- Verify database connection

## Next Steps

1. Implement Gateway middleware (highest priority)
2. Implement Frontend OAuth2 client
3. Add GraphQL authorization directives
4. Create integration tests
5. Add monitoring and logging
6. Implement token refresh logic
7. Add logout functionality
8. Create user profile page

## Resources

- [Ory Hydra Documentation](https://www.ory.sh/docs/hydra)
- [OAuth 2.0 Specification](https://oauth.net/2/)
- [OpenID Connect Specification](https://openid.net/connect/)
- [PKCE Specification](https://oauth.net/2/pkce/)

---

*Last Updated: 2025-11-30*
