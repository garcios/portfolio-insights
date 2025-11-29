package repository

import (
	"context"
	"database/sql"

	"github.com/garcios/portfolio-insights/services/transaction-service/internal/domain"
)

type postgresTransactionRepo struct {
	db *sql.DB
}

func NewPostgresTransactionRepository(db *sql.DB) domain.TransactionRepository {
	return &postgresTransactionRepo{db: db}
}

func (r *postgresTransactionRepo) Create(ctx context.Context, transaction *domain.Transaction) error {
	query := `
		INSERT INTO txn.transactions (user_id, symbol, type, quantity, price_per_share, executed_at, brokerage, notes, price_currency, brokerage_currency)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
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
}

func (r *postgresTransactionRepo) BulkCreate(ctx context.Context, transactions []*domain.Transaction) error {
	if len(transactions) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO txn.transactions (user_id, symbol, type, quantity, price_per_share, executed_at, created_at, updated_at, brokerage, notes, price_currency, brokerage_currency)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

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

	return tx.Commit()
}

func (r *postgresTransactionRepo) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
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

	return &txn, nil
}

func (r *postgresTransactionRepo) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Transaction, error) {
	query := `
		SELECT id, user_id, symbol, type, quantity, price_per_share, executed_at, created_at, updated_at, brokerage, notes, price_currency, brokerage_currency
		FROM txn.transactions
		WHERE user_id = $1
		ORDER BY executed_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
	return transactions, nil
}

func (r *postgresTransactionRepo) Update(ctx context.Context, txn *domain.Transaction) error {
	query := `
		UPDATE txn.transactions
		SET symbol = $1, type = $2, quantity = $3, price_per_share = $4, executed_at = $5, updated_at = $6, brokerage = $7, notes = $8, price_currency = $9, brokerage_currency = $10
		WHERE id = $11
	`
	_, err := r.db.ExecContext(ctx, query,
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
	return err
}

func (r *postgresTransactionRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM txn.transactions WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
