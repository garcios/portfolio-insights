---
description: Prometheus & Grafana monitoring implementation plan
---

# Prometheus & Grafana Monitoring Implementation Plan

This document provides a detailed, phased approach to implementing production-ready monitoring for the portfolio-insights microservices using Prometheus and Grafana.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Grafana Dashboards                       │
│  (Visualization, Alerting UI, Dashboard Management)         │
└────────────────────────┬────────────────────────────────────┘
                         │ PromQL Queries
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   Prometheus Server                          │
│  (Metrics Storage, PromQL Engine, Alert Manager)            │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTP Pull (Scrape /metrics)
         ┌───────────────┼───────────────┬─────────────────┐
         ▼               ▼               ▼                 ▼
┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│   Gateway   │  │User Service │  │Transaction  │  │ Portfolio   │
│   Service   │  │             │  │  Service    │  │  Service    │
│ :8080/metrics│ │:8081/metrics│  │:8082/metrics│  │:8083/metrics│
└─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘
```

---

## Phase 1: Planning, Setup & Tooling

### 1.1 Define Key Metrics

#### Golden Signals for Each Microservice:

**1. Latency (Request Duration)**
- Metric: `http_request_duration_seconds` (Histogram)
- Labels: `service`, `handler`, `method`, `status_code`
- Percentiles: P50, P95, P99

**2. Traffic (Request Rate)**
- Metric: `http_requests_total` (Counter)
- Labels: `service`, `handler`, `method`, `status_code`
- Aggregation: Rate over time windows (1m, 5m, 15m)

**3. Errors (Error Rate)**
- Metric: `http_requests_total{status_code=~"5.."}`
- Calculation: Error rate as percentage of total requests
- Alert threshold: >1% error rate sustained for 5 minutes

**4. Saturation (Resource Utilization)**
- Go Runtime Metrics:
  - `go_goroutines` (Gauge) - Number of goroutines
  - `go_memstats_heap_alloc_bytes` (Gauge) - Heap memory allocated
  - `go_memstats_heap_inuse_bytes` (Gauge) - Heap memory in use
  - `go_gc_duration_seconds` (Summary) - GC pause duration
- System Metrics:
  - `process_cpu_seconds_total` (Counter) - CPU time
  - `process_resident_memory_bytes` (Gauge) - RSS memory

#### Business-Specific Metrics:

**Transaction Service:**
- `transaction_created_total` (Counter) - Total transactions created
- `transaction_value_total` (Counter) - Total transaction value in USD
- `transaction_processing_duration_seconds` (Histogram) - Processing time

**Portfolio Service:**
- `portfolio_holdings_total` (Gauge) - Total holdings tracked
- `portfolio_update_duration_seconds` (Histogram) - Portfolio update latency
- `nats_events_processed_total` (Counter) - NATS events processed
- `nats_events_failed_total` (Counter) - NATS event processing failures

**User Service:**
- `user_registrations_total` (Counter) - Total user registrations
- `user_auth_attempts_total` (Counter) - Authentication attempts
- `user_auth_failures_total` (Counter) - Failed authentications

**Gateway Service:**
- `graphql_query_duration_seconds` (Histogram) - GraphQL query latency
- `graphql_queries_total` (Counter) - Total GraphQL queries
- `grpc_client_requests_total` (Counter) - gRPC client requests to backends

### 1.2 Architecture Diagram

See ASCII diagram above. Key points:
- **Pull Model**: Prometheus scrapes `/metrics` endpoints from each service
- **Service Discovery**: Static configuration initially, Kubernetes SD for production
- **Data Flow**: Services expose metrics → Prometheus scrapes → Grafana queries Prometheus
- **Alerting**: Prometheus evaluates rules → AlertManager → Notifications (Slack/Email)

### 1.3 Environment Setup

#### Docker Compose Configuration

Create `deployments/monitoring/docker-compose.yml`:

```yaml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:v2.48.0
    container_name: prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
      - ./prometheus/alerts.yml:/etc/prometheus/alerts.yml
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/usr/share/prometheus/console_libraries'
      - '--web.console.templates=/usr/share/prometheus/consoles'
      - '--web.enable-lifecycle'
    networks:
      - monitoring

  grafana:
    image: grafana/grafana:10.2.2
    container_name: grafana
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - ./grafana/provisioning:/etc/grafana/provisioning
      - ./grafana/dashboards:/var/lib/grafana/dashboards
      - grafana-data:/var/lib/grafana
    networks:
      - monitoring
    depends_on:
      - prometheus

  alertmanager:
    image: prom/alertmanager:v0.26.0
    container_name: alertmanager
    ports:
      - "9093:9093"
    volumes:
      - ./alertmanager/alertmanager.yml:/etc/alertmanager/alertmanager.yml
      - alertmanager-data:/alertmanager
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--storage.path=/alertmanager'
    networks:
      - monitoring

