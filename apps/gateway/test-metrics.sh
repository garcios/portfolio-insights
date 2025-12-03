#!/bin/bash
# Test Gateway Prometheus Metrics
# This script verifies that the gateway is exposing Prometheus metrics correctly

set -e

echo "🔍 Testing Gateway Prometheus Metrics..."
echo ""

# Check if gateway is running
echo "1. Checking if gateway is accessible..."
if curl -s http://localhost:8080 > /dev/null 2>&1; then
    echo "   ✅ Gateway is running on port 8080"
else
    echo "   ❌ Gateway is not accessible on port 8080"
    echo "   Please start the gateway service first"
    exit 1
fi

# Check metrics endpoint
echo ""
echo "2. Checking /metrics endpoint..."
if curl -s http://localhost:9095/metrics > /dev/null 2>&1; then
    echo "   ✅ Metrics endpoint is accessible on port 9095"
else
    echo "   ❌ Metrics endpoint is not accessible on port 9095"
    echo "   Make sure the gateway is exposing port 9095"
    exit 1
fi

# Check for gateway-specific metrics
echo ""
echo "3. Checking for gateway HTTP metrics..."
METRICS=$(curl -s http://localhost:9095/metrics)

if echo "$METRICS" | grep -q "gateway_http_requests_total"; then
    echo "   ✅ Found gateway_http_requests_total metric"
else
    echo "   ⚠️  gateway_http_requests_total metric not found"
fi

if echo "$METRICS" | grep -q "gateway_http_request_duration_seconds"; then
    echo "   ✅ Found gateway_http_request_duration_seconds metric"
else
    echo "   ⚠️  gateway_http_request_duration_seconds metric not found"
fi

# Check for standard Go metrics
echo ""
echo "4. Checking for standard Go metrics..."
if echo "$METRICS" | grep -q "go_goroutines"; then
    echo "   ✅ Found go_goroutines metric"
else
    echo "   ⚠️  go_goroutines metric not found"
fi

if echo "$METRICS" | grep -q "process_cpu_seconds_total"; then
    echo "   ✅ Found process_cpu_seconds_total metric"
else
    echo "   ⚠️  process_cpu_seconds_total metric not found"
fi

# Generate some traffic to create metrics
echo ""
echo "5. Generating test traffic..."
for i in {1..5}; do
    curl -s http://localhost:8080 > /dev/null 2>&1 || true
done
echo "   ✅ Sent 5 test requests"

# Check if metrics were recorded
echo ""
echo "6. Verifying metrics were recorded..."
METRICS_AFTER=$(curl -s http://localhost:9095/metrics)

if echo "$METRICS_AFTER" | grep "gateway_http_requests_total" | grep -q "path=\"/\""; then
    echo "   ✅ HTTP requests are being tracked"
    echo ""
    echo "Sample metrics:"
    echo "$METRICS_AFTER" | grep "gateway_http_requests_total" | head -3
else
    echo "   ⚠️  No HTTP requests tracked yet"
fi

echo ""
echo "✅ Gateway Prometheus metrics test complete!"
echo ""
echo "📊 View metrics at: http://localhost:9095/metrics"
echo "📈 View in Prometheus: http://localhost:9081"
echo "📉 View in Grafana: http://localhost:3001"
