package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMetricsRegistration(t *testing.T) {
	// Test that metrics are properly registered
	if HttpRequestsTotal == nil {
		t.Error("HttpRequestsTotal metric is nil")
	}
	if HttpRequestDuration == nil {
		t.Error("HttpRequestDuration metric is nil")
	}
}

func TestRecordHttpRequest(t *testing.T) {
	// Reset metrics
	HttpRequestsTotal.Reset()
	HttpRequestDuration.Reset()

	// Record a request
	RecordHttpRequest("GET", "/query", "200", 0.5)

	// Verify counter was incremented
	count := testutil.CollectAndCount(HttpRequestsTotal)
	if count == 0 {
		t.Error("Expected HttpRequestsTotal to have metrics, got 0")
	}

	// Verify histogram was updated
	histCount := testutil.CollectAndCount(HttpRequestDuration)
	if histCount == 0 {
		t.Error("Expected HttpRequestDuration to have metrics, got 0")
	}
}
