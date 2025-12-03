# Grafana Dashboard - Portfolio Insights

## 📊 Dashboard Overview

This dashboard provides comprehensive monitoring for the Portfolio Insights microservices application.

---

## 🎯 What's Monitored

### **1. HTTP/gRPC Metrics**
- Request rate (requests per second)
- 95th percentile response time
- gRPC request rate by service
- gRPC error rate by code
- Success rate percentage

### **2. Business Metrics**
- Transactions created (rate)
- Total users created
- Total portfolio holdings

### **3. Resource Usage**
- Memory usage by service
- CPU usage by service
- Goroutines by service

### **4. NATS Messaging**
- Message publish rate
- Message consume rate

---

## 📥 How to Import

### Method 1: Via Grafana UI (Recommended)

1. **Access Grafana**
   ```
   http://localhost:3000
   ```
   - Username: `admin`
   - Password: `admin` (change on first login)

2. **Import Dashboard**
   - Click the `+` icon in the left sidebar
   - Select `Import dashboard`
   - Click `Upload JSON file`
   - Select `portfolio-insights-dashboard.json`
   - Click `Load`

3. **Configure Data Source**
   - Select `Prometheus` as the data source
   - Click `Import`

### Method 2: Via File System (Auto-provisioning)

If using the monitoring stack with provisioning:

1. **Copy dashboard to provisioning folder**
   ```bash
   cp deployments/monitoring/grafana/dashboards/portfolio-insights-dashboard.json \
      deployments/monitoring/grafana/provisioning/dashboards/
   ```

2. **Restart Grafana**
   ```bash
   make monitoring-down
   make monitoring-up
   ```

3. **Dashboard auto-loads** on startup

---

## 📊 Dashboard Panels

### Row 1: Request Metrics

**Panel 1: HTTP Request Rate**
- Metric: `rate(http_requests_total[5m])`
- Shows: Requests per second
- Type: Time series graph

**Panel 2: 95th Percentile Response Time**
- Metric: `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))`
- Shows: Response time in milliseconds
- Type: Gauge
- Thresholds:
  - Green: < 100ms
  - Yellow: 100-500ms
  - Red: > 500ms

### Row 2: gRPC Metrics

**Panel 3: gRPC Request Rate by Service**
- Metric: `sum by (service) (rate(grpc_server_handled_total[5m]))`
- Shows: Requests per second per service
- Type: Time series graph
- Legend: Service names

**Panel 4: gRPC Error Rate by Code**
- Metric: `sum by (grpc_code) (rate(grpc_server_handled_total{grpc_code!="OK"}[5m]))`
- Shows: Error rate by gRPC status code
- Type: Stacked time series
- Legend: gRPC status codes

### Row 3: Business Metrics

**Panel 5: gRPC Success Rate**
- Metric: `100 * (1 - (errors / total))`
- Shows: Success rate percentage
- Type: Gauge
- Thresholds:
  - Green: > 80%
  - Red: < 80%

**Panel 6: Transactions Created**
- Metric: `sum(rate(transactions_created_total[5m]))`
- Shows: Transaction creation rate
- Type: Stat panel

**Panel 7: Total Users Created**
- Metric: `sum(users_created_total)`
- Shows: Cumulative user count
- Type: Stat panel

**Panel 8: Total Portfolio Holdings**
- Metric: `sum(portfolio_holdings_total)`
- Shows: Total holdings across all users
- Type: Stat panel

### Row 4: Resource Usage

**Panel 9: Memory Usage by Service**
- Metric: `process_resident_memory_bytes`
- Shows: Memory consumption per service
- Type: Time series graph
- Unit: Bytes

**Panel 10: CPU Usage by Service**
- Metric: `rate(process_cpu_seconds_total[5m]) * 100`
- Shows: CPU usage percentage per service
- Type: Time series graph
- Unit: Percent

### Row 5: Application Metrics

**Panel 11: Goroutines by Service**
- Metric: `go_goroutines`
- Shows: Number of goroutines per service
- Type: Time series graph
- Helps identify: Goroutine leaks

**Panel 12: NATS Message Rate**
- Metrics:
  - `rate(nats_messages_published_total[5m])` - Published
  - `rate(nats_messages_consumed_total[5m])` - Consumed
- Shows: Message throughput
- Type: Time series graph

---

## 🔧 Dashboard Settings

### Time Range
- Default: Last 1 hour
- Adjustable via time picker (top right)

### Refresh Rate
- Default: 5 seconds
- Auto-refresh enabled
- Configurable in dashboard settings

### Variables
- None currently defined
- Can be added for filtering by service, environment, etc.

---

## 📈 Using the Dashboard

### Monitor Overall Health

1. **Check Success Rate** (Panel 5)
   - Should be > 95%
   - If < 80%, investigate errors

2. **Check Response Times** (Panel 2)
   - Should be < 100ms for most requests
   - Spikes indicate performance issues

3. **Check Error Rates** (Panel 4)
   - Should be near zero
   - Identify which error codes are occurring

### Monitor Service Performance

1. **gRPC Request Rate** (Panel 3)
   - See which services are busiest
   - Identify traffic patterns

