# Redfish Server Implementation in Golang

A Redfish-compliant server implementation following DSP0266 (Redfish Protocol) and DSP0268 (Redfish Data Model) specifications.

## Project Status

**Current Stage:** Stage 1 - Project Setup and Architecture Planning (Completed)

### Completed Stages
- ✅ Stage 1: Project Setup and Architecture Planning (Completed)
  - Go module initialized
  - Directory structure created
  - Git repository initialized
  - Basic server skeleton implemented
  - Configuration management added
  - Build system configured

### Upcoming Stages
- Stage 2: Core HTTP Server and TLS Implementation
- Stage 3: Authentication and Authorization
- Stage 4: Core Resource Models and Data Structures
- Stage 5: REST API Handlers Implementation
- Stage 6: Query Parameters Support
- Stage 7: Actions Implementation
- Stage 8: Eventing System
- Stage 9: Asynchronous Operations (Tasks)
- Stage 10: OEM Extensions and Registries
- Stage 11: Testing and Conformance Validation

## Architecture

This implementation follows Clean Architecture principles with the following structure:

```
cmd/server/          # Application entry points
internal/            # Private application code
├── handlers/        # HTTP request handlers
├── models/          # Data models and structs
├── services/        # Business logic services
├── middleware/      # HTTP middleware
└── config/          # Configuration management
pkg/                 # Public packages
├── utils/           # Utility functions
└── storage/         # Storage interfaces
api/                 # API specifications
docs/                # Documentation
```

## Technology Choices

- **Language:** Go 1.21+
- **HTTP Server:** Standard `net/http` with `gorilla/mux` for routing
- **JSON Handling:** Standard `encoding/json`
- **Authentication:** Custom implementation with JWT support
- **Testing:** Standard `testing` package with `testify`
- **Configuration:** `viper` for flexible config management
- **Logging:** `logrus` for structured logging

## Building and Running

```bash
# Install dependencies
go mod tidy

# Build the server
go build -o bin/server cmd/server/main.go

# Run the server
./bin/server
```

## Development

See [PLAN.md](PLAN.md) for detailed implementation plan and progress tracking.

## License

This project is licensed under the MIT License.