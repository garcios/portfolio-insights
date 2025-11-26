# GraphQL Field-Level Resolvers Optimization

## Problem
The original `Portfolio` query resolver was fetching **all** data (holdings AND summary) regardless of what fields the client actually requested. This is inefficient because:

1. If a query only requests `holdings`, we still fetched the `summary` (unnecessary gRPC call)
2. If a query only requests `summary`, we still fetched `holdings` (unnecessary gRPC call)
3. This violates GraphQL's core principle of "ask for what you need"

## Solution
Implemented **field-level resolvers** for the `Portfolio` type:

### Before (Over-fetching)
```go
func (r *queryResolver) Portfolio(ctx context.Context, id string) (*model.Portfolio, error) {
    // Always fetch holdings
    holdingsResp, err := r.PortfolioClient.GetHoldings(ctx, holdingsReq)
    // ...
    
    // Always fetch summary
    summaryResp, err := r.PortfolioClient.GetPortfolioSummary(ctx, summaryReq)
    // ...
    
    return &model.Portfolio{
        Holdings: holdings,
        Summary:  summary,
    }, nil
}
```

### After (Optimized)
```go
// Portfolio query resolver - returns basic info only
func (r *queryResolver) Portfolio(ctx context.Context, id string) (*model.Portfolio, error) {
    return &model.Portfolio{
        ID:     id,
        UserID: id,
        Name:   "My Portfolio",
    }, nil
}

// Holdings field resolver - only executes if holdings are requested
func (r *portfolioResolver) Holdings(ctx context.Context, obj *model.Portfolio) ([]*model.Holding, error) {
    holdingsResp, err := r.PortfolioClient.GetHoldings(ctx, holdingsReq)
    // ...
    return holdings, nil
}

// Summary field resolver - only executes if summary is requested
func (r *portfolioResolver) Summary(ctx context.Context, obj *model.Portfolio) (*model.PortfolioSummary, error) {
    summaryResp, err := r.PortfolioClient.GetPortfolioSummary(ctx, summaryReq)
    // ...
    return summary, nil
}
```

## How It Works

GraphQL's execution engine automatically calls the appropriate field resolvers based on the query:

### Query 1: Only Holdings
```graphql
query {
  portfolio(id: "user-123") {
    id
    holdings {
      symbol
      quantity
    }
  }
}
```
**Execution:**
1. ✅ `Portfolio()` resolver called → returns basic info
2. ✅ `Holdings()` resolver called → fetches from gRPC
3. ❌ `Summary()` resolver **NOT** called → saves 1 gRPC call

### Query 2: Only Summary
```graphql
query {
  portfolio(id: "user-123") {
    id
    summary {
      totalValue
      currency
    }
  }
}
```
**Execution:**
1. ✅ `Portfolio()` resolver called → returns basic info
2. ❌ `Holdings()` resolver **NOT** called → saves 1 gRPC call
3. ✅ `Summary()` resolver called → fetches from gRPC

### Query 3: Both
```graphql
query {
  portfolio(id: "user-123") {
    id
    holdings { symbol }
    summary { totalValue }
  }
}
```
**Execution:**
1. ✅ `Portfolio()` resolver called
2. ✅ `Holdings()` resolver called
3. ✅ `Summary()` resolver called

## Performance Impact

| Scenario | Before | After | Savings |
|----------|--------|-------|---------|
| Only holdings requested | 2 gRPC calls | 1 gRPC call | **50%** |
| Only summary requested | 2 gRPC calls | 1 gRPC call | **50%** |
| Both requested | 2 gRPC calls | 2 gRPC calls | 0% (no overhead) |

## Implementation Details

### Files Modified
1. **`apps/gateway/graph/schema.resolvers.go`**
   - Split `Portfolio` resolver into 3 parts:
     - `Portfolio()` - returns basic fields
     - `Holdings()` - field-level resolver
     - `Summary()` - field-level resolver
   - Added `portfolioResolver` type
   - Added `Portfolio()` method to wire up the resolver

2. **`apps/gateway/graph/generated/generated.go`**
   - Added `PortfolioResolver` interface
   - Updated `ResolverRoot` to include `Portfolio()` method

### Key Concepts
- **Query Resolver**: Resolves top-level queries (`portfolio(id: ID!)`)
- **Field Resolver**: Resolves specific fields on a type (`Portfolio.holdings`, `Portfolio.summary`)
- **Lazy Evaluation**: Field resolvers only execute when their field is requested

## Best Practices
✅ Use field-level resolvers for:
- Fields that require expensive operations (database queries, API calls)
- Fields that are optional or not always needed
- Fields with complex business logic

❌ Don't use field-level resolvers for:
- Simple scalar fields that are already in memory
- Fields that are always requested together
- Fields with trivial computation

## Testing
To verify the optimization works:

```bash
# Start the gateway
cd apps/gateway
go run cmd/server/main.go

# Test with only holdings
curl -X POST http://localhost:8080/query \
  -H "Content-Type: application/json" \
  -d '{"query": "{ portfolio(id: \"user-123\") { holdings { symbol } } }"}'

# Test with only summary
curl -X POST http://localhost:8080/query \
  -H "Content-Type: application/json" \
  -d '{"query": "{ portfolio(id: \"user-123\") { summary { totalValue } } }"}'
```

Check the logs to confirm only the requested gRPC calls are made.

## References
- [GraphQL Best Practices - Resolver Design](https://graphql.org/learn/execution/)
- [gqlgen Documentation - Field Resolvers](https://gqlgen.com/reference/resolvers/)
