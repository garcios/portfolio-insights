# CSV Upload Quick Reference

## Endpoint
```
POST http://localhost:8081/upload-csv
```

## Quick Start
```bash
curl -X POST \
  -F "file=@transactions.csv" \
  "http://localhost:8081/upload-csv?user_id=YOUR_USER_ID"
```

## CSV Format
```csv
symbol,executed_at,quantity,price_per_share,type
AAPL,2024-01-15,100,150.50,BUY
GOOGL,2024-01-16,50,2800.00,BUY
```

## Required Columns
- `symbol` - Asset symbol (e.g., AAPL)
- `executed_at` - Date (YYYY-MM-DD)
- `quantity` - Positive number
- `price_per_share` - Positive number
- `type` - BUY or SELL

## Response Codes
- `200` - All records successful
- `206` - Partial success (some rows failed)
- `400` - Bad request (invalid file/missing user_id)

## Common Errors
| Error | Solution |
|-------|----------|
| "user_id is required" | Add `?user_id=XXX` to URL |
| "file must be a CSV" | Use `.csv` file extension |
| "missing required column" | Check CSV headers |
| "symbol X does not exist" | Add asset to marketdata-service first |

## Test
```bash
cd services/transaction-service
./test_csv_upload.sh
```

## Documentation
See `CSV_UPLOAD.md` for full documentation.
