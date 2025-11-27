# Portfolio History Tracking Strategy

## Overview
Implement a system to capture and store portfolio value snapshots over time to power the performance chart in the frontend. This will replace the current mock data generation with real historical data.

---

## Database Schema (Already Exists)

```sql
CREATE TABLE investments.portfolio_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    total_value DECIMAL(20, 8) NOT NULL,
    total_cost_basis DECIMAL(20, 8) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_portfolio_history_user_id_timestamp 
    ON investments.portfolio_history(user_id, timestamp DESC);
```

**Note**: The schema is already in place from migration `000004_create_portfolio_tables.up.sql`.

---

## Strategy Options

### Option 1: Scheduled Snapshot Job (RECOMMENDED)
**Best for**: Production use, accurate historical tracking, minimal overhead

#### How It Works:
1. **Cron Job / Scheduled Task**: Run a background job at regular intervals (e.g., daily at market close, hourly during trading hours)
2. **Snapshot Process**:
   - For each user with holdings:
     - Calculate current portfolio value using `GetPortfolioSummary` logic
     - Insert snapshot into `portfolio_history` table
3. **Data Retention**: Keep snapshots indefinitely or implement retention policy (e.g., keep daily for 1 year, weekly for 5 years)

#### Implementation Components:

**A. Scheduled Job (New Service or Cron)**
```go
// services/portfolio-service/cmd/snapshot/main.go
// OR
// services/portfolio-service/internal/jobs/snapshot_job.go

func SnapshotAllPortfolios(ctx context.Context) error {
    // 1. Get all unique user_ids from holdings table
    // 2. For each user:
    //    - Call GetPortfolioSummary usecase
    //    - Insert into portfolio_history
    // 3. Log results
}
```

**B. Repository Method**
```go
// services/portfolio-service/internal/repository/portfolio_history_repo.go

type PortfolioHistoryRepository interface {
    CreateSnapshot(ctx context.Context, snapshot *domain.PortfolioSnapshot) error
    GetHistory(ctx context.Context, userID string, from, to time.Time) ([]*domain.PortfolioSnapshot, error)
    GetHistoryByPeriod(ctx context.Context, userID string, period string) ([]*domain.PortfolioSnapshot, error)
}
```

**C. Domain Model**
```go
// services/portfolio-service/internal/domain/portfolio_history.go

type PortfolioSnapshot struct {
    ID            string
    UserID        string
    TotalValue    float64
    TotalCostBasis float64
    Timestamp     time.Time
    CreatedAt     time.Time
}
```

**D. Scheduling Options**:
- **Option 1a**: Kubernetes CronJob (if using K8s)
- **Option 1b**: Standalone cron service in Docker Compose
- **Option 1c**: Go scheduler library (e.g., `github.com/robfig/cron`)
- **Option 1d**: PostgreSQL pg_cron extension

#### Pros:
- ✅ Predictable, consistent snapshots
- ✅ Low runtime overhead (runs off-peak)
- ✅ Easy to backfill historical data
- ✅ Can batch process all users efficiently
- ✅ Decoupled from user activity

#### Cons:
- ❌ Requires additional infrastructure (cron/scheduler)
- ❌ Snapshots only at scheduled times (not real-time)
- ❌ May miss intraday fluctuations

---

### Option 1B: Go Routine Worker (SIMPLEST - HIGHLY RECOMMENDED)
**Best for**: Simple deployment, no extra infrastructure, easy testing

#### How It Works:
1. **Background Goroutine**: Start a worker goroutine in `portfolio-service` that runs continuously
2. **Ticker-Based Scheduling**: Use `time.Ticker` to trigger snapshots at regular intervals
3. **Graceful Shutdown**: Integrate with existing service lifecycle
4. **No External Dependencies**: Everything runs within the existing service

#### Implementation:

