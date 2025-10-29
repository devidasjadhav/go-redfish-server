#!/bin/bash

echo "Testing \$filter query parameter"
echo "==============================="

# Start server
go run cmd/server/main.go &
SERVER_PID=$!
sleep 2

echo "1. Basic collection (should return 1 member):"
curl -k -s -u admin:password https://localhost:8443/redfish/v1/Systems | jq '.Members | length'

echo -e "\n2. \$filter=PowerState eq 'On' (should return 1 member):"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems?\$filter=PowerState%20eq%20'On'" | jq '.Members | length'

echo -e "\n3. \$filter=PowerState eq 'Off' (should return 0 members):"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems?\$filter=PowerState%20eq%20'Off'" | jq '.Members | length'

# Kill server
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo -e "\n\$filter testing complete!"
