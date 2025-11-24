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
		INSERT INTO txn.transactions (user_id, symbol, type, quantity, price_per_share, executed_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		transaction.UserID,
		transaction.Symbol,
		transaction.Type,
		transaction.Quantity,
		transaction.PricePerShare,
		transaction.ExecutedAt,
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
		INSERT INTO txn.transactions (user_id, symbol, type, quantity, price_per_share, executed_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, transaction := range transactions {
		err = stmt.QueryRowContext(ctx,
			transaction.UserID,
			transaction.Symbol,
			transaction.Type,
			transaction.Quantity,
			transaction.PricePerShare,
			transaction.ExecutedAt,
		).Scan(&transaction.ID, &transaction.CreatedAt, &transaction.UpdatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *postgresTransactionRepo) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	query := `
		SELECT id, user_id, symbol, type, quantity, price_per_share, executed_at, created_at, updated_at
		FROM txn.transactions
		WHERE id = $1
	`
	var t domain.Transaction
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.UserID, &t.Symbol, &t.Type, &t.Quantity, &t.PricePerShare, &t.ExecutedAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *postgresTransactionRepo) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Transaction, error) {
	query := `
		SELECT id, user_id, symbol, type, quantity, price_per_share, executed_at, created_at, updated_at
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
		var t domain.Transaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.Symbol, &t.Type, &t.Quantity, &t.PricePerShare, &t.ExecutedAt, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, &t)
	}
	return transactions, nil
}

func (r *postgresTransactionRepo) Update(ctx context.Context, transaction *domain.Transaction) error {
	query := `
		UPDATE txn.transactions
		SET symbol = $1, type = $2, quantity = $3, price_per_share = $4, executed_at = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		transaction.Symbol,
		transaction.Type,
		transaction.Quantity,
		transaction.PricePerShare,
		transaction.ExecutedAt,
		transaction.ID,
	).Scan(&transaction.UpdatedAt)
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
