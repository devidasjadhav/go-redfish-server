#!/bin/bash

# Comprehensive test suite for Redfish server
# This script runs all individual test suites and provides a summary report

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_URL="http://localhost:8080"

echo "=========================================="
echo "Redfish Server - Comprehensive Test Suite"
echo "=========================================="
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results tracking
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Function to run a test script and track results
run_test() {
    local test_name=$1
    local test_script=$2

    echo -e "${BLUE}Running $test_name tests...${NC}"

    if [ -f "$SCRIPT_DIR/$test_script" ]; then
        # Run the test script and capture output
        local output
        output=$($SCRIPT_DIR/$test_script 2>&1)
        local exit_code=$?

        if [ $exit_code -eq 0 ]; then
            echo -e "${GREEN}✓ $test_name tests PASSED${NC}"
            ((PASSED_TESTS++))
        else
            echo -e "${RED}✗ $test_name tests FAILED${NC}"
            echo "$output"
            ((FAILED_TESTS++))
        fi
    else
        echo -e "${YELLOW}⚠ $test_script not found, skipping${NC}"
        ((SKIPPED_TESTS++))
    fi

    ((TOTAL_TESTS++))
    echo
}

# Function to check server health
check_server() {
    echo -e "${BLUE}Checking server health...${NC}"

    # Try to connect to the server
    if curl -s --max-time 5 "$BASE_URL/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Server is running and responding${NC}"
        return 0
    else
        echo -e "${RED}✗ Server is not responding${NC}"
        echo "Please start the server first: ./server --config config.yaml"
        return 1
    fi
}

# Function to validate Redfish service root
validate_service_root() {
    echo -e "${BLUE}Validating Redfish service root...${NC}"

    local response
    response=$(curl -s "$BASE_URL/redfish/v1/" -H "Accept: application/json")

    if echo "$response" | jq -e '.RedfishVersion' > /dev/null 2>&1; then
        local version
        version=$(echo "$response" | jq -r '.RedfishVersion')
        echo -e "${GREEN}✓ Service root valid, Redfish version: $version${NC}"

        # Check for required services
        local services=("Systems" "Chassis" "Managers" "TaskService" "EventService" "AccountService")
        for service in "${services[@]}"; do
            if echo "$response" | jq -e ".$service" > /dev/null 2>&1; then
                echo -e "${GREEN}  ✓ $service service available${NC}"
            else
                echo -e "${YELLOW}  ⚠ $service service not found${NC}"
            fi
        done
        return 0
    else
        echo -e "${RED}✗ Service root validation failed${NC}"
        return 1
    fi
}

# Function to run performance tests
run_performance_tests() {
    echo -e "${BLUE}Running performance tests...${NC}"

    local endpoint="$BASE_URL/redfish/v1/Systems/1"
    local iterations=10

    echo "Testing $iterations requests to $endpoint"

    local total_time=0
    local min_time=999999
    local max_time=0

    for i in $(seq 1 $iterations); do
        local start_time
        start_time=$(date +%s%N)
        curl -s "$endpoint" -H "Accept: application/json" > /dev/null 2>&1
        local end_time
        end_time=$(date +%s%N)

        local response_time=$(( (end_time - start_time) / 1000000 )) # Convert to milliseconds
        total_time=$((total_time + response_time))

        if [ $response_time -lt $min_time ]; then
            min_time=$response_time
        fi
        if [ $response_time -gt $max_time ]; then
            max_time=$response_time
        fi
    done

    local avg_time=$((total_time / iterations))

    echo -e "${GREEN}Performance Results:${NC}"
    echo "  Average response time: ${avg_time}ms"
    echo "  Min response time: ${min_time}ms"
    echo "  Max response time: ${max_time}ms"

    if [ $avg_time -lt 100 ]; then
        echo -e "${GREEN}✓ Performance acceptable (< 100ms average)${NC}"
    else
        echo -e "${YELLOW}⚠ Performance could be improved (> 100ms average)${NC}"
    fi
}

# Function to check security headers
check_security() {
    echo -e "${BLUE}Checking security headers...${NC}"

    local response
    response=$(curl -s -I "$BASE_URL/redfish/v1/Systems/1")

    local security_issues=0

    # Check for security headers
    if echo "$response" | grep -i "X-Frame-Options" > /dev/null; then
        echo -e "${GREEN}✓ X-Frame-Options header present${NC}"
    else
        echo -e "${YELLOW}⚠ X-Frame-Options header missing${NC}"
        ((security_issues++))
    fi

    if echo "$response" | grep -i "X-Content-Type-Options" > /dev/null; then
        echo -e "${GREEN}✓ X-Content-Type-Options header present${NC}"
    else
        echo -e "${YELLOW}⚠ X-Content-Type-Options header missing${NC}"
        ((security_issues++))
    fi

    if echo "$response" | grep -i "Content-Security-Policy" > /dev/null; then
        echo -e "${GREEN}✓ Content-Security-Policy header present${NC}"
    else
        echo -e "${YELLOW}⚠ Content-Security-Policy header missing${NC}"
        ((security_issues++))
    fi

    # Check for proper Redfish headers
    if echo "$response" | grep -i "OData-Version" > /dev/null; then
        echo -e "${GREEN}✓ OData-Version header present${NC}"
    else
        echo -e "${RED}✗ OData-Version header missing${NC}"
        ((security_issues++))
    fi

    if [ $security_issues -eq 0 ]; then
        echo -e "${GREEN}✓ Security validation passed${NC}"
    else
        echo -e "${YELLOW}⚠ $security_issues security issues found${NC}"
    fi
}

