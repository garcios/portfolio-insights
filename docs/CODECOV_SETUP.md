# Test Coverage with Codecov

This document explains how to set up and use Codecov for test coverage reporting in the Portfolio Insights project.

## Overview

The project uses GitHub Actions to automatically run tests and upload coverage reports to Codecov for all services and the frontend application.

## Setup Instructions

### 1. Create Codecov Account

1. Go to [codecov.io](https://codecov.io/)
2. Sign in with your GitHub account
3. Authorize Codecov to access your repositories
4. Select the `portfolio-insights` repository

### 2. Get Codecov Token

1. Navigate to your repository settings on Codecov
2. Copy the **Upload Token** (also called Repository Upload Token)
3. Keep this token secure - you'll need it for GitHub Actions

### 3. Add Token to GitHub Secrets

1. Go to your GitHub repository
2. Navigate to **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Name: `CODECOV_TOKEN`
5. Value: Paste the token from Codecov
6. Click **Add secret**

### 4. Verify Workflow

The workflow will automatically run on:
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches

You can manually trigger it from the **Actions** tab.

## Workflow Structure

### Services Tested

The workflow tests the following components:

#### Go Services
1. **User Service** (`services/user-service`)
2. **Transaction Service** (`services/transaction-service`)
3. **Portfolio Service** (`services/portfolio-service`)
4. **MarketData Service** (`services/marketdata-service`)
5. **Gateway** (`apps/gateway`)

#### Frontend
- **React Application** (`apps/frontend`)

### Test Environment

The workflow sets up the following services:
- **PostgreSQL 16** (port 5432)
- **Redis 7** (port 6379)
- **NATS** (ports 4222, 8222)

Database migrations are automatically applied before running tests.

### Coverage Flags

Each service has its own coverage flag for separate tracking:
- `user-service`
- `transaction-service`
- `portfolio-service`
- `marketdata-service`
- `gateway`
- `frontend`
- `combined` (all services together)

## Running Tests Locally

### Go Services

```bash
# Run tests for a specific service
cd services/portfolio-service
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# View coverage report
go tool cover -html=coverage.out
```

### Frontend

```bash
cd apps/frontend

# Install dependencies (if not already installed)
npm install --save-dev vitest @vitest/coverage-v8

# Run tests
npm test

# Run tests with coverage
npm run test:coverage
```

## Coverage Configuration

### codecov.yml

The `codecov.yml` file in the project root configures:

- **Coverage targets**: 70% minimum coverage
- **Threshold**: 5% change tolerance
- **Ignored files**: 
  - Test files (`*_test.go`)
  - Generated code (`*.pb.go`, `*_grpc.pb.go`)
  - Proto definitions
  - Mocks and test data
  - Node modules and build artifacts

### Coverage Targets

- **Project Coverage**: 70% minimum
- **Patch Coverage**: 70% minimum for new code
- **Threshold**: 5% decrease allowed

## Viewing Coverage Reports

### On Codecov Dashboard

1. Go to [codecov.io](https://codecov.io/)
2. Navigate to your repository
3. View coverage metrics:
   - Overall coverage percentage
   - Coverage by service (using flags)
   - Coverage trends over time
   - File-by-file coverage
   - Uncovered lines

### On Pull Requests

Codecov automatically comments on pull requests with:
- Coverage change summary
- Affected files
- Coverage diff
- Links to detailed reports

### Coverage Badges

Add coverage badges to your README:

```markdown
[![codecov](https://codecov.io/gh/YOUR_USERNAME/portfolio-insights/branch/main/graph/badge.svg)](https://codecov.io/gh/YOUR_USERNAME/portfolio-insights)
```

## Troubleshooting

### Workflow Fails

**Issue**: Tests fail in CI but pass locally

**Solutions**:
1. Check service dependencies (PostgreSQL, Redis, NATS)
2. Verify environment variables
3. Check database migrations
4. Review workflow logs in GitHub Actions

### Coverage Not Uploading

**Issue**: Coverage reports not appearing on Codecov

**Solutions**:
1. Verify `CODECOV_TOKEN` is set correctly in GitHub Secrets
2. Check workflow logs for upload errors
3. Ensure coverage files are generated (`coverage.out`)
4. Verify Codecov action version is up to date

### Low Coverage Warnings

**Issue**: Coverage below 70% threshold

**Solutions**:
1. Add unit tests for uncovered code
2. Add integration tests for critical paths
3. Review coverage report to identify gaps
4. Consider adjusting coverage targets if appropriate

## Best Practices

### Writing Tests

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test service interactions
3. **Table-Driven Tests**: Use Go's table-driven test pattern
4. **Mocking**: Mock external dependencies (database, gRPC clients)
5. **Edge Cases**: Test error conditions and boundary cases

### Coverage Goals

- **Critical Code**: Aim for 90%+ coverage
  - Authentication/authorization
  - Financial calculations
  - Data persistence
  
- **Business Logic**: Aim for 80%+ coverage
  - Portfolio calculations
  - Transaction processing
  - Market data handling

- **Infrastructure Code**: Aim for 60%+ coverage
  - HTTP handlers
  - gRPC servers
  - Middleware

### Maintaining Coverage

1. **Pre-commit**: Run tests locally before committing
2. **PR Reviews**: Check coverage reports on pull requests
3. **Regular Audits**: Review coverage trends monthly
4. **Refactoring**: Improve testability during refactoring

## CI/CD Integration

### GitHub Actions Workflow

The workflow runs in three jobs:

1. **test-go-services**: Tests all Go services with coverage
2. **test-frontend**: Tests React application with coverage
3. **coverage-report**: Combines and uploads all coverage data

### Workflow Triggers

```yaml
on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]
```

### Service Dependencies

Services are started using GitHub Actions service containers:
- Automatic health checks
- Isolated test environment
- Consistent with production setup

## Adding New Services

To add coverage for a new service:

1. **Add test step** in `.github/workflows/coverage.yml`:
   ```yaml
   - name: Test New Service
     env:
       DATABASE_URL: postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable
     run: |
       cd services/new-service
       go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
   ```

2. **Add upload step**:
   ```yaml
   - name: Upload New Service coverage to Codecov
     uses: codecov/codecov-action@v4
     with:
       token: ${{ secrets.CODECOV_TOKEN }}
       files: ./services/new-service/coverage.out
       flags: new-service
       name: new-service-coverage
   ```

3. **Add flag** in `codecov.yml`:
   ```yaml
   flags:
     new-service:
       paths:
         - services/new-service/
       carryforward: true
   ```

## Resources

- [Codecov Documentation](https://docs.codecov.com/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Vitest Documentation](https://vitest.dev/)

## Support

For issues or questions:
1. Check GitHub Actions workflow logs
2. Review Codecov dashboard for upload status
3. Consult this documentation
4. Open an issue in the repository

---

*Last Updated: 2025-11-28*
