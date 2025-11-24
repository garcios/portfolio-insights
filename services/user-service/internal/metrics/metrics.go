package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// gRPC Metrics
	GrpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)

	GrpcRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "user_grpc_request_duration_seconds",
			Help:    "Duration of gRPC requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	// Database Metrics
	DatabaseQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "user_database_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"operation", "table"},
	)

	DatabaseErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_database_errors_total",
			Help: "Total number of database errors",
		},
		[]string{"operation", "table"},
	)

	// Business Metrics
	UsersCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "user_users_created_total",
			Help: "Total number of users created",
		},
	)

	TotalUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "user_total_users",
			Help: "Total number of users in the system",
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

// RecordUserCreated records a new user creation
func RecordUserCreated() {
	UsersCreatedTotal.Inc()
}
