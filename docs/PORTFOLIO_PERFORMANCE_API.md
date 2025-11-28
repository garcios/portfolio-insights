# Portfolio Performance API Documentation

## Overview

The Portfolio Performance API provides historical portfolio value data over time, enabling users to track their investment performance across different time periods. This feature is essential for visualizing portfolio growth, analyzing trends, and making informed investment decisions.

## Architecture

### Data Flow

```
GraphQL Gateway (portfolioPerformance query)
    ↓
Portfolio Service (GetPortfolioPerformance RPC)
    ↓
Portfolio History Repository
    ↓
PostgreSQL (investments.portfolio_history table)
```

### Components

1. **GraphQL Gateway** (`apps/gateway`)
   - Exposes the `portfolioPerformance` query
   - Handles request validation and response formatting
   - Located in: `apps/gateway/graph/schema.resolvers.go`

2. **Portfolio Service** (`services/portfolio-service`)
   - Implements the gRPC `GetPortfolioPerformance` endpoint
   - Queries historical snapshots from the database
   - Located in: `services/portfolio-service/internal/handler/grpc/portfolio_handler.go`

3. **Database**
   - Table: `investments.portfolio_history`
   - Stores daily portfolio value snapshots
   - Schema defined in: `infra/db/000004_create_portfolio_tables.up.sql`

## GraphQL API

### Query

```graphql
portfolioPerformance(userId: ID!, period: String!): [PortfolioPerformancePoint!]!
```

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `userId` | ID | Yes | The user ID to retrieve performance data for |
| `period` | String | Yes | Time period for the data (see supported periods below) |

### Supported Periods

| Period | Description | Example Use Case |
|--------|-------------|------------------|
| `1d` | Last 1 day | Intraday performance tracking |
| `1w` | Last 1 week | Weekly performance review |
| `1m` | Last 1 month | Monthly performance analysis |
| `3m` | Last 3 months | Quarterly performance review |
| `1y` | Last 1 year | Annual performance tracking |
| `all` | All available history | Long-term performance analysis |

### Response Type

```graphql
type PortfolioPerformancePoint {
  timestamp: String!  # ISO 8601 format (e.g., "2023-11-01T00:00:00Z")
  value: Float!       # Total portfolio value in default currency
}
```

## Usage Examples

### Basic Query

```graphql
query GetMonthlyPerformance {
  portfolioPerformance(userId: "user-123", period: "1m") {
    timestamp
    value
  }
}
```

**Response:**
```json
{
  "data": {
    "portfolioPerformance": [
      {
        "timestamp": "2023-11-01T00:00:00Z",
        "value": 15234.50
      },
      {
        "timestamp": "2023-11-02T00:00:00Z",
        "value": 15456.75
      },
      {
        "timestamp": "2023-11-03T00:00:00Z",
        "value": 15123.25
      }
    ]
  }
}
```

### With Variables

```graphql
query GetPortfolioPerformance($userId: ID!, $period: String!) {
  portfolioPerformance(userId: $userId, period: $period) {
    timestamp
    value
  }
}
```

**Variables:**
```json
{
  "userId": "user-123",
  "period": "1y"
}
```

### Combined with Other Queries

```graphql
query PortfolioDashboard($userId: ID!) {
  portfolio(id: $userId) {
    id
    name
    summary {
      totalValue
      totalGainLoss
      totalGainLossPercentage
    }
  }
  portfolioPerformance(userId: $userId, period: "1m") {
    timestamp
    value
  }
}
```

## Frontend Integration

### React Example with Apollo Client

```typescript
import { useQuery, gql } from '@apollo/client';

const GET_PORTFOLIO_PERFORMANCE = gql`
  query GetPortfolioPerformance($userId: ID!, $period: String!) {
    portfolioPerformance(userId: $userId, period: $period) {
      timestamp
      value
    }
  }
`;

function PortfolioChart({ userId, period }) {
  const { loading, error, data } = useQuery(GET_PORTFOLIO_PERFORMANCE, {
    variables: { userId, period }
  });

  if (loading) return <p>Loading...</p>;
  if (error) return <p>Error: {error.message}</p>;

  const chartData = data.portfolioPerformance.map(point => ({
    date: new Date(point.timestamp),
    value: point.value
  }));

  return <LineChart data={chartData} />;
}
```

### Chart.js Integration

