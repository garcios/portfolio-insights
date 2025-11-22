# NATS Event-Driven Architecture

This document describes the NATS-based event-driven architecture implemented for the portfolio tracking system.

## Overview

The system uses NATS as a message broker to enable asynchronous communication between microservices. When a transaction is created, the `transaction-service` publishes an event that the `portfolio-service` subscribes to and processes.

## Event Flow

```
User -> Transaction Service -> NATS -> Portfolio Service
         (Creates Transaction)  (Publishes Event)  (Updates Holdings)
```

## Event Schema

### Transaction Created Event

**Topic:** `transaction-service.transaction.created`

**Payload:**
```json
{
  "transaction_id": "string",
  "user_id": "string",
  "asset_symbol": "string",
  "price_per_share": 150.50,
  "quantity": 10.0,
  "type": "BUY|SELL",
  "executed_at": "2025-11-22T10:00:00Z"
}
```

## Services

### Transaction Service (Publisher)

**Location:** `/services/transaction-service/internal/infrastructure/nats_publisher.go`

- Publishes `TransactionCreatedEvent` after successfully creating a transaction
- Uses NATS connection configured via `NATS_URL` environment variable
- Default URL: `nats://nats:4222`

**Key Components:**
- `EventPublisher` interface in domain layer
- `natsEventPublisher` implementation in infrastructure layer
- Integrated into `CreateTransaction` usecase

### Portfolio Service (Subscriber)

**Location:** `/services/portfolio-service/internal/infrastructure/nats_subscriber.go`

- Subscribes to `transaction-service.transaction.created` topic
- Updates portfolio holdings based on transaction events
- Calculates average cost for BUY transactions
- Reduces quantity for SELL transactions

**Key Components:**
- `NATSSubscriber` struct that manages subscription lifecycle
- `handleTransactionCreated` method that processes events
- In-memory repository for storing holdings (can be replaced with database)

**Portfolio Update Logic:**
- **BUY transactions:** Increases quantity and recalculates average cost
  ```
  new_avg_cost = (old_quantity * old_avg_cost + new_quantity * new_price) / total_quantity
  ```
- **SELL transactions:** Decreases quantity, maintains average cost

## Configuration

### Environment Variables

Both services require the `NATS_URL` environment variable:

```bash
NATS_URL=nats://nats:4222
```

### Docker Compose

The NATS server is configured in `docker-compose.yml`:

```yaml
nats:
  image: nats:latest
  command: "-js"  # Enable JetStream
  ports:
    - "4222:4222"  # Client connections
    - "8222:8222"  # HTTP monitoring
```

## Running the System

1. **Start all services:**
   ```bash
   cd deployments/docker-compose
   docker-compose up
   ```

2. **Create a transaction:**
   The transaction service will automatically publish an event to NATS.

3. **Monitor portfolio service logs:**
   You should see log messages indicating the event was received and the portfolio was updated.

## Testing

### Transaction Service Tests

The transaction service includes mock event publishers for testing:

```bash
cd services/transaction-service
go test ./internal/usecase/...
```

### Manual Testing

You can use the NATS CLI to monitor events:

```bash
# Subscribe to all transaction events
nats sub "transaction-service.>"

# Publish a test event
nats pub transaction-service.transaction.created '{"transaction_id":"test","user_id":"user1","asset_symbol":"AAPL","price_per_share":150.0,"quantity":10.0,"type":"BUY","executed_at":"2025-11-22T10:00:00Z"}'
```

## Future Enhancements

1. **Persistent Storage:** Replace in-memory repository with PostgreSQL
2. **Event Sourcing:** Store all events for audit trail
3. **Dead Letter Queue:** Handle failed event processing
4. **Event Replay:** Rebuild portfolio state from transaction history
5. **NATS JetStream:** Add persistence and guaranteed delivery
6. **Monitoring:** Add metrics for event publishing/processing
7. **Schema Validation:** Use Protocol Buffers or JSON Schema for event validation

## Troubleshooting

### Events not being received

1. Check NATS connection:
   ```bash
   docker-compose logs nats
   ```

2. Verify NATS_URL environment variable is set correctly

3. Check service logs for connection errors

### Portfolio not updating

1. Verify the event payload matches the expected schema
2. Check portfolio service logs for processing errors
3. Ensure the repository is functioning correctly

## Architecture Benefits

- **Loose Coupling:** Services don't need to know about each other
- **Scalability:** Multiple portfolio service instances can subscribe to events
- **Resilience:** If portfolio service is down, events can be queued (with JetStream)
- **Extensibility:** Easy to add new subscribers for other features (notifications, analytics, etc.)
