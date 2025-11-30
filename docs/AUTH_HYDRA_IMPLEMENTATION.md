# OAuth2/OIDC Integration with Ory Hydra

## Table of Contents
1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Setup Instructions](#setup-instructions)
4. [Implementation Details](#implementation-details)
5. [Testing](#testing)
6. [Security Considerations](#security-considerations)

## Overview

This document describes the complete OAuth2/OpenID Connect integration for the Portfolio Insights GraphQL Gateway using Ory Hydra as the authorization server.

### Components
- **Ory Hydra:** OAuth2/OIDC authorization server
- **Login/Consent Provider:** User-facing authentication UI
- **GraphQL Gateway:** Token validation and authorization
- **Frontend SPA:** OAuth2 client with PKCE flow

### Flow Diagram

```
┌─────────────┐                                    ┌──────────────┐
│   Browser   │                                    │   Hydra      │
│   (SPA)     │                                    │   Public     │
└──────┬──────┘                                    └───────┬──────┘
       │                                                   │
       │ 1. Redirect to /oauth2/auth (with PKCE)         │
       │──────────────────────────────────────────────────>│
       │                                                   │
       │                                            ┌──────▼──────┐
       │                                            │   Hydra     │
       │                                            │   Admin     │
       │                                            └──────┬──────┘
       │                                                   │
       │ 2. Redirect to Login Provider                    │
       │<──────────────────────────────────────────────────│
       │                                                   │
┌──────▼──────┐                                           │
│   Login/    │                                           │
│   Consent   │                                           │
│   Provider  │                                           │
└──────┬──────┘                                           │
       │                                                   │
       │ 3. Accept Login (POST to Hydra Admin)           │
       │──────────────────────────────────────────────────>│
       │                                                   │
       │ 4. Redirect to Consent                           │
       │<──────────────────────────────────────────────────│
       │                                                   │
       │ 5. Accept Consent (POST to Hydra Admin)         │
       │──────────────────────────────────────────────────>│
       │                                                   │
       │ 6. Redirect back with auth code                  │
       │<──────────────────────────────────────────────────│
       │                                                   │
       │ 7. Exchange code for tokens                      │
       │──────────────────────────────────────────────────>│
       │                                                   │
       │ 8. Access + ID tokens                            │
       │<──────────────────────────────────────────────────│
       │                                                   │
┌──────▼──────┐                                           │
│   GraphQL   │                                           │
│   Gateway   │                                           │
└──────┬──────┘                                           │
       │                                                   │
       │ 9. Validate JWT via JWKS                         │
       │──────────────────────────────────────────────────>│
       │                                                   │
       │ 10. JWKS response                                │
       │<──────────────────────────────────────────────────│
       │                                                   │
       │ 11. Execute GraphQL query                        │
       │                                                   │
```

## Architecture

### Authentication Flow

1. **User initiates login:** Frontend redirects to Hydra's `/oauth2/auth` endpoint
2. **Hydra redirects to Login Provider:** User enters credentials
3. **Login Provider accepts login:** Calls Hydra Admin API
4. **Hydra redirects to Consent Provider:** User grants permissions
5. **Consent Provider accepts consent:** Calls Hydra Admin API
6. **Hydra returns authorization code:** Redirects back to frontend
7. **Frontend exchanges code for tokens:** Calls Hydra's `/oauth2/token` endpoint
8. **Frontend calls GraphQL Gateway:** Includes access token in Authorization header
9. **Gateway validates token:** Fetches JWKS from Hydra and validates JWT
10. **Gateway executes query:** Injects user context into resolvers

### Token Validation

The Gateway validates access tokens by:
1. Extracting JWT from `Authorization: Bearer <token>` header
2. Fetching Hydra's JWKS from `/.well-known/jwks.json`
3. Validating JWT signature, issuer, audience, and expiration
4. Extracting claims (user ID, email, scopes)
5. Injecting auth context into GraphQL resolvers

## Setup Instructions

See the following files for detailed setup:
- `deployments/docker-compose/docker-compose.hydra.yml` - Hydra services
- `apps/login-consent-provider/` - Login/Consent application
- `apps/gateway/internal/auth/` - Gateway authentication middleware
- `apps/frontend/src/auth/` - Frontend OAuth2 client

## Implementation Details

### 1. Ory Hydra Setup

**Docker Compose Configuration:**
- Hydra Public (port 4444): OAuth2 endpoints for clients
- Hydra Admin (port 4445): Admin API for Login/Consent provider
- PostgreSQL: Hydra persistence
- Hydra migrations: Database schema initialization

**OAuth2 Client Registration:**
```bash
# Create OAuth2 client for SPA
docker exec hydra \
  hydra clients create \
    --endpoint http://localhost:4445 \
    --id portfolio-insights-spa \
    --secret "" \
    --grant-types authorization_code,refresh_token \
    --response-types code \
    --scope openid,offline,profile,email,portfolio:read,portfolio:write \
    --callbacks http://localhost:5173/auth/callback \
    --token-endpoint-auth-method none
```

### 2. Login/Consent Provider

**Technology:** Go with Gin framework

**Endpoints:**
- `GET /login` - Display login form
- `POST /login` - Process login and accept with Hydra
- `GET /consent` - Display consent screen
- `POST /consent` - Accept consent with Hydra

**Key Features:**
- Session management
- User authentication (database lookup)
- Hydra Admin API integration
- CSRF protection

### 3. Gateway Authentication Middleware

**Implementation:**
- Gin middleware for token validation
- JWKS caching with TTL
- Context population with user claims
- Error handling for invalid/expired tokens

**Auth Context:**
```go
type AuthContext struct {
    UserID    string
    Email     string
    Scopes    []string
    Claims    map[string]interface{}
}
```

### 4. GraphQL Authorization

**Custom Directive:**
```graphql
directive @auth(requiredScopes: [String!]) on FIELD_DEFINITION

type Query {
  portfolio(id: ID!): Portfolio @auth(requiredScopes: ["portfolio:read"])
}
```

**Resolver Middleware:**
- Checks auth context presence
- Validates required scopes
- Returns authentication errors

### 5. Frontend Integration

**OAuth2 Flow:**
1. Generate PKCE code verifier and challenge
2. Redirect to Hydra with challenge
3. Handle callback with authorization code
- [x] **Extend GraphQL schema with authorization directives**
    - [x] Add `@auth` directive to schema
    - [x] Implement directive logic in Gateway
    - [x] Register directive in server setup
4. Exchange code for tokens (with verifier)
5. Store tokens securely
6. Include access token in GraphQL requests

**Token Storage:**
- Access token: Memory or sessionStorage
- Refresh token: HttpOnly cookie (if using BFF pattern)
- ID token: Memory for user info display

## Testing

### Manual Testing

1. **Start services:**
   ```bash
   make podman-up
   ```

2. **Create OAuth2 client:**
   ```bash
   ./scripts/create-oauth-client.sh
   ```

3. **Access frontend:**
   ```
   http://localhost:5173
   ```

4. **Login flow:**
   - Click "Login"
   - Enter credentials
   - Grant consent
   - Verify redirect with tokens

5. **GraphQL request:**
   ```bash
   curl -X POST http://localhost:8080/query \
     -H "Authorization: Bearer <access_token>" \
     -H "Content-Type: application/json" \
     -d '{"query": "{ me { id email } }"}'
   ```

### Automated Testing

See `tests/auth/` for integration tests covering:
- Login flow
- Token validation
- Scope enforcement
- Token refresh
- Error scenarios

## Security Considerations

### Best Practices

1. **PKCE Required:** Prevents authorization code interception
2. **State Parameter:** Prevents CSRF attacks
3. **Nonce in ID Token:** Prevents replay attacks
4. **Short-lived Access Tokens:** 15-minute expiration
5. **Refresh Token Rotation:** New refresh token on each use
6. **HTTPS Only:** All production traffic over TLS
7. **Secure Cookie Flags:** HttpOnly, Secure, SameSite=Strict
8. **JWKS Caching:** Reduce load on Hydra, 1-hour TTL

### Token Validation

Gateway validates:
- **Signature:** Using Hydra's public key from JWKS
- **Issuer (iss):** Must match Hydra's URL
- **Audience (aud):** Must include gateway client ID
- **Expiration (exp):** Token must not be expired
- **Not Before (nbf):** Token must be valid now
- **Subject (sub):** User ID must be present

### Scope Management

Scopes define permissions:
- `openid`: Required for OIDC
- `offline`: Enables refresh tokens
- `profile`: User profile information
- `email`: User email address
- `portfolio:read`: Read portfolio data
- `portfolio:write`: Modify portfolio data
- `transactions:read`: Read transactions
- `transactions:write`: Create/modify transactions

## Troubleshooting

### Common Issues

1. **"Invalid redirect URI"**
   - Ensure client callback URL matches exactly
   - Check for trailing slashes

2. **"Token validation failed"**
   - Verify JWKS endpoint is accessible
   - Check token expiration
   - Validate issuer and audience claims

3. **"Consent screen not showing"**
   - Check Login/Consent provider is running
   - Verify `URLS_CONSENT` and `URLS_LOGIN` in Hydra config

4. **"CORS errors"**
   - Add Hydra public URL to Gateway CORS whitelist
   - Ensure credentials are included in requests

## References

- [Ory Hydra Documentation](https://www.ory.sh/docs/hydra)
- [OAuth 2.0 RFC 6749](https://tools.ietf.org/html/rfc6749)
- [OpenID Connect Core 1.0](https://openid.net/specs/openid-connect-core-1_0.html)
- [PKCE RFC 7636](https://tools.ietf.org/html/rfc7636)
- [JWT RFC 7519](https://tools.ietf.org/html/rfc7519)

---

*Last Updated: 2025-11-30*
