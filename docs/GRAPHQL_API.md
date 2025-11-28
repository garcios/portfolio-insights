# GraphQL Gateway API Documentation

## Overview

The GraphQL Gateway provides a unified API interface to all Portfolio Insights microservices. It's built using [gqlgen](https://gqlgen.com/) and runs on port `8080` by default.

## Quick Start

### Starting the Gateway

```bash
# From the project root
make podman-up

# Or run the gateway directly
cd apps/gateway
go run cmd/server/main.go
```

### Accessing the API

- **GraphQL Endpoint**: `http://localhost:8080/query`
- **GraphQL Playground**: `http://localhost:8080/`

## Using the Postman Collection

### Import the Collection

1. Open Postman
2. Click **Import** button
3. Select the file: `docs/graphql_gateway.postman_collection.json`
4. The collection will be imported with all queries, mutations, and examples

### Collection Structure

The collection is organized into the following folders:

#### 1. **Queries**
- `Get Current User (me)` - Retrieve authenticated user information
- `Get Portfolio by ID` - Fetch a specific portfolio with all holdings
- `Get Portfolio with Minimal Fields` - Demonstrates selective field retrieval
- `Get User and Portfolio Combined` - Multiple queries in one request

#### 2. **Mutations**
- `Create User` - Create a new user account
- `Create User - Minimal Response` - Create user with selective response fields

#### 3. **Introspection**
- `Get Schema Types` - List all types in the schema
- `Get Query Type Fields` - Discover available queries
- `Get Mutation Type Fields` - Discover available mutations
- `Get Type Details - User` - Detailed User type information
- `Get Type Details - Portfolio` - Detailed Portfolio type information

#### 4. **Error Handling Examples**
- `Query with Missing Required Variable` - Validation error example
- `Query with Invalid Field` - Schema validation error
- `Mutation with Invalid Input` - Input validation error

#### 5. **Health Check**
- `GraphQL Playground` - Open the interactive playground

### Environment Variables

The collection uses the following variable:

| Variable | Default Value | Description |
|----------|---------------|-------------|
| `gateway_url` | `http://localhost:8080` | Base URL for the GraphQL Gateway |

To change the gateway URL:
1. Click on the collection name
2. Go to **Variables** tab
3. Update the `gateway_url` value

## GraphQL Schema

### Types

#### User
```graphql
type User {
  id: ID!
  username: String!
  email: String!
}
```

#### Portfolio
```graphql
type Portfolio {
  id: ID!
  userId: ID!
  name: String!
  holdings: [Holding!]!
}
```

#### Holding
```graphql
type Holding {
  symbol: String!
  quantity: Float!
  value: Float!
}
```

#### PortfolioPerformancePoint
```graphql
type PortfolioPerformancePoint {
  timestamp: String!
  value: Float!
}
```

Represents a single data point in portfolio performance history.

- **timestamp**: ISO 8601 formatted date-time string (e.g., "2023-11-01T00:00:00Z")
- **value**: Total portfolio value at that point in time


### Queries

#### me
Retrieves the currently authenticated user.

```graphql
query GetCurrentUser {
  me {
    id
    username
    email
  }
}
```

**Response:**
```json
{
  "data": {
    "me": {
      "id": "user-123",
      "username": "johndoe",
      "email": "john.doe@example.com"
    }
  }
}
```

#### portfolio(id: ID!)
Retrieves a specific portfolio by ID.

```graphql
query GetPortfolio($id: ID!) {
  portfolio(id: $id) {
    id
    userId
    name
    holdings {
      symbol
      quantity
      value
    }
  }
}
```

**Variables:**
```json
{
  "id": "portfolio-123"
}
```

**Response:**
```json
{
  "data": {
    "portfolio": {
      "id": "portfolio-123",
      "userId": "user-123",
      "name": "My Investment Portfolio",
      "holdings": [
        {
          "symbol": "AAPL",
          "quantity": 10,
          "value": 1750.50
        },
        {
          "symbol": "GOOGL",
          "quantity": 5,
          "value": 750.25
        }
      ]
    }
  }
}
```

