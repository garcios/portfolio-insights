// Package entity defines the domain entities for the gateway.
package entity

import "time"

// TransactionType represents the type of transaction
type TransactionType string

const (
	// TransactionTypeBuy represents a buy transaction.
	TransactionTypeBuy TransactionType = "BUY"
	// TransactionTypeSell represents a sell transaction.
	TransactionTypeSell TransactionType = "SELL"
	// TransactionTypeSplit represents a split transaction.
	TransactionTypeSplit TransactionType = "SPLIT"
	// TransactionTypeDividend represents a dividend transaction.
	TransactionTypeDividend TransactionType = "DIVIDEND"
)

// Transaction represents a financial transaction
type Transaction struct {
	ID                string
	UserID            string
	Symbol            string
	Type              TransactionType
	Quantity          float64
	PricePerShare     float64
	PriceCurrency     string
	ExecutedAt        time.Time
	Notes             string
	Brokerage         float64
	BrokerageCurrency string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// NewTransaction creates a new Transaction entity
func NewTransaction(
	id, userID, symbol string,
	txType TransactionType,
	quantity, pricePerShare float64,
	priceCurrency string,
	executedAt time.Time,
) *Transaction {
	now := time.Now()
	return &Transaction{
		ID:            id,
		UserID:        userID,
		Symbol:        symbol,
		Type:          txType,
		Quantity:      quantity,
		PricePerShare: pricePerShare,
		PriceCurrency: priceCurrency,
		ExecutedAt:    executedAt,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// WithNotes adds notes to the transaction
func (t *Transaction) WithNotes(notes string) *Transaction {
	t.Notes = notes
	return t
}

// WithBrokerage adds brokerage fee to the transaction
func (t *Transaction) WithBrokerage(brokerage float64, currency string) *Transaction {
	t.Brokerage = brokerage
	t.BrokerageCurrency = currency
	return t
}
