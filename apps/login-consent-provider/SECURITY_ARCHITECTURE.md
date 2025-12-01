# Security Architecture: Service-to-Service Authentication

## Overview

The `login-consent-provider` has been refactored to use a **service-to-service architecture** for user authentication instead of direct database access. This improves security, maintainability, and follows microservices best practices.

## Previous Architecture (Insecure)

```
┌─────────────────────────┐
│ Login-Consent-Provider  │
│                         │
│  - Direct SQL queries   │
│  - Password validation  │
│  - User data access     │
└────────────┬────────────┘
             │
             │ Direct DB Connection
             ▼
   ┌─────────────────┐
   │   PostgreSQL    │
   │ customers.users │
   └─────────────────┘
```

### Problems with Previous Approach:

1. **Security Risks**:
   - Multiple services with direct database credentials
   - Increased attack surface (credentials in multiple places)
   - Password hashing logic duplicated across services

2. **Tight Coupling**:
   - Login provider tightly coupled to database schema
   - Schema changes require updates in multiple services

3. **Violation of Separation of Concerns**:
   - Authentication logic scattered across services
   - No single source of truth for user operations

4. **Difficult to Audit**:
   - User access from multiple points
   - Harder to track who accessed what

## New Architecture (Secure)

```
┌─────────────────────────┐
│ Login-Consent-Provider  │
│                         │
│  - gRPC client          │
│  - No DB access         │
│  - No credentials       │
└────────────┬────────────┘
             │
             │ gRPC (TLS in production)
             ▼
   ┌─────────────────────┐
   │   User-Service      │
   │                     │
   │  - VerifyUser RPC   │
   │  - GetUser RPC      │
   │  - Password logic   │
   └──────────┬──────────┘
              │
              │ DB Connection
              ▼
    ┌─────────────────┐
    │   PostgreSQL    │
    │ customers.users │
    └─────────────────┘
```

### Benefits of New Approach:

1. **Enhanced Security**:
   - ✅ Only user-service has database credentials
   - ✅ Reduced attack surface
   - ✅ Centralized password validation
   - ✅ Easier to implement rate limiting and monitoring
   - ✅ Can add mTLS for service-to-service auth

2. **Better Separation of Concerns**:
   - ✅ User-service owns all user data operations
   - ✅ Login-consent-provider focuses on OAuth flows
   - ✅ Single responsibility principle

3. **Improved Maintainability**:
   - ✅ Schema changes only affect user-service
   - ✅ Password hashing logic in one place
   - ✅ Easier to test (mock gRPC client)

4. **Scalability**:
   - ✅ User-service can be scaled independently
   - ✅ Can add caching at service level
   - ✅ Better load distribution

5. **Audit Trail**:
   - ✅ All user access goes through user-service
   - ✅ Centralized logging and monitoring
   - ✅ Easier compliance reporting

## Implementation Details

### New gRPC Methods in User-Service

```protobuf
service UserService {
  rpc GetUser (GetUserRequest) returns (GetUserResponse);
  rpc CreateUser (CreateUserRequest) returns (CreateUserResponse);
  rpc VerifyUser (VerifyUserRequest) returns (VerifyUserResponse);  // NEW
}

message VerifyUserRequest {
  string email = 1;
  string password = 2;
}

message VerifyUserResponse {
  bool valid = 1;
  string id = 2;
  string username = 3;
  string email = 4;
}
```

### Login-Consent-Provider Changes

**Before:**
```go
func (app *App) authenticateUser(ctx context.Context, email, password string) (*User, error) {
    var user User
    query := `SELECT id, email, password_hash, username FROM customers.users WHERE email = $1`
    err := app.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Username)
    // ... password verification with bcrypt
}
```

**After:**
```go
func (app *App) authenticateUser(ctx context.Context, email, password string) (*User, error) {
    return app.userClient.VerifyUser(ctx, email, password)
}
```

### Configuration Changes

**Environment Variables:**

Before:
```bash
DATABASE_URL=postgres://garcios:Password123@postgres:5432/portfolio?sslmode=disable
```

After:
```bash
USER_SERVICE_ADDR=user-service:50051
```

## Security Best Practices Implemented

