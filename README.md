Here is a comprehensive, production-ready directory structure for your microservices-based stock portfolio application. This structure is designed for a monorepo setup, ensuring code sharing, unified tooling, and streamlined CI/CD.

```
portfolio-insights/
├── .github/
│   └── workflows/
│       ├── ci-backend.yml           # Go services CI pipeline
│       ├── ci-frontend.yml          # React frontend CI pipeline
│       └── cd-deploy.yml            # Deployment pipeline
├── configs/                         # Centralized configuration templates
│   ├── gateway.yaml
│   ├── prometheus.yml
│   └── services/
│       ├── marketdata.dev.yaml
│       └── portfolio.prod.yaml
├── deployments/                     # Infrastructure as Code & Orchestration
│   ├── docker-compose.yml           # Local development orchestration
│   ├── k8s/                         # Kubernetes manifests
│   │   ├── base/
│   │   └── overlays/
│   │       ├── dev/
│   │       └── prod/
│   └── terraform/                   # Cloud infrastructure provisioning
├── docs/                            # Architecture decision records (ADRs) & API docs
│   ├── adr/
│   └── api/
├── gateway/                         # GraphQL Federation Gateway (Apollo/Yoga)
│   ├── src/
│   │   ├── config/
│   │   ├── data-sources/            # REST/gRPC data source classes
│   │   ├── directives/              # Custom schema directives (auth, etc.)
│   │   ├── plugins/                 # Logging, tracing plugins
│   │   ├── index.ts                 # Server entrypoint
│   │   └── types.ts
│   ├── Dockerfile
│   ├── package.json
│   ├── supergraph.yaml              # Federation config
│   └── tsconfig.json
├── pkg/                             # Shared Go Libraries (Internal)
│   ├── config/                      # Viper/Config loading
│   ├── database/                    # Postgres/Redis connection helpers
│   ├── errors/                      # Standardized error types
│   ├── logger/                      # Structured logging (Zap/Logrus)
│   ├── middleware/                  # Auth, tracing, recovery middleware
│   └── proto/                       # Shared gRPC Protobuf definitions
│       ├── marketdata/
│       │   └── marketdata.proto
│       └── portfolio/
│           └── portfolio.proto
├── scripts/                         # Dev automation & utilities
│   ├── generate-proto.sh
│   ├── init-db.sh
│   └── seed-data.go
├── services/                        # Backend Microservices (Go)
│   ├── marketdata-service/
│   │   ├── cmd/
│   │   │   └── server/
│   │   │       └── main.go          # Service entrypoint
│   │   ├── configs/                 # Service-specific config
│   │   ├── internal/                # Private application code (Clean Arch)
│   │   │   ├── adapters/            # Infrastructure implementations
│   │   │   │   ├── grpc/            # gRPC server implementation
│   │   │   │   ├── http/            # REST handlers (if needed)
│   │   │   │   └── thirdparty/      # External API clients (e.g., AlphaVantage)
│   │   │   ├── core/
│   │   │   │   ├── domain/          # Enterprise entities (Pure Go structs)
│   │   │   │   └── ports/           # Interfaces (Input/Output ports)
│   │   │   └── service/             # Application business logic (Use cases)
│   │   ├── migrations/              # SQL migration files
│   │   │   └── 000001_init_market.up.sql
│   │   ├── test/                    # Integration tests
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── Makefile
│   ├── portfolio-service/
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   ├── adapters/
│   │   │   │   └── repository/      # Postgres implementation of repo ports
│   │   │   ├── core/
│   │   │   │   ├── domain/          # e.g., Portfolio, Holding structs
│   │   │   │   └── ports/
│   │   │   └── service/             # e.g., CalculatePerformance logic
│   │   ├── migrations/
│   │   ├── Dockerfile
│   │   └── go.mod
│   ├── transaction-service/         # Handles buy/sell orders & history
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   └── ...
│   │   └── ...
│   └── user-service/                # Auth & User profiles
│       ├── cmd/server/main.go
│       ├── internal/
│       │   └── ...
│       └── ...
├── web/                             # Frontend (React + TypeScript)
│   ├── public/
│   ├── src/
│   │   ├── assets/
│   │   ├── components/              # Shared UI Components (Atoms/Molecules)
│   │   │   ├── Button/
│   │   │   ├── Card/
│   │   │   └── Layout/
│   │   ├── config/                  # Env vars & constants
│   │   ├── context/                 # Global state (Theme, Auth)
│   │   ├── features/                # Feature-based modules
│   │   │   ├── auth/
│   │   │   │   ├── api/             # Feature-specific API calls
│   │   │   │   ├── components/      # LoginForm, RegisterForm
│   │   │   │   └── hooks/           # useAuth
│   │   │   ├── dashboard/
│   │   │   │   ├── DashboardPage.tsx
│   │   │   │   └── widgets/
│   │   │   ├── market/
│   │   │   │   ├── components/      # StockChart, Ticker
│   │   │   │   └── types/
│   │   │   └── portfolio/
│   │   │       ├── components/      # HoldingsTable, AllocationPie
│   │   │       └── hooks/           # usePortfolioMetrics
│   │   ├── graphql/                 # Generated hooks & operations
│   │   │   ├── generated/
│   │   │   └── queries/
│   │   ├── hooks/                   # Shared hooks (useDebounce, etc.)
│   │   ├── lib/                     # Utils (Apollo Client setup, formatting)
│   │   ├── styles/                  # Global styles/Tailwind config
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── .eslintrc.json
│   ├── Dockerfile
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
├── .gitignore
├── Makefile                         # Root Makefile for monorepo commands
├── README.md
└── go.work                          # Go Workspace file for multi-module dev
```

