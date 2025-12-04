package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	exitSuccess = 0
	exitError   = 1
)

// BrokerTransaction represents a row from the broker CSV
type BrokerTransaction struct {
	Symbol      string
	Exchange    string
	Name        string
	Date        string
	Action      string
	Quantity    string
	Price       string
	Currency    string
	Fee         string
	FeeCurrency string
	FXRate      string
	Total       string
}

// Transaction represents the output transaction format
type Transaction struct {
	Symbol         string
	ExecutedAt     string
	Quantity       float64
	PricePerShare  float64
	Type           string
	Brokerage      float64
	Notes          string
	PriceCurrency  string
	BrokerCurrency string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: csvtxn <input.csv>\n")
		fmt.Fprintf(os.Stderr, "Output will be written to transactions.csv\n")
		os.Exit(exitError)
	}

	inputFile := os.Args[1]
	outputFile := "transactions.csv"

	if err := convertCSV(inputFile, outputFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitError)
	}

	fmt.Printf("✓ Successfully converted %s to %s\n", inputFile, outputFile)
	os.Exit(exitSuccess)
}

func convertCSV(inputFile, outputFile string) error {
	// Open input file
	inFile, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inFile.Close()

	// Read input CSV
	reader := csv.NewReader(inFile)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("input file is empty")
	}

	// Parse transactions
	var transactions []Transaction
	var errors []string

	for i, record := range records {
		// Skip empty rows
		if len(record) == 0 || (len(record) == 1 && strings.TrimSpace(record[0]) == "") {
			continue
		}

		// Ensure we have enough fields
		if len(record) < 8 {
			errors = append(errors, fmt.Sprintf("Row %d: insufficient fields (expected 13, got %d)", i+1, len(record)))
			continue
		}

		brokerTx := BrokerTransaction{
			Symbol:   strings.TrimSpace(record[0]),
			Exchange: strings.TrimSpace(record[1]),
			Name:     strings.TrimSpace(record[2]),
			Date:     strings.TrimSpace(record[3]),
			Action:   strings.TrimSpace(record[4]),
			Quantity: strings.TrimSpace(record[5]),
			Price:    strings.TrimSpace(record[6]),
			Currency: strings.TrimSpace(record[7]),
		}

		if len(record) > 8 {
			brokerTx.Fee = strings.TrimSpace(record[9])
			brokerTx.FeeCurrency = strings.TrimSpace(record[10])
			brokerTx.FXRate = strings.TrimSpace(record[11])
			brokerTx.Total = strings.TrimSpace(record[12])
		}

		tx, err := convertTransaction(brokerTx)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Row %d (%s): %v", i+1, brokerTx.Symbol, err))
			continue
		}

		transactions = append(transactions, tx)
	}

	// Report errors if any
	if len(errors) > 0 {
		fmt.Fprintf(os.Stderr, "Warnings during conversion:\n")
		for _, err := range errors {
			fmt.Fprintf(os.Stderr, "  - %s\n", err)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	if len(transactions) == 0 {
		return fmt.Errorf("no valid transactions found")
	}

	// Write output CSV
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	// Write header
	header := []string{"symbol", "executed_at", "quantity", "price_per_share", "type", "brokerage", "notes", "price_currency", "brokerage_currency"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write transactions
	for _, tx := range transactions {
		record := []string{
			tx.Symbol,
			tx.ExecutedAt,
			formatFloat(tx.Quantity),
			formatFloat(tx.PricePerShare),
			tx.Type,
			formatFloat(tx.Brokerage),
			tx.Notes,
			tx.PriceCurrency,
			tx.BrokerCurrency,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write transaction: %w", err)
		}
	}

	if err := writer.Error(); err != nil {
		return fmt.Errorf("CSV writer error: %w", err)
	}

	fmt.Printf("Converted %d transactions\n", len(transactions))
	if len(errors) > 0 {
		fmt.Printf("Skipped %d rows due to errors\n", len(errors))
	}

	return nil
}

func convertTransaction(bt BrokerTransaction) (Transaction, error) {
	var tx Transaction

	// Symbol
	if bt.Symbol == "" {
		return tx, fmt.Errorf("symbol is empty")
	}
	tx.Symbol = strings.ToUpper(bt.Symbol)

	// Date - parse and convert to YYYY-MM-DD format
	executedAt, err := parseDate(bt.Date)
	if err != nil {
		return tx, fmt.Errorf("invalid date: %w", err)
	}
	tx.ExecutedAt = executedAt.Format("2006-01-02")

	// Quantity - remove commas and parse
	quantity, err := parseNumber(bt.Quantity)
	if err != nil {
		return tx, fmt.Errorf("invalid quantity: %w", err)
	}
	if quantity == 0 {
		return tx, fmt.Errorf("quantity cannot be zero")
	}

	tx.Notes = "System Upload"

	// Handle negative quantities (SELL transactions)
	isNegative := quantity < 0
	absQuantity := quantity
	if isNegative {
		absQuantity = -quantity
	}
	tx.Quantity = absQuantity

	// Price - remove commas and parse
	price, err := parseNumber(bt.Price)
	if err != nil && bt.Action != "Split" {
		fmt.Printf("Invalid price: %s\n", bt.Action)
		return tx, fmt.Errorf("invalid price: %w", err)
	}
	if price <= 0 && bt.Action != "Split" {
		return tx, fmt.Errorf("price must be positive")
	}
	tx.PricePerShare = price

	// Type - determine from action or quantity sign
	action := strings.ToUpper(bt.Action)

	// If quantity is negative, it's a SELL regardless of action field
	if isNegative {
		tx.Type = "SELL"
	} else if action == "BUY" || action == "B" {
		tx.Type = "BUY"
	} else if action == "SELL" || action == "S" {
		tx.Type = "SELL"
	} else if action == "SPLIT" {
		tx.Type = "SPLIT"
		fmt.Printf("SPLIT: %s\n", bt.Symbol)
	} else {
		// Default to BUY for positive quantities if action is unclear
		tx.Type = "BUY"
	}

	if tx.Type != "SPLIT" {
		// Fee
		fee, err := parseNumber(bt.Fee)
		if err != nil {
			return tx, fmt.Errorf("invalid fee: %w", err)
		}
		tx.Brokerage = fee
		tx.BrokerCurrency = bt.FeeCurrency
	}

	tx.PriceCurrency = bt.Currency

	return tx, nil
}

// parseDate tries multiple date formats
func parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"01/02/2006",
		"02/01/2006",
		"01-02-2006",
		"02-01-2006",
		"2006/01/02",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// parseNumber removes commas and parses float
func parseNumber(numStr string) (float64, error) {
	// Remove commas, quotes, and spaces
	cleaned := strings.ReplaceAll(numStr, ",", "")
	cleaned = strings.ReplaceAll(cleaned, "\"", "")
	cleaned = strings.TrimSpace(cleaned)

	if cleaned == "" {
		return 0, fmt.Errorf("empty number")
	}

	return strconv.ParseFloat(cleaned, 64)
}

// formatFloat formats a float with appropriate precision
func formatFloat(f float64) string {
	// Remove trailing zeros
	s := strconv.FormatFloat(f, 'f', -1, 64)
	return s
}
