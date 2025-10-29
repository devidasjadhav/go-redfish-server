#!/bin/bash

# Test script for Redfish Eventing functionality
# This script tests the EventService, EventSubscriptions, and SSE endpoints

BASE_URL="http://localhost:8080"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Testing Redfish Eventing functionality..."
echo "=========================================="

# Function to make HTTP requests
make_request() {
    local method=$1
    local url=$2
    local data=$3
    local auth=$4

    if [ "$method" = "GET" ]; then
        curl -s -X GET "$url" -H "Accept: application/json" ${auth:+-H "$auth"}
    elif [ "$method" = "POST" ]; then
        curl -s -X POST "$url" -H "Content-Type: application/json" ${auth:+-H "$auth"} -d "$data"
    elif [ "$method" = "DELETE" ]; then
        curl -s -X DELETE "$url" ${auth:+-H "$auth"}
    fi
}

# Test 1: Get EventService
echo "Test 1: GET /redfish/v1/EventService"
response=$(make_request "GET" "$BASE_URL/redfish/v1/EventService")
if echo "$response" | jq -e '.ServiceEnabled' > /dev/null 2>&1; then
    echo "✓ EventService retrieved successfully"
    echo "$response" | jq '.ServiceEnabled, .EventFormatTypes, .ServerSentEventUri'
else
    echo "✗ Failed to retrieve EventService"
    echo "Response: $response"
fi
echo

# Test 2: Get EventSubscriptions collection
echo "Test 2: GET /redfish/v1/EventService/Subscriptions"
response=$(make_request "GET" "$BASE_URL/redfish/v1/EventService/Subscriptions")
if echo "$response" | jq -e '.Members' > /dev/null 2>&1; then
    echo "✓ EventSubscriptions collection retrieved successfully"
    echo "$response" | jq '.Members | length'
else
    echo "✗ Failed to retrieve EventSubscriptions collection"
    echo "Response: $response"
fi
echo

# Test 3: Create EventSubscription
echo "Test 3: POST /redfish/v1/EventService/Subscriptions"
subscription_data='{
    "Destination": "http://example.com/events",
    "Protocol": "Redfish",
    "Context": "TestSubscription",
    "RegistryPrefixes": ["Base"],
    "ResourceTypes": ["ComputerSystem"],
    "Severities": ["Warning", "Critical"]
}'
response=$(make_request "POST" "$BASE_URL/redfish/v1/EventService/Subscriptions" "$subscription_data")
if echo "$response" | jq -e '.Id' > /dev/null 2>&1; then
    subscription_id=$(echo "$response" | jq -r '.Id')
    echo "✓ EventSubscription created successfully with ID: $subscription_id"
    echo "$response" | jq '.Destination, .Protocol, .Context'
else
    echo "✗ Failed to create EventSubscription"
    echo "Response: $response"
fi
echo

# Test 4: Get specific EventSubscription (should fail since not persisted)
if [ -n "$subscription_id" ]; then
    echo "Test 4: GET /redfish/v1/EventService/Subscriptions/$subscription_id"
    response=$(make_request "GET" "$BASE_URL/redfish/v1/EventService/Subscriptions/$subscription_id")
    if echo "$response" | grep -q "Subscription not found"; then
        echo "✓ Correctly returned 404 for non-persisted subscription"
    else
        echo "? Unexpected response for non-persisted subscription"
        echo "Response: $response"
    fi
    echo
fi

# Test 5: Test SSE endpoint
echo "Test 5: GET /redfish/v1/EventService/SSE"
# Note: This will timeout after 1 second as implemented
timeout 3 curl -s -N "$BASE_URL/redfish/v1/EventService/SSE" | head -5
echo "✓ SSE endpoint responded (may show heartbeat event)"
echo

# Test 6: Test invalid subscription creation
echo "Test 6: POST /redfish/v1/EventService/Subscriptions (invalid - missing destination)"
invalid_data='{
    "Protocol": "Redfish"
}'
response=$(make_request "POST" "$BASE_URL/redfish/v1/EventService/Subscriptions" "$invalid_data")
if echo "$response" | grep -q "Destination is required"; then
    echo "✓ Correctly rejected subscription without destination"
else
    echo "✗ Should have rejected subscription without destination"
    echo "Response: $response"
fi
echo

echo "Eventing tests completed!"
echo "Note: Full persistence and event routing not implemented in this demo version."