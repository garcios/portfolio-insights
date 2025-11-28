# MarketData Service Unit Tests

This document describes the unit tests for the marketdata-service.

## Overview

The marketdata-service unit tests focus on the business logic layer (usecase) to ensure proper handling of asset data, price information, and currency rates.

## Test Coverage

- **Usecase Layer**: **76.2%** coverage
- **Handler Layer**: 9.1% coverage (basic tests exist)
- **Repository Layer**: 0% (integration tests recommended)

## Test Structure

### Mock Repository

The tests use a mock repository implementation (`MockMarketDataRepository`) that simulates database operations in memory:

```go
type MockMarketDataRepository struct {
    assets        map[string]*domain.Asset
    prices        map[string][]*domain.AssetPrice
    currencyRates map[string]*domain.CurrencyRate
}
```

This allows testing business logic without database dependencies.

## Test Cases

### Asset Management

#### `TestGetAsset`
- **Success**: Retrieves an existing asset by symbol
- **NotFound**: Returns error for non-existent asset
- **Coverage**: Asset lookup, error handling

#### `TestListAssets`
- **FirstPage**: Tests pagination with page size limit
- **AllAssets**: Retrieves all assets when page size exceeds total
- **Coverage**: Pagination logic, page token generation

### Price Operations

#### `TestGetLatestPrice`
- **Success**: Retrieves the most recent price for a symbol
- **NotFound**: Returns error when no prices exist
- **Coverage**: Latest price selection, error handling

#### `TestGetLatestPrices`
- **Success**: Retrieves latest prices for multiple symbols
- **PartialResults**: Handles mix of valid and invalid symbols
- **Coverage**: Bulk price retrieval, partial success scenarios

#### `TestGetHistoricalPrices`
- **Success**: Retrieves price history within date range
- **EmptyResult**: Returns empty array for symbols with no history
- **Coverage**: Historical data filtering, empty result handling

### Currency Rate Operations

#### `TestGetLatestCurrencyRate`
- **Success**: Retrieves current exchange rate for currency pair
- **NotFound**: Returns error for unavailable currency pairs
- **Coverage**: Currency rate lookup, error handling

#### `TestGetHistoricalCurrencyRates`
- **Success**: Retrieves historical exchange rates
- **Coverage**: Date range filtering for currency rates

## Running Tests

### Run all tests
```bash
cd services/marketdata-service
go test ./... -v
```

### Run usecase tests only
```bash
cd services/marketdata-service
go test ./internal/usecase/... -v
```

### Run with coverage
```bash
cd services/marketdata-service
go test ./... -cover -coverprofile=coverage.out
```

### View coverage report
```bash
cd services/marketdata-service
go tool cover -html=coverage.out
```

## CI Integration

MarketData service tests are automatically run in the GitHub Actions workflow:

1. **Test Execution**: Tests run with race detection and coverage tracking
2. **Coverage Upload**: Results uploaded to Codecov with `marketdata-service` flag
3. **Environment**: Tests use mock repository, no database required

## Test Patterns

### Mock Setup Pattern
```go
repo := NewMockRepo()
repo.assets["AAPL"] = &domain.Asset{
    Symbol: "AAPL",
    Name:   "Apple Inc.",
    Type:   "stock",
}

uc := NewMarketDataUsecase(repo)
```

### Testing Success and Error Cases
```go
t.Run("Success", func(t *testing.T) {
    asset, err := uc.GetAsset("AAPL")
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    // Assertions...
})

t.Run("NotFound", func(t *testing.T) {
    _, err := uc.GetAsset("INVALID")
    if err == nil {
        t.Error("expected error, got nil")
    }
})
```

## Coverage Analysis

### Well-Covered Areas (76.2%)
- ✅ Asset retrieval and listing
- ✅ Price queries (latest and historical)
- ✅ Currency rate lookups
- ✅ Pagination logic
- ✅ Error handling

### Areas Not Covered
- ⚠️ Repository layer (database operations)
- ⚠️ gRPC handlers (minimal coverage)
- ⚠️ Worker processes
- ⚠️ Middleware
- ⚠️ Metrics collection

## Future Improvements

1. **Handler Tests**: Add more comprehensive gRPC handler tests
2. **Integration Tests**: Add database integration tests for repository layer
3. **Worker Tests**: Test background price update workers
4. **Error Scenarios**: Add more edge case and error condition tests
5. **Performance Tests**: Add benchmarks for critical paths

## Why Repository Tests Were Removed

Initial attempts to test the repository layer with `sqlmock` proved challenging due to:
- Complex SQL queries with schema prefixes (`marketdata.assets`)
- JOIN operations that are hard to mock accurately
- Tight coupling between SQL structure and mock expectations

**Recommendation**: Use integration tests with a test database for repository layer instead of mocking SQL.

## Related Documentation

- [CI Fixes Documentation](../../docs/CI_FIXES.md)
- [Codecov Setup](../../docs/CODECOV_SETUP.md)
- [Domain Models](./internal/domain/marketdata.go)