2. **Resource Usage** (Panels 9-10)
   - Monitor memory and CPU
   - Identify resource-hungry services
   - Detect memory leaks (steadily increasing)

3. **Goroutines** (Panel 11)
   - Should be relatively stable
   - Rapid increase indicates goroutine leak

### Monitor Business Metrics

1. **Transaction Rate** (Panel 6)
   - Track business activity
   - Identify usage patterns

2. **User Growth** (Panel 7)
   - Monitor user acquisition
   - Track cumulative growth

3. **Portfolio Holdings** (Panel 8)
   - Monitor total holdings
   - Track portfolio growth

### Monitor Messaging

1. **NATS Message Rate** (Panel 12)
   - Published vs Consumed should match
   - Lag indicates processing issues

---

## 🎨 Customization

### Add New Panel

1. Click `Add panel` (top right)
2. Select `Add a new panel`
3. Configure query:
   ```promql
   your_metric_name[5m]
   ```
4. Choose visualization type
5. Set title and description
6. Click `Apply`

### Modify Existing Panel

1. Click panel title
2. Select `Edit`
3. Modify query, visualization, or settings
4. Click `Apply`

### Add Variables

1. Click dashboard settings (gear icon)
2. Select `Variables`
3. Click `Add variable`
4. Configure:
   - Name: `service`
   - Type: `Query`
   - Query: `label_values(up, job)`
5. Use in queries: `{job="$service"}`

---

## 🔍 Common Queries

### Request Rate
```promql
rate(http_requests_total[5m])
```

### Error Rate
```promql
rate(http_requests_total{status=~"5.."}[5m])
```

### Response Time (p95)
```promql
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

### Memory Usage
```promql
process_resident_memory_bytes
```

### CPU Usage
```promql
rate(process_cpu_seconds_total[5m]) * 100
```

### Goroutines
```promql
go_goroutines
```

### gRPC Success Rate
```promql
100 * (
  sum(rate(grpc_server_handled_total{grpc_code="OK"}[5m]))
  /
  sum(rate(grpc_server_handled_total[5m]))
)
```

---

## 🚨 Alerts (Future Enhancement)

### Suggested Alerts

1. **High Error Rate**
   ```promql
   rate(grpc_server_handled_total{grpc_code!="OK"}[5m]) > 0.1
   ```
   - Threshold: > 10% error rate
   - Severity: Warning

2. **High Response Time**
   ```promql
   histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
   ```
   - Threshold: > 1 second
   - Severity: Warning

3. **High Memory Usage**
   ```promql
   process_resident_memory_bytes > 1e9
   ```
   - Threshold: > 1GB
   - Severity: Warning

4. **Service Down**
   ```promql
   up == 0
   ```
   - Threshold: Service unreachable
   - Severity: Critical

---

## 📊 Dashboard JSON Structure

```json
{
  "title": "Portfolio Insights - Application Dashboard",
  "uid": "portfolio-insights-app",
  "tags": ["portfolio-insights", "microservices"],
  "panels": [...],
  "refresh": "5s",
  "time": {
    "from": "now-1h",
    "to": "now"
  }
}
```

---

## 🔧 Troubleshooting

### Dashboard Shows "No Data"

**Check:**
1. Prometheus is running: `http://localhost:9081`
2. Services are exposing metrics
3. Prometheus is scraping targets
4. Data source is configured correctly

**Fix:**
```bash
# Check Prometheus targets
curl http://localhost:9081/api/v1/targets

# Check if metrics exist
curl http://localhost:9081/api/v1/query?query=up
```

### Panels Show Errors

**Common Issues:**
1. **Invalid query syntax**
   - Check PromQL syntax
   - Test in Prometheus UI first

2. **Missing metrics**
   - Verify services expose metrics
   - Check metric names match

3. **Data source not configured**
   - Go to Configuration → Data Sources
   - Add Prometheus data source
   - URL: `http://prometheus:9081`

### Slow Dashboard Loading

**Solutions:**
1. Reduce time range (e.g., last 15 minutes)
2. Increase refresh interval (e.g., 30 seconds)
3. Simplify complex queries
4. Add query result caching

---

## 📚 Resources

### Grafana Documentation
- [Grafana Dashboards](https://grafana.com/docs/grafana/latest/dashboards/)
- [Panel Types](https://grafana.com/docs/grafana/latest/panels/)
- [PromQL Queries](https://prometheus.io/docs/prometheus/latest/querying/basics/)

### Prometheus Metrics
- [Go Metrics](https://prometheus.io/docs/guides/go-application/)
- [gRPC Metrics](https://github.com/grpc-ecosystem/go-grpc-prometheus)
- [HTTP Metrics](https://prometheus.io/docs/instrumenting/writing_exporters/)

---

## 🎯 Next Steps

1. ✅ Import dashboard
2. ⏳ Verify all panels show data
3. ⏳ Customize for your needs
4. ⏳ Add alerts
5. ⏳ Create additional dashboards:
   - Database performance
   - Redis cache metrics
   - Business analytics
   - User behavior

---

**Dashboard File**: `portfolio-insights-dashboard.json`  
**Version**: 1.0  
**Last Updated**: 2025-11-24  
**Compatible With**: Grafana 9.0+, Prometheus 2.0+

