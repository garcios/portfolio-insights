# CSV Transaction Upload Feature

## Overview

The transaction-service now supports bulk transaction uploads via CSV files. This feature allows users to import historical transaction data efficiently.

## API Endpoint

**POST** `/upload-csv`

### Request

- **Method**: `POST`
- **Content-Type**: `multipart/form-data`
- **Port**: `8081` (HTTP)

### Parameters

- **user_id** (required): Can be provided via:
  - Query parameter: `?user_id=<user_id>`
  - Header: `X-User-ID: <user_id>`
- **file** (required): CSV file to upload

### CSV Format

#### Required Columns

| Column | Type | Description | Example |
|--------|------|-------------|---------|
| `symbol` | string | Asset symbol (case-insensitive) | AAPL, GOOGL |
| `executed_at` | date/datetime | Transaction execution date | 2024-01-15 or 2024-01-15 10:30:00 |
| `quantity` | float | Number of shares (must be positive) | 100, 25.5 |
| `price_per_share` | float | Price per share (must be positive) | 150.50 |
| `type` | string | Transaction type: BUY or SELL | BUY, SELL |

#### Supported Date Formats

- `2006-01-02` (YYYY-MM-DD)
- `2006-01-02 15:04:05` (YYYY-MM-DD HH:MM:SS)
- `01/02/2006` (MM/DD/YYYY)
- `01-02-2006` (MM-DD-YYYY)
- RFC3339

#### Example CSV

```csv
symbol,executed_at,quantity,price_per_share,type
AAPL,2024-01-15,100,150.50,BUY
GOOGL,2024-01-16,50,2800.00,BUY
MSFT,2024-01-17,75,380.25,BUY
TSLA,2024-01-18,30,245.00,BUY
AAPL,2024-02-01,25,155.00,SELL
```

## Response

### Success Response (200 OK)

```json
{
  "total_records": 5,
  "successful_records": 5,
  "failed_records": 0,
  "errors": []
}
```

### Partial Success Response (206 Partial Content)

```json
{
  "total_records": 5,
  "successful_records": 3,
  "failed_records": 2,
  "errors": [
    {
      "row_number": 3,
      "row": {
        "symbol": "INVALID",
        "executed_at": "2024-01-17",
        "quantity": "75",
        "price_per_share": "380.25",
        "type": "BUY"
      },
      "error": "symbol INVALID does not exist"
    },
    {
      "row_number": 5,
      "row": {
        "symbol": "AAPL",
        "executed_at": "invalid-date",
        "quantity": "25",
        "price_per_share": "155.00",
        "type": "SELL"
      },
      "error": "invalid executed_at format: unable to parse date: invalid-date"
    }
  ]
}
```

### Error Response (400 Bad Request)

```json
{
  "error": "missing required column: symbol"
}
```

## Usage Examples

### Using cURL

```bash
# Upload CSV with user_id as query parameter
curl -X POST \
  -F "file=@transactions.csv" \
  "http://localhost:8081/upload-csv?user_id=user-123"

# Upload CSV with user_id as header
curl -X POST \
  -H "X-User-ID: user-123" \
  -F "file=@transactions.csv" \
  http://localhost:8081/upload-csv
```

### Using Postman

1. Set method to `POST`
2. URL: `http://localhost:8081/upload-csv?user_id=user-123`
3. Go to **Body** tab
4. Select **form-data**
5. Add key `file` with type `File`
6. Choose your CSV file
7. Click **Send**

### Using Python

```python
import requests

url = "http://localhost:8081/upload-csv"
params = {"user_id": "user-123"}
files = {"file": open("transactions.csv", "rb")}

response = requests.post(url, params=params, files=files)
print(response.json())
```

### Using JavaScript (Node.js)

```javascript
const FormData = require('form-data');
const fs = require('fs');
const axios = require('axios');

const form = new FormData();
form.append('file', fs.createReadStream('transactions.csv'));

axios.post('http://localhost:8081/upload-csv?user_id=user-123', form, {
  headers: form.getHeaders()
})
.then(response => console.log(response.data))
.catch(error => console.error(error));
```

## Validation Rules

### File-Level Validation

- File must be in CSV format (`.csv` extension or `text/csv` content type)
- File size limit: 10 MB
- All required columns must be present in the header

### Row-Level Validation

- **User ID**: Must exist in the user-service
- **Symbol**: Must exist in the marketdata-service
- **Quantity**: Must be a positive number
- **Price Per Share**: Must be a positive number
- **Type**: Must be either "BUY" or "SELL" (case-insensitive)
- **Executed At**: Must be a valid date in one of the supported formats

### Error Handling

- **File-level errors**: The entire upload is rejected
- **Row-level errors**: Invalid rows are skipped, valid rows are processed
- A detailed error report is returned with:
  - Row number
  - Row data
  - Error message

## Implementation Details

### Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ POST /upload-csv
       ▼
┌─────────────────────┐
│  HTTP Handler       │
│  (csv_upload_       │
│   handler.go)       │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│  CSV Upload         │
│  Usecase            │
│  (csv_upload_       │
│   usecase.go)       │
└──────┬──────────────┘
       │
       ├──► UserGateway (validate user)
       ├──► MarketDataGateway (validate symbols)
       ├──► TransactionRepository (bulk insert)
       └──► EventPublisher (publish events)
```

### Database Operations

- Uses `BulkCreate` method for efficient batch insertion
- Wrapped in a database transaction for atomicity
- All valid transactions are inserted or none (if transaction fails)

### Event Publishing

- After successful bulk insert, events are published for each transaction
- Event publishing failures are logged but don't fail the upload
- Events trigger portfolio updates in the portfolio-service

## Performance Considerations

- **Batch Size**: No hard limit, but recommended max 1000 rows per file
- **Processing Time**: ~100-200ms for 100 transactions
- **Database**: Uses prepared statements for efficiency
- **Memory**: File is read into memory (10 MB limit)

## Testing

### Sample Test File

A sample CSV file is provided at:
```
services/transaction-service/sample_transactions.csv
```

### Test Scenarios

1. **Valid CSV**: All rows should be processed successfully
2. **Invalid Symbol**: Rows with non-existent symbols should be skipped
3. **Invalid Date**: Rows with invalid dates should be skipped
4. **Missing Column**: Upload should be rejected
5. **Invalid User**: Upload should be rejected

## Troubleshooting

### Common Issues

1. **"user_id is required"**
   - Ensure you're passing `user_id` as query param or header

2. **"file must be a CSV"**
   - Check file extension is `.csv`
   - Verify Content-Type is `text/csv`

3. **"missing required column: X"**
   - Ensure all required columns are in the CSV header
   - Check for typos in column names

4. **"symbol X does not exist"**
   - Ensure the asset exists in marketdata-service
   - Run asset ingestion first if needed

5. **"user X does not exist"**
   - Create the user in user-service first
   - Verify the user_id is correct

## Future Enhancements

- [ ] Support for larger files (streaming)
- [ ] Async processing with job queue
- [ ] Progress tracking for large uploads
- [ ] Support for additional transaction types
- [ ] CSV template download endpoint
- [ ] Duplicate detection
