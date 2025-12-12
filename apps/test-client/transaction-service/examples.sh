#!/bin/bash

# Example usage script for transaction-service test client
# This script demonstrates how to use the test client to interact with the transaction-service

set -e

# Configuration
SERVER_ADDR="${SERVER_ADDR:-localhost:50053}"
CLIENT="./transaction-client"
USER_ID="6b2ff832-51d8-44e3-ba7f-ab540af1ab96"

echo "=== Transaction Service Test Client Examples ==="
echo "Server: $SERVER_ADDR"
echo "User ID: $USER_ID"
echo ""

# Check if client binary exists
if [ ! -f "$CLIENT" ]; then
    echo "Building client..."
    go build -o transaction-client
    echo ""
fi

# Example 1: Create a BUY transaction
echo "1. Creating a BUY transaction..."
BUY_OUTPUT=$($CLIENT -addr "$SERVER_ADDR" \
    -op create \
    -user-id "$USER_ID" \
    -type BUY \
    -symbol AMZN \
    -quantity 10 \
    -price 150.50 \
    -executed-at "2024-01-15T10:30:00Z" \
    -brokerage 5.00 \
    -notes "Initial purchase" 2>&1)

echo "$BUY_OUTPUT"
echo ""

# Extract the Transaction ID from the output
BUY_TXN_ID=$(echo "$BUY_OUTPUT" | grep "Transaction ID:" | awk '{print $3}')

if [ -z "$BUY_TXN_ID" ]; then
    echo "Error: Failed to extract transaction ID from create operation"
    exit 1
fi

# Example 2: Create a SELL transaction
echo "2. Creating a SELL transaction..."
SELL_OUTPUT=$($CLIENT -addr "$SERVER_ADDR" \
    -op create \
    -user-id "$USER_ID" \
    -type SELL \
    -symbol AMZN \
    -quantity 5 \
    -price 155.75 \
    -executed-at "2024-02-20T14:45:00Z" \
    -brokerage 5.00 2>&1)

echo "$SELL_OUTPUT"
echo ""

SELL_TXN_ID=$(echo "$SELL_OUTPUT" | grep "Transaction ID:" | awk '{print $3}')

# Example 3: Create a DEP (deposit) transaction
echo "3. Creating a DEP (deposit) transaction..."
$CLIENT -addr "$SERVER_ADDR" \
    -op create \
    -user-id "$USER_ID" \
    -type DEP \
    -amount 5000.00 \
    -executed-at "2024-01-01T09:00:00Z" \
    -notes "Initial deposit"
echo ""

# Example 4: Create a DIV (dividend) transaction
echo "4. Creating a DIV (dividend) transaction..."
$CLIENT -addr "$SERVER_ADDR" \
    -op create \
    -user-id "$USER_ID" \
    -type DIV \
    -symbol AMZN \
    -amount 25.50 \
    -executed-at "2024-03-15T12:00:00Z" \
    -notes "Quarterly dividend"
echo ""

# Example 5: List all transactions
echo "5. Listing all transactions for user..."
$CLIENT -addr "$SERVER_ADDR" \
    -op list \
    -user-id "$USER_ID"
echo ""

# Example 6: List transactions filtered by symbol
echo "6. Listing transactions filtered by symbol (AMZN)..."
$CLIENT -addr "$SERVER_ADDR" \
    -op list \
    -user-id "$USER_ID" \
    -filter-symbol AMZN
echo ""

# Example 7: Get specific transaction
echo "7. Retrieving specific transaction..."
$CLIENT -addr "$SERVER_ADDR" \
    -op get \
    -user-id "$USER_ID" \
    -transaction-id "$BUY_TXN_ID"
echo ""

# Example 8: Update transaction notes
echo "8. Updating transaction notes..."
$CLIENT -addr "$SERVER_ADDR" \
    -op update \
    -user-id "$USER_ID" \
    -transaction-id "$BUY_TXN_ID" \
    -notes "Updated: Initial AMZN purchase" \
    -update-fields notes
echo ""

# Example 9: Get oldest transaction
echo "9. Getting oldest transaction for user..."
$CLIENT -addr "$SERVER_ADDR" \
    -op oldest \
    -user-id "$USER_ID"
echo ""

# Example 10: Delete a transaction
echo "10. Deleting SELL transaction..."
$CLIENT -addr "$SERVER_ADDR" \
    -op delete \
    -user-id "$USER_ID" \
    -transaction-id "$SELL_TXN_ID"
echo ""

# Example 11: Verify deletion by listing again
echo "11. Verifying deletion by listing transactions..."
$CLIENT -addr "$SERVER_ADDR" \
    -op list \
    -user-id "$USER_ID"
echo ""

echo "=== Examples completed ==="
