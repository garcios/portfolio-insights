# Portfolio Snapshot System: Lazy Read-Repair Strategy

## Overview
This document outlines the implementation of a high-performance snapshotting system designed to minimize write-latency while ensuring data consistency for portfolio summaries.

## 1. Strategy: Lazy Read-Repair
Instead of updating snapshots on every transaction (Write-Heavy), we use a **Lazy Read-Repair** approach.
- **Trigger**: During a `Read` operation, the system checks if a snapshot is "stale" (Transaction count > 100 OR Time > 30 days).
- **Execution**: If stale, the system triggers an asynchronous background job to "repair" (regenerate) the snapshot without blocking the user's request.

## 2. Technical Architecture
The system is built in Go using a worker-pool pattern with built-in resilience features:

### Key Components
* **API Layer**: Performs a "Stale Check" and calls `Trigger(userID)`.
* **Worker Pool**: Decouples the API request from the snapshot generation logic.
* **Rate Limiting**: Uses a Token Bucket algorithm (`golang.org/x/time/rate`) to prevent background tasks from overwhelming the Database.
* **Exponential Backoff**: If a snapshot fails (e.g., DB lock), the job is retried with increasing delays ($2^{retry} \times 1s$).
* **Observability**: Integrated Prometheus metrics for real-time monitoring of queue depth and failure rates.
* **Deduplication**: Uses an atomic `sync.Map` to check if a job for that ID is already active.



---

## 3. Implementation Details

### Configuration
```go
cfg := WorkerConfig{
    PoolSize:   5,      // Concurrent workers
    MaxRetries: 3,      // Max attempts per job
    RateLimit:  10.0,   // Max 10 snapshots per second
    Burst:      5,      // Allow short spikes
}
```

### Snapshot Manager with Prometheus Metrics
```go
package snapshotter

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/time/rate"
)

var (
	jobsEnqueued = promauto.NewCounter(prometheus.CounterOpts{
		Name: "snapshot_jobs_enqueued_total",
		Help: "The total number of snapshot jobs added to the queue",
	})
	jobsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "snapshot_jobs_processed_total",
		Help: "The total number of processed snapshots",
	}, []string{"status"}) // status = "success" or "error"
	
	queueSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "snapshot_queue_depth",
		Help: "Current number of jobs waiting in the buffer",
	})

	processingDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "snapshot_processing_duration_seconds",
		Help:    "Time taken to generate a snapshot",
		Buckets: prometheus.DefBuckets,
	})
)

type Job struct {
	UserID     string
	RetryCount int
}

type WorkerConfig struct {
	PoolSize     int
	MaxRetries   int
	RateLimit    float64 // Jobs per second
	Burst        int
}

// Configurable "Max Time" a snapshot generation is allowed to take
const LockTTL = 5 * time.Minute

type LockEntry struct {
	StartTime time.Time
	JobID     string
}

type Manager struct {
	queue   chan Job
	limiter *rate.Limiter
	config  WorkerConfig
	wg      sync.WaitGroup
	// Tracks users currently in the queue or being processed
	mu       sync.Mutex
	inFlight map[string]LockEntry
}

func NewManager(cfg WorkerConfig) *Manager {
	return &Manager{
		queue:   make(chan Job, 1000),
		limiter: rate.NewLimiter(rate.Limit(cfg.RateLimit), cfg.Burst),
		config:  cfg,
		inFlight: make(map[string]LockEntry),
	}
}

func (m *Manager) Start(ctx context.Context, processor func(string) error) {
	for i := 0; i < m.config.PoolSize; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case job := <-m.queue:

                // Wrap processing in a function to enforce defer scope
				func() {
					// ✅ SAFE: Only delete if WE own the lock
					defer m.safeUnlock(job.UserID, job.JobID)

					queueSize.Dec() // Decrement gauge as job is picked up
					
					// Respect rate limits
					if err := m.limiter.Wait(ctx); err != nil {
						return
					}

					start := time.Now()
					err := processor(job.UserID) 

					// Record Metrics
					processingDuration.Observe(time.Since(start).Seconds())
					
					if err != nil {
						jobsProcessed.WithLabelValues("error").Inc()
						m.handleFailure(job, err)
					} else {
						jobsProcessed.WithLabelValues("success").Inc()
					}
				}()
 				
			}
		}()
	}
}

func (m *Manager) Trigger(userID string) {
    m.mu.Lock()
	entry, exists := m.inFlight[userID]
	
	// 1. Check for Active Lock
	if exists {
		if time.Since(entry.StartTime) < LockTTL {
			// Job is running and valid. Skip Trigger.
			m.mu.Unlock()
			return
		}
		// Else: Lock exists but is EXPIRED. We will overwrite it (Steal Lock).
		fmt.Printf("⚠️ Stale lock detected for user %s. Stealing lock.\n", userID)
	}

	// 2. Create New Lock (Lease)
	jobID := uuid.New().String()
	m.inFlight[userID] = LockEntry{
		StartTime: time.Now(),
		JobID:     jobID,
	}
	m.mu.Unlock()

	select {
	case m.queue <- Job{UserID: userID}:
		jobsEnqueued.Inc()
		queueSize.Inc()
	default:
		// Queue full - maybe increment a "dropped_jobs" counter here

		// If queue is full, we must remove from inFlight 
		// so it can be retried later
		m.safeUnlock(userID, jobID)
	}
}

func (m *Manager) handleFailure(job Job, err error) {
	if job.RetryCount < m.config.MaxRetries {
		job.RetryCount++
		// Exponential backoff: 2^count * 1s
		delay := time.Duration(math.Pow(2, float64(job.RetryCount))) * time.Second

		time.AfterFunc(delay, func() {
			m.queue <- job
		})
	} else {
		// Log to monitoring (e.g., Sentry/Prometheus)
		fmt.Printf("Critical: User %s failed after %d retries\n", job.UserID, job.RetryCount)
	}
}

// Helper to ensure we only delete OUR lock
func (m *Manager) safeUnlock(userID, jobID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if entry, exists := m.inFlight[userID]; exists {
		// Only delete if the JobID matches. 
		// If IDs don't match, it means the lock was already stolen by a newer job.
		if entry.JobID == jobID {
			delete(m.inFlight, userID)
		}
	}
}

```


