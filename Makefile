.PHONY: build run clean test fmt vet mod-tidy

# Build the server
build:
	go build -o bin/server cmd/server/main.go

# Run the server
run: build
	./bin/server

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

# All checks
check: fmt vet test