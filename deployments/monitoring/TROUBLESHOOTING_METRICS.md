# Troubleshooting Grafana Metrics

If you are not seeing data in the Grafana dashboard, follow these steps to diagnose and fix the issue.

## 1. Verify Services are Running

Ensure all services are up and running:

```bash
make podman-up
# OR
podman-compose -f deployments/docker-compose/docker-compose.yml up -d
```

Check status:
```bash
podman ps
```
You should see `portfolio-service`, `transaction-service`, `prometheus`, and `grafana`.

## 2. Verify Metrics Endpoints (Local)

Check if the services are exposing metrics on the expected ports.

**Portfolio Service:**
```bash
curl http://localhost:9098/metrics | grep portfolio_
```
*Expected Output:*
```
portfolio_grpc_requests_total{method="/portfolio.PortfolioService/GetHoldings",status="OK"} 0
...
```

**Transaction Service:**
```bash
curl http://localhost:9097/metrics
```

If these fail, the ports are not correctly exposed in `docker-compose.yml`.

**Fix:**
Ensure `docker-compose.yml` has:
```yaml
  portfolio-service:
    ports:
      - "50052:50052"
      - "9098:9098"  # Metrics
```

## 3. Verify Prometheus Targets

Check if Prometheus is successfully scraping the targets.

1. Open Prometheus UI: http://localhost:9081
2. Go to **Status** -> **Targets**
3. Check the status of `portfolio-service` and `transaction-service`.

- **UP (Green)**: Scraping is working.
- **DOWN (Red)**: Scraping failed. Check the error message.

**Common Errors:**
- `connection refused`: Prometheus cannot reach the host.
  - **Fix**: Ensure `host.docker.internal` is working. On Linux, you might need extra configuration.
  - **Workaround**: Use the host IP address in `prometheus.yml` instead of `host.docker.internal`.

## 4. Verify Dashboard Queries

If Prometheus has data but Grafana is empty:

1. Open the dashboard in Grafana.
2. Click the title of an empty panel -> **Edit**.
3. Check the **PromQL Query**.
4. Run the same query in Prometheus UI (http://localhost:9081/graph) to see if it returns data.

**Correct Metric Names:**
- `portfolio_grpc_requests_total`
- `portfolio_grpc_request_duration_seconds_bucket`
- `portfolio_holdings_total`

## 5. Generate Traffic

Metrics might be empty simply because there is no traffic.

Generate some requests:

```bash
# Get Holdings
grpcurl -plaintext -d '{"user_id": "user-1"}' localhost:50052 portfolio.PortfolioService/GetHoldings
```

Check Grafana again after 15-30 seconds.

## 6. Re-import Dashboard

If you updated the dashboard JSON, make sure to re-import it in Grafana:
1. Go to **Dashboards** -> **Import**.
2. Upload the updated `portfolio-insights-dashboard.json`.
3. Select "Overwrite" if prompted.
