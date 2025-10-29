#!/bin/bash

echo "Testing Redfish Actions Implementation"
echo "====================================="

# Start server
go run cmd/server/main.go &
SERVER_PID=$!
sleep 2

echo "1. Testing ComputerSystem.Reset action (default parameters):"
curl -k -s -u admin:password -X POST \
  -H "Content-Type: application/json" \
  -w "%{http_code}" \
  https://localhost:8443/redfish/v1/Systems/1/Actions/ComputerSystem.Reset

echo -e "\n2. Testing ComputerSystem.Reset action (with ResetType):"
curl -k -s -u admin:password -X POST \
  -H "Content-Type: application/json" \
  -d '{"ResetType": "ForceRestart"}' \
  -w "%{http_code}" \
  https://localhost:8443/redfish/v1/Systems/1/Actions/ComputerSystem.Reset

echo -e "\n3. Testing ComputerSystem.Reset action (invalid ResetType):"
curl -k -s -u admin:password -X POST \
  -H "Content-Type: application/json" \
  -d '{"ResetType": "InvalidType"}' \
  https://localhost:8443/redfish/v1/Systems/1/Actions/ComputerSystem.Reset | jq '.error.code'

echo -e "\n4. Testing Manager.Reset action:"
curl -k -s -u admin:password -X POST \
  -H "Content-Type: application/json" \
  -w "%{http_code}" \
  https://localhost:8443/redfish/v1/Managers/1/Actions/Manager.Reset

echo -e "\n5. Testing invalid action:"
curl -k -s -u admin:password -X POST \
  -H "Content-Type: application/json" \
  https://localhost:8443/redfish/v1/Systems/1/Actions/InvalidAction | jq '.error.code'

echo -e "\n6. Testing GET on action URI (should return ActionInfo):"
curl -k -s -u admin:password \
  https://localhost:8443/redfish/v1/Systems/1/Actions/ComputerSystem.Reset | jq '.Id'

# Kill server
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo -e "\nActions testing complete!"
