# Portfolio Backfill Admin Collection (gRPC)

This Postman collection provides administrative gRPC endpoints for backfilling portfolio history snapshots.

## Overview

The `BackfillHistory` endpoint allows administrators to create historical portfolio value snapshots for users. This is useful for:
- Populating historical data after setting up the system
- Filling in gaps in portfolio history
- Regenerating snapshots with corrected data

## Prerequisites

### 1. Import Proto Files into Postman

Postman's native gRPC support requires importing the proto files:

1. Open Postman
2. Go to **APIs** in the left sidebar
3. Click **Import** → **Upload Files**
4. Navigate to your project's `proto/portfolio/portfolio.proto` file
5. Import the file (Postman will automatically detect dependencies)

Alternatively, you can set the `proto_path` variable to point to your proto directory.

### 2. Set Admin Token

The portfolio-service must have the `ADMIN_TOKEN` environment variable set. This token is required for authentication.

**On the portfolio-service:**
```bash
export ADMIN_TOKEN="your-secret-admin-token-here"
```

Or in your `docker-compose.yml`:
```yaml
services:
  portfolio-service:
    environment:
      - ADMIN_TOKEN=your-secret-admin-token-here
```

### 3. Configure Postman Variables

In Postman, set the following variables (either in the collection or in an environment):

- `portfolio_host`: The hostname of your portfolio-service (default: `localhost`)
- `portfolio_port`: The port of your portfolio-service (default: `50052`)
- `proto_path`: Absolute path to your proto directory (e.g., `/Users/yourname/portfolio-insights/proto`)
- `admin_token`: Your admin token (must match the `ADMIN_TOKEN` on the server)

**Note**: The `proto_path` variable is optional if you've imported the proto files directly into Postman's API section.

## Using gRPC in Postman

This collection uses Postman's native gRPC support. Here's how to use it:

1. **Import the Collection**: Import `portfolio_backfill.postman_collection.json` into Postman
2. **Select a Request**: Choose one of the BackfillHistory requests
3. **Configure the Server**: The URL will be `grpc://localhost:50052` (or your configured host/port)
4. **Select the Method**: The method `portfolio.PortfolioService/BackfillHistory` should be auto-selected
5. **Edit the Message**: Modify the JSON message body with your parameters
6. **Invoke**: Click the "Invoke" button to send the gRPC request

**Note**: Unlike HTTP/REST, gRPC requests in Postman use the proto file definitions to validate and serialize the request/response messages.

## Available Requests

### 1. BackfillHistory - Single User

Backfill portfolio history for a specific user.

**Request Body:**
```json
{
    "user_id": "user-123",
    "start_date": "2023-01-01",
    "end_date": "2023-12-31",
    "dry_run": false,
    "admin_token": "{{admin_token}}"
}
```

**Use Case:** Backfill an entire year of data for one user.

### 2. BackfillHistory - All Users

Backfill portfolio history for all users with holdings.

**Request Body:**
```json
{
    "user_id": "",
    "start_date": "2023-01-01",
    "end_date": "2023-01-31",
    "dry_run": false,
    "admin_token": "{{admin_token}}"
}
```

**Use Case:** Backfill data for all users. Leave `user_id` empty to process everyone.

### 3. BackfillHistory - Dry Run

Preview what would be backfilled without actually creating snapshots.

**Request Body:**
```json
{
    "user_id": "user-123",
    "start_date": "2023-01-01",
    "end_date": "2023-01-07",
    "dry_run": true,
    "admin_token": "{{admin_token}}"
}
```

**Use Case:** Test the backfill operation before running it. Returns counts without creating data.

### 4. BackfillHistory - Recent Week

Backfill a recent week of data.

**Request Body:**
```json
{
    "user_id": "user-123",
    "start_date": "2024-01-01",
    "end_date": "2024-01-07",
    "dry_run": false,
    "admin_token": "{{admin_token}}"
}
```

**Use Case:** Fill in a small date range, useful for testing or filling gaps.

### 5. BackfillHistory - Default End Date

Backfill from a start date to today.

**Request Body:**
```json
{
    "user_id": "user-123",
    "start_date": "2024-01-01",
    "end_date": "",
    "dry_run": false,
    "admin_token": "{{admin_token}}"
}
```

**Use Case:** Backfill from a specific date until now. Leave `end_date` empty to default to today.

## Request Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `user_id` | string | No | User ID to backfill. Leave empty for all users. |
| `start_date` | string | Yes | Start date in `YYYY-MM-DD` format. |
| `end_date` | string | No | End date in `YYYY-MM-DD` format. Defaults to today if empty. |
| `dry_run` | boolean | No | If `true`, preview without creating snapshots. Default: `false`. |
| `admin_token` | string | Yes | Admin authentication token. |

## Response Format

```json
{
    "snapshots_created": 365,
    "snapshots_skipped": 0,
    "errors": 0,
    "error_messages": [],
    "status": "success"
}
```

**Response Fields:**
- `snapshots_created`: Number of new snapshots created
- `snapshots_skipped`: Number of snapshots that already existed
- `errors`: Number of errors encountered
- `error_messages`: Array of error messages (if any)
- `status`: Overall status (`success`, `partial`, or `failed`)

## Important Notes

### Historical Prices

The backfill uses **historical prices** from the marketdata-service. Ensure you have:
1. Historical price data loaded in the `marketdata.asset_prices` table
2. Historical currency rates loaded in the `marketdata.currency_rates` table

### Current Holdings Assumption

The backfill assumes **current holdings** were held on the historical dates. This means:
- If a user bought a stock in 2024, backfilling 2023 will show that stock in their 2023 portfolio
- For accurate historical data, you would need to replay transactions (not currently implemented)

### Duplicate Prevention

The endpoint automatically skips dates that already have snapshots, so it's safe to run multiple times.

### Performance Considerations

- Backfilling large date ranges for many users can take time
- Use dry run first to estimate the workload
- Consider backfilling in smaller batches (e.g., month by month)

## Example Workflow

1. **Test with Dry Run:**
   ```json
   {
       "user_id": "user-123",
       "start_date": "2023-01-01",
       "end_date": "2023-01-31",
       "dry_run": true,
       "admin_token": "{{admin_token}}"
   }
   ```

2. **Review the response** to see how many snapshots would be created

3. **Run the actual backfill:**
   ```json
   {
       "user_id": "user-123",
       "start_date": "2023-01-01",
       "end_date": "2023-01-31",
       "dry_run": false,
       "admin_token": "{{admin_token}}"
   }
   ```

4. **Check the response** for any errors

5. **Repeat for other date ranges** as needed

## Troubleshooting

### "invalid admin token" Error

- Ensure the `admin_token` in your request matches the `ADMIN_TOKEN` environment variable on the portfolio-service
- Check that the portfolio-service has been restarted after setting the environment variable

### "no price found" Warnings

- Ensure historical price data is loaded in the marketdata-service
- Check the `marketdata.asset_prices` table for the relevant symbols and dates

### "no currency rate found" Warnings

- Ensure historical currency rate data is loaded
- Check the `marketdata.currency_rates` table for the relevant currency pairs and dates

## Security

⚠️ **Important Security Notes:**

- The admin token should be kept secret and not committed to version control
- Only use this endpoint in secure, administrative contexts
- Consider implementing additional authentication mechanisms (JWT, OAuth) for production use
- Rotate the admin token regularly

## Related Documentation

- [Portfolio History Strategy](../PORTFOLIO_HISTORY_STRATEGY.md)
- [Main Postman Collection](./portfolio_insights.postman_collection.json)
