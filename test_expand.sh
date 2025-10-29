#!/bin/bash

echo "Testing \$expand query parameter"
echo "==============================="

# Start server
go run cmd/server/main.go &
SERVER_PID=$!
sleep 2

echo "1. Basic system request:"
curl -k -s -u admin:password https://localhost:8443/redfish/v1/Systems/1 | jq '.Links.Chassis, .Links.ManagedBy'

echo -e "\n2. \$expand=Chassis:"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems/1?\$expand=Chassis" | jq '.Links.Chassis, .Links.ManagedBy'

echo -e "\n3. \$expand=ManagedBy:"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems/1?\$expand=ManagedBy" | jq '.Links.Chassis, .Links.ManagedBy'

echo -e "\n4. \$expand=Chassis,ManagedBy:"
curl -k -s -u admin:password "https://localhost:8443/redfish/v1/Systems/1?\$expand=Chassis,ManagedBy" | jq '.Links.Chassis, .Links.ManagedBy'

# Kill server
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo -e "\n\$expand testing complete!"