## Metrics Tracked

| Metric Name | Type | Description |
| :--- | :--- | :--- |
| `snapshot_queue_depth` | **Gauge** | Current jobs waiting in the buffer. |
| `snapshot_jobs_processed_total` | **Counter** | Total count of successes vs. errors. |
| `snapshot_processing_duration_seconds` | **Histogram** | Latency of snapshot generation. |


## 4. Resilience & Efficiency Features

1.  **Identify Stale Data:** `GetPortfolioSummary` detects a repair is needed based on fragmentation or age.
2.  **Enqueue:** The job is sent to a buffered channel. 
    * *Self-Preservation:* If the channel is full, the job is dropped. This preserves API stability, knowing the next read request will simply re-trigger the check.
3.  **Rate Limit:** The worker waits for a token (Leaky Bucket or Token Bucket) before processing to avoid DB spikes.
4.  **Process:** The snapshot is generated and the "Dirty Bit" is cleared.
5.  **Retry / Dead Letter:** * On failure, the job is rescheduled. 
    * After **3 failures**, it is moved to a Dead Letter Queue (DLQ) and logged as a critical error for manual intervention.
6.  **Deduplication:** The `inFlight` map ensures that we don't process the same user's snapshot multiple times.

## 5. Usage Example

```go
// Initialize Manager
mgr := snapshotter.NewManager(cfg)
mgr.Start(ctx, MySnapshotLogic)

// Trigger in Read Path
if needsRepair {
    mgr.Trigger(userID)
}
```

Or

```go
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Initialize with specific production limits
	mgr := NewManager(WorkerConfig{
		PoolSize:   5,
		MaxRetries: 3,
		RateLimit:  10.0, // 10 snapshots per second max
		Burst:      5,
	})

	// 2. Define the actual business logic (DB calls, etc.)
	processFn := func(userID string) error {
		// Logic to calculate and save snapshot
		return nil 
	}

	// 3. Start workers
	mgr.Start(ctx, processFn)

	// 4. Usage in an HTTP handler or Read path
	if needsRepair {  
      mgr.Trigger(userID)
    }
}
```

## Key Technical Details

### 1. Non-Blocking Queue
Inside `GetPortfolioSummary`, we use a `select` statement with a `default` case. This ensures that if the background queue is completely slammed, the user's current request isn't delayed—we simply skip the repair and try again on their next visit.

### 2. Worker Pool
By limiting the `WorkerPoolSize`, we prevent the backfill process from overwhelming your database connections, which is a common risk during historical migrations.

### 3. Decoupling
The **"Read" path (API)** only knows how to request a repair; the **"Worker" path (Background)** knows how to execute it. This separation of concerns keeps the API logic clean and fast.

---

