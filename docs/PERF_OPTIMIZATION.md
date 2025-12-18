# Portfolio Service Performance Optimization

**Date:** December 16, 2025
**Scope:** `GetPortfolioSummary` Method Optimization

This document outlines the performance bottlenecks identified in the `portfolio-service` and the optimization strategies implemented to resolve them.

## 1. Currency Conversion in Transaction Replay

| Original Bottleneck | Optimization Strategy |
| :--- | :--- |
| **N+1 Network Calls:** The `calculatePeriodGains` method iterated through every transaction and made a sequential gRPC call (`GetCurrencyRateOnDate`) for each one where the currency differed from the reporting currency. This caused latency to grow linearly with the number of transactions ($O(N)$). | **Batch Pre-fetching:** Implemented a pre-fetch step that identifies all unique currencies and the full date range of the transaction history. Logic now calls `GetHistoricalCurrencyRates` once per currency to fetch all required rates in a single batch (or range) operation before the replay loop starts. Access during replay is now a fast map lookup ($O(1)$). |

## 2. Sequential Execution of Independent Operations

| Original Bottleneck | Optimization Strategy |
| :--- | :--- |
| **Sequential Processing:** The `GetPortfolioSummary` method executed three major heavy operations strictly one after another:<br>1. Transaction Replay (History)<br>2. Current Holdings Price Fetching<br>3. Previous Day's Summary (for Day Change)<br><br>The total latency was the sum of all three operations ($T_{total} = T_{replay} + T_{current} + T_{prev}$). | **Parallel Concurrency:** Utilized `errgroup` to execute these three independent tasks concurrently. The total latency is now effectively determined by the slowest single operation ($T_{total} \approx \max(T_{replay}, T_{current}, T_{prev})$). |

## 3. Repeated Fetching of Immutable Historical Data

| Original Bottleneck | Optimization Strategy |
| :--- | :--- |
| **Redundant API Calls:** Historical market prices and currency rates are immutable (they do not change once set). However, the service was repeatedly fetching this data from the `marketdata-service` on every request, wasting network bandwidth and increasing latency. | **Persistent Caching (Redis):** Implemented a `HistoricalCache` using Redis with an infinite (1-year) TTL. This cache stores the results of `GetPriceOnDate` and `GetCurrencyRateOnDate`. <br><br>The `MarketDataGateway` now checks this cache first. Subsequent requests for the same date/asset are instant (0ms network latency). |

## 4. Redundant Price Fetching

| Original Bottleneck | Optimization Strategy |
| :--- | :--- |
| **Double Fetching:** The service would often fetch current prices for holdings in one step, and then re-fetch them individually inside helper methods or for different calculations within the same request. | **Shared Data Context:** Refactored the data flow so that prices fetched during the "Current Holdings" parallel task are stored in a shared map (`marketQuotes`). This map is successfully passed down and reused for valuation and unrealized gain calculations, ensuring each asset price is fetched exactly once per request. |

## Impact Summary

- **Latency:** Drastically reduced, especially for users with large transaction histories or diverse currency exposure.
- **Throughput:** Improved service capacity by offloading redundant requests from the `marketdata-service`.
- **Resilience:** Reduced dependency on `marketdata-service` uptime for frequently accessed historical data.
