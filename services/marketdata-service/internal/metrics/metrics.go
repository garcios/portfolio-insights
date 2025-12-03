package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// gRPC Metrics
	GrpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "marketdata_grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)

	GrpcRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "marketdata_grpc_request_duration_seconds",
			Help:    "Duration of gRPC requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	// Database Metrics
	DatabaseQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "marketdata_database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "marketdata_database_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"operation", "table"},
	)

	DatabaseErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "marketdata_database_errors_total",
			Help: "Total number of database errors",
		},
		[]string{"operation", "table"},
	)

	// Ingestion Metrics
	IngestionJobsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "marketdata_ingestion_jobs_total",
			Help: "Total number of ingestion jobs",
		},
		[]string{"type", "status"},
	)

	IngestionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "marketdata_ingestion_duration_seconds",
			Help:    "Duration of ingestion jobs in seconds",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
		},
		[]string{"type"},
	)

	PricesIngestedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "marketdata_prices_ingested_total",
			Help: "Total number of price records ingested",
		},
	)

	CurrenciesIngestedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "marketdata_currencies_ingested_total",
			Help: "Total number of currency rates ingested",
		},
	)

	// Business Metrics
	TotalAssets = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "marketdata_total_assets",
			Help: "Total number of assets in the system",
		},
	)

	TotalPrices = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "marketdata_total_prices",
			Help: "Total number of price records in the system",
		},
	)
)

// RecordGrpcRequest records a gRPC request
func RecordGrpcRequest(method, status string, duration float64) {
	GrpcRequestsTotal.WithLabelValues(method, status).Inc()
	GrpcRequestDuration.WithLabelValues(method).Observe(duration)
}

// RecordDatabaseQuery records a database query
func RecordDatabaseQuery(operation, table string, duration float64, err error) {
	DatabaseQueriesTotal.WithLabelValues(operation, table).Inc()
	DatabaseQueryDuration.WithLabelValues(operation, table).Observe(duration)
	if err != nil {
		DatabaseErrorsTotal.WithLabelValues(operation, table).Inc()
	}
}

// RecordIngestionJob records an ingestion job
func RecordIngestionJob(jobType, status string, duration float64) {
	IngestionJobsTotal.WithLabelValues(jobType, status).Inc()
	IngestionDuration.WithLabelValues(jobType).Observe(duration)
}

// RecordPricesIngested records the number of prices ingested
func RecordPricesIngested(count int) {
	PricesIngestedTotal.Add(float64(count))
}

// RecordCurrenciesIngested records the number of currency rates ingested
func RecordCurrenciesIngested(count int) {
	CurrenciesIngestedTotal.Add(float64(count))
}