volumes:
  prometheus-data:
  grafana-data:
  alertmanager-data:

networks:
  monitoring:
    driver: bridge
```

### 1.4 Configuration Management

#### Prometheus Configuration (`deployments/monitoring/prometheus/prometheus.yml`)

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: 'portfolio-insights-dev'
    environment: 'development'

# Alertmanager configuration
alerting:
  alertmanagers:
    - static_configs:
        - targets:
            - alertmanager:9093

# Load rules once and periodically evaluate them
rule_files:
  - "alerts.yml"

# Scrape configurations
scrape_configs:
  # Prometheus self-monitoring
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  # Gateway Service
  - job_name: 'gateway-service'
    static_configs:
      - targets: ['host.docker.internal:8080']
        labels:
          service: 'gateway'
          tier: 'api'

  # User Service
  - job_name: 'user-service'
    static_configs:
      - targets: ['host.docker.internal:8081']
        labels:
          service: 'user'
          tier: 'backend'

  # Transaction Service
  - job_name: 'transaction-service'
    static_configs:
      - targets: ['host.docker.internal:8082']
        labels:
          service: 'transaction'
          tier: 'backend'

  # Portfolio Service
  - job_name: 'portfolio-service'
    static_configs:
      - targets: ['host.docker.internal:8083']
        labels:
          service: 'portfolio'
          tier: 'backend'

  # Market Data Service
  - job_name: 'marketdata-service'
    static_configs:
      - targets: ['host.docker.internal:8084']
        labels:
          service: 'marketdata'
          tier: 'backend'
```

#### Grafana Datasource Provisioning (`deployments/monitoring/grafana/provisioning/datasources/prometheus.yml`)

```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
    jsonData:
      timeInterval: "15s"
      queryTimeout: "60s"
```

---

## Phase 2: Golang Microservice Instrumentation

### 2.1 Dependency Integration

Add to each service's `go.mod`:

```bash
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promauto
go get github.com/prometheus/client_golang/prometheus/promhttp
```

### 2.2 Create Metrics Package

Create `internal/metrics/metrics.go` for each service:

```go
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP Metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"service", "handler", "method", "status_code"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets, // [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
		},
		[]string{"service", "handler", "method", "status_code"},
	)

	HTTPInFlightRequests = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_in_flight_requests",
			Help: "Current number of in-flight HTTP requests",
		},
		[]string{"service", "handler"},
	)

	// gRPC Metrics
	GRPCRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"service", "method", "status"},
	)

	GRPCRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "status"},
	)
)

// Service-specific metrics will be defined in separate files
```

### 2.3 HTTP Metrics Middleware

Create `internal/middleware/metrics.go`:

```go
package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/garcios/portfolio-insights/services/[SERVICE_NAME]/internal/metrics"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// MetricsMiddleware instruments HTTP handlers with Prometheus metrics
func MetricsMiddleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Track in-flight requests
			metrics.HTTPInFlightRequests.WithLabelValues(serviceName, r.URL.Path).Inc()
			defer metrics.HTTPInFlightRequests.WithLabelValues(serviceName, r.URL.Path).Dec()

			// Wrap response writer to capture status code
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			
			// Process request
			next.ServeHTTP(rw, r)
			
			// Record metrics
			duration := time.Since(start).Seconds()
			statusCode := strconv.Itoa(rw.statusCode)
			
			metrics.HTTPRequestsTotal.WithLabelValues(
				serviceName,
				r.URL.Path,
				r.Method,
				statusCode,
			).Inc()
			
			metrics.HTTPRequestDuration.WithLabelValues(
				serviceName,
				r.URL.Path,
				r.Method,
				statusCode,
			).Observe(duration)
		})
	}
}
```

### 2.4 gRPC Metrics Interceptor

Create `internal/middleware/grpc_metrics.go`:

```go
package middleware

import (
	"context"
	"time"

	"github.com/garcios/portfolio-insights/services/[SERVICE_NAME]/internal/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns a gRPC unary server interceptor for metrics
func UnaryServerInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		
		// Call the handler
		resp, err := handler(ctx, req)
		
		// Record metrics
		duration := time.Since(start).Seconds()
		statusCode := status.Code(err).String()
		
		metrics.GRPCRequestsTotal.WithLabelValues(
			serviceName,
			info.FullMethod,
			statusCode,
		).Inc()
		
		metrics.GRPCRequestDuration.WithLabelValues(
			serviceName,
			info.FullMethod,
			statusCode,
		).Observe(duration)
		
		return resp, err
	}
}
```

### 2.5 Expose /metrics Endpoint

Update each service's `main.go` or HTTP server setup:

```go
package main

import (
	"net/http"
	
	"github.com/prometheus/client_golang/prometheus/promhttp"
	// ... other imports
)

func main() {
	// ... existing setup ...
	
	// Create metrics endpoint
	metricsHandler := promhttp.Handler()
	
	// If using a separate metrics server (recommended for production)
	go func() {
		metricsMux := http.NewServeMux()
		metricsMux.Handle("/metrics", metricsHandler)
		metricsMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		
		log.Printf("Metrics server listening on :9090")
		if err := http.ListenAndServe(":9090", metricsMux); err != nil {
			log.Fatalf("Metrics server failed: %v", err)
		}
	}()
	
	// ... start main application server ...
}
```

### 2.6 Service-Specific Metrics Examples

#### Transaction Service (`internal/metrics/transaction_metrics.go`)

```go
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	TransactionsCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transactions_created_total",
			Help: "Total number of transactions created",
		},
		[]string{"type"}, // BUY or SELL
	)

	TransactionValueTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transaction_value_total",
			Help: "Total transaction value in USD",
		},
		[]string{"type"},
	)

	TransactionProcessingDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "transaction_processing_duration_seconds",
			Help:    "Time spent processing transaction business logic",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
	)

	UserValidationDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "user_validation_duration_seconds",
			Help:    "Time spent validating user via gRPC",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
		},
	)

	AssetValidationDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "asset_validation_duration_seconds",
			Help:    "Time spent validating asset via gRPC",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
		},
	)
)
```

#### Portfolio Service (`internal/metrics/portfolio_metrics.go`)

```go
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	PortfolioHoldingsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "portfolio_holdings_total",
			Help: "Current number of portfolio holdings tracked",
		},
	)

	PortfolioUpdateDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "portfolio_update_duration_seconds",
			Help:    "Time spent updating portfolio holdings",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25},
		},
	)

	NATSEventsProcessedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nats_events_processed_total",
			Help: "Total number of NATS events processed",
		},
		[]string{"event_type", "status"}, // status: success, failed
	)

	NATSEventProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "nats_event_processing_duration_seconds",
			Help:    "Time spent processing NATS events",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
		},
		[]string{"event_type"},
	)
)
```

