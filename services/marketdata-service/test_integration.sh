#!/bin/bash

# Integration Test for MarketData Service
# Tests all gRPC endpoints: GetAsset, ListAssets, GetLatestPrice, GetLatestPrices,
# GetHistoricalPrices, GetLatestCurrencyRate, GetHistoricalCurrencyRates

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

DB_URL="postgres://garcios:Password123@localhost:5432/portfolio"
GRPC_HOST="localhost:50054"

echo -e "${YELLOW}=== MarketData Service Integration Tests ===${NC}"
echo ""

PASSED=0
FAILED=0

# Generate unique test identifier
TEST_ID=$(date +%s)
TEST_SYMBOL="TEST${TEST_ID}"
TEST_SYMBOL2="TST2${TEST_ID}"

# Cleanup function
cleanup() {
    echo -e "${YELLOW}Cleaning up test data...${NC}"
    
    # Delete test prices first (foreign key constraint)
    psql "$DB_URL" -c "DELETE FROM marketdata.asset_prices WHERE asset_id IN (SELECT id FROM marketdata.assets WHERE symbol LIKE 'TEST%' OR symbol LIKE 'TST%');" > /dev/null 2>&1 || true
    
    # Delete test assets
    psql "$DB_URL" -c "DELETE FROM marketdata.assets WHERE symbol LIKE 'TEST%' OR symbol LIKE 'TST%';" > /dev/null 2>&1 || true
    
    # Delete test currency rates
    psql "$DB_URL" -c "DELETE FROM marketdata.currency_rates WHERE base_currency='XXX' OR target_currency='XXX';" > /dev/null 2>&1 || true
    
    echo -e "${GREEN}✓ Cleanup complete${NC}"
}

# Trap to ensure cleanup runs on exit
trap cleanup EXIT

# Cleanup any existing test data first
cleanup
echo ""

echo -e "${YELLOW}=== Test 1: GetAsset RPC ===${NC}"

# Insert test asset
ASSET_ID=$(psql "$DB_URL" -A -t -c "INSERT INTO marketdata.assets (symbol, name, type, exchange, currency) VALUES ('$TEST_SYMBOL', 'Test Asset $TEST_ID', 'EQUITY', 'NASDAQ', 'USD') RETURNING id;" | head -1)

if [ -z "$ASSET_ID" ]; then
    echo -e "${RED}✗ FAILED - Could not insert test asset${NC}"
    ((FAILED++))
else
    echo "  Created test asset: $ASSET_ID"
    
    # Test GetAsset
    json='{"symbol":"'$TEST_SYMBOL'"}'
    
    if response=$(grpcurl -plaintext -import-path ../../proto -proto marketdata/marketdata.proto \
        -d "$json" $GRPC_HOST marketdata.MarketDataService/GetAsset 2>&1); then
        
        RESP_SYMBOL=$(echo "$response" | jq -r '.asset.symbol' 2>/dev/null || echo "")
        RESP_NAME=$(echo "$response" | jq -r '.asset.name' 2>/dev/null || echo "")
        RESP_TYPE=$(echo "$response" | jq -r '.asset.type' 2>/dev/null || echo "")
        
        if [ "$RESP_SYMBOL" = "$TEST_SYMBOL" ] && [ "$RESP_TYPE" = "EQUITY" ]; then
            echo -e "${GREEN}✓ PASSED${NC}"
            echo "  Symbol: $RESP_SYMBOL"
            echo "  Name: $RESP_NAME"
            echo "  Type: $RESP_TYPE"
            ((PASSED++))
        else
            echo -e "${RED}✗ FAILED - Response data mismatch${NC}"
            echo "  Response: $response"
            ((FAILED++))
        fi
    else
        echo -e "${RED}✗ FAILED - gRPC error${NC}"
        echo "$response"
        ((FAILED++))
    fi
fi
echo ""

echo -e "${YELLOW}=== Test 2: ListAssets RPC ===${NC}"

