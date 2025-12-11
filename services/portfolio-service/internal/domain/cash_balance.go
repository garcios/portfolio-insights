package domain

import "time"

// CashBalance represents a user's cash balance in a specific currency
type CashBalance struct {
	UserID    string
	Currency  string
	Balance   float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CashBalanceRepository defines operations for cash balances
type CashBalanceRepository interface {
	// Upsert creates or updates a cash balance
	Upsert(balance *CashBalance) error

	// GetByUserAndCurrency retrieves a cash balance for a specific currency
	GetByUserAndCurrency(userID, currency string) (*CashBalance, error)

	// ListByUser retrieves all cash balances for a user
	ListByUser(userID string) ([]*CashBalance, error)

	// AddAmount adds (or subtracts if negative) to a cash balance
	// This is the primary method for updating cash from transactions
	AddAmount(userID, currency string, amount float64) error
}
