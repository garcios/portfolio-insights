# GraphQL Gateway Testing Guide

## Overview
This guide explains how to test the GraphQL Gateway API using the Postman collection.

## Files Created
- **`docs/graphql_queries.postman_collection.json`**: Postman collection with GraphQL queries and mutations

## Issue Fixed
The original error was:
```
rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing: dial tcp [::1]:50052: connect: connection refused"
```

**Root Cause**: The gateway service was trying to connect to `localhost:50052`, but inside the Docker container, the portfolio-service is accessible at `portfolio-service:50052`.

**Solution**: Added the `PORTFOLIO_SERVICE_ADDR` environment variable to the gateway service in `docker-compose.yml`:
```yaml
gateway:
  environment:
    - PORTFOLIO_SERVICE_ADDR=portfolio-service:50052
```

## Test Data Setup
Created test data in the database:
- **User ID**: `0fdc618a-0cd2-4c04-aec4-44646c813e6a`
- **Email**: `test@example.com`
- **Holdings**: 
  - AAPL: 10.5 shares @ $150.25 avg cost
  - GOOGL: 5.0 shares @ $2800.50 avg cost

## Available Queries

### 1. Get Portfolio
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

**Variables**:
```json
{
  "id": "0fdc618a-0cd2-4c04-aec4-44646c813e6a"
}
```

### 2. Get Current User
```graphql
query {
  me {
    id
    username
    email
  }
}
```

### 3. Create User
```graphql
mutation CreateUser($input: NewUser!) {
  createUser(input: $input) {
    id
    username
    email
  }
}
```

**Variables**:
```json
{
  "input": {
    "username": "testuser",
    "email": "test@example.com"
  }
}
```

## Testing with cURL

```bash
# Get Portfolio
curl -X POST -H "Content-Type: application/json" \
  -d '{"query": "query GetPortfolio($id: ID!) { portfolio(id: $id) { id userId name holdings { symbol quantity value } } }", "variables": {"id": "0fdc618a-0cd2-4c04-aec4-44646c813e6a"}}' \
  http://localhost:8080/query

# Expected Response:
# {"data":{"portfolio":{"id":"0fdc618a-0cd2-4c04-aec4-44646c813e6a","userId":"0fdc618a-0cd2-4c04-aec4-44646c813e6a","name":"My Portfolio","holdings":[]}}}
```

## GraphQL Playground
Access the interactive GraphQL Playground at: **http://localhost:8080/**

## Notes
- The gateway is now successfully connecting to the portfolio-service
- Holdings may appear empty if the market data service hasn't populated price information
- All queries use the `/query` endpoint
- The base URL is `http://localhost:8080`

## Importing to Postman
1. Open Postman
2. Click **Import**
3. Select `docs/graphql_queries.postman_collection.json`
4. The collection will include all queries with example variables
