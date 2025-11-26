# GraphQL Holdings Enhancement

## Overview
This document describes the enhancement of the `Holding` type in the GraphQL schema to include comprehensive holding information including pricing, gains/losses, and currency data.

## Changes Made

### 1. GraphQL Schema Updates
**File**: `apps/gateway/graph/schema.graphqls`

#### Updated Holding Type
**Before:**
```graphql
type Holding {
  symbol: String!
  quantity: Float!
  value: Float!
}
```

**After:**
```graphql
type Holding {
  symbol: String!
  quantity: Float!
  averagePrice: Float!
  currentPrice: Float!
  currentValue: Float!
  gainLoss: Float!
  gainLossPercentage: Float!
  currency: String!
}
```

### 2. Resolver Implementation
**File**: `apps/gateway/graph/schema.resolvers.go`

Updated the holdings mapping in the `Portfolio` resolver to include all fields from the proto:

**Before:**
```go
var holdings []*model.Holding
for _, h := range holdingsResp.Holdings {
    holdings = append(holdings, &model.Holding{
        Symbol:   h.Symbol,
        Quantity: h.Quantity,
        Value:    h.CurrentValue,
    })
}
```

**After:**
```go
var holdings []*model.Holding
for _, h := range holdingsResp.Holdings {
    holdings = append(holdings, &model.Holding{
        Symbol:              h.Symbol,
        Quantity:            h.Quantity,
        AveragePrice:        h.AveragePrice,
        CurrentPrice:        h.CurrentPrice,
        CurrentValue:        h.CurrentValue,
        GainLoss:            h.GainLoss,
        GainLossPercentage:  h.GainLossPercentage,
        Currency:            h.Currency,
    })
}
```

## Field Descriptions

### Holding Fields

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `symbol` | `String!` | Stock ticker symbol | `"AAPL"` |
| `quantity` | `Float!` | Number of shares held | `10.5` |
| `averagePrice` | `Float!` | Average purchase price per share | `150.25` |
| `currentPrice` | `Float!` | Current market price per share | `175.50` |
| `currentValue` | `Float!` | Total current value (quantity × currentPrice) | `1842.75` |
| `gainLoss` | `Float!` | Total profit/loss amount | `265.125` |
| `gainLossPercentage` | `Float!` | Percentage gain/loss | `16.82` |
| `currency` | `String!` | Currency code | `"USD"` |

## Proto to GraphQL Mapping

| Proto Field | GraphQL Field | Notes |
|-------------|---------------|-------|
| `symbol` | `symbol` | Direct mapping |
| `quantity` | `quantity` | Direct mapping |
| `average_price` | `averagePrice` | Camel case conversion |
| `current_price` | `currentPrice` | Camel case conversion |
| `current_value` | `currentValue` | Camel case conversion |
| `gain_loss` | `gainLoss` | Camel case conversion |
| `gain_loss_percentage` | `gainLossPercentage` | Camel case conversion |
| `currency` | `currency` | Direct mapping |

## GraphQL Queries

### Complete Portfolio Query
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
      averagePrice
      currentPrice
      currentValue
      gainLoss
      gainLossPercentage
      currency
    }
  }
}
```

### Holdings Only Query
```graphql
query GetHoldings($id: ID!) {
  portfolio(id: $id) {
    holdings {
      symbol
      quantity
      averagePrice
      currentPrice
      currentValue
      gainLoss
      gainLossPercentage
      currency
    }
  }
}
```

### Minimal Holdings Query
```graphql
query GetMinimalHoldings($id: ID!) {
  portfolio(id: $id) {
    holdings {
      symbol
      quantity
      currentValue
    }
  }
}
```

## Testing

### cURL Example
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"query": "query GetPortfolio($id: ID!) { portfolio(id: $id) { id name holdings { symbol quantity averagePrice currentPrice currentValue gainLoss gainLossPercentage currency } } }", "variables": {"id": "0fdc618a-0cd2-4c04-aec4-44646c813e6a"}}' \
  http://localhost:8080/query
```

### Example Response (with holdings)
```json
{
  "data": {
    "portfolio": {
      "id": "user-123",
      "name": "My Portfolio",
      "holdings": [
        {
          "symbol": "AAPL",
          "quantity": 10.5,
          "averagePrice": 150.25,
          "currentPrice": 175.50,
          "currentValue": 1842.75,
          "gainLoss": 265.125,
          "gainLossPercentage": 16.82,
          "currency": "USD"
        },
        {
          "symbol": "GOOGL",
          "quantity": 5.0,
          "averagePrice": 2800.50,
          "currentPrice": 2950.00,
          "currentValue": 14750.00,
          "gainLoss": 747.50,
          "gainLossPercentage": 5.34,
          "currency": "USD"
        }
      ]
    }
  }
}
```

### Example Response (empty holdings)
```json
{
  "data": {
    "portfolio": {
      "id": "0fdc618a-0cd2-4c04-aec4-44646c813e6a",
      "name": "My Portfolio",
      "holdings": []
    }
  }
}
```

