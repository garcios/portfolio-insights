package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

// InputRecord represents a row in the input CSV
type InputRecord struct {
	Date        string
	Type        string
	Description string
	Debit       string
	Credit      string
	Balance     string
}

// OutputRecord represents a row in the target CSV
type OutputRecord struct {
	Symbol            string
	ExecutedAt        string
	Quantity          string
	PricePerShare     string
	Type              string
	Brokerage         string
	Notes             string
	PriceCurrency     string
	BrokerageCurrency string
}

func main() {
	reader := csv.NewReader(os.Stdin)
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write Header
	header := []string{"symbol", "executed_at", "quantity", "price_per_share", "type", "brokerage", "notes", "price_currency", "brokerage_currency"}
	if err := writer.Write(header); err != nil {
		log.Fatalf("Failed to write header: %v", err)
	}

	// Read Header from input
	_, err := reader.Read()
	if err != nil {
		log.Fatalf("Failed to read input header: %v", err)
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to read record: %v", err)
		}

		// Parse Input
		in := InputRecord{
			Date:        record[0],
			Type:        record[1],
			Description: record[2],
			Debit:       record[3],
			Credit:      record[4],
			Balance:     record[5],
		}

		out := transform(in)
		if out == nil {
			continue
		}

		// Write Output
		row := []string{
			out.Symbol,
			out.ExecutedAt,
			out.Quantity,
			out.PricePerShare,
			out.Type,
			out.Brokerage,
			out.Notes,
			out.PriceCurrency,
			out.BrokerageCurrency,
		}
		if err := writer.Write(row); err != nil {
			log.Fatalf("Failed to write record: %v", err)
		}
	}
}

func transform(in InputRecord) *OutputRecord {
	out := &OutputRecord{
		ExecutedAt:        in.Date,
		Notes:             in.Description,
		BrokerageCurrency: "AUD", // Default
		PriceCurrency:     "AUD", // Default
	}

	typeUpper := strings.ToUpper(in.Type)
	descUpper := strings.ToUpper(in.Description)

	// Exit early for types we don't process
	if typeUpper == "INTERESTCHANGE" {
		return nil
	}

	// Determine the output type
	switch {
	case typeUpper == "INTEREST":
		out.Type = "INT"
	case strings.Contains(descUpper, "DIVIDEND"):
		out.Type = "DIV"
	case strings.Contains(descUpper, "BUY"):
		out.Type = "WTH"
	case strings.Contains(descUpper, "DEPOSIT") || strings.Contains(descUpper, "SELL"):
		out.Type = "DEP"
	default:
		return nil // No valid type found
	}

	// Rules for Brokerage
	if out.Type == "DIV" || out.Type == "INT" || out.Type == "DEP" {
		out.Brokerage = "0"
	} else {
		out.Brokerage = "9.95"
	}
	// Correction for DEP/INT based on user request "Default to 0 for Dividends... and 9.95 for all other"
	// Wait, user said: "Default to '0' for Dividends and '9.95' for all other transactions unless specified."
	// And example output showed DEP and INT having 9.95.
	// Re-reading example:
	// AUD,2023-09-29,...,DEP,9.95,...
	// AUD,2025-11-28,...,INT,9.95,...
	// So my previous assumption in plan (0 for DEP/INT) was wrong based on the example. The example has 9.95 for DEP/INT.
	// I will stick to the rule "9.95 for all other".
	if out.Type == "DIV" {
		out.Brokerage = "0"
	} else {
		out.Brokerage = "9.95"
	}

	// Extract Symbol, Qty, Price
	switch out.Type {
	case "BUY":
		// Format: "BUY [Ticker] [Qty] [Currency] [Price] ..."
		// Example: "BUY MA.NYS 1 USD 545.03 ..."
		parts := strings.Fields(in.Description)
		if len(parts) >= 2 {
			rawTicker := parts[1]
			out.Symbol = extractTicker(rawTicker)
		}
		if len(parts) >= 3 {
			out.Quantity = parts[2]
		}
		if len(parts) >= 5 {
			// Assuming format "BUY Ticker Qty Curr Price"
			out.PricePerShare = parts[4]
			// Check for US currency
			if parts[3] == "USD" {
				out.PriceCurrency = "USD"
			}
		}

	case "DIV":
		// Format: "DIVIDEND on TSM.US ..."
		re := regexp.MustCompile(`DIVIDEND on (\S+)`)
		match := re.FindStringSubmatch(in.Description)
		if len(match) > 1 {
			out.Symbol = extractTicker(match[1])
		} else {
			out.Symbol = "AUD"
		}
		out.Quantity = "0"
		out.PricePerShare = in.Credit // Dividend amount in PricePerShare?
		// Rule: "For 'DIVIDEND', use the 'Credit' value" for PricePerShare.

		// Currency check
		if strings.Contains(descUpper, "USD") || strings.Contains(descUpper, ".US") || strings.Contains(descUpper, ".NYS") || strings.Contains(descUpper, ".NAS") {
			// Example: "DIVIDEND on TSM.US ... USD to AUD"
			// If the dividends are paid in USD but converted, the original transaction might be in USD context.
			// But the Credit value `35.53` in the example `2026-01-20,Credit,DIVIDEND on TSM.US ... 35.53`
			// The description says "USD to AUD @ 1.4879". So the 35.53 is likely AUD.
			// However, rule says: "price_currency ... Default to 'AUD', but use 'USD' ... if the description indicates a US stock transaction."
			// Example output for DIV TSM: PriceCurrency = USD.
			out.PriceCurrency = "USD"
		}

	case "DEP", "INT":
		out.Symbol = "AUD"
		// Rule: "For 'FUNDS TRANSFER', or 'INTEREST', use ... the currency amount if it represents a deposit quantity."
		// And example output shows quantity populated for DEP/INT.
		if in.Credit != "" {
			out.Quantity = in.Credit
		} else {
			out.Quantity = "0"
		}
		out.PricePerShare = "0"

	default:
		out.Symbol = "AUD"
		out.Quantity = "0"
		out.PricePerShare = "0"
	}

	return out
}

func extractTicker(input string) string {
	// Removes suffix like .US, .NYS, .AX
	parts := strings.Split(input, ".")
	return parts[0]
}