**A. Snapshot Worker**
```go
// services/portfolio-service/internal/infrastructure/snapshot_worker.go

package infrastructure

import (
    "context"
    "os"
    "time"
    
    "github.com/garcios/portfolio-insights/pkg/logger"
    "github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
    "github.com/garcios/portfolio-insights/services/portfolio-service/internal/usecase"
)

type SnapshotWorker struct {
    portfolioUsecase usecase.PortfolioUsecase
    historyRepo      domain.PortfolioHistoryRepository
    logger           *logger.Logger
    ticker           *time.Ticker
    stopChan         chan struct{}
}

func NewSnapshotWorker(
    uc usecase.PortfolioUsecase,
    repo domain.PortfolioHistoryRepository,
    log *logger.Logger,
) *SnapshotWorker {
    // Default: snapshot every 24 hours
    // Configurable via SNAPSHOT_INTERVAL env var (e.g., "24h", "1h", "30m")
    interval := 24 * time.Hour
    if envInterval := os.Getenv("SNAPSHOT_INTERVAL"); envInterval != "" {
        if d, err := time.ParseDuration(envInterval); err == nil {
            interval = d
        }
    }
    
    return &SnapshotWorker{
        portfolioUsecase: uc,
        historyRepo:      repo,
        logger:           log,
        ticker:           time.NewTicker(interval),
        stopChan:         make(chan struct{}),
    }
}

func (w *SnapshotWorker) Start() {
    w.logger.Info("Snapshot worker started")
    
    // Optional: Run immediately on startup
    // go w.createSnapshots()
    
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

func (w *SnapshotWorker) Stop() {
    w.ticker.Stop()
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
```

**B. Integration in Main**
```go
// services/portfolio-service/cmd/server/main.go

func main() {
    l := logger.New()
    l.Info("Portfolio Service starting...")
    
    // ... existing setup (DB, Redis, repos, usecases) ...
    
    // Initialize snapshot worker
    snapshotWorker := infrastructure.NewSnapshotWorker(
        portfolioUsecase,
        historyRepo,
        l,
    )
    
    // Start snapshot worker in background
    go snapshotWorker.Start()
    defer snapshotWorker.Stop()
    
    // ... rest of your main (gRPC server, NATS subscriber, etc.) ...
}
```

**C. Advanced: Schedule at Specific Time**
For running at a specific time (e.g., 4:00 PM EST daily):

```go
func (w *SnapshotWorker) Start() {
    w.logger.Info("Snapshot worker started")
    
    for {
        // Calculate next run time (4 PM EST = 21:00 UTC)
        nextRun := w.calculateNextRunTime()
        duration := time.Until(nextRun)
        
        w.logger.Info("Next snapshot scheduled", "time", nextRun)
        
        select {
        case <-time.After(duration):
            w.logger.Info("Starting scheduled portfolio snapshots")
            w.createSnapshots()
        case <-w.stopChan:
            w.logger.Info("Snapshot worker stopped")
            return
        }
    }
}

func (w *SnapshotWorker) calculateNextRunTime() time.Time {
    now := time.Now()
    targetHour := 21  // 4 PM EST = 21:00 UTC
    targetMinute := 0
    
    next := time.Date(
        now.Year(), now.Month(), now.Day(),
        targetHour, targetMinute, 0, 0,
        time.UTC,
    )
    
    // If we've passed today's target time, schedule for tomorrow
    if now.After(next) {
        next = next.Add(24 * time.Hour)
    }
    
    return next
}
```

**D. Using Cron Library (Most Flexible)**
```go
import "github.com/robfig/cron/v3"

type SnapshotWorker struct {
    // ... existing fields ...
    cron *cron.Cron
}

func NewSnapshotWorker(...) *SnapshotWorker {
    c := cron.New()
    
    worker := &SnapshotWorker{
        // ... existing fields ...
        cron: c,
    }
    
    // Schedule: Daily at 4 PM EST (9 PM UTC)
    // Cron format: "minute hour day month weekday"
    c.AddFunc("0 21 * * *", func() {
        worker.logger.Info("Starting scheduled portfolio snapshots")
        worker.createSnapshots()
    })
    
    return worker
}

func (w *SnapshotWorker) Start() {
    w.cron.Start()
    w.logger.Info("Snapshot worker started with cron schedule")
}

func (w *SnapshotWorker) Stop() {
    w.cron.Stop()
}
```

#### Configuration:
```yaml
# docker-compose.yml
portfolio-service:
  environment:
    - SNAPSHOT_INTERVAL=24h  # Simple interval
    # OR use cron library with schedule in code
```

#### Handling Multiple Service Instances:
If running multiple instances of `portfolio-service`, use distributed locking:

```go
func (w *SnapshotWorker) createSnapshots() {
    ctx := context.Background()
    
    // Try to acquire distributed lock (using Redis)
    lockKey := "snapshot:lock"
    lockValue := "1"
    lockTTL := 5 * time.Minute
    
    acquired, err := w.redisClient.SetNX(ctx, lockKey, lockValue, lockTTL).Result()
    if err != nil || !acquired {
        w.logger.Info("Another instance is running snapshots, skipping")
        return
    }
    defer w.redisClient.Del(ctx, lockKey)
    
    // ... rest of snapshot logic ...
}
```

#### Pros:
- ✅ **Simplest implementation** - No external infrastructure needed
- ✅ **Part of existing service** - Runs alongside gRPC server
- ✅ **Easy to test** - Can trigger via `TriggerNow()` method
- ✅ **Graceful shutdown** - Integrates with service lifecycle
- ✅ **Consistent with architecture** - Already using goroutines (NATS subscriber)
- ✅ **No deployment complexity** - Just update existing service
- ✅ **Same logging/metrics** - Uses existing observability stack
- ✅ **Configurable** - Via environment variables

#### Cons:
- ⚠️ **Runs on every instance** - Needs distributed locking if scaled horizontally
- ⚠️ **Coupled to service** - Restarts when service restarts (usually fine)
- ⚠️ **Memory overhead** - Minimal (1 goroutine + ticker)

---

### Option 2: Event-Driven Snapshots
**Best for**: Real-time tracking, event-based architecture

#### How It Works:
1. **Trigger**: When a transaction is processed (via NATS subscriber)
2. **Action**: After updating holdings, create a snapshot
3. **Optimization**: Debounce to avoid too many snapshots (e.g., max 1 per hour per user)

#### Implementation:
```go
// services/portfolio-service/internal/infrastructure/nats_subscriber.go

func (s *NATSSubscriber) handleTransactionCreated(msg *nats.Msg) error {
    // ... existing transaction processing ...
    
    // After updating holdings:
    if s.shouldCreateSnapshot(userID) {
        summary, _ := s.portfolioUsecase.GetPortfolioSummary(ctx, userID)
        s.historyRepo.CreateSnapshot(ctx, &domain.PortfolioSnapshot{
            UserID:         userID,
            TotalValue:     summary.TotalValue,
            TotalCostBasis: summary.TotalCost,
            Timestamp:      time.Now(),
        })
        s.updateLastSnapshotTime(userID)
    }
}
```

#### Pros:
- ✅ Real-time snapshots on portfolio changes
- ✅ No separate infrastructure needed
- ✅ Captures exact state at transaction time

#### Cons:
- ❌ Can create many snapshots (storage overhead)
- ❌ Requires debouncing logic
- ❌ Coupled to transaction events
- ❌ Doesn't capture market price changes without transactions

---

### Option 3: Hybrid Approach (BEST OF BOTH WORLDS)
**Best for**: Production systems requiring both accuracy and real-time updates

#### How It Works:
1. **Daily Scheduled Snapshots**: Run at market close (e.g., 4:00 PM EST) for all users
2. **Event-Driven Snapshots**: Create snapshot on significant transactions (debounced)
3. **Deduplication**: Ensure only one snapshot per user per day

#### Implementation:
```go
func (r *portfolioHistoryRepo) CreateSnapshot(ctx context.Context, snapshot *domain.PortfolioSnapshot) error {
    // Upsert: Insert or update if snapshot for same user+day exists
    query := `
        INSERT INTO investments.portfolio_history 
            (user_id, total_value, total_cost_basis, timestamp)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id, DATE(timestamp))
        DO UPDATE SET 
            total_value = EXCLUDED.total_value,
            total_cost_basis = EXCLUDED.total_cost_basis,
            timestamp = EXCLUDED.timestamp
    `
    // Note: Requires adding unique constraint on (user_id, DATE(timestamp))
}
```

#### Pros:
- ✅ Guaranteed daily snapshots
- ✅ Real-time updates on transactions
- ✅ Efficient storage (max 1 per day per user)
- ✅ Best data quality

#### Cons:
- ❌ More complex implementation
- ❌ Requires schema modification (unique constraint)

---

## Recommended Implementation Plan

### Phase 1: Basic Scheduled Snapshots (Week 1)
1. **Create Repository Layer**
   - `internal/repository/portfolio_history_repo.go`
   - Implement `CreateSnapshot()` and `GetHistoryByPeriod()`

