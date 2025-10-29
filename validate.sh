#!/bin/bash
set -e

echo "Building server..."
go build -o server cmd/server/main.go

echo "Starting server..."
SERVER_ADDRESS=:8443 TLS_ENABLED=true ./server &
SERVER_PID=$!

sleep 3

echo "Running Redfish Protocol Validator..."
python3 Redfish-Protocol-Validator/rf_protocol_validator.py --user admin --password password --rhost https://127.0.0.1:8443 --no-cert-check
EXIT_CODE=$?

echo "Stopping server..."
kill $SERVER_PID 2>/dev/null || pkill -f "./server" 2>/dev/null || true

if [ $EXIT_CODE -eq 0 ]; then
    echo "Validation complete. Check reports/ for results."
else
    echo "Validation failed with exit code $EXIT_CODE. Check reports/ for details."
fi

exit $EXIT_CODE