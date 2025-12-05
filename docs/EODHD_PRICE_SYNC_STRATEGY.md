# EODHD Price Synchronization Strategy

## Overview

This document outlines a comprehensive strategy for retrieving the latest asset prices from the EODHD API and synchronizing them with the `marketdata.asset_prices` table in the portfolio-insights system.

## Architecture

### Components

1. **EODHD Client** - HTTP client for interacting with EODHD API
2. **Price Sync Worker** - Background worker for scheduled price updates
3. **HTTP Handler** - On-demand endpoint to trigger price synchronization
4. **Repository Layer** - Database operations for price management
5. **Metrics** - Prometheus metrics for monitoring

### Data Flow

```
┌─────────────────┐
│  HTTP Endpoint  │ (On-demand trigger)
│  /sync-prices   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Price Sync     │
│  Orchestrator   │
└────────┬────────┘
         │
         ├──────────────────┐
         ▼                  ▼
┌─────────────────┐  ┌──────────────────┐
│  EODHD Client   │  │  Repository      │
│  (API Calls)    │  │  (Read existing) │
└────────┬────────┘  └────────┬─────────┘
         │                    │
         │                    │
         ▼                    ▼
┌─────────────────────────────────┐
│  Determine Missing/Stale Data   │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────┐
│  Batch Insert/  │
│  Update Prices  │
└─────────────────┘
```

## Implementation Plan

### 1. EODHD API Client

**File**: `services/marketdata-service/internal/client/eodhd_client.go`

#### Features:
- HTTP client with retry logic and rate limiting
- Support for both real-time and historical price endpoints
- Configurable timeout and connection pooling
- Error handling for API rate limits (429) and server errors

#### API Endpoints:
```
Real-time: https://eodhd.com/api/real-time/{TICKER}.{EXCHANGE}?api_token={TOKEN}&fmt=json
Historical: https://eodhd.com/api/eod/{TICKER}.{EXCHANGE}?from={DATE}&to={DATE}&api_token={TOKEN}&fmt=json
```

#### Response Format:
```json
{
  "code": "AAPL.US",
  "timestamp": 1609459200,
  "gmtoffset": 0,
  "open": 133.52,
  "high": 133.61,
  "low": 126.76,
  "close": 129.41,
  "volume": 143301900,
  "previousClose": 132.69,
  "change": -3.28,
  "change_p": -2.4722
}
```

For historical data (EOD endpoint):
```json
[
  {
    "date": "2022-01-03",
    "open": 177.83,
    "high": 182.88,
    "low": 177.71,
    "close": 182.01,
    "adjusted_close": 178.2703,
    "volume": 104487900
  }
]
```

#### Client Interface:
```go
type EODHDClient interface {
    GetRealTimePrice(ctx context.Context, ticker, exchange string) (*RealTimePrice, error)
    GetHistoricalPrices(ctx context.Context, ticker, exchange string, from, to time.Time) ([]*HistoricalPrice, error)
    GetBulkPrices(ctx context.Context, exchange string) (map[string]*RealTimePrice, error)
}
```

### 2. Repository Enhancements

**File**: `services/marketdata-service/internal/repository/postgres_repo.go`

#### New Methods:

```go
// GetAssetsRequiringPriceUpdate returns assets that need price updates
// based on the last price timestamp
func (r *postgresMarketDataRepo) GetAssetsRequiringPriceUpdate(staleDuration time.Duration) ([]*domain.Asset, error)

// GetLatestPriceTimestamp returns the most recent price timestamp for an asset
func (r *postgresMarketDataRepo) GetLatestPriceTimestamp(assetID string) (*time.Time, error)

// GetMissingPriceDates returns dates where prices are missing for an asset
// within a given date range
func (r *postgresMarketDataRepo) GetMissingPriceDates(assetID string, start, end time.Time) ([]time.Time, error)
```

#### Query Strategy:

**Identify Stale Prices:**
```sql
SELECT DISTINCT a.id, a.symbol, a.exchange, MAX(p.timestamp) as last_price_time
FROM marketdata.assets a
LEFT JOIN marketdata.asset_prices p ON a.id = p.asset_id
GROUP BY a.id, a.symbol, a.exchange
HAVING MAX(p.timestamp) IS NULL 
   OR MAX(p.timestamp) < NOW() - INTERVAL '1 day'
ORDER BY last_price_time ASC NULLS FIRST;
```

**Find Missing Dates:**
```sql
WITH date_series AS (
    SELECT generate_series(
        $1::date,
        $2::date,
        '1 day'::interval
    )::date AS date
),
existing_prices AS (
    SELECT DATE(timestamp) as price_date
    FROM marketdata.asset_prices
    WHERE asset_id = $3
    AND timestamp >= $1
    AND timestamp <= $2
)
SELECT ds.date
FROM date_series ds
LEFT JOIN existing_prices ep ON ds.date = ep.price_date
WHERE ep.price_date IS NULL
ORDER BY ds.date;
```

### 3. Price Sync Worker

**File**: `services/marketdata-service/internal/worker/eodhd_price_sync.go`

#### Responsibilities:
1. Periodically check for assets requiring price updates
2. Fetch missing/stale prices from EODHD API
3. Batch insert/update prices in the database
4. Handle errors and retry logic
5. Emit metrics for monitoring

#### Configuration:
```go
type PriceSyncConfig struct {
    SyncInterval     time.Duration // How often to run sync (e.g., 1 hour)
    StaleDuration    time.Duration // How old before price is considered stale (e.g., 24 hours)
    BatchSize        int           // Number of assets to process per batch
    MaxConcurrency   int           // Max concurrent API calls
    HistoricalDays   int           // How many days of historical data to backfill
}
```

#### Worker Structure:
```go
type EODHDPriceSyncWorker struct {
    repo         domain.MarketDataRepository
    eodhd        EODHDClient
    config       PriceSyncConfig
    apiToken     string
}

func (w *EODHDPriceSyncWorker) Start(ctx context.Context) {
    ticker := time.NewTicker(w.config.SyncInterval)
    defer ticker.Stop()
    
    // Run immediately on startup
    w.syncPrices(ctx)
    
    for {
        select {
        case <-ticker.C:
            w.syncPrices(ctx)
        case <-ctx.Done():
            return
        }
    }
}

func (w *EODHDPriceSyncWorker) syncPrices(ctx context.Context) error {
    // 1. Get assets requiring updates
    // 2. Process in batches with concurrency control
    // 3. For each asset, determine missing dates
    // 4. Fetch from EODHD API
    // 5. Batch insert to database
    // 6. Record metrics
}
```

#### Sync Algorithm:

```
1. Query assets requiring updates (stale or missing prices)
2. For each asset:
   a. Get latest price timestamp from DB
   b. If no prices exist:
      - Fetch historical prices for last N days
   c. If prices exist but stale:
      - Fetch prices from last_timestamp to now
   d. Rate limit API calls (e.g., 1 request/second)
3. Transform EODHD response to domain.AssetPrice
4. Batch insert/upsert prices (1000 per batch)
5. Update metrics (prices_synced, api_calls, errors)
```

### 4. HTTP Endpoint

**File**: `services/marketdata-service/internal/handler/http/sync_handler.go`

#### Endpoint Design:

**POST /api/v1/sync-prices**

Request Body:
```json
{
  "symbols": ["AAPL", "GOOGL", "MSFT"],  // Optional: specific symbols
  "exchange": "US",                       // Optional: default "US"
  "from_date": "2024-01-01",             // Optional: start date
  "to_date": "2024-12-05",               // Optional: end date
  "force": false                          // Optional: force refresh even if not stale
}
```

Response:
```json
{
  "job_id": "sync-20241205-171300",
  "status": "started",
  "assets_queued": 3,
  "message": "Price synchronization job started"
}
```

**GET /api/v1/sync-prices/{job_id}**

