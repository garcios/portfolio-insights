# Incremental Aggregation Implementation Plan

This document details the technical implementation of the Incremental Aggregation strategy outlined in [OPTIMIZATION_PLAN.md](./OPTIMIZATION_PLAN.md) (Section 3.4).

## 1. Data Modeling

We will introduce a new `portfolio_snapshots` table in the `investments` schema. This table acts as a checkpoint, storing the complete state of a user's portfolio at a specific point in time.

### 1.1. Schema Definition

```sql
CREATE TABLE IF NOT EXISTS investments.portfolio_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- State Snapshots (Stored as JSONB for flexibility and performance)
    -- Map<Symbol, {Quantity: string, CostBasis: string}>
    holdings_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    
    -- Map<Currency, Balance: string>
    cash_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    
    -- Map<Currency, Amount: string>
    realized_gains_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    
    transaction_count INTEGER NOT NULL DEFAULT 0, -- Checksum/Validation
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES customers.users(id) ON DELETE CASCADE,
    UNIQUE (user_id, timestamp)
);

CREATE INDEX idx_portfolio_snapshots_user_timestamp 
ON investments.portfolio_snapshots(user_id, timestamp DESC);
```

### 1.2. JSONB Structure Details

**holdings_snapshot**:
```json
{
  "AAPL": {"quantity": "10.0000", "cost_basis": "1500.00"},
  "GOOGL": {"quantity": "5.0000", "cost_basis": "12000.00"}
}
```

**realized_gains_snapshot** (Cumulative realized gains since inception):
```json
{
  "USD": "500.25",
  "AUD": "120.00"
}
```

**cash_snapshot**:
```json
{
  "USD": "1050.00",
  "AUD": "50.00"
}
```

*Note: We use string representations for decimals in JSON to ensure precision is maintained.*

## 2. Service Logic (portfolio-service)

### 2.1. Updated Calculation Flow (`GetPortfolioSummary`)

The `calculatePeriodGains` function will be refactored to:

1.  **Fetch Latest Snapshot**: Query `portfolio_snapshots` for the most recent snapshot where `timestamp <= request.StartDate`.
2.  **Initialize State**:
    *   **If Snapshot Found**: Load `Holdings`, `Cash`, and `RealizedGains` from the snapshot. Set `replayStartDate = snapshot.timestamp`.
    *   **If Valid Snapshot Not Found**: Initialize empty state. Set `replayStartDate = user.CreatedAt`.
3.  **Fetch Transactions**: Query `TransactionRepository` for transactions where `executed_at > replayStartDate` AND `executed_at <= request.EndDate`.
4.  **Replay**: Iterate through fetched transactions and apply them to the current state (update holdings, cash, realized gains).
5.  **Finalize**: Return the projected state.

### 2.2. Construction of PortfolioSummary Response

The "Projected State" (result of replay) maps directly to the `PortfolioSummary` response object:

*   **Current Holdings**: The `holdings` map from the projected state contains the final Quantity and Cost Basis for each asset.
    *   *Enrichment*: Current prices and FX rates are fetched in bulk (per Section 3.1 of Optimization Plan) to calculate `MarketValue`, `UnrealizedGain`, and `Allocation` percentage for each holding.
*   **Total Equity**: Sum of all holdings' `MarketValue`.
*   **Cash Balances**: The `cash` map from the projected state provides the current balance for each currency.
*   **Total Portfolio Value**: `Total Equity + Total Cash (converted to target currency)`.
*   **Period Gains**:
    *   *Realized Gains*: Derived from the `realized_gains` map in the projected state.
    *   *Unrealized Gains*: Calculated dynamically as `(Current Market Value - Cost Basis)` for all active holdings.
    *   *Total Gain*: `Realized Gains + Unrealized Gains`.

This approach ensures that valid snapshots provide 90%+ of the data instantly, with only the small delta of recent transactions needing processing.

### 2.2. Triggering Updates (Snapshot Generation)

We will adopt a **Lazy Read-Repair** strategy with **Async Backfill** to minimize write penalties on critical paths.

1.  **Lazy Creation (On Read)**:
    *   During `GetPortfolioSummary`, if the number of replayed transactions exceeds a threshold (e.g., `MaxReplayCount = 100`) OR time since last snapshot > `SnapshotInterval (e.g., 30 days)`:
    *   Queue a background job (via internal Go channel or worker) to generate a new snapshot at `Now()` (or `request.EndDate`).

2.  **Async Backfill (Historical Data)**:
    *   A background migration worker can iterate through active users with high transaction counts and generate monthly snapshots to seed the system.

## 3. Consistency & Concurrency

Maintaining data integrity between the raw ledger (`txn.transactions`) and aggregate snapshots is critical.

