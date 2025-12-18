# Technical Design: Optimization of GetCurrentPrices

**Author**: Antigravity  
**Date**: 2025-12-18  
**Status**: DRAFT  

## 1. Problem Statement
The `GetPortfolioSummary` endpoint is experiencing extreme latency, specifically traced to the `GetCurrentPrices` step. A recent profiling log captured a duration of **134,585ms (approx. 2.2 minutes)** for a single call.

**Log Evidence:**
```text
[GetPortfolioSummary-1766031991733390000] Step: GetCurrentPrices | Duration: 134585ms | Call Count: 1
```

This latency renders the portfolio summary unusable for users with large portfolios or during cold-cache scenarios. The issue appears to be server-side processing time within the `marketdata-service` or its underlying database interactions.

## 2. Root Cause Analysis
Based on code analysis of the `marketdata-service` repository (`postgres_repo.go`), the `GetLatestPrices` method uses a **correlated subquery** inside a `WHERE IN` clause to fetch the latest price for each symbol.

**Current SQL Implementation:**
```sql
SELECT ...
FROM marketdata.asset_prices p
JOIN marketdata.assets a ON p.asset_id = a.id
WHERE a.symbol IN ($1, $2, ...)
AND p.id IN (
    SELECT p2.id
    FROM marketdata.asset_prices p2
    JOIN marketdata.assets a2 ON p2.asset_id = a2.id
    WHERE a2.symbol = a.symbol -- Correlated reference
    ORDER BY p2.timestamp DESC
    LIMIT 1
)
```

**Deficiencies:**
1.  **N+1 Complexity**: For every symbol in the input list, the database executes the inner subquery. If `asset_prices` is large (likely millions of rows), this results in minimal index usage efficiency and high I/O.
2.  **Join Overhead**: The inner query joins `asset_prices` to `assets` again, doubling the work for every row check.
3.  **Lack of Server-Side Caching**: While the *client* (Portfolio Service) has a cache, the *server* (Market Data Service) hits the database directly for every request. A cold client cache triggers a massive DB load.

## 3. Proposed Solutions

### Strategy A: Database Query Optimization (Recommended)
Replace the correlated subquery with a Postgres-specific `DISTINCT ON` or a `LATERAL JOIN`. This is the impactful low-hanging fruit.

**Proposed SQL:**
```sql
SELECT DISTINCT ON (a.symbol) 
    a.symbol, p.id, p.price, p.timestamp
FROM marketdata.asset_prices p
JOIN marketdata.assets a ON p.asset_id = a.id
WHERE a.symbol = ANY($1)
ORDER BY a.symbol, p.timestamp DESC;
```

*   **Pros**:
    *   Drastically reduces complexity from O(N*M) to O(N log M) or better (depending on indexes).
    *   Native Postgres optimization; no infrastructure changes.
    *   Immediate impact on latency.
*   **Cons**:
    *   Requires migration/validation of SQL logic.
    *   `DISTINCT ON` is non-standard SQL (Postgres specific), slightly reducing portability (low concern).

### Strategy B: Server-Side Redis Caching
Implement a shared Redis cache layer within the `marketdata-service`.

*   **Pros**:
    *   Protects the database from read storms.
    *   Serves repeated requests for popular assets (e.g., BTC, AAPL) in sub-millisecond time.
*   **Cons**:
    *   Introduces infrastructure complexity (Redis dependency).
    *   Cache invalidation challenges (need to invalidate on new price ingestion).

### Strategy C: Database Partitioning & Read Replicas
Partition the `asset_prices` table by time (e.g., monthly) and offload reads to a replica.

*   **Pros**:
    *   Keeps the "active" partition small, speeding up index scans.
    *   Scales read throughput horizontally.
*   **Cons**:
    *   High operational cost and complexity.
    *   Overkill if the query itself is inefficient (Strategy A should be done first).

## 4. Implementation Plan

ref: Task `OPTIMIZATION-GET-CURRENT-PRICES`

1.  **Benchmark Current Query**: Run `EXPLAIN ANALYZE` on the existing query with a batch of 100 symbols to establish a baseline cost.
2.  **Refactor Repository**: Modify `GetLatestPrices` in `postgres_repo.go` to use `SELECT DISTINCT ON`.
3.  **Verify Indices**: Ensure `asset_prices` has a composite index on `(asset_id, timestamp DESC)` to support the optimized query.
4.  **Load Test**: Re-run the profiling scenario to verify latency drops from ~134s to <1s.
