# EODHD Price Sync - Implementation Checklist

## Phase 1: Core Infrastructure ✓

### 1.1 EODHD Client Package
- [ ] Create `internal/client/eodhd_client.go`
  - [ ] Define `EODHDClient` interface
  - [ ] Implement HTTP client with timeout
  - [ ] Add rate limiter (1 req/sec default)
  - [ ] Implement retry logic with exponential backoff
  - [ ] Handle 429 (rate limit) responses
  - [ ] Parse JSON responses to domain models

### 1.2 Domain Models
- [ ] Create `internal/domain/eodhd.go`
  - [ ] `RealTimePrice` struct
  - [ ] `HistoricalPrice` struct
  - [ ] `SyncJob` struct
  - [ ] `AssetSyncDetail` struct

### 1.3 Repository Enhancements
- [ ] Update `internal/repository/postgres_repo.go`
  - [ ] `GetAssetsRequiringPriceUpdate(staleDuration time.Duration) ([]*Asset, error)`
  - [ ] `GetLatestPriceTimestamp(assetID string) (*time.Time, error)`
  - [ ] `GetMissingPriceDates(assetID, start, end) ([]time.Time, error)`

### 1.4 Metrics
- [ ] Create `internal/metrics/eodhd_metrics.go`
  - [ ] `EODHDAPICallsTotal` counter
  - [ ] `EODHDAPILatency` histogram
  - [ ] `PricesSyncedTotal` counter
  - [ ] `SyncJobDuration` histogram
  - [ ] `SyncJobErrors` counter

## Phase 2: Price Sync Worker ✓

### 2.1 Worker Implementation
- [ ] Create `internal/worker/eodhd_price_sync.go`
  - [ ] `EODHDPriceSyncWorker` struct
  - [ ] `NewEODHDPriceSyncWorker()` constructor
  - [ ] `Start(ctx)` method with ticker
  - [ ] `syncPrices(ctx)` main sync logic
  - [ ] `syncAsset(ctx, asset)` per-asset sync
  - [ ] `determineMissingDates(asset)` helper
  - [ ] `fetchAndStorePrices(asset, dates)` helper

### 2.2 Configuration
- [ ] Add environment variables
  - [ ] `EODHD_API_TOKEN`
  - [ ] `EODHD_API_BASE_URL`
  - [ ] `EODHD_RATE_LIMIT`
  - [ ] `PRICE_SYNC_INTERVAL`
  - [ ] `PRICE_STALE_DURATION`
  - [ ] `PRICE_SYNC_BATCH_SIZE`
  - [ ] `PRICE_SYNC_MAX_CONCURRENCY`
  - [ ] `PRICE_SYNC_HISTORICAL_DAYS`

### 2.3 Integration
- [ ] Update `cmd/server/main.go`
  - [ ] Initialize EODHD client
  - [ ] Create price sync worker
  - [ ] Start worker in goroutine
  - [ ] Handle graceful shutdown

## Phase 3: HTTP API ✓

### 3.1 HTTP Server Setup
- [ ] Create `internal/handler/http/server.go`
  - [ ] HTTP server initialization
  - [ ] Router setup (chi/mux)
  - [ ] Middleware (logging, recovery, CORS)

### 3.2 Sync Endpoints
- [ ] Create `internal/handler/http/sync_handler.go`
  - [ ] `POST /api/v1/sync-prices` handler
  - [ ] `GET /api/v1/sync-prices/:job_id` handler
  - [ ] Request/response DTOs
  - [ ] Job tracking (in-memory map)
  - [ ] Async job execution

### 3.3 Health & Status
- [ ] Create `internal/handler/http/health_handler.go`
  - [ ] `GET /health` endpoint
  - [ ] `GET /ready` endpoint
  - [ ] Database connectivity check

## Phase 4: Testing ✓

### 4.1 Unit Tests
- [ ] `internal/client/eodhd_client_test.go`
  - [ ] Test successful API calls
  - [ ] Test retry logic
  - [ ] Test rate limiting
  - [ ] Test error handling
  - [ ] Mock HTTP responses

- [ ] `internal/repository/postgres_repo_test.go`
  - [ ] Test `GetAssetsRequiringPriceUpdate`
  - [ ] Test `GetMissingPriceDates`
  - [ ] Test batch insert performance

- [ ] `internal/worker/eodhd_price_sync_test.go`
  - [ ] Test sync algorithm
  - [ ] Test batch processing
  - [ ] Test error recovery
  - [ ] Mock repository and client

### 4.2 Integration Tests
- [ ] `tests/integration/price_sync_test.go`
  - [ ] End-to-end sync test
  - [ ] Test with real database
  - [ ] Test HTTP endpoints
  - [ ] Verify metrics

## Phase 5: Documentation ✓

### 5.1 Code Documentation
- [ ] Add godoc comments to all public functions
- [ ] Document configuration options
- [ ] Add usage examples

### 5.2 Operational Documentation
- [ ] Update `README.md`
- [ ] Create runbook for common issues
- [ ] Document monitoring setup
- [ ] Add troubleshooting guide

## Phase 6: Deployment ✓

### 6.1 Database
- [ ] Review existing indexes
- [ ] Add performance indexes if needed
- [ ] Test query performance

