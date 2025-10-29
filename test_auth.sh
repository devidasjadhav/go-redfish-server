#!/bin/bash

# Authentication test suite for Redfish server
# Tests basic auth, session auth, and access control

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_URL="http://localhost:8080"

echo "=========================================="
echo "Redfish Authentication Test Suite"
echo "=========================================="
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run a test
run_test() {
    local test_name=$1
    local command=$2
    local expected_exit=$3

    ((TOTAL_TESTS++))

    echo -n "[$test_name] "

    # Run the test
    eval "$command" > /dev/null 2>&1
    local actual_exit=$?

    if [ $actual_exit -eq $expected_exit ]; then
        echo -e "${GREEN}PASS${NC}"
        ((PASSED_TESTS++))
    else
        echo -e "${RED}FAIL${NC} (expected $expected_exit, got $actual_exit)"
        ((FAILED_TESTS++))
    fi
}

# Function to test HTTP response code
test_response_code() {
    local test_name=$1
    local url=$2
    local expected_code=$3
    local auth_header=$4

    ((TOTAL_TESTS++))

    echo -n "[$test_name] "

    local cmd="curl -s -o /dev/null -w '%{http_code}'"
    if [ -n "$auth_header" ]; then
        cmd="$cmd -H '$auth_header'"
    fi
    cmd="$cmd '$url'"

    local response_code
    response_code=$(eval "$cmd")

    if [ "$response_code" = "$expected_code" ]; then
        echo -e "${GREEN}PASS${NC}"
        ((PASSED_TESTS++))
    else
        echo -e "${RED}FAIL${NC} (expected $expected_code, got $response_code)"
        ((FAILED_TESTS++))
    fi
}

# Function to test JSON response contains property
test_json_property() {
    local test_name=$1
    local url=$2
    local property=$3
    local auth_header=$4

    ((TOTAL_TESTS++))

    echo -n "[$test_name] "

    local cmd="curl -s"
    if [ -n "$auth_header" ]; then
        cmd="$cmd -H '$auth_header'"
    fi
    cmd="$cmd '$url'"

    local response
    response=$(eval "$cmd")

    if echo "$response" | jq -e "$property" > /dev/null 2>&1; then
        echo -e "${GREEN}PASS${NC}"
        ((PASSED_TESTS++))
    else
        echo -e "${RED}FAIL${NC} (property $property not found)"
        ((FAILED_TESTS++))
    fi
}

echo "Testing server availability..."
if ! curl -s --max-time 5 "$BASE_URL/health" > /dev/null 2>&1; then
    echo -e "${RED}Server not available. Please start the server first.${NC}"
    exit 1
fi
echo -e "${GREEN}Server is running${NC}"
echo

echo "=========================================="
echo "Public Endpoint Access Tests"
echo "=========================================="

# Test public endpoints (should work without auth)
test_response_code "Public: Health endpoint" "$BASE_URL/health" "200"
test_response_code "Public: Service root" "$BASE_URL/redfish/v1/" "200"
test_response_code "Public: Metadata" "$BASE_URL/redfish/v1/\$metadata" "200"
test_response_code "Public: OData service" "$BASE_URL/redfish/v1/odata" "200"
test_response_code "Public: Session service" "$BASE_URL/redfish/v1/SessionService" "200"
test_response_code "Public: Sessions collection" "$BASE_URL/redfish/v1/SessionService/Sessions" "200"

echo
echo "=========================================="
echo "Protected Endpoint Access Tests (No Auth)"
echo "=========================================="

# Test protected endpoints without auth (should fail)
test_response_code "Protected: Systems (no auth)" "$BASE_URL/redfish/v1/Systems/1" "401"
test_response_code "Protected: Chassis (no auth)" "$BASE_URL/redfish/v1/Chassis/1" "401"
test_response_code "Protected: Managers (no auth)" "$BASE_URL/redfish/v1/Managers/1" "401"
test_response_code "Protected: Account service (no auth)" "$BASE_URL/redfish/v1/AccountService" "401"

echo
echo "=========================================="
echo "Basic Authentication Tests"
echo "=========================================="

# Test basic auth with valid credentials
test_response_code "Basic Auth: Valid admin" "$BASE_URL/redfish/v1/Systems/1" "200" "Authorization: Basic $(echo -n 'admin:password' | base64)"
test_response_code "Basic Auth: Valid operator" "$BASE_URL/redfish/v1/Systems/1" "200" "Authorization: Basic $(echo -n 'operator:password' | base64)"