#### portfolioPerformance(userId: ID!, period: String!)
Retrieves historical portfolio performance data over a specified time period.

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
  "period": "1m"
}
```

**Supported Periods:**
- `"1d"` - Last 1 day
- `"1w"` - Last 1 week
- `"1m"` - Last 1 month
- `"3m"` - Last 3 months
- `"1y"` - Last 1 year
- `"all"` - All available history

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

**Use Cases:**

1. **Performance Chart**: Display portfolio value over time
   ```graphql
   query PortfolioChart($userId: ID!) {
     portfolioPerformance(userId: $userId, period: "1m") {
       timestamp
       value
     }
   }
   ```

2. **Year-to-Date Performance**: Track annual performance
   ```graphql
   query YTDPerformance($userId: ID!) {
     portfolioPerformance(userId: $userId, period: "1y") {
       timestamp
       value
     }
   }
   ```

3. **All-Time Performance**: View complete portfolio history
   ```graphql
   query AllTimePerformance($userId: ID!) {
     portfolioPerformance(userId: $userId, period: "all") {
       timestamp
       value
     }
   }
   ```

**Notes:**
- Data points are returned in chronological order
- Timestamps are in ISO 8601 format
- Values represent total portfolio value in the default currency
- Empty array is returned if no historical data exists for the period


### Mutations

#### createUser(input: NewUser!)
Creates a new user account.

```graphql
mutation CreateUser($input: NewUser!) {
  createUser(input: $input) {
    id
    username
    email
  }
}
```

**Variables:**
```json
{
  "input": {
    "username": "johndoe",
    "email": "john.doe@example.com"
  }
}
```

**Response:**
```json
{
  "data": {
    "createUser": {
      "id": "user-456",
      "username": "johndoe",
      "email": "john.doe@example.com"
    }
  }
}
```

## GraphQL Best Practices

### 1. Request Only What You Need
GraphQL allows you to specify exactly which fields you want:

```graphql
# Good - Only request needed fields
query GetPortfolioName($id: ID!) {
  portfolio(id: $id) {
    name
  }
}

# Avoid - Requesting all fields when not needed
query GetPortfolio($id: ID!) {
  portfolio(id: $id) {
    id
    userId
    name
    holdings {
      symbol
      quantity
      value
    }
  }
}
```

### 2. Use Variables for Dynamic Values
Always use variables instead of inline values:

```graphql
# Good
query GetPortfolio($id: ID!) {
  portfolio(id: $id) {
    name
  }
}

# Avoid
query GetPortfolio {
  portfolio(id: "portfolio-123") {
    name
  }
}
```

### 3. Name Your Operations
Always name your queries and mutations:

```graphql
# Good
query GetUserProfile {
  me {
    username
  }
}

# Avoid
{
  me {
    username
  }
}
```

### 4. Batch Multiple Queries
Combine multiple queries in a single request:

```graphql
query GetUserAndPortfolio($portfolioId: ID!) {
  me {
    id
    username
  }
  portfolio(id: $portfolioId) {
    name
    holdings {
      symbol
    }
  }
}
```

## Error Handling

GraphQL returns errors in a standardized format:

```json
{
  "errors": [
    {
      "message": "Variable \"$id\" of required type \"ID!\" was not provided.",
      "locations": [
        {
          "line": 1,
          "column": 23
        }
      ],
      "extensions": {
        "code": "GRAPHQL_VALIDATION_FAILED"
      }
    }
  ]
}
```

### Common Error Types

1. **Validation Errors** - Schema or input validation failures
2. **Resolver Errors** - Errors from resolver functions
3. **Network Errors** - Connection or timeout issues

## Testing with cURL

You can also test the GraphQL API using cURL:

```bash
# Query example
curl -X POST http://localhost:8080/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query { me { id username email } }"
  }'

# Mutation with variables
curl -X POST http://localhost:8080/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation CreateUser($input: NewUser!) { createUser(input: $input) { id username email } }",
    "variables": {
      "input": {
        "username": "johndoe",
        "email": "john.doe@example.com"
      }
    }
  }'
```

## Development

### Regenerating GraphQL Code

When you modify the schema (`apps/gateway/graph/schema.graphqls`), regenerate the code:

```bash
cd apps/gateway
go run github.com/99designs/gqlgen generate
```

### Adding New Queries or Mutations

1. Update `apps/gateway/graph/schema.graphqls`
2. Run `go run github.com/99designs/gqlgen generate`
3. Implement the resolver in `apps/gateway/graph/schema.resolvers.go`
4. Update this Postman collection with new examples

## Additional Resources

- [GraphQL Official Documentation](https://graphql.org/)
- [gqlgen Documentation](https://gqlgen.com/)
- [GraphQL Best Practices](https://graphql.org/learn/best-practices/)
- [Postman GraphQL Documentation](https://learning.postman.com/docs/sending-requests/supported-api-frameworks/graphql/)

## Support

For issues or questions:
1. Check the GraphQL Playground at `http://localhost:8080/`
2. Review the schema using introspection queries
3. Check service logs for detailed error messages