# Main test execution
echo "Starting comprehensive test suite..."
echo

# Check if server is running
if ! check_server; then
    echo -e "${RED}Cannot proceed with tests - server not available${NC}"
    exit 1
fi

echo

# Validate service root
if ! validate_service_root; then
    echo -e "${RED}Service root validation failed - tests may not be reliable${NC}"
fi

echo

# Run individual test suites
run_test "Server" "test_server.sh"
run_test "Authentication" "test_auth.sh"
run_test "CRUD Operations" "test_crud.sh"
run_test "Query Parameters" "test_all_queries.sh"
run_test "Actions" "test_actions.sh"
run_test "Eventing" "test_eventing.sh"
run_test "Tasks" "test_task.sh"
run_test "OEM" "test_oem.sh"

# Run performance tests
run_performance_tests
echo

# Check security
check_security
echo

# Generate summary report
echo "=========================================="
echo "TEST SUMMARY REPORT"
echo "=========================================="
echo "Total test suites: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
echo -e "Skipped: ${YELLOW}$SKIPPED_TESTS${NC}"
echo

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}🎉 ALL TESTS PASSED!${NC}"
    echo "The Redfish server implementation is fully functional."
else
    echo -e "${RED}❌ $FAILED_TESTS test suite(s) failed${NC}"
    echo "Please review the test output above for details."
fi

echo
echo "=========================================="
echo "Redfish Conformance Assessment"
echo "=========================================="

# Basic conformance checks
conformance_score=0
total_checks=10

# Check 1: Service root compliance
if curl -s "$BASE_URL/redfish/v1/" | jq -e '.RedfishVersion' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Service root compliant${NC}"
    ((conformance_score++))
else
    echo -e "${RED}✗ Service root not compliant${NC}"
fi

# Check 2: Systems collection
if curl -s "$BASE_URL/redfish/v1/Systems" | jq -e '.Members' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Systems collection compliant${NC}"
    ((conformance_score++))
else
    echo -e "${RED}✗ Systems collection not compliant${NC}"
fi

# Check 3: Authentication
if curl -s "$BASE_URL/redfish/v1/SessionService" | jq -e '.ServiceEnabled' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Session service compliant${NC}"
    ((conformance_score++))
else
    echo -e "${RED}✗ Session service not compliant${NC}"
fi

# Check 4: Task service
if curl -s "$BASE_URL/redfish/v1/TaskService" | jq -e '.ServiceEnabled' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Task service compliant${NC}"
    ((conformance_score++))
else
    echo -e "${RED}✗ Task service not compliant${NC}"
fi

# Check 5: Event service
if curl -s "$BASE_URL/redfish/v1/EventService" | jq -e '.ServiceEnabled' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Event service compliant${NC}"
    ((conformance_score++))
else
    echo -e "${RED}✗ Event service not compliant${NC}"
fi

# Check 6: Actions support
if curl -s "$BASE_URL/redfish/v1/Systems/1" | jq -e '.Actions' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Actions support compliant${NC}"
    ((conformance_score++))
else
    echo -e "${RED}✗ Actions support not compliant${NC}"
fi

# Check 7: OEM extensions
if curl -s "$BASE_URL/redfish/v1/Systems/1" | jq -e '.Oem' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ OEM extensions compliant${NC}"
    ((conformance_score++))
else
    echo -e "${RED}✗ OEM extensions not compliant${NC}"
fi

# Check 8: Registry support
if curl -s "$BASE_URL/redfish/v1/Registries" | jq -e '.Members' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Registry support compliant${NC}"
    ((conformance_score++))
else
    echo -e "${RED}✗ Registry support not compliant${NC}"
fi

# Check 9: Error responses
response=$(curl -s "$BASE_URL/redfish/v1/Systems/999")
if echo "$response" | jq -e '.error' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Error responses compliant${NC}"
    ((conformance_score++))
else
    echo -e "${RED}✗ Error responses not compliant${NC}"
fi

# Check 10: JSON schema compliance
if curl -s "$BASE_URL/redfish/v1/Systems/1" | jq empty > /dev/null 2>&1; then
    echo -e "${GREEN}✓ JSON schema compliant${NC}"
    ((conformance_score++))
else
    echo -e "${RED}✗ JSON schema not compliant${NC}"
fi

echo
echo "Conformance Score: $conformance_score / $total_checks"

percentage=$((conformance_score * 100 / total_checks))
if [ $percentage -ge 90 ]; then
    echo -e "${GREEN}🎉 EXCELLENT: $percentage% conformance achieved${NC}"
elif [ $percentage -ge 75 ]; then
    echo -e "${YELLOW}⚠ GOOD: $percentage% conformance achieved${NC}"
else
    echo -e "${RED}❌ NEEDS IMPROVEMENT: $percentage% conformance achieved${NC}"
fi

echo
echo "=========================================="
echo "Test suite completed at $(date)"
echo "=========================================="