package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// gRPC Metrics
	GrpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "portfolio_grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)

	GrpcRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "portfolio_grpc_request_duration_seconds",
			Help:    "Duration of gRPC requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	// Business Metrics
	HoldingsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "portfolio_holdings_total",
			Help: "Total number of holdings across all users",
		},
	)

	HoldingsByUser = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "portfolio_holdings_by_user",
			Help: "Number of holdings per user",
		},
		[]string{"user_id"},
	)

	PortfolioValue = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "portfolio_total_value",
			Help: "Total portfolio value in USD",
		},
		[]string{"user_id"},
	)

	// Cache Metrics
	CacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "portfolio_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type"},
	)

	CacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "portfolio_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_type"},
	)

	CacheOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "portfolio_cache_operation_duration_seconds",
			Help:    "Duration of cache operations in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"operation", "cache_type"},
	)

	// Database Metrics
	DatabaseQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "portfolio_database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "portfolio_database_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"operation", "table"},
	)

	DatabaseErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "portfolio_database_errors_total",
			Help: "Total number of database errors",
		},
		[]string{"operation", "table"},
	)

	// Market Data Metrics
	MarketDataRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "portfolio_marketdata_requests_total",
			Help: "Total number of market data requests",
		},
		[]string{"operation", "status"},
	)

	MarketDataRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "portfolio_marketdata_request_duration_seconds",
			Help:    "Duration of market data requests in seconds",
			Buckets: []float64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		},
		[]string{"operation"},
	)

	PricesFetched = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "portfolio_prices_fetched_total",
			Help: "Total number of prices fetched",
		},
		[]string{"source"}, // "cache" or "service"
	)

	// NATS Metrics
	NatsMessagesConsumed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "portfolio_nats_messages_consumed_total",
			Help: "Total number of NATS messages consumed",
		},
		[]string{"subject", "status"},
	)

	NatsMessageProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "portfolio_nats_message_processing_duration_seconds",
			Help:    "Duration of NATS message processing in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"subject"},
	)

	// Error Metrics
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "portfolio_errors_total",
			Help: "Total number of errors",
		},
		[]string{"component", "error_type"},
	)
)

// RecordGrpcRequest records a gRPC request
func RecordGrpcRequest(method, status string, duration float64) {
	GrpcRequestsTotal.WithLabelValues(method, status).Inc()
	GrpcRequestDuration.WithLabelValues(method).Observe(duration)
}

// RecordCacheOperation records a cache operation
func RecordCacheOperation(operation, cacheType string, hit bool, duration float64) {
	if hit {
		CacheHitsTotal.WithLabelValues(cacheType).Inc()
	} else {
		CacheMissesTotal.WithLabelValues(cacheType).Inc()
	}
	CacheOperationDuration.WithLabelValues(operation, cacheType).Observe(duration)
}

// RecordDatabaseQuery records a database query
func RecordDatabaseQuery(operation, table string, duration float64, err error) {
	DatabaseQueriesTotal.WithLabelValues(operation, table).Inc()
	DatabaseQueryDuration.WithLabelValues(operation, table).Observe(duration)
	if err != nil {
		DatabaseErrorsTotal.WithLabelValues(operation, table).Inc()
	}
}

// RecordMarketDataRequest records a market data request
func RecordMarketDataRequest(operation, status string, duration float64) {
	MarketDataRequestsTotal.WithLabelValues(operation, status).Inc()
	MarketDataRequestDuration.WithLabelValues(operation).Observe(duration)
}

// RecordPriceFetch records a price fetch
func RecordPriceFetch(source string, count int) {
	PricesFetched.WithLabelValues(source).Add(float64(count))
}

// RecordNatsMessage records a NATS message
func RecordNatsMessage(subject, status string, duration float64) {
	NatsMessagesConsumed.WithLabelValues(subject, status).Inc()
	NatsMessageProcessingDuration.WithLabelValues(subject).Observe(duration)
}

// RecordError records an error
func RecordError(component, errorType string) {
	ErrorsTotal.WithLabelValues(component, errorType).Inc()
}

// UpdateHoldingsMetrics updates holdings-related metrics
func UpdateHoldingsMetrics(totalHoldings int, userHoldings map[string]int, userValues map[string]float64) {
	HoldingsTotal.Set(float64(totalHoldings))

	for userID, count := range userHoldings {
		HoldingsByUser.WithLabelValues(userID).Set(float64(count))
	}

	for userID, value := range userValues {
		PortfolioValue.WithLabelValues(userID).Set(value)
	}
}
