#!/bin/bash

# Integration Test for User Service
# Tests all gRPC endpoints: CreateUser, GetUser, VerifyUser
# Updated for AIP-compliant resource names

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

DB_URL="postgres://garcios:Password123@localhost:5432/portfolio"
GRPC_HOST="localhost:50051"

echo -e "${YELLOW}=== User Service Integration Tests (AIP-Compliant) ===${NC}"
echo ""

PASSED=0
FAILED=0

# Generate unique test identifier
TEST_ID=$(date +%s)
TEST_EMAIL="test-${TEST_ID}@example.com"
TEST_USERNAME="testuser${TEST_ID}"
TEST_PASSWORD="TestPassword123!"

# Cleanup function
cleanup() {
    echo -e "${YELLOW}Cleaning up test data...${NC}"
    psql "$DB_URL" -c "DELETE FROM customers.users WHERE email='$TEST_EMAIL';" > /dev/null 2>&1 || true
    echo -e "${GREEN}✓ Cleanup complete${NC}"
}

# Trap to ensure cleanup runs on exit
trap cleanup EXIT

echo -e "${YELLOW}=== Test 1: CreateUser RPC (AIP-133 Compliant) ===${NC}"

# Create a new user with User object structure
json='{
  "user": {
    "email": "'$TEST_EMAIL'",
    "username": "'$TEST_USERNAME'",
    "password": "'$TEST_PASSWORD'"
  }
}'

if response=$(grpcurl -plaintext -import-path ../../proto -proto user/user.proto \
    -d "$json" $GRPC_HOST user.UserService/CreateUser 2>&1); then
    
    # Response now returns User object with name and userId fields
    USER_NAME=$(echo "$response" | jq -r '.name' 2>/dev/null || echo "")
    USER_ID=$(echo "$response" | jq -r '.userId' 2>/dev/null || echo "")
    
    if [ -n "$USER_ID" ] && [ "$USER_ID" != "null" ]; then
        echo -e "${GREEN}✓ User created successfully${NC}"
        echo "  Resource Name: $USER_NAME"
        echo "  User ID: $USER_ID"
        
        # Verify resource name format
        EXPECTED_NAME="users/$USER_ID"
        if [ "$USER_NAME" = "$EXPECTED_NAME" ]; then
            echo -e "${GREEN}✓ Resource name format correct${NC}"
        else
            echo -e "${RED}✗ Resource name format incorrect${NC}"
            echo "  Expected: $EXPECTED_NAME"
            echo "  Got: $USER_NAME"
        fi
        
        # Verify in database
        DB_EMAIL=$(psql "$DB_URL" -t -c "SELECT email FROM customers.users WHERE id='$USER_ID';" | xargs)
        DB_USERNAME=$(psql "$DB_URL" -t -c "SELECT username FROM customers.users WHERE id='$USER_ID';" | xargs)
        DB_PASSWORD_HASH=$(psql "$DB_URL" -t -c "SELECT password_hash FROM customers.users WHERE id='$USER_ID';" | xargs)
        
        if [ "$DB_EMAIL" = "$TEST_EMAIL" ] && [ "$DB_USERNAME" = "$TEST_USERNAME" ]; then
            echo -e "${GREEN}✓ Database verification passed${NC}"
            echo "  Email: $DB_EMAIL"
            echo "  Username: $DB_USERNAME"
            
            # Verify password is hashed (not plaintext)
            if [ "$DB_PASSWORD_HASH" != "$TEST_PASSWORD" ] && [ -n "$DB_PASSWORD_HASH" ]; then
                echo -e "${GREEN}✓ Password is hashed (not plaintext)${NC}"
                ((PASSED++))
            else
                echo -e "${RED}✗ FAILED - Password not properly hashed${NC}"
                ((FAILED++))
            fi
        else
            echo -e "${RED}✗ FAILED - Database verification failed${NC}"
            echo "  Expected email: $TEST_EMAIL, got: $DB_EMAIL"
            echo "  Expected username: $TEST_USERNAME, got: $DB_USERNAME"
            ((FAILED++))
        fi
    else
        echo -e "${RED}✗ FAILED - No user ID returned${NC}"
        echo "$response"
        ((FAILED++))
    fi
else
    echo -e "${RED}✗ FAILED - gRPC error${NC}"
    echo "$response"
    ((FAILED++))
fi
echo ""

# Only proceed with remaining tests if user was created
if [ -z "$USER_ID" ] || [ "$USER_ID" = "null" ]; then
    echo -e "${RED}Cannot proceed with remaining tests - user creation failed${NC}"
    exit 1
fi

echo -e "${YELLOW}=== Test 2: GetUser RPC (AIP-131 Compliant) ===${NC}"