Response:
```json
{
  "job_id": "sync-20241205-171300",
  "status": "completed",
  "started_at": "2024-12-05T17:13:00Z",
  "completed_at": "2024-12-05T17:15:30Z",
  "assets_processed": 3,
  "prices_synced": 450,
  "errors": 0,
  "details": [
    {
      "symbol": "AAPL",
      "prices_added": 150,
      "status": "success"
    }
  ]
}
```

#### Handler Implementation:
```go
type SyncHandler struct {
    worker *EODHDPriceSyncWorker
    jobs   map[string]*SyncJob // In-memory job tracking (consider Redis for production)
    mu     sync.RWMutex
}

func (h *SyncHandler) HandleSyncRequest(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request
    // 2. Validate symbols/dates
    // 3. Create job ID
    // 4. Start async sync operation
    // 5. Return job ID immediately
}

func (h *SyncHandler) HandleJobStatus(w http.ResponseWriter, r *http.Request) {
    // 1. Extract job ID
    // 2. Look up job status
    // 3. Return current status and progress
}
```

### 5. Domain Models

**File**: `services/marketdata-service/internal/domain/eodhd.go`

```go
type RealTimePrice struct {
    Code          string    `json:"code"`
    Timestamp     int64     `json:"timestamp"`
    Open          float64   `json:"open"`
    High          float64   `json:"high"`
    Low           float64   `json:"low"`
    Close         float64   `json:"close"`
    Volume        int64     `json:"volume"`
    PreviousClose float64   `json:"previousClose"`
}

type HistoricalPrice struct {
    Date          string  `json:"date"`
    Open          float64 `json:"open"`
    High          float64 `json:"high"`
    Low           float64 `json:"low"`
    Close         float64 `json:"close"`
    AdjustedClose float64 `json:"adjusted_close"`
    Volume        int64   `json:"volume"`
}

type SyncJob struct {
    ID              string
    Status          string // "pending", "running", "completed", "failed"
    StartedAt       time.Time
    CompletedAt     *time.Time
    AssetsProcessed int
    PricesSynced    int
    Errors          int
    Details         []AssetSyncDetail
}

type AssetSyncDetail struct {
    Symbol      string
    PricesAdded int
    Status      string
    Error       string
}
```

### 6. Error Handling & Retry Logic

#### API Error Handling:

```go
type RetryConfig struct {
    MaxRetries     int
    InitialBackoff time.Duration
    MaxBackoff     time.Duration
    Multiplier     float64
}

func (c *EODHDClient) callWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
    backoff := c.retryConfig.InitialBackoff
    
    for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
        resp, err := c.httpClient.Do(req)
        
        if err == nil && resp.StatusCode < 500 {
            return resp, nil
        }
        
        // Handle rate limiting
        if resp != nil && resp.StatusCode == 429 {
            retryAfter := parseRetryAfter(resp.Header)
            time.Sleep(retryAfter)
            continue
        }
        
        // Exponential backoff
        if attempt < c.retryConfig.MaxRetries {
            time.Sleep(backoff)
            backoff = time.Duration(float64(backoff) * c.retryConfig.Multiplier)
            if backoff > c.retryConfig.MaxBackoff {
                backoff = c.retryConfig.MaxBackoff
            }
        }
    }
    
    return nil, fmt.Errorf("max retries exceeded")
}
```

#### Database Error Handling:

- Use transactions for batch inserts
- Implement deadlock retry logic
- Log failed batches for manual review
- Continue processing on partial failures

### 7. Rate Limiting

#### EODHD API Limits:
- Free tier: 20 requests/day
- Standard: 100,000 requests/day
- Professional: 1,000,000 requests/day

#### Implementation:

```go
type RateLimiter struct {
    limiter *rate.Limiter
}

func NewRateLimiter(requestsPerSecond float64) *RateLimiter {
    return &RateLimiter{
        limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), 1),
    }
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
    return rl.limiter.Wait(ctx)
}
```

