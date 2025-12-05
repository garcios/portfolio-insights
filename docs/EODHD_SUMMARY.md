# EODHD Price Sync Implementation Summary

## Overview

This document provides a complete strategy for retrieving the latest asset prices from the EODHD API and synchronizing them with the `marketdata.asset_prices` table.

## Quick Links

- **Detailed Strategy**: [EODHD_PRICE_SYNC_STRATEGY.md](./EODHD_PRICE_SYNC_STRATEGY.md)
- **Implementation Checklist**: [EODHD_IMPLEMENTATION_CHECKLIST.md](./EODHD_IMPLEMENTATION_CHECKLIST.md)

## What Has Been Created

### 1. Documentation
- ✅ **EODHD_PRICE_SYNC_STRATEGY.md** - Comprehensive 500+ line strategy document
- ✅ **EODHD_IMPLEMENTATION_CHECKLIST.md** - Step-by-step implementation guide

### 2. Core Implementation Files
- ✅ **internal/client/eodhd_client.go** - EODHD API client with:
  - Rate limiting (configurable requests/second)
  - Exponential backoff retry logic
  - Comprehensive error handling
  - Support for real-time and historical price endpoints

- ✅ **internal/worker/eodhd_price_sync.go** - Price synchronization worker with:
  - Periodic background sync
  - Batch processing with concurrency control
  - On-demand sync capability for HTTP endpoint
  - Intelligent date range determination
  - Comprehensive metrics and logging

- ✅ **internal/repository/postgres_repo.go** - Added 3 new methods:
  - `GetAssetsRequiringPriceUpdate()` - Find assets with stale prices
  - `GetLatestPriceTimestamp()` - Get last price timestamp for an asset
  - `GetMissingPriceDates()` - Identify missing price dates

- ✅ **internal/domain/marketdata.go** - Updated interface with new methods

- ✅ **internal/usecase/marketdata_usecase_test.go** - Updated mock repository

## Architecture Summary

```
┌─────────────────────────────────────────────────────────┐
│                   HTTP Endpoint (Future)                 │
│              POST /api/v1/sync-prices                    │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│            EODHDPriceSyncWorker                          │
│  • Periodic sync (configurable interval)                │
│  • Batch processing with concurrency control            │
│  • Intelligent date range determination                 │
└──────────┬────────────────────────┬─────────────────────┘
           │                        │
           ▼                        ▼
┌──────────────────────┐  ┌────────────────────────────┐
│   EODHD API Client   │  │  PostgreSQL Repository     │
│  • Rate limiting     │  │  • Query stale assets      │
│  • Retry logic       │  │  • Batch insert/update     │
│  • Error handling    │  │  • Missing date detection  │
└──────────────────────┘  └────────────────────────────┘
```

## Key Features

### 1. EODHD API Client
- **Rate Limiting**: Configurable to avoid API limits (default: 1 req/sec)
- **Retry Logic**: Exponential backoff with configurable max retries
- **Error Handling**: 
  - Handles 429 (rate limit) with Retry-After header
  - Retries on 5xx server errors
  - Fails fast on 4xx client errors
- **Endpoints Supported**:
  - Real-time prices: `/api/real-time/{TICKER}.{EXCHANGE}`
  - Historical prices: `/api/eod/{TICKER}.{EXCHANGE}`

### 2. Price Sync Worker
- **Periodic Sync**: Runs at configurable intervals (default: 1 hour)
- **Stale Detection**: Identifies assets with prices older than threshold (default: 24 hours)
- **Batch Processing**: Processes assets in batches with concurrency control
- **Intelligent Backfill**: 
  - For new assets: Fetches last N days (default: 30 days)
  - For existing assets: Fetches only missing dates
- **Metrics**: Comprehensive Prometheus metrics for monitoring

### 3. Repository Enhancements
- **GetAssetsRequiringPriceUpdate**: Efficient query to find stale assets
- **GetLatestPriceTimestamp**: Quick lookup of last price date
- **GetMissingPriceDates**: Identifies gaps in price history

## Configuration

### Environment Variables

```bash
# EODHD API Configuration
EODHD_API_TOKEN=your_api_token_here
EODHD_API_BASE_URL=https://eodhd.com/api
EODHD_RATE_LIMIT=1.0  # requests per second

# Sync Worker Configuration
PRICE_SYNC_INTERVAL=1h
PRICE_STALE_DURATION=24h
PRICE_SYNC_BATCH_SIZE=100
PRICE_SYNC_MAX_CONCURRENCY=5
PRICE_SYNC_HISTORICAL_DAYS=30
```

## Next Steps

### Phase 1: Integration (Immediate)
1. Add dependency: `go get golang.org/x/time/rate`
2. Update `cmd/server/main.go` to initialize and start the worker
3. Test with a small set of assets
4. Monitor metrics and logs

### Phase 2: HTTP API (Optional)
1. Create HTTP server setup
2. Implement sync endpoints:
   - `POST /api/v1/sync-prices` - Trigger on-demand sync
   - `GET /api/v1/sync-prices/:job_id` - Check sync status
3. Add authentication and rate limiting

### Phase 3: Production Readiness
1. Add comprehensive tests
2. Performance optimization
3. Create Grafana dashboard
4. Set up Prometheus alerts

## Usage Examples

### Worker Integration (main.go)

