# Gateway Clean Architecture - Visual Overview

## Layer Dependency Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                     PRESENTATION LAYER                          │
│                      (graph/)                                   │
│                                                                 │
│  ┌──────────────┐         ┌──────────────┐                    │
│  │   Resolvers  │────────▶│   Mappers    │                    │
│  │              │         │  (GraphQL ↔  │                    │
│  │ - Query      │         │   Domain)    │                    │
│  │ - Mutation   │         │              │                    │
│  │ - Field      │         └──────────────┘                    │
│  └──────────────┘                                              │
└────────────────────┬────────────────────────────────────────────┘
                     │ Uses
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│                      USE CASE LAYER                             │
│                    (internal/usecase/)                          │
│                                                                 │
│  ┌─────────────────┐  ┌──────────────────┐  ┌───────────────┐ │
│  │  UserUseCase    │  │ PortfolioUseCase │  │TransactionUC  │ │
│  │                 │  │                  │  │               │ │
│  │ - GetUser       │  │ - GetPortfolio   │  │ - Create      │ │
│  │ - CreateUser    │  │ - GetSummary     │  │ - Validate    │ │
│  │ - GetCurrent    │  │ - GetHoldings    │  │               │ │
│  └─────────────────┘  └──────────────────┘  └───────────────┘ │
└────────────────────┬────────────────────────────────────────────┘
                     │ Depends on
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│                       DOMAIN LAYER                              │
│                    (internal/domain/)                           │
│                                                                 │
│  ┌────────────────────────────────────────────────────────┐   │
│  │                    Entities                            │   │
│  │  ┌──────────┐  ┌──────────────┐  ┌──────────────┐    │   │
│  │  │   User   │  │  Portfolio   │  │ Transaction  │    │   │
│  │  └──────────┘  └──────────────┘  └──────────────┘    │   │
│  └────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌────────────────────────────────────────────────────────┐   │
│  │              Gateway Interfaces                        │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐ │   │
│  │  │UserGateway   │  │PortfolioGW   │  │TransactionGW│ │   │
│  │  │(interface)   │  │(interface)   │  │(interface)  │ │   │
│  │  └──────────────┘  └──────────────┘  └─────────────┘ │   │
│  └────────────────────────────────────────────────────────┘   │
└────────────────────┬────────────────────────────────────────────┘
                     │ Implemented by
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│                   INFRASTRUCTURE LAYER                          │
│                 (internal/infrastructure/)                      │
│                                                                 │
│  ┌────────────────────────────────────────────────────────┐   │
│  │              gRPC Gateway Implementations              │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐ │   │
│  │  │UserGRPCGW    │  │PortfolioGRPC │  │TransactionGW│ │   │
│  │  │              │  │Gateway       │  │             │ │   │
│  │  └──────────────┘  └──────────────┘  └─────────────┘ │   │
│  └────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌────────────────────────────────────────────────────────┐   │
│  │                    Mappers                             │   │
│  │         (Protobuf ↔ Domain Entities)                   │   │
│  └────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌────────────────────────────────────────────────────────┐   │
│  │              External gRPC Clients                     │   │
│  │  • UserServiceClient                                   │   │
│  │  • PortfolioServiceClient                              │   │
│  │  • TransactionServiceClient                            │   │
│  └────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## Request Flow Example: Create Transaction

```
1. GraphQL Request
   POST /query
   mutation { createTransaction(...) }
        │
        ▼
2. Resolver (Presentation)
   schema.resolvers.go:CreateTransaction()
        │
        ├─▶ Extract userID from context
        │
        ├─▶ GraphQL Mapper
        │   GraphQLNewTransactionToUseCaseInput()
        │   (GraphQL model → Use Case DTO)
        │
        ▼
3. Use Case (Application Logic)
   TransactionUseCase.CreateTransaction()
        │
        ├─▶ Validate input
        │   - Check required fields
        │   - Validate transaction type
        │   - Check date constraints
        │
        ├─▶ Convert to Gateway input
        │
        ▼
4. Gateway Interface (Domain)
   TransactionGateway.CreateTransaction()
        │
        ▼
5. gRPC Implementation (Infrastructure)
   TransactionGRPCGateway.CreateTransaction()
        │
        ├─▶ Proto Mapper
        │   CreateTransactionInputToProto()
        │   (Domain → Protobuf)
        │
        ├─▶ gRPC Client Call
        │   transactionClient.CreateTransaction()
        │
        ├─▶ Proto Mapper
        │   ProtoToTransactionEntity()
        │   (Protobuf → Domain)
        │
        ▼
6. Domain Entity
   Transaction entity returned
        │
        ▼
7. Use Case returns to Resolver
        │
        ▼
8. GraphQL Mapper
   TransactionEntityToGraphQL()
   (Domain → GraphQL model)
        │
        ▼
9. GraphQL Response
   { data: { createTransaction: {...} } }
```

## Dependency Injection Container

```
┌─────────────────────────────────────────────────────────────┐
│                    main.go                                  │
│                                                             │
│  1. Create gRPC Clients                                    │
│     • userClient                                           │
│     • portfolioClient                                      │
│     • transactionClient                                    │
│                                                             │
│  2. Initialize Container                                   │
│     container.NewContainer(clients...)                     │
│        │                                                    │
│        ├─▶ Creates Gateways                               │
│        │   (Infrastructure Layer)                          │
│        │                                                    │
│        └─▶ Creates Use Cases                              │
│            (Application Layer)                             │
│                                                             │
│  3. Create Resolver                                        │
│     &Resolver{ Container: c }                             │
│                                                             │
│  4. Start GraphQL Server                                   │
└─────────────────────────────────────────────────────────────┘
```

## Testing Strategy

```
┌──────────────────────────────────────────────────────────────┐
│                    Unit Tests                                │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Use Case Tests                                             │
│  ┌────────────────────────────────────────────────────┐    │
│  │  • Mock gateway interfaces                         │    │
│  │  • Test business logic in isolation                │    │
│  │  • Verify validation rules                         │    │
│  └────────────────────────────────────────────────────┘    │
│                                                              │
│  Gateway Tests                                              │
│  ┌────────────────────────────────────────────────────┐    │
│  │  • Mock gRPC clients                               │    │
│  │  • Test mapping logic                              │    │
│  │  • Verify error handling                           │    │
│  └────────────────────────────────────────────────────┘    │
│                                                              │
└──────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│                 Integration Tests                            │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Resolver Tests                                             │
│  ┌────────────────────────────────────────────────────┐    │
│  │  • Use container with mock clients                 │    │
│  │  • Test full resolver flow                         │    │
│  │  • Verify GraphQL schema compliance                │    │
│  └────────────────────────────────────────────────────┘    │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

## Key Principles

### 1. Dependency Rule
**Dependencies point inward**
- Presentation → Use Case → Domain
- Infrastructure implements Domain interfaces
- Domain has NO dependencies on outer layers

### 2. Interface Segregation
**Small, focused interfaces**
- Each gateway interface serves one purpose
- Easy to mock and test
- Clear contracts between layers

### 3. Single Responsibility
**Each layer has one job**
- Domain: Business entities and rules
- Use Case: Application logic
- Infrastructure: External communication
- Presentation: HTTP/GraphQL handling

### 4. Dependency Inversion
**Depend on abstractions, not concretions**
- Use cases depend on gateway interfaces
- Infrastructure implements those interfaces
- Easy to swap implementations
