# Clean Architecture Refactoring - Gateway Service

## Overview

The GraphQL Gateway has been successfully refactored to follow **Clean Architecture** principles. This refactoring separates concerns into distinct layers, making the codebase more maintainable, testable, and flexible.

## Architecture Layers

### 1. Domain Layer (`internal/domain/`)

The innermost layer containing business entities and interfaces. This layer has **no dependencies** on external frameworks or infrastructure.

#### Entities (`internal/domain/entity/`)
- **`user.go`**: User entity representing a user in the system
- **`portfolio.go`**: Portfolio, PortfolioSummary, Holding, and PortfolioPerformancePoint entities
- **`transaction.go`**: Transaction entity with TransactionType enum

#### Gateway Interfaces (`internal/domain/gateway/`)
- **`user_gateway.go`**: Interface for user service interactions
- **`portfolio_gateway.go`**: Interface for portfolio service interactions
- **`transaction_gateway.go`**: Interface for transaction service interactions

These interfaces define **what** the application needs from external services, without specifying **how** to get it.

### 2. Use Case Layer (`internal/usecase/`)

Contains application-specific business logic. Use cases orchestrate the flow of data between entities and gateways.

- **`user_usecase.go`**: User-related operations (GetUser, CreateUser, GetCurrentUser)
- **`portfolio_usecase.go`**: Portfolio operations (GetPortfolio, GetPortfolioSummary, GetHoldings, GetPortfolioPerformance)
- **`transaction_usecase.go`**: Transaction operations with validation (CreateTransaction)

**Key Features:**
- Input validation
- Business rule enforcement
- Orchestration of domain entities and gateways
- Independent of infrastructure details

### 3. Infrastructure Layer (`internal/infrastructure/`)

Implements the gateway interfaces defined in the domain layer. This layer handles all external communication.

#### gRPC Implementations (`internal/infrastructure/grpc/`)
- **`user_gateway.go`**: gRPC implementation of UserGateway
- **`portfolio_gateway.go`**: gRPC implementation of PortfolioGateway
- **`transaction_gateway.go`**: gRPC implementation of TransactionGateway

#### Mappers (`internal/infrastructure/mapper/`)
- **`proto_mapper.go`**: Converts between protobuf messages and domain entities

### 4. Presentation Layer (`graph/`)

The GraphQL layer that handles HTTP requests and responses.

#### Resolvers (`graph/`)
- **`resolver.go`**: Main resolver struct with dependency injection container
- **`schema.resolvers.go`**: GraphQL resolver implementations (thin layer delegating to use cases)

#### Mappers (`graph/mapper/`)
- **`graphql_mapper.go`**: Converts between GraphQL models and domain entities

### 5. Dependency Injection (`internal/container/`)

- **`container.go`**: Wires all dependencies together, creating a single source of truth for dependency management

## Data Flow

```
GraphQL Request
    ↓
Resolver (Presentation Layer)
    ↓
GraphQL Mapper (converts GraphQL input → Use Case input)
    ↓
Use Case (Application Logic)
    ↓
Gateway Interface (Domain)
    ↓
gRPC Gateway Implementation (Infrastructure)
    ↓
Proto Mapper (converts protobuf → Domain Entity)
    ↓
Domain Entity
    ↓
(reverse flow back to GraphQL Response)
```

## Key Benefits

### 1. **Testability**
- Each layer can be tested independently
- Mock implementations are easy to create
- Use cases can be tested without any infrastructure

### 2. **Maintainability**
- Clear separation of concerns
- Each layer has a single responsibility
- Easy to locate and modify code

### 3. **Flexibility**
- Easy to swap implementations (e.g., REST instead of gRPC)
- Infrastructure changes don't affect business logic
- Can add new features without modifying existing code

### 4. **Domain Focus**
- Business logic is isolated in the domain and use case layers
- No infrastructure concerns leak into business logic
- Domain entities are pure Go structs with no external dependencies

## Example: Creating a Transaction