### 6.2 Container Configuration
- [ ] Update `Dockerfile` if needed
- [ ] Update `deployments/podman-compose.yml`
  - [ ] Add environment variables
  - [ ] Expose HTTP port (8080)
  - [ ] Add health check

### 6.3 Observability
- [ ] Update Prometheus scrape config
- [ ] Create Grafana dashboard
  - [ ] Price sync metrics
  - [ ] API call metrics
  - [ ] Error rates
  - [ ] Job duration

### 6.4 Alerts
- [ ] Configure Prometheus alerts
  - [ ] High error rate
  - [ ] Sync duration threshold
  - [ ] API rate limit warnings
  - [ ] Stale price count

## Phase 7: Production Readiness ✓

### 7.1 Security
- [ ] Validate API token is not logged
- [ ] Add authentication to HTTP endpoints
- [ ] Input validation and sanitization
- [ ] Rate limiting on HTTP endpoints

### 7.2 Performance
- [ ] Load test with 1000+ assets
- [ ] Optimize database queries
- [ ] Tune batch sizes
- [ ] Configure connection pooling

### 7.3 Reliability
- [ ] Test graceful shutdown
- [ ] Test recovery from failures
- [ ] Verify retry logic
- [ ] Test concurrent sync requests

## Quick Start Commands

### Development
```bash
# Set API token
export EODHD_API_TOKEN="your_token_here"

# Run tests
cd services/marketdata-service
go test ./...

# Run service locally
go run cmd/server/main.go
```

### Testing Sync Endpoint
```bash
# Trigger sync
curl -X POST http://localhost:8080/api/v1/sync-prices \
  -H "Content-Type: application/json" \
  -d '{
    "symbols": ["AAPL"],
    "exchange": "US",
    "from_date": "2024-11-01",
    "to_date": "2024-12-05"
  }'

# Check job status
curl http://localhost:8080/api/v1/sync-prices/{job_id}
```

### Monitoring
```bash
# Check metrics
curl http://localhost:9099/metrics | grep eodhd

# Check Prometheus targets
open http://localhost:9090/targets

# Open Grafana dashboard
open http://localhost:3001
```

## File Structure

```
services/marketdata-service/
├── cmd/
│   └── server/
│       └── main.go                    # Updated with HTTP server
├── internal/
│   ├── client/
│   │   ├── eodhd_client.go           # NEW: EODHD API client
│   │   └── eodhd_client_test.go      # NEW: Client tests
│   ├── domain/
│   │   ├── marketdata.go              # Existing
│   │   └── eodhd.go                   # NEW: EODHD models
│   ├── handler/
│   │   ├── grpc/
│   │   │   └── handler.go             # Existing
│   │   └── http/                      # NEW: HTTP handlers
│   │       ├── server.go
│   │       ├── sync_handler.go
│   │       └── health_handler.go
│   ├── metrics/
│   │   ├── metrics.go                 # Existing
│   │   └── eodhd_metrics.go          # NEW: EODHD metrics
│   ├── repository/
│   │   └── postgres_repo.go           # Updated with new methods
│   └── worker/
│       ├── ingestion.go               # Existing
│       ├── price_ingestion.go         # Existing
│       ├── currency_ingestion.go      # Existing
│       └── eodhd_price_sync.go       # NEW: EODHD sync worker
├── tests/
│   └── integration/
│       └── price_sync_test.go         # NEW: Integration tests
└── docs/
    ├── EODHD_PRICE_SYNC_STRATEGY.md  # This strategy doc
    └── EODHD_IMPLEMENTATION_CHECKLIST.md  # This checklist
```

## Dependencies to Add

```bash
cd services/marketdata-service

# Rate limiting
go get golang.org/x/time/rate

# HTTP router (if using chi)
go get github.com/go-chi/chi/v5

# HTTP middleware
go get github.com/go-chi/cors
```

## Environment Variables Template

Add to `.env` or deployment config:

```bash
# EODHD Configuration
EODHD_API_TOKEN=your_api_token_here
EODHD_API_BASE_URL=https://eodhd.com/api
EODHD_RATE_LIMIT=1.0

# Sync Worker Configuration
PRICE_SYNC_INTERVAL=1h
PRICE_STALE_DURATION=24h
PRICE_SYNC_BATCH_SIZE=100
PRICE_SYNC_MAX_CONCURRENCY=5
PRICE_SYNC_HISTORICAL_DAYS=30

# HTTP Server
HTTP_SERVER_PORT=8080
HTTP_READ_TIMEOUT=30s
HTTP_WRITE_TIMEOUT=30s
HTTP_IDLE_TIMEOUT=120s
```

## Success Criteria

- ✅ Worker successfully syncs prices every hour
- ✅ HTTP endpoint triggers on-demand sync
- ✅ Metrics are exposed and scraped by Prometheus
- ✅ No API rate limit violations
- ✅ Database performance is acceptable (< 100ms for batch insert)
- ✅ Error rate < 5%
- ✅ All tests passing
- ✅ Documentation complete

## Notes

- Start with a small set of assets for testing
- Monitor API usage to avoid rate limits
- Consider using EODHD's bulk endpoints for efficiency
- Implement circuit breaker if API becomes unreliable
- Use database transactions for data consistency
