# MarketData Service

The MarketData Service provides access to financial market data including asset information, price data, and currency exchange rates.

## Overview

This service manages:
- **Assets**: Financial instruments (stocks, cryptocurrencies, etc.)
- **Prices**: Historical and real-time price data for assets
- **Currency Rates**: Exchange rates between different currencies

## Database Schema

### Table: `marketdata.assets`

```sql
CREATE TABLE marketdata.assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    exchange VARCHAR(50),
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Table: `marketdata.asset_prices`

```sql
CREATE TABLE marketdata.asset_prices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id UUID NOT NULL REFERENCES marketdata.assets(id) ON DELETE CASCADE,
    price DECIMAL(20, 8) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Table: `marketdata.currency_rates`

```sql
CREATE TABLE marketdata.currency_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base_currency VARCHAR(3) NOT NULL,
    target_currency VARCHAR(3) NOT NULL,
    rate DECIMAL(20, 8) NOT NULL,
    rate_date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_currency_rate UNIQUE (base_currency, target_currency, rate_date)
);
```

## Configuration

The service uses environment variables for configuration:

### Required Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `DB_HOST` | PostgreSQL host | `localhost` | `postgres` |
| `DB_PORT` | PostgreSQL port | `5432` | `5432` |
| `DB_USER` | Database user | `garcios` | `garcios` |
| `DB_PASSWORD` | Database password | `Password123` | `Password123` |
| `DB_NAME` | Database name | `portfolio` | `portfolio` |
| `DB_SSLMODE` | SSL mode | `disable` | `disable` or `require` |
| `PORT` | gRPC server port | `50054` | `50054` |
| `METRICS_PORT` | Metrics server port | `9099` | `9099` |

### Optional Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `EODHD_API_TOKEN` | EODHD API token for data sync | - |
| `MINIO_ENDPOINT` | MinIO endpoint for data storage | `minio:9000` |
| `MINIO_ACCESS_KEY` | MinIO access key | `minioadmin` |
| `MINIO_SECRET_KEY` | MinIO secret key | - |

## gRPC Endpoints

### Asset Operations

#### GetAsset
Retrieves detailed information about a specific asset by its symbol.

```protobuf
rpc GetAsset(GetAssetRequest) returns (GetAssetResponse);
```

#### ListAssets
Returns a paginated list of available assets.

```protobuf
rpc ListAssets(ListAssetsRequest) returns (ListAssetsResponse);
```

### Price Operations

#### GetLatestPrice
Retrieves the latest price for a specific asset.

```protobuf
rpc GetLatestPrice(GetLatestPriceRequest) returns (GetLatestPriceResponse);
```

#### GetLatestPrices
Retrieves the latest prices for multiple assets in a single batch call.

```protobuf
rpc GetLatestPrices(GetLatestPricesRequest) returns (GetLatestPricesResponse);
```

#### GetHistoricalPrices
Retrieves historical price data for a specific asset within a time range.

```protobuf
rpc GetHistoricalPrices(GetHistoricalPricesRequest) returns (GetHistoricalPricesResponse);
```

### Currency Operations

#### GetLatestCurrencyRate
Retrieves the latest exchange rate between two currencies.

```protobuf
rpc GetLatestCurrencyRate(GetLatestCurrencyRateRequest) returns (GetLatestCurrencyRateResponse);
```

#### GetHistoricalCurrencyRates
Retrieves historical exchange rates between two currencies within a time range.

```protobuf
rpc GetHistoricalCurrencyRates(GetHistoricalCurrencyRatesRequest) returns (GetHistoricalCurrencyRatesResponse);
```

## Integration Testing

The service includes a comprehensive integration test suite that tests all gRPC endpoints and database operations.

### Prerequisites

- MarketData service running on `localhost:50054`
- PostgreSQL running on `localhost:5432`
- `grpcurl` installed (`brew install grpcurl`)
- `psql` installed
- `jq` installed (`brew install jq`)

### Running Integration Tests

```bash
cd services/marketdata-service
chmod +x test_integration.sh
./test_integration.sh
```

### What the Tests Cover

- **GetAsset RPC** - Retrieves asset details by symbol
- **ListAssets RPC** - Lists assets with pagination
- **GetLatestPrice RPC** - Gets latest price for an asset
- **GetLatestPrices RPC** - Batch retrieves latest prices
- **GetHistoricalPrices RPC** - Gets historical price data
- **GetLatestCurrencyRate RPC** - Gets latest exchange rate
- **GetHistoricalCurrencyRates RPC** - Gets historical exchange rates
- **Database Schema Validation** - Verifies table structure and indexes

The tests automatically clean up test data after execution.

## Running the Service

### Local Development

```bash
# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=garcios
export DB_PASSWORD=Password123
export DB_NAME=portfolio
export DB_SSLMODE=disable

# Run the service
go run cmd/server/main.go
```

### Docker/Podman

The service is configured in `docker-compose.yml`:

```yaml
marketdata-service:
  build:
    context: ../../
    dockerfile: services/marketdata-service/Dockerfile
  ports:
    - "50054:50054"
    - "9099:9099"
  environment:
    - DB_HOST=postgres
    - DB_USER=garcios
    - DB_PASSWORD=Password123
    - DB_NAME=portfolio
  depends_on:
    - postgres
    - minio
```

Start with:
```bash
make services-up
```

## Background Workers

The service includes several background workers:

### Ingestion Workers
- **Asset Ingestion Worker**: Processes CSV files with asset data
- **Price Ingestion Worker**: Processes CSV files with price data
- **Currency Ingestion Worker**: Processes CSV files with currency rates

### Sync Workers
- **EODHD Price Sync Worker**: Syncs price data from EODHD API
- **EODHD Currency Sync Worker**: Syncs currency rates from EODHD API

## Metrics

The service exposes Prometheus metrics on the metrics port (default: 9099):

- `marketdata_total_assets` - Total number of assets in the database
- `marketdata_total_prices` - Total number of price records
- gRPC request metrics (latency, count, errors)

Access metrics at: `http://localhost:9099/metrics`

## Admin Endpoints

### Manual Currency Sync

Trigger a manual currency rate sync:

```bash
curl -X POST http://localhost:9099/sync/currencies
```

## Dependencies

- `database/sql` - Standard library SQL interface
- `github.com/lib/pq` - PostgreSQL driver
- `google.golang.org/grpc` - gRPC framework
- `github.com/minio/minio-go` - MinIO client for object storage

---

**Last Updated**: 2025-12-11  
**Service Version**: 1.0.0
