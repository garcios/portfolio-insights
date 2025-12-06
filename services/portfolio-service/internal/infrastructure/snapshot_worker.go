package infrastructure

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
	"github.com/garcios/portfolio-insights/services/portfolio-service/internal/usecase"
)

// SnapshotWorker handles periodic portfolio snapshots
type SnapshotWorker struct {
	portfolioUsecase usecase.PortfolioUsecase
	historyRepo      domain.PortfolioHistoryRepository
	logger           *slog.Logger
	ticker           *time.Ticker
	stopChan         chan struct{}
	interval         time.Duration
}

// NewSnapshotWorker creates a new SnapshotWorker
func NewSnapshotWorker(
	uc usecase.PortfolioUsecase,
	repo domain.PortfolioHistoryRepository,
	log *slog.Logger,
) *SnapshotWorker {
	// Default: snapshot every 24 hours
	// Configurable via SNAPSHOT_INTERVAL env var (e.g., "24h", "1h", "30m")
	interval := 24 * time.Hour

	// We can also allow smaller intervals for debugging
	if envInterval := os.Getenv("SNAPSHOT_INTERVAL"); envInterval != "" {
		if d, err := time.ParseDuration(envInterval); err == nil {
			interval = d
		}
	}

	return &SnapshotWorker{
		portfolioUsecase: uc,
		historyRepo:      repo,
		logger:           log,
		interval:         interval,
		stopChan:         make(chan struct{}),
	}
}

// Start the worker
func (w *SnapshotWorker) Start() {
	w.logger.Info("Snapshot worker started", "interval", w.interval)

	w.ticker = time.NewTicker(w.interval)

	// Run immediately on startup in a goroutine
	go w.createSnapshots()

	for {
		select {
		case <-w.ticker.C:
			w.logger.Info("Starting scheduled portfolio snapshots")
			w.createSnapshots()
		case <-w.stopChan:
			w.logger.Info("Snapshot worker stopped")
			return
		}
	}
}

// Stop the worker
func (w *SnapshotWorker) Stop() {
	if w.ticker != nil {
		w.ticker.Stop()
	}
	close(w.stopChan)
}

func (w *SnapshotWorker) createSnapshots() {
	ctx := context.Background()

	// Get all unique user IDs from holdings
	userIDs, err := w.historyRepo.GetAllUserIDs(ctx)
	if err != nil {
		w.logger.Error("failed to get user IDs", "error", err)
		return
	}

	successCount := 0
	errorCount := 0

	for _, userID := range userIDs {
		if err := w.createSnapshotForUser(ctx, userID); err != nil {
			w.logger.Error("failed to create snapshot", "user_id", userID, "error", err)
			errorCount++
		} else {
			successCount++
		}
	}

	w.logger.Info("Snapshot batch completed",
		"success", successCount,
		"errors", errorCount,
		"total", len(userIDs),
	)
}

func (w *SnapshotWorker) createSnapshotForUser(ctx context.Context, userID string) error {
	// Get current portfolio summary
	summary, err := w.portfolioUsecase.GetPortfolioSummary(ctx, userID)
	if err != nil {
		return err
	}

	// Create snapshot
	snapshot := &domain.PortfolioSnapshot{
		UserID:         userID,
		TotalValue:     summary.TotalValue,
		TotalCostBasis: summary.TotalCost,
		Timestamp:      time.Now(),
	}

	return w.historyRepo.CreateSnapshot(ctx, snapshot)
}

// TriggerNow allows manual triggering (useful for testing)
func (w *SnapshotWorker) TriggerNow() {
	go w.createSnapshots()
}
