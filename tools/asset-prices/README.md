# Asset Prices Tool

This tool downloads historical price data and currency exchange rates from the EODHD (End of Day Historical Data) API.

## Prerequisites

### API Token

You need an EODHD API token to use this tool. Get one from [https://eodhd.com/](https://eodhd.com/).

Set your API token as an environment variable:

```bash
export EODHD_API_TOKEN="your_api_token_here"
```

Alternatively, you can pass it directly using the `-token` flag when running the command.

## Quick Start

### Download Stock Prices

```bash
make download-prices
```

This downloads the latest prices for all tickers listed in `tickers.txt` and saves them to `prices.csv`.

### Download Currency Exchange Rates

```bash
make download-currency-rates
```

This downloads currency exchange rates for all forex pairs listed in `forex.txt` and saves them to `currency_rates.csv`.

## Configuration Files

### `tickers.txt`

List stock/ETF tickers you want to download, one per line.

**Format:**
- US stocks: Use the ticker symbol (e.g., `GOOGL`, `MSFT`)
- International stocks: Add exchange suffix (e.g., `IVV.AU` for Australian exchange)

**Example:**
```
GOOGL
AMZN
MSFT
IVV.AU
NDQ.AU
```

### `forex.txt`

List forex pairs you want to download, one per line.

**Format:**
- Use the format: `{CURRENCY}.FOREX` (e.g., `AUD.FOREX` for AUD/USD exchange rate)

**Example:**
```
AUD.FOREX
EUR.FOREX
GBP.FOREX
```

## Output Files

### `prices.csv`

Stock/ETF price data in the format required by the marketdata-service.

**Format:**
```csv
symbol,price,timestamp
GOOGL,319.95,2025-11-26
AMZN,229.16,2025-11-26
```

**Columns:**
- `symbol`: Stock ticker (exchange suffix removed, e.g., `IVV.AU` → `IVV`)
- `price`: Adjusted close price
- `timestamp`: Date in YYYY-MM-DD format

### `currency_rates.csv`

Currency exchange rate data in the format required by the marketdata-service.

**Format:**
```csv
base_currency,target_currency,rate,rate_date
USD,AUD,1.47,2023-01-02
USD,EUR,0.93,2023-01-02
```

**Columns:**
- `base_currency`: Base currency (always `USD`)
- `target_currency`: Target currency (e.g., `AUD`, `EUR`)
- `rate`: Exchange rate (how many target currency units per 1 base currency unit)
- `rate_date`: Date in YYYY-MM-DD format

## Advanced Usage

### Manual Command Execution

You can run the tool directly with custom parameters:

```bash
go run eodhd/asset_prices_eodhd.go \
  -tickers=tickers.txt \
  -from=20230101 \
  -to=20251127 \
  -output=prices.csv
```

### Command-Line Flags

| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `-tickers` | Path to tickers file | `tickers.txt` | `-tickers=my_stocks.txt` |
| `-from` | Start date (YYYYMMDD) | `20000101` | `-from=20230101` |
| `-to` | End date (YYYYMMDD) | Current date | `-to=20251127` |
| `-output` | Output CSV file | `all_prices.csv` | `-output=prices.csv` |
| `-token` | EODHD API token | From `EODHD_API_TOKEN` env var | `-token=your_token` |

### Date Range Examples

**Download today's prices only:**
```bash
go run eodhd/asset_prices_eodhd.go \
  -tickers=tickers.txt \
  -from=20251126 \
  -to=20251126 \
  -output=prices.csv
```

**Download historical data for the past year:**
```bash
go run eodhd/asset_prices_eodhd.go \
  -tickers=tickers.txt \
  -from=20240101 \
  -to=20251127 \
  -output=historical_prices.csv
```

## How It Works

1. **Reads tickers** from the specified file (one ticker per line)
2. **Fetches data** from EODHD API for each ticker
   - Sleeps 1 second between requests to respect rate limits
   - Automatically detects forex data (tickers containing `.FOREX`)
3. **Processes data:**
   - Removes exchange suffixes (`.AU`, `.FOREX`, etc.)
   - Deduplicates entries (by ticker + date)
   - Sorts by date, then ticker
4. **Writes output:**
   - Stock data → `symbol,price,timestamp` format
   - Forex data → `base_currency,target_currency,rate,rate_date` format

## Troubleshooting

### Error: API token is required

**Problem:** The EODHD API token is not set.

**Solution:**
```bash
export EODHD_API_TOKEN="your_api_token_here"
```

### Error: bad status: 401 Unauthorized

**Problem:** Invalid API token.

**Solution:** Check that your API token is correct and active.

### Error: bad status: 404 Not Found

**Problem:** Invalid ticker symbol or exchange suffix.

**Solution:** 
- Verify ticker symbols are correct
- For international stocks, ensure you're using the correct exchange suffix (e.g., `.AU` for Australia)
- Check EODHD documentation for supported exchanges

### No data returned for a ticker

**Problem:** The ticker might not be available on EODHD or the date range has no data.

**Solution:**
- Verify the ticker exists on EODHD
- Check the date range is valid
- Some tickers may have limited historical data

## Integration with MarketData Service

The output CSV files are designed to be uploaded to MinIO and processed by the marketdata-service workers:

1. **Upload to MinIO:**
   ```bash
   # Upload prices.csv to the market-data bucket
   mc cp prices.csv minio/market-data/price.csv
   
   # Upload currency_rates.csv to the market-data bucket
   mc cp currency_rates.csv minio/market-data/currency_rate.csv
   ```

2. **Workers will automatically process:**
   - `price_ingestion.go` worker processes `price.csv`
   - `currency_ingestion.go` worker processes `currency_rate.csv`

## API Rate Limits

The tool includes a 1-second delay between API requests to respect EODHD rate limits. For large ticker lists, this means:

- 10 tickers ≈ 10 seconds
- 50 tickers ≈ 50 seconds
- 100 tickers ≈ 100 seconds

## Examples

### Example 1: Daily Price Update

Download today's prices for your portfolio:

```bash
# Update tickers.txt with your holdings
echo "GOOGL
AMZN
MSFT" > tickers.txt

# Download today's prices
make download-prices

# Upload to MinIO
mc cp prices.csv minio/market-data/price.csv
```

### Example 2: Historical Backfill

Download 2 years of historical data:

```bash
go run eodhd/asset_prices_eodhd.go \
  -tickers=tickers.txt \
  -from=20230101 \
  -to=20251127 \
  -output=historical_prices.csv
```

### Example 3: Multiple Currency Rates

Download multiple forex pairs:

```bash
# Create forex.txt with multiple currencies
echo "AUD.FOREX
EUR.FOREX
GBP.FOREX
JPY.FOREX" > forex.txt

# Download rates
make download-currency-rates
```

## Support

For issues with the EODHD API, refer to their documentation: [https://eodhd.com/financial-apis/](https://eodhd.com/financial-apis/)

For issues with this tool, check the error messages and troubleshooting section above.
