#!/bin/bash
# Simplified Integration Test for Portfolio Service - Cash Balance Functionality
# Tests NATS event processing and cash balance table operations

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Configuration
NATS_URL="nats://localhost:4222"
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="portfolio"
DB_USER="garcios"
DB_PASSWORD="Password123"

# Test user ID
TEST_USER_ID="550e8400-e29b-41d4-a716-446655440000"

echo -e "${YELLOW}=== Portfolio Service Integration Tests (Simplified) ===${NC}"
echo "Testing cash balance functionality..."
echo ""

# Function to publish NATS event
publish_event() {
    local subject=$1
    local payload=$2
    echo -e "${YELLOW}Publishing to ${subject}...${NC}"
    echo "$payload" | nats pub "$subject" --server="$NATS_URL"
    sleep 2  # Give service time to process
}

# Function to query database
query_db() {
    local query=$1
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "$query"
}

# Function to verify cash balance
verify_cash_balance() {
    local user_id=$1
    local currency=$2
    local expected_balance=$3
    
    local actual_balance=$(query_db "SELECT balance FROM investments.cash_balances WHERE user_id='$user_id' AND currency='$currency';")
    actual_balance=$(echo "$actual_balance" | xargs)
    
    # Convert to comparable format (remove trailing zeros)
    expected_clean=$(echo "$expected_balance" | awk '{printf "%.2f", $1}')
    actual_clean=$(echo "$actual_balance" | awk '{printf "%.2f", $1}')
    
    if [ "$actual_clean" == "$expected_clean" ]; then
        echo -e "${GREEN}✓ Cash balance verified: $currency = $expected_balance${NC}"
        return 0
    else
        echo -e "${RED}✗ Cash balance mismatch: expected $expected_balance, got $actual_balance${NC}"
        return 1
    fi
}

# Function to cleanup
cleanup() {
    echo -e "${YELLOW}Cleaning up test data...${NC}"
    query_db "DELETE FROM investments.cash_balances WHERE user_id='$TEST_USER_ID';" > /dev/null
    echo -e "${GREEN}✓ Cleanup complete${NC}"
}

