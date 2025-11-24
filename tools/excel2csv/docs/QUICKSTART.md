# excel2csv - Quick Start Guide

## Installation

```bash
cd tools/excel2csv
make build
```

The binary will be created as `./excel2csv`

## Basic Usage

### Convert all sheets

```bash
./excel2csv myfile.xlsx
```

Output:
- `myfile_Sheet1.csv`
- `myfile_Sheet2.csv`
- etc.

### Convert specific sheets

```bash
./excel2csv --sheets "Data,Summary" myfile.xlsx
```

Output:
- `myfile_Data.csv`
- `myfile_Summary.csv`

## Common Use Cases

### 1. Convert Financial Data

```bash
# Convert all sheets from a financial report
./excel2csv financial_report_2024.xlsx

# Result:
# - financial_report_2024_Income.csv
# - financial_report_2024_Expenses.csv
# - financial_report_2024_Balance.csv
```

### 2. Selective Sheet Conversion

```bash
# Only convert specific sheets
./excel2csv --sheets "Q4 Results,Summary" annual_report.xlsx
```

### 3. Batch Processing

```bash
# Convert all Excel files in a directory
for file in *.xlsx; do
    echo "Converting $file..."
    ./excel2csv "$file"
done
```

### 4. Integration with CSV Upload

```bash
# Convert Excel to CSV, then upload to transaction service
./excel2csv transactions.xlsx

# Upload the generated CSV
curl -X POST \
  -F "file=@transactions_Data.csv" \
  "http://localhost:8081/upload-csv?user_id=user-123"
```

## Features

✅ Automatic data trimming (removes empty rows/columns)  
✅ Proper date/time formatting  
✅ Currency formatting preserved  
✅ Handles special characters in sheet names  
✅ Standard CSV format with proper quoting  

## Troubleshooting

### Build fails

```bash
# Clean and rebuild
make clean
make build
```

### "input file does not exist"

Check the file path:
```bash
ls -la myfile.xlsx
```

### "failed to open Excel file"

Ensure the file is a valid .xlsx (Excel 2007+) file.

## Full Documentation

See [README.md](README.md) for complete documentation.
