#!/bin/bash

# Test script for Redfish server
set -e

echo "🧪 Testing Redfish Server Implementation"
echo "========================================"

# Build the server
echo "📦 Building server..."
make build

# Start server in background
echo "🚀 Starting server..."
./bin/server &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Test health endpoint
echo "🏥 Testing health endpoint..."
HEALTH_RESPONSE=$(curl -k -s https://localhost:8443/health)
if [[ "$HEALTH_RESPONSE" == *"redfish-server"* ]]; then
    echo "✅ Health endpoint working"
else
    echo "❌ Health endpoint failed"
    exit 1
fi

# Test Redfish service root
echo "🔍 Testing Redfish service root..."
ROOT_RESPONSE=$(curl -k -s https://localhost:8443/redfish/v1/)
if [[ "$ROOT_RESPONSE" == *"@odata.type"* ]]; then
    echo "✅ Service root endpoint working"
else
    echo "❌ Service root endpoint failed"
    exit 1
fi

# Test OData metadata
echo "📄 Testing OData metadata..."
METADATA_RESPONSE=$(curl -k -s https://localhost:8443/redfish/v1/\$metadata)
if [[ "$METADATA_RESPONSE" == *"edmx:Edmx"* ]]; then
    echo "✅ OData metadata working"
else
    echo "❌ OData metadata failed"
    exit 1
fi

# Test CORS
echo "🌐 Testing CORS..."
CORS_HEADERS=$(curl -k -s -I https://localhost:8443/redfish/v1/ | grep -i access-control | wc -l)
if [[ "$CORS_HEADERS" -gt 0 ]]; then
    echo "✅ CORS headers present"
else
    echo "❌ CORS headers missing"
    exit 1
fi

# Test TLS
echo "🔒 Testing TLS..."
TLS_INFO=$(curl -k -v https://localhost:8443/health 2>&1 | grep "TLSv1.3" | wc -l)
if [[ "$TLS_INFO" -gt 0 ]]; then
    echo "✅ TLS 1.3 working"
else
    echo "❌ TLS not working"
    exit 1
fi

# Run unit tests
echo "🧪 Running unit tests..."
go test ./...
echo "✅ Unit tests passed"

# Shutdown server
echo "🛑 Shutting down server..."
kill $SERVER_PID
sleep 2

echo ""
echo "🎉 All tests passed! Server implementation is working correctly."
echo "📊 Test Summary:"
echo "   • HTTPS/TLS server ✅"
echo "   • Health endpoint ✅"
echo "   • Redfish service root ✅"
echo "   • OData metadata ✅"
echo "   • CORS support ✅"
echo "   • Unit tests ✅"