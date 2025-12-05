// Package middleware implements middleware for the gateway.
package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/garcios/portfolio-insights/apps/gateway/internal/metrics"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// NewResponseWriter creates a new response writer.
func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// MetricsMiddleware records metrics for each request.
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := NewResponseWriter(w)
		next.ServeHTTP(rw, r)
		duration := time.Since(start).Seconds()

		path := r.URL.Path

		metrics.RecordHttpRequest(r.Method, path, strconv.Itoa(rw.statusCode), duration)
	})
}
