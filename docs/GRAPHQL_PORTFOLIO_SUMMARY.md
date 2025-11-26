# GraphQL Portfolio Summary Integration

## Overview
This document describes the integration of portfolio summary information into the GraphQL Gateway, enabling clients to fetch comprehensive portfolio metrics including total value, gains/losses, and performance data.

## Changes Made

### 1. GraphQL Schema Updates
**File**: `apps/gateway/graph/schema.graphqls`

#### Added PortfolioSummary Type
```graphql
type PortfolioSummary {
  totalValue: Float!
  totalGainLoss: Float!
  totalGainLossPercentage: Float!
  currency: String!
  lastUpdated: String!
}
```

#### Updated Portfolio Type
```graphql
type Portfolio {
  id: ID!
  userId: ID!
  name: String!
  summary: PortfolioSummary      # NEW - nullable field
  holdings: [Holding!]!
}
```

### 2. Resolver Implementation
**File**: `apps/gateway/graph/schema.resolvers.go`

The `Portfolio` resolver was updated to fetch both holdings and summary data from the portfolio-service:

```go
func (r *queryResolver) Portfolio(ctx context.Context, id string) (*model.Portfolio, error) {
    // Fetch holdings
    holdingsReq := &portfoliopb.GetHoldingsRequest{
        UserId: id,
    }
    holdingsResp, err := r.PortfolioClient.GetHoldings(ctx, holdingsReq)
    if err != nil {
        return nil, fmt.Errorf("failed to get holdings: %w", err)
    }

    var holdings []*model.Holding
    for _, h := range holdingsResp.Holdings {
        holdings = append(holdings, &model.Holding{
            Symbol:   h.Symbol,
            Quantity: h.Quantity,
            Value:    h.CurrentValue,
        })
    }

    // Fetch portfolio summary
    summaryReq := &portfoliopb.GetPortfolioSummaryRequest{
        UserId: id,
    }
    summaryResp, err := r.PortfolioClient.GetPortfolioSummary(ctx, summaryReq)
    if err != nil {
        // If summary fails, return portfolio without summary
        return &model.Portfolio{
            ID:       id,
            UserID:   id,
            Name:     "My Portfolio",
            Summary:  nil,
            Holdings: holdings,
        }, nil
    }

    var summary *model.PortfolioSummary
    if summaryResp.Summary != nil {
        summary = &model.PortfolioSummary{
            TotalValue:              summaryResp.Summary.TotalValue,
            TotalGainLoss:           summaryResp.Summary.TotalGainLoss,
            TotalGainLossPercentage: summaryResp.Summary.TotalGainLossPercentage,
            Currency:                summaryResp.Summary.Currency,
            LastUpdated:             summaryResp.Summary.LastUpdated.AsTime().Format("2006-01-02T15:04:05Z07:00"),
        }
    }

    return &model.Portfolio{
        ID:       id,
        UserID:   id,
        Name:     "My Portfolio",
        Summary:  summary,
        Holdings: holdings,
    }, nil
}
```

### Key Implementation Details

1. **Graceful Degradation**: If the summary fetch fails, the resolver returns the portfolio with `summary: null` instead of failing the entire query
2. **Timestamp Conversion**: The protobuf timestamp is converted to ISO 8601 format string
3. **Parallel Data Fetching**: Holdings and summary are fetched in sequence (could be optimized with goroutines)

## GraphQL Schema Mapping

### Proto to GraphQL Type Mapping

| Proto Field (PortfolioSummary) | GraphQL Field | Type | Notes |
|-------------------------------|---------------|------|-------|
| `total_value` | `totalValue` | `Float!` | Total portfolio value |
| `total_gain_loss` | `totalGainLoss` | `Float!` | Total profit/loss |
| `total_gain_loss_percentage` | `totalGainLossPercentage` | `Float!` | Percentage gain/loss |
| `currency` | `currency` | `String!` | Currency code (e.g., "AUD") |
| `last_updated` | `lastUpdated` | `String!` | ISO 8601 timestamp |

## Testing

### Query with Summary
```graphql
query GetPortfolio($id: ID!) {
  portfolio(id: $id) {
    id
    userId
    name
    summary {
      totalValue
      totalGainLoss
      totalGainLossPercentage
      currency
      lastUpdated
    }
    holdings {
      symbol
      quantity
      value
    }
  }
}
```

### Example cURL Request
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"query": "query GetPortfolio($id: ID!) { portfolio(id: $id) { id userId name summary { totalValue totalGainLoss totalGainLossPercentage currency lastUpdated } holdings { symbol quantity value } } }", "variables": {"id": "0fdc618a-0cd2-4c04-aec4-44646c813e6a"}}' \
  http://localhost:8080/query
