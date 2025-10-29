#!/bin/bash

echo "Testing ETag functionality"
echo "=========================="

# Start server
go run cmd/server/main.go &
SERVER_PID=$!
sleep 2

echo "1. Testing ETag header presence"
RESPONSE=$(curl -k -s -u admin:password -D - https://localhost:8443/redfish/v1/Systems/1 2>/dev/null)
echo "Response headers:"
echo "$RESPONSE" | head -10

ETAG=$(echo "$RESPONSE" | grep -i etag | head -1 | sed 's/.*ETag: "\([^"]*\)".*/\1/')
echo "Extracted ETag: '$ETAG'"

if [ -n "$ETAG" ]; then
    echo -e "\n2. Testing conditional GET with matching ETag (should return 304)"
    STATUS=$(curl -k -s -u admin:password -H "If-None-Match: \"$ETAG\"" -w "%{http_code}" -o /dev/null https://localhost:8443/redfish/v1/Systems/1)
    echo "Status code: $STATUS"
    
    echo -e "\n3. Testing conditional GET with different ETag (should return 200)"
    STATUS=$(curl -k -s -u admin:password -H "If-None-Match: \"different-etag\"" -w "%{http_code}" -o /dev/null https://localhost:8443/redfish/v1/Systems/1)
    echo "Status code: $STATUS"
else
    echo "ETag not found in response"
fi

# Kill server
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo -e "\nETag testing complete!"
