# GitHub Actions CI Workflow Fixes

This document summarizes the fixes applied to resolve issues in the GitHub Actions test coverage workflow.

## Issues Encountered and Resolutions

### 1. NATS Service Container Failure

**Error:**
```
Service container nats failed.
Error: Failed to initialize container nats:latest
```

**Root Cause:**
The health check command used flags (`--no-verbose --tries=1`) not supported by the minimal `wget` in the `nats:latest` image.

**Fix:**
- Changed image from `nats:latest` to `nats:2-alpine`
- Updated health check to use compatible flags: `wget -q --spider http://localhost:8222/healthz || exit 1`

**File Modified:** `.github/workflows/coverage.yml`

---

### 2. Go Module Download Failure

**Error:**
```
go: no modules specified (see 'go help mod download')
```

**Root Cause:**
The workflow was running `go mod download` in the root directory, but this is a Go workspace with no root `go.mod` file.

**Fix:**
Removed the "Install dependencies" step. The `go test` commands in subsequent steps automatically download required modules.

**File Modified:** `.github/workflows/coverage.yml`

---

### 3. Missing Protobuf Generated Files

**Error:**
```
no required module provides package github.com/garcios/portfolio-insights/services/user-service/proto/user
```

**Root Cause:**
Generated protobuf files (`*.pb.go`) are ignored by `.gitignore` and not present in the repository. They need to be generated during CI.

**Fix:**
Added three new steps to the workflow:
1. **Install protoc**: Downloads and installs Protocol Buffers compiler v25.1
2. **Install Go protoc plugins**: Installs `protoc-gen-go` and `protoc-gen-go-grpc`
3. **Generate protobuf files**: Runs `make proto-gen` to generate all `.pb.go` files

**File Modified:** `.github/workflows/coverage.yml`

---

### 4. Cross-Module Dependencies Not Resolved

**Error:**
```
no required module provides package github.com/garcios/portfolio-insights/services/marketdata-service/proto/marketdata
```

**Root Cause:**
The `go.work` file (which manages workspace dependencies) is ignored by `.gitignore`, so cross-module references don't work in CI without explicit `replace` directives.

**Fix:**
Added cross-module dependencies with `replace` directives:

**`services/portfolio-service/go.mod`:**
```go
require (
    github.com/garcios/portfolio-insights/services/marketdata-service v0.0.0
    // ... other deps
)

replace github.com/garcios/portfolio-insights/services/marketdata-service => ../marketdata-service
```

**Files Modified:**
- `services/portfolio-service/go.mod`
- `services/portfolio-service/go.sum` (via `go mod tidy`)

---

### 5. Missing gRPC Dependencies

**Error:**
```
no required module provides package google.golang.org/grpc
```

**Root Cause:**
The `marketdata-service/go.mod` didn't explicitly list `google.golang.org/grpc` as a dependency, even though the generated protobuf code requires it.

**Fix:**
Added gRPC and protobuf as direct dependencies:

**`services/marketdata-service/go.mod`:**
```go
require (
    github.com/garcios/portfolio-insights/pkg v0.0.0
    google.golang.org/grpc v1.77.0
    google.golang.org/protobuf v1.36.10
)
```

**Files Modified:**
- `services/marketdata-service/go.mod`
- `services/marketdata-service/go.sum` (via `go mod tidy`)

---

### 6. MarketData Service Test Mock Errors

**Error:**
```
*MockMarketDataRepository does not implement domain.MarketDataRepository (missing method GetHistoricalCurrencyRates)
```

**Root Cause:**
The mock implementations in test files were missing the `GetHistoricalCurrencyRates` method that was added to the domain interfaces.

**Fix:**
Added the missing method to both mock implementations:

**Files Modified:**
- `services/marketdata-service/internal/usecase/marketdata_usecase_test.go`
- `services/marketdata-service/internal/handler/grpc/handler_test.go`

---

## Summary of Files Modified

### GitHub Actions Workflow
- `.github/workflows/coverage.yml`

### Go Module Files
- `services/portfolio-service/go.mod`
- `services/portfolio-service/go.sum`
- `services/marketdata-service/go.mod`
- `services/marketdata-service/go.sum`

### Test Mocks
- `services/marketdata-service/internal/usecase/marketdata_usecase_test.go`
- `services/marketdata-service/internal/handler/grpc/handler_test.go`

---

## Verification

After these fixes, the CI workflow should:
1. ✅ Start all service containers (PostgreSQL, Redis, NATS) successfully
2. ✅ Install protoc and Go plugins
3. ✅ Generate all protobuf files
4. ✅ Run database migrations
5. ✅ Execute tests for all Go services with coverage
6. ✅ Execute frontend tests with coverage
7. ✅ Upload coverage reports to Codecov

---

## Related Documentation

- [Codecov Setup Guide](./CODECOV_SETUP.md)
- [Frontend Tests](./FRONTEND_TESTS.md)
- [Development Journal](./JOURNAL.md)
