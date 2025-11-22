# Prometheus & Grafana Monitoring

This directory contains the monitoring stack configuration for the Portfolio Insights microservices.

## Stack Components

- **Prometheus** (port 9090): Metrics collection and storage
- **Grafana** (port 3001): Visualization and dashboarding
- **AlertManager** (port 9093): Alert routing and management

## Quick Start

### 1. Start the Monitoring Stack

```bash
cd deployments/monitoring
docker-compose up -d
```

### 2. Access the UIs

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3001 (admin/admin)
- **AlertManager**: http://localhost:9093

### 3. Verify Metrics Collection

Check that Prometheus is scraping all services:
```bash
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'
```

### 4. Test Transaction Service Metrics

```bash
# Check if metrics endpoint is accessible
curl http://localhost:9097/metrics

# Check health endpoint
curl http://localhost:9097/health
```

## Service Metrics Ports

Each service exposes metrics on a dedicated port:

| Service | Metrics Port |
|---------|--------------|
| Gateway | 9095 |
| User Service | 9096 |
| Transaction Service | 9097 |
| Portfolio Service | 9098 |
| Market Data Service | 9099 |

## Key Metrics

### Golden Signals

1. **Latency**: `http_request_duration_seconds`, `grpc_request_duration_seconds`
2. **Traffic**: `http_requests_total`, `grpc_requests_total`
3. **Errors**: `http_requests_total{status_code=~"5.."}`
4. **Saturation**: `go_goroutines`, `go_memstats_heap_alloc_bytes`

### Business Metrics

**Transaction Service:**
- `transactions_created_total` - Total transactions created
- `transaction_value_total` - Total transaction value
- `transaction_processing_duration_seconds` - Processing time
- `user_validation_duration_seconds` - User validation latency
- `asset_validation_duration_seconds` - Asset validation latency

**Portfolio Service:**
- `portfolio_holdings_total` - Current holdings count
- `nats_events_processed_total` - NATS events processed
- `nats_event_processing_duration_seconds` - Event processing time

## Example PromQL Queries

### Request Rate (RPS)
```promql
sum(rate(grpc_requests_total[5m])) by (service)
```

### Error Rate
```promql
sum(rate(grpc_requests_total{status!="OK"}[5m])) by (service) 
/ 
sum(rate(grpc_requests_total[5m])) by (service)
```

### P95 Latency
```promql
histogram_quantile(0.95, 
  sum(rate(grpc_request_duration_seconds_bucket[5m])) by (service, le)
)
```

### Transaction Creation Rate
```promql
sum(rate(transactions_created_total[5m])) by (type)
```

### NATS Event Success Rate
```promql
sum(rate(nats_events_processed_total{status="success"}[5m])) 
/ 
sum(rate(nats_events_processed_total[5m]))
```

## Alerting

Alerts are defined in `prometheus/alerts.yml` and include:

- **HighErrorRate**: Error rate > 1% for 5 minutes
- **HighLatency**: P95 latency > 1s for 10 minutes
- **ServiceDown**: Service unreachable for 2 minutes
- **HighMemoryUsage**: Heap usage > 90% for 15 minutes
- **GoroutineLeak**: Goroutine count > 1000 and growing
- **HighTransactionFailureRate**: Transaction failures > 5%
- **HighNATSEventFailureRate**: NATS event failures > 5%
- **SlowGRPCCalls**: gRPC P95 > 0.5s for 10 minutes

## Grafana Dashboards

### Creating Dashboards

1. Log in to Grafana (http://localhost:3001)
2. Click "+" → "Dashboard"
3. Add panels with PromQL queries
4. Save the dashboard

### Recommended Dashboards

1. **Golden Signals Dashboard**
   - Request rate by service
   - Error rate by service
   - P50/P95/P99 latency
   - In-flight requests

2. **Go Runtime Dashboard**
   - Goroutine count
   - Heap memory usage
   - GC pause duration
   - CPU usage

3. **Business Metrics Dashboard**
   - Transaction volume by type
   - Transaction value by type
   - Portfolio holdings count
   - NATS event processing rate

## Troubleshooting

### Prometheus not scraping services

1. Check if services are running:
```bash
curl http://localhost:9097/health
```

2. Check Prometheus targets:
```bash
open http://localhost:9090/targets
```

3. Verify network connectivity from Docker:
```bash
docker exec prometheus wget -O- http://host.docker.internal:9097/metrics
```

### No data in Grafana

1. Verify Prometheus datasource:
   - Go to Configuration → Data Sources
   - Test the Prometheus connection

2. Check if metrics exist in Prometheus:
```bash
curl 'http://localhost:9090/api/v1/query?query=up'
```

### Alerts not firing

1. Check alert rules are loaded:
```bash
curl http://localhost:9090/api/v1/rules
```

2. Verify AlertManager is receiving alerts:
```bash
curl http://localhost:9093/api/v1/alerts
```

## Stopping the Stack

```bash
cd deployments/monitoring
docker-compose down
```

To remove volumes (deletes all metrics data):
```bash
docker-compose down -v
```

## Next Steps

1. ✅ Transaction service instrumented
2. ⏳ Instrument remaining services (user, portfolio, marketdata, gateway)
3. ⏳ Create Grafana dashboards
4. ⏳ Configure Slack/email alerts
5. ⏳ Set up long-term metrics storage
6. ⏳ Deploy to production with Kubernetes service discovery
