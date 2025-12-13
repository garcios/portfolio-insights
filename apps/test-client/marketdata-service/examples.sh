#!/bin/bash

# Example usage script for marketdata-service test client
# This script demonstrates how to use the test client to interact with the marketdata-service

set -e

# Configuration
SERVER_ADDR="${SERVER_ADDR:-localhost:50054}"
CLIENT="./marketdata-client"

echo "=== Market Data Service Test Client Examples ==="
echo "Server: $SERVER_ADDR"
echo ""

# Check if client binary exists
if [ ! -f "$CLIENT" ]; then
    echo "Building client..."
    go build -o marketdata-client
    echo ""
fi

# Example 1: Get asset by symbol
echo "1. Getting asset by symbol (AMZN)..."
ASSET_OUTPUT=$($CLIENT -addr "$SERVER_ADDR" \
    -op get-asset-by-symbol \
    -symbol "AMZN" 2>&1)

echo "$ASSET_OUTPUT"
echo ""

# Extract the resource name from the output
ASSET_NAME=$(echo "$ASSET_OUTPUT" | grep "Resource Name:" | awk '{print $3}')

if [ -z "$ASSET_NAME" ]; then
    echo "Error: Failed to extract asset resource name"
    exit 1
fi

# Example 2: Get asset by resource name
echo "2. Getting asset by resource name ($ASSET_NAME)..."
$CLIENT -addr "$SERVER_ADDR" \
    -op get-asset \
    -asset-name "$ASSET_NAME"
echo ""

# Example 3: Get latest price for the asset
echo "3. Getting latest price for $ASSET_NAME..."
$CLIENT -addr "$SERVER_ADDR" \
    -op get-price \
    -asset-name "$ASSET_NAME"
echo ""

# Example 4: List assets with pagination
echo "4. Listing assets (page size: 10)..."
$CLIENT -addr "$SERVER_ADDR" \
    -op list-assets \
    -page-size 10
echo ""

# Example 5: Get batch prices for multiple assets
echo "5. Getting batch prices for multiple assets..."
$CLIENT -addr "$SERVER_ADDR" \
    -op get-prices \
    -asset-names "assets/AMZN,assets/GOOGL,assets/MSFT" || true
echo ""

# Example 6: Get historical prices
echo "6. Getting historical prices (last 7 days)..."
END_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
START_TIME=$(date -u -v-7d +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "7 days ago" +"%Y-%m-%dT%H:%M:%SZ")

$CLIENT -addr "$SERVER_ADDR" \
    -op get-historical \
    -asset-name "$ASSET_NAME" \
    -start-time "$START_TIME" \
    -end-time "$END_TIME" \
    -interval "1d" || true
echo ""

# Example 7: Get currency exchange rate
echo "7. Getting currency exchange rate (USD -> EUR)..."
$CLIENT -addr "$SERVER_ADDR" \
    -op get-currency-rate \
    -base-currency "USD" \
    -target-currency "EUR" || true
echo ""

echo "=== Examples completed ==="