# Insert another test asset
ASSET_ID2=$(psql "$DB_URL" -A -t -c "INSERT INTO marketdata.assets (symbol, name, type, exchange, currency) VALUES ('$TEST_SYMBOL2', 'Test Asset 2 $TEST_ID', 'CRYPTO', 'BINANCE', 'USD') RETURNING id;" | head -1)

json='{"page_size":10}'

if response=$(grpcurl -plaintext -import-path ../../proto -proto marketdata/marketdata.proto \
    -d "$json" $GRPC_HOST marketdata.MarketDataService/ListAssets 2>&1); then
    
    ASSET_COUNT=$(echo "$response" | jq -r '.assets | length' 2>/dev/null || echo "0")
    
    if [ "$ASSET_COUNT" -gt "0" ]; then
        echo -e "${GREEN}✓ PASSED${NC}"
        echo "  Returned $ASSET_COUNT assets"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAILED - No assets returned${NC}"
        echo "  Response: $response"
        ((FAILED++))
    fi
else
    echo -e "${RED}✗ FAILED - gRPC error${NC}"
    echo "$response"
    ((FAILED++))
fi
echo ""

echo -e "${YELLOW}=== Test 3: GetLatestPrice RPC ===${NC}"

# Insert test price
PRICE_TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)
psql "$DB_URL" -c "INSERT INTO marketdata.asset_prices (asset_id, price, timestamp) VALUES ('$ASSET_ID', 150.50, '$PRICE_TIMESTAMP');" > /dev/null

json='{"symbol":"'$TEST_SYMBOL'"}'

if response=$(grpcurl -plaintext -import-path ../../proto -proto marketdata/marketdata.proto \
    -d "$json" $GRPC_HOST marketdata.MarketDataService/GetLatestPrice 2>&1); then
    
    RESP_PRICE=$(echo "$response" | jq -r '.price.price' 2>/dev/null || echo "")
    
    if [ -n "$RESP_PRICE" ] && [ "$RESP_PRICE" != "null" ]; then
        echo -e "${GREEN}✓ PASSED${NC}"
        echo "  Price: $RESP_PRICE"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAILED - No price returned${NC}"
        echo "  Response: $response"
        ((FAILED++))
    fi
else
    echo -e "${RED}✗ FAILED - gRPC error${NC}"
    echo "$response"
    ((FAILED++))
fi
echo ""

echo -e "${YELLOW}=== Test 4: GetLatestPrices RPC ===${NC}"

# Insert price for second asset
psql "$DB_URL" -c "INSERT INTO marketdata.asset_prices (asset_id, price, timestamp) VALUES ('$ASSET_ID2', 0.00025, '$PRICE_TIMESTAMP');" > /dev/null

json='{"symbols":["'$TEST_SYMBOL'","'$TEST_SYMBOL2'"]}'

if response=$(grpcurl -plaintext -import-path ../../proto -proto marketdata/marketdata.proto \
    -d "$json" $GRPC_HOST marketdata.MarketDataService/GetLatestPrices 2>&1); then
    
    PRICE_COUNT=$(echo "$response" | jq -r '.prices | length' 2>/dev/null || echo "0")
    
    if [ "$PRICE_COUNT" -ge "1" ]; then
        echo -e "${GREEN}✓ PASSED${NC}"
        echo "  Returned prices for $PRICE_COUNT symbols"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAILED - No prices returned${NC}"
        echo "  Response: $response"
        ((FAILED++))
    fi
else
    echo -e "${RED}✗ FAILED - gRPC error${NC}"
    echo "$response"
    ((FAILED++))
fi
echo ""

echo -e "${YELLOW}=== Test 5: GetHistoricalPrices RPC ===${NC}"

# Insert historical prices
YESTERDAY=$(date -u -v-1d +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -d '1 day ago' +%Y-%m-%dT%H:%M:%SZ)
TWO_DAYS_AGO=$(date -u -v-2d +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -d '2 days ago' +%Y-%m-%dT%H:%M:%SZ)

