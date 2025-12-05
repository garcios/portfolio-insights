# Grafana Loki Integration Plan

## Executive Summary

This plan integrates Grafana Loki into the existing observability stack (Prometheus + Jaeger) running on Podman, targeting 1TB/day log ingestion with HA capabilities.

## 1. Architecture Decision

### Deployment Mode: **Microservices Mode**

**Rationale for 1TB/day:**
- Single binary mode caps at ~100GB/day
- Microservices mode enables horizontal scaling
- Component-level HA and independent scaling
- Better resource utilization

### Components Required:
- **Distributor** (2 replicas): Receives logs, validates, forwards to ingesters
- **Ingester** (3 replicas): Writes logs to storage, creates chunks
- **Querier** (2 replicas): Handles LogQL queries
- **Query Frontend** (2 replicas): Query optimization, caching
- **Compactor** (1 replica): Background compaction
- **Index Gateway** (2 replicas): Index query optimization

## 2. Loki Agent Selection

### Recommended: **Promtail**

**Why Promtail:**
- Native Loki integration
- Service discovery matches Prometheus patterns
- Low resource overhead
- JSON log parsing built-in
- Podman container log support

**Alternatives Considered:**
- Grafana Alloy: More features but heavier
- Fluentd/Fluent Bit: Overkill for this use case

## 3. Storage Backend

### Primary: **MinIO (S3-compatible)**

**Configuration:**
```yaml
Storage Layout:
- Chunks: MinIO bucket (loki-chunks)
- Index: MinIO bucket (loki-index)
- Ruler: MinIO bucket (loki-ruler)

Retention:
- 30 days for all logs
- Compaction every 2 hours
```

**Fallback: Filesystem** (for development/testing only)

## 4. Label Strategy

### Consistent with Prometheus Service Discovery

**Static Labels (Low Cardinality):**
```yaml
- service: <service-name>      # matches Prometheus
- tier: <api|backend|auth>     # matches Prometheus
- environment: development     # matches Prometheus
- cluster: portfolio-insights-dev
- container_name: <name>
- namespace: <podman-network>
```

**Indexed Labels:**
```yaml
- job: <job-name>              # matches Prometheus job_name
- level: <info|warn|error>     # log level
- app: <application-name>
```

**Avoid High Cardinality:**
- ❌ user_id, request_id, trace_id as labels
- ✅ Extract these in queries using LogQL parsers

## 5. Step-by-Step Integration

### Phase 1: Infrastructure Setup (Week 1)

#### Step 1.1: MinIO Deployment
```bash
# Add to deployments/monitoring/docker-compose.yml
```

**File:** `deployments/monitoring/docker-compose.yml`
```yaml
  minio:
    image: minio/minio:RELEASE.2024-01-16T16-07-38Z
    container_name: minio
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      MINIO_ROOT_USER: loki
      MINIO_ROOT_PASSWORD: lokipassword
    volumes:
      - minio-data:/data
    command: server /data --console-address ":9001"
    networks:
      - monitoring
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3
```

**Action:** Create MinIO buckets
```bash
# After MinIO starts, create buckets
podman exec -it minio mc alias set local http://localhost:9000 loki lokipassword
podman exec -it minio mc mb local/loki-chunks
podman exec -it minio mc mb local/loki-index
podman exec -it minio mc mb local/loki-ruler
```

#### Step 1.2: Loki Configuration

**File:** `deployments/monitoring/loki/loki-config.yml`
```yaml
auth_enabled: false

server:
  http_listen_port: 3100
  grpc_listen_port: 9096

common:
  path_prefix: /loki
  storage:
    s3:
      endpoint: minio:9000
      bucketnames: loki-chunks
      access_key_id: loki
      secret_access_key: lokipassword
      s3forcepathstyle: true
      insecure: true
  replication_factor: 2
  ring:
    kvstore:
      store: memberlist

memberlist:
  join_members:
    - loki-1:7946
    - loki-2:7946
    - loki-3:7946

schema_config:
  configs:
    - from: 2024-01-01
      store: tsdb
      object_store: s3
      schema: v12
      index:
        prefix: index_
        period: 24h

storage_config:
  tsdb_shipper:
    active_index_directory: /loki/tsdb-index
    cache_location: /loki/tsdb-cache
    shared_store: s3
  
  aws:
    s3: s3://loki:lokipassword@minio:9000/loki-chunks
    s3forcepathstyle: true
    insecure: true

compactor:
  working_directory: /loki/compactor
  shared_store: s3
  compaction_interval: 2h

limits_config:
  retention_period: 720h  # 30 days
  ingestion_rate_mb: 50
  ingestion_burst_size_mb: 100
  max_query_series: 10000
  max_query_parallelism: 32

chunk_store_config:
  max_look_back_period: 720h

table_manager:
  retention_deletes_enabled: true
  retention_period: 720h

query_range:
  align_queries_with_step: true
  cache_results: true
  results_cache:
    cache:
      embedded_cache:
        enabled: true
        max_size_mb: 500

ruler:
  storage:
    type: s3
    s3:
      bucketnames: loki-ruler
```