### 2.7 Instrument Transaction Usecase

Update `internal/usecase/transaction_usecase.go`:

```go
func (uc *transactionUsecase) CreateTransaction(ctx context.Context, userID, symbol, txType string, quantity, price float64, executedAt time.Time) (*domain.Transaction, error) {
	start := time.Now()
	defer func() {
		metrics.TransactionProcessingDuration.Observe(time.Since(start).Seconds())
	}()

	// Validate User
	userValidationStart := time.Now()
	exists, err := uc.userGateway.Exists(ctx, userID)
	metrics.UserValidationDuration.Observe(time.Since(userValidationStart).Seconds())
	if err != nil {
		return nil, fmt.Errorf("failed to validate user: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user %s does not exist", userID)
	}

	// Validate Asset
	assetValidationStart := time.Now()
	exists, err = uc.marketDataGateway.Exists(ctx, symbol)
	metrics.AssetValidationDuration.Observe(time.Since(assetValidationStart).Seconds())
	if err != nil {
		return nil, fmt.Errorf("failed to validate asset: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("asset %s does not exist", symbol)
	}

	tx := &domain.Transaction{
		UserID:        userID,
		Symbol:        symbol,
		Type:          txType,
		Quantity:      quantity,
		PricePerShare: price,
		ExecutedAt:    executedAt,
	}
	
	if err := uc.repo.Create(ctx, tx); err != nil {
		return nil, err
	}

	// Record business metrics
	metrics.TransactionsCreatedTotal.WithLabelValues(txType).Inc()
	metrics.TransactionValueTotal.WithLabelValues(txType).Add(quantity * price)

	// Publish transaction created event
	if err := uc.eventPublisher.PublishTransactionCreated(ctx, tx); err != nil {
		fmt.Printf("failed to publish transaction created event: %v\n", err)
	}

	return tx, nil
}
```

### 2.8 Instrument Portfolio NATS Subscriber

Update `internal/infrastructure/nats_subscriber.go`:

```go
func (s *NATSSubscriber) handleTransactionCreated(msg *nats.Msg) {
	start := time.Now()
	eventType := "transaction.created"
	
	var event TransactionCreatedEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		s.logger.Error("failed to unmarshal transaction created event", "error", err)
		metrics.NATSEventsProcessedTotal.WithLabelValues(eventType, "failed").Inc()
		return
	}

	s.logger.Info("Received transaction created event",
		"transaction_id", event.TransactionID,
		"user_id", event.UserID,
		"symbol", event.AssetSymbol,
		"type", event.Type,
		"quantity", event.Quantity,
	)

	// ... existing holding update logic ...

	// Save updated holding
	if err := s.repo.Upsert(holding); err != nil {
		s.logger.Error("failed to update holding", "error", err, "user_id", event.UserID, "symbol", event.AssetSymbol)
		metrics.NATSEventsProcessedTotal.WithLabelValues(eventType, "failed").Inc()
		return
	}

	// Record success metrics
	metrics.NATSEventsProcessedTotal.WithLabelValues(eventType, "success").Inc()
	metrics.NATSEventProcessingDuration.WithLabelValues(eventType).Observe(time.Since(start).Seconds())
	
	// Update holdings gauge (you'll need to implement a count method)
	// metrics.PortfolioHoldingsTotal.Set(float64(s.repo.Count()))

	s.logger.Info("Updated portfolio holding",
		"user_id", event.UserID,
		"symbol", event.AssetSymbol,
		"new_quantity", holding.Quantity,
		"average_cost", holding.AverageCost,
	)
}
```

---

## Phase 3: Prometheus Configuration and Alerting

### 3.1 Advanced Scraping Configuration

For Kubernetes environments, create `deployments/monitoring/prometheus/prometheus-k8s.yml`:

```yaml
scrape_configs:
  # Kubernetes pod discovery
  - job_name: 'kubernetes-pods'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      # Only scrape pods with prometheus.io/scrape annotation
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      
      # Use custom metrics path if specified
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      
      # Use custom port if specified
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
        target_label: __address__
      
      # Add service label from pod label
      - source_labels: [__meta_kubernetes_pod_label_app]
        action: replace
        target_label: service
      
      # Add namespace label
      - source_labels: [__meta_kubernetes_namespace]
        action: replace
        target_label: namespace
      
      # Add pod name label
      - source_labels: [__meta_kubernetes_pod_name]
        action: replace
        target_label: pod
```

### 3.2 Alerting Rules

Create `deployments/monitoring/prometheus/alerts.yml`:

```yaml
groups:
  - name: service_health
    interval: 30s
    rules:
      # High Error Rate Alert
      - alert: HighErrorRate
        expr: |
          (
            sum(rate(http_requests_total{status_code=~"5.."}[5m])) by (service)
            /
            sum(rate(http_requests_total[5m])) by (service)
          ) > 0.01
        for: 5m
        labels:
          severity: critical
          team: backend
        annotations:
          summary: "High error rate on {{ $labels.service }}"
          description: "Service {{ $labels.service }} has error rate of {{ $value | humanizePercentage }} (threshold: 1%)"

      # High Latency Alert (P95 > 1s)
      - alert: HighLatency
        expr: |
          histogram_quantile(0.95,
            sum(rate(http_request_duration_seconds_bucket[5m])) by (service, le)
          ) > 1
        for: 10m
        labels:
          severity: warning
          team: backend
        annotations:
          summary: "High latency on {{ $labels.service }}"
          description: "Service {{ $labels.service }} P95 latency is {{ $value }}s (threshold: 1s)"

      # Service Down Alert
      - alert: ServiceDown
        expr: up{job=~".*-service"} == 0
        for: 2m
        labels:
          severity: critical
          team: backend
        annotations:
          summary: "Service {{ $labels.job }} is down"
          description: "Service {{ $labels.job }} has been down for more than 2 minutes"

      # High Memory Usage
      - alert: HighMemoryUsage
        expr: |
          (
            go_memstats_heap_alloc_bytes
            /
            go_memstats_heap_sys_bytes
          ) > 0.9
        for: 15m
        labels:
          severity: warning
          team: backend
        annotations:
          summary: "High memory usage on {{ $labels.service }}"
          description: "Service {{ $labels.service }} is using {{ $value | humanizePercentage }} of heap (threshold: 90%)"

      # Goroutine Leak Detection
      - alert: GoroutineLeak
        expr: |
          (
            go_goroutines > 1000
            and
            rate(go_goroutines[10m]) > 10
          )
        for: 15m
        labels:
          severity: warning
          team: backend
        annotations:
          summary: "Potential goroutine leak on {{ $labels.service }}"
          description: "Service {{ $labels.service }} has {{ $value }} goroutines and growing"

  - name: business_metrics
    interval: 1m
    rules:
      # Transaction Processing Failures
      - alert: HighTransactionFailureRate
        expr: |
          (
            sum(rate(grpc_requests_total{method=~".*CreateTransaction", status!="OK"}[5m]))
            /
            sum(rate(grpc_requests_total{method=~".*CreateTransaction"}[5m]))
          ) > 0.05
        for: 5m
        labels:
          severity: critical
          team: backend
        annotations:
          summary: "High transaction failure rate"
          description: "Transaction creation failure rate is {{ $value | humanizePercentage }} (threshold: 5%)"

      # NATS Event Processing Failures
      - alert: HighNATSEventFailureRate
        expr: |
          (
            sum(rate(nats_events_processed_total{status="failed"}[5m]))
            /
            sum(rate(nats_events_processed_total[5m]))
          ) > 0.05
        for: 5m
        labels:
          severity: warning
          team: backend
        annotations:
          summary: "High NATS event processing failure rate"
          description: "NATS event failure rate is {{ $value | humanizePercentage }} (threshold: 5%)"

      # Slow gRPC Calls
      - alert: SlowGRPCCalls
        expr: |
          histogram_quantile(0.95,
            sum(rate(grpc_request_duration_seconds_bucket[5m])) by (service, method, le)
          ) > 0.5
        for: 10m
        labels:
          severity: warning
          team: backend
        annotations:
          summary: "Slow gRPC calls on {{ $labels.service }}"
          description: "Method {{ $labels.method }} P95 latency is {{ $value }}s (threshold: 0.5s)"
```

