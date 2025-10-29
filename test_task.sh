#!/bin/bash

# Test script for Redfish Task Service functionality
# This script tests the TaskService, Tasks collection, and individual task management

BASE_URL="http://localhost:8080"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Testing Redfish Task Service functionality..."
echo "============================================"

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

# Test 1: Get TaskService
echo "Test 1: GET /redfish/v1/TaskService"
response=$(make_request "GET" "$BASE_URL/redfish/v1/TaskService")
if echo "$response" | jq -e '.ServiceEnabled' > /dev/null 2>&1; then
    echo "✓ TaskService retrieved successfully"
    echo "$response" | jq '.ServiceEnabled, .CompletedTaskOverWritePolicy, .LifeCycleEventOnTaskStateChange'
else
    echo "✗ Failed to retrieve TaskService"
    echo "Response: $response"
fi
echo

# Test 2: Get Tasks collection (should be empty initially)
echo "Test 2: GET /redfish/v1/TaskService/Tasks"
response=$(make_request "GET" "$BASE_URL/redfish/v1/TaskService/Tasks")
if echo "$response" | jq -e '.Members' > /dev/null 2>&1; then
    echo "✓ Tasks collection retrieved successfully"
    member_count=$(echo "$response" | jq '.Members | length')
    echo "Initial task count: $member_count"
else
    echo "✗ Failed to retrieve Tasks collection"
    echo "Response: $response"
fi
echo

# Test 3: Create a new task
echo "Test 3: POST /redfish/v1/TaskService/Tasks"
response=$(make_request "POST" "$BASE_URL/redfish/v1/TaskService/Tasks")
if echo "$response" | jq -e '.Id' > /dev/null 2>&1; then
    task_id=$(echo "$response" | jq -r '.Id')
    task_uri=$(echo "$response" | jq -r '.@odata.id')
    echo "✓ Task created successfully with ID: $task_id"
    echo "Task URI: $task_uri"
    echo "$response" | jq '.TaskState, .PercentComplete'
else
    echo "✗ Failed to create task"
    echo "Response: $response"
fi
echo

# Test 4: Get the created task
if [ -n "$task_id" ]; then
    echo "Test 4: GET /redfish/v1/TaskService/Tasks/$task_id"
    response=$(make_request "GET" "$BASE_URL/redfish/v1/TaskService/Tasks/$task_id")
    if echo "$response" | jq -e '.TaskState' > /dev/null 2>&1; then
        echo "✓ Task retrieved successfully"
        echo "$response" | jq '.TaskState, .PercentComplete, .StartTime'
    else
        echo "✗ Failed to retrieve task"
        echo "Response: $response"
    fi
    echo

    # Wait a bit for task to progress
    echo "Waiting 5 seconds for task to progress..."
    sleep 5

    # Test 5: Check task progress
    echo "Test 5: GET /redfish/v1/TaskService/Tasks/$task_id (after progress)"
    response=$(make_request "GET" "$BASE_URL/redfish/v1/TaskService/Tasks/$task_id")
    if echo "$response" | jq -e '.TaskState' > /dev/null 2>&1; then
        task_state=$(echo "$response" | jq -r '.TaskState')
        percent_complete=$(echo "$response" | jq -r '.PercentComplete')
        echo "✓ Task state: $task_state, Progress: $percent_complete%"
        if [ "$task_state" = "Completed" ] && [ "$percent_complete" = "100" ]; then
            echo "✓ Task completed successfully"
        fi
    else
        echo "✗ Failed to check task progress"
        echo "Response: $response"
    fi
    echo
fi

# Test 6: Test ComputerSystem.Reset action creates a task
echo "Test 6: POST /redfish/v1/Systems/1/Actions/ComputerSystem.Reset (creates task)"
reset_data='{"ResetType": "ForceRestart"}'
response=$(make_request "POST" "$BASE_URL/redfish/v1/Systems/1/Actions/ComputerSystem.Reset" "$reset_data")
if echo "$response" | jq -e '.Id' > /dev/null 2>&1; then
    reset_task_id=$(echo "$response" | jq -r '.Id')
    echo "✓ ComputerSystem.Reset created task with ID: $reset_task_id"
    echo "$response" | jq '.Name, .@odata.id'
else
    echo "✗ ComputerSystem.Reset failed to create task"
    echo "Response: $response"
fi
echo

# Test 7: Test Manager.Reset action creates a task
echo "Test 7: POST /redfish/v1/Managers/1/Actions/Manager.Reset (creates task)"
mgr_reset_data='{"ResetType": "GracefulRestart"}'
response=$(make_request "POST" "$BASE_URL/redfish/v1/Managers/1/Actions/Manager.Reset" "$mgr_reset_data")
if echo "$response" | jq -e '.Id' > /dev/null 2>&1; then
    mgr_task_id=$(echo "$response" | jq -r '.Id')
    echo "✓ Manager.Reset created task with ID: $mgr_task_id"
    echo "$response" | jq '.Name, .@odata.id'
else
    echo "✗ Manager.Reset failed to create task"
    echo "Response: $response"
fi
echo

# Test 8: Check updated Tasks collection
echo "Test 8: GET /redfish/v1/TaskService/Tasks (after creating tasks)"
response=$(make_request "GET" "$BASE_URL/redfish/v1/TaskService/Tasks")
if echo "$response" | jq -e '.Members' > /dev/null 2>&1; then
    member_count=$(echo "$response" | jq '.Members | length')
    echo "✓ Tasks collection now has $member_count tasks"
    echo "$response" | jq '.Members[]'
else
    echo "✗ Failed to retrieve updated Tasks collection"
    echo "Response: $response"
fi
echo

# Test 9: Delete a task
if [ -n "$task_id" ]; then
    echo "Test 9: DELETE /redfish/v1/TaskService/Tasks/$task_id"
    response=$(make_request "DELETE" "$BASE_URL/redfish/v1/TaskService/Tasks/$task_id")
    if [ $? -eq 0 ]; then
        echo "✓ Task deleted successfully"
    else
        echo "✗ Failed to delete task"
        echo "Response: $response"
    fi
    echo
fi

echo "Task Service tests completed!"
echo "Note: Tasks are stored in memory and will be lost on server restart."