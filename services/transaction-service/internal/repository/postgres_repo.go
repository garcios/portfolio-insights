// Package repository implements the persistence layer for the transaction service.
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/garcios/portfolio-insights/pkg/database"
	"github.com/garcios/portfolio-insights/services/transaction-service/internal/domain"
)

type postgresTransactionRepo struct {
	db *sql.DB
}

// NewPostgresTransactionRepository creates a new PostgreSQL transaction repository.
func NewPostgresTransactionRepository(db *sql.DB) domain.TransactionRepository {
	return &postgresTransactionRepo{db: db}
}

func (r *postgresTransactionRepo) Create(ctx context.Context, transaction *domain.Transaction) error {
	start := time.Now()
	query := `
		INSERT INTO txn.transactions (user_id, symbol, type, quantity, price_per_share, executed_at, brokerage, notes, price_currency, brokerage_currency)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		transaction.UserID,
		transaction.Symbol,
		transaction.Type,
		transaction.Quantity,
		transaction.PricePerShare,
		transaction.ExecutedAt,
		transaction.Brokerage,
		transaction.Notes,
		transaction.PriceCurrency,
		transaction.BrokerageCurrency,
	).Scan(&transaction.ID, &transaction.CreatedAt, &transaction.UpdatedAt)

	database.RecordQuery("create", "transactions", time.Since(start).Seconds(), err)
	if err == nil {
		database.RecordRowsAffected("create", "transactions", 1)
	}
	return err
}

func (r *postgresTransactionRepo) BulkCreate(ctx context.Context, transactions []*domain.Transaction) error {
	start := time.Now()
	if len(transactions) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			// Just log or ignore if already committed
			fmt.Printf("failed to rollback transaction: %v\n", err)
		}
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO txn.transactions (user_id, symbol, type, quantity, price_per_share, executed_at, created_at, updated_at, brokerage, notes, price_currency, brokerage_currency)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`)
	if err != nil {
		return err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			fmt.Printf("failed to close statement: %v\n", err)
		}
	}()

	for _, txn := range transactions {
		_, err = stmt.ExecContext(ctx,
			txn.UserID,
			txn.Symbol,
			txn.Type,
			txn.Quantity,
			txn.PricePerShare,
			txn.ExecutedAt,
			txn.CreatedAt,
			txn.UpdatedAt,
			txn.Brokerage,
			txn.Notes,
			txn.PriceCurrency,
			txn.BrokerageCurrency,
		)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	database.RecordQuery("bulk_create", "transactions", time.Since(start).Seconds(), err)
	if err == nil {
		database.RecordRowsAffected("bulk_create", "transactions", int64(len(transactions)))
	}
	return err
}