Usage:
```go
// For standard tier: ~1 request/second to stay under daily limit
rateLimiter := NewRateLimiter(1.0)

for _, asset := range assets {
    if err := rateLimiter.Wait(ctx); err != nil {
        return err
    }
    prices, err := client.GetHistoricalPrices(ctx, asset.Symbol, asset.Exchange, from, to)
    // ...
}
```

### 8. Metrics & Monitoring

**File**: `services/marketdata-service/internal/metrics/eodhd_metrics.go`

```go
var (
    EODHDAPICallsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "eodhd_api_calls_total",
            Help: "Total number of EODHD API calls",
        },
        []string{"endpoint", "status"},
    )
    
    EODHDAPILatency = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "eodhd_api_latency_seconds",
            Help:    "EODHD API call latency",
            Buckets: prometheus.DefBuckets,
        },
        []string{"endpoint"},
    )
    
    PricesSyncedTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "prices_synced_total",
            Help: "Total number of prices synchronized",
        },
        []string{"source"},
    )
    
    SyncJobDuration = promauto.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "price_sync_job_duration_seconds",
            Help:    "Duration of price sync jobs",
            Buckets: []float64{1, 5, 10, 30, 60, 120, 300, 600},
        },
    )
    
    SyncJobErrors = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "price_sync_errors_total",
            Help: "Total number of price sync errors",
        },
        []string{"error_type"},
    )
)
```

### 9. Configuration

**Environment Variables:**

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

# HTTP Server Configuration
HTTP_SERVER_PORT=8080
HTTP_READ_TIMEOUT=30s
HTTP_WRITE_TIMEOUT=30s
```

### 10. Testing Strategy

#### Unit Tests:

1. **EODHD Client Tests**
   - Mock HTTP responses
   - Test error handling
   - Test retry logic
   - Test rate limiting

2. **Repository Tests**
   - Test query for stale assets
   - Test missing date detection
   - Test batch insert/update

3. **Worker Tests**
   - Test sync algorithm
   - Test batch processing
   - Test error recovery

#### Integration Tests:

1. **End-to-End Sync Test**
   - Insert test assets
   - Trigger sync (with mock EODHD API)
   - Verify prices inserted
   - Check metrics

2. **HTTP Endpoint Tests**
   - Test request validation
   - Test job creation
   - Test status retrieval

#### Load Tests:

- Test syncing 1000+ assets
- Test concurrent sync requests
- Test database performance under load

### 11. Deployment Considerations

#### Database Indexes:

Already exists:
```sql
CREATE INDEX idx_asset_prices_asset_id_timestamp 
ON marketdata.asset_prices(asset_id, timestamp DESC);
```

Additional recommended:
```sql
-- For finding stale prices
CREATE INDEX idx_asset_prices_timestamp 
ON marketdata.asset_prices(timestamp DESC);

-- For symbol lookups
CREATE INDEX idx_assets_symbol_exchange 
ON marketdata.assets(symbol, exchange);
```

#### Podman Configuration:

Update `deployments/podman-compose.yml`:

```yaml
marketdata-service:
  environment:
    - EODHD_API_TOKEN=${EODHD_API_TOKEN}
    - PRICE_SYNC_INTERVAL=1h
    - PRICE_STALE_DURATION=24h
  ports:
    - "50054:50054"  # gRPC
    - "9099:9099"    # Metrics
    - "8080:8080"    # HTTP API (new)
```

#### Prometheus Scrape Config:

```yaml
scrape_configs:
  - job_name: 'marketdata-service'
    static_configs:
      - targets: ['marketdata-service:9099']
```

### 12. Logging Strategy

Use structured logging with context:

```go
logger.Info("Starting price sync",
    "job_id", jobID,
    "assets_count", len(assets),
    "from_date", fromDate,
    "to_date", toDate,
)