#### Step 1.3: Loki Microservices Deployment

**File:** `deployments/monitoring/docker-compose.loki.yml`
```yaml
services:
  # Read path
  loki-read:
    image: grafana/loki:2.9.3
    container_name: loki-read
    command: "-config.file=/etc/loki/config.yml -target=read"
    ports:
      - "3100:3100"
      - "7946"
      - "9095"
    volumes:
      - ./loki/loki-config.yml:/etc/loki/config.yml:z
      - loki-read-data:/loki
    networks:
      - monitoring
    depends_on:
      - minio
    deploy:
      replicas: 2

  # Write path
  loki-write:
    image: grafana/loki:2.9.3
    container_name: loki-write
    command: "-config.file=/etc/loki/config.yml -target=write"
    ports:
      - "3101:3100"
      - "7946"
      - "9095"
    volumes:
      - ./loki/loki-config.yml:/etc/loki/config.yml:z
      - loki-write-data:/loki
    networks:
      - monitoring
    depends_on:
      - minio
    deploy:
      replicas: 3

  # Backend (compactor, ruler, etc)
  loki-backend:
    image: grafana/loki:2.9.3
    container_name: loki-backend
    command: "-config.file=/etc/loki/config.yml -target=backend"
    ports:
      - "3102:3100"
      - "7946"
      - "9095"
    volumes:
      - ./loki/loki-config.yml:/etc/loki/config.yml:z
      - loki-backend-data:/loki
    networks:
      - monitoring
    depends_on:
      - minio

volumes:
  loki-read-data:
  loki-write-data:
  loki-backend-data:
  minio-data:
```

#### Step 1.4: Promtail Configuration

**File:** `deployments/monitoring/promtail/promtail-config.yml`
```yaml
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki-write:3100/loki/api/v1/push
    tenant_id: portfolio-insights

scrape_configs:
  # Podman container logs
  - job_name: podman-containers
    static_configs:
      - targets:
          - localhost
        labels:
          job: podman-containers
          __path__: /var/lib/containers/storage/overlay-containers/*/userdata/ctr.log
    
    pipeline_stages:
      - json:
          expressions:
            timestamp: time
            stream: stream
            log: log
            container_name: attrs.io.kubernetes.container.name
      
      - labels:
          container_name:
          stream:
      
      - timestamp:
          source: timestamp
          format: RFC3339Nano

  # Gateway service (JSON logs)
  - job_name: gateway-service
    static_configs:
      - targets:
          - localhost
        labels:
          job: gateway-service
          service: gateway
          tier: api
          environment: development
          cluster: portfolio-insights-dev
          __path__: /var/log/portfolio-insights/gateway/*.log
    
    pipeline_stages:
      - json:
          expressions:
            level: level
            timestamp: timestamp
            message: message
            trace_id: trace_id
            span_id: span_id
            service: service
      
      - labels:
          level:
          service:
      
      - timestamp:
          source: timestamp
          format: RFC3339

  # User service
  - job_name: user-service
    static_configs:
      - targets:
          - localhost
        labels:
          job: user-service
          service: user
          tier: backend
          environment: development
          cluster: portfolio-insights-dev
          __path__: /var/log/portfolio-insights/user-service/*.log
    
    pipeline_stages:
      - json:
          expressions:
            level: level
            timestamp: timestamp
            message: message
      
      - labels:
          level:
      
      - timestamp:
          source: timestamp
          format: RFC3339

  # Transaction service
  - job_name: transaction-service
    static_configs:
      - targets:
          - localhost
        labels:
          job: transaction-service
          service: transaction
          tier: backend
          environment: development
          cluster: portfolio-insights-dev
          __path__: /var/log/portfolio-insights/transaction-service/*.log
    
    pipeline_stages:
      - json:
          expressions:
            level: level
            timestamp: timestamp
            message: message
      
      - labels:
          level:
      
      - timestamp:
          source: timestamp
          format: RFC3339

  # Portfolio service
  - job_name: portfolio-service
    static_configs:
      - targets:
          - localhost
        labels:
          job: portfolio-service
          service: portfolio
          tier: backend
          environment: development
          cluster: portfolio-insights-dev
          __path__: /var/log/portfolio-insights/portfolio-service/*.log
    
    pipeline_stages:
      - json:
          expressions:
            level: level
            timestamp: timestamp
            message: message
      
      - labels:
          level:
      
      - timestamp:
          source: timestamp
          format: RFC3339

  # Market data service
  - job_name: marketdata-service
    static_configs:
      - targets:
          - localhost
        labels:
          job: marketdata-service
          service: marketdata
          tier: backend
          environment: development
          cluster: portfolio-insights-dev
          __path__: /var/log/portfolio-insights/marketdata-service/*.log
    
    pipeline_stages:
      - json:
          expressions:
            level: level
            timestamp: timestamp
            message: message
      
      - labels:
          level:
      
      - timestamp:
          source: timestamp
          format: RFC3339

  # Login consent provider
  - job_name: login-consent-provider
    static_configs:
      - targets:
          - localhost
        labels:
          job: login-consent-provider
          service: login-consent
          tier: auth
          environment: development
          cluster: portfolio-insights-dev
          __path__: /var/log/portfolio-insights/login-consent/*.log
    
    pipeline_stages:
      - json:
          expressions:
            level: level
            timestamp: timestamp
            message: message
      
      - labels:
          level:
      
      - timestamp:
          source: timestamp
          format: RFC3339
```