# Use resource name instead of ID
json='{"name":"users/'$USER_ID'"}'

if response=$(grpcurl -plaintext -import-path ../../proto -proto user/user.proto \
    -d "$json" $GRPC_HOST user.UserService/GetUser 2>&1); then
    
    # Response is now User object
    RESP_NAME=$(echo "$response" | jq -r '.name' 2>/dev/null || echo "")
    RESP_USER_ID=$(echo "$response" | jq -r '.userId' 2>/dev/null || echo "")
    RESP_EMAIL=$(echo "$response" | jq -r '.email' 2>/dev/null || echo "")
    RESP_USERNAME=$(echo "$response" | jq -r '.username' 2>/dev/null || echo "")
    
    if [ "$RESP_USER_ID" = "$USER_ID" ] && [ "$RESP_EMAIL" = "$TEST_EMAIL" ] && [ "$RESP_USERNAME" = "$TEST_USERNAME" ]; then
        echo -e "${GREEN}✓ PASSED${NC}"
        echo "  Resource Name: $RESP_NAME"
        echo "  User ID: $RESP_USER_ID"
        echo "  Email: $RESP_EMAIL"
        echo "  Username: $RESP_USERNAME"
        
        # Verify resource name format
        if [ "$RESP_NAME" = "users/$USER_ID" ]; then
            echo -e "${GREEN}✓ Resource name format correct${NC}"
        else
            echo -e "${YELLOW}⚠ Resource name format unexpected: $RESP_NAME${NC}"
        fi
        
        ((PASSED++))
    else
        echo -e "${RED}✗ FAILED - Response data mismatch${NC}"
        echo "  Expected user ID: $USER_ID, got: $RESP_USER_ID"
        echo "  Expected email: $TEST_EMAIL, got: $RESP_EMAIL"
        echo "  Expected username: $TEST_USERNAME, got: $RESP_USERNAME"
        ((FAILED++))
    fi
else
    echo -e "${RED}✗ FAILED - gRPC error${NC}"
    echo "$response"
    ((FAILED++))
fi
echo ""

echo -e "${YELLOW}=== Test 3: GetUser with Invalid Resource Name ===${NC}"

# Test with invalid resource name format
json='{"name":"invalid-name"}'

if response=$(grpcurl -plaintext -import-path ../../proto -proto user/user.proto \
    -d "$json" $GRPC_HOST user.UserService/GetUser 2>&1); then
    
    # Should return error for invalid resource name
    if echo "$response" | grep -q "invalid resource name"; then
        echo -e "${GREEN}✓ PASSED - Correctly rejected invalid resource name${NC}"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAILED - Should have rejected invalid resource name${NC}"
        echo "$response"
        ((FAILED++))
    fi
else
    # grpcurl returns non-zero for gRPC errors, which is expected
    if echo "$response" | grep -q "invalid resource name\|InvalidArgument"; then
        echo -e "${GREEN}✓ PASSED - Correctly rejected invalid resource name${NC}"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAILED - Unexpected error${NC}"
        echo "$response"
        ((FAILED++))
    fi
fi
echo ""

echo -e "${YELLOW}=== Test 4: VerifyUser RPC (Valid Credentials) ===${NC}"

json='{"email":"'$TEST_EMAIL'","password":"'$TEST_PASSWORD'"}'

if response=$(grpcurl -plaintext -import-path ../../proto -proto user/user.proto \
    -d "$json" $GRPC_HOST user.UserService/VerifyUser 2>&1); then
    
    VALID=$(echo "$response" | jq -r '.valid' 2>/dev/null || echo "")
    
    # User object is now nested in response
    RESP_USER_ID=$(echo "$response" | jq -r '.user.userId' 2>/dev/null || echo "")
    RESP_NAME=$(echo "$response" | jq -r '.user.name' 2>/dev/null || echo "")
    RESP_EMAIL=$(echo "$response" | jq -r '.user.email' 2>/dev/null || echo "")
    RESP_USERNAME=$(echo "$response" | jq -r '.user.username' 2>/dev/null || echo "")
    
    if [ "$VALID" = "true" ] && [ "$RESP_USER_ID" = "$USER_ID" ] && [ "$RESP_EMAIL" = "$TEST_EMAIL" ]; then
        echo -e "${GREEN}✓ PASSED${NC}"
        echo "  Valid: $VALID"
        echo "  Resource Name: $RESP_NAME"
        echo "  User ID: $RESP_USER_ID"
        echo "  Email: $RESP_EMAIL"
        echo "  Username: $RESP_USERNAME"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAILED - Invalid response${NC}"
        echo "  Expected valid=true, got: $VALID"
        echo "$response"
        ((FAILED++))
    fi
