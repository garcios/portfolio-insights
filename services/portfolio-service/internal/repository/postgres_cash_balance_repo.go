package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
)

// PostgresCashBalanceRepository implements domain.CashBalanceRepository using PostgreSQL
type PostgresCashBalanceRepository struct {
	db *sql.DB
}

// NewPostgresCashBalanceRepository creates a new PostgreSQL cash balance repository
func NewPostgresCashBalanceRepository(db *sql.DB) *PostgresCashBalanceRepository {
	return &PostgresCashBalanceRepository{db: db}
}

// Upsert creates or updates a cash balance
func (r *PostgresCashBalanceRepository) Upsert(balance *domain.CashBalance) error {
	query := `
		INSERT INTO investments.cash_balances (user_id, currency, balance, notes, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, currency) 
		DO UPDATE SET balance = $3, notes = $4, updated_at = $5
	`
	_, err := r.db.Exec(query, balance.UserID, balance.Currency, balance.Balance, balance.Notes, time.Now())

	if err != nil {
		return fmt.Errorf("failed to upsert cash balance: %w", err)
	}

	return nil
}

// GetByUserAndCurrency retrieves a cash balance for a specific currency
func (r *PostgresCashBalanceRepository) GetByUserAndCurrency(userID, currency string) (*domain.CashBalance, error) {
	query := `
		SELECT user_id, currency, balance, notes, created_at, updated_at
		FROM investments.cash_balances
		WHERE user_id = $1 AND currency = $2
	`
	balance := &domain.CashBalance{}
	err := r.db.QueryRow(query, userID, currency).Scan(
		&balance.UserID,
		&balance.Currency,
		&balance.Balance,
		&balance.Notes,
		&balance.CreatedAt,
		&balance.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("cash balance not found for user %s currency %s", userID, currency)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cash balance: %w", err)
	}

	return balance, nil
}

// ListByUser retrieves all cash balances for a user
func (r *PostgresCashBalanceRepository) ListByUser(userID string) ([]*domain.CashBalance, error) {
	query := `
		SELECT user_id, currency, balance, notes, created_at, updated_at
		FROM investments.cash_balances
		WHERE user_id = $1
		ORDER BY currency
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list cash balances: %w", err)
	}
	defer func() {
		_ = rows.Close() // Ignore close error in defer
	}()

	var balances []*domain.CashBalance
	for rows.Next() {
		balance := &domain.CashBalance{}
		if err := rows.Scan(
			&balance.UserID,
			&balance.Currency,
			&balance.Balance,
			&balance.Notes,
			&balance.CreatedAt,
			&balance.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan cash balance: %w", err)
		}
		balances = append(balances, balance)
	}

	return balances, nil
}

// AddAmount adds (or subtracts if negative) to a cash balance
// This is the primary method for updating cash from transactions
// Notes parameter allows setting notes on new balances; existing notes are preserved on updates
func (r *PostgresCashBalanceRepository) AddAmount(userID, currency string, amount float64, notes string) error {
	query := `
		INSERT INTO investments.cash_balances (user_id, currency, balance, notes, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id, currency)
		DO UPDATE SET 
			balance = investments.cash_balances.balance + $3,
			notes = CASE 
				WHEN $4 != '' THEN $4 
				ELSE investments.cash_balances.notes 
			END,
			updated_at = NOW()
	`
	_, err := r.db.Exec(query, userID, currency, amount, notes)

	if err != nil {
		return fmt.Errorf("failed to add amount to cash balance: %w", err)
	}

	return nil
}