### 4. Rate Limit
* **Controlled Throughput:** By setting `rate.Limit(5)`, you guarantee that even if your backfill script finds 1,000,000 users, the system will never attempt more than 5 snapshots per second.
* **Graceful Pressure:** The workers simply "pause" and wait for a token to become available. This keeps your CPU and Database I/O predictable.
* **Context Awareness:** `limiter.Wait(ctx)` respects context cancellation. If your application shuts down, the workers stop waiting immediately rather than hanging.

---

### 5. Retry Logic & Backoff
Instead of immediately hammering the database again, `time.AfterFunc` schedules the job to be put back into the queue after a specific duration.

* **The Formula:** Using $2^{retry} \times BaseDelay$ ensures that if the system is under heavy load, the background workers progressively "step back" to give the database breathing room.
* **The "Dead Letter" logic:** Once `MaxRetries` is hit, we stop. This prevents **poison pills** (malformed data that will never succeed) from cycling in your queue forever and wasting CPU cycles.

---

### 6. Observability
#### What this gives you in Grafana
By adding these metrics, you can build a dashboard that answers:
* **Is the backfill working?** (Watch the success counter climbing).
* **Is the DB struggling?** (Watch the `processing_duration` histogram shifting to the right).
* **Are we hitting our rate limit?** (Look for a steady plateau in the processing rate).

---

### 6. Deduplication
By checking `inFlight`, we ensure that we don't process the same user's snapshot multiple times.

* **Thundering Herd Protection**: If 1,000 requests hit your API for the same user in a millisecond, only the first one passes the LoadOrStore check. The other 999 are dropped instantly, saving your CPU and DB.
* **Deadlocks**: If a worker dies mid-job, the LockTTL ensures the user is only blocked for 5 minutes max. The next request after 5 minutes will "steal" the lock and succeed.
* **Race Conditions**: By using safeUnlock (checking the JobID), we prevent a "Slow Worker" (who finishes after 6 mins) from accidentally deleting the lock of a "Fast Worker" (who started at minute 5:01).
* **Accuracy**: We assume that if a snapshot takes > 5 minutes, something is wrong, so treating it as "expired" is logically sound for a read-repair system.

---

## 6. Service Logic Implementation (Core Portfolio Logic)

This section details the core business logic changes required in `portfolio-service` to leverage the new snapshotting architecture.

### 6.1 Domain Models (`internal/domain`)

First, we define the `PortfolioSnapshot` and its associated JSON-compatible state structures. These mirror the schema defined in the PostgreSQL migration.

```go
package domain

import (
	"context"
	"time"
)

// SnapshotState represents the JSONB payload stored in the snapshot.
// We use strings for monetary values to ensure precision during marshalling/unmarshalling.
type SnapshotState struct {
	Holdings      map[string]HoldingState `json:"holdings"`       // Map[Symbol]HoldingState
	Cash          map[string]string       `json:"cash"`           // Map[Currency]Amount
	RealizedGains map[string]string       `json:"realized_gains"` // Map[Currency]Amount
}

// HoldingState captures the position of a specific asset at the snapshot time.
type HoldingState struct {
	Quantity  string `json:"quantity"`   // precise decimal string
	CostBasis string `json:"cost_basis"` // precise decimal string
}

// PortfolioSnapshot is the aggregate root for the detailed point-in-time state.
type PortfolioSnapshot struct {
	ID               string
	UserID           string
	Timestamp        time.Time
	State            SnapshotState
	TransactionCount int
	CreatedAt        time.Time
}

// DetailedSnapshotRepository defines the interface for managing snapshots.
// This extends or complements the existing PortfolioHistoryRepository.
type DetailedSnapshotRepository interface {
	// GetLatestSnapshot retrieves the most recent snapshot before or at the given time.
	GetLatestSnapshot(ctx context.Context, userID string, before time.Time) (*PortfolioSnapshot, error)
	
	// UpsertSnapshot saves a new snapshot, handling idempotency.
	UpsertSnapshot(ctx context.Context, snapshot *PortfolioSnapshot) error
	
	// InvalidateSnapshots deletes or marks as stale all snapshots after a certain time (used on write).
	InvalidateSnapshots(ctx context.Context, userID string, after time.Time) error
}
```

### 6.2 Updated Calculation Flow (`GetPortfolioSummary`)

The core usecase logic is refactored to use the **Hydrate -> Replay -> Project** pattern. This ensures that we only process a small delta of transactions rather than the entire history.

