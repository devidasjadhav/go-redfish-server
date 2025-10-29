#!/bin/bash

echo "Debug pagination"
echo "================"

# Start server
go run cmd/server/main.go &
SERVER_PID=$!
sleep 2

echo "Basic collection:"
curl -k -s -u admin:password https://localhost:8443/redfish/v1/Systems | jq '.'

echo -e "\nWith \$skip=1:"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems?\$skip=1" | jq '.'

# Kill server
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null
