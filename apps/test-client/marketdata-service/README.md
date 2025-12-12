# Market Data Service Test Client

A gRPC test client for the Market Data Service that allows you to test all available RPC methods.

## Prerequisites

- Go 1.24.0 or later
- Market Data Service running (default: `localhost:50054`)

## Installation

```bash
cd apps/test-client/marketdata-service
go mod download
go build -o marketdata-client
```

## Usage

The test client supports eight operations: `get-asset`, `get-asset-by-symbol`, `list-assets`, `get-price`, `get-prices`, `get-historical`, `get-currency-rate`, and `test-errors`.

### Get Asset by Resource Name

Retrieve an asset by its resource name:

```bash
./marketdata-client -op get-asset -asset-name assets/AAPL
```

With custom server address:

```bash
./marketdata-client -addr localhost:50054 -op get-asset -asset-name assets/AAPL
```

### Get Asset by Symbol

Retrieve an asset by its symbol (more convenient than resource name):

```bash
./marketdata-client -op get-asset-by-symbol -symbol AAPL
```

### List Assets

List all available assets with pagination:

```bash
./marketdata-client -op list-assets -page-size 20
```

With pagination token:

```bash
./marketdata-client -op list-assets -page-size 20 -page-token "next-page-token"
```

### Get Latest Price

Get the latest price for a specific asset:

```bash
./marketdata-client -op get-price -asset-name assets/AAPL
```

### Get Latest Prices (Batch)

Get latest prices for multiple assets in a single call:

```bash
./marketdata-client -op get-prices -asset-names "assets/AAPL,assets/GOOGL,assets/MSFT"
```

### Get Historical Prices

Get historical price data for an asset within a time range:

```bash
./marketdata-client -op get-historical \
    -asset-name assets/AAPL \
    -start-time "2024-01-01T00:00:00Z" \
    -end-time "2024-01-07T00:00:00Z" \
    -interval "1d"
```

### Get Currency Exchange Rate

Get the latest exchange rate between two currencies:

```bash
./marketdata-client -op get-currency-rate \
    -base-currency USD \
    -target-currency EUR
```

### Test Error Handling

Run a comprehensive suite of error tests to validate input validation and error responses:

```bash
./marketdata-client -op test-errors
```

This will run 12 different error test cases including:
- Empty and invalid resource names
- Missing required fields
- Non-existent assets and symbols
- Invalid time ranges
- Empty currency codes

Use the `-verbose` flag to see additional error details:

```bash
./marketdata-client -op test-errors -verbose
```

## Command-Line Flags

| Flag | Description | Default | Required |
|------|-------------|---------|----------|
| `-addr` | Server address (host:port) | `localhost:50054` | No |
| `-op` | Operation to perform | `get-asset-by-symbol` | No |
| `-asset-name` | Asset resource name (e.g., `assets/AAPL`) | - | Yes (for get-asset, get-price, get-historical) |
| `-symbol` | Asset symbol (e.g., `AAPL`) | - | Yes (for get-asset-by-symbol) |
| `-asset-names` | Comma-separated asset names | - | Yes (for get-prices) |
| `-page-size` | Page size for list operations | `50` | No |
| `-page-token` | Page token for pagination | - | No |
| `-start-time` | Start time (RFC3339 format) | - | Yes (for get-historical) |
| `-end-time` | End time (RFC3339 format) | - | Yes (for get-historical) |
| `-interval` | Data interval (e.g., `1d`, `1h`) | `1d` | No |
| `-base-currency` | Base currency code (e.g., `USD`) | - | Yes (for get-currency-rate) |
| `-target-currency` | Target currency code (e.g., `EUR`) | - | Yes (for get-currency-rate) |
| `-verbose` | Enable verbose error output | `false` | No |

## Examples

### Complete Workflow

1. **Get an asset by symbol:**
   ```bash
   ./marketdata-client -op get-asset-by-symbol -symbol AAPL
   ```
   
   Output:
   ```
   Getting asset by symbol: AAPL
   
   === Asset Details ===
   Resource Name: assets/AAPL
   Asset ID:      AAPL
   Symbol:        AAPL
   Display Name:  Apple Inc.
   Type:          EQUITY
   Exchange:      NASDAQ
   Currency:      USD
   =====================
   ```

2. **Get the latest price:**
   ```bash
   ./marketdata-client -op get-price -asset-name assets/AAPL
   ```
   
   Output:
   ```
   Getting latest price for: assets/AAPL
   
   === Asset Price ===
   Asset:     assets/AAPL
   Price:     $150.25
   Timestamp: 2024-01-15T16:00:00Z
   ===================
   ```

3. **List available assets:**
   ```bash
   ./marketdata-client -op list-assets -page-size 5
   ```

4. **Get batch prices:**
   ```bash
   ./marketdata-client -op get-prices -asset-names "assets/AAPL,assets/GOOGL,assets/MSFT"
   ```

### Testing Against Different Environments

**Local development:**
```bash
./marketdata-client -addr localhost:50054 -op get-asset-by-symbol -symbol AAPL
```

**Docker environment:**
```bash
./marketdata-client -addr marketdata-service:50054 -op get-asset-by-symbol -symbol AAPL
```

## Error Handling

The client provides clear error messages for common issues:

- **Connection failures:** Check if the server is running and the address is correct
- **Missing required fields:** The client will indicate which fields are required for each operation
- **Invalid resource names:** Resource names must follow the format `assets/{asset}`
- **Asset not found:** Get operations will return an error if the asset doesn't exist
- **Invalid time ranges:** Historical data requests require valid start and end times

## Quick Start Scripts

### Run Examples

```bash
chmod +x examples.sh
./examples.sh
```

This script demonstrates all major operations with real examples.

### Run Error Tests

```bash
chmod +x demo-errors.sh
./demo-errors.sh
```

This script demonstrates error handling and validation.

## Development

To run without building:

```bash
go run main.go -op get-asset-by-symbol -symbol AAPL
```

To build for different platforms:

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o marketdata-client-linux

# macOS
GOOS=darwin GOARCH=amd64 go build -o marketdata-client-macos

# Windows
GOOS=windows GOARCH=amd64 go build -o marketdata-client.exe
```

## Time Format

All timestamps use RFC3339 format: `YYYY-MM-DDTHH:MM:SSZ`

Examples:
- `2024-01-01T00:00:00Z`
- `2024-12-31T23:59:59Z`

You can generate timestamps using:

```bash
# Current time
date -u +"%Y-%m-%dT%H:%M:%SZ"

# 7 days ago (macOS)
date -u -v-7d +"%Y-%m-%dT%H:%M:%SZ"

# 7 days ago (Linux)
date -u -d "7 days ago" +"%Y-%m-%dT%H:%M:%SZ"
```