2. **Create Snapshot Job**
   - `cmd/snapshot/main.go` (standalone binary)
   - Iterate all users, calculate summary, insert snapshot

3. **Add to Docker Compose**
   ```yaml
   portfolio-snapshot:
     build:
       context: .
       dockerfile: services/portfolio-service/Dockerfile.snapshot
     environment:
       - POSTGRES_DSN=...
     # Run daily at 4 PM UTC
     command: ["crond", "-f"]
   ```

4. **Test with Manual Execution**
   ```bash
   go run services/portfolio-service/cmd/snapshot/main.go
   ```

### Phase 2: Implement gRPC Endpoint (Week 1-2)
1. **Update Proto** (already defined in `portfolio.proto`)
   - `GetPortfolioPerformance` RPC exists
   - Period options: "1d", "1w", "1m", "3m", "1y", "all"

2. **Implement Handler**
   ```go
   func (h *PortfolioHandler) GetPortfolioPerformance(
       ctx context.Context, 
       req *pb.GetPortfolioPerformanceRequest,
   ) (*pb.GetPortfolioPerformanceResponse, error) {
       history, err := h.historyRepo.GetHistoryByPeriod(ctx, req.UserId, req.Period)
       // Convert to protobuf response
   }
   ```

3. **Period Calculation Logic**
   ```go
   func calculateTimeRange(period string) (from, to time.Time) {
       to = time.Now()
       switch period {
       case "1d": from = to.AddDate(0, 0, -1)
       case "1w": from = to.AddDate(0, 0, -7)
       case "1m": from = to.AddDate(0, -1, 0)
       case "3m": from = to.AddDate(0, -3, 0)
       case "1y": from = to.AddDate(-1, 0, 0)
       case "all": from = time.Time{} // Beginning of time
       }
       return
   }
   ```

### Phase 3: GraphQL Integration (Week 2)
1. **Update Schema**
   ```graphql
   type Portfolio {
       id: ID!
       userId: ID!
       name: String!
       summary: PortfolioSummary
       holdings: [Holding!]!
       performance(period: String!): [PerformancePoint!]!
   }
   
   type PerformancePoint {
       timestamp: String!
       value: Float!
   }
   ```

2. **Add Resolver**
   ```go
   func (r *portfolioResolver) Performance(
       ctx context.Context, 
       obj *model.Portfolio, 
       period string,
   ) ([]*model.PerformancePoint, error) {
       resp, err := r.PortfolioClient.GetPortfolioPerformance(ctx, &portfoliopb.GetPortfolioPerformanceRequest{
           UserId: obj.UserID,
           Period: period,
       })
       // Map to GraphQL model
   }
   ```

### Phase 4: Frontend Integration (Week 2)
1. **Update GraphQL Query**
   ```graphql
   query GetPortfolio($id: ID!, $period: String!) {
       portfolio(id: $id) {
           # ... existing fields ...
           performance(period: $period) {
               timestamp
               value
           }
       }
   }
   ```

2. **Remove Mock Data Generator**
   - Delete `generateMockPerformance()` function
   - Use real data from GraphQL query

3. **Add Period Selector State**
   ```typescript
   const [selectedPeriod, setSelectedPeriod] = useState('1M');
   ```

---

## Data Retention Strategy

### Option A: Keep All Data
- Simple, no cleanup needed
- Good for compliance/auditing
- Storage grows linearly

### Option B: Aggregation Strategy
- **Daily snapshots**: Keep for 1 year
- **Weekly snapshots**: Aggregate daily → weekly after 1 year
- **Monthly snapshots**: Aggregate weekly → monthly after 5 years

### Option C: Sliding Window
- Keep last N days (e.g., 365 days)
- Delete older snapshots automatically

**Recommendation**: Start with Option A, implement Option B if storage becomes an issue.

---

## Performance Considerations

### Query Optimization
```sql
-- Efficient query with index usage
SELECT timestamp, total_value
FROM investments.portfolio_history
WHERE user_id = $1 
  AND timestamp >= $2 
  AND timestamp <= $3
ORDER BY timestamp ASC;
```

### Caching Strategy
- Cache recent performance data (last 24 hours) in Redis
- TTL: 5 minutes
- Key: `portfolio:performance:{user_id}:{period}`