### 3.3 AlertManager Configuration

Create `deployments/monitoring/alertmanager/alertmanager.yml`:

```yaml
global:
  resolve_timeout: 5m
  slack_api_url: 'YOUR_SLACK_WEBHOOK_URL'

route:
  group_by: ['alertname', 'service']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h
  receiver: 'default'
  routes:
    - match:
        severity: critical
      receiver: 'critical-alerts'
      continue: true
    - match:
        severity: warning
      receiver: 'warning-alerts'

receivers:
  - name: 'default'
    slack_configs:
      - channel: '#alerts'
        title: 'Portfolio Insights Alert'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'

  - name: 'critical-alerts'
    slack_configs:
      - channel: '#critical-alerts'
        title: '🚨 CRITICAL: {{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
        send_resolved: true

  - name: 'warning-alerts'
    slack_configs:
      - channel: '#alerts'
        title: '⚠️ WARNING: {{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
        send_resolved: true

inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'service']
```

---

## Phase 4: Grafana Visualization and Dashboarding

### 4.1 Dashboard Provisioning

Create `deployments/monitoring/grafana/provisioning/dashboards/dashboard.yml`:

```yaml
apiVersion: 1

providers:
  - name: 'Default'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /var/lib/grafana/dashboards
```

### 4.2 Golden Signals Dashboard

Create `deployments/monitoring/grafana/dashboards/golden-signals.json`:

This is a comprehensive dashboard JSON. Key panels include:

**Panel 1: Request Rate (Traffic)**
```json
{
  "title": "Request Rate by Service",
  "targets": [
    {
      "expr": "sum(rate(http_requests_total[5m])) by (service)",
      "legendFormat": "{{ service }}"
    }
  ],
  "type": "graph"
}
```

**Panel 2: Error Rate**
```json
{
  "title": "Error Rate by Service",
  "targets": [
    {
      "expr": "sum(rate(http_requests_total{status_code=~\"5..\"}[5m])) by (service) / sum(rate(http_requests_total[5m])) by (service)",
      "legendFormat": "{{ service }}"
    }
  ],
  "type": "graph",
  "yaxes": [
    {
      "format": "percentunit"
    }
  ]
}
```

**Panel 3: Latency (P50, P95, P99)**
```json
{
  "title": "Request Latency by Service",
  "targets": [
    {
      "expr": "histogram_quantile(0.50, sum(rate(http_request_duration_seconds_bucket[5m])) by (service, le))",
      "legendFormat": "{{ service }} P50"
    },
    {
      "expr": "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (service, le))",
      "legendFormat": "{{ service }} P95"
    },
    {
      "expr": "histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket[5m])) by (service, le))",
      "legendFormat": "{{ service }} P99"
    }
  ],
  "type": "graph"
}
```

**Panel 4: Saturation (In-Flight Requests)**
```json
{
  "title": "In-Flight Requests",
  "targets": [
    {
      "expr": "sum(http_in_flight_requests) by (service)",
      "legendFormat": "{{ service }}"
    }
  ],
  "type": "graph"
}
```

### 4.3 Go Runtime Dashboard

Key PromQL queries for Go runtime monitoring:

**Goroutines:**
```promql
go_goroutines{service="$service"}
```

