#!/bin/bash

echo "========================================="
echo "Portfolio Insights Monitoring Test"
echo "========================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Service metrics ports
declare -A SERVICES=(
    ["transaction-service"]="9097"
    ["gateway-service"]="9095"
    ["user-service"]="9096"
    ["portfolio-service"]="9098"
    ["marketdata-service"]="9099"
)

echo "=== Testing Service Metrics Endpoints ==="
echo ""

for service in "${!SERVICES[@]}"; do
    port="${SERVICES[$service]}"
    echo -n "Testing $service on port $port... "
    
    # Check if metrics endpoint is accessible
    if curl -s "http://localhost:$port/metrics" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Accessible${NC}"
        
        # Check for standard Go metrics
        if curl -s "http://localhost:$port/metrics" | grep -q "go_goroutines"; then
            echo "  ${GREEN}✓${NC} Go runtime metrics present"
        else
            echo "  ${RED}✗${NC} Go runtime metrics missing"
        fi
        
        # Check for gRPC metrics
        if curl -s "http://localhost:$port/metrics" | grep -q "grpc_requests_total"; then
            echo "  ${GREEN}✓${NC} gRPC metrics present"
        else
            echo "  ${YELLOW}⚠${NC} gRPC metrics not found (may not be initialized yet)"
        fi
        
        # Service-specific checks
        if [ "$service" == "transaction-service" ]; then
            if curl -s "http://localhost:$port/metrics" | grep -q "transactions_created_total"; then
                echo "  ${GREEN}✓${NC} Transaction business metrics present"
            else
                echo "  ${YELLOW}⚠${NC} Transaction business metrics not found (create a transaction to initialize)"
            fi
        fi
    else
        echo -e "${RED}✗ Not accessible${NC}"
        echo "  ${YELLOW}→${NC} Make sure the service is running"
    fi
    echo ""
done

echo "=== Testing Prometheus ==="
echo ""

echo -n "Checking Prometheus server... "
if curl -s "http://localhost:9090/-/healthy" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Running${NC}"
else
    echo -e "${RED}✗ Not running${NC}"
    echo "  ${YELLOW}→${NC} Start with: cd deployments/monitoring && docker-compose up -d"
    exit 1
fi

echo -n "Checking Prometheus targets... "
targets_response=$(curl -s "http://localhost:9090/api/v1/targets")
if echo "$targets_response" | jq -e '.data.activeTargets' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Targets configured${NC}"
    echo ""
    echo "Target Status:"
    echo "$targets_response" | jq -r '.data.activeTargets[] | "  \(.labels.job): \(.health)"' | while read line; do
        if echo "$line" | grep -q "up"; then
            echo -e "  ${GREEN}$line${NC}"
        else
            echo -e "  ${RED}$line${NC}"
        fi
    done
else
    echo -e "${RED}✗ Failed to retrieve targets${NC}"
fi

echo ""
echo "=== Testing Grafana ==="
echo ""

echo -n "Checking Grafana server... "
if curl -s "http://localhost:3001/api/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Running${NC}"
    
    echo -n "Checking Grafana datasources... "
    datasources=$(curl -s -u admin:admin "http://localhost:3001/api/datasources")
    if echo "$datasources" | jq -e '.[].name' > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Configured${NC}"
        echo "  Datasources:"
        echo "$datasources" | jq -r '.[].name' | sed 's/^/    /'
    else
        echo -e "${YELLOW}⚠ No datasources configured${NC}"
    fi
else
    echo -e "${RED}✗ Not running${NC}"
    echo "  ${YELLOW}→${NC} Start with: cd deployments/monitoring && docker-compose up -d"
fi

echo ""
echo "=== Testing AlertManager ==="
echo ""

echo -n "Checking AlertManager... "
if curl -s "http://localhost:9093/-/healthy" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Running${NC}"
else
    echo -e "${RED}✗ Not running${NC}"
fi

echo ""
echo "========================================="
echo "Test Summary"
echo "========================================="
echo ""
echo "Prometheus UI: http://localhost:9090"
echo "Grafana UI:    http://localhost:3001 (admin/admin)"
echo "AlertManager:  http://localhost:9093"
echo ""
echo "Next steps:"
echo "1. Create a transaction to generate metrics"
echo "2. View metrics in Prometheus"
echo "3. Create dashboards in Grafana"
echo ""
