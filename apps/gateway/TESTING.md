# Gateway Unit Tests

This document describes the unit tests for the GraphQL gateway application.

## Overview

The gateway unit tests verify the functionality of GraphQL resolvers that act as a bridge between the GraphQL API and the underlying gRPC microservices.

## Test Coverage

Current coverage: **71.8%** of statements in the `graph` package.

## Test Structure

### Mock Clients

The tests use mock gRPC clients to simulate responses from backend services:

- **MockUserServiceClient**: Simulates the user-service gRPC client
- **MockPortfolioServiceClient**: Simulates the portfolio-service gRPC client

These mocks allow us to test resolver logic in isolation without requiring actual service connections.

## Test Cases

### Query Resolvers

#### `TestQueryResolver_User`
- **Purpose**: Verifies that the `user` query correctly calls the user service and transforms the response
- **Coverage**: User lookup by ID
- **Assertions**: User ID, username, and email are correctly mapped

#### `TestQueryResolver_Portfolio`
- **Purpose**: Tests the basic portfolio query that returns portfolio metadata
- **Coverage**: Portfolio ID and user ID mapping
- **Assertions**: Correct portfolio structure is returned

#### `TestQueryResolver_PortfolioPerformance`
- **Purpose**: Validates portfolio performance data retrieval over time
- **Coverage**: Performance data points with timestamps and values
- **Assertions**: Correct number of data points, proper value mapping, timestamp formatting

### Mutation Resolvers

#### `TestMutationResolver_CreateUser`
- **Purpose**: Tests user creation mutation
- **Coverage**: User creation request and response handling
- **Assertions**: New user ID is returned, input data is preserved

### Field Resolvers

#### `TestPortfolioResolver_Summary`
- **Purpose**: Tests the portfolio summary field resolver
- **Coverage**: Portfolio summary data including total value, gain/loss, and currency
- **Assertions**: All summary fields are correctly mapped, timestamps are formatted properly

#### `TestPortfolioResolver_Holdings`
- **Purpose**: Validates holdings data retrieval and transformation
- **Coverage**: Multiple holdings with complete data (symbol, quantity, prices, gain/loss)
- **Assertions**: Correct number of holdings, all fields properly mapped

## Running Tests

### Run all gateway tests
```bash
cd apps/gateway
go test ./graph/... -v
```

### Run with coverage
```bash
cd apps/gateway
go test ./graph/... -cover -coverprofile=coverage.out
```

### View coverage report
```bash
cd apps/gateway
go tool cover -html=coverage.out
```

## CI Integration

Gateway tests are automatically run in the GitHub Actions workflow:

1. **Test Execution**: Tests run with race detection and coverage tracking
2. **Coverage Upload**: Results are uploaded to Codecov with the `gateway` flag
3. **Environment**: Tests run in isolation without requiring actual service connections

## Test Patterns

### Mock Setup Pattern
```go
mockClient := &MockUserServiceClient{
    GetUserFunc: func(ctx context.Context, in *userpb.GetUserRequest, opts ...grpc.CallOption) (*userpb.GetUserResponse, error) {
        return &userpb.GetUserResponse{
            Id:    "user-123",
            Name:  "John Doe",
            Email: "john@example.com",
        }, nil
    },
}
```

### Resolver Testing Pattern
```go
resolver := &Resolver{
    UserClient: mockClient,
}

queryResolver := &queryResolver{resolver}

result, err := queryResolver.User(context.Background(), "user-123")
// Assert on result and err
```

## Future Improvements

1. **Error Handling Tests**: Add tests for error scenarios (service unavailable, invalid input, etc.)
2. **Context Propagation**: Test that context values (auth, tracing) are properly passed to services
3. **Field-Level Authorization**: Add tests for authorization logic when implemented
4. **Pagination Tests**: Add tests for paginated queries when implemented
5. **Integration Tests**: Add end-to-end tests with actual gRPC services running

## Related Documentation

- [GraphQL Schema](../graph/schema.graphqls)
- [Field-Level Resolvers Guide](../FIELD_LEVEL_RESOLVERS.md)
- [CI Fixes Documentation](../../docs/CI_FIXES.md)
- [Codecov Setup](../../docs/CODECOV_SETUP.md)
