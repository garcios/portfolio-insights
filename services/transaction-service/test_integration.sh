#!/bin/bash

# Simple Integration Test - Focus on Cash Transactions
# Tests INT, DEP, WIT (no asset validation required)

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

DB_URL="postgres://garcios:Password123@localhost:5432/portfolio"
GRPC_HOST="localhost:50053"

echo -e "${YELLOW}=== Transaction Service Integration Tests (Cash Transactions) ===${NC}"
echo ""

# Get test user
USER_ID=$(psql "$DB_URL" -t -c "SELECT id FROM customers.users LIMIT 1;" | xargs)

if [ -z "$USER_ID" ]; then
    echo -e "${RED}No users found${NC}"
    exit 1
fi

echo -e "${GREEN}Test User: $USER_ID${NC}"
echo ""

PASSED=0
FAILED=0

# Test cash transactions (no asset validation)
test_cash_transaction() {
    local name=$1
    local type=$2
    local amount=$3
    
    echo -e "${YELLOW}Test: $name${NC}"
    
    json='{"user_id":"'$USER_ID'","type":"'$type'","amount":'$amount',"executed_at":"'$(date -u +%Y-%m-%dT%H:%M:%SZ)'","price_currency":"USD","brokerage_currency":"USD","brokerage":0,"notes":"Integration test"}'
    
    if response=$(grpcurl -plaintext -import-path ../../proto -proto transaction/transaction.proto \
        -d "$json" $GRPC_HOST transaction.TransactionService/CreateTransaction 2>&1); then
        
        TXN_ID=$(echo "$response" | jq -r '.transaction.id' 2>/dev/null || echo "")
        
        if [ -n "$TXN_ID" ] && [ "$TXN_ID" != "null" ]; then
            # Verify in database
            DB_TYPE=$(psql "$DB_URL" -t -c "SELECT type FROM txn.transactions WHERE id='$TXN_ID';" | xargs)
            DB_AMOUNT=$(psql "$DB_URL" -t -c "SELECT amount FROM txn.transactions WHERE id='$TXN_ID';" | xargs)
            
            if [ "$DB_TYPE" = "$type" ] && [ -n "$DB_AMOUNT" ]; then
                echo -e "${GREEN}✓ PASSED${NC}"
                echo "  ID: $TXN_ID | Type: $DB_TYPE | Amount: $DB_AMOUNT"
                ((PASSED++))
            else
                echo -e "${RED}✗ FAILED - DB verification failed${NC}"
                ((FAILED++))
            fi
        else
            echo -e "${RED}✗ FAILED - No transaction ID${NC}"
            echo "$response"
            ((FAILED++))
        fi
    else
        echo -e "${RED}✗ FAILED - gRPC error${NC}"
        echo "$response"
        ((FAILED++))
    fi
    echo ""
}

# Run tests
test_cash_transaction "INT - Interest Income" "INT" "25.50"
test_cash_transaction "DEP - Cash Deposit" "DEP" "1000.00"
test_cash_transaction "WIT - Cash Withdrawal" "WIT" "500.00"

# Verify schema changes
echo -e "${YELLOW}Verifying database schema...${NC}"
SCHEMA_CHECK=$(psql "$DB_URL" -t -c "
    SELECT COUNT(*) FROM information_schema.columns 
    WHERE table_schema='txn' AND table_name='transactions' 
    AND column_name='amount' AND is_nullable='YES';
" | xargs)

if [ "$SCHEMA_CHECK" = "1" ]; then
    echo -e "${GREEN}✓ Amount column exists and is nullable${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Schema check failed${NC}"
    ((FAILED++))
fi
echo ""

# Check nullable fields
NULLABLE_FIELDS=$(psql "$DB_URL" -t -c "
    SELECT COUNT(*) FROM information_schema.columns 
    WHERE table_schema='txn' AND table_name='transactions' 
    AND column_name IN ('symbol', 'quantity', 'price_per_share') 
    AND is_nullable='YES';
" | xargs)

if [ "$NULLABLE_FIELDS" = "3" ]; then
    echo -e "${GREEN}✓ Equity fields are nullable${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Nullable fields check failed${NC}"
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
    
    # Show sample transactions
    echo ""
    echo -e "${YELLOW}Sample transactions created:${NC}"
    psql "$DB_URL" -c "SELECT id, type, amount, symbol, quantity, price_per_share, executed_at FROM txn.transactions WHERE user_id='$USER_ID' ORDER BY created_at DESC LIMIT 5;"
    
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
