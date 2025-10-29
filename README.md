# Redfish Server Implementation in Golang

A Redfish-compliant server implementation following DSP0266 (Redfish Protocol) and DSP0268 (Redfish Data Model) specifications.

## Project Status

**Current Stage:** Stage 6 - Query Parameters Support (In Progress)

### Completed Stages
- ✅ Stage 1: Project Setup and Architecture Planning (Completed)
  - Go module initialized
  - Directory structure created
  - Git repository initialized
  - Basic server skeleton implemented
  - Configuration management added
  - Build system configured

- ✅ Stage 2: Core HTTP Server and TLS Implementation (Completed & Tested)
  - HTTPS server with TLS 1.3 support implemented
  - Self-signed certificates generated for development
  - CORS middleware added for cross-origin requests
  - Logging middleware with request timing implemented
  - Redfish-required headers (OData-Version: 4.0) added
  - Health check, service root, and metadata endpoints implemented
  - Graceful shutdown handling with context cancellation
  - Comprehensive unit tests and integration tests added
  - **All tests passed** - see [stage2_report.md](stage2_report.md)

- ✅ Stage 3: Authentication and Authorization (Completed & Tested)
  - HTTP Basic Authentication implemented
  - Redfish Session Service with token management
  - Account Service with user enumeration
  - Role-based access control for protected endpoints
  - Session persistence and validation
  - TLS-secured authentication traffic
  - **All authentication tests passed** - see [stage3_report.md](stage3_report.md)

- ✅ Stage 4: Core Resource Models and Data Structures (Completed & Tested)
  - Common Redfish types (Status, Location, Actions, Links, etc.) implemented
  - ComputerSystem, Chassis, Manager, and AccountService models created
  - JSON marshaling/unmarshaling with proper OData annotations
  - Collection and individual resource support
  - **All model tests passed** - see [stage4_report.md](stage4_report.md)

- ✅ Stage 5: REST API Handlers Implementation (Completed & Tested)
  - Full CRUD operations (GET, POST, PATCH, PUT, DELETE) implemented
  - Proper HTTP status codes (200, 304, 401, 405, 404) returned
  - ETag support for optimistic concurrency control
  - Conditional GET with If-None-Match header support
  - Redfish-compliant error responses with extended information
  - **All handler tests passed** - see [stage5_report.md](stage5_report.md)

### Upcoming Stages
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
├── auth/            # Authentication and session management
├── config/          # Configuration management
├── middleware/      # HTTP middleware (CORS, logging, auth)
├── models/          # Redfish data models and structs
└── server/          # HTTP server and request handlers
pkg/                 # Public packages (future use)
api/                 # API specifications (future use)
docs/                # Documentation and specifications
```

## API Endpoints

The server implements the following Redfish API endpoints:

### Public Endpoints (No Authentication Required)
- `GET /health` - Health check
- `GET /redfish/v1/` - Service root
- `GET /redfish/v1/$metadata` - OData metadata
- `GET /redfish/v1/odata` - OData service document
- `POST /redfish/v1/SessionService/Sessions` - Session login
- `GET /redfish/v1/SessionService` - Session service info

### Protected Endpoints (Authentication Required)
- `GET /redfish/v1/Systems` - Computer systems collection
- `GET /redfish/v1/Systems/1` - Individual computer system
- `GET /redfish/v1/Chassis` - Chassis collection
- `GET /redfish/v1/Chassis/1` - Individual chassis
- `GET /redfish/v1/Managers` - Managers collection
- `GET /redfish/v1/Managers/1` - Individual manager
- `GET /redfish/v1/AccountService` - Account service
- `GET /redfish/v1/AccountService/Accounts` - Accounts collection
- `GET /redfish/v1/AccountService/Accounts/{username}` - Individual account

### Supported Features
- ✅ HTTP Basic Authentication
- ✅ Session-based authentication
- ✅ ETag support for caching
- ✅ Conditional GET requests
- ✅ Redfish-compliant error responses
- ✅ TLS 1.3 encryption

## Technology Choices

- **Language:** Go 1.21+
- **HTTP Server:** Standard `net/http` with custom routing
- **JSON Handling:** Standard `encoding/json`
- **Authentication:** Custom implementation with session tokens
- **Testing:** Standard `testing` package
- **Configuration:** Custom config management
- **Logging:** Standard `log` package with middleware

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