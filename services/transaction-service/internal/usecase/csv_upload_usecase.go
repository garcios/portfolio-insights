package usecase

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/garcios/portfolio-insights/services/transaction-service/internal/domain"
)

type csvUploadUsecase struct {
	repo           domain.TransactionRepository
	userGateway    domain.UserGateway
	marketGateway  domain.MarketDataGateway
	eventPublisher domain.EventPublisher
}

func NewCSVUploadUsecase(
	repo domain.TransactionRepository,
	userGateway domain.UserGateway,
	marketGateway domain.MarketDataGateway,
	eventPublisher domain.EventPublisher,
) domain.CSVUploadUsecase {
	return &csvUploadUsecase{
		repo:           repo,
		userGateway:    userGateway,
		marketGateway:  marketGateway,
		eventPublisher: eventPublisher,
	}
}

func (uc *csvUploadUsecase) UploadCSV(userID string, csvData []byte) (*domain.CSVUploadResult, error) {
	ctx := context.Background()

	// Validate user exists
	exists, err := uc.userGateway.Exists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate user: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user %s does not exist", userID)
	}

	// Parse CSV
	reader := csv.NewReader(strings.NewReader(string(csvData)))
	reader.TrimLeadingSpace = true

	// Read header
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	// Validate required columns
	requiredColumns := []string{"symbol", "executed_at", "quantity", "price_per_share", "type", "brokerage", "notes", "price_currency", "brokerage_currency"}
	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(header))] = i
	}

	for _, required := range requiredColumns {
		if _, ok := headerMap[required]; !ok {
			return nil, fmt.Errorf("missing required column: %s", required)
		}
	}

	// Parse rows
	var validTransactions []*domain.Transaction
	var errors []domain.CSVRowError
	rowNumber := 1 // Start at 1 (header is row 0)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			errors = append(errors, domain.CSVRowError{
				RowNumber: rowNumber,
				Error:     fmt.Sprintf("failed to parse row: %v", err),
			})
			rowNumber++
			continue
		}

		rowNumber++

		// Create row map for error reporting
		rowMap := make(map[string]string)
		for i, value := range record {
			if i < len(headers) {
				rowMap[headers[i]] = value
			}
		}

		// Parse transaction
		tx, err := uc.parseCSVRow(record, headerMap, userID)
		if err != nil {
			errors = append(errors, domain.CSVRowError{
				RowNumber: rowNumber,
				Row:       rowMap,
				Error:     err.Error(),
			})
			continue
		}

		// Validate symbol exists
		symbolExists, err := uc.marketGateway.Exists(ctx, tx.Symbol)
		if err != nil {
			errors = append(errors, domain.CSVRowError{
				RowNumber: rowNumber,
				Row:       rowMap,
				Error:     fmt.Sprintf("failed to validate symbol: %v", err),
			})
			continue
		}
		if !symbolExists {
			errors = append(errors, domain.CSVRowError{
				RowNumber: rowNumber,
				Row:       rowMap,
				Error:     fmt.Sprintf("symbol %s does not exist", tx.Symbol),
			})
			continue
		}

		validTransactions = append(validTransactions, tx)
	}

	// Bulk insert valid transactions
	if len(validTransactions) > 0 {
		if err := uc.repo.BulkCreate(ctx, validTransactions); err != nil {
			return nil, fmt.Errorf("failed to insert transactions: %w", err)
		}

		// Publish events for each transaction
		for _, tx := range validTransactions {
			if err := uc.eventPublisher.PublishTransactionCreated(ctx, tx); err != nil {
				// Log error but don't fail the upload
				fmt.Printf("failed to publish event for transaction %s: %v\n", tx.ID, err)
			}
		}
	}

	result := &domain.CSVUploadResult{
		TotalRecords:      rowNumber - 1, // Exclude header
		SuccessfulRecords: len(validTransactions),
		FailedRecords:     len(errors),
		Errors:            errors,
	}

	return result, nil
}

func (uc *csvUploadUsecase) parseCSVRow(record []string, headerMap map[string]int, userID string) (*domain.Transaction, error) {
	// Helper function to get value by column name
	getValue := func(colName string) (string, error) {
		idx, ok := headerMap[colName]
		if !ok || idx >= len(record) {
			return "", fmt.Errorf("column %s not found", colName)
		}
		return strings.TrimSpace(record[idx]), nil
	}

	// Parse symbol
	symbol, err := getValue("symbol")
	if err != nil {
		return nil, err
	}
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	// Parse executed_at
	executedAtStr, err := getValue("executed_at")
	if err != nil {
		return nil, err
	}
	executedAt, err := parseDate(executedAtStr)
	if err != nil {
		return nil, fmt.Errorf("invalid executed_at format: %w", err)
	}

	// Parse quantity
	quantityStr, err := getValue("quantity")
	if err != nil {
		return nil, err
	}
	quantity, err := strconv.ParseFloat(quantityStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid quantity: %w", err)
	}
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}

	// Parse type
	txType, err := getValue("type")
	if err != nil {
		return nil, err
	}
	txType = strings.ToUpper(txType)
	if txType != "BUY" && txType != "SELL" && txType != "SPLIT" {
		return nil, fmt.Errorf("type must be BUY, SELL, or SPLIT")
	}

	// Parse price_per_share
	// For SPLIT transactions, price can be 0 or empty
	priceStr, err := getValue("price_per_share")
	if err != nil {
		return nil, err
	}

	var price float64
	if priceStr == "" || priceStr == "0" {
		// Empty or zero price is only allowed for SPLIT transactions
		if txType != "SPLIT" {
			return nil, fmt.Errorf("price_per_share is required for %s transactions", txType)
		}
		price = 0
	} else {
		price, err = strconv.ParseFloat(priceStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid price_per_share: %w", err)
		}
		// Price must be positive for BUY and SELL, but can be 0 for SPLIT
		if price <= 0 && txType != "SPLIT" {
			return nil, fmt.Errorf("price_per_share must be positive for %s transactions", txType)
		}
	}

	// Parse brokerage
	brokerageStr, err := getValue("brokerage")
	if err != nil {
		return nil, err
	}

	var brokerage float64
	if brokerageStr != "" {
		brokerage, err = strconv.ParseFloat(brokerageStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid brokerage: %w", err)
		}
	}

	// Parse notes
	notes, err := getValue("notes")
	if err != nil {
		return nil, err
	}

	// Parse price_currency
	priceCurrency, err := getValue("price_currency")
	if err != nil {
		return nil, err
	}

	// Parse brokerage_currency
	brokerageCurrency, err := getValue("brokerage_currency")
	if err != nil {
		return nil, err
	}

	return &domain.Transaction{
		UserID:            userID,
		Symbol:            strings.ToUpper(symbol),
		Type:              txType,
		Quantity:          quantity,
		PricePerShare:     price,
		ExecutedAt:        executedAt,
		Brokerage:         brokerage,
		Notes:             notes,
		PriceCurrency:     priceCurrency,
		BrokerageCurrency: brokerageCurrency,
	}, nil
}

// parseDate tries multiple date formats
func parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"01/02/2006",
		"01-02-2006",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}
