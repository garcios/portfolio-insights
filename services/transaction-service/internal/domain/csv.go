package domain

import "time"

// CSVTransaction represents a transaction row from CSV
type CSVTransaction struct {
	Symbol        string
	ExecutedAt    time.Time
	Quantity      float64
	PricePerShare float64
	Type          string
}

// CSVUploadResult represents the result of a CSV upload operation
type CSVUploadResult struct {
	TotalRecords      int
	SuccessfulRecords int
	FailedRecords     int
	Errors            []CSVRowError
}

// CSVRowError represents an error for a specific row
type CSVRowError struct {
	RowNumber int
	Row       map[string]string
	Error     string
}

// CSVUploadUsecase defines the interface for CSV upload operations
type CSVUploadUsecase interface {
	UploadCSV(userID string, csvData []byte) (*CSVUploadResult, error)
}
