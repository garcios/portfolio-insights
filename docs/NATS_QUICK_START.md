# Quick Start: Testing NATS Integration

This guide will help you test the NATS event-driven architecture between the transaction and portfolio services.

## Prerequisites

- Docker and Docker Compose (or Podman)
- gRPC client tool (e.g., `grpcurl` or BloomRPC)

## Step 1: Start All Services

```bash
cd /Users/oscargarcia/Documents/workspace/portfolio-insights
make podman-up
# or
cd deployments/docker-compose
docker-compose up
```

Wait for all services to start. You should see:
- ✅ `Portfolio Service is running and listening for events...`
- ✅ `Transaction Service listening on port 50053`
- ✅ `Subscribed to NATS topic: transaction-service.transaction.created`

## Step 2: Install grpcurl (if not already installed)

```bash
# macOS
brew install grpcurl

# Or download from: https://github.com/fullstorydev/grpcurl/releases
```

## Step 3: Create a User (if needed)

First, ensure you have a user in the system:

```bash
grpcurl -plaintext \
  -d '{
    "name": "Test User",
    "email": "test@example.com"
  }' \
  localhost:50051 \
  user.UserService/CreateUser
```

Note the `user_id` from the response.

## Step 4: Create an Asset (if needed)

Ensure the asset exists in the market data service:

```bash
grpcurl -plaintext \
  -d '{
    "symbol": "AAPL",
    "name": "Apple Inc.",
    "asset_type": "STOCK"
  }' \
  localhost:50054 \
  marketdata.MarketDataService/CreateAsset
```

## Step 5: Create a BUY Transaction

```bash
grpcurl -plaintext \
  -d '{
    "user_id": "YOUR_USER_ID_HERE",
    "symbol": "AAPL",
    "type": "BUY",
    "quantity": 10,
    "price_per_share": 150.50,
    "executed_at": "2025-11-22T10:00:00Z"
  }' \
  localhost:50053 \
  transaction.TransactionService/CreateTransaction
```

## Step 6: Monitor the Logs

### Transaction Service Logs
```bash
docker-compose logs -f transaction-service
```

You should see:
```
INFO Transaction Service listening on port 50053
INFO Creating transaction...
INFO Transaction created successfully
```

### Portfolio Service Logs
```bash
docker-compose logs -f portfolio-service
```

You should see:
```
INFO Subscribed to NATS topic topic=transaction-service.transaction.created
INFO Received transaction created event transaction_id=xxx user_id=xxx symbol=AAPL type=BUY quantity=10
INFO Updated portfolio holding user_id=xxx symbol=AAPL new_quantity=10 average_cost=150.5
```

## Step 7: Create Another BUY Transaction

```bash
grpcurl -plaintext \
  -d '{
    "user_id": "YOUR_USER_ID_HERE",
    "symbol": "AAPL",
    "type": "BUY",
    "quantity": 5,
    "price_per_share": 155.00,
    "executed_at": "2025-11-22T11:00:00Z"
  }' \
  localhost:50053 \
  transaction.TransactionService/CreateTransaction
```

Check portfolio service logs - you should see the average cost recalculated:
```
INFO Updated portfolio holding user_id=xxx symbol=AAPL new_quantity=15 average_cost=152
```

**Calculation:**
- Total Cost = (10 × $150.50) + (5 × $155.00) = $2,280
- New Quantity = 10 + 5 = 15
- New Average Cost = $2,280 ÷ 15 = $152.00

## Step 8: Create a SELL Transaction

```bash
grpcurl -plaintext \
  -d '{
    "user_id": "YOUR_USER_ID_HERE",
    "symbol": "AAPL",
    "type": "SELL",
    "quantity": 5,
    "price_per_share": 160.00,
    "executed_at": "2025-11-22T12:00:00Z"
  }' \
  localhost:50053 \
  transaction.TransactionService/CreateTransaction
```

Check portfolio service logs:
```
INFO Updated portfolio holding user_id=xxx symbol=AAPL new_quantity=10 average_cost=152
```

Note: Average cost remains $152 (unchanged on SELL), but quantity decreased to 10.

## Step 9: Monitor NATS Directly (Optional)

If you have the NATS CLI installed:

```bash
# Install NATS CLI
brew install nats-io/nats-tools/nats

# Subscribe to all transaction events
nats sub "transaction-service.>"
```

Then create a transaction and watch the event appear in real-time!

## Troubleshooting

### Events not appearing in portfolio service

1. **Check NATS is running:**
   ```bash
   docker-compose ps nats
   ```

2. **Check NATS logs:**
   ```bash
   docker-compose logs nats
   ```

3. **Verify environment variables:**
   ```bash
   docker-compose exec transaction-service env | grep NATS
   docker-compose exec portfolio-service env | grep NATS
   ```

### Transaction creation fails

1. **Check user exists:**
   ```bash
   grpcurl -plaintext localhost:50051 user.UserService/ListUsers
   ```

2. **Check asset exists:**
   ```bash
   grpcurl -plaintext \
     -d '{"symbol": "AAPL"}' \
     localhost:50054 \
     marketdata.MarketDataService/GetAsset
   ```

3. **Check transaction service logs:**
   ```bash
   docker-compose logs transaction-service
   ```

### Portfolio service not starting

1. **Check for port conflicts:**
   ```bash
   lsof -i :50052
   ```

2. **Check service logs:**
   ```bash
   docker-compose logs portfolio-service
   ```

## Expected Behavior Summary

| Action | Transaction Service | NATS | Portfolio Service |
|--------|-------------------|------|------------------|
| Create BUY transaction | ✅ Saves to DB<br>✅ Publishes event | ✅ Routes event | ✅ Receives event<br>✅ Increases quantity<br>✅ Recalculates avg cost |
| Create SELL transaction | ✅ Saves to DB<br>✅ Publishes event | ✅ Routes event | ✅ Receives event<br>✅ Decreases quantity<br>✅ Keeps avg cost same |

## Next Steps

1. **Add gRPC endpoint** to portfolio service to query holdings
2. **Persist holdings** to PostgreSQL instead of in-memory
3. **Add monitoring** dashboard to visualize event flow
4. **Enable NATS JetStream** for guaranteed delivery
5. **Add integration tests** that verify end-to-end flow

## Useful Commands

```bash
# View all running services
docker-compose ps

# Restart a specific service
docker-compose restart portfolio-service

# View logs for all services
docker-compose logs -f

# Stop all services
docker-compose down

# Rebuild and restart
docker-compose up --build
```

## Success Criteria

✅ Transaction service publishes events to NATS
✅ Portfolio service receives and processes events
✅ Holdings are updated correctly for BUY transactions
✅ Holdings are updated correctly for SELL transactions
✅ Average cost is calculated properly
✅ No errors in service logs
✅ System handles multiple transactions for same user/symbol