func (r *postgresTransactionRepo) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	start := time.Now()

	query := `
		SELECT id, user_id, symbol, type, quantity, price_per_share, executed_at, created_at, updated_at, brokerage, notes, price_currency, brokerage_currency
		FROM txn.transactions
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var txn domain.Transaction
	var brokerage sql.NullFloat64
	var notes sql.NullString
	var priceCurrency sql.NullString
	var brokerageCurrency sql.NullString

	err := row.Scan(
		&txn.ID,
		&txn.UserID,
		&txn.Symbol,
		&txn.Type,
		&txn.Quantity,
		&txn.PricePerShare,
		&txn.ExecutedAt,
		&txn.CreatedAt,
		&txn.UpdatedAt,
		&brokerage,
		&notes,
		&priceCurrency,
		&brokerageCurrency,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if brokerage.Valid {
		txn.Brokerage = brokerage.Float64
	}
	if notes.Valid {
		txn.Notes = notes.String
	}
	if priceCurrency.Valid {
		txn.PriceCurrency = priceCurrency.String
	}
	if brokerageCurrency.Valid {
		txn.BrokerageCurrency = brokerageCurrency.String
	}

	database.RecordQuery("get_by_id", "transactions", time.Since(start).Seconds(), nil)
	return &txn, nil
}

func (r *postgresTransactionRepo) ListByUserID(ctx context.Context, userID string, filter domain.TransactionFilter, limit, offset int) ([]*domain.Transaction, error) {
	start := time.Now()

	query := `
		SELECT id, user_id, symbol, type, quantity, price_per_share, executed_at, created_at, updated_at, brokerage, notes, price_currency, brokerage_currency
		FROM txn.transactions
		WHERE user_id = $1
	`
	args := []interface{}{userID}
	argIdx := 2

	if filter.Symbol != "" {
		query += fmt.Sprintf(" AND symbol ILIKE $%d", argIdx)
		args = append(args, "%"+filter.Symbol+"%")
		argIdx++
	}
	if filter.Type != "" {
		query += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, filter.Type)
		argIdx++
	}
	if !filter.FromExecutedAt.IsZero() {
		query += fmt.Sprintf(" AND executed_at >= $%d", argIdx)
		args = append(args, filter.FromExecutedAt)
		argIdx++
	}
	if !filter.ToExecutedAt.IsZero() {
		query += fmt.Sprintf(" AND executed_at <= $%d", argIdx)
		args = append(args, filter.ToExecutedAt)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY executed_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		database.RecordQuery("list_by_user_id", "transactions", time.Since(start).Seconds(), err)
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("failed to close rows: %v\n", err)
			database.RecordQuery("list_by_user_id_close", "transactions", 0, err)
		}
	}()

	var transactions []*domain.Transaction
	for rows.Next() {
		var txn domain.Transaction
		var brokerage sql.NullFloat64
		var notes sql.NullString
		var priceCurrency sql.NullString
		var brokerageCurrency sql.NullString

		if err := rows.Scan(
			&txn.ID,
			&txn.UserID,
			&txn.Symbol,
			&txn.Type,
			&txn.Quantity,
			&txn.PricePerShare,
			&txn.ExecutedAt,
			&txn.CreatedAt,
			&txn.UpdatedAt,
			&brokerage,
			&notes,
			&priceCurrency,
			&brokerageCurrency,
		); err != nil {
			return nil, err
		}

		if brokerage.Valid {
			txn.Brokerage = brokerage.Float64
		}
		if notes.Valid {
			txn.Notes = notes.String
		}
		if priceCurrency.Valid {
			txn.PriceCurrency = priceCurrency.String
		}
		if brokerageCurrency.Valid {
			txn.BrokerageCurrency = brokerageCurrency.String
		}

		transactions = append(transactions, &txn)
	}
	database.RecordQuery("list_by_user_id", "transactions", time.Since(start).Seconds(), nil)
	return transactions, nil
}

func (r *postgresTransactionRepo) Update(ctx context.Context, txn *domain.Transaction) error {
	start := time.Now()
	query := `
		UPDATE txn.transactions
		SET symbol = $1, type = $2, quantity = $3, price_per_share = $4, executed_at = $5, updated_at = $6, brokerage = $7, notes = $8, price_currency = $9, brokerage_currency = $10
		WHERE id = $11
	`
	result, err := r.db.ExecContext(ctx, query,
		txn.Symbol,
		txn.Type,
		txn.Quantity,
		txn.PricePerShare,
		txn.ExecutedAt,
		txn.UpdatedAt,
		txn.Brokerage,
		txn.Notes,
		txn.PriceCurrency,
		txn.BrokerageCurrency,
		txn.ID,
	)
	database.RecordQuery("update", "transactions", time.Since(start).Seconds(), err)
	if err == nil {
		rowsAffected, _ := result.RowsAffected()
		database.RecordRowsAffected("update", "transactions", rowsAffected)
	}
	return err
}

func (r *postgresTransactionRepo) Delete(ctx context.Context, id string) error {
	start := time.Now()
	query := `DELETE FROM txn.transactions WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		database.RecordQuery("delete", "transactions", time.Since(start).Seconds(), err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		err = sql.ErrNoRows
		database.RecordQuery("delete", "transactions", time.Since(start).Seconds(), err)
		return err
	}
	database.RecordQuery("delete", "transactions", time.Since(start).Seconds(), nil)
	database.RecordRowsAffected("delete", "transactions", rowsAffected)
	return nil
}

func (r *postgresTransactionRepo) Count() (int, error) {
	start := time.Now()
	var count int
	query := `SELECT COUNT(*) FROM txn.transactions`
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		database.RecordQuery("count", "transactions", time.Since(start).Seconds(), err)
		return 0, err
	}
	database.RecordQuery("count", "transactions", time.Since(start).Seconds(), nil)
	return count, nil
}

func (r *postgresTransactionRepo) GetOldestByUserID(ctx context.Context, userID string) (*domain.Transaction, error) {
	start := time.Now()

	query := `
		SELECT id, user_id, symbol, type, quantity, price_per_share, executed_at, created_at, updated_at, brokerage, notes, price_currency, brokerage_currency
		FROM txn.transactions
		WHERE user_id = $1
		ORDER BY executed_at ASC
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, userID)

	var txn domain.Transaction
	var brokerage sql.NullFloat64
	var notes sql.NullString
	var priceCurrency sql.NullString
	var brokerageCurrency sql.NullString

	err := row.Scan(
		&txn.ID,
		&txn.UserID,
		&txn.Symbol,
		&txn.Type,
		&txn.Quantity,
		&txn.PricePerShare,
		&txn.ExecutedAt,
		&txn.CreatedAt,
		&txn.UpdatedAt,
		&brokerage,
		&notes,
		&priceCurrency,
		&brokerageCurrency,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if brokerage.Valid {
		txn.Brokerage = brokerage.Float64
	}
	if notes.Valid {
		txn.Notes = notes.String
	}
	if priceCurrency.Valid {
		txn.PriceCurrency = priceCurrency.String
	}
	if brokerageCurrency.Valid {
		txn.BrokerageCurrency = brokerageCurrency.String
	}

	database.RecordQuery("get_oldest_by_user_id", "transactions", time.Since(start).Seconds(), nil)
	return &txn, nil
}
