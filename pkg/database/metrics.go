package database

import (
	"database/sql"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// queriesTotal counts the total number of database queries.
	queriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	// queryDuration histograms the duration of database queries in seconds.
	queryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"operation", "table"},
	)

	// errorsTotal counts the total number of database errors.
	errorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_errors_total",
			Help: "Total number of database errors",
		},
		[]string{"operation", "table"},
	)

	// rowsAffectedTotal counts the total number of rows affected by database operations.
	rowsAffectedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_rows_affected_total",
			Help: "Total number of rows affected by database operations",
		},
		[]string{"operation", "table"},
	)

	// dbConnectionsOpen reports the number of established connections both in use and idle.
	dbConnectionsOpen = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_connections_open",
			Help: "Number of established connections both in use and idle",
		},
		[]string{"db_name"},
	)

	// dbConnectionsInUse reports the number of connections currently in use.
	dbConnectionsInUse = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_connections_in_use",
			Help: "Number of connections currently in use",
		},
		[]string{"db_name"},
	)

	// dbConnectionsIdle reports the number of idle connections.
	dbConnectionsIdle = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Number of idle connections",
		},
		[]string{"db_name"},
	)

	// dbConnectionsWaitCount reports the total number of connections waited for.
	dbConnectionsWaitCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_connections_wait_count",
			Help: "Total number of connections waited for",
		},
		[]string{"db_name"},
	)

	// dbConnectionsWaitDuration reports the total time blocked waiting for a new connection.
	dbConnectionsWaitDuration = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_connections_wait_duration_seconds",
			Help: "Total time blocked waiting for a new connection",
		},
		[]string{"db_name"},
	)

	// dbConnectionsMaxOpen reports the maximum number of open connections to the database.
	dbConnectionsMaxOpen = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_connections_max_open",
			Help: "Maximum number of open connections to the database",
		},
		[]string{"db_name"},
	)
)

// RecordQuery records a database query metric.
func RecordQuery(operation, table string, duration float64, err error) {
	queriesTotal.WithLabelValues(operation, table).Inc()
	queryDuration.WithLabelValues(operation, table).Observe(duration)
	if err != nil {
		errorsTotal.WithLabelValues(operation, table).Inc()
	}
}

// RecordRowsAffected records the number of rows affected by a database operation.
func RecordRowsAffected(operation, table string, count int64) {
	if count > 0 {
		rowsAffectedTotal.WithLabelValues(operation, table).Add(float64(count))
	}
}

// StartMonitoring starts a goroutine that periodically records database connection metrics.
func StartMonitoring(db *sql.DB, dbName string) {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			stats := db.Stats()
			dbConnectionsOpen.WithLabelValues(dbName).Set(float64(stats.OpenConnections))
			dbConnectionsInUse.WithLabelValues(dbName).Set(float64(stats.InUse))
			dbConnectionsIdle.WithLabelValues(dbName).Set(float64(stats.Idle))
			dbConnectionsWaitCount.WithLabelValues(dbName).Set(float64(stats.WaitCount))
			dbConnectionsWaitDuration.WithLabelValues(dbName).Set(stats.WaitDuration.Seconds())
			dbConnectionsMaxOpen.WithLabelValues(dbName).Set(float64(stats.MaxOpenConnections))
		}
	}()
}
