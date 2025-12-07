# Portfolio Returns Computation Strategy

This document outlines the technical strategy for implementing Internal Rate of Return (IRR) / Money-Weighted Rate of Return (MWRR) and Time-Weighted Rate of Return (TWRR) calculations within the `portfolio-service`.

## 1. Data Requirements

To accurately calculate IRR and TWRR, the following data points are required:

### 1.1. Transactions (Cash Flows)
From `transaction-service`, we require the full history of transactions for the user.
*   **Source**: `TransactionRepository.ListByUserID` (via gRPC or events).
*   **Fields Needed**:
    *   `ExecutedAt`: The exact date/time of the cash flow.
    *   `Type`: To determine direction (Inflow vs Outflow).
        *   `DEPOSIT` (or implied via `BUY` if funding from outside): Negative flow for XIRR (Investment).
        *   `WITHDRAWAL` (or implied via `SELL` if withdrawing to outside): Positive flow for XIRR (Return).
    *   `Amount`: The total value of the transaction (Price * Quantity + Fees).
        *   *Note*: For TWRR, we strictly need "External Cash Flows". Rebalancing (Sell Stock A to Buy Stock B) is internal and should not trigger a sub-period split unless it involves external capital.

### 1.2. Portfolio Valuations (Market Values)
We need the total value of the portfolio at specific points in time.
*   **Source**: `MarketDataService` (historical prices) + `HoldingRepository` (snapshot reconstruction).
*   **Required Points**:
    *   **Start Date**: Value at the beginning of the period ($V_{start}$).
    *   **End Date**: current value ($V_{end}$).
    *   **Cash Flow Dates**: For TWRR, we need the Portfolio Market Value immediately *before* each external cash flow occurs ($V_{t}$).

### 1.3. Market Data
*   **Source**: `MarketDataRepository`.
*   **Fields**: `Price` for each asset held on the required dates.

---

## 2. IRR / MWRR Methodology

The Money-Weighted Rate of Return is equivalent to the Internal Rate of Return (IRR). We will use the **XIRR** algorithm to account for irregular time periods between cash flows.

### 2.1. Formula
XIRR is the rate $r$ that satisfies:

$$ \sum_{i=0}^{N} \frac{C_i}{(1+r)^{\frac{d_i - d_0}{365}}} = 0 $$

Where:
*   $C_i$: Cash flow amount at index $i$.
*   $d_i$: Date of cash flow $i$.
*   $d_0$: Start date.

### 2.2. Input Array Structure
The input to the XIRR algorithm will be constructed as follows:

| Order | Date ($d_i$) | Standard XIRR Sign Convention ($C_i$) | Description |
| :--- | :--- | :--- | :--- |
| 1 | $t_{start}$ | Negative (-) | Initial Portfolio Value (treat as initial investment) |
| 2..N | $t_{flow}$ | Negative (-) | **Deposits** (External Inflows) |
| 2..N | $t_{flow}$ | Positive (+) | **Withdrawals** (External Outflows) |
| Last | $t_{end}$ | Positive (+) | **Current Portfolio Value** (Terminal Value) |

### 2.3. Algorithm Selection
Since an analytical solution does not exist for XIRR, we will use a numerical method:
*   **Newton-Raphson Method**: Fast convergence.
*   **Bisection Method**: Fallback if Newton-Raphson fails to converge.

---

## 3. TWRR Methodology

The Time-Weighted Rate of Return eliminates the distorting effects of inflows and outflows of money. It measures the compound rate of growth of $1 over the period.

### 3.1. Process (Modified Dietz or True Time-Weighted)
Given the potential for high frequency data, we will aim for **True TWRR** by valuing the portfolio at every external cash flow.

1.  **Partitioning**: Break the total period into sub-periods determined by the dates of external cash flows.
    *   Period 1: From Start to $t_{flow1}$
    *   Period 2: From $t_{flow1}$ to $t_{flow2}$
    *   ...
    *   Period N: From $t_{flowN}$ to End.

2.  **Valuations**:
    *   Calculate $MV_{start}$ (Market Value at start of sub-period).
    *   Calculate $MV_{end}$ (Market Value at end of sub-period, *before* the cash flow).
    *   Identify $CF$ (The external cash flow occurring at the split point).

3.  **Sub-Period Return Calculation**:
    $$ r_n = \frac{MV_{end} - MV_{start}}{MV_{start}} $$
    *(Note: If the Cash Flow occurs at the start of the next period, it is not included in calculating the return of the previous period).*

4.  **Chaining (Geometric Linking)**:
    $$ TWRR = \left[ (1 + r_1) \times (1 + r_2) \times ... \times (1 + r_n) \right] - 1 $$

---

## 4. Implementation Plan & Code Snippets

### 4.1. Domain Models
**File**: `services/portfolio-service/internal/domain/performance.go` (New File)

We will introduce a new domain entity to hold the calculation results.

```go
package domain

import "time"

type PortfolioPerformance struct {
    UserID      string
    Period      string // "YTD", "1Y", "ALL"
    TWRR        float64
    MWRR        float64 // IRR
    CalcDate    time.Time
}

type CashFlow struct {
    Date   time.Time
    Amount float64
    Type   string // "INFLOW" (Deposit), "OUTFLOW" (Withdrawal)
}
```

