#!/bin/bash

# Test CRUD operations for Stage 5

echo "Testing Stage 5 CRUD Operations"
echo "================================"

# Start server in background
go run cmd/server/main.go &
SERVER_PID=$!
sleep 2

echo "1. Testing session creation"
SESSION_RESPONSE=$(curl -k -s -X POST -u admin:password https://localhost:8443/redfish/v1/SessionService/Sessions)
echo "Session response: $SESSION_RESPONSE"
TOKEN=$(echo $SESSION_RESPONSE | grep -o '"Id":"[^"]*"' | cut -d'"' -f4)
echo "Extracted token: '$TOKEN'"

if [ -z "$TOKEN" ]; then
    echo "Failed to get token, using Basic Auth instead"
    AUTH="-u admin:password"
else
    AUTH="-H X-Auth-Token:$TOKEN"
fi

echo -e "\n2. Testing GET /redfish/v1/Systems (should work)"
curl -k -s $AUTH https://localhost:8443/redfish/v1/Systems | head -5

echo -e "\n3. Testing POST /redfish/v1/Systems (should return 405)"
curl -k -s $AUTH -X POST https://localhost:8443/redfish/v1/Systems | head -3

echo -e "\n4. Testing GET /redfish/v1/Systems/1 (should work)"
RESPONSE=$(curl -k -s $AUTH https://localhost:8443/redfish/v1/Systems/1)
echo $RESPONSE | head -5

echo -e "\n5. Testing PATCH /redfish/v1/Systems/1 (should return 405)"
curl -k -s $AUTH -X PATCH https://localhost:8443/redfish/v1/Systems/1 | head -3

echo -e "\n6. Testing conditional GET with ETag"
ETAG=$(curl -k -s $AUTH https://localhost:8443/redfish/v1/Systems/1 -D - 2>/dev/null | grep ETag | cut -d'"' -f2)
echo "ETag: '$ETAG'"
if [ -n "$ETAG" ]; then
    STATUS=$(curl -k -s $AUTH -H "If-None-Match: \"$ETAG\"" https://localhost:8443/redfish/v1/Systems/1 -w "%{http_code}" -o /dev/null)
    echo "Conditional GET status: $STATUS"
fi

# Kill server
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo -e "\nStage 5 CRUD testing complete!"