### Pagination
For "all" period with many snapshots:
```sql
SELECT timestamp, total_value
FROM investments.portfolio_history
WHERE user_id = $1
ORDER BY timestamp DESC
LIMIT 1000;  -- Reasonable limit
```

---

## Testing Strategy

### Unit Tests
- Repository layer: Mock database, test CRUD operations
- Usecase layer: Mock repository, test business logic
- Handler layer: Mock usecase, test gRPC responses

### Integration Tests
1. **Snapshot Creation**
   ```go
   func TestSnapshotJob(t *testing.T) {
       // Create test user with holdings
       // Run snapshot job
       // Verify snapshot in database
   }
   ```

2. **Historical Query**
   ```go
   func TestGetPerformance(t *testing.T) {
       // Insert test snapshots
       // Query for different periods
       // Verify correct data returned
   }
   ```

### Manual Testing
```bash
# 1. Create snapshot manually
go run services/portfolio-service/cmd/snapshot/main.go

# 2. Verify in database
psql -d investments -c "SELECT * FROM investments.portfolio_history ORDER BY timestamp DESC LIMIT 10;"

# 3. Test gRPC endpoint
grpcurl -plaintext -d '{"user_id":"test-user","period":"1m"}' \
  localhost:50052 portfolio.PortfolioService/GetPortfolioPerformance
```

---

## Migration & Backfill Strategy

### Overview
Backfilling is a **one-time or infrequent operation** to populate historical portfolio snapshots. This is useful when:
- Initially implementing the history tracking feature
- Importing historical data from another system
- Filling gaps from snapshot worker failures
- Correcting data after bug fixes

### Backfill Approaches

#### Option A: Admin API Endpoint (RECOMMENDED)
**Best for**: Flexibility, remote triggering, production environments

Provides a secure gRPC endpoint that can be called programmatically or via CLI tools.

**1. Update Proto Definition**
```protobuf
// proto/portfolio/portfolio.proto

service PortfolioService {
  rpc GetPortfolioSummary(GetPortfolioSummaryRequest) returns (GetPortfolioSummaryResponse);
  rpc GetPortfolioPerformance(GetPortfolioPerformanceRequest) returns (GetPortfolioPerformanceResponse);
  rpc GetHoldings(GetHoldingsRequest) returns (GetHoldingsResponse);
  
  // Admin endpoint for backfilling history
  rpc BackfillHistory(BackfillHistoryRequest) returns (BackfillHistoryResponse);
}

message BackfillHistoryRequest {
  string user_id = 1;           // Empty = all users
  string start_date = 2;        // YYYY-MM-DD format
  string end_date = 3;          // YYYY-MM-DD format (optional, defaults to today)
  bool dry_run = 4;             // Preview without writing
  string admin_token = 5;       // Authentication token
}

message BackfillHistoryResponse {
  int32 snapshots_created = 1;
  int32 snapshots_skipped = 2;
  int32 errors = 3;
  repeated string error_messages = 4;
  string status = 5;            // "success", "partial", "failed"
}
```

