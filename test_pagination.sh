#!/bin/bash

echo "Testing pagination query parameters"
echo "==================================="

# Start server
go run cmd/server/main.go &
SERVER_PID=$!
sleep 2

echo "1. Testing basic collection (should return 1 member)"
curl -k -s -u admin:password https://localhost:8443/redfish/v1/Systems | jq '.Members | length'

echo -e "\n2. Testing \$top=1 (should return 1 member)"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems?\$top=1" | jq '.Members | length'

echo -e "\n3. Testing \$skip=1 (should return 0 members)"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems?\$skip=1" | jq '.Members | length'

echo -e "\n4. Testing invalid \$top parameter (should return error)"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems?\$top=invalid" | jq '.error.code'

echo -e "\n5. Testing invalid \$skip parameter (should return error)"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems?\$skip=invalid" | jq '.error.code'

# Kill server
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo -e "\nPagination testing complete!"
