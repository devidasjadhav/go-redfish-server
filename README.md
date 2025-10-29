# Redfish Server Implementation in Golang

A Redfish-compliant server implementation following DSP0266 (Redfish Protocol) and DSP0268 (Redfish Data Model) specifications.

## Project Status

**Current Stage:** Stage 10 - OEM Extensions and Registries (Completed)

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

- ✅ Stage 6: Query Parameters Support (Completed & Tested)
  - OData query parameters ($top, $skip, $select, $expand, $filter) implemented
  - Pagination, filtering, and resource expansion support
  - Query parameter parsing with validation and error handling
  - Combined parameter processing with proper precedence
  - **All query parameter tests passed** - see [stage6_report.md](stage6_report.md)

- ✅ Stage 7: Actions Implementation (Completed & Tested)
  - ComputerSystem.Reset and Manager.Reset actions implemented
  - Action URI parsing and parameter validation added
  - ActionInfo resources for action metadata implemented
  - Proper HTTP status codes (204 No Content for POST, 200 for GET ActionInfo)
  - Parameter descriptions and allowable values support
  - **All action tests passed** - see [stage7_report.md](stage7_report.md)

- ✅ Stage 8: Eventing System (Completed & Tested)
  - EventService and EventSubscription data models implemented
  - EventService endpoints with configuration and capabilities
  - EventSubscriptions collection and individual subscription management
  - Server-Sent Events (SSE) infrastructure for real-time event delivery
  - Basic event filtering and routing framework
  - Comprehensive eventing tests with validation
  - **All eventing tests passed** - see [stage8_report.md](stage8_report.md)

- ✅ Stage 9: Asynchronous Operations (Tasks) (Completed & Tested)
  - Task and TaskService data models with full Redfish schema compliance
  - TaskService endpoints with configuration and task lifecycle management
  - Task collection and individual task CRUD operations
  - Asynchronous task execution with status monitoring and progress tracking
  - Integration with existing actions (ComputerSystem.Reset, Manager.Reset)
  - Comprehensive task testing with lifecycle validation
  - **All task tests passed** - see [stage9_report.md](stage9_report.md)

- ✅ Stage 10: OEM Extensions and Registries (Completed & Tested)
  - OEM extension framework with Contoso vendor-specific properties
  - Message Registry data models with complete message definitions
  - Registry collection and individual registry file endpoints
  - OEM properties integrated into ComputerSystem resources
  - OEM-specific custom actions and functionality
  - Comprehensive OEM and registry testing with validation
  - **All OEM and registry tests passed** - see [stage10_report.md](stage10_report.md)

### Upcoming Stages
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
- `POST /redfish/v1/Systems/1/Actions/ComputerSystem.Reset` - Reset computer system
- `GET /redfish/v1/Systems/1/Actions/ComputerSystem.Reset` - ComputerSystem.Reset action info
- `GET /redfish/v1/Chassis` - Chassis collection
- `GET /redfish/v1/Chassis/1` - Individual chassis
- `GET /redfish/v1/Managers` - Managers collection
- `GET /redfish/v1/Managers/1` - Individual manager
- `POST /redfish/v1/Managers/1/Actions/Manager.Reset` - Reset manager
- `GET /redfish/v1/Managers/1/Actions/Manager.Reset` - Manager.Reset action info
- `GET /redfish/v1/AccountService` - Account service
- `GET /redfish/v1/AccountService/Accounts` - Accounts collection
- `GET /redfish/v1/AccountService/Accounts/{username}` - Individual account
- `GET /redfish/v1/EventService` - Event service configuration
- `GET /redfish/v1/EventService/Subscriptions` - Event subscriptions collection
- `POST /redfish/v1/EventService/Subscriptions` - Create event subscription
- `GET /redfish/v1/EventService/Subscriptions/{id}` - Individual event subscription
- `DELETE /redfish/v1/EventService/Subscriptions/{id}` - Delete event subscription
- `GET /redfish/v1/EventService/SSE` - Server-Sent Events stream
- `GET /redfish/v1/TaskService` - Task service configuration
- `GET /redfish/v1/TaskService/Tasks` - Tasks collection
- `POST /redfish/v1/TaskService/Tasks` - Create new task
- `GET /redfish/v1/TaskService/Tasks/{id}` - Individual task status
- `DELETE /redfish/v1/TaskService/Tasks/{id}` - Delete completed task
- `GET /redfish/v1/Registries` - Message registries collection
- `GET /redfish/v1/Registries/{id}` - Individual message registry file
- `POST /redfish/v1/Oem/Contoso/CustomAction` - OEM custom action

### Supported Features
- ✅ HTTP Basic Authentication
- ✅ Session-based authentication
- ✅ ETag support for caching
- ✅ Conditional GET requests
- ✅ Redfish-compliant error responses
- ✅ TLS 1.3 encryption
- ✅ Redfish Actions (ComputerSystem.Reset, Manager.Reset)
- ✅ ActionInfo metadata for action parameters
- ✅ Redfish Eventing System with subscriptions and SSE
- ✅ Event filtering and routing framework
- ✅ Redfish Task Service for asynchronous operations
- ✅ Task lifecycle management with progress tracking
- ✅ OEM Extensions framework with vendor-specific properties
- ✅ Message Registry support with standard message definitions

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
make build
# or manually:
go build -o server cmd/server/main.go

# Run the server
make run
# or manually:
./server
```

## Redfish Protocol Validation

The server includes automated validation against the Redfish Protocol Validator to ensure compliance with DSP0266.

### Quick Validation

```bash
# Run full validation (build + start server + validate + cleanup)
make test-validation
# or
./validate.sh
```

This will:
1. Build the server
2. Start it with TLS on port 8443
3. Run the Redfish Protocol Validator
4. Generate HTML and TSV reports in `reports/`
5. Stop the server

### Manual Validation Steps

```bash
# 1. Build the server
make build

# 2. Start server in background
SERVER_ADDRESS=:8443 TLS_ENABLED=true ./server &

# 3. Run validator
python3 Redfish-Protocol-Validator/rf_protocol_validator.py \
  --user admin \
  --password password \
  --rhost https://127.0.0.1:8443 \
  --no-cert-check

# 4. Stop server
pkill -f server
```

### Current Compliance Status ✅

- **PASS:** 301 tests
- **WARN:** 0 tests
- **FAIL:** 0 tests
- **NOT_TESTED:** 42 tests

**🎉 FULL REDFISH PROTOCOL COMPLIANCE ACHIEVED!**

The server now passes all Redfish Protocol Validator tests. See `reports/` directory for detailed validation reports.

## Development

See [PLAN.md](PLAN.md) for detailed implementation plan and progress tracking.

## License

This project is licensed under the MIT License.