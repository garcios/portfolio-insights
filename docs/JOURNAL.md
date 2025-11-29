# Portfolio Insights Development Journal

A chronological record of development progress, features implemented, and technical decisions made during the development of the Portfolio Insights application.

---

## 2025-11-29 - Frontend Implementation & UI Polish

### Overview
Focused on building out the core frontend pages, improving navigation, and establishing a robust testing foundation. Implemented the Transactions page, Authentication UI, and a responsive Header component, backed by comprehensive unit tests.

### Features Implemented

#### 1. Responsive Navigation & Header
**Status:** ✅ Complete

**Components:**
- **Header:** Sticky top navigation with glassmorphism effect.
- **Mobile Menu:** Slide-over menu for small screens with smooth animations.
- **User Menu:** Dropdown with avatar, settings, and theme toggle (UI).
- **Routing:** Client-side routing using `react-router-dom` v6.

**Key Features:**
- Responsive design adapting to mobile/desktop.
- Active state highlighting for navigation links.
- Accessible ARIA attributes for menu interactions.
- "Click outside" detection for closing menus.

#### 2. Transactions Management UI
**Status:** ✅ Complete (Mock Data)

**Components:**
- **Page:** `TransactionsPage.tsx`
- **Table:** `TransactionsTable.tsx` with sorting and filtering.
- **Modal:** `AddTransactionModal.tsx` for creating new entries.

**Features:**
- **List View:** Display transactions with formatted dates, currency, and type-based coloring.
- **Sorting:** Sort by date, ticker, type, quantity, price, or total.
- **Filtering:** Real-time search by ticker and filter by transaction type (Buy/Sell/Split/Dividend).
- **Creation:** Modal form with auto-calculation of totals (Quantity * Price + Brokerage).
- **Validation:** Client-side validation for required fields and numeric values.

#### 3. Authentication UI
**Status:** ✅ Complete (UI Only)

**Components:**
- **Page:** `AuthPage.tsx`
- **Component:** `Input.tsx` (Reusable UI component)

**Features:**
- **Dual Mode:** Tabbed interface for Login and Register.
- **Validation:** Email format, password length, and password matching validation.
- **Feedback:** Loading states and error messaging.
- **Design:** Modern, clean aesthetic consistent with the rest of the app.

#### 4. Frontend Testing Suite
**Status:** ✅ Complete

**Coverage:**
- **Header:** 100% coverage
- **MobileMenu:** 100% coverage
- **StatsCard:** 100% coverage
- **LoadingSpinner:** 100% coverage
- **UserMenu:** ~95% coverage
- **NavLink:** ~71% coverage

**Tooling:**
- **Framework:** Vitest + React Testing Library
- **Features Tested:** Rendering, user interactions (clicks, keyboard), accessibility attributes, and conditional styling.

### Technical Decisions

#### 1. Client-Side Routing
**Decision:** Migrated from conditional rendering to `react-router-dom`.

**Rationale:**
- Enables deep linking (e.g., `/transactions`).
- Supports browser history navigation (back/forward).
- Standard practice for SPA (Single Page Applications).

#### 2. Mock Data Strategy
**Decision:** Used local mock data for Transactions and Auth.

**Rationale:**
- Allowed rapid UI development without waiting for backend endpoints.
- Decoupled frontend progress from backend readiness.
- Easy to swap with Apollo Client `useQuery` hooks later.

#### 3. Component-First Testing
**Decision:** Wrote tests for individual UI components before page integration.

**Rationale:**
- Ensures building blocks are solid.
- easier to debug isolated components.
- Provides immediate feedback on accessibility compliance.

### Next Steps
1. **Backend Integration:** Connect Transactions and Auth pages to real GraphQL endpoints.
2. **State Management:** Implement global auth state (Context or Redux) to handle login sessions.
3. **E2E Testing:** Add Cypress or Playwright tests for full user flows.

---

## 2025-11-28 - Portfolio Performance Feature Complete

### Overview
Completed end-to-end implementation of portfolio performance tracking and visualization, including historical data backfilling, GraphQL API, and frontend integration.

### Features Implemented

#### 1. Portfolio History Backfilling System
**Status:** ✅ Complete