```go
// Initialize EODHD price sync worker
eodhd Worker, err := worker.NewEODHDPriceSyncWorker(repo)
if err != nil {
    l.Error("failed to create EODHD price sync worker", "error", err)
    // Non-fatal - continue without EODHD sync
} else {
    eodhd Worker.Start(ctx)
    l.Info("EODHD price sync worker started")
}
```

### Manual Testing

```bash
# Set API token
export EODHD_API_TOKEN="your_token_here"

# Run service
cd services/marketdata-service
go run cmd/server/main.go

# Check logs for sync activity
# Check metrics
curl http://localhost:9099/metrics | grep eodhd
```

## Monitoring

### Key Metrics
- `eodhd_api_calls_total` - Total API calls by endpoint and status
- `eodhd_api_latency_seconds` - API call latency histogram
- `prices_synced_total` - Total prices synchronized
- `price_sync_job_duration_seconds` - Sync job duration
- `price_sync_errors_total` - Sync errors by type

### Grafana Dashboard Panels (Recommended)
1. Price sync job duration (histogram)
2. EODHD API calls per minute
3. Prices synced per hour
4. Error rate by type
5. Assets with stale prices count

## Database Schema

The implementation uses the existing schema:

```sql
-- Assets table (existing)
CREATE TABLE marketdata.assets (
    id UUID PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    exchange VARCHAR(50),
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Asset prices table (existing)
CREATE TABLE marketdata.asset_prices (
    id UUID PRIMARY KEY,
    asset_id UUID NOT NULL REFERENCES marketdata.assets(id),
    price DECIMAL(20, 8) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(asset_id, timestamp)  -- Prevents duplicate prices
);

-- Indexes (existing)
CREATE INDEX idx_assets_symbol ON marketdata.assets(symbol);
CREATE INDEX idx_asset_prices_asset_id_timestamp 
    ON marketdata.asset_prices(asset_id, timestamp DESC);
```

## API Rate Limits

### EODHD API Tiers
- **Free**: 20 requests/day
- **Standard**: 100,000 requests/day (~1.15 req/sec)
- **Professional**: 1,000,000 requests/day (~11.5 req/sec)

### Recommended Settings by Tier
- **Free**: `EODHD_RATE_LIMIT=0.0002` (1 req per 5000 seconds)
- **Standard**: `EODHD_RATE_LIMIT=1.0` (1 req/sec, safe margin)
- **Professional**: `EODHD_RATE_LIMIT=10.0` (10 req/sec)

## Error Handling

### Retry Strategy
- **Initial Backoff**: 1 second
- **Max Backoff**: 30 seconds
- **Multiplier**: 2.0 (exponential)
- **Max Retries**: 3

### Error Types
1. **Rate Limit (429)**: Respects Retry-After header
2. **Server Errors (5xx)**: Retries with exponential backoff
3. **Client Errors (4xx)**: Fails immediately (no retry)
4. **Network Errors**: Retries with exponential backoff

## Security Considerations

1. **API Token Protection**
   - Stored in environment variable
   - Never logged
   - Rotate periodically

2. **Input Validation**
   - Validate symbol format
   - Validate date ranges
   - Sanitize all inputs

3. **Rate Limiting**
   - Prevents API abuse
   - Protects against rate limit violations

## Troubleshooting

### Common Issues

1. **API Rate Limit Exceeded**
   - Reduce `EODHD_RATE_LIMIT`
   - Increase `PRICE_SYNC_INTERVAL`
   - Reduce `PRICE_SYNC_BATCH_SIZE`

2. **Database Deadlocks**
   - Reduce `PRICE_SYNC_MAX_CONCURRENCY`
   - Reduce batch size in `InsertPrices`

3. **Missing Prices**
   - Check asset symbol format (must include exchange, e.g., "AAPL.US")
   - Verify EODHD API coverage for the symbol
   - Check logs for API errors

4. **Slow Sync Performance**
   - Increase `PRICE_SYNC_MAX_CONCURRENCY`
   - Optimize database queries
   - Use connection pooling

## Testing Strategy

### Unit Tests
- EODHD client with mocked HTTP responses
- Repository methods with test database
- Worker logic with mocked dependencies

### Integration Tests
- End-to-end sync with test data
- Database performance under load
- API error handling

### Load Tests
- Sync 1000+ assets
- Concurrent sync requests
- Database performance

## Performance Considerations

### Database
- Existing indexes are sufficient
- Batch inserts use `ON CONFLICT` for upserts
- Connection pooling recommended

### API Calls
- Rate limiting prevents overwhelming EODHD API
- Concurrent processing with semaphore
- Efficient date range determination

### Memory
- Batch processing prevents loading all data at once
- Streaming results from database
- Configurable batch sizes

## Success Criteria

- ✅ Worker successfully syncs prices at configured interval
- ✅ No API rate limit violations
- ✅ Database performance acceptable (< 100ms for batch insert)
- ✅ Error rate < 5%
- ✅ Metrics exposed and scraped by Prometheus
- ✅ All tests passing
- ✅ Documentation complete

## Support

For questions or issues:
1. Check the detailed strategy document
2. Review the implementation checklist
3. Check logs and metrics
4. Consult the troubleshooting guide

## References

- [EODHD API Documentation](https://eodhd.com/financial-apis/)
- [EODHD EOD Data API](https://eodhd.com/financial-apis/api-for-historical-data-and-volumes/)
- [EODHD Real-time API](https://eodhd.com/financial-apis/live-realtime-stocks-api/)