psql "$DB_URL" -c "INSERT INTO marketdata.asset_prices (asset_id, price, timestamp) VALUES ('$ASSET_ID', 148.00, '$YESTERDAY');" > /dev/null
psql "$DB_URL" -c "INSERT INTO marketdata.asset_prices (asset_id, price, timestamp) VALUES ('$ASSET_ID', 145.50, '$TWO_DAYS_AGO');" > /dev/null

START_TIME=$(date -u -v-3d +%Y-%m-%dT00:00:00Z 2>/dev/null || date -u -d '3 days ago' +%Y-%m-%dT00:00:00Z)
END_TIME=$(date -u +%Y-%m-%dT23:59:59Z)

json='{"symbol":"'$TEST_SYMBOL'","start_time":"'$START_TIME'","end_time":"'$END_TIME'","interval":"1d"}'

if response=$(grpcurl -plaintext -import-path ../../proto -proto marketdata/marketdata.proto \
    -d "$json" $GRPC_HOST marketdata.MarketDataService/GetHistoricalPrices 2>&1); then
    
    PRICE_COUNT=$(echo "$response" | jq -r '.prices | length' 2>/dev/null || echo "0")
    
    if [ "$PRICE_COUNT" -ge "1" ]; then
        echo -e "${GREEN}✓ PASSED${NC}"
        echo "  Returned $PRICE_COUNT historical prices"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAILED - No historical prices returned${NC}"
        echo "  Response: $response"
        ((FAILED++))
    fi
else
    echo -e "${RED}✗ FAILED - gRPC error${NC}"
    echo "$response"
    ((FAILED++))
fi
echo ""

echo -e "${YELLOW}=== Test 6: GetLatestCurrencyRate RPC ===${NC}"

# Insert test currency rate
RATE_DATE=$(date -u +%Y-%m-%d)
psql "$DB_URL" -c "INSERT INTO marketdata.currency_rates (base_currency, target_currency, rate, rate_date) VALUES ('XXX', 'USD', 1.25, '$RATE_DATE');" > /dev/null

json='{"base_currency":"XXX","target_currency":"USD"}'

if response=$(grpcurl -plaintext -import-path ../../proto -proto marketdata/marketdata.proto \
    -d "$json" $GRPC_HOST marketdata.MarketDataService/GetLatestCurrencyRate 2>&1); then
    
    RESP_RATE=$(echo "$response" | jq -r '.currencyRate.rate' 2>/dev/null || echo "")
    RESP_BASE=$(echo "$response" | jq -r '.currencyRate.baseCurrency' 2>/dev/null || echo "")
    
    if [ "$RESP_BASE" = "XXX" ] && [ -n "$RESP_RATE" ] && [ "$RESP_RATE" != "null" ]; then
        echo -e "${GREEN}✓ PASSED${NC}"
        echo "  Base: $RESP_BASE"
        echo "  Rate: $RESP_RATE"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAILED - Invalid currency rate response${NC}"
        echo "  Expected base: XXX, got: $RESP_BASE"
        echo "  Expected rate: not null, got: $RESP_RATE"
        echo "  Response: $response"
        ((FAILED++))
    fi
else
    echo -e "${RED}✗ FAILED - gRPC error${NC}"
    echo "$response"
    ((FAILED++))
fi
echo ""

echo -e "${YELLOW}=== Test 7: GetHistoricalCurrencyRates RPC ===${NC}"

# Insert historical currency rates
YESTERDAY_DATE=$(date -u -v-1d +%Y-%m-%d 2>/dev/null || date -u -d '1 day ago' +%Y-%m-%d)
psql "$DB_URL" -c "INSERT INTO marketdata.currency_rates (base_currency, target_currency, rate, rate_date) VALUES ('XXX', 'USD', 1.23, '$YESTERDAY_DATE');" > /dev/null

