# GraphQL Gateway User Integration

## Overview
This document describes the integration of user-service into the GraphQL Gateway, enabling user queries and mutations through the GraphQL API.

## Changes Made

### 1. GraphQL Schema Updates
**File**: `apps/gateway/graph/schema.graphqls`

Added a new `user` query to fetch user information by ID:
```graphql
type Query {
  me: User
  user(id: ID!): User  # NEW
  portfolio(id: ID!): Portfolio
}
```

### 2. Resolver Updates
**File**: `apps/gateway/graph/resolver.go`

Added `UserClient` to the Resolver struct:
```go
type Resolver struct {
    PortfolioClient portfoliopb.PortfolioServiceClient
    UserClient      userpb.UserServiceClient  // NEW
}
```

### 3. Resolver Implementation
**File**: `apps/gateway/graph/schema.resolvers.go`

Implemented three resolvers:

#### a. `User(id: ID!)` Query
Fetches a user by ID from the user-service:
```go
func (r *queryResolver) User(ctx context.Context, id string) (*model.User, error) {
    req := &userpb.GetUserRequest{
        Id: id,
    }
    
    resp, err := r.UserClient.GetUser(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    
    return &model.User{
        ID:       resp.Id,
        Username: resp.Name,
        Email:    resp.Email,
    }, nil
}
```

#### b. `Me` Query
Returns the currently authenticated user (currently hardcoded for demo):
```go
func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
    // For now, return a hardcoded user ID
    // In production, this should come from authentication context
    userID := "3a4b3185-5abb-4899-9835-829ddb91e3a6"
    return r.User(ctx, userID)
}
```

#### c. `CreateUser` Mutation
Creates a new user via the user-service:
```go
func (r *mutationResolver) CreateUser(ctx context.Context, input model.NewUser) (*model.User, error) {
    req := &userpb.CreateUserRequest{
        Email:    input.Email,
        Name:     input.Username,
        Password: "defaultPassword123", // In production, this should come from input
    }
    
    resp, err := r.UserClient.CreateUser(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    return &model.User{
        ID:       resp.Id,
        Username: input.Username,
        Email:    input.Email,
    }, nil
}
```

### 4. Gateway Main Updates
**File**: `apps/gateway/cmd/server/main.go`

Added user-service gRPC client connection:
```go
// Connect to user service
userServiceAddr := os.Getenv("USER_SERVICE_ADDR")
if userServiceAddr == "" {
    userServiceAddr = "localhost:50051"
}

userConn, err := grpc.NewClient(userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
if err != nil {
    l.Error("Failed to connect to user service", "error", err)
    os.Exit(1)
}
defer userConn.Close()

userClient := userpb.NewUserServiceClient(userConn)

srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
    PortfolioClient: portfolioClient,
    UserClient:      userClient,  // NEW
}}))
```

### 5. Docker Compose Configuration
**File**: `deployments/docker-compose/docker-compose.yml`

Added `USER_SERVICE_ADDR` environment variable to the gateway service:
```yaml
gateway:
  environment:
    - PORTFOLIO_SERVICE_ADDR=portfolio-service:50052
    - USER_SERVICE_ADDR=user-service:50051  # NEW
```

### 6. Go Module Dependencies
**File**: `apps/gateway/go.mod`

Added user-service dependency:
```go
require (
    github.com/garcios/portfolio-insights/services/user-service v0.0.0-00010101000000-000000000000
)

replace github.com/garcios/portfolio-insights/services/user-service => ../../services/user-service
```

## Testing

### 1. Get Current User (Me)
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"query": "query { me { id username email } }"}' \
  http://localhost:8080/query
```

**Response**:
```json
{
  "data": {
    "me": {
      "id": "3a4b3185-5abb-4899-9835-829ddb91e3a6",
      "username": "John Doe",
      "email": "john.doe@example.com"
    }
  }
}
```

### 2. Get User by ID
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"query": "query GetUser($id: ID!) { user(id: $id) { id username email } }", "variables": {"id": "3a4b3185-5abb-4899-9835-829ddb91e3a6"}}' \
  http://localhost:8080/query
```

**Response**:
```json
{
  "data": {
    "user": {
      "id": "3a4b3185-5abb-4899-9835-829ddb91e3a6",
      "username": "John Doe",
      "email": "john.doe@example.com"
    }
  }
}
```

### 3. Create User
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"query": "mutation CreateUser($input: NewUser!) { createUser(input: $input) { id username email } }", "variables": {"input": {"username": "Jane Smith", "email": "jane.smith@example.com"}}}' \
  http://localhost:8080/query
```

**Response**:
```json
{
  "data": {
    "createUser": {
      "id": "new-user-id",
      "username": "Jane Smith",
      "email": "jane.smith@example.com"
    }
  }
}
```

## Postman Collection
The updated Postman collection (`docs/graphql_queries.postman_collection.json`) includes:
- **Get Me**: Query to get the current user
- **Get User**: Query to get a user by ID
- **Get Portfolio**: Query to get a portfolio by ID
- **Create User**: Mutation to create a new user

## Architecture
```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ GraphQL Query
       ▼
┌─────────────────────┐
│  GraphQL Gateway    │
│  (Port 8080)        │
└──────┬──────┬───────┘
       │      │
       │      │ gRPC
       │      ▼
       │  ┌──────────────────┐
       │  │  User Service    │
       │  │  (Port 50051)    │
       │  └──────────────────┘
       │
       │ gRPC
       ▼
┌──────────────────┐
│ Portfolio Service│
│ (Port 50052)     │
└──────────────────┘
```

## Future Improvements
1. **Authentication**: Implement proper authentication and extract user ID from JWT tokens
2. **Password Input**: Add password field to the `NewUser` input type
3. **Error Handling**: Improve error messages and add proper error codes
4. **Caching**: Add caching layer for frequently accessed user data
5. **Field Resolvers**: Add nested resolvers to fetch user's portfolio directly from the User type