**Components:**
- **Admin Endpoint:** `BackfillHistory` RPC in portfolio-service
  - Location: `services/portfolio-service/internal/handler/grpc/portfolio_handler.go`
  - Authentication: Admin token via environment variable `ADMIN_TOKEN`
  - Features:
    - Backfill for single user or all users
    - Date range selection (start_date, end_date)
    - Dry-run mode for previewing operations
    - Automatic duplicate prevention
    - Skip snapshots with zero total value

**Database Schema:**
```sql
CREATE TABLE investments.portfolio_history (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    total_value DECIMAL(20, 2) NOT NULL,
    total_cost_basis DECIMAL(20, 2) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, timestamp)
);
```

**Key Implementation Details:**
- Uses historical prices from `marketdata-service`
- Uses historical currency rates for multi-currency portfolios
- Falls back to previous 6 days if exact date has no price data
- Assumes current holdings for all historical dates

**Documentation:**
- Postman collection: `docs/postman/portfolio_backfill.postman_collection.json`
- README: `docs/postman/README_BACKFILL.md`
- Strategy document: `docs/PORTFOLIO_HISTORY_STRATEGY.md`

#### 2. Historical Currency Rates
**Status:** ✅ Complete

**Implementation:**
- Added `GetHistoricalCurrencyRates` RPC to marketdata-service
- Proto definition: `proto/marketdata/marketdata.proto`
- Repository method: `GetHistoricalCurrencyRates` in postgres_repo.go
- Database table: `marketdata.currency_rates`

**Features:**
- Query currency rates by date range
- Support for base/target currency pairs
- Integration with portfolio backfill for accurate historical valuations

#### 3. GraphQL Portfolio Performance API
**Status:** ✅ Complete

**Components:**
- **GraphQL Schema:** `apps/gateway/graph/schema.graphqls`
  - Query: `portfolioPerformance(userId: ID!, period: String!)`
  - Type: `PortfolioPerformancePoint { timestamp: String!, value: Float! }`
  
- **Resolver:** `apps/gateway/graph/schema.resolvers.go`
  - Calls portfolio-service gRPC endpoint
  - Transforms protobuf timestamps to ISO 8601 strings

- **Portfolio Service RPC:** `GetPortfolioPerformance`
  - Location: `services/portfolio-service/internal/handler/grpc/portfolio_handler.go`
  - Queries `portfolio_history` table via repository
  - Supports periods: 1d, 1w, 1m, 3m, 1y, all

**Documentation:**
- API Reference: `docs/GRAPHQL_API.md`
- Comprehensive Guide: `docs/PORTFOLIO_PERFORMANCE_API.md`

#### 4. Frontend Performance Chart
**Status:** ✅ Complete

**Implementation:**
- **File:** `apps/frontend/src/App.tsx`
- **Changes:**
  - Removed mock data generation
  - Added GraphQL query for portfolio performance
  - Implemented period selection (1D, 1W, 1M, 3M, 1Y, ALL)
  - Dynamic period description
  - Loading states and error handling
  - Auto-refresh every 60 seconds

**Features:**
- Real-time data from GraphQL API
- Interactive period selection buttons
- Responsive chart using Recharts library
- Gradient fill based on positive/negative performance
- Custom tooltips with formatted currency

#### 5. Historical Price Fallback Logic
**Status:** ✅ Complete

**Implementation:**
- **File:** `services/portfolio-service/internal/infrastructure/marketdata_gateway.go`
- **Methods:** `GetPriceOnDate`, `GetCurrencyRateOnDate`

**Features:**
- Attempts to find price on requested date
- Falls back to previous 6 days if no data found
- Prevents backfill failures due to weekends/holidays
- Detailed error messages for debugging

### Technical Decisions

#### 1. Historical Data Approach
**Decision:** Store daily snapshots in `portfolio_history` table

**Rationale:**
- Pre-computed values for fast queries
- Avoids complex historical calculations at query time
- Enables efficient time-series queries
- Supports multiple time periods without performance degradation

**Trade-offs:**
- Requires backfilling for historical data
- Storage overhead (minimal for daily snapshots)
- Need for scheduled jobs to maintain current data

#### 2. Backfill Strategy
**Decision:** Use current holdings with historical prices

**Rationale:**
- Simplifies implementation
- Provides reasonable approximation of historical value
- Avoids need to track historical holding changes

