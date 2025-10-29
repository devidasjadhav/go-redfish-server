#!/bin/bash

echo "Testing \$select query parameter"
echo "==============================="

# Start server
go run cmd/server/main.go &
SERVER_PID=$!
sleep 2

echo "1. Basic system request (all properties):"
curl -k -s -u admin:password https://localhost:8443/redfish/v1/Systems/1 | jq 'keys | length'

echo -e "\n2. \$select single property:"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems/1?\$select=Id" | jq 'keys | length'

echo -e "\n3. \$select multiple properties:"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems/1?\$select=Id,Name,Status" | jq 'keys | length'

echo -e "\n4. \$select with invalid property (should still work):"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems/1?\$select=Id,InvalidProperty" | jq 'keys | length'

# Kill server
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo -e "\n\$select testing complete!"