**File:** `deployments/monitoring/docker-compose.yml` (add Promtail)
```yaml
  promtail:
    image: grafana/promtail:2.9.3
    container_name: promtail
    volumes:
      - ./promtail/promtail-config.yml:/etc/promtail/config.yml:z
      - /var/log/portfolio-insights:/var/log/portfolio-insights:ro,z
      - /var/lib/containers/storage:/var/lib/containers/storage:ro,z
    command: -config.file=/etc/promtail/config.yml
    networks:
      - monitoring
    depends_on:
      - loki-write
```

### Phase 2: Application Instrumentation (Week 2)

#### Step 2.1: Update Services to Write JSON Logs

**Example for Go services:**

```go
// pkg/logger/logger.go
package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func NewLogger(serviceName string) (*zap.Logger, error) {
    config := zap.NewProductionConfig()
    config.OutputPaths = []string{
        "stdout",
        fmt.Sprintf("/var/log/portfolio-insights/%s/app.log", serviceName),
    }
    config.EncoderConfig.TimeKey = "timestamp"
    config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    
    logger, err := config.Build(
        zap.Fields(
            zap.String("service", serviceName),
            zap.String("environment", "development"),
            zap.String("cluster", "portfolio-insights-dev"),
        ),
    )
    return logger, err
}
```

**Action Items:**
1. Update each service's logger initialization
2. Ensure log directory `/var/log/portfolio-insights/<service>/` exists
3. Mount log directories in Podman containers

#### Step 2.2: Update Podman Compose Files

**File:** `deployments/docker-compose/docker-compose.yml` (example for gateway)
```yaml
  gateway:
    # ... existing config ...
    volumes:
      - /var/log/portfolio-insights/gateway:/var/log/portfolio-insights/gateway:z
```

### Phase 3: Grafana Integration (Week 2)

#### Step 3.1: Add Loki Data Source

**File:** `deployments/monitoring/grafana/provisioning/datasources/loki.yml`
```yaml
apiVersion: 1

datasources:
  - name: Loki
    type: loki
    access: proxy
    url: http://loki-read:3100
    jsonData:
      maxLines: 1000
      derivedFields:
        - datasourceUid: jaeger
          matcherRegex: "trace_id=(\\w+)"
          name: TraceID
          url: "$${__value.raw}"
```

#### Step 3.2: Create Combined Dashboard

**File:** `deployments/monitoring/grafana/dashboards/logs-metrics-traces.json`

Key panels:
1. **Log Volume** (Loki): `sum(rate({job=~".+"} [1m])) by (service)`
2. **Error Rate** (Loki): `sum(rate({level="error"} [5m])) by (service)`
3. **Request Rate** (Prometheus): Existing panels
4. **Latency** (Prometheus): Existing panels
5. **Logs Explorer** (Loki): Interactive log search
6. **Trace Correlation** (Loki→Jaeger): Click trace_id to jump to Jaeger

**Example LogQL Queries:**

```logql
# All logs for gateway service
{service="gateway"}

# Error logs across all services
{job=~".+"} |= "level=error"

# Logs with trace correlation
{service="gateway"} | json | trace_id != ""

# Request logs with latency > 1s
{service="gateway"} | json | duration > 1000

# Aggregation: Error rate by service
sum(rate({level="error"}[5m])) by (service)

# Pattern matching: Failed authentication
{service="login-consent"} |~ "authentication failed"
```

### Phase 4: Testing & Validation (Week 3)

#### Step 4.1: Verify Log Ingestion

```bash
# Check Promtail is scraping
curl http://localhost:9080/metrics | grep promtail_targets_active_total

# Check Loki is receiving logs
curl http://localhost:3100/loki/api/v1/label/service/values

# Query logs via API
curl -G -s "http://localhost:3100/loki/api/v1/query" \
  --data-urlencode 'query={service="gateway"}' \
  --data-urlencode 'limit=10'
```

