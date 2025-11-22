# Portfolio Insights - MVP Development Plan
**Goal**: Launch MVP in 3 Months (12 Weeks)  
**Role**: Solo Developer  
**Current Status**: Foundation laid (Monorepo, User Service, DB Schemas, Frontend Skeleton)

---

## 📅 Phase 1: Backend Core & Data Ingestion (Weeks 1-4)
**Focus**: getting the microservices functional and flowing real data.

### Week 1: Authentication & User Service Polish (✅ Partially Done)
- [x] **User Service**: Basic CRUD & DB Connection.
- [ ] **Auth Implementation**: Implement JWT (JSON Web Tokens) generation in `user-service`.
- [ ] **Gateway Auth**: Implement auth middleware in `apps/gateway` to validate JWTs.
- [ ] **Frontend Auth**: Create Login/Signup pages and store JWT in local storage/cookies.

### Week 2: Market Data Service (The Data Engine)
- [ ] **External API Integration**: Choose a provider (Yahoo Finance, AlphaVantage, or CoinGecko) and implement a client in `marketdata-service`.
- [x] **Data Ingestion**: Create a background worker (Cron job or Go routine) to fetch and update asset prices daily.
- [x] **gRPC Endpoints**: Implement `GetAsset`, `GetLatestPrice`, `GetHistoricalPrices`.
- [x] **DB Integration**: Store data in `marketdata.assets` and `marketdata.asset_prices`.

### Week 3: Transaction Service (The Input)
- [x] **Core Logic**: Implement `Buy`, `Sell` logic.
- [x] **Validation**: Ensure transactions reference valid Assets (via gRPC call to MarketData) and Users.
- [x] **DB Integration**: Store in `txn.transactions`.
- [x] **API**: Expose `CreateTransaction`, `ListTransactions`, `GetTransaction`.

### Week 4: Portfolio Service (The Aggregator)
- [ ] **Calculation Engine**: Implement logic to calculate current holdings based on transaction history.
- [ ] **Value Calculation**: Fetch latest prices from `marketdata-service` to calculate total portfolio value.
- [ ] **History Tracking**: Implement snapshots of portfolio value over time for the performance chart.
- [ ] **DB Integration**: Store in `investments.holdings` and `investments.portfolio_history`.

---

## 🔗 Phase 2: Integration & Frontend Connectivity (Weeks 5-8)
**Focus**: Connecting the frontend to the backend via GraphQL.

### Week 5: GraphQL Gateway & Federation
- [ ] **Schema Stitching**: Map gRPC services (Market, Transaction, Portfolio) to the GraphQL Schema in `apps/gateway`.
- [ ] **Resolvers**: Implement resolvers to call the gRPC clients.
- [ ] **Data Loaders**: Implement DataLoaders to solve the N+1 problem (efficiently fetching nested data).

### Week 6: Dashboard Integration
- [ ] **Real Data**: Replace Frontend mock data with Apollo Client queries.
- [ ] **Portfolio Chart**: Connect the Area Chart to the `PortfolioService` history endpoint.
- [ ] **Stats Cards**: Wire up "Total Value", "Day Change" to real calculations.

### Week 7: Transaction Management UI
- [ ] **Add Transaction Modal**: Create a form to search for assets and add Buy/Sell records.
- [ ] **Holdings Table**: Connect the table to real holdings data.
- [ ] **Asset Search**: Implement an autocomplete search bar for finding stocks/crypto.

### Week 8: User Experience & Polish
- [ ] **Loading States**: Add skeletons and spinners for data fetching.
- [ ] **Error Handling**: Graceful error messages (e.g., "API Down", "Invalid Password").
- [ ] **Responsive Design**: Ensure dashboard looks good on mobile.

---

## 🚀 Phase 3: Advanced Features & Optimization (Weeks 9-10)
**Focus**: Adding value and ensuring stability.

### Week 9: Analytics & Insights
- [ ] **Allocation Charts**: Add a Pie chart for Asset Allocation (Stocks vs Crypto vs Cash).
- [ ] **Performance Metrics**: Calculate "All Time High", "Best Performer", "Worst Performer".
- [ ] **Currency Support**: (Optional) Allow viewing portfolio in USD/EUR/GBP.

### Week 10: Infrastructure & Observability
- [ ] **Observability Stack**: Deploy Prometheus and Grafana containers via Docker Compose.
- [ ] **Instrumentation**: Add Prometheus metrics middleware to all Go microservices (Request latency, Error rates, Active goroutines).
- [ ] **Dashboards**: Create Grafana dashboards to visualize service health, system metrics (CPU/Memory), and business metrics (e.g., New Users, Total Transactions).
- [ ] **CI/CD**: Setup GitHub Actions for automated testing and linting.
- [ ] **Docker Optimization**: Multi-stage builds for smaller images.
- [ ] **Environment Config**: Ensure strict separation of Dev/Prod configs.

---

## 🏁 Phase 4: Testing & Launch (Weeks 11-12)
**Focus**: Quality assurance and deployment.

### Week 11: Testing & Bug Fixes
- [ ] **E2E Testing**: Use Cypress or Playwright to test critical flows (Signup -> Add Transaction -> View Dashboard).
- [ ] **Integration Tests**: Verify microservices communicate correctly.
- [ ] **Bug Bash**: Dedicate a week to fixing all known issues.

### Week 12: Deployment (MVP Launch)
- [ ] **Cloud Provider**: Select provider (AWS, DigitalOcean, or Render).
- [ ] **Database**: Provision a managed PostgreSQL instance.
- [ ] **Deploy**: Deploy services via Docker Compose or Kubernetes (if ambitious).
- [ ] **Launch**: Share with initial users!

---

## 🔑 Key Milestones

1.  **Milestone 1 (Week 4)**: All Backend Services running and communicating locally.
2.  **Milestone 2 (Week 6)**: Frontend displaying real data from at least one service.
3.  **Milestone 3 (Week 8)**: Full "Write" capability (Users can sign up and add transactions).
4.  **Milestone 4 (Week 12)**: Production Deployment.