**2. Implement Handler**
```go
// services/portfolio-service/internal/handler/grpc/portfolio_handler.go

func (h *PortfolioHandler) BackfillHistory(
    ctx context.Context,
    req *pb.BackfillHistoryRequest,
) (*pb.BackfillHistoryResponse, error) {
    // 1. Validate admin token
    if !h.validateAdminToken(req.AdminToken) {
        return nil, status.Error(codes.Unauthenticated, "invalid admin token")
    }
    
    // 2. Parse dates
    startDate, err := time.Parse("2006-01-02", req.StartDate)
    if err != nil {
        return nil, status.Errorf(codes.InvalidArgument, "invalid start_date: %v", err)
    }
    
    endDate := time.Now()
    if req.EndDate != "" {
        endDate, err = time.Parse("2006-01-02", req.EndDate)
        if err != nil {
            return nil, status.Errorf(codes.InvalidArgument, "invalid end_date: %v", err)
        }
    }
    
    // 3. Determine users to backfill
    var userIDs []string
    if req.UserId != "" {
        userIDs = []string{req.UserId}
    } else {
        // Get all users with holdings
        userIDs, err = h.historyRepo.GetAllUserIDs(ctx)
        if err != nil {
            return nil, status.Errorf(codes.Internal, "failed to get users: %v", err)
        }
    }
    
    // 4. Run backfill
    result := h.runBackfill(ctx, userIDs, startDate, endDate, req.DryRun)
    
    return &pb.BackfillHistoryResponse{
        SnapshotsCreated: int32(result.Created),
        SnapshotsSkipped: int32(result.Skipped),
        Errors:           int32(result.Errors),
        ErrorMessages:    result.ErrorMessages,
        Status:           result.Status,
    }, nil
}

type BackfillResult struct {
    Created       int
    Skipped       int
    Errors        int
    ErrorMessages []string
    Status        string
}

func (h *PortfolioHandler) runBackfill(
    ctx context.Context,
    userIDs []string,
    startDate, endDate time.Time,
    dryRun bool,
) BackfillResult {
    result := BackfillResult{
        ErrorMessages: []string{},
    }
    
    for _, userID := range userIDs {
        // Backfill for each day in range
        for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
            // Check if snapshot already exists
            exists, _ := h.historyRepo.SnapshotExists(ctx, userID, date)
            if exists {
                result.Skipped++
                continue
            }
            
            if dryRun {
                h.logger.Info("DRY RUN: Would create snapshot",
                    "user_id", userID,
                    "date", date.Format("2006-01-02"),
                )
                result.Created++
                continue
            }
            
            // Create snapshot for this date
            if err := h.createHistoricalSnapshot(ctx, userID, date); err != nil {
                result.Errors++
                result.ErrorMessages = append(result.ErrorMessages,
                    fmt.Sprintf("user=%s date=%s: %v", userID, date.Format("2006-01-02"), err),
                )
                continue
            }
            
            result.Created++
        }
    }
    
    // Determine overall status
    if result.Errors == 0 {
        result.Status = "success"
    } else if result.Created > 0 {
        result.Status = "partial"
    } else {
        result.Status = "failed"
    }
    
    return result
}

func (h *PortfolioHandler) createHistoricalSnapshot(
    ctx context.Context,
    userID string,
    date time.Time,
) error {
    // Get portfolio summary for this user
    // Note: This uses CURRENT prices, not historical prices
    summary, err := h.portfolioUsecase.GetPortfolioSummary(ctx, userID)
    if err != nil {
        return fmt.Errorf("failed to get summary: %w", err)
    }
    
    // Create snapshot with the specified date
    snapshot := &domain.PortfolioSnapshot{
        UserID:         userID,
        TotalValue:     summary.TotalValue,
        TotalCostBasis: summary.TotalCost,
        Timestamp:      date,
    }
    
    return h.historyRepo.CreateSnapshot(ctx, snapshot)
}

func (h *PortfolioHandler) validateAdminToken(token string) bool {
    // Simple token validation
    // In production, use proper authentication (JWT, OAuth, etc.)
    adminToken := os.Getenv("ADMIN_TOKEN")
    return token != "" && token == adminToken
}
```

**3. Add Repository Method**
```go
// services/portfolio-service/internal/repository/portfolio_history_repo.go

func (r *portfolioHistoryRepo) SnapshotExists(
    ctx context.Context,
    userID string,
    date time.Time,
) (bool, error) {
    query := `
        SELECT EXISTS(
            SELECT 1 FROM investments.portfolio_history
            WHERE user_id = $1 
              AND DATE(timestamp) = DATE($2)
        )
    `
    
    var exists bool
    err := r.db.QueryRowContext(ctx, query, userID, date).Scan(&exists)
    return exists, err
}

func (r *portfolioHistoryRepo) GetAllUserIDs(ctx context.Context) ([]string, error) {
    query := `
        SELECT DISTINCT user_id 
        FROM investments.holdings
        WHERE quantity > 0
    `
    
    rows, err := r.db.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var userIDs []string
    for rows.Next() {
        var userID string
        if err := rows.Scan(&userID); err != nil {
            return nil, err
        }
        userIDs = append(userIDs, userID)
    }
    
    return userIDs, rows.Err()
}
```

**4. Usage via grpcurl**
```bash
# Dry run for single user
grpcurl -plaintext \
  -d '{
    "user_id": "user-123",
    "start_date": "2024-01-01",
    "end_date": "2024-12-31",
    "dry_run": true,
    "admin_token": "your-secret-token"
  }' \
  localhost:50052 portfolio.PortfolioService/BackfillHistory

# Actual backfill for all users
grpcurl -plaintext \
  -d '{
    "start_date": "2024-01-01",
    "admin_token": "your-secret-token"
  }' \
  localhost:50052 portfolio.PortfolioService/BackfillHistory
```

