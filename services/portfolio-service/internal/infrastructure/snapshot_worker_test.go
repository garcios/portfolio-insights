package infrastructure

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
)

type mockPortfolioUsecase struct {
	summary *domain.PortfolioSummary
	err     error
}

func (m *mockPortfolioUsecase) GetHoldings(ctx context.Context, userID string) ([]*domain.Holding, error) {
	return nil, nil
}

func (m *mockPortfolioUsecase) GetPortfolioSummary(ctx context.Context, userID string) (*domain.PortfolioSummary, error) {
	if m.err != nil {
		return nil, m.err
	}
	// Return a summary that matches the user ID if possible, or just the default one
	s := *m.summary
	s.UserID = userID
	return &s, nil
}

func (m *mockPortfolioUsecase) GetHistoricalPortfolioSummary(ctx context.Context, userID string, date time.Time) (*domain.PortfolioSummary, error) {
	return nil, nil
}

type mockPortfolioHistoryRepository struct {
	snapshots []*domain.PortfolioSnapshot
	userIDs   []string
	err       error
}

func (m *mockPortfolioHistoryRepository) CreateSnapshot(ctx context.Context, snapshot *domain.PortfolioSnapshot) error {
	if m.err != nil {
		return m.err
	}
	m.snapshots = append(m.snapshots, snapshot)
	return nil
}

func (m *mockPortfolioHistoryRepository) GetHistory(ctx context.Context, userID string, from, to time.Time) ([]*domain.PortfolioSnapshot, error) {
	return nil, nil
}

func (m *mockPortfolioHistoryRepository) GetHistoryByPeriod(ctx context.Context, userID string, period string) ([]*domain.PortfolioSnapshot, error) {
	return nil, nil
}

func (m *mockPortfolioHistoryRepository) SnapshotExists(ctx context.Context, userID string, date time.Time) (bool, error) {
	return false, nil
}

func (m *mockPortfolioHistoryRepository) GetAllUserIDs(ctx context.Context) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.userIDs, nil
}

func TestSnapshotWorker_CreateSnapshots(t *testing.T) {
	// Setup
	uc := &mockPortfolioUsecase{
		summary: &domain.PortfolioSummary{
			UserID:     "default",
			TotalValue: 1000.0,
			TotalCost:  800.0,
		},
	}
	repo := &mockPortfolioHistoryRepository{
		userIDs: []string{"user-1", "user-2"},
	}
	logger := slog.Default()

	worker := NewSnapshotWorker(uc, repo, logger)

	// Execute via TriggerNow which runs in a goroutine
	worker.TriggerNow()

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Assert
	if len(repo.snapshots) != 2 {
		t.Errorf("Expected 2 snapshots, got %d", len(repo.snapshots))
	}

	// Verify snapshot content
	var s1 *domain.PortfolioSnapshot
	for _, s := range repo.snapshots {
		if s.UserID == "user-1" {
			s1 = s
			break
		}
	}

	if s1 == nil {
		t.Fatal("Snapshot for user-1 not found")
	}

	if s1.TotalValue != 1000.0 {
		t.Errorf("Expected total value 1000.0, got %f", s1.TotalValue)
	}
	if s1.TotalCostBasis != 800.0 {
		t.Errorf("Expected total cost 800.0, got %f", s1.TotalCostBasis)
	}
}