### 4.2. Calculation Service / Usecase Logic
**File**: `services/portfolio-service/internal/usecase/performance_calculator.go` (New File)

This service will encapsulate the complex logic of reconstructing portfolio state and running the math.

```go
package usecase

import (
    "math"
    "time"
    "github.com/garcios/portfolio-insights/services/portfolio-service/internal/domain"
)

// CalculateIRR computes the Money-Weighted Rate of Return using Newton-Raphson
func CalculateIRR(flows []domain.CashFlow, currentValue float64, currentDate time.Time) (float64, error) {
    // 1. Construct standard input array: 
    //    - Flows (Deposits = negative, Withdrawals = positive)
    //    - Terminal Value = positive at currentDate
    
    // 2. Define function f(r) = sum( C_i / (1+r)^((d_i-d_0)/365) )
    // 3. Define derivative f'(r)
    // 4. Iterate r_new = r_old - f(r)/f'(r) until |r_new - r_old| < tolerance
    
    return rate, nil
}

// CalculateTWRR computes Time-Weighted return by geometric linking
func CalculateTWRR(
    startValue float64, 
    flows []domain.CashFlow, 
    valuationsAtFlows map[time.Time]float64, 
    currentValue float64,
) float64 {
    // 1. Sort flows by date
    // 2. Iterate through flows:
    //      r_sub = (ValuationCheckPoint - PeriodStartValue) / PeriodStartValue
    //      chainProduct *= (1 + r_sub)
    //      PeriodStartValue = ValuationCheckPoint + FlowAmount
    
    // 3. Final period: (CurrentValue - PeriodStartValue) / PeriodStartValue
    // 4. chainProduct *= (1 + r_final)
    
    return chainProduct - 1
}
```

### 4.3. Integration
**File**: `services/portfolio-service/internal/usecase/portfolio_usecase.go`

Update the `PortfolioUsecase` interface and implementation to include performance retrieval.

```go
type PortfolioUsecase interface {
    // ... existing methods ...
    GetPortfolioPerformance(ctx context.Context, userID string, period string) (*domain.PortfolioPerformance, error)
}

// Implementation
func (u *portfolioUsecase) GetPortfolioPerformance(ctx context.Context, userID string, period string) (*domain.PortfolioPerformance, error) {
    // 1. Fetch all transactions for user
    // 2. Fetch current portfolio value
    // 3. If Period != "ALL", fetch value at period start (need historical snapshot)
    
    // 4. Identify External Cash Flows (Deposits/Withdrawals)
    //    TODO: Define logic to distinguish Rebalancing vs External Flows based on Transaction Type
    
    // 5. For TWRR: Fetch portfolio value at each cash flow date
    //    (Requires looping through flow dates and querying MarketDataService for historical prices of holding snapshot at that time)

    // 6. Call CalculateIRR(...)
    // 7. Call CalculateTWRR(...)
    
    return &domain.PortfolioPerformance{...}, nil
}
```

---

## 5. Gap Analysis & Remediation Strategy

Based on the analysis of the data requirements and current system state, the following gaps have been identified along with their proposed remediation.

### 5.1. Identify External vs. Internal Cash Flows
**The Gap**: TWRR requires partitioning performance periods *only* on external cash flows (Deposits/Withdrawals). Currently, the `Transaction` struct lacks distinguishing between `DEPOSIT` (external) and `BUY` (potentially internal rebalancing).
**Remedy**:
*   **Transaction Types**: Update `transaction-service` to explicitly support `DEPOSIT` and `WITHDRAWAL` in the `TransactionType` enum.
*   **Logic Update**: Treat `BUY` and `SELL` as internal portfolio actions.

### 5.2. Missing "Cash" Asset Handling
**The Gap**: If a user sells a stock, the visible "Portfolio Value" drops because the resulting cash is not tracked as a holding. This distorts TWRR as a false withdrawal.
**Remedy**:
*   **Model Cash as a Holding**: Introduce a `CASH` or `USD` asset in `portfolio-service`.
*   **Flow Logic**: 
    *   `SELL` -> Increase `CASH` holding.
    *   `BUY` -> Decrease `CASH` holding.
    *   `DEPOSIT` -> Increase `CASH` (External Inflow).
    *   `WITHDRAWAL` -> Decrease `CASH` (External Outflow).

### 5.3. Point-in-Time Portfolio Valuation ("Time Travel")
**The Gap**: True TWRR requires the total portfolio value at the *exact moment* of an external cash flow. Daily snapshots are insufficient for intraday accuracy or past dates without snapshots.
**Remedy**:
*   **Reconstruction Method**: Implement `ReconstructPortfolioState(date)` in `portfolio-service`.
*   **Logic**: Replay all transactions up to `date` to determine exact asset quantities, then query `marketdata-service` for prices at that specific date.

### 5.4. Historical Price Availability
**The Gap**: Prices must be available for every asset on every cash flow date. Missing prices (weekends/holidays) will cause calculation errors.
**Remedy**:
*   **Fill-Forward Pricing**: Update `marketdata-service` to support "Nearest Previous Price" retrieval. If a price is missing for a requested date, return the most recent valid price.
