# Portfolio Service - Observability Metrics Implementation

## ✅ Implementation Complete

Successfully implemented comprehensive observability metrics for the portfolio-service using Prometheus.

---

## 📋 What Was Added

### **1. Metrics Infrastructure** (`internal/metrics/metrics.go`)

**gRPC Metrics:**
- `portfolio_grpc_requests_total`: Counter by method and status
- `portfolio_grpc_request_duration_seconds`: Histogram of request latency

**Business Metrics:**
- `portfolio_holdings_total`: Total number of holdings (Gauge)
- `portfolio_holdings_by_user`: Holdings per user (Gauge)
- `portfolio_total_value`: Portfolio value per user (Gauge)

**Cache Metrics:**
- `portfolio_cache_hits_total`: Counter by cache type
- `portfolio_cache_misses_total`: Counter by cache type
- `portfolio_cache_operation_duration_seconds`: Histogram of cache latency

**Database Metrics:**
- `portfolio_database_queries_total`: Counter by operation and table
- `portfolio_database_query_duration_seconds`: Histogram of query latency
- `portfolio_database_errors_total`: Counter of DB errors

**Market Data Metrics:**
- `portfolio_marketdata_requests_total`: Counter by operation and status
- `portfolio_marketdata_request_duration_seconds`: Histogram of request latency
- `portfolio_prices_fetched_total`: Counter by source (cache vs service)

**NATS Metrics:**
- `portfolio_nats_messages_consumed_total`: Counter by subject and status
- `portfolio_nats_message_processing_duration_seconds`: Histogram of processing time

### **2. Instrumentation Points**

**gRPC Middleware (`internal/middleware/metrics.go`):**
- Automatically records all gRPC request durations and status codes
- Zero-code instrumentation for new RPC methods

**Repository Layer (`internal/repository/postgres_holding_repo.go`):**
- Records all DB query durations and errors
- Operations: Upsert, Get, List, Count, Delete

**MarketData Gateway (`internal/infrastructure/marketdata_gateway.go`):**
- Records external service call durations
- Tracks cache hits/misses
- Tracks price fetch sources

**NATS Subscriber (`internal/infrastructure/nats_subscriber.go`):**
- Records message consumption rates
- Tracks message processing duration
- Tracks processing errors

### **3. Metrics Server (`cmd/server/main.go`):**
- Exposes metrics at `http://localhost:9090/metrics`
- Runs on a separate port from the gRPC server
- Includes background worker to update business metrics (e.g., total holdings)

---

## 🚀 Usage

### Access Metrics

```bash
curl http://localhost:9090/metrics
```

### Example Output

```text
# HELP portfolio_grpc_requests_total Total number of gRPC requests
# TYPE portfolio_grpc_requests_total counter
portfolio_grpc_requests_total{method="/portfolio.PortfolioService/GetHoldings",status="OK"} 42

# HELP portfolio_cache_hits_total Total number of cache hits
# TYPE portfolio_cache_hits_total counter
portfolio_cache_hits_total{cache_type="price"} 150

# HELP portfolio_database_query_duration_seconds Duration of database queries in seconds
# TYPE portfolio_database_query_duration_seconds histogram
portfolio_database_query_duration_seconds_bucket{operation="list_by_user",table="holdings",le="0.005"} 10
...
```

---

## 📊 Grafana Integration

The dashboard created in `deployments/monitoring/grafana/dashboards/portfolio-insights-dashboard.json` is already configured to visualize these metrics.

**Key Panels:**
- **Request Rate**: `rate(portfolio_grpc_requests_total[5m])`
- **Cache Hit Rate**: `rate(portfolio_cache_hits_total[5m]) / (rate(portfolio_cache_hits_total[5m]) + rate(portfolio_cache_misses_total[5m]))`
- **DB Latency**: `histogram_quantile(0.95, rate(portfolio_database_query_duration_seconds_bucket[5m]))`
- **Market Data Latency**: `histogram_quantile(0.95, rate(portfolio_marketdata_request_duration_seconds_bucket[5m]))`

---

## ⚠️ Important Configuration

### Docker Compose Port Exposure

For Prometheus to scrape metrics, the metrics port (`9090` internally) must be exposed to the host.

In `deployments/docker-compose/docker-compose.yml`:

```yaml
  portfolio-service:
    ports:
      - "50052:50052"  # gRPC
      - "9098:9090"    # Metrics (Host:Container)
```

**Note:** We map to `9098` on the host to avoid conflicts if `9090` is used by Prometheus itself.

---

## 🧪 Testing

### Verify Metrics Server

```bash
# Start service
go run cmd/server/main.go

# Check metrics endpoint (Local Run)
curl http://localhost:9090/metrics | grep portfolio_

# Check metrics endpoint (Docker/Podman)
curl http://localhost:9098/metrics | grep portfolio_
```

### Verify Instrumentation

1. **Make a gRPC request:**
   ```bash
   grpcurl ... portfolio.PortfolioService/GetHoldings
   ```
   - Check `portfolio_grpc_requests_total` increases
   - Check `portfolio_database_queries_total` increases

2. **Wait for NATS message:**
   - Publish a transaction event
   - Check `portfolio_nats_messages_consumed_total` increases

3. **Check Cache:**
   - Request holdings twice
   - First time: `portfolio_cache_misses_total` increases
   - Second time: `portfolio_cache_hits_total` increases

---

## 📚 Files Modified

1. ✅ `internal/metrics/metrics.go` (Created)
2. ✅ `internal/middleware/metrics.go` (Created)
3. ✅ `internal/repository/postgres_holding_repo.go` (Modified)
4. ✅ `internal/infrastructure/marketdata_gateway.go` (Modified)
5. ✅ `internal/infrastructure/nats_subscriber.go` (Modified)
6. ✅ `cmd/server/main.go` (Modified)
7. ✅ `go.mod` (Updated)

---

**Status**: ✅ **Observability metrics successfully implemented!**

The portfolio-service is now fully instrumented and ready for production monitoring.

