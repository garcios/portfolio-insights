# EODHD Price Sync - Data Flow Diagram

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Portfolio Insights System                        │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│                      MarketData Service (Port 50054)                     │
│                                                                           │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    HTTP Server (Port 8080)                       │   │
│  │  ┌──────────────────────────────────────────────────────────┐   │   │
│  │  │  POST /api/v1/sync-prices                                │   │   │
│  │  │  GET  /api/v1/sync-prices/:job_id                        │   │   │
│  │  └──────────────────────────────────────────────────────────┘   │   │
│  └─────────────────────────┬───────────────────────────────────────┘   │
│                             │                                            │
│                             ▼                                            │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │              EODHDPriceSyncWorker                                │   │
│  │  ┌───────────────────────────────────────────────────────────┐  │   │
│  │  │  Periodic Sync (every 1 hour)                             │  │   │
│  │  │  1. Query stale assets                                    │  │   │
│  │  │  2. Determine missing dates                               │  │   │
│  │  │  3. Fetch from EODHD API                                  │  │   │
│  │  │  4. Batch insert to database                              │  │   │
│  │  │  5. Record metrics                                        │  │   │
│  │  └───────────────────────────────────────────────────────────┘  │   │
│  └────────┬──────────────────────────────────┬─────────────────────┘   │
│           │                                   │                          │
│           ▼                                   ▼                          │
│  ┌────────────────────┐            ┌──────────────────────────┐        │
│  │  EODHD API Client  │            │  PostgreSQL Repository   │        │
│  │  ┌──────────────┐  │            │  ┌────────────────────┐  │        │
│  │  │ Rate Limiter │  │            │  │  GetAssetsRequiring│  │        │
│  │  │ (1 req/sec)  │  │            │  │  PriceUpdate()     │  │        │
│  │  └──────────────┘  │            │  └────────────────────┘  │        │
│  │  ┌──────────────┐  │            │  ┌────────────────────┐  │        │
│  │  │ Retry Logic  │  │            │  │  GetLatestPrice    │  │        │
│  │  │ (Exp backoff)│  │            │  │  Timestamp()       │  │        │
│  │  └──────────────┘  │            │  └────────────────────┘  │        │
│  │  ┌──────────────┐  │            │  ┌────────────────────┐  │        │
│  │  │ HTTP Client  │  │            │  │  InsertPrices()    │  │        │
│  │  └──────────────┘  │            │  │  (Batch upsert)    │  │        │
│  └────────┬───────────┘            │  └────────────────────┘  │        │
│           │                         └──────────┬───────────────┘        │
│           │                                    │                         │
└───────────┼────────────────────────────────────┼─────────────────────────┘
            │                                    │
            ▼                                    ▼
┌───────────────────────┐          ┌──────────────────────────────┐
│   EODHD API           │          │   PostgreSQL Database        │
│   https://eodhd.com   │          │   ┌────────────────────────┐ │
│                       │          │   │ marketdata.assets      │ │
│  /api/real-time/      │          │   │  - id                  │ │
│  /api/eod/            │          │   │  - symbol              │ │
│                       │          │   │  - exchange            │ │
│  Response:            │          │   └────────────────────────┘ │
│  {                    │          │   ┌────────────────────────┐ │
│    "date": "...",     │          │   │ marketdata.asset_prices│ │
│    "open": 177.83,    │          │   │  - id                  │ │
│    "high": 182.88,    │          │   │  - asset_id (FK)       │ │
│    "low": 177.71,     │          │   │  - price               │ │
│    "close": 182.01,   │          │   │  - timestamp           │ │
│    "adjusted_close":  │          │   │  UNIQUE(asset_id,      │ │
│      178.2703,        │          │   │         timestamp)     │ │
│    "volume": 104487900│          │   └────────────────────────┘ │
│  }                    │          └──────────────────────────────┘
└───────────────────────┘
```

## Sync Flow Sequence

```
┌────────┐         ┌──────────┐         ┌─────────┐         ┌──────────┐
│ Timer  │         │  Worker  │         │  EODHD  │         │ Database │
└───┬────┘         └────┬─────┘         └────┬────┘         └────┬─────┘
    │                   │                     │                   │
    │ Tick (1 hour)     │                     │                   │
    ├──────────────────>│                     │                   │
    │                   │                     │                   │
    │                   │ Query stale assets  │                   │
    │                   ├─────────────────────────────────────────>│
    │                   │                     │                   │
    │                   │<─────────────────────────────────────────┤
    │                   │ [AAPL, GOOGL, MSFT]│                   │
    │                   │                     │                   │
    │                   │ For each asset:     │                   │
    │                   │                     │                   │
    │                   │ Get latest timestamp│                   │
    │                   ├─────────────────────────────────────────>│
    │                   │<─────────────────────────────────────────┤
    │                   │ 2024-11-01          │                   │
    │                   │                     │                   │
    │                   │ GET /api/eod/AAPL.US?from=2024-11-02&to=2024-12-05
    │                   ├────────────────────>│                   │
    │                   │                     │                   │
    │                   │<────────────────────┤                   │
    │                   │ [prices array]      │                   │
    │                   │                     │                   │
    │                   │ Batch insert prices │                   │
    │                   ├─────────────────────────────────────────>│
    │                   │                     │                   │
    │                   │<─────────────────────────────────────────┤
    │                   │ OK (34 rows)        │                   │
    │                   │                     │                   │
    │                   │ Record metrics      │                   │
    │                   │ (prices_synced: 34) │                   │
    │                   │                     │                   │
