# excel2csv - Excel to CSV Converter

A command-line tool to convert Excel (.xlsx) files to CSV format with advanced features.

## Features

- ✅ Convert all sheets or specific sheets from Excel files
- ✅ Automatic data trimming (removes leading/trailing empty rows and columns)
- ✅ Proper CSV formatting with quoted fields
- ✅ Handles dates, times, and currencies correctly
- ✅ Sanitizes sheet names for valid filenames
- ✅ Comprehensive error handling

## Installation

### Build from source

```bash
cd tools/excel2csv
go build -o excel2csv excel2csv.go
```

### Install globally

```bash
cd tools/excel2csv
go install
```

## Usage

### Basic Usage (Convert all sheets)

```bash
excel2csv input.xlsx
```

This will create CSV files for all sheets:
- `input_Sheet1.csv`
- `input_Sheet2.csv`
- etc.

### Convert Specific Sheets

```bash
excel2csv --sheets "Sheet1,Summary" input.xlsx
```

This will only convert the specified sheets:
- `input_Sheet1.csv`
- `input_Summary.csv`

### Examples

```bash
# Convert all sheets in report.xlsx
excel2csv report.xlsx

# Convert only "Data" and "Summary" sheets
excel2csv --sheets "Data,Summary" report.xlsx

# Convert a file in a different directory
excel2csv /path/to/data/report.xlsx

# Get help
excel2csv --help
```

## Output

- **File Naming**: `<original_filename>_<sheet_name>.csv`
- **Location**: Same directory as the input Excel file
- **Format**: Standard CSV with comma separator
- **Encoding**: UTF-8

## Features in Detail

### Data Trimming

The tool automatically:
- Removes leading empty rows and columns
- Removes trailing empty rows and columns
- Starts conversion from the first cell with actual data

### Cell Formatting

- **Dates**: Converted to readable format (e.g., `2023-10-27`)
- **Times**: Converted to readable format
- **Currencies**: Formatted values preserved
- **Numbers**: Decimal precision maintained

### CSV Standards

- Fields containing commas are quoted
- Fields containing newlines are quoted
- Fields containing double quotes are escaped and quoted
- Standard comma (`,`) separator

### Error Handling

The tool will:
- Exit with error if input file doesn't exist
- Exit with error if file is not a valid .xlsx file
- Warn and skip if a requested sheet doesn't exist
- Continue processing other sheets if one fails

## Exit Codes

- `0` - Success
- `1` - Error (invalid file, no sheets converted, etc.)

## Requirements

- Go 1.24 or higher
- Valid .xlsx Excel file (Excel 2007+)

## Dependencies

- [excelize](https://github.com/xuri/excelize) v2.9.0 - Excel file processing

## Troubleshooting

### "input file does not exist"

Ensure the file path is correct and the file exists.

```bash
# Check if file exists
ls -la report.xlsx
```

### "failed to open Excel file"

The file may be:
- Corrupted
- Not a valid .xlsx file (try opening in Excel first)
- In use by another program

### "sheet 'X' not found"

The sheet name is case-sensitive. Check the exact sheet name in Excel.

```bash
# Convert all sheets to see available names
excel2csv report.xlsx
```

### Permission denied

Ensure you have:
- Read permission on the input file
- Write permission on the output directory

```bash
# Check permissions
ls -la report.xlsx
ls -ld /path/to/output/directory
```

## Examples with Sample Data

### Example 1: Financial Report

```bash
# Input: financial_report.xlsx with sheets: "Income", "Expenses", "Balance"
excel2csv financial_report.xlsx

# Output:
# - financial_report_Income.csv
# - financial_report_Expenses.csv
# - financial_report_Balance.csv
```

### Example 2: Selective Conversion

```bash
# Only convert specific sheets
excel2csv --sheets "Q4 Data,Summary" annual_report.xlsx

# Output:
# - annual_report_Q4_Data.csv
# - annual_report_Summary.csv
```

### Example 3: Batch Processing

```bash
# Convert multiple Excel files
for file in *.xlsx; do
    echo "Converting $file..."
    excel2csv "$file"
done
```

## Advanced Usage

### Integration with Other Tools

```bash
# Convert and immediately process with another tool
excel2csv data.xlsx && process_csv data_Sheet1.csv

# Convert and upload to database
excel2csv transactions.xlsx
psql -d mydb -c "\COPY transactions FROM 'transactions_Data.csv' CSV HEADER"
```

### Scripting

```bash
#!/bin/bash
# Convert Excel file and validate output

INPUT="report.xlsx"
excel2csv "$INPUT"

if [ $? -eq 0 ]; then
    echo "✓ Conversion successful"
    # Process CSV files
    for csv in report_*.csv; do
        echo "Processing $csv..."
        # Your processing logic here
    done
else
    echo "✗ Conversion failed"
    exit 1
fi
```

## Limitations

- Only supports .xlsx format (Excel 2007+)
- Does not support .xls (older Excel format)
- Does not preserve formulas (converts to values)
- Does not preserve formatting (colors, fonts, etc.)
- Maximum file size depends on available memory

## Contributing

To modify or extend the tool:

1. Edit `excel2csv.go`
2. Test your changes:
   ```bash
   go run excel2csv.go test.xlsx
   ```
3. Rebuild:
   ```bash
   go build -o excel2csv excel2csv.go
   ```

## License

Part of the Portfolio Insights project.

## Support

For issues or questions, refer to the main project documentation.
