package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

const (
	exitSuccess = 0
	exitError   = 1
)

// Config holds the CLI configuration
type Config struct {
	inputFile string
	sheets    []string
}

func main() {
	config, err := parseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitError)
	}

	if err := run(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitError)
	}

	os.Exit(exitSuccess)
}

// parseFlags parses command-line flags and arguments
func parseFlags() (*Config, error) {
	sheetsFlag := flag.String("sheets", "", "Comma-separated list of sheet names to convert (optional)")
	flag.Parse()

	// Get input file from positional argument
	args := flag.Args()
	if len(args) < 1 {
		return nil, fmt.Errorf("usage: excel2csv [--sheets \"Sheet1,Sheet2\"] <input.xlsx>")
	}

	inputFile := args[0]

	// Validate input file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("input file does not exist: %s", inputFile)
	}

	// Parse sheets if provided
	var sheets []string
	if *sheetsFlag != "" {
		sheets = strings.Split(*sheetsFlag, ",")
		for i := range sheets {
			sheets[i] = strings.TrimSpace(sheets[i])
		}
	}

	return &Config{
		inputFile: inputFile,
		sheets:    sheets,
	}, nil
}

// run executes the main conversion logic
func run(config *Config) error {
	// Open Excel file
	f, err := excelize.OpenFile(config.inputFile)
	if err != nil {
		return fmt.Errorf("failed to open Excel file: %w (ensure it's a valid .xlsx file)", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close Excel file: %v\n", err)
		}
	}()

	// Get all sheet names
	allSheets := f.GetSheetList()
	if len(allSheets) == 0 {
		return fmt.Errorf("no sheets found in Excel file")
	}

	// Determine which sheets to convert
	sheetsToConvert := config.sheets
	if len(sheetsToConvert) == 0 {
		sheetsToConvert = allSheets
	} else {
		// Validate requested sheets exist
		sheetsToConvert = validateSheets(sheetsToConvert, allSheets)
	}

	if len(sheetsToConvert) == 0 {
		return fmt.Errorf("no valid sheets to convert")
	}

	// Get base filename and directory
	baseFilename := getBaseFilename(config.inputFile)
	outputDir := filepath.Dir(config.inputFile)

	// Convert each sheet
	successCount := 0
	for _, sheetName := range sheetsToConvert {
		outputFile := filepath.Join(outputDir, fmt.Sprintf("%s_%s.csv", baseFilename, sanitizeFilename(sheetName)))

		if err := convertSheet(f, sheetName, outputFile); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to convert sheet '%s': %v\n", sheetName, err)
			continue
		}

		fmt.Printf("✓ Converted sheet '%s' to %s\n", sheetName, outputFile)
		successCount++
	}

	if successCount == 0 {
		return fmt.Errorf("failed to convert any sheets")
	}

	fmt.Printf("\nSuccessfully converted %d/%d sheets\n", successCount, len(sheetsToConvert))
	return nil
}

// validateSheets checks which requested sheets exist and returns valid ones
func validateSheets(requested []string, available []string) []string {
	availableMap := make(map[string]bool)
	for _, sheet := range available {
		availableMap[sheet] = true
	}

	var valid []string
	for _, sheet := range requested {
		if availableMap[sheet] {
			valid = append(valid, sheet)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: sheet '%s' not found, skipping\n", sheet)
		}
	}

	return valid
}

// getBaseFilename extracts the filename without extension
func getBaseFilename(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// sanitizeFilename removes or replaces characters that are problematic in filenames
func sanitizeFilename(name string) string {
	// Replace problematic characters with underscores
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	)
	return replacer.Replace(name)
}

// convertSheet converts a single Excel sheet to CSV
func convertSheet(f *excelize.File, sheetName, outputFile string) error {
	// Get all rows from the sheet
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("failed to read rows: %w", err)
	}

	if len(rows) == 0 {
		return fmt.Errorf("sheet is empty")
	}

	// Trim empty rows and columns
	trimmedRows := trimEmptyRowsAndColumns(rows)
	if len(trimmedRows) == 0 {
		return fmt.Errorf("sheet contains no data after trimming")
	}

	// Create output CSV file
	csvFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer csvFile.Close()

	// Create CSV writer
	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Write all rows to CSV
	for _, row := range trimmedRows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	if err := writer.Error(); err != nil {
		return fmt.Errorf("CSV writer error: %w", err)
	}

	return nil
}

// trimEmptyRowsAndColumns removes leading and trailing empty rows and columns
func trimEmptyRowsAndColumns(rows [][]string) [][]string {
	if len(rows) == 0 {
		return rows
	}

	// Find first non-empty row
	firstRow := -1
	for i, row := range rows {
		if !isEmptyRow(row) {
			firstRow = i
			break
		}
	}

	if firstRow == -1 {
		return [][]string{} // All rows are empty
	}

	// Find last non-empty row
	lastRow := -1
	for i := len(rows) - 1; i >= 0; i-- {
		if !isEmptyRow(rows[i]) {
			lastRow = i
			break
		}
	}

	// Trim rows
	trimmedRows := rows[firstRow : lastRow+1]

	// Find first non-empty column
	firstCol := -1
	maxCols := getMaxColumns(trimmedRows)
	for col := 0; col < maxCols; col++ {
		if !isEmptyColumn(trimmedRows, col) {
			firstCol = col
			break
		}
	}

	if firstCol == -1 {
		return [][]string{} // All columns are empty
	}

	// Find last non-empty column
	lastCol := -1
	for col := maxCols - 1; col >= 0; col-- {
		if !isEmptyColumn(trimmedRows, col) {
			lastCol = col
			break
		}
	}

	// Trim columns
	result := make([][]string, len(trimmedRows))
	for i, row := range trimmedRows {
		// Ensure row has enough columns
		if len(row) <= firstCol {
			result[i] = []string{}
			continue
		}

		endCol := lastCol + 1
		if endCol > len(row) {
			endCol = len(row)
		}

		result[i] = row[firstCol:endCol]
	}

	return result
}

// isEmptyRow checks if a row contains only empty cells
func isEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

// isEmptyColumn checks if a column contains only empty cells
func isEmptyColumn(rows [][]string, col int) bool {
	for _, row := range rows {
		if col < len(row) && strings.TrimSpace(row[col]) != "" {
			return false
		}
	}
	return true
}

// getMaxColumns returns the maximum number of columns across all rows
func getMaxColumns(rows [][]string) int {
	max := 0
	for _, row := range rows {
		if len(row) > max {
			max = len(row)
		}
	}
	return max
}
