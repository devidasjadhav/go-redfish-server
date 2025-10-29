#!/bin/bash

# Redfish Conformance Validation Script
# Validates implementation against DSP0266 and DSP0268 specifications

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_URL="http://localhost:8080"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "=========================================="
echo "Redfish Conformance Validation"
echo "=========================================="
echo

# Validation counters
total_checks=0
passed_checks=0
failed_checks=0
warnings=0

# Function to check a requirement
check_requirement() {
    local requirement=$1
    local description=$2
    local command=$3
    local expected_exit=$4

    ((total_checks++))

    echo -n "[$requirement] $description... "

    # Run the check
    eval "$command" > /dev/null 2>&1
    local actual_exit=$?

    if [ $actual_exit -eq $expected_exit ]; then
        echo -e "${GREEN}PASS${NC}"
        ((passed_checks++))
    else
        echo -e "${RED}FAIL${NC}"
        ((failed_checks++))
    fi
}

# Function to check JSON property
check_json_property() {
    local url=$1
    local property=$2
    local requirement=$3
    local description=$4
    local auth_required=${5:-false}

    ((total_checks++))

    echo -n "[$requirement] $description... "

    local response
    if [ "$auth_required" = "true" ]; then
        response=$(curl -s "$url" -H "Accept: application/json" -H "Authorization: Basic $(echo -n 'admin:password' | base64)")
    else
        response=$(curl -s "$url" -H "Accept: application/json")
    fi

    if echo "$response" | jq -e "$property" > /dev/null 2>&1; then
        echo -e "${GREEN}PASS${NC}"
        ((passed_checks++))
    else
        echo -e "${RED}FAIL${NC}"
        ((failed_checks++))
    fi
}

# Function to check HTTP header
check_http_header() {
    local url=$1
    local header=$2
    local requirement=$3
    local description=$4

    ((total_checks++))

    echo -n "[$requirement] $description... "

    local response
    response=$(curl -s -I "$url")

    if echo "$response" | grep -i "$header" > /dev/null; then
        echo -e "${GREEN}PASS${NC}"
        ((passed_checks++))
    else
        echo -e "${RED}FAIL${NC}"
        ((failed_checks++))
    fi
}

# Function to check response code
check_response_code() {
    local url=$1
    local expected_code=$2
    local requirement=$3
    local description=$4

    ((total_checks++))

    echo -n "[$requirement] $description... "

    local response_code
    response_code=$(curl -s -o /dev/null -w "%{http_code}" "$url")

    if [ "$response_code" = "$expected_code" ]; then
        echo -e "${GREEN}PASS${NC}"
        ((passed_checks++))
    else
        echo -e "${RED}FAIL${NC} (got $response_code, expected $expected_code)"
        ((failed_checks++))
    fi
}

echo "Checking server availability..."
if ! curl -s --max-time 5 "$BASE_URL/health" > /dev/null 2>&1; then
    echo -e "${RED}Server not available. Please start the server first.${NC}"
    exit 1
fi
echo -e "${GREEN}Server is running${NC}"
echo

echo "=========================================="
echo "DSP0266 (Redfish Protocol) Conformance"
echo "=========================================="

# 4.1 Service Root
check_json_property "$BASE_URL/redfish/v1/" ".RedfishVersion" "DSP0266-4.1.1" "Service root contains RedfishVersion"
check_json_property "$BASE_URL/redfish/v1/" ".UUID" "DSP0266-4.1.2" "Service root contains UUID"
check_json_property "$BASE_URL/redfish/v1/" ".Systems" "DSP0266-4.1.3" "Service root contains Systems link"
check_json_property "$BASE_URL/redfish/v1/" ".Chassis" "DSP0266-4.1.4" "Service root contains Chassis link"
check_json_property "$BASE_URL/redfish/v1/" ".Managers" "DSP0266-4.1.5" "Service root contains Managers link"

# 4.2 Metadata
check_response_code "$BASE_URL/redfish/v1/\$metadata" "200" "DSP0266-4.2" "Metadata endpoint returns 200"

# 4.3 OData Service Document
check_response_code "$BASE_URL/redfish/v1/odata" "200" "DSP0266-4.3" "OData service document returns 200"

# 5.1 HTTP Headers
check_http_header "$BASE_URL/redfish/v1/" "OData-Version" "DSP0266-5.1.1" "OData-Version header present"
check_http_header "$BASE_URL/redfish/v1/" "Content-Type" "DSP0266-5.1.2" "Content-Type header present"

# 6.1 Authentication
check_response_code "$BASE_URL/redfish/v1/SessionService" "200" "DSP0266-6.1" "Session service accessible"

# 7.1 Resources
check_json_property "$BASE_URL/redfish/v1/Systems/1" ".Id" "DSP0266-7.1.1" "Resource contains Id property" "true"
check_json_property "$BASE_URL/redfish/v1/Systems/1" ".Name" "DSP0266-7.1.2" "Resource contains Name property" "true"
check_json_property "$BASE_URL/redfish/v1/Systems/1" ".\"@odata.id\"" "DSP0266-7.1.3" "Resource contains @odata.id" "true"
check_json_property "$BASE_URL/redfish/v1/Systems/1" ".\"@odata.type\"" "DSP0266-7.1.4" "Resource contains @odata.type" "true"

# 7.2 Collections
check_json_property "$BASE_URL/redfish/v1/Systems" ".Members" "DSP0266-7.2.1" "Collection contains Members array" "true"
check_json_property "$BASE_URL/redfish/v1/Systems" ".\"Members@odata.count\"" "DSP0266-7.2.2" "Collection contains member count" "true"

# 7.3 Actions
check_json_property "$BASE_URL/redfish/v1/Systems/1" ".Actions" "DSP0266-7.3" "Resource contains Actions property" "true"