else
    echo -e "${RED}✗ FAILED - gRPC error${NC}"
    echo "$response"
    ((FAILED++))
fi
echo ""

echo -e "${YELLOW}=== Test 5: VerifyUser RPC (Invalid Password) ===${NC}"

json='{"email":"'$TEST_EMAIL'","password":"WrongPassword123!"}'

if response=$(grpcurl -plaintext -import-path ../../proto -proto user/user.proto \
    -d "$json" $GRPC_HOST user.UserService/VerifyUser 2>&1); then
    
    # grpcurl omits default values, so empty {} means valid=false (default)
    # We check if valid field is either missing or explicitly false
    VALID=$(echo "$response" | jq -r '.valid // false' 2>/dev/null || echo "false")
    
    if [ "$VALID" = "false" ]; then
        echo -e "${GREEN}✓ PASSED${NC}"
        echo "  Valid: $VALID (correctly rejected invalid password)"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAILED - Should have rejected invalid password${NC}"
        echo "  Response: $response"
        ((FAILED++))
    fi
else
    echo -e "${RED}✗ FAILED - gRPC error${NC}"
    echo "$response"
    ((FAILED++))
fi
echo ""

echo -e "${YELLOW}=== Test 6: VerifyUser RPC (Non-existent User) ===${NC}"

json='{"email":"nonexistent-'$TEST_ID'@example.com","password":"SomePassword123!"}'

if response=$(grpcurl -plaintext -import-path ../../proto -proto user/user.proto \
    -d "$json" $GRPC_HOST user.UserService/VerifyUser 2>&1); then
    
    # grpcurl omits default values, so empty {} means valid=false (default)
    VALID=$(echo "$response" | jq -r '.valid // false' 2>/dev/null || echo "false")
    
    if [ "$VALID" = "false" ]; then
        echo -e "${GREEN}✓ PASSED${NC}"
        echo "  Valid: $VALID (correctly rejected non-existent user)"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAILED - Should have rejected non-existent user${NC}"
        echo "  Response: $response"
        ((FAILED++))
    fi
else
    echo -e "${RED}✗ FAILED - gRPC error${NC}"
    echo "$response"
    ((FAILED++))
fi
echo ""

echo -e "${YELLOW}=== Test 7: Database Schema Validation ===${NC}"

# Check if table exists
TABLE_EXISTS=$(psql "$DB_URL" -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'customers' AND table_name = 'users');" | xargs)

if [ "$TABLE_EXISTS" = "t" ]; then
    echo -e "${GREEN}✓ customers.users table exists${NC}"
    
    # Check required columns
    REQUIRED_COLUMNS=("id" "email" "username" "password_hash" "created_at" "updated_at")
    COLUMNS_OK=true
    
    for col in "${REQUIRED_COLUMNS[@]}"; do
        COL_EXISTS=$(psql "$DB_URL" -t -c "SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = 'customers' AND table_name = 'users' AND column_name = '$col');" | xargs)
        if [ "$COL_EXISTS" = "t" ]; then
            echo -e "${GREEN}  ✓ Column '$col' exists${NC}"
        else
            echo -e "${RED}  ✗ Column '$col' missing${NC}"
            COLUMNS_OK=false
        fi
    done
    
    # Check email unique constraint
    UNIQUE_CONSTRAINT=$(psql "$DB_URL" -t -c "SELECT COUNT(*) FROM information_schema.table_constraints WHERE table_schema = 'customers' AND table_name = 'users' AND constraint_type = 'UNIQUE' AND constraint_name LIKE '%email%';" | xargs)
    
    if [ "$UNIQUE_CONSTRAINT" -gt "0" ]; then
        echo -e "${GREEN}  ✓ Email has UNIQUE constraint${NC}"
    else
        echo -e "${YELLOW}  ⚠ Email UNIQUE constraint not found (may use different naming)${NC}"
    fi
    
    if [ "$COLUMNS_OK" = true ]; then
        ((PASSED++))
    else
        ((FAILED++))
    fi
else
    echo -e "${RED}✗ FAILED - customers.users table does not exist${NC}"
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
    echo "  ✓ CreateUser RPC working (AIP-133 compliant)"
    echo "  ✓ GetUser RPC working (AIP-131 compliant)"
    echo "  ✓ GetUser validates resource name format"
    echo "  ✓ VerifyUser RPC working (valid credentials)"
    echo "  ✓ VerifyUser RPC working (invalid password)"
    echo "  ✓ VerifyUser RPC working (non-existent user)"
    echo "  ✓ Database schema validated"
    echo "  ✓ Password hashing verified"
    echo "  ✓ Resource names follow users/{user} pattern"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