**5. Configuration**
```yaml
# docker-compose.yml
portfolio-service:
  environment:
    - ADMIN_TOKEN=${ADMIN_TOKEN:-change-me-in-production}
```

```bash
# .env
ADMIN_TOKEN=super-secret-admin-token-12345
```

#### Option B: Standalone CLI Tool
**Best for**: One-time migrations, simple scripts

```go
// cmd/backfill/main.go
package main

import (
    "context"
    "flag"
    "fmt"
    "time"
    
    "github.com/garcios/portfolio-insights/pkg/logger"
    "github.com/garcios/portfolio-insights/services/portfolio-service/internal/infrastructure"
    "github.com/garcios/portfolio-insights/services/portfolio-service/internal/repository"
    "github.com/garcios/portfolio-insights/services/portfolio-service/internal/usecase"
)

func main() {
    userID := flag.String("user-id", "", "User ID to backfill (empty = all users)")
    fromDate := flag.String("from", "", "Start date (YYYY-MM-DD)")
    toDate := flag.String("to", "", "End date (YYYY-MM-DD, default: today)")
    dryRun := flag.Bool("dry-run", false, "Preview without writing")
    flag.Parse()
    
    if *fromDate == "" {
        fmt.Println("Error: --from date is required")
        flag.Usage()
        return
    }
    
    // Parse dates
    startDate, err := time.Parse("2006-01-02", *fromDate)
    if err != nil {
        fmt.Printf("Error parsing from date: %v\n", err)
        return
    }
    
    endDate := time.Now()
    if *toDate != "" {
        endDate, err = time.Parse("2006-01-02", *toDate)
        if err != nil {
            fmt.Printf("Error parsing to date: %v\n", err)
            return
        }
    }
    
    // Initialize dependencies
    l := logger.New()
    db, err := infrastructure.NewPostgresDB()
    if err != nil {
        l.Error("failed to connect to database", "error", err)
        return
    }
    defer db.Close()
    
    historyRepo := repository.NewPortfolioHistoryRepository(db)
    holdingRepo := repository.NewPostgresHoldingRepository(db)
    
    // ... rest of backfill logic ...
    
    fmt.Printf("Backfill completed: %d snapshots created\n", count)
}
```

### Important Considerations

#### 1. Historical Price Data Challenge
**Problem**: Backfilling requires historical prices, which you may not have.

**Solutions**:
- **Option A**: Use current prices (approximation)
  ```go
  // Simple but inaccurate - uses today's prices for all dates
  summary, _ := portfolioUsecase.GetPortfolioSummary(ctx, userID)
  ```

- **Option B**: Integrate historical price API
  ```go
  // Fetch historical prices from external API
  price, _ := alphaVantageClient.GetHistoricalPrice(symbol, date)
  ```

- **Option C**: Store prices as you go (for future backfills)
  ```sql
  -- Keep historical prices in marketdata-service
  CREATE TABLE marketdata.price_history (
      symbol VARCHAR(20),
      price DECIMAL(20, 8),
      date DATE,
      PRIMARY KEY (symbol, date)
  );
  ```

#### 2. Performance Optimization
For large backfills:

```go
// Batch insert snapshots
func (r *portfolioHistoryRepo) CreateSnapshotsBatch(
    ctx context.Context,
    snapshots []*domain.PortfolioSnapshot,
) error {
    // Use COPY or multi-row INSERT for better performance
    query := `
        INSERT INTO investments.portfolio_history 
            (user_id, total_value, total_cost_basis, timestamp)
        VALUES ($1, $2, $3, $4)
    `
    
    tx, _ := r.db.BeginTx(ctx, nil)
    defer tx.Rollback()
    
    stmt, _ := tx.PrepareContext(ctx, query)
    defer stmt.Close()
    
    for _, snapshot := range snapshots {
        _, err := stmt.ExecContext(ctx,
            snapshot.UserID,
            snapshot.TotalValue,
            snapshot.TotalCostBasis,
            snapshot.Timestamp,
        )
        if err != nil {
            return err
        }
    }
    
    return tx.Commit()
}
```

#### 3. Idempotency
Ensure backfill can be run multiple times safely:

```go
// Check before inserting
if exists, _ := repo.SnapshotExists(ctx, userID, date); exists {
    continue // Skip existing snapshots
}
```

