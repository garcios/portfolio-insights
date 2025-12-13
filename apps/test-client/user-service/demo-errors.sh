#!/bin/bash

# Demonstration script showing error handling and detailed error messages
# This script demonstrates the enhanced error testing capabilities

set -e

SERVER_ADDR="${SERVER_ADDR:-localhost:50051}"
CLIENT="./user-client"

echo "=== User Service Test Client - Error Handling Demo ==="
echo "Server: $SERVER_ADDR"
echo ""

# Check if client binary exists
if [ ! -f "$CLIENT" ]; then
    echo "Building client..."
    go build -o user-client
    echo ""
fi

echo "1. Testing with invalid user ID (non-UUID format)..."
echo "   Command: ./user-client -op get -user-id 123"
$CLIENT -addr "$SERVER_ADDR" -op get -user-id "123" 2>&1 || true
echo ""

echo "2. Testing with empty resource name..."
echo "   Command: ./user-client -op get -user-id ''"
echo "   (This will fail client-side validation)"
echo ""

echo "3. Testing with non-existent UUID..."
echo "   Command: ./user-client -op get -user-id 00000000-0000-0000-0000-000000000000"
$CLIENT -addr "$SERVER_ADDR" -op get -user-id "00000000-0000-0000-0000-000000000000" 2>&1 || true
echo ""

echo "4. Running comprehensive error test suite..."
echo "   Command: ./user-client -op test-errors"
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