logger.Error("Failed to fetch prices",
    "symbol", symbol,
    "exchange", exchange,
    "error", err,
    "attempt", attempt,
)
```

### 13. Migration Path

#### Phase 1: Core Implementation
1. Implement EODHD client
2. Add repository methods
3. Create sync worker
4. Add metrics

#### Phase 2: HTTP API
1. Create HTTP handlers
2. Add job tracking
3. Implement status endpoint

#### Phase 3: Production Readiness
1. Add comprehensive tests
2. Performance optimization
3. Documentation
4. Monitoring dashboards

#### Phase 4: Enhancements
1. Support for multiple exchanges
2. Real-time price streaming
3. Price validation/anomaly detection
4. Historical data backfill tool

## Usage Examples

### Manual Sync via HTTP:

```bash
# Sync specific symbols
curl -X POST http://localhost:8080/api/v1/sync-prices \
  -H "Content-Type: application/json" \
  -d '{
    "symbols": ["AAPL", "GOOGL"],
    "exchange": "US",
    "from_date": "2024-11-01",
    "to_date": "2024-12-05"
  }'

# Response
{
  "job_id": "sync-20241205-171300",
  "status": "started",
  "assets_queued": 2
}

# Check status
curl http://localhost:8080/api/v1/sync-prices/sync-20241205-171300
```

### Programmatic Usage:

```go
// In another service
client := marketdata.NewClient("localhost:50054")

// Trigger sync via gRPC (if we add this to proto)
resp, err := client.SyncPrices(ctx, &pb.SyncPricesRequest{
    Symbols: []string{"AAPL", "GOOGL"},
    Exchange: "US",
})
```

## Security Considerations

1. **API Token Protection**
   - Store in environment variables
   - Never log API token
   - Rotate periodically

2. **Rate Limiting**
   - Implement per-client rate limiting on HTTP endpoint
   - Prevent abuse of sync endpoint

3. **Input Validation**
   - Validate symbol format
   - Validate date ranges
   - Sanitize inputs

4. **Authentication**
   - Add authentication to HTTP endpoints
   - Use API keys or JWT tokens

## Monitoring & Alerts

### Key Metrics to Monitor:

1. **Sync Success Rate**
   - Alert if < 95% success rate

2. **API Error Rate**
   - Alert on 429 (rate limit) errors
   - Alert on 5xx errors

3. **Sync Duration**
   - Alert if sync takes > 10 minutes

4. **Price Staleness**
   - Alert if > 10% of assets have stale prices

5. **Database Performance**
   - Monitor insert latency
   - Monitor query performance

### Grafana Dashboard Panels:

1. Price sync job duration (histogram)
2. EODHD API calls per minute
3. Prices synced per hour
4. Error rate by type
5. Assets with stale prices count

## Troubleshooting Guide

### Common Issues:

1. **API Rate Limit Exceeded**
   - Reduce sync frequency
   - Increase rate limiter delay
   - Upgrade EODHD plan

2. **Database Deadlocks**
   - Reduce batch size
   - Add retry logic
   - Check for long-running transactions

3. **Missing Prices**
   - Check asset symbol format
   - Verify exchange suffix
   - Check EODHD API coverage

4. **Slow Sync Performance**
   - Increase concurrency
   - Optimize database queries
   - Use connection pooling

## Future Enhancements

1. **Intelligent Scheduling**
   - Sync during market hours only
   - Different frequencies for different asset types

2. **Price Validation**
   - Detect anomalies (e.g., price spikes)
   - Cross-reference with multiple sources

3. **Caching Layer**
   - Redis cache for frequently accessed prices
   - Reduce database load

4. **Webhook Support**
   - Notify other services when prices update
   - Event-driven architecture

5. **Multi-Source Support**
   - Fallback to alternative data providers
   - Aggregate prices from multiple sources

## Conclusion

This strategy provides a robust, scalable solution for synchronizing asset prices from the EODHD API. The implementation follows clean architecture principles, includes comprehensive error handling, and provides observability through metrics and logging.

The phased approach allows for incremental development and testing, ensuring production readiness before full deployment.
