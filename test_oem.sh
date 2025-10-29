#!/bin/bash

# Test script for Redfish OEM Extensions and Message Registries
# This script tests OEM properties, registry endpoints, and OEM functionality

BASE_URL="http://localhost:8080"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Testing Redfish OEM Extensions and Message Registries..."
echo "========================================================="

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
    fi
}

# Test 1: Get ComputerSystem with OEM properties
echo "Test 1: GET /redfish/v1/Systems/1 (with OEM properties)"
response=$(make_request "GET" "$BASE_URL/redfish/v1/Systems/1")
if echo "$response" | jq -e '.Oem.Contoso' > /dev/null 2>&1; then
    echo "✓ ComputerSystem includes OEM properties"
    echo "$response" | jq '.Oem.Contoso | {VendorId, ProductId, SerialNumber, FirmwareVersion, CustomProperties}'
else
    echo "✗ ComputerSystem missing OEM properties"
    echo "Response: $response"
fi
echo

# Test 2: Get Registries collection
echo "Test 2: GET /redfish/v1/Registries"
response=$(make_request "GET" "$BASE_URL/redfish/v1/Registries")
if echo "$response" | jq -e '.Members' > /dev/null 2>&1; then
    echo "✓ Registries collection retrieved successfully"
    member_count=$(echo "$response" | jq '.Members | length')
    echo "Registry count: $member_count"
    echo "$response" | jq '.Members[]'
else
    echo "✗ Failed to retrieve Registries collection"
    echo "Response: $response"
fi
echo

# Test 3: Get specific registry file
echo "Test 3: GET /redfish/v1/Registries/Base.1.0.0"
response=$(make_request "GET" "$BASE_URL/redfish/v1/Registries/Base.1.0.0")
if echo "$response" | jq -e '.Registry' > /dev/null 2>&1; then
    echo "✓ Registry file retrieved successfully"
    echo "$response" | jq '{Id, Name, Registry, Languages, Location: [.Location[] | {Language, Uri}]}'
else
    echo "✗ Failed to retrieve registry file"
    echo "Response: $response"
fi
echo

# Test 4: Get another registry file
echo "Test 4: GET /redfish/v1/Registries/Task.1.0.0"
response=$(make_request "GET" "$BASE_URL/redfish/v1/Registries/Task.1.0.0")
if echo "$response" | jq -e '.Registry' > /dev/null 2>&1; then
    echo "✓ Task registry file retrieved successfully"
    echo "$response" | jq '{Id, Name, Registry}'
else
    echo "✗ Failed to retrieve task registry file"
    echo "Response: $response"
fi
echo

# Test 5: Test non-existent registry
echo "Test 5: GET /redfish/v1/Registries/NonExistent.1.0.0"
response=$(make_request "GET" "$BASE_URL/redfish/v1/Registries/NonExistent.1.0.0")
if echo "$response" | grep -q "Registry not found"; then
    echo "✓ Correctly returned 404 for non-existent registry"
else
    echo "? Unexpected response for non-existent registry"
    echo "Response: $response"
fi
echo

# Test 6: Test OEM custom action
echo "Test 6: POST /redfish/v1/Oem/Contoso/CustomAction"
custom_action_data='{
    "Action": "CustomDiagnostic",
    "Parameters": {
        "TestMode": true,
        "Timeout": 30,
        "Verbose": false
    }
}'
response=$(make_request "POST" "$BASE_URL/redfish/v1/Oem/Contoso/CustomAction" "$custom_action_data")
if echo "$response" | jq -e '.Status' > /dev/null 2>&1; then
    echo "✓ OEM custom action executed successfully"
    echo "$response" | jq '{Action, Status, Message, Timestamp, Parameters}'
else
    echo "✗ OEM custom action failed"
    echo "Response: $response"
fi
echo

# Test 7: Test OEM custom action without parameters
echo "Test 7: POST /redfish/v1/Oem/Contoso/CustomAction (no parameters)"
simple_action_data='{"Action": "SimpleAction"}'
response=$(make_request "POST" "$BASE_URL/redfish/v1/Oem/Contoso/CustomAction" "$simple_action_data")
if echo "$response" | jq -e '.Status' > /dev/null 2>&1; then
    echo "✓ OEM custom action without parameters executed successfully"
    echo "$response" | jq '{Action, Status, Message}'
else
    echo "✗ OEM custom action without parameters failed"
    echo "Response: $response"
fi
echo

# Test 8: Verify OEM properties in different resources
echo "Test 8: Check OEM properties consistency"
system_response=$(make_request "GET" "$BASE_URL/redfish/v1/Systems/1")
if echo "$system_response" | jq -e '.Oem.Contoso.VendorId' > /dev/null 2>&1; then
    vendor_id=$(echo "$system_response" | jq -r '.Oem.Contoso.VendorId')
    echo "✓ OEM VendorId found: $vendor_id"
else
    echo "✗ OEM VendorId not found in ComputerSystem"
fi
echo

echo "OEM Extensions and Message Registries tests completed!"
echo "Note: OEM extensions provide vendor-specific customizations and extensions."