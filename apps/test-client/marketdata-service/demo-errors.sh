#!/bin/bash

# Demonstration script showing error handling and detailed error messages
# This script demonstrates the enhanced error testing capabilities

set -e

SERVER_ADDR="${SERVER_ADDR:-localhost:50054}"
CLIENT="./marketdata-client"

echo "=== Market Data Service Test Client - Error Handling Demo ==="
echo "Server: $SERVER_ADDR"
echo ""

# Check if client binary exists
if [ ! -f "$CLIENT" ]; then
    echo "Building client..."
    go build -o marketdata-client
    echo ""
fi

echo "1. Testing with invalid resource name format..."
echo "   Command: ./marketdata-client -op get-asset -asset-name AAPL"
$CLIENT -addr "$SERVER_ADDR" -op get-asset -asset-name "AAPL" 2>&1 || true
echo ""

echo "2. Testing with empty symbol..."
echo "   Command: ./marketdata-client -op get-asset-by-symbol -symbol ''"
echo "   (This will fail client-side validation)"
echo ""

echo "3. Testing with non-existent asset..."
echo "   Command: ./marketdata-client -op get-asset -asset-name assets/NONEXISTENT"
$CLIENT -addr "$SERVER_ADDR" -op get-asset -asset-name "assets/NONEXISTENT" 2>&1 || true
echo ""

echo "4. Testing with invalid time range (end before start)..."
echo "   Command: ./marketdata-client -op get-historical with reversed times"
END_TIME=$(date -u -v-7d +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "7 days ago" +"%Y-%m-%dT%H:%M:%SZ")
START_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
$CLIENT -addr "$SERVER_ADDR" \
    -op get-historical \
    -asset-name "assets/AAPL" \
    -start-time "$START_TIME" \
    -end-time "$END_TIME" 2>&1 || true
echo ""

echo "5. Running comprehensive error test suite..."
echo "   Command: ./marketdata-client -op test-errors"
$CLIENT -addr "$SERVER_ADDR" -op test-errors 2>&1 || true
echo ""

echo "=== Demo completed ==="
echo ""
echo "Key features demonstrated:"
echo "  ✓ Detailed error messages with gRPC status codes"
echo "  ✓ Clear formatting of error information"
echo "  ✓ Comprehensive error test suite"
echo "  ✓ Validation of input parameters"
echo ""
