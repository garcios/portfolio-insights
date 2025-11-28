# Transaction Service Unit Tests

This document describes the unit tests for the transaction-service.

## Overview

The transaction-service unit tests focus on the business logic layer (usecase) to ensure proper handling of transaction creation, retrieval, updates, and deletion, as well as validation logic.

## Test Coverage

- **Usecase Layer**: **30.1%** coverage (Core transaction logic covered, CSV upload logic pending)
- **Handler Layer**: 18.4% coverage (basic tests exist)

## Test Structure

### Mock Repository

The tests use a mock repository implementation (`MockTransactionRepository`) that simulates database operations in memory:

```go
type MockTransactionRepository struct {
    transactions map[string]*domain.Transaction
    createError  error
    getError     error
}
```

### Mock Gateways

Mock implementations for external services:
- `MockUserGateway`: Simulates user existence checks
- `MockMarketDataGateway`: Simulates asset existence checks
- `MockEventPublisher`: Captures published events for verification

## Test Cases

### Transaction Management

#### `TestCreateTransaction`
- **Success_BUY**: Creates a valid BUY transaction
- **Success_SELL**: Creates a valid SELL transaction
- **UserNotFound**: Validates user existence check
- **AssetNotFound**: Validates asset existence check
- **InvalidTransactionType**: Validates type (BUY/SELL)
- **ZeroQuantity**: Validates quantity > 0
- **NegativePrice**: Validates price >= 0

#### `TestGetTransaction`
- **Success**: Retrieves existing transaction
- **NotFound**: Returns error for non-existent transaction

#### `TestListTransactions`
- **FilterByUser**: Retrieves transactions for specific user
- **Pagination**: Tests limit/offset logic
- **EmptyResult**: Returns empty list for user with no transactions

#### `TestUpdateTransaction`
- **Success**: Updates transaction fields
- **NotFound**: Returns error for non-existent transaction

#### `TestDeleteTransaction`
- **Success**: Deletes transaction
- **NotFound**: Returns error for non-existent transaction

## Running Tests

### Run all tests
```bash
cd services/transaction-service
go test ./... -v
```

### Run usecase tests only
```bash
cd services/transaction-service
go test ./internal/usecase/... -v
```

### Run with coverage
```bash
cd services/transaction-service
go test ./... -cover -coverprofile=coverage.out
```

### View coverage report
```bash
cd services/transaction-service
go tool cover -html=coverage.out
```

## CI Integration

Transaction service tests are automatically run in the GitHub Actions workflow:

1. **Test Execution**: Tests run with race detection and coverage tracking
2. **Coverage Upload**: Results uploaded to Codecov with `transaction-service` flag
3. **Environment**: Tests use mock repository, no database required

## Test Patterns

### Mock Setup Pattern
```go
repo := NewMockRepo()
userGateway := &MockUserGateway{exists: true}
marketGateway := &MockMarketDataGateway{exists: true}
eventPublisher := &MockEventPublisher{}
uc := NewTransactionUsecase(repo, userGateway, marketGateway, eventPublisher)
```

## Future Improvements

1. **CSV Upload Tests**: Add tests for `csv_upload_usecase.go` to improve coverage
2. **Handler Tests**: Add more comprehensive gRPC handler tests
3. **Integration Tests**: Add database integration tests for repository layer
4. **Event Publishing**: Verify event payload content in tests

## Related Documentation

- [CI Fixes Documentation](../../docs/CI_FIXES.md)
- [Codecov Setup](../../docs/CODECOV_SETUP.md)
- [Domain Models](./internal/domain/transaction.go)
