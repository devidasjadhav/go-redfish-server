#!/bin/bash

echo "Final pagination testing"
echo "========================"

# Start server
go run cmd/server/main.go &
SERVER_PID=$!
sleep 2

echo "1. Basic collection (should return 1 member):"
curl -k -s -u admin:password https://localhost:8443/redfish/v1/Systems | jq '.Members | length, ."Members@odata.count"'

echo -e "\n2. \$top=1 (should return 1 member):"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems?\$top=1" | jq '.Members | length, ."Members@odata.count"'

echo -e "\n3. \$skip=1 (should return 0 members):"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems?\$skip=1" | jq '.Members | length, ."Members@odata.count"'

echo -e "\n4. \$top=0 (should return 0 members):"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems?\$top=0" | jq '.Members | length, ."Members@odata.count"'

echo -e "\n5. Invalid \$top parameter (should return error):"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems?\$top=invalid" | jq '.error.code // empty'

echo -e "\n6. Invalid \$skip parameter (should return error):"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems?\$skip=invalid" | jq '.error.code // empty'

# Kill server
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo -e "\nPagination testing complete!"