```javascript
const performanceData = data.portfolioPerformance;

const chartConfig = {
  type: 'line',
  data: {
    labels: performanceData.map(p => new Date(p.timestamp).toLocaleDateString()),
    datasets: [{
      label: 'Portfolio Value',
      data: performanceData.map(p => p.value),
      borderColor: 'rgb(75, 192, 192)',
      tension: 0.1
    }]
  },
  options: {
    responsive: true,
    scales: {
      y: {
        beginAtZero: false,
        ticks: {
          callback: function(value) {
            return '$' + value.toLocaleString();
          }
        }
      }
    }
  }
};
```

## Data Population

### Backfilling Historical Data

Historical portfolio snapshots are created using the `BackfillHistory` admin endpoint:

```bash
# Using Postman or gRPC client
POST /portfolio.PortfolioService/BackfillHistory
{
  "user_id": "user-123",
  "start_date": "2023-01-01",
  "end_date": "2023-12-31",
  "dry_run": false,
  "admin_token": "your-admin-token"
}
```

For detailed backfill documentation, see: [docs/postman/README_BACKFILL.md](./postman/README_BACKFILL.md)

### Automated Daily Snapshots

To keep performance data current, implement a scheduled job to create daily snapshots:

```go
// Example cron job (pseudo-code)
func dailySnapshotJob() {
    users := getAllActiveUsers()
    for _, user := range users {
        createSnapshot(user.ID, time.Now())
    }
}
```

## Performance Considerations

### Data Volume

- Each user gets one snapshot per day
- For 1000 users over 1 year: ~365,000 records
- Indexed on `(user_id, timestamp)` for fast queries

### Query Optimization

1. **Period Filtering**: The service filters data by period on the database level
2. **Indexing**: Composite index on `(user_id, timestamp)` ensures fast lookups
3. **Caching**: Consider implementing Redis caching for frequently accessed periods

### Recommended Practices

1. **Limit Data Points**: For long periods, consider aggregating data (e.g., weekly instead of daily for "all" period)
2. **Pagination**: For very large datasets, implement cursor-based pagination
3. **Client-Side Caching**: Cache performance data on the frontend with appropriate TTL

## Error Handling

### Empty Results

If no historical data exists for the requested period, an empty array is returned:

```json
{
  "data": {
    "portfolioPerformance": []
  }
}
```

### Invalid Period

If an unsupported period is provided, the service may return an error or default behavior. Always use documented period values.

### Missing User

If the user ID doesn't exist, an empty array is returned (not an error).

## Testing

### GraphQL Playground

1. Navigate to `http://localhost:8080/`
2. Use the following query:

```graphql
query TestPerformance {
  portfolioPerformance(userId: "user-123", period: "1m") {
    timestamp
    value
  }
}
```

### cURL Example

```bash
curl -X POST http://localhost:8080/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query GetPerformance($userId: ID!, $period: String!) { portfolioPerformance(userId: $userId, period: $period) { timestamp value } }",
    "variables": {
      "userId": "user-123",
      "period": "1m"
    }
  }'
```

### Postman Collection

Import the GraphQL Gateway collection from `docs/postman/graphql_gateway.postman_collection.json` for pre-configured examples.

## Database Schema

### portfolio_history Table

```sql
CREATE TABLE IF NOT EXISTS investments.portfolio_history (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    total_value DECIMAL(20, 2) NOT NULL,
    total_cost_basis DECIMAL(20, 2) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, timestamp)
);

CREATE INDEX idx_portfolio_history_user_timestamp 
ON investments.portfolio_history(user_id, timestamp DESC);
```

## Related Documentation

- [GraphQL API Documentation](./GRAPHQL_API.md)
- [Portfolio History Strategy](./PORTFOLIO_HISTORY_STRATEGY.md)
- [Backfill Admin Guide](./postman/README_BACKFILL.md)
- [Portfolio Service gRPC API](../proto/portfolio/portfolio.proto)

## Future Enhancements

### Planned Features

1. **Aggregated Data**: Support for weekly/monthly aggregations for long periods
2. **Performance Metrics**: Add gain/loss percentage to each data point
3. **Benchmark Comparison**: Compare portfolio performance against market indices
4. **Custom Date Ranges**: Allow arbitrary start/end dates instead of predefined periods
5. **Real-time Updates**: WebSocket support for live performance updates

### Contribution

To contribute to this feature:

1. Review the architecture diagram above
2. Check existing issues/PRs related to portfolio performance
3. Follow the development workflow in the main README
4. Add tests for any new functionality

## Support

For questions or issues:

1. Check the GraphQL Playground schema documentation
2. Review service logs for detailed error messages
3. Consult the [Portfolio History Strategy](./PORTFOLIO_HISTORY_STRATEGY.md) document
4. Open an issue on GitHub with reproduction steps