## Use Cases

### 1. Portfolio Dashboard
Display comprehensive holding information including:
- Current positions
- Cost basis (average price)
- Current market value
- Unrealized gains/losses

### 2. Performance Analysis
Calculate and display:
- Individual holding performance
- Best/worst performers
- Total portfolio gains/losses

### 3. Tax Reporting
Provide data for:
- Cost basis reporting
- Capital gains calculations
- Holding period analysis

### 4. Trading Decisions
Support decision-making with:
- Current vs. purchase price comparison
- Percentage gains/losses
- Position sizing information

## Frontend Integration

### React Example
```typescript
interface Holding {
  symbol: string;
  quantity: number;
  averagePrice: number;
  currentPrice: number;
  currentValue: number;
  gainLoss: number;
  gainLossPercentage: number;
  currency: string;
}

const HoldingsTable: React.FC<{ holdings: Holding[] }> = ({ holdings }) => {
  return (
    <table>
      <thead>
        <tr>
          <th>Symbol</th>
          <th>Quantity</th>
          <th>Avg Price</th>
          <th>Current Price</th>
          <th>Current Value</th>
          <th>Gain/Loss</th>
          <th>Gain/Loss %</th>
          <th>Currency</th>
        </tr>
      </thead>
      <tbody>
        {holdings.map((holding) => (
          <tr key={holding.symbol}>
            <td>{holding.symbol}</td>
            <td>{holding.quantity.toFixed(2)}</td>
            <td>{holding.currency} {holding.averagePrice.toFixed(2)}</td>
            <td>{holding.currency} {holding.currentPrice.toFixed(2)}</td>
            <td>{holding.currency} {holding.currentValue.toFixed(2)}</td>
            <td className={holding.gainLoss >= 0 ? 'positive' : 'negative'}>
              {holding.currency} {holding.gainLoss.toFixed(2)}
            </td>
            <td className={holding.gainLossPercentage >= 0 ? 'positive' : 'negative'}>
              {holding.gainLossPercentage.toFixed(2)}%
            </td>
            <td>{holding.currency}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
};
```

### GraphQL Query Hook
```typescript
import { useQuery, gql } from '@apollo/client';

const GET_PORTFOLIO = gql`
  query GetPortfolio($id: ID!) {
    portfolio(id: $id) {
      id
      name
      summary {
        totalValue
        totalGainLoss
        totalGainLossPercentage
        currency
      }
      holdings {
        symbol
        quantity
        averagePrice
        currentPrice
        currentValue
        gainLoss
        gainLossPercentage
        currency
      }
    }
  }
`;

function usePortfolio(portfolioId: string) {
  const { data, loading, error } = useQuery(GET_PORTFOLIO, {
    variables: { id: portfolioId },
  });

  return {
    portfolio: data?.portfolio,
    loading,
    error,
  };
}
```

## Calculations

### Gain/Loss Calculation
```
gainLoss = currentValue - (quantity × averagePrice)
gainLoss = (quantity × currentPrice) - (quantity × averagePrice)
gainLoss = quantity × (currentPrice - averagePrice)
```

### Gain/Loss Percentage Calculation
```
gainLossPercentage = (gainLoss / (quantity × averagePrice)) × 100
gainLossPercentage = ((currentPrice - averagePrice) / averagePrice) × 100
```

## Files Modified

1. **`apps/gateway/graph/schema.graphqls`**
   - Expanded `Holding` type with 5 new fields

2. **`apps/gateway/graph/schema.resolvers.go`**
   - Updated holdings mapping to include all proto fields

3. **`apps/gateway/graph/model/models_gen.go`** (auto-generated)
   - Generated updated `Holding` struct

4. **`docs/graphql_queries.postman_collection.json`**
   - Updated "Get Portfolio" query with all holding fields

## Backward Compatibility

### Breaking Change
⚠️ **This is a breaking change** - the `value` field has been replaced with `currentValue`.

Clients using the old schema will need to update their queries:
- Old: `holdings { value }`
- New: `holdings { currentValue }`

### Migration Guide
1. Update all GraphQL queries to use `currentValue` instead of `value`
2. Update TypeScript/JavaScript types to include new fields
3. Update UI components to display new fields
4. Test thoroughly before deploying to production

## Future Enhancements

1. **Lot Tracking**: Add support for tracking individual purchase lots
2. **Realized Gains**: Add fields for realized gains/losses from sales
3. **Holding Period**: Add purchase date and holding period
4. **Dividend Data**: Add dividend yield and total dividends received
5. **Cost Basis Methods**: Support FIFO, LIFO, and specific lot identification
6. **Multi-Currency**: Enhanced support for holdings in different currencies
7. **Historical Data**: Add time-series data for holding value over time

## Related Documentation

- [GraphQL Portfolio Summary](./GRAPHQL_PORTFOLIO_SUMMARY.md)
- [GraphQL User Integration](./GRAPHQL_USER_INTEGRATION.md)
- [Portfolio Service Proto](../proto/portfolio/portfolio.proto)