START_DATE=$(date -u -v-3d +%Y-%m-%dT00:00:00Z 2>/dev/null || date -u -d '3 days ago' +%Y-%m-%dT00:00:00Z)
END_DATE=$(date -u +%Y-%m-%dT23:59:59Z)

json='{"base_currency":"XXX","target_currency":"USD","start_time":"'$START_DATE'","end_time":"'$END_DATE'"}'

if response=$(grpcurl -plaintext -import-path ../../proto -proto marketdata/marketdata.proto \
    -d "$json" $GRPC_HOST marketdata.MarketDataService/GetHistoricalCurrencyRates 2>&1); then
    
    RATE_COUNT=$(echo "$response" | jq -r '.rates | length' 2>/dev/null || echo "0")
    
    if [ "$RATE_COUNT" -ge "1" ]; then
        echo -e "${GREEN}✓ PASSED${NC}"
        echo "  Returned $RATE_COUNT historical rates"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAILED - No historical rates returned${NC}"
        echo "  Response: $response"
        ((FAILED++))
    fi
else
    echo -e "${RED}✗ FAILED - gRPC error${NC}"
    echo "$response"
    ((FAILED++))
fi
echo ""

echo -e "${YELLOW}=== Test 8: Database Schema Validation ===${NC}"

# Check if tables exist
ASSETS_TABLE=$(psql "$DB_URL" -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'marketdata' AND table_name = 'assets');" | xargs)
PRICES_TABLE=$(psql "$DB_URL" -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'marketdata' AND table_name = 'asset_prices');" | xargs)
CURRENCY_TABLE=$(psql "$DB_URL" -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'marketdata' AND table_name = 'currency_rates');" | xargs)

SCHEMA_OK=true

if [ "$ASSETS_TABLE" = "t" ]; then
    echo -e "${GREEN}  ✓ marketdata.assets table exists${NC}"
else
    echo -e "${RED}  ✗ marketdata.assets table missing${NC}"
    SCHEMA_OK=false
fi

if [ "$PRICES_TABLE" = "t" ]; then
    echo -e "${GREEN}  ✓ marketdata.asset_prices table exists${NC}"
else
    echo -e "${RED}  ✗ marketdata.asset_prices table missing${NC}"
    SCHEMA_OK=false
fi

if [ "$CURRENCY_TABLE" = "t" ]; then
    echo -e "${GREEN}  ✓ marketdata.currency_rates table exists${NC}"
else
    echo -e "${RED}  ✗ marketdata.currency_rates table missing${NC}"
    SCHEMA_OK=false
fi

# Check key indexes
SYMBOL_INDEX=$(psql "$DB_URL" -t -c "SELECT EXISTS (SELECT FROM pg_indexes WHERE schemaname = 'marketdata' AND tablename = 'assets' AND indexname = 'idx_assets_symbol');" | xargs)

if [ "$SYMBOL_INDEX" = "t" ]; then
    echo -e "${GREEN}  ✓ idx_assets_symbol index exists${NC}"
else
    echo -e "${YELLOW}  ⚠ idx_assets_symbol index missing${NC}"
fi

if [ "$SCHEMA_OK" = true ]; then
    echo -e "${GREEN}✓ PASSED${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ FAILED${NC}"
    ((FAILED++))
fi
echo ""

# Summary
echo -e "${YELLOW}=== Test Summary ===${NC}"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All integration tests passed!${NC}"
    echo ""
    echo "Summary:"
    echo "  ✓ GetAsset RPC working"
    echo "  ✓ ListAssets RPC working"
    echo "  ✓ GetLatestPrice RPC working"
    echo "  ✓ GetLatestPrices RPC working"
    echo "  ✓ GetHistoricalPrices RPC working"
    echo "  ✓ GetLatestCurrencyRate RPC working"
    echo "  ✓ GetHistoricalCurrencyRates RPC working"
    echo "  ✓ Database schema validated"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