**Limitations:**
- Assumes holdings haven't changed (acceptable for initial implementation)
- Future enhancement: Track holding changes via transaction events

#### 3. Period Selection
**Decision:** Support predefined periods (1d, 1w, 1m, 3m, 1y, all)

**Rationale:**
- Covers common use cases
- Simplifies frontend UI
- Efficient database queries with indexed timestamps

**Future Enhancement:**
- Custom date range selection
- Comparison with market benchmarks

### Testing & Validation

#### Postman Collections
1. **GraphQL Gateway:** `docs/postman/graphql_gateway.postman_collection.json`
   - Portfolio performance queries
   - Variable period selection
   - Error handling examples

2. **Portfolio Backfill:** `docs/postman/portfolio_backfill.postman_collection.json`
   - Single user backfill
   - All users backfill
   - Dry-run mode
   - Date range examples

#### Build Verification
- ✅ Gateway builds successfully
- ✅ Portfolio-service builds successfully
- ✅ Frontend builds successfully
- ✅ All services start without errors

### Known Issues & Future Work

#### Current Limitations
1. **Empty Performance Data:** If no historical snapshots exist, returns empty array
   - **Solution:** Run backfill endpoint to populate historical data
   - **Future:** Implement scheduled daily snapshot job

2. **Test Mocks Missing:** MarketData service test mocks need `GetHistoricalCurrencyRates`
   - **Impact:** Test compilation errors (non-blocking)
   - **Priority:** Medium
   - **Files:**
     - `services/marketdata-service/internal/usecase/marketdata_usecase_test.go`
     - `services/marketdata-service/internal/handler/grpc/handler_test.go`

#### Planned Enhancements
1. **Automated Daily Snapshots**
   - Scheduled job to create daily portfolio snapshots
   - Cron job or background worker
   - Prevents manual backfill for current data

2. **Performance Metrics**
   - Add gain/loss percentage to each data point
   - Calculate returns (daily, weekly, monthly)
   - Benchmark comparison (S&P 500, etc.)

3. **Aggregated Data**
   - Weekly/monthly aggregations for long periods
   - Reduces data points for "all" period
   - Improves chart rendering performance

4. **Real-time Updates**
   - WebSocket support for live performance updates
   - Push notifications for significant changes

### Documentation Created

1. **API Documentation**
   - `docs/GRAPHQL_API.md` - Updated with portfolio performance query
   - `docs/PORTFOLIO_PERFORMANCE_API.md` - Comprehensive feature guide

2. **Admin Documentation**
   - `docs/postman/README_BACKFILL.md` - Backfill endpoint usage guide

3. **Strategy Documents**
   - `docs/PORTFOLIO_HISTORY_STRATEGY.md` - Historical tracking architecture

### Deployment Notes

#### Environment Variables
```bash
# Portfolio Service
ADMIN_TOKEN=secret  # Required for backfill endpoint
```

#### Database Migrations
All migrations applied successfully:
- `000001_create_users_table`
- `000002_create_transactions_table`
- `000003_create_market_data_tables`
- `000004_create_portfolio_tables`
- `000005_add_currency_to_holdings`
- `000006_create_currency_rates_table`
- `000007_add_unique_constraint_asset_prices`

#### Service Status
- ✅ Gateway (port 8080)
- ✅ Portfolio Service (port 50052)
- ✅ MarketData Service (port 50054)
- ✅ Transaction Service (port 50053)
- ✅ User Service (port 50051)
- ✅ Frontend (port 5173)

### Performance Considerations

#### Database Indexing
```sql
CREATE INDEX idx_portfolio_history_user_timestamp 
ON investments.portfolio_history(user_id, timestamp DESC);
```

**Impact:**
- Fast queries for user-specific performance data
- Efficient date range filtering
- Optimized for chronological ordering

#### Query Optimization
- Period filtering done at database level
- Minimal data transformation in application layer
- Efficient protobuf serialization

#### Caching Strategy
- Frontend: 60-second polling interval
- Consider Redis caching for frequently accessed periods
- Cache invalidation on new snapshots

---

## Previous Development Sessions

### Asset Price Integration
**Date:** 2025-11-27

