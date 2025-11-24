#!/bin/bash

# CSV Upload Test Script
# This script tests the CSV upload functionality

set -e

echo "🧪 Testing CSV Upload Functionality"
echo "===================================="
echo ""

# Configuration
BASE_URL="http://localhost:8081"
USER_ID="test-user-123"
CSV_FILE="sample_transactions.csv"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if CSV file exists
if [ ! -f "$CSV_FILE" ]; then
    echo -e "${RED}❌ Error: $CSV_FILE not found${NC}"
    echo "Please run this script from the transaction-service directory"
    exit 1
fi

echo "📋 Configuration:"
echo "  - Base URL: $BASE_URL"
echo "  - User ID: $USER_ID"
echo "  - CSV File: $CSV_FILE"
echo ""

# Test 1: Health Check
echo "🏥 Test 1: Health Check"
echo "----------------------"
HEALTH_RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/health")
HTTP_CODE=$(echo "$HEALTH_RESPONSE" | tail -n1)
BODY=$(echo "$HEALTH_RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✅ Service is healthy${NC}"
else
    echo -e "${RED}❌ Service health check failed (HTTP $HTTP_CODE)${NC}"
    exit 1
fi
echo ""

# Test 2: Upload CSV with user_id as query parameter
echo "📤 Test 2: Upload CSV (user_id as query param)"
echo "----------------------------------------------"
UPLOAD_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
    -F "file=@$CSV_FILE" \
    "$BASE_URL/upload-csv?user_id=$USER_ID")

HTTP_CODE=$(echo "$UPLOAD_RESPONSE" | tail -n1)
BODY=$(echo "$UPLOAD_RESPONSE" | head -n-1)

echo "Response (HTTP $HTTP_CODE):"
echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"

if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "206" ]; then
    echo -e "${GREEN}✅ Upload successful${NC}"
    
    # Parse response
    TOTAL=$(echo "$BODY" | jq -r '.total_records' 2>/dev/null || echo "N/A")
    SUCCESS=$(echo "$BODY" | jq -r '.successful_records' 2>/dev/null || echo "N/A")
    FAILED=$(echo "$BODY" | jq -r '.failed_records' 2>/dev/null || echo "N/A")
    
    echo ""
    echo "📊 Summary:"
    echo "  - Total Records: $TOTAL"
    echo "  - Successful: $SUCCESS"
    echo "  - Failed: $FAILED"
    
    if [ "$FAILED" != "0" ] && [ "$FAILED" != "N/A" ]; then
        echo -e "${YELLOW}⚠️  Some records failed to import${NC}"
        echo "$BODY" | jq '.errors' 2>/dev/null || true
    fi
else
    echo -e "${RED}❌ Upload failed (HTTP $HTTP_CODE)${NC}"
    echo "Response: $BODY"
    exit 1
fi
echo ""

# Test 3: Upload CSV with user_id as header
echo "📤 Test 3: Upload CSV (user_id as header)"
echo "-----------------------------------------"
UPLOAD_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
    -H "X-User-ID: $USER_ID" \
    -F "file=@$CSV_FILE" \
    "$BASE_URL/upload-csv")

HTTP_CODE=$(echo "$UPLOAD_RESPONSE" | tail -n1)
BODY=$(echo "$UPLOAD_RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "206" ]; then
    echo -e "${GREEN}✅ Upload with header successful${NC}"
else
    echo -e "${RED}❌ Upload with header failed (HTTP $HTTP_CODE)${NC}"
fi
echo ""

# Test 4: Upload without user_id (should fail)
echo "❌ Test 4: Upload without user_id (should fail)"
echo "-----------------------------------------------"
UPLOAD_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
    -F "file=@$CSV_FILE" \
    "$BASE_URL/upload-csv")

HTTP_CODE=$(echo "$UPLOAD_RESPONSE" | tail -n1)

if [ "$HTTP_CODE" = "400" ]; then
    echo -e "${GREEN}✅ Correctly rejected (HTTP 400)${NC}"
else
    echo -e "${RED}❌ Expected HTTP 400, got $HTTP_CODE${NC}"
fi
echo ""

# Test 5: Upload non-CSV file (should fail)
echo "❌ Test 5: Upload non-CSV file (should fail)"
echo "--------------------------------------------"
echo "test" > /tmp/test.txt
UPLOAD_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
    -F "file=@/tmp/test.txt" \
    "$BASE_URL/upload-csv?user_id=$USER_ID")

HTTP_CODE=$(echo "$UPLOAD_RESPONSE" | tail -n1)

if [ "$HTTP_CODE" = "400" ]; then
    echo -e "${GREEN}✅ Correctly rejected non-CSV file (HTTP 400)${NC}"
else
    echo -e "${YELLOW}⚠️  Expected HTTP 400, got $HTTP_CODE${NC}"
fi
rm /tmp/test.txt
echo ""

echo "===================================="
echo -e "${GREEN}✅ All tests completed!${NC}"
echo "===================================="
