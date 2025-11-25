# Currency Rates Ingestion Worker

## Overview
The Currency Ingestion Worker is a background service that automatically ingests currency exchange rate data from CSV files stored in MinIO into the `marketdata.currency_rates` table.

## Features
- **Automatic Ingestion**: Runs daily at midnight to process new currency rate data
- **CSV Format Support**: Reads currency rates from standardized CSV files
- **Batch Processing**: Efficiently inserts rates in batches of 1000 records
- **Duplicate Prevention**: Uses unique constraint to prevent duplicate entries for the same currency pair and date
- **Error Handling**: Validates currency codes, rates, and dates with detailed logging
- **MinIO Integration**: Fetches CSV files from MinIO object storage

## CSV File Format

The worker expects a CSV file named `currency_rates.csv` in the MinIO bucket with the following format:

```csv
base_currency,target_currency,rate,rate_date
USD,EUR,0.92,2025-11-25
USD,GBP,0.79,2025-11-25
USD,JPY,149.50,2025-11-25
```

### Column Specifications:
1. **base_currency** (VARCHAR(3)): The base currency code (e.g., 'USD', 'EUR')
2. **target_currency** (VARCHAR(3)): The target currency code (e.g., 'GBP', 'JPY')
3. **rate** (DECIMAL): The exchange rate from base to target currency
4. **rate_date** (DATE): The date the rate is valid for (format: YYYY-MM-DD)

## Database Schema

The worker inserts data into the `marketdata.currency_rates` table:

```sql
CREATE TABLE marketdata.currency_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base_currency VARCHAR(3) NOT NULL,
    target_currency VARCHAR(3) NOT NULL,
    rate DECIMAL(20,8) NOT NULL,
    rate_date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_currency_rate UNIQUE (base_currency, target_currency, rate_date)
);
```

## Configuration

The worker uses the following environment variables:

- **MINIO_ENDPOINT**: MinIO server endpoint (default: `localhost:9000`)
- **MINIO_ACCESS_KEY**: MinIO access key
- **MINIO_SECRET_KEY**: MinIO secret key
- **MINIO_USE_SSL**: Whether to use SSL (default: `false`)
- **MINIO_BUCKET_NAME**: Bucket name (default: `market-data`)

The CSV file must be named `currency_rates.csv` and placed in the configured bucket.

## Worker Behavior

### Startup
- Runs immediately on service startup
- Processes any existing `currency_rates.csv` file in MinIO

### Scheduled Execution
- Runs every 24 hours (daily at midnight)
- Automatically processes updated CSV files

### Data Processing
1. Fetches `currency_rates.csv` from MinIO
2. Validates each row:
   - Currency codes must be exactly 3 characters
   - Rate must be a valid decimal number
   - Date must be in YYYY-MM-DD format
3. Batch inserts valid records (1000 per batch)
4. On conflict (duplicate currency pair + date), updates the rate

### Error Handling
- Invalid rows are logged and skipped
- Processing continues even if individual rows fail validation
- Database errors are logged and returned

## Usage Example

### Upload CSV to MinIO
```bash
# Using MinIO client (mc)
mc cp currency_rates.csv myminio/market-data/currency_rates.csv
```

### Monitor Worker Logs
```bash
# Check marketdata-service logs
docker logs -f marketdata-service

# Expected output:
# Currency Worker: Starting ingestion...
# Currency Worker: Successfully ingested 150 currency rates.
```

### Query Ingested Data
```sql
-- Get latest rates for USD
SELECT * FROM marketdata.currency_rates 
WHERE base_currency = 'USD' 
ORDER BY rate_date DESC;

-- Get specific currency pair rate
SELECT rate FROM marketdata.currency_rates 
WHERE base_currency = 'USD' 
  AND target_currency = 'EUR' 
  AND rate_date = '2025-11-25';
```

## Testing

A sample CSV file is provided at:
```
services/marketdata-service/sample_currency_rates.csv
```

To test the worker:
1. Upload the sample CSV to MinIO
2. Restart the marketdata-service or wait for the next scheduled run
3. Check the logs for successful ingestion
4. Query the database to verify the data

## Metrics

The worker records database query metrics using Prometheus:
- Query duration
- Success/failure rates
- Batch insert performance

Access metrics at: `http://localhost:9099/metrics`

## Implementation Details

### Files
- **Worker**: `internal/worker/currency_ingestion.go`
- **Domain Model**: `internal/domain/marketdata.go` (CurrencyRate struct)
- **Repository**: `internal/repository/postgres_repo.go` (InsertCurrencyRates method)
- **Main**: `cmd/server/main.go` (worker initialization)

### Key Functions
- `NewCurrencyIngestionWorker()`: Initializes the worker with MinIO client
- `Start()`: Starts the background goroutine with ticker
- `processFile()`: Fetches and parses the CSV file
- `batchInsert()`: Validates and inserts currency rates in batches