**Heap Memory:**
```promql
go_memstats_heap_alloc_bytes{service="$service"}
go_memstats_heap_inuse_bytes{service="$service"}
go_memstats_heap_sys_bytes{service="$service"}
```

**GC Metrics:**
```promql
rate(go_gc_duration_seconds_sum[5m])
rate(go_memstats_gc_sys_bytes[5m])
```

**Memory Allocation Rate:**
```promql
rate(go_memstats_alloc_bytes_total[5m])
```

### 4.4 Business Metrics Dashboard

**Transaction Volume:**
```promql
sum(rate(transactions_created_total[5m])) by (type)
```

**Transaction Value:**
```promql
sum(rate(transaction_value_total[5m])) by (type)
```

**Portfolio Holdings:**
```promql
portfolio_holdings_total
```

**NATS Event Processing:**
```promql
sum(rate(nats_events_processed_total[5m])) by (event_type, status)
```

### 4.5 Dashboard Templating

Add variables to dashboards for dynamic filtering:

```json
{
  "templating": {
    "list": [
      {
        "name": "service",
        "type": "query",
        "datasource": "Prometheus",
        "query": "label_values(http_requests_total, service)",
        "multi": true,
        "includeAll": true
      },
      {
        "name": "interval",
        "type": "interval",
        "query": "1m,5m,10m,30m,1h",
        "auto": true
      }
    ]
  }
}
```

---

## Phase 5: Testing, Validation, and Rollout

### 5.1 End-to-End Testing

#### Test Script (`scripts/test-metrics.sh`)

```bash
#!/bin/bash

echo "=== Testing Metrics Endpoints ==="

services=(
  "gateway:8080"
  "user-service:8081"
  "transaction-service:8082"
  "portfolio-service:8083"
  "marketdata-service:8084"
)

for service in "${services[@]}"; do
  IFS=':' read -r name port <<< "$service"
  echo ""
  echo "Testing $name on port $port..."
  
  # Check if metrics endpoint is accessible
  if curl -s "http://localhost:$port/metrics" > /dev/null; then
    echo "✓ Metrics endpoint accessible"
    
    # Check for standard Go metrics
    if curl -s "http://localhost:$port/metrics" | grep -q "go_goroutines"; then
      echo "✓ Go runtime metrics present"
    else
      echo "✗ Go runtime metrics missing"
    fi
    
    # Check for HTTP metrics
    if curl -s "http://localhost:$port/metrics" | grep -q "http_requests_total"; then
      echo "✓ HTTP metrics present"
    else
      echo "✗ HTTP metrics missing"
    fi
  else
    echo "✗ Metrics endpoint not accessible"
  fi
done

echo ""
echo "=== Testing Prometheus Scraping ==="
if curl -s "http://localhost:9090/api/v1/targets" | jq -r '.data.activeTargets[] | "\(.labels.job): \(.health)"'; then
  echo "✓ Prometheus targets status retrieved"
else
  echo "✗ Failed to retrieve Prometheus targets"
fi

echo ""
echo "=== Testing Grafana Datasource ==="
if curl -s -u admin:admin "http://localhost:3001/api/datasources" | jq -r '.[].name'; then
  echo "✓ Grafana datasources configured"
else
  echo "✗ Failed to retrieve Grafana datasources"
fi
```

#### Load Testing Script (`scripts/load-test.sh`)

```bash
#!/bin/bash

# Generate load to test metrics collection
echo "Generating load on transaction service..."

for i in {1..100}; do
  curl -X POST http://localhost:8082/api/transactions \
    -H "Content-Type: application/json" \
    -d '{
      "user_id": "user-123",
      "symbol": "AAPL",
      "type": "BUY",
      "quantity": 10,
      "price_per_share": 150.00
    }' &
done

wait
echo "Load generation complete. Check Grafana dashboards."
```

### 5.2 Performance Baseline

#### Benchmark Script (`scripts/benchmark-metrics.sh`)