### Before (Tightly Coupled)
```go
// Resolver directly calls gRPC and handles mapping
func (r *mutationResolver) CreateTransaction(ctx context.Context, input model.NewTransaction) (*model.Transaction, error) {
    // Parsing, validation, gRPC call, and mapping all mixed together
    executedAt, _ := parseTimestamp(input.ExecutedAt)
    req := &transactionpb.CreateTransactionRequest{...}
    resp, err := r.TransactionClient.CreateTransaction(ctx, req)
    // Map response...
}
```

### After (Clean Architecture)
```go
// Resolver delegates to use case
func (r *mutationResolver) CreateTransaction(ctx context.Context, input model.NewTransaction) (*model.Transaction, error) {
    userID, _ := auth.UserIDFromContext(ctx)
    useCaseInput, _ := mapper.GraphQLNewTransactionToUseCaseInput(input)
    tx, err := r.Container.TransactionUseCase.CreateTransaction(ctx, userID, useCaseInput)
    return mapper.TransactionEntityToGraphQL(tx), nil
}

// Use case handles validation and business logic
func (uc *TransactionUseCase) CreateTransaction(ctx context.Context, userID string, input CreateTransactionInput) (*entity.Transaction, error) {
    if err := uc.validateCreateTransactionInput(input); err != nil {
        return nil, err
    }
    return uc.transactionGateway.CreateTransaction(ctx, gatewayInput)
}

// Gateway handles infrastructure
func (g *TransactionGRPCGateway) CreateTransaction(ctx context.Context, input gateway.CreateTransactionInput) (*entity.Transaction, error) {
    req := mapper.CreateTransactionInputToProto(input)
    resp, err := g.client.CreateTransaction(ctx, req)
    return mapper.ProtoToTransactionEntity(resp.Transaction), nil
}
```

## Migration Guide

### For New Features

1. **Add domain entity** (if needed) in `internal/domain/entity/`
2. **Define gateway interface** (if needed) in `internal/domain/gateway/`
3. **Implement use case** in `internal/usecase/`
4. **Implement gateway** in `internal/infrastructure/grpc/`
5. **Add mappers** in `internal/infrastructure/mapper/` and `graph/mapper/`
6. **Update resolver** in `graph/schema.resolvers.go`
7. **Wire dependencies** in `internal/container/container.go`

### Testing Strategy

- **Unit tests for use cases**: Mock gateway interfaces
- **Unit tests for gateways**: Mock gRPC clients
- **Integration tests for resolvers**: Use real container with mock clients
- **End-to-end tests**: Test the entire flow

## Directory Structure

```
apps/gateway/
├── cmd/
│   └── server/
│       └── main.go                    # Entry point, initializes container
├── graph/
│   ├── mapper/
│   │   └── graphql_mapper.go         # GraphQL ↔ Domain mapping
│   ├── resolver.go                    # Resolver with container
│   └── schema.resolvers.go            # Thin resolver implementations
├── internal/
│   ├── container/
│   │   └── container.go               # Dependency injection
│   ├── domain/
│   │   ├── entity/                    # Business entities
│   │   │   ├── user.go
│   │   │   ├── portfolio.go
│   │   │   └── transaction.go
│   │   └── gateway/                   # Gateway interfaces
│   │       ├── user_gateway.go
│   │       ├── portfolio_gateway.go
│   │       └── transaction_gateway.go
│   ├── infrastructure/
│   │   ├── grpc/                      # gRPC implementations
│   │   │   ├── user_gateway.go
│   │   │   ├── portfolio_gateway.go
│   │   │   └── transaction_gateway.go
│   │   └── mapper/
│   │       └── proto_mapper.go        # Protobuf ↔ Domain mapping
│   └── usecase/                       # Application logic
│       ├── user_usecase.go
│       ├── portfolio_usecase.go
│       └── transaction_usecase.go
└── internal/
    ├── auth/                          # Cross-cutting concerns
    └── util/                          # Utilities
```

## Conclusion

This refactoring transforms the gateway from a simple proxy into a well-architected application that:
- Separates business logic from infrastructure
- Makes testing straightforward
- Enables easy feature additions and modifications
- Follows industry best practices for maintainable software
