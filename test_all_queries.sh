#!/bin/bash

echo "Comprehensive Query Parameters Test"
echo "==================================="

# Start server
go run cmd/server/main.go &
SERVER_PID=$!
sleep 2

echo "1. Basic requests (no query params):"
echo "Systems collection:"
curl -k -s -u admin:password https://localhost:8443/redfish/v1/Systems | jq '.Members | length, ."Members@odata.count"'
echo "Individual system:"
curl -k -s -u admin:password https://localhost:8443/redfish/v1/Systems/1 | jq 'keys | length'

echo -e "\n2. Pagination (\$top, \$skip):"
echo "\$top=1:"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems?\$top=1" | jq '.Members | length'
echo "\$skip=1:"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems?\$skip=1" | jq '.Members | length'

echo -e "\n3. Filtering (\$filter):"
echo "\$filter=PowerState eq 'On':"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems?\$filter=PowerState%20eq%20'On'" | jq '.Members | length'
echo "\$filter=PowerState eq 'Off':"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems?\$filter=PowerState%20eq%20'Off'" | jq '.Members | length'

echo -e "\n4. Selection (\$select):"
echo "\$select=Id,Name:"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems/1?\$select=Id,Name" | jq 'keys | length'

echo -e "\n5. Expansion (\$expand):"
echo "\$expand=Chassis:"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems/1?\$expand=Chassis" | jq '.Links.Chassis'

echo -e "\n6. Combined query parameters:"
echo "\$top=1&\$filter=PowerState eq 'On':"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems?\$top=1&\$filter=PowerState%20eq%20'On'" | jq '.Members | length'

echo -e "\n7. Error handling:"
echo "Invalid \$top:"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems?\$top=invalid" | jq '.error.code'
echo "Invalid \$skip:"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems?\$skip=invalid" | jq '.error.code'

# Kill server
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo -e "\nQuery parameters testing complete!"