```bash
#!/bin/bash

echo "=== Benchmarking Metrics Overhead ==="

# Test without metrics
echo "Running baseline test (no metrics)..."
ab -n 10000 -c 100 http://localhost:8082/health > baseline.txt

# Test with metrics
echo "Running test with metrics enabled..."
ab -n 10000 -c 100 http://localhost:8082/metrics > with-metrics.txt

# Compare results
echo ""
echo "Baseline (no metrics):"
grep "Requests per second" baseline.txt
grep "Time per request" baseline.txt

echo ""
echo "With metrics:"
grep "Requests per second" with-metrics.txt
grep "Time per request" with-metrics.txt

# Calculate overhead
echo ""
echo "Expected overhead: < 5% latency increase"
```

### 5.3 Rollout Strategy

#### Phase 5.3.1: Development Environment

1. **Deploy monitoring stack:**
```bash
cd deployments/monitoring
docker-compose up -d
```

2. **Verify Prometheus targets:**
   - Navigate to http://localhost:9090/targets
   - Ensure all services show as "UP"

3. **Import Grafana dashboards:**
   - Navigate to http://localhost:3001 (admin/admin)
   - Verify datasource connection
   - Import golden signals dashboard

#### Phase 5.3.2: Staging Environment

1. **Deploy with Kubernetes:**
```bash
kubectl apply -f deployments/k8s/monitoring/
```

2. **Validate service discovery:**
```bash
kubectl get servicemonitors -n monitoring
```

3. **Run integration tests:**
```bash
./scripts/test-metrics.sh
./scripts/load-test.sh
```

#### Phase 5.3.3: Production Rollout

**Week 1: Canary Deployment**
- Deploy instrumented services to 10% of traffic
- Monitor for performance degradation
- Validate metrics accuracy

**Week 2: Gradual Rollout**
- Increase to 50% of traffic
- Set up alerting rules
- Train team on dashboards

**Week 3: Full Deployment**
- Deploy to 100% of traffic
- Enable all alerting rules
- Document runbooks

### 5.4 Validation Checklist

- [ ] All services expose `/metrics` endpoint
- [ ] Prometheus successfully scrapes all targets
- [ ] Grafana datasource connected to Prometheus
- [ ] Golden signals dashboard shows data for all services
- [ ] Go runtime dashboard shows goroutines and memory metrics
- [ ] Business metrics dashboard shows transaction and portfolio data
- [ ] Alert rules are loaded in Prometheus
- [ ] AlertManager is configured and sending test alerts
- [ ] Performance overhead is < 5%
- [ ] Team is trained on using dashboards
- [ ] Runbooks are documented for common alerts

---

## Appendix: Quick Start Commands

### Start Monitoring Stack
```bash
cd deployments/monitoring
docker-compose up -d
```

### View Logs
```bash
docker-compose logs -f prometheus
docker-compose logs -f grafana
```

### Reload Prometheus Config
```bash
curl -X POST http://localhost:9090/-/reload
```

### Test Metrics Endpoint
```bash
curl http://localhost:8082/metrics
```

### Query Prometheus
```bash
# Request rate
curl 'http://localhost:9090/api/v1/query?query=rate(http_requests_total[5m])'

# Error rate
curl 'http://localhost:9090/api/v1/query?query=rate(http_requests_total{status_code=~"5.."}[5m])'
```

### Access UIs
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3001 (admin/admin)
- AlertManager: http://localhost:9093

---

## Next Steps

1. **Implement Phase 2** - Start with instrumenting one service (transaction-service)
2. **Create metrics package** - Build reusable metrics library
3. **Add middleware** - Implement HTTP and gRPC interceptors
4. **Deploy monitoring stack** - Use Docker Compose for local testing
5. **Create dashboards** - Build golden signals and Go runtime dashboards
6. **Set up alerts** - Configure critical alerting rules
7. **Load test** - Validate performance overhead
8. **Roll out** - Deploy to production incrementally