### 3.1. Immutable History & Invalidation
Snapshots are derived data. If the source of truth (transactions) changes, related snapshots become stale.

**Strategy: Invalidation on Write**
*   **Event Listener**: The `portfolio-service` already listens to NATS events (`TransactionCreated`, `TransactionUpdated`, `TransactionDeleted`).
*   **Action**: Upon receiving an event for a modified transaction determining its `event.ExecutedAt`:
    *   **Delete** all snapshots for that `user_id` where `snapshot.timestamp >= event.ExecutedAt`.
    *   *Alternative (Optimization)*: Mark them as `dirty` if we implement a repair queue, but deletion is safer and simpler for V1.

### 3.2. Race Conditions
*   **Problem**: A user reads their portfolio (triggering snapshot creation) while simultaneously adding a backdated transaction.
*   **Solution**:
    *   Snapshot creation should happen in a transaction.
    *   Before committing the new snapshot, verify no new transactions have appeared in the window `[replayStartDate, snapshotTimestamp]` that were not included in the replay.
    *   However, since we use "Invalidation on Write", even if a race occurs, the "Write" side will delete the just-created snapshot correctly.

## 4. Performance Goals

| Metric | Current Baseline | Target (with Optimization) |
| :--- | :--- | :--- |
| **Replay Complexity** | O(Total History) | O(Recent History) |
| **P99 Latency (Heavy User)** | > 2000ms | < 200ms |
| **DB Reads (Rows)** | All Transactions | 1 Snapshot + < 100 Transactions |

*   **Latency**: The primary goal is to decouple read latency from account age. A 10-year user should see the same performance as a 1-week user.
*   **Storage**: Snapshot storage cost is negligible compared to the CPU/Latency savings. We will retain snapshots indefinitely for now, but can implement a retention policy (e.g., "Keep 1 per month for years > 1") later.

## Notes

# High-Performance Systems: Lazy Read-Repair

In the context of high-performance systems (like a portfolio tracker), **Lazy Read-Repair** is a strategy where you defer the "expensive" work of updating or fixing data until the moment someone actually tries to read it.

Instead of keeping your snapshots perfectly up-to-date every time a transaction occurs (which would slow down the write path), you wait for a user to request their portfolio and check if the data is "stale" or "messy."

---

## How it works in your specific context
Imagine a user has 500 transactions. Calculating their portfolio balance from scratch every time they log in is slow. To fix this, you use **Snapshots** (cached totals at a specific point in time).

### 1. The "Lazy" Detection
When `GetPortfolioSummary` is called, the system performs a quick health check on the data:

* **Is it too fragmented?** (e.g., "Are there more than 100 new transactions since the last snapshot?")
* **Is it too old?** (e.g., "Was the last snapshot taken more than 30 days ago?")

> [!IMPORTANT]
> If either condition is true, the system realizes the data needs a "repair" (a new snapshot).

### 2. The "Async" Repair (Non-Blocking)
The "Repair" part is the generation of a new snapshot. In a Lazy approach, you don't want the user to wait while you crunch those 500 transactions.

1.  **Immediate Response:** You serve the user the current data (even if it takes a second to calculate manually this one time).
2.  **Background Trigger:** You "fire and forget" a job to a background worker. This worker calculates the new snapshot and saves it to the database quietly.

**The Result:** The next time the user visits, the data is already repaired and loads instantly.

---

## Why use this instead of Eager Updates?



| Feature | Eager Update (Write-Through) | Lazy Read-Repair (Your Context) |
| :--- | :--- | :--- |
| **Write Latency** | High (Every transaction must update a snapshot). | Low (Transactions are just appended). |
| **Resource Usage** | High (Updates snapshots for inactive users). | Optimized (Only updates active users). |
| **Data Freshness** | Always perfect. | Eventually consistent (slightly delayed). |
| **Complexity** | Simple, but heavy on DB. | Slightly higher (requires background workers). |


## Summary of the "Backfill"

The **Async Backfill** is the proactive sibling to **Lazy Read-Repair**. While Lazy Repair waits for a user to show up, the Backfill worker looks for "heavy" users (those with many transactions) and generates snapshots for them during off-peak hours so they don't experience a slow first load.

---

### Proactive vs. Reactive
| Strategy | Trigger | Purpose |
| :--- | :--- | :--- |
| **Lazy Read-Repair** | User Activity (Read) | Reactive maintenance to keep active data fresh. |
| **Async Backfill** | System Scheduled (Cron/Batch) | Proactive optimization for high-volume users. |

> [!TIP]
> Combining both strategies ensures that "heavy" users always have a fast experience, while the system doesn't waste resources on inactive accounts.