# 8.1 Error Responses
check_response_code "$BASE_URL/redfish/v1/Systems/999" "401" "DSP0266-8.1" "Invalid resource returns 401 (auth required)"
check_json_property "$BASE_URL/redfish/v1/Systems/999" ".error" "DSP0266-8.1.1" "Error response contains error object"

echo
echo "=========================================="
echo "DSP0268 (Redfish Data Model) Conformance"
echo "=========================================="

# ComputerSystem Schema
check_json_property "$BASE_URL/redfish/v1/Systems/1" ".SystemType" "DSP0268-ComputerSystem" "ComputerSystem has SystemType" "true"
check_json_property "$BASE_URL/redfish/v1/Systems/1" ".PowerState" "DSP0268-ComputerSystem" "ComputerSystem has PowerState" "true"
check_json_property "$BASE_URL/redfish/v1/Systems/1" ".Boot" "DSP0268-ComputerSystem" "ComputerSystem has Boot configuration" "true"

# Manager Schema
check_json_property "$BASE_URL/redfish/v1/Managers/1" ".ManagerType" "DSP0268-Manager" "Manager has ManagerType" "true"
check_json_property "$BASE_URL/redfish/v1/Managers/1" ".Status" "DSP0268-Manager" "Manager has Status" "true"

# Task Service
check_json_property "$BASE_URL/redfish/v1/TaskService" ".ServiceEnabled" "DSP0268-TaskService" "TaskService has ServiceEnabled" "true"
check_json_property "$BASE_URL/redfish/v1/TaskService" ".Tasks" "DSP0268-TaskService" "TaskService has Tasks link" "true"

# Event Service
check_json_property "$BASE_URL/redfish/v1/EventService" ".ServiceEnabled" "DSP0268-EventService" "EventService has ServiceEnabled" "true"
check_json_property "$BASE_URL/redfish/v1/EventService" ".Links.Subscriptions" "DSP0268-EventService" "EventService has Subscriptions link" "true"

# Account Service
check_json_property "$BASE_URL/redfish/v1/AccountService" ".ServiceEnabled" "DSP0268-AccountService" "AccountService has ServiceEnabled" "true"
check_json_property "$BASE_URL/redfish/v1/AccountService" ".Accounts" "DSP0268-AccountService" "AccountService has Accounts link" "true"

echo
echo "=========================================="
echo "Security Validation"
echo "=========================================="

# Security checks
check_requirement "SEC-1" "Server uses HTTPS" "curl -s --max-time 5 https://localhost:8443/health > /dev/null 2>&1" 0
check_requirement "SEC-2" "HTTP redirects to HTTPS" "curl -s -I http://localhost:8080/ | grep -i 'location.*https' > /dev/null" 0
check_http_header "$BASE_URL/redfish/v1/" "X-Frame-Options" "SEC-3" "X-Frame-Options header present"
check_http_header "$BASE_URL/redfish/v1/" "X-Content-Type-Options" "SEC-4" "X-Content-Type-Options header present"

echo
echo "=========================================="
echo "Performance Validation"
echo "=========================================="

# Performance checks
echo -n "[PERF-1] Response time under 100ms... "
start_time=$(date +%s%N)
curl -s "$BASE_URL/redfish/v1/Systems/1" -H "Authorization: Basic $(echo -n 'admin:password' | base64)" > /dev/null
end_time=$(date +%s%N)
response_time=$(( (end_time - start_time) / 1000000 ))

((total_checks++))
if [ $response_time -lt 100 ]; then
    echo -e "${GREEN}PASS${NC} (${response_time}ms)"
    ((passed_checks++))
else
    echo -e "${YELLOW}WARN${NC} (${response_time}ms)"
    ((warnings++))
fi

echo
echo "=========================================="
echo "CONFORMANCE VALIDATION SUMMARY"
echo "=========================================="
echo "Total Checks: $total_checks"
echo -e "Passed: ${GREEN}$passed_checks${NC}"
echo -e "Failed: ${RED}$failed_checks${NC}"
echo -e "Warnings: ${YELLOW}$warnings${NC}"
echo

# Calculate compliance percentage
if [ $total_checks -gt 0 ]; then
    compliance=$(( (passed_checks * 100) / total_checks ))
    echo "Compliance Score: $compliance%"

    if [ $compliance -ge 95 ]; then
        echo -e "${GREEN}🎉 EXCELLENT COMPLIANCE${NC}"
        echo "Implementation is highly conformant with Redfish specifications."
    elif [ $compliance -ge 85 ]; then
        echo -e "${GREEN}✅ GOOD COMPLIANCE${NC}"
        echo "Implementation is conformant with minor issues."
    elif [ $compliance -ge 75 ]; then
        echo -e "${YELLOW}⚠ FAIR COMPLIANCE${NC}"
        echo "Implementation has some conformance issues to address."
    else
        echo -e "${RED}❌ POOR COMPLIANCE${NC}"
        echo "Implementation needs significant conformance improvements."
    fi
fi

echo
echo "=========================================="
echo "Recommendations"
echo "=========================================="

if [ $failed_checks -gt 0 ]; then
    echo -e "${RED}Failed Checks:${NC}"
    echo "Review the failed tests above and implement missing features."
    echo
fi

if [ $warnings -gt 0 ]; then
    echo -e "${YELLOW}Warnings:${NC}"
    echo "Consider addressing performance and optional security headers."
    echo
fi

echo "For full Redfish certification, ensure:"
echo "• All DSP0266 and DSP0268 requirements are met"
echo "• Proper error handling and status codes"
echo "• Complete schema compliance"
echo "• Security best practices implementation"
echo "• Performance meets application requirements"

echo
echo "=========================================="
echo "Validation completed at $(date)"
echo "=========================================="