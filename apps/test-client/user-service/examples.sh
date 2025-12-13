#!/bin/bash

# Example usage script for user-service test client
# This script demonstrates how to use the test client to interact with the user-service

set -e

# Configuration
SERVER_ADDR="${SERVER_ADDR:-localhost:50051}"
CLIENT="./user-client"

echo "=== User Service Test Client Examples ==="
echo "Server: $SERVER_ADDR"
echo ""

# Check if client binary exists
if [ ! -f "$CLIENT" ]; then
    echo "Building client..."
    go build -o user-client
    echo ""
fi

# Example 1: Create a user and capture the user ID
echo "1. Creating a new user..."
CREATE_OUTPUT=$($CLIENT -addr "$SERVER_ADDR" \
    -op create \
    -email "test@example.com" \
    -username "testuser" \
    -password "testpass123" 2>&1)

echo "$CREATE_OUTPUT"
echo ""

# Extract the User ID from the output
USER_ID=$(echo "$CREATE_OUTPUT" | grep "User ID:" | awk '{print $3}')

if [ -z "$USER_ID" ]; then
    echo "Error: Failed to extract user ID from create operation"
    exit 1
fi

# Example 2: Get the user using the captured ID
echo "2. Retrieving user with ID $USER_ID..."
$CLIENT -addr "$SERVER_ADDR" \
    -op get \
    -user-id "$USER_ID"
echo ""

# Example 3: Verify user credentials (valid)
echo "3. Verifying valid credentials..."
$CLIENT -addr "$SERVER_ADDR" \
    -op verify \
    -email "test@example.com" \
    -password "testpass123"
echo ""

# Example 4: Verify user credentials (invalid)
echo "4. Verifying invalid credentials..."
$CLIENT -addr "$SERVER_ADDR" \
    -op verify \
    -email "test@example.com" \
    -password "wrongpassword" || true
echo ""

echo "=== Examples completed ==="