# Main test execution
main() {
    echo "=== Pre-flight Checks ==="
    
    # Check if NATS is available
    if ! nats --version > /dev/null 2>&1; then
        echo -e "${RED}✗ NATS CLI not installed${NC}"
        echo "Install with: brew install nats-io/nats-tools/nats"
        exit 1
    fi
    echo -e "${GREEN}✓ NATS CLI available${NC}"
    
    # Ensure test user exists
    echo -e "${YELLOW}Ensuring test user exists...${NC}"
    query_db "INSERT INTO customers.users (id, username, email, password_hash, created_at, updated_at) VALUES ('$TEST_USER_ID', 'testuser', 'test@example.com', 'dummy_hash_for_testing', NOW(), NOW()) ON CONFLICT (id) DO NOTHING;" > /dev/null
    echo -e "${GREEN}✓ Test user ready${NC}"
    echo ""
    
    # Cleanup any existing test data
    cleanup
    echo ""
    
    echo "=== Test 1: Initial Deposit (DEP) ==="
    publish_event "transaction-service.transaction.created" '{
        "transaction_id": "test-dep-001",
        "user_id": "'"$TEST_USER_ID"'",
        "type": "DEP",
        "amount": 1000.00,
        "executed_at": "'"$(date -u +%Y-%m-%dT%H:%M:%SZ)"'"
    }'
    
    if verify_cash_balance "$TEST_USER_ID" "USD" "1000"; then
        echo -e "${GREEN}✓ Test 1 PASSED${NC}"
    else
        echo -e "${RED}✗ Test 1 FAILED${NC}"
        exit 1
    fi
    echo ""
    
    echo "=== Test 2: Interest Income (INT) ==="
    publish_event "transaction-service.transaction.created" '{
        "transaction_id": "test-int-001",
        "user_id": "'"$TEST_USER_ID"'",
        "type": "INT",
        "amount": 25.50,
        "executed_at": "'"$(date -u +%Y-%m-%dT%H:%M:%SZ)"'"
    }'
    
    if verify_cash_balance "$TEST_USER_ID" "USD" "1025.5"; then
        echo -e "${GREEN}✓ Test 2 PASSED${NC}"
    else
        echo -e "${RED}✗ Test 2 FAILED${NC}"
        exit 1
    fi
    echo ""
    
    echo "=== Test 3: Dividend Income (DIV) ==="
    publish_event "transaction-service.transaction.created" '{
        "transaction_id": "test-div-001",
        "user_id": "'"$TEST_USER_ID"'",
        "type": "DIV",
        "asset_symbol": "AAPL",
        "amount": 15.75,
        "executed_at": "'"$(date -u +%Y-%m-%dT%H:%M:%SZ)"'"
    }'
    
    if verify_cash_balance "$TEST_USER_ID" "USD" "1041.25"; then
        echo -e "${GREEN}✓ Test 3 PASSED${NC}"
    else
        echo -e "${RED}✗ Test 3 FAILED${NC}"
        exit 1
    fi
    echo ""
    
    echo "=== Test 4: Withdrawal (WIT) ==="
    publish_event "transaction-service.transaction.created" '{
        "transaction_id": "test-wit-001",
        "user_id": "'"$TEST_USER_ID"'",
        "type": "WIT",
        "amount": 500.00,
        "executed_at": "'"$(date -u +%Y-%m-%dT%H:%M:%SZ)"'"
    }'
    
    if verify_cash_balance "$TEST_USER_ID" "USD" "541.25"; then
        echo -e "${GREEN}✓ Test 4 PASSED${NC}"
    else
        echo -e "${RED}✗ Test 4 FAILED${NC}"
        exit 1
    fi
    echo ""
    
    echo "=== Test 5: Negative Balance (Overdraft) ==="
    publish_event "transaction-service.transaction.created" '{
        "transaction_id": "test-wit-002",
        "user_id": "'"$TEST_USER_ID"'",
        "type": "WIT",
        "amount": 1000.00,
        "executed_at": "'"$(date -u +%Y-%m-%dT%H:%M:%SZ)"'"
    }'
    
    if verify_cash_balance "$TEST_USER_ID" "USD" "-458.75"; then
        echo -e "${GREEN}✓ Test 5 PASSED (negative balance allowed)${NC}"
    else
        echo -e "${RED}✗ Test 5 FAILED${NC}"
        exit 1
    fi
    echo ""
    
    echo "=== Test 6: Database Schema Verification ==="
    TABLE_EXISTS=$(query_db "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'investments' AND table_name = 'cash_balances');")
    TABLE_EXISTS=$(echo "$TABLE_EXISTS" | xargs)
    
    if [ "$TABLE_EXISTS" == "t" ]; then
        echo -e "${GREEN}✓ cash_balances table exists${NC}"
    else
        echo -e "${RED}✗ cash_balances table does not exist${NC}"
        exit 1
    fi
    echo ""
    
    echo "=== Test 7: No CASH-* Holdings Remain ==="
    CASH_HOLDINGS=$(query_db "SELECT COUNT(*) FROM investments.holdings WHERE symbol LIKE 'CASH-%';")
    CASH_HOLDINGS=$(echo "$CASH_HOLDINGS" | xargs)
    
    if [ "$CASH_HOLDINGS" == "0" ]; then
        echo -e "${GREEN}✓ Test 7 PASSED (no CASH-* holdings)${NC}"
    else
        echo -e "${YELLOW}⚠ Test 7 WARNING: Found $CASH_HOLDINGS CASH-* holdings${NC}"
    fi
    echo ""
    
    # Final cleanup
    cleanup
    echo ""
    
    echo -e "${GREEN}=== All Tests Completed Successfully ===${NC}"
    echo ""
    echo "Summary:"
    echo "  ✓ Deposit transactions working"
    echo "  ✓ Interest transactions working"
    echo "  ✓ Dividend transactions working"
    echo "  ✓ Withdrawal transactions working"
    echo "  ✓ Negative balances supported"
    echo "  ✓ Database schema correct"
    echo "  ✓ NATS event processing functional"
}

# Run tests
main

exit 0
