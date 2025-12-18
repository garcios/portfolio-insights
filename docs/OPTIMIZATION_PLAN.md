# Technical Optimization Document: GetPortfolioSummary

## 1. Executive Summary
The `GetPortfolioSummary` method is currently suffering from high latency due to **O(N) sequential network calls** in the critical path, **redundant data fetching**, and **full transaction history replay**. This document outlines a plan to reduce latency by implementation of **batch processing**, **parallel concurrency**, **multi-level caching**, and **incremental aggregation**.

## 2. Current State Analysis & Bottlenecks

### 2.1. N+1 Network Calls (Critical)
**Location:** `calculatePeriodGains` (Lines 342-493) and `getCurrentHoldingsSummary` (Lines 616-678).
**Issue:**
- Inside the transaction replay loop, `getAmountAndRate` calls `uc.marketDataGateway.GetCurrencyRateOnDate` for *every* transaction where currency differs.
- If a user has 1,000 transactions, this results in potentially 1,000 sequential RPC calls to the Market Data Service.
- Similarly, `getCurrentHoldingsSummary` iterates holdings and calls `GetPriceOnDate` and `GetCurrencyRateOnDate` sequentially for each asset.

### 2.2. Full History Replay (Scalability)
**Location:** `calculatePeriodGains`
**Issue:**
- The logic fetches *ALL* transactions for a user (`PageToken` loop) and replays them from scratch to determine the current state.
- It does *not* utilize existing `PortfolioSnapshot`s to initialize the state.
- **Impact:** Performance degrades linearly with the age of the account and number of transactions.

### 2.3. Redundant Data Fetching
**Location:** `GetPortfolioSummary` vs `getCurrentHoldingsSummary`
**Issue:**
- `GetPortfolioSummary` fetches current prices (batch) effectively in Step 3a.
- However, it sets `summary.TotalValue` by calling `getCurrentHoldingsSummary` (Line 276), which *ignores* the prices just fetched and re-fetches them one-by-one.

## 3. Proposed Optimizations

### 3.1. Elimination of N+1 Queries (Batching & Prefetching)
**Strategy:**
Instead of fetching rates/prices inside loops, fetch distinct required data in parallel batches.

**Implementation:**
1.  **Transactions:** Collect all unique `(Currency, Date)` tuples required for the replay.
2.  **Market Data:** Add a `GetHistoricalRates(ctx, requests)` bulk endpoint to Market Data Service (or Parallel Fetch in Portfolio Service).
3.  **Holdings:** Utilize the `marketQuotes` and `fxRates` maps already populated in `GetPortfolioSummary` for the final calculation instead of calling `getCurrentHoldingsSummary`.

**Expected Impact:** Major. Reduces network round-trips from `O(Txn)` to `O(1)` (or `O(Days)`).

### 3.2. Concurrency with `errgroup`
**Strategy:**
Use `golang.org/x/sync/errgroup` to parallelize independent fetch operations.

**Implementation:**
```go
g, ctx := errgroup.WithContext(ctx)

// 1. Fetch Transactions
g.Go(func() error {
    // Fetch and sort transactions
    return nil
})

// 2. Fetch Current Market Data (if isCurrent)
g.Go(func() error {
    // Fetch current prices for user's holdings
    return nil
})

if err := g.Wait(); err != nil {
    return nil, err
}
```

**Expected Impact:** Moderate to High. Reduces total latency to the slowest dependency rather than the sum of all.

### 3.3. Caching Strategy (Redis + In-Memory)
**Strategy:**
Market data is largely static (historical prices never change).

**Recommendations:**
1.  **L1 Request-Scoped Cache:** Use a simple `map[string]float64` within the request context to avoid re-fetching the same USD/AUD rate for the same date multiple times during replay.
2.  **L2 App Cache (Redis):** Cache outcomes of `GetPriceOnDate` and `GetCurrencyRateOnDate`.
    - **Key:** `market:price:{symbol}:{date}` or `market:fx:{base}:{target}:{date}`
    - **TTL:** Indefinite (for historical dates).

**Expected Impact:** High. Historical replays will hit cache >90% of the time.

### 3.4. Database & Logic Tuning (Incremental Aggregation)
**Strategy:**
Avoid replaying the entire history `calculatePeriodGains`.

**Implementation:**
1.  **Snapshots:** Retrieve the latest `PortfolioSnapshot` *before* `startDate`.
2.  **Filter:** Only fetch transactions occurring *after* the snapshot timestamp.
3.  **Replay:** Initialize `ReplayResult` state (positions, cost basis) from the snapshot, then replay only the new transactions.

**Expected Impact:** High. Caps the computation time to "time since last snapshot" rather than "time since inception".

## 4. Implementation Plan Summary

1.  **Refactor Replay Logic:** Introduce `RequestCache` struct to memoize FX calls within the loop.
2.  **Add Batch Methods:** Update `MarketDataGateway` to support `GetHistoricalPrices(map[string]time.Time)`.
3.  **Switch to Incremental Replay:** Update `calculatePeriodGains` to look for the nearest previous snapshot.
4.  **Parallelize:** Wrap independent fetches in `errgroup`.

## 5. Conclusion
Implementing these changes will likely reduce the P99 latency of `GetPortfolioSummary` by **10-50x** for users with significant transaction history, transforming it from a linear `O(N)` operation to a near-constant time operation dependent mainly on the "days since last login".
