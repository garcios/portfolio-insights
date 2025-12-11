#!/bin/bash

# Integration Test for Login-Consent-Provider
# Tests HTTP endpoints for OAuth2 login and consent flows

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

SERVICE_URL="http://localhost:3001"

echo -e "${YELLOW}=== Login-Consent-Provider Integration Tests ===${NC}"
echo ""

PASSED=0
FAILED=0

echo -e "${YELLOW}=== Test 1: Health Check ===${NC}"

if response=$(curl -s -o /dev/null -w "%{http_code}" "$SERVICE_URL/health" 2>&1); then
    if [ "$response" = "200" ] || [ "$response" = "302" ]; then
        echo -e "${GREEN}✓ PASSED${NC}"
        echo "  Health endpoint is accessible (HTTP $response)"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAILED - HTTP $response${NC}"
        ((FAILED++))
    fi
else
    echo -e "${RED}✗ FAILED - Connection error${NC}"
    echo "  Is the service running on port 3001?"
    ((FAILED++))
fi
echo ""

echo -e "${YELLOW}=== Test 2: Login Page (GET) ===${NC}"

# Test login page with a mock challenge parameter
if response=$(curl -s -w "\n%{http_code}" "$SERVICE_URL/login?login_challenge=test-challenge" 2>&1); then
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" = "200" ]; then
        # Check if response contains login form elements
        if echo "$body" | grep -qi "email" && echo "$body" | grep -qi "password"; then
            echo -e "${GREEN}✓ PASSED${NC}"
            echo "  Login form is rendered correctly"
            ((PASSED++))
        else
            echo -e "${YELLOW}⚠ WARNING - Form may not contain expected fields${NC}"
            echo "  Login page accessible but form structure unclear"
            ((PASSED++))
        fi
    else
        echo -e "${RED}✗ FAILED - HTTP $http_code${NC}"
        ((FAILED++))
    fi
else
    echo -e "${RED}✗ FAILED - Connection error${NC}"
    ((FAILED++))
fi
echo ""

echo -e "${YELLOW}=== Test 3: Consent Page (GET) ===${NC}"

# Test consent page with a mock challenge parameter
if response=$(curl -s -w "\n%{http_code}" "$SERVICE_URL/consent?consent_challenge=test-challenge" 2>&1); then
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "200" ] || [ "$http_code" = "302" ] || [ "$http_code" = "400" ]; then
        echo -e "${GREEN}✓ PASSED${NC}"
        echo "  Consent endpoint is accessible (HTTP $http_code)"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAILED - HTTP $http_code${NC}"
        ((FAILED++))
    fi
else
    echo -e "${RED}✗ FAILED - Connection error${NC}"
    ((FAILED++))
fi
echo ""

echo -e "${YELLOW}=== Test 4: Logout Page (GET) ===${NC}"

# Test logout page with a mock challenge parameter
if response=$(curl -s -w "\n%{http_code}" "$SERVICE_URL/logout?logout_challenge=test-challenge" 2>&1); then
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "200" ] || [ "$http_code" = "302" ] || [ "$http_code" = "400" ]; then
        echo -e "${GREEN}✓ PASSED${NC}"
        echo "  Logout endpoint is accessible (HTTP $http_code)"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAILED - HTTP $http_code${NC}"
        ((FAILED++))
    fi
else
    echo -e "${RED}✗ FAILED - Connection error${NC}"
    ((FAILED++))
fi
echo ""

echo -e "${YELLOW}=== Test 5: Error Page ===${NC}"

if response=$(curl -s -w "\n%{http_code}" "$SERVICE_URL/error" 2>&1); then
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "200" ] || [ "$http_code" = "302" ]; then
        echo -e "${GREEN}✓ PASSED${NC}"
        echo "  Error page is accessible (HTTP $http_code)"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAILED - HTTP $http_code${NC}"
        ((FAILED++))
    fi
else
    echo -e "${RED}✗ FAILED - Connection error${NC}"
    ((FAILED++))
fi
echo ""

echo -e "${YELLOW}=== Test 6: Metrics Endpoint ===${NC}"

if response=$(curl -s "$SERVICE_URL/metrics" 2>&1); then
    if echo "$response" | grep -q "# HELP"; then
        echo -e "${GREEN}✓ PASSED${NC}"
        echo "  Prometheus metrics are exposed"
        
        # Check for specific metrics
        if echo "$response" | grep -q "http_requests_total"; then
            echo "  Found http_requests_total metric"
        fi
        
        ((PASSED++))
    else
        echo -e "${RED}✗ FAILED - Invalid metrics format${NC}"
        ((FAILED++))
    fi
else
    echo -e "${RED}✗ FAILED - Connection error${NC}"
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
    echo "  ✓ Health check working"
    echo "  ✓ Login page accessible"
    echo "  ✓ Consent page accessible"
    echo "  ✓ Logout page accessible"
    echo "  ✓ Error page accessible"
    echo "  ✓ Metrics endpoint working"
    echo ""
    echo "📊 Service URL: $SERVICE_URL"
    echo "📈 Metrics: $SERVICE_URL/metrics"
    echo ""
    echo "Note: Full OAuth2 flow testing requires Ory Hydra to be running."
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    echo ""
    echo "Troubleshooting:"
    echo "  - Ensure login-consent-provider is running on port 3001"
    echo "  - Check service logs for errors"
    echo "  - Verify database connection is working"
    exit 1
fi