#### Step 4.2: Performance Testing

```bash
# Generate load
hey -z 5m -c 50 http://localhost:8080/health

# Monitor Loki metrics
curl http://localhost:3100/metrics | grep loki_ingester_chunks_created_total

# Check MinIO storage
podman exec minio mc du local/loki-chunks
```

#### Step 4.3: HA Testing

```bash
# Stop one Loki write instance
podman stop loki-write-1

# Verify logs still ingesting
# Check Grafana dashboards

# Restart instance
podman start loki-write-1
```

### Phase 5: Production Hardening (Week 4)

#### Step 5.1: Resource Limits

**File:** `deployments/monitoring/docker-compose.loki.yml`
```yaml
  loki-write:
    # ... existing config ...
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G
        reservations:
          cpus: '1'
          memory: 2G
```

#### Step 5.2: Monitoring Loki Itself

Add to Prometheus scrape config:
```yaml
  - job_name: 'loki'
    static_configs:
      - targets: 
          - 'loki-read:3100'
          - 'loki-write:3100'
          - 'loki-backend:3100'
        labels:
          service: 'loki'
          tier: 'observability'
```

#### Step 5.3: Alerting Rules

**File:** `deployments/monitoring/loki/alerts.yml`
```yaml
groups:
  - name: loki_alerts
    interval: 1m
    rules:
      - alert: LokiHighIngestionRate
        expr: rate(loki_distributor_bytes_received_total[5m]) > 50000000
        for: 5m
        annotations:
          summary: "Loki ingestion rate is high"
      
      - alert: LokiIngesterDown
        expr: up{job="loki"} == 0
        for: 5m
        annotations:
          summary: "Loki ingester is down"
```

## 6. Deployment Commands

```bash
# Create log directories
sudo mkdir -p /var/log/portfolio-insights/{gateway,user-service,transaction-service,portfolio-service,marketdata-service,login-consent}
sudo chown -R $(id -u):$(id -g) /var/log/portfolio-insights

# Create Loki config directory
mkdir -p deployments/monitoring/loki
mkdir -p deployments/monitoring/promtail

# Start MinIO
cd deployments/monitoring
podman-compose -f docker-compose.yml up -d minio

# Create buckets (wait 30s for MinIO to start)
sleep 30
podman exec -it minio sh -c "mc alias set local http://localhost:9000 loki lokipassword && \
  mc mb local/loki-chunks && \
  mc mb local/loki-index && \
  mc mb local/loki-ruler"

# Start Loki
podman-compose -f docker-compose.loki.yml up -d

# Start Promtail
podman-compose -f docker-compose.yml up -d promtail

# Restart Grafana to pick up new datasource
podman-compose -f docker-compose.yml restart grafana
```

## 7. Verification Checklist

- [ ] MinIO accessible at http://localhost:9001
- [ ] Loki API responding: `curl http://localhost:3100/ready`
- [ ] Promtail scraping: `curl http://localhost:9080/metrics`
- [ ] Logs visible in Grafana Explore
- [ ] Labels match Prometheus service discovery
- [ ] Trace correlation working (Loki → Jaeger)
- [ ] Combined dashboard showing metrics + logs
- [ ] HA failover tested
- [ ] Retention policy working (30 days)
- [ ] Alerts configured and firing

## 8. Troubleshooting

### Issue: Promtail not finding logs
```bash
# Check file permissions
ls -la /var/log/portfolio-insights/

# Check Promtail targets
curl http://localhost:9080/targets
```

### Issue: Loki not storing data
```bash
# Check MinIO buckets
podman exec minio mc ls local/

# Check Loki logs
podman logs loki-write
```

### Issue: High cardinality warnings
```bash
# Check label cardinality
curl http://localhost:3100/loki/api/v1/labels | jq
```

## 9. Cost & Resource Estimates

**For 1TB/day:**
- **Storage**: ~30TB/month (with compression ~10TB)
- **MinIO**: 500GB SSD recommended
- **Loki Write**: 3 instances × 4GB RAM = 12GB
- **Loki Read**: 2 instances × 4GB RAM = 8GB
- **Promtail**: 512MB RAM
- **Total**: ~21GB RAM, 500GB storage

## 10. Next Steps

1. **Week 1**: Deploy infrastructure (MinIO, Loki, Promtail)
2. **Week 2**: Instrument applications, configure Grafana
3. **Week 3**: Test and validate
4. **Week 4**: Production hardening, monitoring, alerts

## References

- [Loki Documentation](https://grafana.com/docs/loki/latest/)
- [Promtail Configuration](https://grafana.com/docs/loki/latest/clients/promtail/)
- [LogQL Guide](https://grafana.com/docs/loki/latest/logql/)
