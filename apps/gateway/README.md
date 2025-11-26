# GraphQL Gateway – Portfolio Service

This README explains how to use the **GraphQL gateway** to query a user's portfolio holdings.

## Overview
The gateway (`apps/gateway`) exposes a GraphQL API that forwards portfolio‑related requests to the **portfolio‑service** via gRPC. The key query we expose is `portfolio(id: ID!)` which returns a `Portfolio` object containing the user's holdings.

## Running the Gateway
```bash
# Start all services (including the gateway)
make podman-up
```
The gateway listens on **`http://localhost:8080/`**. You can open the GraphQL Playground at that URL.

## Environment Variables
| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP port for the GraphQL server | `8080` |
| `PORTFOLIO_SERVICE_ADDR` | Address of the portfolio‑service gRPC endpoint | `localhost:50052` |

## GraphQL Schema (relevant parts)
```graphql
type Query {
  portfolio(id: ID!): Portfolio
}

type Portfolio {
  id: ID!
  userId: ID!
  name: String!
  holdings: [Holding!]!
}

type Holding {
  symbol: String!
  quantity: Float!
  value: Float!
}
```

## Example Query
```graphql
query GetMyPortfolio {
  portfolio(id: "user-123") {
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
The resolver calls `portfolio-service`'s `GetHoldings` RPC, converts the protobuf response to the GraphQL types, and returns the data.

## Expected Response
```json
{
  "data": {
    "portfolio": {
      "id": "user-123",
      "userId": "user-123",
      "name": "My Portfolio",
      "holdings": [
        {"symbol": "AAPL", "quantity": 10, "value": 1500.0},
        {"symbol": "GOOGL", "quantity": 5, "value": 2800.0}
      ]
    }
  }
}
```
If the gRPC call fails, the resolver returns an error which appears in the GraphQL response under `errors`.

## Troubleshooting
- **Connection errors** – Ensure `PORTFOLIO_SERVICE_ADDR` points to a running portfolio‑service container.
- **Missing holdings** – Verify that the portfolio‑service has data for the given user ID.
- **Container issues** – Run `make podman-down && make podman-up` to restart the stack.

---
*This file was generated to help developers quickly understand and use the GraphQL gateway.*
