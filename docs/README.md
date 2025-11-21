# Portfolio Insights - API Collections

This directory contains Postman collections and documentation for testing the Portfolio Insights API.

## Available Collections

### 1. GraphQL Gateway Collection
**File**: `graphql_gateway.postman_collection.json`  
**Documentation**: `GRAPHQL_API.md`  
**Service**: GraphQL Gateway (Port 8080)

This collection provides comprehensive testing for the GraphQL Gateway API, which serves as the unified entry point for all microservices.

**Includes**:
- ✅ All GraphQL queries (me, portfolio)
- ✅ All GraphQL mutations (createUser)
- ✅ Schema introspection queries
- ✅ Error handling examples
- ✅ Health check endpoints

**Use this when**: You want to test the GraphQL API or need a unified interface to query multiple services.

---

### 2. gRPC Services Collection
**File**: `portfolio_insights.postman_collection.json`  
**Service**: Direct gRPC microservices

This collection provides direct access to the underlying gRPC microservices.

**Includes**:
- MarketData Service (Port 50051)
- Portfolio Service (Port 50052)
- Transaction Service (Port 50053)
- User Service (Port 50054)

**Use this when**: You need to test individual microservices directly or debug service-specific issues.

---

## Quick Start

### Import Collections into Postman

1. Open Postman
2. Click **Import** button
3. Select one or both collection files:
   - `graphql_gateway.postman_collection.json`
   - `portfolio_insights.postman_collection.json`
4. Collections will appear in your Postman workspace

### Start the Services

```bash
# Start all services using Podman
make podman-up

# Or start individual services
cd apps/gateway && go run cmd/server/main.go
```

### Test the APIs

#### GraphQL Gateway
```bash
# Open GraphQL Playground in browser
open http://localhost:8080/

# Or use curl
curl -X POST http://localhost:8080/query \
  -H "Content-Type: application/json" \
  -d '{"query": "{ me { id username email } }"}'
```

#### gRPC Services
Use the Postman collection or gRPC clients to test individual services.

---

## Service Endpoints

| Service | Type | Port | Endpoint |
|---------|------|------|----------|
| GraphQL Gateway | HTTP/GraphQL | 8080 | http://localhost:8080/query |
| GraphQL Playground | HTTP | 8080 | http://localhost:8080/ |
| MarketData Service | gRPC | 50051 | localhost:50051 |
| Portfolio Service | gRPC | 50052 | localhost:50052 |
| Transaction Service | gRPC | 50053 | localhost:50053 |
| User Service | gRPC | 50054 | localhost:50054 |

---

## Documentation

- **GraphQL API**: See [GRAPHQL_API.md](./GRAPHQL_API.md) for detailed GraphQL documentation
- **Podman Setup**: See [PODMAN_SETUP.md](./PODMAN_SETUP.md) for container setup instructions

---

## Recommended Workflow

### For Frontend Development
👉 **Use the GraphQL Gateway Collection**
- Provides a clean, unified API
- Easy to query multiple services in one request
- Better for frontend integration

### For Backend Development
👉 **Use both collections**
- GraphQL collection for end-to-end testing
- gRPC collection for debugging individual services

### For API Exploration
👉 **Use the GraphQL Playground**
- Interactive interface at http://localhost:8080/
- Auto-complete and documentation
- Easy to experiment with queries

---

## Collection Features

### GraphQL Gateway Collection Features

✨ **Organized by Operation Type**
- Queries folder
- Mutations folder
- Introspection queries
- Error handling examples

✨ **Complete Examples**
- All queries with sample variables
- Multiple variations (full vs minimal fields)
- Combined queries (multiple operations in one request)

✨ **Developer-Friendly**
- Detailed descriptions for each request
- Example responses
- Best practices documentation

### gRPC Collection Features

✨ **Service-Organized**
- Separate folders for each microservice
- All RPC methods included

✨ **Ready-to-Use**
- Pre-configured endpoints
- Sample request bodies
- Environment variables for easy configuration

---

## Environment Variables

### GraphQL Gateway Collection
| Variable | Default | Description |
|----------|---------|-------------|
| `gateway_url` | http://localhost:8080 | GraphQL Gateway base URL |

### gRPC Collection
| Variable | Default | Description |
|----------|---------|-------------|
| `marketdata_url` | http://localhost:50051 | MarketData Service URL |
| `portfolio_url` | http://localhost:50052 | Portfolio Service URL |
| `transaction_url` | http://localhost:50053 | Transaction Service URL |
| `user_url` | http://localhost:50054 | User Service URL |

---

## Troubleshooting

### Services Not Running
```bash
# Check if services are running
make podman-ps

# View logs
make podman-logs

# Restart services
make podman-down
make podman-up
```

### GraphQL Errors
1. Check the GraphQL Playground at http://localhost:8080/
2. Use introspection queries to verify schema
3. Review error messages in the response

### gRPC Connection Issues
1. Verify service ports are not in use
2. Check service logs for errors
3. Ensure gRPC services are properly started

---

## Contributing

When adding new API endpoints:

1. Update the GraphQL schema (`apps/gateway/graph/schema.graphqls`)
2. Regenerate GraphQL code: `cd apps/gateway && go run github.com/99designs/gqlgen generate`
3. Add new requests to the appropriate Postman collection
4. Update documentation in `GRAPHQL_API.md`
5. Test all requests before committing

---

## Additional Resources

- [GraphQL Documentation](https://graphql.org/)
- [Postman GraphQL Guide](https://learning.postman.com/docs/sending-requests/supported-api-frameworks/graphql/)
- [gRPC Documentation](https://grpc.io/docs/)
- [gqlgen Documentation](https://gqlgen.com/)
