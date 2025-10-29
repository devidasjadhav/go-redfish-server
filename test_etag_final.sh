#!/bin/bash

echo "Testing ETag functionality (final test)"
echo "======================================="

# Start server
go run cmd/server/main.go &
SERVER_PID=$!
sleep 2

echo "1. Testing ETag header presence"
RESPONSE=$(curl -k -s -u admin:password -D - https://localhost:8443/redfish/v1/Systems/1 2>/dev/null)
echo "ETag header:"
echo "$RESPONSE" | grep -i etag

# Extract ETag more carefully
ETAG_LINE=$(echo "$RESPONSE" | grep -i etag)
ETAG=$(echo "$ETAG_LINE" | sed 's/.*etag: *"\([^"]*\)".*/\1/i')
echo "Extracted ETag: '$ETAG'"

if [ -n "$ETAG" ] && [ "$ETAG" != "$ETAG_LINE" ]; then
    echo -e "\n2. Testing conditional GET with matching ETag (should return 304)"
    STATUS=$(curl -k -s -u admin:password -H "If-None-Match: \"$ETAG\"" -w "%{http_code}" -o /dev/null https://localhost:8443/redfish/v1/Systems/1)
    echo "Status code: $STATUS (expected: 304)"
    
    echo -e "\n3. Testing conditional GET with different ETag (should return 200)"
    STATUS=$(curl -k -s -u admin:password -H "If-None-Match: \"different\"" -w "%{http_code}" -o /dev/null https://localhost:8443/redfish/v1/Systems/1)
    echo "Status code: $STATUS (expected: 200)"
else
    echo "ETag extraction failed"
fi

# Kill server
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo -e "\nETag testing complete!"