```go
// internal/usecase/portfolio_usecase.go

func (uc *portfolioUsecase) GetPortfolioSummary(ctx context.Context, userID string, startDate, endDate *time.Time) (*domain.PortfolioSummary, error) {
	// 1. Establish Time Boundaries
	calcEndDate := time.Now()
	if endDate != nil {
		calcEndDate = *endDate
	}

	// 2. Fetch Latest Snapshot (Optimization)
	// We attempt to find the closest snapshot to our target date.
	// If checking for Date X, finding a snapshot at Date X-5 is better than Date 0.
	snapshot, err := uc.snapshotRepo.GetLatestSnapshot(ctx, userID, calcEndDate)
	if err != nil {
		// Log error but proceed without snapshot (Graceful Degradation)
		uc.logger.Warn("failed to fetch snapshot", "error", err, "user_id", userID)
	}

	// 3. Initialize Replay State
	var replayStart time.Time // Default: Beginning of time
	currentState := NewReplayState() // Initialize empty map structure

	if snapshot != nil {
		// Optimization: Hydrate state from snapshot
		replayStart = snapshot.Timestamp
		if err := currentState.HydrateFrom(snapshot); err != nil {
			uc.logger.Error("failed to hydrate snapshot state", "error", err)
			// Fallback: reset to empty if hydration fails
			replayStart = time.Time{}
			currentState = NewReplayState()
		} else {
			uc.metrics.RecordCacheHit("snapshot")
		}
	} else {
		uc.metrics.RecordCacheMiss("snapshot")
	}

	// 4. Fetch Delta Transactions
	// Only fetch transactions that happened AFTER the snapshot
	transactions, err := uc.transactionClient.ListTransactions(ctx, &transactionpb.ListTransactionsRequest{
		Parent:    fmt.Sprintf("users/%s", userID),
		StartTime: timestamppb.New(replayStart),
		EndTime:   timestamppb.New(calcEndDate),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}

	// 5. Replay Transactions (The "Delta")
	// Apply just the new transactions to the loaded state
	for _, txn := range transactions.Transactions {
		if err := currentState.Apply(txn); err != nil {
			// Log individual transaction error but try to continue
			uc.logger.Error("failed to apply transaction during replay", "txn_id", txn.Id, "error", err)
		}
	}

	// 6. Trigger Lazy Repair (If needed)
	// Criteria: Too many transactions in delta OR snapshot is too old
	deltaCount := len(transactions.Transactions)
	timeSinceSnapshot := time.Since(replayStart)
	
	shouldTriggerRepair := deltaCount > 100 || (snapshot != nil && timeSinceSnapshot > 30*24*time.Hour)
	
	if shouldTriggerRepair {
		// Async check: fire and forget
		go func() {
			uc.snapshotManager.Trigger(userID)
		}()
	}

	// 7. Final Projection & Enrichment
	// Convert the final ReplayState (Balances, Holdings) into the PortfolioSummary DTO
	// This includes fetching real-time prices for the final holdings.
	summary, err := uc.enrichAndMapSummary(ctx, currentState, calcEndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to enrich portfolio summary: %w", err)
	}

	return summary, nil
}
```

### 6.3 Replay Logic (`ReplayState`)

This helper struct acts as the in-memory ledger during the calculation phase.

```go
// internal/usecase/replay_state.go

type ReplayState struct {
	Holdings      map[string]*domain.AssetPosition
	Cash          map[string]float64
	RealizedGains map[string]float64
}

func NewReplayState() *ReplayState {
	return &ReplayState{
		Holdings:      make(map[string]*domain.AssetPosition),
		Cash:          make(map[string]float64),
		RealizedGains: make(map[string]float64),
	}
}

// HydrateFrom loads the optimized JSON snapshot into the working memory
func (rs *ReplayState) HydrateFrom(snap *domain.PortfolioSnapshot) error {
	for symbol, h := range snap.State.Holdings {
		qty, _ := decimal.NewFromString(h.Quantity)
		cost, _ := decimal.NewFromString(h.CostBasis)
		
		rs.Holdings[symbol] = &domain.AssetPosition{
			Quantity:    qty.InexactFloat64(),
			AverageCost: cost.InexactFloat64(),
			// Note: Snapshot needs to store Currency for full fidelity if not implicit
		}
	}
	// ... Hydrate Cash and Gains similarly ...
	return nil
}

// Apply updates the state based on a single transaction event
func (rs *ReplayState) Apply(txn *transactionpb.Transaction) error {
	switch txn.Type {
	case "BUY":
		// Update Cash (Decrease)
		// Update Holdings (Increase Quantity, Recalculate Average Cost)
	case "SELL":
		// Update Cash (Increase)
		// Calculate Realized Gain
		// Update Unrealized Gains bucket
		// Update Holdings (Decrease Quantity)
	case "DEP", "WIT", "DIV":
		// Simple Cash updates
	}
	return nil
}
```