# Test basic auth with invalid credentials
test_response_code "Basic Auth: Invalid user" "$BASE_URL/redfish/v1/Systems/1" "401" "Authorization: Basic $(echo -n 'invalid:password' | base64)"
test_response_code "Basic Auth: Wrong password" "$BASE_URL/redfish/v1/Systems/1" "401" "Authorization: Basic $(echo -n 'admin:wrong' | base64)"
test_response_code "Basic Auth: Empty credentials" "$BASE_URL/redfish/v1/Systems/1" "401" "Authorization: Basic $(echo -n '' | base64)"

echo
echo "=========================================="
echo "Session Authentication Tests"
echo "=========================================="

# Create a session for admin user
echo -n "[Session: Create admin session] "
SESSION_TOKEN=$(curl -s -X POST \
    -H "Authorization: Basic $(echo -n 'admin:password' | base64)" \
    -H "Content-Type: application/json" \
    "$BASE_URL/redfish/v1/SessionService/Sessions" | jq -r '.Id' 2>/dev/null)

if [ -n "$SESSION_TOKEN" ] && [ "$SESSION_TOKEN" != "null" ]; then
    echo -e "${GREEN}PASS${NC}"
    ((PASSED_TESTS++))
    ((TOTAL_TESTS++))
else
    echo -e "${RED}FAIL${NC} (could not create session)"
    ((FAILED_TESTS++))
    ((TOTAL_TESTS++))
    SESSION_TOKEN=""
fi

# Test session authentication if session was created
if [ -n "$SESSION_TOKEN" ]; then
    test_response_code "Session Auth: Systems" "$BASE_URL/redfish/v1/Systems/1" "200" "X-Auth-Token: $SESSION_TOKEN"
    test_response_code "Session Auth: Chassis" "$BASE_URL/redfish/v1/Chassis/1" "200" "X-Auth-Token: $SESSION_TOKEN"
    test_response_code "Session Auth: Managers" "$BASE_URL/redfish/v1/Managers/1" "200" "X-Auth-Token: $SESSION_TOKEN"
    test_response_code "Session Auth: Account service" "$BASE_URL/redfish/v1/AccountService" "200" "X-Auth-Token: $SESSION_TOKEN"

    # Test invalid session token
    test_response_code "Session Auth: Invalid token" "$BASE_URL/redfish/v1/Systems/1" "401" "X-Auth-Token: invalid-token-12345"
else
    echo -e "${YELLOW}Skipping session tests due to session creation failure${NC}"
fi

echo
echo "=========================================="
echo "Session Service Tests"
echo "=========================================="

# Test session service properties
test_json_property "Session Service: ServiceEnabled" "$BASE_URL/redfish/v1/SessionService" ".ServiceEnabled" ""
test_json_property "Session Service: Sessions link" "$BASE_URL/redfish/v1/SessionService" ".Sessions" ""

# Test sessions collection
test_json_property "Sessions Collection: Members array" "$BASE_URL/redfish/v1/SessionService/Sessions" ".Members" ""

echo
echo "=========================================="
echo "Account Service Tests"
echo "=========================================="

# Test account service with session auth (if available)
if [ -n "$SESSION_TOKEN" ]; then
    test_json_property "Account Service: ServiceEnabled" "$BASE_URL/redfish/v1/AccountService" ".ServiceEnabled" "X-Auth-Token: $SESSION_TOKEN"
    test_json_property "Account Service: Accounts link" "$BASE_URL/redfish/v1/AccountService" ".Accounts" "X-Auth-Token: $SESSION_TOKEN"

    # Test accounts collection
    test_json_property "Accounts Collection: Members" "$BASE_URL/redfish/v1/AccountService/Accounts" ".Members" "X-Auth-Token: $SESSION_TOKEN"
else
    echo -e "${YELLOW}Skipping account service tests due to no valid session${NC}"
fi

echo
echo "=========================================="
echo "Authentication Test Summary"
echo "=========================================="
echo "Total Tests: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed: ${RED}$FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}🎉 ALL AUTHENTICATION TESTS PASSED!${NC}"
    echo "Authentication implementation is working correctly."
    exit 0
else
    echo -e "${RED}❌ $FAILED_TESTS authentication test(s) failed${NC}"
    echo "Please review the authentication implementation."
    exit 1
fi