### Practical Recommendations

1. **Start Fresh**: Don't backfill initially - let snapshots build naturally from today
2. **Demo Data**: Use fake data generator for testing the chart UI
3. **Real Backfill**: Only if you have historical transaction data AND historical prices
4. **Admin API**: Preferred for production - more flexible and secure
5. **Monitoring**: Log all backfill operations for audit trail

### Demo Data Generator (For Testing)
```go
// cmd/generate-demo-history/main.go
// Creates fake historical data for demonstration

func generateDemoHistory(userID string, days int) {
    baseValue := 10000.0
    
    for i := days; i >= 0; i-- {
        date := time.Now().AddDate(0, 0, -i)
        
        // Simulate growth with noise
        trend := float64(days-i) * 50      // $50/day growth
        noise := (rand.Float64() - 0.5) * 300  // ±$150 random
        value := baseValue + trend + noise
        
        snapshot := &domain.PortfolioSnapshot{
            UserID:         userID,
            TotalValue:     value,
            TotalCostBasis: baseValue,
            Timestamp:      date,
        }
        
        repo.CreateSnapshot(ctx, snapshot)
    }
}
```

**Note**: For production backfilling with accurate historical data, you'll need to integrate with a historical price data provider (Alpha Vantage, Yahoo Finance, Polygon.io, etc.).

---

## Monitoring & Alerting

### Metrics to Track
- Snapshot job execution time
- Number of snapshots created per run
- Failed snapshot attempts
- Query performance for historical data

### Prometheus Metrics
```go
var (
    snapshotDuration = prometheus.NewHistogram(...)
    snapshotsCreated = prometheus.NewCounter(...)
    snapshotErrors   = prometheus.NewCounter(...)
)
```

---

## Summary & Recommendation

**Recommended Approach**: **Option 1B (Go Routine Worker)** for initial implementation

### Why?
1. ✅ **Simplest to implement** - No external infrastructure or containers needed
2. ✅ **Part of existing service** - Runs alongside your gRPC server
3. ✅ **Easy to test** - Can trigger manually via `TriggerNow()` method
4. ✅ **Existing schema supports it** - Database tables already in place
5. ✅ **Consistent with your architecture** - Already using goroutines (NATS subscriber)
6. ✅ **Graceful shutdown** - Integrates with service lifecycle
7. ✅ **No deployment complexity** - Just update existing `portfolio-service`
8. ✅ **Can evolve** - Easy to add event-driven snapshots or switch to external cron later

### Comparison: Go Routine vs External Cron

| Aspect | Go Routine Worker | External Cron/Job |
|--------|------------------|-------------------|
| **Simplicity** | ✅ Very simple | ❌ More complex |
| **Deployment** | ✅ Part of service | ❌ Separate container |
| **Resource Usage** | ✅ Minimal (1 goroutine) | ❌ Separate process |
| **Testing** | ✅ Easy (`TriggerNow()`) | ⚠️ Harder |
| **Observability** | ✅ Same logs/metrics | ⚠️ Separate logs |
| **Failure Recovery** | ✅ Auto-restarts | ❌ Needs monitoring |
| **Configuration** | ✅ Env vars | ⚠️ Cron syntax |

### Implementation Timeline
- **Week 1**: Repository + Snapshot Worker + Integration
- **Week 2**: gRPC endpoint + GraphQL resolver + Frontend integration
- **Week 3**: Testing, monitoring, optimization

### Next Steps
1. Create `internal/domain/portfolio_history.go`
2. Create `internal/repository/portfolio_history_repo.go`
3. Create `internal/infrastructure/snapshot_worker.go` ⭐ **NEW**
4. Wire up in `cmd/server/main.go`
5. Test with `SNAPSHOT_INTERVAL=1m` for quick validation
6. Implement gRPC handler
7. Update GraphQL schema and resolver
8. Update frontend to use real data

### Quick Start Configuration
```yaml
# docker-compose.yml
portfolio-service:
  environment:
    # Run snapshots every 24 hours
    - SNAPSHOT_INTERVAL=24h
    
    # For testing: Run every 1 minute
    # - SNAPSHOT_INTERVAL=1m
```

This approach provides the simplest foundation that can be enhanced with event-driven snapshots, external cron jobs, or real-time updates in the future if needed.
