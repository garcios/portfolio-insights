#!/bin/bash

# Demonstration script showing error handling and detailed error messages
# This script demonstrates the enhanced error testing capabilities

set -e

SERVER_ADDR="${SERVER_ADDR:-localhost:50053}"
CLIENT="./transaction-client"

echo "=== Transaction Service Test Client - Error Handling Demo ==="
echo "Server: $SERVER_ADDR"
echo ""

# Check if client binary exists
if [ ! -f "$CLIENT" ]; then
    echo "Building client..."
    go build -o transaction-client
    echo ""
fi

echo "1. Testing with missing required field (type)..."
echo "   Command: ./transaction-client -op create -user-id test-user -executed-at 2024-01-15T10:30:00Z"
$CLIENT -addr "$SERVER_ADDR" -op create -user-id test-user -executed-at "2024-01-15T10:30:00Z" 2>&1 || true
echo ""

echo "2. Testing with invalid parent format..."
echo "   Command: ./transaction-client -op list -user-id ''"
echo "   (This will fail client-side validation)"
echo ""

echo "3. Testing with non-existent transaction..."
echo "   Command: ./transaction-client -op get -user-id test-user -transaction-id 00000000-0000-0000-0000-000000000000"
$CLIENT -addr "$SERVER_ADDR" -op get -user-id test-user -transaction-id "00000000-0000-0000-0000-000000000000" 2>&1 || true
echo ""

echo "4. Testing with invalid resource name format..."
echo "   Command: ./transaction-client -op delete -user-id test-user -transaction-id invalid-id"
$CLIENT -addr "$SERVER_ADDR" -op delete -user-id test-user -transaction-id "invalid-id" 2>&1 || true
echo ""

echo "5. Running comprehensive error test suite..."
echo "   Command: ./transaction-client -op test-errors"
$CLIENT -addr "$SERVER_ADDR" -op test-errors 2>&1 || true
echo ""

echo "=== Demo completed ==="
echo ""
echo "Key features demonstrated:"
echo "  ✓ Detailed error messages with gRPC status codes"
echo "  ✓ Clear formatting of error information"
echo "  ✓ Comprehensive error test suite (16 test cases)"
echo "  ✓ Validation of input parameters"
echo "  ✓ User-friendly error responses"
echo ""
