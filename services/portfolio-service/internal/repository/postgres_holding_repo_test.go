package repository

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
)

func TestUpsert_Insert(t *testing.T) {
	// Setup
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresHoldingRepository(db)

	holding := &domain.Holding{
		UserID:      "user-123",
		Symbol:      "AAPL",
		Quantity:    10,
		AverageCost: 150.00,
		LastUpdated: time.Now(),
	}

	// Expect INSERT query
	mock.ExpectExec("INSERT INTO investments.holdings").
		WithArgs(holding.UserID, holding.Symbol, holding.Quantity, holding.AverageCost, holding.Currency, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute
	err = repo.Upsert(holding)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestUpsert_Update(t *testing.T) {
	// Setup
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresHoldingRepository(db)

	holding := &domain.Holding{
		UserID:      "user-123",
		Symbol:      "AAPL",
		Quantity:    20,
		AverageCost: 160.00,
		LastUpdated: time.Now(),
	}

	// Expect UPSERT query (ON CONFLICT DO UPDATE)
	mock.ExpectExec("INSERT INTO investments.holdings").
		WithArgs(holding.UserID, holding.Symbol, holding.Quantity, holding.AverageCost, holding.Currency, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute
	err = repo.Upsert(holding)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestUpsert_Error(t *testing.T) {
	// Setup
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresHoldingRepository(db)

	holding := &domain.Holding{
		UserID:      "user-123",
		Symbol:      "AAPL",
		Quantity:    10,
		AverageCost: 150.00,
		LastUpdated: time.Now(),
	}

	// Expect query to fail
	mock.ExpectExec("INSERT INTO investments.holdings").
		WithArgs(holding.UserID, holding.Symbol, holding.Quantity, holding.AverageCost, holding.Currency, sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	// Execute
	err = repo.Upsert(holding)

	// Assert
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestGetByUserAndSymbol_Success(t *testing.T) {
	// Setup
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresHoldingRepository(db)

	userID := "user-123"
	symbol := "AAPL"
	expectedTime := time.Now()

	// Expect SELECT query
	rows := sqlmock.NewRows([]string{"user_id", "symbol", "quantity", "average_cost_basis", "currency", "updated_at"}).
		AddRow(userID, symbol, 10.0, 150.0, "USD", expectedTime)

	mock.ExpectQuery("SELECT user_id, symbol, quantity, average_cost_basis, currency, updated_at FROM investments.holdings").
		WithArgs(userID, symbol).
		WillReturnRows(rows)

	// Execute
	holding, err := repo.GetByUserAndSymbol(userID, symbol)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if holding.UserID != userID {
		t.Errorf("Expected user_id %s, got %s", userID, holding.UserID)
	}

	if holding.Symbol != symbol {
		t.Errorf("Expected symbol %s, got %s", symbol, holding.Symbol)
	}

	if holding.Quantity != 10.0 {
		t.Errorf("Expected quantity 10.0, got %f", holding.Quantity)
	}

	if holding.AverageCost != 150.0 {
		t.Errorf("Expected average cost 150.0, got %f", holding.AverageCost)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestGetByUserAndSymbol_NotFound(t *testing.T) {
	// Setup
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresHoldingRepository(db)

	userID := "user-123"
	symbol := "AAPL"

	// Expect SELECT query to return no rows
	mock.ExpectQuery("SELECT user_id, symbol, quantity, average_cost_basis, currency, updated_at FROM investments.holdings").
		WithArgs(userID, symbol).
		WillReturnError(sql.ErrNoRows)

	// Execute
	_, err = repo.GetByUserAndSymbol(userID, symbol)

	// Assert
	if err == nil {
		t.Error("Expected error for not found, got nil")
	}

	if err.Error() != "holding not found" {
		t.Errorf("Expected 'holding not found', got '%s'", err.Error())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestListByUser_Success(t *testing.T) {
	// Setup
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresHoldingRepository(db)

	userID := "user-123"
	now := time.Now()

	// Expect SELECT query
	rows := sqlmock.NewRows([]string{"user_id", "symbol", "quantity", "average_cost_basis", "currency", "updated_at"}).
		AddRow(userID, "AAPL", 10.0, 150.0, "USD", now).
		AddRow(userID, "GOOGL", 5.0, 2800.0, "USD", now)

	mock.ExpectQuery("SELECT user_id, symbol, quantity, average_cost_basis, currency, updated_at FROM investments.holdings").
		WithArgs(userID).
		WillReturnRows(rows)

	// Execute
	holdings, err := repo.ListByUser(userID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(holdings) != 2 {
		t.Fatalf("Expected 2 holdings, got %d", len(holdings))
	}

	// Verify first holding
	if holdings[0].Symbol != "AAPL" {
		t.Errorf("Expected first symbol AAPL, got %s", holdings[0].Symbol)
	}

	if holdings[0].Quantity != 10.0 {
		t.Errorf("Expected first quantity 10.0, got %f", holdings[0].Quantity)
	}

	// Verify second holding
	if holdings[1].Symbol != "GOOGL" {
		t.Errorf("Expected second symbol GOOGL, got %s", holdings[1].Symbol)
	}

	if holdings[1].Quantity != 5.0 {
		t.Errorf("Expected second quantity 5.0, got %f", holdings[1].Quantity)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestListByUser_Empty(t *testing.T) {
	// Setup
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresHoldingRepository(db)

	userID := "user-456"

	// Expect SELECT query to return no rows
	rows := sqlmock.NewRows([]string{"user_id", "symbol", "quantity", "average_cost_basis", "currency", "updated_at"})

	mock.ExpectQuery("SELECT user_id, symbol, quantity, average_cost_basis, currency, updated_at FROM investments.holdings").
		WithArgs(userID).
		WillReturnRows(rows)

	// Execute
	holdings, err := repo.ListByUser(userID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(holdings) != 0 {
		t.Errorf("Expected 0 holdings, got %d", len(holdings))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestListByUser_QueryError(t *testing.T) {
	// Setup
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresHoldingRepository(db)

	userID := "user-123"

	// Expect query to fail
	mock.ExpectQuery("SELECT user_id, symbol, quantity, average_cost_basis, currency, updated_at FROM investments.holdings").
		WithArgs(userID).
		WillReturnError(sql.ErrConnDone)

	// Execute
	_, err = repo.ListByUser(userID)

	// Assert
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestCount_Success(t *testing.T) {
	// Setup
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresHoldingRepository(db)

	// Expect COUNT query
	rows := sqlmock.NewRows([]string{"count"}).AddRow(42)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM investments.holdings").
		WillReturnRows(rows)

	// Execute
	count, err := repo.Count()

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if count != 42 {
		t.Errorf("Expected count 42, got %d", count)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestDeleteZeroQuantityHoldings_Success(t *testing.T) {
	// Setup
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresHoldingRepository(db)

	// Expect DELETE query
	mock.ExpectExec("DELETE FROM investments.holdings WHERE quantity <= 0").
		WillReturnResult(sqlmock.NewResult(0, 3))

	// Execute
	err = repo.DeleteZeroQuantityHoldings()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}
