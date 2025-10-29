.PHONY: build run clean test fmt vet mod-tidy

# Build the server
build:
	go build -o server cmd/server/main.go

# Run the server
run: build
	./server

# Clean build artifacts
clean:
	rm -rf bin/

# Run tests
test:
	go test ./...

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Tidy modules
mod-tidy:
	go mod tidy

# Development setup
dev-setup: mod-tidy fmt vet

# Run Redfish Protocol Validator
.PHONY: test-validation
test-validation: build
	@echo "Starting server..."
	@SERVER_ADDRESS=:8443 TLS_ENABLED=true ./server &
	@SERVER_PID=$$!
	@sleep 3
	@echo "Running Redfish Protocol Validator..."
	@-python3 Redfish-Protocol-Validator/rf_protocol_validator.py --user admin --password password --rhost https://127.0.0.1:8443 --no-cert-check; \
	EXIT_CODE=$$?; \
	kill $$SERVER_PID 2>/dev/null || pkill -f "./server" 2>/dev/null || true; \
	exit $$EXIT_CODE
	@echo "Validation complete. Check reports/ for results."

# All checks
check: fmt vet test