**Achievements:**
- Created `tools/asset-prices/eodhd/asset_prices_eodhd.go`
- Support for stock prices and forex rates
- CSV output format for database ingestion
- Automatic detection of forex vs. stock data
- Rate limiting and error handling

**Documentation:**
- `tools/asset-prices/README.md`

### Price Ingestion Worker
**Date:** 2025-11-27

**Achievements:**
- Fixed date parsing in `price_ingestion.go`
- Support for YYYY-MM-DD format
- Resolved "ON CONFLICT" duplicate errors
- Batch insertion with 1000-row batches

### NATS Event-Driven Architecture
**Date:** 2025-11-22

**Achievements:**
- Transaction events published to NATS
- Portfolio service subscribes to transaction events
- Automatic holding updates on transactions
- Support for BUY, SELL, SPLIT transaction types

**Documentation:**
- `docs/NATS_ARCHITECTURE.md`
- `docs/NATS_EVENT_FLOW.md`
- `docs/NATS_IMPLEMENTATION_SUMMARY.md`

### GraphQL Gateway Implementation
**Date:** 2025-11-25

**Achievements:**
- Apollo Server setup with gqlgen
- Portfolio and holdings queries
- User management mutations
- CORS configuration
- Frontend integration

**Documentation:**
- `docs/GRAPHQL_API.md`
- `docs/GRAPHQL_TESTING_GUIDE.md`

### Monitoring Stack
**Date:** Earlier

**Achievements:**
- Prometheus metrics collection
- Grafana dashboards
- AlertManager configuration
- Service health monitoring

**Documentation:**
- `docs/MONITORING_IMPLEMENTATION.md`
- `docs/MONITORING_QUICKSTART.md`

---

## Development Metrics

### Code Statistics
- **Total Services:** 5 (gateway, user, portfolio, transaction, marketdata)
- **Total Endpoints:** 15+ gRPC RPCs, 5+ GraphQL queries
- **Database Tables:** 8 tables across 2 schemas
- **Documentation Files:** 20+ markdown files
- **Test Coverage:** In progress

### Technology Stack
- **Backend:** Go 1.21+
- **Frontend:** React 18 + TypeScript + Vite
- **API Gateway:** GraphQL (gqlgen)
- **RPC:** gRPC with Protocol Buffers
- **Database:** PostgreSQL 16
- **Caching:** Redis 7
- **Messaging:** NATS with JetStream
- **Storage:** MinIO (S3-compatible)
- **Monitoring:** Prometheus + Grafana
- **Containerization:** Podman/Docker Compose

---

## Next Steps

### Immediate Priorities
1. ✅ Complete portfolio performance feature
2. ⏳ Fix test mocks for GetHistoricalCurrencyRates
3. ⏳ Implement automated daily snapshot job
4. ⏳ Add performance metrics calculations
5. ⏳ Implement Transaction Search Page

### Short-term Goals
1. Implement custom date range selection
2. Add benchmark comparison feature
3. Create performance analytics dashboard
4. Implement data aggregation for long periods

### Long-term Vision
1. Real-time portfolio updates via WebSocket
2. Machine learning for portfolio insights
3. Tax reporting features
4. Multi-portfolio support
5. Social features (portfolio sharing, leaderboards)

---

## Lessons Learned

### What Went Well
1. **Modular Architecture:** Clean separation of concerns enables rapid feature development
2. **Repository Pattern:** Easy to swap implementations and test
3. **gRPC + GraphQL:** Best of both worlds - efficient internal communication, flexible client API
4. **Documentation-First:** Writing docs alongside code improves clarity

### Challenges Overcome
1. **Historical Data Complexity:** Solved with fallback logic for missing dates
2. **Multi-Currency Support:** Implemented historical currency rate tracking
3. **Performance Optimization:** Database indexing and efficient queries
4. **Frontend State Management:** Apollo Client handles caching and updates elegantly

### Areas for Improvement
1. **Test Coverage:** Need comprehensive unit and integration tests
2. **Error Handling:** More granular error types and better user feedback
3. **Logging:** Structured logging with correlation IDs
4. **Observability:** Distributed tracing across services

---

## Contributors

- Oscar Garcia (@garcios) - Lead Developer

---

## License

Portfolio Insights is a personal project for investment tracking and analysis.

---

*Last Updated: 2025-11-28 15:30 AEDT*