```

### Example Response
```json
{
  "data": {
    "portfolio": {
      "id": "0fdc618a-0cd2-4c04-aec4-44646c813e6a",
      "userId": "0fdc618a-0cd2-4c04-aec4-44646c813e6a",
      "name": "My Portfolio",
      "summary": {
        "totalValue": 15750.50,
        "totalGainLoss": 1250.50,
        "totalGainLossPercentage": 8.62,
        "currency": "AUD",
        "lastUpdated": "2025-11-26T03:32:50Z"
      },
      "holdings": [
        {
          "symbol": "AAPL",
          "quantity": 10.5,
          "value": 1577.625
        }
      ]
    }
  }
}
```

### Query without Summary (Optional Field)
Since `summary` is nullable, clients can choose not to request it:

```graphql
query GetPortfolioBasic($id: ID!) {
  portfolio(id: $id) {
    id
    name
    holdings {
      symbol
      quantity
    }
  }
}
```

## Error Handling

### Summary Fetch Failure
If the `GetPortfolioSummary` gRPC call fails:
- The query **does not fail**
- The `summary` field returns `null`
- Holdings are still returned
- This allows partial data to be displayed to users

Example response when summary fails:
```json
{
  "data": {
    "portfolio": {
      "id": "user-123",
      "userId": "user-123",
      "name": "My Portfolio",
      "summary": null,
      "holdings": [...]
    }
  }
}
```

### Holdings Fetch Failure
If the `GetHoldings` gRPC call fails:
- The entire query fails
- An error is returned in the GraphQL response

## Performance Considerations

### Current Implementation
- Holdings and summary are fetched sequentially
- Two separate gRPC calls per portfolio query

### Potential Optimizations
1. **Parallel Fetching**: Use goroutines to fetch holdings and summary concurrently
2. **DataLoader**: Implement DataLoader pattern for batch fetching
3. **Caching**: Add Redis caching for frequently accessed portfolios
4. **Field-Level Resolvers**: Only fetch summary when explicitly requested

### Example Parallel Fetching
```go
var (
    holdings []*model.Holding
    summary  *model.PortfolioSummary
    holdingsErr, summaryErr error
)

var wg sync.WaitGroup
wg.Add(2)

// Fetch holdings
go func() {
    defer wg.Done()
    // ... fetch holdings
}()

// Fetch summary
go func() {
    defer wg.Done()
    // ... fetch summary
}()

wg.Wait()
```

## Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ GraphQL Query
       │ portfolio(id: "user-123") {
       │   summary { totalValue }
       │   holdings { symbol }
       │ }
       ▼
┌─────────────────────────┐
│  GraphQL Gateway        │
│  Portfolio Resolver     │
└──────┬──────────┬───────┘
       │          │
       │          │ gRPC: GetPortfolioSummary
       │          ▼
       │      ┌──────────────────────┐
       │      │  Portfolio Service   │
       │      │  - Calculates totals │
       │      │  - Aggregates data   │
       │      └──────────────────────┘
       │
       │ gRPC: GetHoldings
       ▼
   ┌──────────────────────┐
   │  Portfolio Service   │
   │  - Returns holdings  │
   └──────────────────────┘
```

## Files Modified

1. **`apps/gateway/graph/schema.graphqls`**
   - Added `PortfolioSummary` type
   - Added `summary` field to `Portfolio` type

2. **`apps/gateway/graph/schema.resolvers.go`**
   - Updated `Portfolio` resolver to fetch summary
   - Added timestamp formatting
   - Implemented graceful error handling

3. **`apps/gateway/graph/model/models_gen.go`** (auto-generated)
   - Generated `PortfolioSummary` struct
   - Updated `Portfolio` struct with `Summary` field

4. **`docs/graphql_queries.postman_collection.json`**
   - Updated "Get Portfolio" query to include summary fields

## Future Enhancements

1. **Real-time Updates**: Add GraphQL subscriptions for live portfolio updates
2. **Historical Data**: Add time-series data for portfolio value over time
3. **Multi-Currency**: Support viewing portfolio in different currencies
4. **Performance Metrics**: Add additional metrics like Sharpe ratio, volatility
5. **Comparison**: Add ability to compare portfolio performance against benchmarks
6. **Caching Strategy**: Implement intelligent caching based on market hours
7. **Field Resolvers**: Make summary a separate field resolver for better performance

## Related Documentation

- [GraphQL User Integration](./GRAPHQL_USER_INTEGRATION.md)
- [Portfolio Service Proto](../proto/portfolio/portfolio.proto)
- [GraphQL Testing Guide](./GRAPHQL_TESTING_GUIDE.md)