Why this structure is Production-Ready

1. Scalability via Clean Architecture (Go Services):
 - By separating internal/core (domain logic) from internal/adapters (infrastructure), your business logic remains pure and testable. You can swap a PostgreSQL database for MongoDB or switch from REST to gRPC without touching the core business rules.
 - The ports layer defines strict interfaces, ensuring loose coupling between layers.
2. Feature-Based Frontend (web/src/features):
 - Instead of grouping by file type (all controllers together, all views together), grouping by feature (e.g., auth, portfolio) makes the codebase easier to navigate as it grows. A developer working on "Portfolio" only needs to look in one folder.
3. Monorepo Efficiency:
 - pkg/: Shared libraries prevent code duplication. If you improve the logger or fix a security issue in the JWT middleware, every service benefits immediately.
 - go.work: Enables seamless local development across multiple Go modules without complex replace directives.
4. Federated Gateway:
 - The gateway acts as a single entry point. It handles cross-cutting concerns like authentication, rate limiting, and request logging before requests ever reach your microservices. This simplifies the microservices themselves.
5. DevOps & Cloud Native:
 - Dedicated deployments/ and .github/ directories ensure that infrastructure and CI/CD are treated as first-class citizens, not afterthoughts. The structure supports both local Docker Compose development and production Kubernetes deployment out of the box.


Summary of Generated Components

1.  Monorepo Root:
- go.work: Configured for local multi-module development.
- Makefile: Commands to build, run, and test all services.
- deployments/docker-compose.yml: Orchestration for running the stack locally.

2. Backend (Go Microservices):
- pkg/: Shared libraries for logger (Zap) and config (Viper).
- services/marketdata-service/: A reference implementation following Clean Architecture:
    - internal/core/domain: Pure business entities (Stock).
    - internal/core/ports: Interfaces for repositories and services.
    - internal/service: Business logic implementation.
    - internal/adapters: Infrastructure code (gRPC/HTTP).
- services/{portfolio,transaction,user}-service/: Initialized modules ready for implementation.

3. Middleware (GraphQL Gateway):
- gateway/: Apollo Server setup with supergraph.yaml for federation.
- gateway/src/index.ts: Entry point for the federated graph.

4. Frontend (React + TypeScript):
- web/: Vite project structure.
- web/src/features/: Feature-based architecture (Auth, Dashboard, Market, Portfolio).
- web/src/graphql/: Dedicated folder for generated hooks and queries.


Next Steps

1. Install Dependencies:
- Frontend: cd web && npm install
- Gateway: cd gateway && npm install
- Backend: go mod download (in respective directories)
2. Run Locally:
- Use make run-all (once you configure the docker-compose fully) or run services individually.
3. Develop:
- Start implementing the GetStock logic in marketdata-service/internal/adapters.
- Define your GraphQL schemas in the Gateway or Service subgraphs.

