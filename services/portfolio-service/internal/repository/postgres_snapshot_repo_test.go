package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
)

func TestPostgresSnapshotRepo_GetLatestSnapshot(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func() { _ = db.Close() }()

	repo := NewPostgresSnapshotRepository(db)

	userID := "user-123"
	timestamp := time.Now().Truncate(time.Second) // Truncate for DB precision match

	// Setup Expectations
	rows := sqlmock.NewRows([]string{"id", "user_id", "timestamp", "holdings_snapshot", "cash_snapshot", "realized_gains_snapshot", "net_invested", "transaction_count", "created_at"}).
		AddRow("snap-1", userID, timestamp, []byte(`{"AAPL": {"quantity": "10", "cost_basis": "1000", "currency": "USD"}}`), []byte(`{"USD": "500"}`), []byte(`{"USD": "100"}`), "1000", 5, time.Now())

	mock.ExpectQuery("SELECT id, user_id, timestamp, holdings_snapshot, cash_snapshot, realized_gains_snapshot, net_invested, transaction_count, created_at FROM investments.portfolio_snapshots").
		WithArgs(userID, timestamp).
		WillReturnRows(rows)

	// Execute
	snap, err := repo.GetLatestSnapshot(context.Background(), userID, timestamp)

	// Assert
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if snap == nil {
		t.Fatal("Expected snapshot, got nil")
	}
	if snap.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, snap.UserID)
	}
	if len(snap.State.Holdings) != 1 {
		t.Errorf("Expected 1 holding, got %d", len(snap.State.Holdings))
	}
	if snap.State.NetInvested != "1000" {
		t.Errorf("Expected NetInvested 1000, got %s", snap.State.NetInvested)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPostgresSnapshotRepo_UpsertSnapshot(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func() { _ = db.Close() }()

	repo := NewPostgresSnapshotRepository(db)

	snap := &domain.PortfolioSnapshot{
		ID:        "snap-new",
		UserID:    "user-123",
		Timestamp: time.Now(),
		State: domain.SnapshotState{
			Holdings: map[string]domain.HoldingState{
				"AAPL": {Quantity: "10", CostBasis: "1000", Currency: "USD"},
			},
			Cash:          map[string]string{"USD": "0"},
			RealizedGains: map[string]string{"USD": "0"},
			NetInvested:   "1000",
		},
		TransactionCount: 10,
	}

	// Expectation
	// Note: checking JSON arguments with sqlmock can be tricky. We match any args or specific strings.
	mock.ExpectExec("INSERT INTO investments.portfolio_snapshots").
		WithArgs(snap.UserID, snap.Timestamp, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), snap.State.NetInvested, snap.TransactionCount, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute
	err = repo.UpsertSnapshot(context.Background(), snap)

	// Assert
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