### 1. Principle of Least Privilege
- Login-consent-provider no longer has database credentials
- Only user-service has direct database access

### 2. Defense in Depth
- Multiple layers of security:
  - Network isolation (service mesh in production)
  - gRPC authentication (can add mTLS)
  - Service-level authorization

### 3. Single Source of Truth
- User-service is the only authority for user data
- Consistent password validation logic

### 4. Testability
- Easy to mock UserServiceClient in tests
- No need for database fixtures in login-consent-provider tests

## Production Recommendations

### 1. Enable mTLS for gRPC
```go
// Production: Use TLS credentials
creds, err := credentials.NewClientTLSFromFile("cert.pem", "")
conn, err := grpc.Dial(userServiceAddr, grpc.WithTransportCredentials(creds))
```

### 2. Add Service Authentication
- Implement service-to-service authentication tokens
- Use Istio/Linkerd for automatic mTLS

### 3. Rate Limiting
- Add rate limiting in user-service for VerifyUser calls
- Prevent brute force attacks

### 4. Monitoring and Alerting
- Monitor failed authentication attempts
- Alert on unusual patterns
- Track service-to-service call latency

### 5. Circuit Breaker
```go
// Add circuit breaker for resilience
import "github.com/sony/gobreaker"

cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
    Name:        "user-service",
    MaxRequests: 3,
    Timeout:     time.Second * 60,
})
```

## Migration Guide

### Step 1: Deploy Updated User-Service
```bash
# Regenerate protobuf
make proto-gen

# Test user-service
cd services/user-service
go test ./...

# Deploy
make podman-build-user-service
```

### Step 2: Update Login-Consent-Provider
```bash
cd apps/login-consent-provider
go mod tidy
go test ./...
```

### Step 3: Update Docker Compose
```bash
# Update environment variables in docker-compose.hydra.yml
# Remove DATABASE_URL
# Add USER_SERVICE_ADDR=user-service:50051
```

### Step 4: Deploy
```bash
make hydra-build
make hydra-up
```

### Step 5: Verify
```bash
# Check logs
make hydra-logs

# Test login flow
# Navigate to http://localhost:5173 and test authentication
```

## Testing

### Unit Tests
```bash
# Test user-service
cd services/user-service
go test ./internal/handler/grpc -v

# Test login-consent-provider
cd apps/login-consent-provider
go test -v
```

### Integration Tests
```bash
# Start all services
make podman-up
make hydra-up

# Test OAuth flow
./scripts/test-oauth-flow.sh
```

## Rollback Plan

If issues arise, you can temporarily rollback:

1. Revert docker-compose.hydra.yml to use DATABASE_URL
2. Revert handlers.go to use direct DB access
3. Restart services

However, the new architecture is recommended for production use.

## Performance Considerations

### Latency
- gRPC call adds ~1-5ms latency vs direct DB
- Acceptable for authentication flows
- Can be optimized with connection pooling

### Caching
- Consider caching user data in login-consent-provider
- Use Redis for session-based caching
- Implement cache invalidation strategy

## Compliance and Audit

### Benefits for Compliance
- ✅ Centralized access control
- ✅ Better audit logging
- ✅ Easier to implement GDPR data access requests
- ✅ Clear data ownership boundaries

### Audit Logging
```go
// In user-service
log.Info("User verification attempt",
    "email", email,
    "source_service", "login-consent-provider",
    "success", valid,
    "timestamp", time.Now())
```

## Future Enhancements

1. **Add OAuth2 Client Credentials Flow**
   - Service-to-service authentication
   - Token-based authorization

2. **Implement API Gateway**
   - Centralized authentication
   - Rate limiting
   - Request routing

3. **Add Distributed Tracing**
   - OpenTelemetry integration
   - Track requests across services

4. **Implement RBAC**
   - Role-based access control
   - Fine-grained permissions

## Conclusion

The migration from direct database access to service-to-service gRPC calls significantly improves the security posture of the login-consent-provider. This architecture follows microservices best practices and provides a foundation for future security enhancements.

### Key Takeaways:
- ✅ Reduced attack surface
- ✅ Better separation of concerns
- ✅ Improved testability
- ✅ Easier to maintain and scale
- ✅ Production-ready with mTLS and monitoring