```

## Error Handling Flow

```
┌──────────┐         ┌─────────┐         ┌──────────┐
│  Worker  │         │  EODHD  │         │  Retry   │
└────┬─────┘         └────┬────┘         └────┬─────┘
     │                     │                   │
     │ GET /api/eod/...    │                   │
     ├────────────────────>│                   │
     │                     │                   │
     │<────────────────────┤                   │
     │ 429 Too Many Requests                  │
     │ Retry-After: 60     │                   │
     │                     │                   │
     │ Wait 60 seconds     │                   │
     ├─────────────────────────────────────────>│
     │                     │                   │
     │<─────────────────────────────────────────┤
     │                     │                   │
     │ GET /api/eod/... (retry)               │
     ├────────────────────>│                   │
     │                     │                   │
     │<────────────────────┤                   │
     │ 200 OK              │                   │
     │ [prices]            │                   │
```

## Concurrency Model

```
┌─────────────────────────────────────────────────────────┐
│                    Sync Worker                          │
│                                                          │
│  Assets: [AAPL, GOOGL, MSFT, TSLA, AMZN, ...]          │
│                                                          │
│  Batch 1: [AAPL, GOOGL, MSFT, TSLA, AMZN]              │
│                                                          │
│  ┌─────────────────────────────────────────────────┐   │
│  │  Semaphore (Max Concurrency: 5)                 │   │
│  │  ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐       │   │
│  │  │Slot1│ │Slot2│ │Slot3│ │Slot4│ │Slot5│       │   │
│  │  └──┬──┘ └──┬──┘ └──┬──┘ └──┬──┘ └──┬──┘       │   │
│  └─────┼───────┼───────┼───────┼───────┼───────────┘   │
│        │       │       │       │       │                │
│        ▼       ▼       ▼       ▼       ▼                │
│     ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐               │
│     │AAPL│ │GOOGL│MSFT│ │TSLA│ │AMZN│               │
│     └─┬──┘ └─┬──┘ └─┬──┘ └─┬──┘ └─┬──┘               │
│       │      │      │      │      │                    │
│       ▼      ▼      ▼      ▼      ▼                    │
│    ┌────────────────────────────────────┐              │
│    │      Rate Limiter (1 req/sec)      │              │
│    └────────────────────────────────────┘              │
│                      │                                  │
│                      ▼                                  │
│              ┌──────────────┐                          │
│              │  EODHD API   │                          │
│              └──────────────┘                          │
└─────────────────────────────────────────────────────────┘
```

## Metrics Flow

```
┌──────────────────────────────────────────────────────────┐
│                    Sync Worker                           │
│                                                           │
│  On API Call:                                            │
│  eodhd_api_calls_total{endpoint="eod",status="success"}++│
│  eodhd_api_latency_seconds.Observe(duration)            │
│                                                           │
│  On Price Insert:                                        │
│  prices_synced_total{source="eodhd"}+=34                │
│                                                           │
│  On Job Complete:                                        │
│  price_sync_job_duration_seconds.Observe(120.5)         │
│                                                           │
│  On Error:                                               │
│  price_sync_errors_total{error_type="api_error"}++      │
└────────────────────────┬─────────────────────────────────┘
                         │
                         ▼
              ┌──────────────────────┐
              │  Prometheus          │
              │  (Port 9090)         │
              │  Scrapes :9099/metrics│
              └──────────┬───────────┘
                         │
                         ▼
              ┌──────────────────────┐
              │  Grafana             │
              │  (Port 3001)         │
              │  Dashboards & Alerts │
              └──────────────────────┘
```

## Database Upsert Logic

```
┌──────────────────────────────────────────────────────────┐
│                    InsertPrices()                         │
│                                                           │
│  Batch: [                                                │
│    {asset_id: "uuid1", price: 150.0, timestamp: "2024-12-01"},
│    {asset_id: "uuid1", price: 151.0, timestamp: "2024-12-02"},
│    {asset_id: "uuid1", price: 152.0, timestamp: "2024-12-03"}
│  ]                                                        │
│                                                           │
│  SQL:                                                     │
│  INSERT INTO marketdata.asset_prices                     │
│    (asset_id, price, timestamp)                          │
│  VALUES                                                   │
│    ($1, $2, $3),                                         │
│    ($4, $5, $6),                                         │
│    ($7, $8, $9)                                          │
│  ON CONFLICT (asset_id, timestamp)                       │
│  DO UPDATE SET                                           │
│    price = EXCLUDED.price                                │
│                                                           │
│  Result:                                                  │
│  - New records inserted                                  │
│  - Existing records updated (if price changed)           │
│  - No duplicates                                         │
└──────────────────────────────────────────────────────────┘
```

## Configuration Hierarchy

```
┌─────────────────────────────────────────────────────────┐
│                  Environment Variables                   │
│                                                          │
│  EODHD_API_TOKEN          (required)                    │
│  EODHD_API_BASE_URL       (default: https://eodhd.com/api)
│  EODHD_RATE_LIMIT         (default: 1.0 req/sec)        │
│  PRICE_SYNC_INTERVAL      (default: 1h)                 │
│  PRICE_STALE_DURATION     (default: 24h)                │
│  PRICE_SYNC_BATCH_SIZE    (default: 100)                │
│  PRICE_SYNC_MAX_CONCURRENCY (default: 5)                │
│  PRICE_SYNC_HISTORICAL_DAYS (default: 30)               │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
          ┌──────────────────────┐
          │  LoadPriceSyncConfig │
          │  FromEnv()           │
          └──────────┬───────────┘
                     │
                     ▼
          ┌──────────────────────┐
          │  PriceSyncConfig     │
          │  struct              │
          └──────────┬───────────┘
                     │
                     ▼
          ┌──────────────────────┐
          │  EODHDPriceSyncWorker│
          └──────────────────────┘
```
