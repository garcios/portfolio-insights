# csvtxn - Broker CSV to Transaction CSV Converter

A command-line tool to convert broker export CSV files to the transaction service format.

## Purpose

Converts broker transaction exports (with fees, exchange info, etc.) to a simplified CSV format compatible with the transaction upload service.

## Input Format

Broker CSV format (13 columns):
```
Symbol,Exchange,Name,Date,Action,Quantity,Price,Currency,Column8,Fee,FeeCurrency,FXRate,Total
CSL,ASX,Csl Limited,2020-09-08,Buy,5.0000,286.38,AUD,,29.95,AUD,1.00,"1,461.85"
STW,ASX,Spdr S&P/Asx 200 Etf,2020-09-11,Buy,18.0000,54.86,AUD,,29.95,AUD,1.00,"1,017.43"
```

## Output Format

Transaction service CSV format (5 columns):
```
symbol,executed_at,quantity,price_per_share,type
CSL,2020-09-08,5,286.38,BUY
STW,2020-09-11,18,54.86,BUY
```

## Usage

```bash
# Build the tool
GOWORK=off go build -o csvtxn csvtxn.go

# Convert a broker CSV file
./csvtxn AllTradesReport_Combined.csv

# Output will be written to transactions.csv
```

## Features

✅ **Automatic Type Detection**
- Positive quantities → BUY
- Negative quantities → SELL
- Respects Action column (Buy/Sell)

✅ **Data Cleaning**
- Removes commas from numbers
- Removes quotes from values
- Trims whitespace

✅ **Date Handling**
- Supports multiple date formats
- Outputs YYYY-MM-DD format

✅ **Error Handling**
- Skips invalid rows
- Reports warnings for skipped rows
- Continues processing valid rows

## Example

### Input (broker export)
```csv
Symbol,Exchange,Name,Date,Action,Quantity,Price,Currency,,Fee,FeeCurrency,FXRate,Total
CSL,ASX,Csl Limited,2020-09-08,Buy,5.0000,286.38,AUD,,29.95,AUD,1.00,"1,461.85"
AAPL,NASDAQ,Apple Inc,2024-01-15,Buy,100,150.50,USD,,9.95,USD,1.00,"15,059.95"
GOOGL,NASDAQ,Alphabet Inc,2024-02-01,Sell,-25,2800.00,USD,,9.95,USD,1.00,"69,990.05"
```

### Output (transaction service format)
```csv
symbol,executed_at,quantity,price_per_share,type
CSL,2020-09-08,5,286.38,BUY
AAPL,2024-01-15,100,150.50,BUY
GOOGL,2024-02-01,25,2800.00,SELL
```

## Integration with Transaction Upload

```bash
# 1. Convert broker export to transaction format
./csvtxn broker_export.csv

# 2. Upload to transaction service
curl -X POST \
  -F "file=@transactions.csv" \
  "http://localhost:8081/upload-csv?user_id=user-123"
```

## Complete Workflow

```bash
# 1. Export transactions from broker (Excel format)
# 2. Convert Excel to CSV
./excel2csv broker_export.xlsx

# 3. Convert broker CSV to transaction CSV
./csvtxn broker_export_Transactions.csv

# 4. Upload to service
curl -X POST \
  -F "file=@transactions.csv" \
  "http://localhost:8081/upload-csv?user_id=user-123"
```

## Error Messages

| Error | Meaning | Solution |
|-------|---------|----------|
| "insufficient fields" | Row has less than 13 columns | Check CSV format |
| "invalid date" | Date format not recognized | Use YYYY-MM-DD format |
| "quantity cannot be zero" | Quantity is 0 | Remove or fix row |
| "price must be positive" | Price is 0 or negative | Fix price value |

## Supported Date Formats

- `2006-01-02` (YYYY-MM-DD)
- `2006-01-02 15:04:05` (YYYY-MM-DD HH:MM:SS)
- `01/02/2006` (MM/DD/YYYY)
- `02/01/2006` (DD/MM/YYYY)
- `01-02-2006` (MM-DD-YYYY)
- `02-01-2006` (DD-MM-YYYY)
- `2006/01/02` (YYYY/MM/DD)
- RFC3339

## Notes

- Output is always written to `transactions.csv` in the current directory
- Existing `transactions.csv` will be overwritten
- Symbols are converted to uppercase
- Quantities are always positive (type indicates BUY/SELL)
- Fees and exchange info are not included in output

## Troubleshooting

### "no valid transactions found"

Check that:
- Input file has at least one valid row
- Rows have 13 columns
- Dates are in a recognized format
- Quantities and prices are valid numbers

### Negative quantities causing errors

The tool now handles negative quantities automatically:
- Negative quantity → SELL transaction
- Positive quantity → BUY transaction

### Header row being processed

The tool will skip the header row if it can't parse the date.
