package server

import (
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/user/redfish-server/internal/auth"
	"github.com/user/redfish-server/internal/config"
	"github.com/user/redfish-server/internal/middleware"
	"github.com/user/redfish-server/internal/models"
)

// Server represents the Redfish HTTP server
type Server struct {
	httpServer *http.Server
	config     *config.Config
}

// New creates a new Redfish server instance
func New(cfg *config.Config) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	mux := http.NewServeMux()
	setupRoutes(mux)

	// Apply middleware
	handler := middleware.CORSMiddleware(mux)
	handler = middleware.AuthMiddleware(handler)
	handler = middleware.LoggingMiddleware(handler)

	httpServer := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      handler,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	if cfg.TLS.Enabled {
		cert, err := tls.LoadX509KeyPair(cfg.TLS.CertFile, cfg.TLS.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS certificates: %w", err)
		}

		httpServer.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}
	}

	return &Server{
		httpServer: httpServer,
		config:     cfg,
	}, nil
}

// Start starts the server
func (s *Server) Start() error {
	fmt.Printf("Starting Redfish server on %s (TLS: %t)\n", s.config.Server.Address, s.config.TLS.Enabled)

	if s.config.TLS.Enabled {
		fmt.Printf("TLS certificates: %s, %s\n", s.config.TLS.CertFile, s.config.TLS.KeyFile)
		return s.httpServer.ListenAndServeTLS("", "")
	}

	fmt.Println("WARNING: TLS is disabled. Redfish requires TLS in production!")
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(ctx)
}

// setupRoutes configures the HTTP routes
func setupRoutes(mux *http.ServeMux) {
	// Health check endpoint
	mux.HandleFunc("/health", healthHandler)

	// Redfish endpoints - order matters! More specific routes first
	mux.HandleFunc("/redfish/v1/$metadata", metadataHandler)
	mux.HandleFunc("/redfish/v1/odata", odataHandler)

	// Session service endpoints
	mux.HandleFunc("/redfish/v1/SessionService/Sessions", sessionsHandler)
	mux.HandleFunc("/redfish/v1/SessionService", sessionServiceHandler)

	// Account service endpoints
	mux.HandleFunc("/redfish/v1/AccountService/Accounts/", accountHandler)
	mux.HandleFunc("/redfish/v1/AccountService/Accounts", accountsHandler)
	mux.HandleFunc("/redfish/v1/AccountService", accountServiceHandler)

	// Computer system endpoints
	mux.HandleFunc("/redfish/v1/Systems/", systemHandler)
	mux.HandleFunc("/redfish/v1/Systems", systemsHandler)

	// Chassis endpoints
	mux.HandleFunc("/redfish/v1/Chassis/", chassisItemHandler)
	mux.HandleFunc("/redfish/v1/Chassis", chassisHandler)

	// Manager endpoints
	mux.HandleFunc("/redfish/v1/Managers/", managerHandler)
	mux.HandleFunc("/redfish/v1/Managers", managersHandler)

	// Redfish root endpoint - must be last
	mux.HandleFunc("/redfish/v1/", serviceRootHandler)
}

// healthHandler handles health check requests
func healthHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)

	switch r.Method {
	case "GET":
		handleGetHealth(w, r)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetHealth returns health check information
func handleGetHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := `{"status": "ok", "service": "redfish-server"}`
	etag := generateETag(response)
	w.Header().Set("ETag", etag)

	w.Write([]byte(response))
}

// serviceRootHandler handles the Redfish service root
func serviceRootHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)

	switch r.Method {
	case "GET":
		handleGetServiceRoot(w, r)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetServiceRoot returns the service root
func handleGetServiceRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	serviceRoot := models.NewServiceRoot()
	etag := generateETag(serviceRoot)
	w.Header().Set("ETag", etag)

	// Check conditional GET
	if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
		normalizedETag := normalizeETag(etag)
		normalizedIfNoneMatch := normalizeETag(ifNoneMatch)
		if normalizedIfNoneMatch == normalizedETag || ifNoneMatch == "*" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	json.NewEncoder(w).Encode(serviceRoot)
}

// handleGetAccountService returns the account service
func handleGetAccountService(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accountService := models.NewAccountService()
	etag := generateETag(accountService)
	w.Header().Set("ETag", etag)

	// Check conditional GET
	if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
		normalizedETag := normalizeETag(etag)
		normalizedIfNoneMatch := normalizeETag(ifNoneMatch)
		if normalizedIfNoneMatch == normalizedETag || ifNoneMatch == "*" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	json.NewEncoder(w).Encode(accountService)
}

// metadataHandler serves the OData metadata document
func metadataHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)

	switch r.Method {
	case "GET":
		handleGetMetadata(w, r)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetMetadata returns the OData metadata document
func handleGetMetadata(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")

	// Basic metadata document (simplified)
	metadata := `<?xml version="1.0" encoding="utf-8"?>
<edmx:Edmx Version="4.0" xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx">
  <edmx:DataServices>
    <Schema Namespace="Service" xmlns="http://docs.oasis-open.org/odata/ns/edm">
      <EntityType Name="ServiceRoot">
        <Key>
          <PropertyRef Name="Id" />
        </Key>
        <Property Name="Id" Type="Edm.String" Nullable="false" />
        <Property Name="Name" Type="Edm.String" Nullable="false" />
        <Property Name="RedfishVersion" Type="Edm.String" Nullable="false" />
      </EntityType>
    </Schema>
  </edmx:DataServices>
</edmx:Edmx>`

	etag := generateETag(metadata)
	w.Header().Set("ETag", etag)

	// Check conditional GET
	if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
		normalizedETag := normalizeETag(etag)
		normalizedIfNoneMatch := normalizeETag(ifNoneMatch)
		if normalizedIfNoneMatch == normalizedETag || ifNoneMatch == "*" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	w.Write([]byte(metadata))
}

// handleGetOdata returns the OData service document
func handleGetOdata(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := `{
		"@odata.context": "/redfish/v1/$metadata",
		"value": [
			{
				"name": "ServiceRoot",
				"url": "/redfish/v1/"
			}
		]
	}`

	etag := generateETag(response)
	w.Header().Set("ETag", etag)

	// Check conditional GET
	if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
		normalizedETag := normalizeETag(etag)
		normalizedIfNoneMatch := normalizeETag(ifNoneMatch)
		if normalizedIfNoneMatch == normalizedETag || ifNoneMatch == "*" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	w.Write([]byte(response))
}

// odataHandler serves the OData service document
func odataHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)

	switch r.Method {
	case "GET":
		handleGetOdata(w, r)
	default:
		methodNotAllowed(w, r)
	}
}

// sessionServiceHandler handles the SessionService resource
func sessionServiceHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)

	switch r.Method {
	case "GET":
		handleGetSessionService(w, r)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetSessionService returns the session service
func handleGetSessionService(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := `{
		"@odata.context": "/redfish/v1/$metadata#SessionService.SessionService",
		"@odata.id": "/redfish/v1/SessionService",
		"@odata.type": "#SessionService.v1_1_8.SessionService",
		"Id": "SessionService",
		"Name": "Session Service",
		"Status": {
			"State": "Enabled",
			"Health": "OK"
		},
		"ServiceEnabled": true,
		"SessionTimeout": 3600,
		"Sessions": {
			"@odata.id": "/redfish/v1/SessionService/Sessions"
		}
	}`

	etag := generateETag(response)
	w.Header().Set("ETag", etag)

	// Check conditional GET
	if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
		normalizedETag := normalizeETag(etag)
		normalizedIfNoneMatch := normalizeETag(ifNoneMatch)
		if normalizedIfNoneMatch == normalizedETag || ifNoneMatch == "*" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	w.Write([]byte(response))
}

// sessionsHandler handles session collection and creation
func sessionsHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)

	switch r.Method {
	case "GET":
		handleGetSessions(w, r)
	case "POST":
		handleCreateSession(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetSessions returns the sessions collection
func handleGetSessions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := `{
		"@odata.context": "/redfish/v1/$metadata#SessionCollection.SessionCollection",
		"@odata.id": "/redfish/v1/SessionService/Sessions",
		"@odata.type": "#SessionCollection.SessionCollection",
		"Name": "Sessions Collection",
		"Members": [],
		"Members@odata.count": 0
	}`

	etag := generateETag(response)
	w.Header().Set("ETag", etag)

	// Check conditional GET
	if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
		normalizedETag := normalizeETag(etag)
		normalizedIfNoneMatch := normalizeETag(ifNoneMatch)
		if normalizedIfNoneMatch == normalizedETag || ifNoneMatch == "*" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	w.Write([]byte(response))
}

// handleCreateSession creates a new session (login)
func handleCreateSession(w http.ResponseWriter, r *http.Request) {
	// For simplicity, use Basic Auth for login
	// TODO: Support JSON body with UserName/Password
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, `{"error": {"code": "Base.1.0.InsufficientPrivilege", "message": "Basic authentication required"}}`, http.StatusUnauthorized)
		return
	}

	// Validate credentials
	authService := auth.GetAuthService()
	if !authService.ValidateBasicAuth(username, password) {
		http.Error(w, `{"error": {"code": "Base.1.0.InsufficientPrivilege", "message": "Invalid credentials"}}`, http.StatusUnauthorized)
		return
	}

	// Create session
	token, err := authService.CreateSession(username)
	if err != nil {
		http.Error(w, `{"error": {"code": "Base.1.0.InternalError", "message": "Failed to create session"}}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Auth-Token", token)
	w.WriteHeader(http.StatusCreated)

	response := fmt.Sprintf(`{
		"@odata.context": "/redfish/v1/$metadata#Session.Session",
		"@odata.id": "/redfish/v1/SessionService/Sessions/%s",
		"@odata.type": "#Session.v1_1_6.Session",
		"Id": "%s",
		"Name": "User Session",
		"UserName": "%s"
	}`, token, token, username)

	w.Write([]byte(response))
}

// accountServiceHandler handles the AccountService resource
func accountServiceHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)

	switch r.Method {
	case "GET":
		handleGetAccountService(w, r)
	case "PATCH":
		handleUpdateAccountService(w, r)
	default:
		methodNotAllowed(w, r)
	}
}

// handleUpdateAccountService updates the account service (PATCH)
func handleUpdateAccountService(w http.ResponseWriter, r *http.Request) {
	sendRedfishError(w, "MethodNotAllowed", "AccountService updates not implemented", http.StatusMethodNotAllowed)
}

// accountsHandler handles the accounts collection
func accountsHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)

	switch r.Method {
	case "GET":
		handleGetAccounts(w, r)
	case "POST":
		handleCreateAccount(w, r)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetAccounts returns the accounts collection
func handleGetAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accounts := models.NewManagerAccountCollection()
	etag := generateETag(accounts)
	w.Header().Set("ETag", etag)

	// Check conditional GET
	if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
		normalizedETag := normalizeETag(etag)
		normalizedIfNoneMatch := normalizeETag(ifNoneMatch)
		if normalizedIfNoneMatch == normalizedETag || ifNoneMatch == "*" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	json.NewEncoder(w).Encode(accounts)
}

// handleCreateAccount creates a new user account
func handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	// For now, account creation is not implemented
	sendRedfishError(w, "MethodNotAllowed", "Account creation not implemented", http.StatusMethodNotAllowed)
}

// accountHandler handles individual account resources
func accountHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)

	// Extract username from URL path
	path := r.URL.Path
	username := path[len("/redfish/v1/AccountService/Accounts/"):]

	switch r.Method {
	case "GET":
		handleGetAccount(w, r, username)
	case "PATCH":
		handleUpdateAccount(w, r, username)
	case "PUT":
		handleReplaceAccount(w, r, username)
	case "DELETE":
		handleDeleteAccount(w, r, username)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetAccount returns a specific account
func handleGetAccount(w http.ResponseWriter, r *http.Request, username string) {
	w.Header().Set("Content-Type", "application/json")

	// For demo purposes, only support admin and operator accounts
	var account *models.ManagerAccount
	switch username {
	case "admin":
		account = models.NewManagerAccount("admin", "Administrator", true)
	case "operator":
		account = models.NewManagerAccount("operator", "Operator", true)
	default:
		sendRedfishError(w, "ResourceNotFound", "Account not found", http.StatusNotFound)
		return
	}

	etag := generateETag(account)
	w.Header().Set("ETag", etag)

	// Check conditional GET
	if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
		normalizedETag := normalizeETag(etag)
		normalizedIfNoneMatch := normalizeETag(ifNoneMatch)
		if normalizedIfNoneMatch == normalizedETag || ifNoneMatch == "*" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	json.NewEncoder(w).Encode(account)
}

// handleUpdateAccount updates an account (PATCH)
func handleUpdateAccount(w http.ResponseWriter, r *http.Request, username string) {
	sendRedfishError(w, "MethodNotAllowed", "Account updates not implemented", http.StatusMethodNotAllowed)
}

// handleReplaceAccount replaces an account (PUT)
func handleReplaceAccount(w http.ResponseWriter, r *http.Request, username string) {
	sendRedfishError(w, "MethodNotAllowed", "Account replacement not implemented", http.StatusMethodNotAllowed)
}

// handleDeleteAccount deletes an account
func handleDeleteAccount(w http.ResponseWriter, r *http.Request, username string) {
	sendRedfishError(w, "MethodNotAllowed", "Account deletion not implemented", http.StatusMethodNotAllowed)
}

// systemsHandler handles the computer systems collection
func systemsHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)

	switch r.Method {
	case "GET":
		handleGetSystems(w, r)
	case "POST":
		handleCreateSystem(w, r)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetSystems returns the computer systems collection
func handleGetSystems(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	systems := models.NewComputerSystemCollection()

	// Parse query parameters
	queryParams, err := parseQueryParameters(r.URL.Query())
	if err != nil {
		sendRedfishError(w, "QueryParameterError", err.Error(), http.StatusBadRequest)
		return
	}

	// Apply query parameters
	systems = applyQueryParametersToSystems(systems, queryParams)

	etag := generateETag(systems)
	w.Header().Set("ETag", etag)

	// Check conditional GET
	if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
		normalizedETag := normalizeETag(etag)
		normalizedIfNoneMatch := normalizeETag(ifNoneMatch)
		if normalizedIfNoneMatch == normalizedETag || ifNoneMatch == "*" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	json.NewEncoder(w).Encode(systems)
}

// handleCreateSystem creates a new computer system (not typically allowed in Redfish)
func handleCreateSystem(w http.ResponseWriter, r *http.Request) {
	// Computer systems are typically not created via POST in Redfish
	// This would be a BMC implementation detail
	sendRedfishError(w, "MethodNotAllowed", "ComputerSystem creation not supported", http.StatusMethodNotAllowed)
}

// systemHandler handles individual computer system resources and actions
func systemHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)

	path := r.URL.Path

	// Check if this is an action request
	if strings.Contains(path, "/Actions/") {
		handleSystemAction(w, r, path)
		return
	}

	// Extract system ID from URL path
	id := path[len("/redfish/v1/Systems/"):]

	switch r.Method {
	case "GET":
		handleGetSystem(w, r, id)
	case "PATCH":
		handleUpdateSystem(w, r, id)
	case "PUT":
		handleReplaceSystem(w, r, id)
	case "DELETE":
		handleDeleteSystem(w, r, id)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetSystem returns a specific computer system
func handleGetSystem(w http.ResponseWriter, r *http.Request, id string) {
	w.Header().Set("Content-Type", "application/json")

	system := models.NewComputerSystem(id)

	// Parse query parameters
	queryParams, err := parseQueryParameters(r.URL.Query())
	if err != nil {
		sendRedfishError(w, "QueryParameterError", err.Error(), http.StatusBadRequest)
		return
	}

	// Apply $select if specified
	if len(queryParams.Select) > 0 {
		system = applySelectToSystem(system, queryParams.Select)
	}

	// Apply $expand if specified
	if len(queryParams.Expand) > 0 {
		system = applyExpandToSystem(system, queryParams.Expand)
	}

	etag := generateETag(system)
	w.Header().Set("ETag", etag)

	// Check conditional GET
	if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
		normalizedETag := normalizeETag(etag)
		normalizedIfNoneMatch := normalizeETag(ifNoneMatch)
		if normalizedIfNoneMatch == normalizedETag || ifNoneMatch == "*" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	json.NewEncoder(w).Encode(system)
}

// handleUpdateSystem updates a computer system (PATCH)
func handleUpdateSystem(w http.ResponseWriter, r *http.Request, id string) {
	// For now, systems are read-only in this implementation
	sendRedfishError(w, "MethodNotAllowed", "ComputerSystem updates not supported", http.StatusMethodNotAllowed)
}

// handleReplaceSystem replaces a computer system (PUT)
func handleReplaceSystem(w http.ResponseWriter, r *http.Request, id string) {
	// For now, systems are read-only in this implementation
	sendRedfishError(w, "MethodNotAllowed", "ComputerSystem replacement not supported", http.StatusMethodNotAllowed)
}

// handleDeleteSystem deletes a computer system
func handleDeleteSystem(w http.ResponseWriter, r *http.Request, id string) {
	// Computer systems are typically not deleted in Redfish
	sendRedfishError(w, "MethodNotAllowed", "ComputerSystem deletion not supported", http.StatusMethodNotAllowed)
}

// handleSystemAction handles ComputerSystem actions
func handleSystemAction(w http.ResponseWriter, r *http.Request, path string) {
	// Extract action from path: /redfish/v1/Systems/{id}/Actions/{ActionName}
	parts := strings.Split(path, "/")
	if len(parts) < 7 || parts[5] != "Actions" {
		sendRedfishError(w, "InvalidAction", "Invalid action URI format", http.StatusBadRequest)
		return
	}

	actionName := parts[6]
	systemId := parts[4]

	switch r.Method {
	case "GET":
		switch actionName {
		case "ComputerSystem.Reset":
			handleComputerSystemResetActionInfo(w, r, systemId)
		default:
			sendRedfishError(w, "ActionNotSupported", fmt.Sprintf("Action %s not supported for ComputerSystem", actionName), http.StatusBadRequest)
		}
	case "POST":
		switch actionName {
		case "ComputerSystem.Reset":
			handleComputerSystemReset(w, r, systemId)
		default:
			sendRedfishError(w, "ActionNotSupported", fmt.Sprintf("Action %s not supported for ComputerSystem", actionName), http.StatusBadRequest)
		}
	default:
		methodNotAllowed(w, r)
	}
}

// handleComputerSystemResetActionInfo returns ActionInfo for ComputerSystem.Reset
func handleComputerSystemResetActionInfo(w http.ResponseWriter, r *http.Request, systemId string) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"@odata.context": "/redfish/v1/$metadata#ActionInfo.ActionInfo",
		"@odata.id":      fmt.Sprintf("/redfish/v1/Systems/%s/Actions/ComputerSystem.Reset", systemId),
		"@odata.type":    "#ActionInfo.v1_1_2.ActionInfo",
		"Id":             "ComputerSystem.Reset",
		"Name":           "Computer System Reset",
		"Parameters": []map[string]interface{}{
			{
				"Name":            "ResetType",
				"Required":        false,
				"DataType":        "String",
				"AllowableValues": []string{"On", "ForceOff", "ForceRestart", "Nmi", "PushPowerButton", "GracefulRestart", "GracefulShutdown", "ForceOn"},
			},
		},
	}

	etag := generateETag(response)
	w.Header().Set("ETag", etag)

	// Check conditional GET
	if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
		normalizedETag := normalizeETag(etag)
		normalizedIfNoneMatch := normalizeETag(ifNoneMatch)
		if normalizedIfNoneMatch == normalizedETag || ifNoneMatch == "*" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	json.NewEncoder(w).Encode(response)
}

// handleComputerSystemReset handles the ComputerSystem.Reset action
func handleComputerSystemReset(w http.ResponseWriter, r *http.Request, systemId string) {
	// Parse request body for ResetType parameter
	var requestBody struct {
		ResetType string `json:"ResetType"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil && err.Error() != "EOF" {
		sendRedfishError(w, "MalformedJSON", "Invalid JSON in request body", http.StatusBadRequest)
		return
	}

	// Validate ResetType parameter
	validResetTypes := map[string]bool{
		"On":               true,
		"ForceOff":         true,
		"ForceRestart":     true,
		"Nmi":              true,
		"PushPowerButton":  true,
		"GracefulRestart":  true,
		"GracefulShutdown": true,
		"ForceOn":          true,
	}

	resetType := requestBody.ResetType
	if resetType == "" {
		resetType = "On" // Default reset type
	}

	if !validResetTypes[resetType] {
		sendRedfishError(w, "InvalidParameter", fmt.Sprintf("Invalid ResetType: %s", resetType), http.StatusBadRequest)
		return
	}

	// In a real implementation, this would trigger the actual reset
	// For demo purposes, we just return success
	w.WriteHeader(http.StatusNoContent)
}

// chassisHandler handles the chassis collection
func chassisHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)

	switch r.Method {
	case "GET":
		handleGetChassis(w, r)
	case "POST":
		handleCreateChassis(w, r)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetChassis returns the chassis collection
func handleGetChassis(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	chassis := models.NewChassisCollection()

	// Parse query parameters
	queryParams, err := parseQueryParameters(r.URL.Query())
	if err != nil {
		sendRedfishError(w, "QueryParameterError", err.Error(), http.StatusBadRequest)
		return
	}

	// Apply query parameters
	chassis = applyQueryParametersToChassis(chassis, queryParams)

	etag := generateETag(chassis)
	w.Header().Set("ETag", etag)

	// Check conditional GET
	if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
		normalizedETag := normalizeETag(etag)
		normalizedIfNoneMatch := normalizeETag(ifNoneMatch)
		if normalizedIfNoneMatch == normalizedETag || ifNoneMatch == "*" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	json.NewEncoder(w).Encode(chassis)
}

// handleGetChassisItem returns a specific chassis
func handleGetChassisItem(w http.ResponseWriter, r *http.Request, id string) {
	w.Header().Set("Content-Type", "application/json")

	chassis := models.NewChassis(id)
	etag := generateETag(chassis)
	w.Header().Set("ETag", etag)

	// Check conditional GET
	if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
		normalizedETag := normalizeETag(etag)
		normalizedIfNoneMatch := normalizeETag(ifNoneMatch)
		if normalizedIfNoneMatch == normalizedETag || ifNoneMatch == "*" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	json.NewEncoder(w).Encode(chassis)
}

// handleCreateChassis creates a new chassis (not typically allowed)
func handleCreateChassis(w http.ResponseWriter, r *http.Request) {
	sendRedfishError(w, "MethodNotAllowed", "Chassis creation not supported", http.StatusMethodNotAllowed)
}

// chassisItemHandler handles individual chassis resources
func chassisItemHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)

	// Extract chassis ID from URL path
	path := r.URL.Path
	id := path[len("/redfish/v1/Chassis/"):]

	switch r.Method {
	case "GET":
		handleGetChassisItem(w, r, id)
	case "PATCH":
		handleUpdateChassis(w, r, id)
	case "PUT":
		handleReplaceChassis(w, r, id)
	case "DELETE":
		handleDeleteChassis(w, r, id)
	default:
		methodNotAllowed(w, r)
	}
}

// handleUpdateChassis updates a chassis (PATCH)
func handleUpdateChassis(w http.ResponseWriter, r *http.Request, id string) {
	sendRedfishError(w, "MethodNotAllowed", "Chassis updates not supported", http.StatusMethodNotAllowed)
}

// handleReplaceChassis replaces a chassis (PUT)
func handleReplaceChassis(w http.ResponseWriter, r *http.Request, id string) {
	sendRedfishError(w, "MethodNotAllowed", "Chassis replacement not supported", http.StatusMethodNotAllowed)
}

// handleDeleteChassis deletes a chassis
func handleDeleteChassis(w http.ResponseWriter, r *http.Request, id string) {
	sendRedfishError(w, "MethodNotAllowed", "Chassis deletion not supported", http.StatusMethodNotAllowed)
}

// managersHandler handles the managers collection
func managersHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)

	switch r.Method {
	case "GET":
		handleGetManagers(w, r)
	case "POST":
		handleCreateManager(w, r)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetManagers returns the managers collection
func handleGetManagers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	managers := models.NewManagerCollection()

	// Parse query parameters
	queryParams, err := parseQueryParameters(r.URL.Query())
	if err != nil {
		sendRedfishError(w, "QueryParameterError", err.Error(), http.StatusBadRequest)
		return
	}

	// Apply query parameters
	managers = applyQueryParametersToManagers(managers, queryParams)

	etag := generateETag(managers)
	w.Header().Set("ETag", etag)

	// Check conditional GET
	if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
		normalizedETag := normalizeETag(etag)
		normalizedIfNoneMatch := normalizeETag(ifNoneMatch)
		if normalizedIfNoneMatch == normalizedETag || ifNoneMatch == "*" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	json.NewEncoder(w).Encode(managers)
}

// handleGetManager returns a specific manager
func handleGetManager(w http.ResponseWriter, r *http.Request, id string) {
	w.Header().Set("Content-Type", "application/json")

	manager := models.NewManager(id)
	etag := generateETag(manager)
	w.Header().Set("ETag", etag)

	// Check conditional GET
	if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
		normalizedETag := normalizeETag(etag)
		normalizedIfNoneMatch := normalizeETag(ifNoneMatch)
		if normalizedIfNoneMatch == normalizedETag || ifNoneMatch == "*" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	json.NewEncoder(w).Encode(manager)
}

// handleCreateManager creates a new manager (not typically allowed)
func handleCreateManager(w http.ResponseWriter, r *http.Request) {
	sendRedfishError(w, "MethodNotAllowed", "Manager creation not supported", http.StatusMethodNotAllowed)
}

// managerHandler handles individual manager resources and actions
func managerHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)

	path := r.URL.Path

	// Check if this is an action request
	if strings.Contains(path, "/Actions/") {
		handleManagerAction(w, r, path)
		return
	}

	// Extract manager ID from URL path
	id := path[len("/redfish/v1/Managers/"):]

	switch r.Method {
	case "GET":
		handleGetManager(w, r, id)
	case "PATCH":
		handleUpdateManager(w, r, id)
	case "PUT":
		handleReplaceManager(w, r, id)
	case "DELETE":
		handleDeleteManager(w, r, id)
	default:
		methodNotAllowed(w, r)
	}
}

// handleUpdateManager updates a manager (PATCH)
func handleUpdateManager(w http.ResponseWriter, r *http.Request, id string) {
	sendRedfishError(w, "MethodNotAllowed", "Manager updates not supported", http.StatusMethodNotAllowed)
}

// handleReplaceManager replaces a manager (PUT)
func handleReplaceManager(w http.ResponseWriter, r *http.Request, id string) {
	sendRedfishError(w, "MethodNotAllowed", "Manager replacement not supported", http.StatusMethodNotAllowed)
}

// handleDeleteManager deletes a manager
func handleDeleteManager(w http.ResponseWriter, r *http.Request, id string) {
	sendRedfishError(w, "MethodNotAllowed", "Manager deletion not supported", http.StatusMethodNotAllowed)
}

// handleManagerAction handles Manager actions
func handleManagerAction(w http.ResponseWriter, r *http.Request, path string) {
	// Extract action from path: /redfish/v1/Managers/{id}/Actions/{ActionName}
	parts := strings.Split(path, "/")
	if len(parts) < 7 || parts[5] != "Actions" {
		sendRedfishError(w, "InvalidAction", "Invalid action URI format", http.StatusBadRequest)
		return
	}

	actionName := parts[6]
	managerId := parts[4]

	switch r.Method {
	case "GET":
		switch actionName {
		case "Manager.Reset":
			handleManagerResetActionInfo(w, r, managerId)
		default:
			sendRedfishError(w, "ActionNotSupported", fmt.Sprintf("Action %s not supported for Manager", actionName), http.StatusBadRequest)
		}
	case "POST":
		switch actionName {
		case "Manager.Reset":
			handleManagerReset(w, r, managerId)
		default:
			sendRedfishError(w, "ActionNotSupported", fmt.Sprintf("Action %s not supported for Manager", actionName), http.StatusBadRequest)
		}
	default:
		methodNotAllowed(w, r)
	}
}

// handleManagerResetActionInfo returns ActionInfo for Manager.Reset
func handleManagerResetActionInfo(w http.ResponseWriter, r *http.Request, managerId string) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"@odata.context": "/redfish/v1/$metadata#ActionInfo.ActionInfo",
		"@odata.id":      fmt.Sprintf("/redfish/v1/Managers/%s/Actions/Manager.Reset", managerId),
		"@odata.type":    "#ActionInfo.v1_1_2.ActionInfo",
		"Id":             "Manager.Reset",
		"Name":           "Manager Reset",
		"Parameters": []map[string]interface{}{
			{
				"Name":            "ResetType",
				"Required":        false,
				"DataType":        "String",
				"AllowableValues": []string{"ForceRestart", "GracefulRestart"},
			},
		},
	}

	etag := generateETag(response)
	w.Header().Set("ETag", etag)

	// Check conditional GET
	if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
		normalizedETag := normalizeETag(etag)
		normalizedIfNoneMatch := normalizeETag(ifNoneMatch)
		if normalizedIfNoneMatch == normalizedETag || ifNoneMatch == "*" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	json.NewEncoder(w).Encode(response)
}

// handleManagerReset handles the Manager.Reset action
func handleManagerReset(w http.ResponseWriter, r *http.Request, managerId string) {
	// Parse request body for ResetType parameter
	var requestBody struct {
		ResetType string `json:"ResetType"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil && err.Error() != "EOF" {
		sendRedfishError(w, "MalformedJSON", "Invalid JSON in request body", http.StatusBadRequest)
		return
	}

	// Validate ResetType parameter
	validResetTypes := map[string]bool{
		"ForceRestart":    true,
		"GracefulRestart": true,
	}

	resetType := requestBody.ResetType
	if resetType == "" {
		resetType = "GracefulRestart" // Default reset type for managers
	}

	if !validResetTypes[resetType] {
		sendRedfishError(w, "InvalidParameter", fmt.Sprintf("Invalid ResetType: %s", resetType), http.StatusBadRequest)
		return
	}

	// In a real implementation, this would trigger the actual manager reset
	// For demo purposes, we just return success
	w.WriteHeader(http.StatusNoContent)
}

// setRedfishHeaders sets common Redfish headers
func setRedfishHeaders(w http.ResponseWriter) {
	w.Header().Set("OData-Version", "4.0")
	w.Header().Set("Cache-Control", "no-cache")
}

// methodNotAllowed sends a 405 Method Not Allowed response
func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	sendRedfishError(w, "MethodNotAllowed", fmt.Sprintf("HTTP method %s not allowed for this resource", r.Method), http.StatusMethodNotAllowed)
}

// generateETag generates a simple ETag for a resource
func generateETag(data interface{}) string {
	// Simple ETag generation - in production, this should be more sophisticated
	// For now, use a hash of the JSON representation
	jsonBytes, _ := json.Marshal(data)
	hash := fmt.Sprintf("%x", md5.Sum(jsonBytes))
	return fmt.Sprintf(`"%s"`, hash[:8])
}

// normalizeETag normalizes an ETag for comparison (removes quotes if present)
func normalizeETag(etag string) string {
	if len(etag) >= 2 && etag[0] == '"' && etag[len(etag)-1] == '"' {
		return etag[1 : len(etag)-1]
	}
	return etag
}

// sendRedfishError sends a Redfish-compliant error response
func sendRedfishError(w http.ResponseWriter, code, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := models.RedfishError{
		Error: struct {
			Code    string           `json:"code"`
			Message string           `json:"message"`
			Details []models.Message `json:"@Message.ExtendedInfo,omitempty"`
		}{
			Code:    code,
			Message: message,
			Details: []models.Message{
				{
					MessageID:  code,
					Message:    message,
					Severity:   "Critical",
					Resolution: "Check the request method and try again",
				},
			},
		},
	}

	json.NewEncoder(w).Encode(errorResponse)
}

// QueryParameters represents parsed OData query parameters
type QueryParameters struct {
	Top     int      `json:"top,omitempty"`
	Skip    int      `json:"skip,omitempty"`
	Select  []string `json:"select,omitempty"`
	Expand  []string `json:"expand,omitempty"`
	Filter  string   `json:"filter,omitempty"`
	OrderBy string   `json:"orderby,omitempty"`
}

// parseQueryParameters parses OData query parameters from the URL
func parseQueryParameters(query url.Values) (*QueryParameters, error) {
	params := &QueryParameters{}

	// Parse $top
	if topStr := query.Get("$top"); topStr != "" {
		top, err := strconv.Atoi(topStr)
		if err != nil || top < 0 {
			return nil, fmt.Errorf("invalid $top parameter: %s", topStr)
		}
		params.Top = top
	}

	// Parse $skip
	if skipStr := query.Get("$skip"); skipStr != "" {
		skip, err := strconv.Atoi(skipStr)
		if err != nil || skip < 0 {
			return nil, fmt.Errorf("invalid $skip parameter: %s", skipStr)
		}
		params.Skip = skip
	}

	// Parse $select
	if selectStr := query.Get("$select"); selectStr != "" {
		params.Select = strings.Split(strings.ReplaceAll(selectStr, " ", ""), ",")
	}

	// Parse $expand
	if expandStr := query.Get("$expand"); expandStr != "" {
		params.Expand = strings.Split(strings.ReplaceAll(expandStr, " ", ""), ",")
	}

	// Parse $filter
	params.Filter = query.Get("$filter")

	// Parse $orderby
	params.OrderBy = query.Get("$orderby")

	return params, nil
}

// applyQueryParameters applies query parameters to a ComputerSystemCollection
func applyQueryParametersToSystems(collection *models.ComputerSystemCollection, params *QueryParameters) *models.ComputerSystemCollection {
	if params == nil {
		return collection
	}

	result := *collection // Create a copy

	// Apply $filter if specified (basic implementation)
	if params.Filter != "" {
		result = applyFilterToSystems(result, params.Filter)
	}

	// Apply $skip and $top for pagination
	totalMembers := len(result.Members)
	start := params.Skip
	if start > totalMembers {
		start = totalMembers
	}

	end := totalMembers
	if params.Top > 0 && start+params.Top < totalMembers {
		end = start + params.Top
	}

	result.Members = result.Members[start:end]
	result.MembersODataCount = len(result.Members)

	return &result
}

// applyFilterToSystems applies basic $filter to ComputerSystemCollection
func applyFilterToSystems(collection models.ComputerSystemCollection, filter string) models.ComputerSystemCollection {
	// Very basic filter implementation
	// In a real implementation, this would parse OData filter expressions

	result := collection

	// For demo purposes, support simple equality filters
	// Note: URL decoding happens in parseQueryParameters
	if strings.Contains(filter, "PowerState eq 'On'") || strings.Contains(filter, "PowerState eq \"On\"") {
		// Keep all members (since our demo system is 'On')
	} else if strings.Contains(filter, "PowerState eq 'Off'") || strings.Contains(filter, "PowerState eq \"Off\"") {
		// Remove all members (since our demo system is not 'Off')
		result.Members = []models.ODataID{}
		result.MembersODataCount = 0
	}

	return result
}

// applyQueryParametersToChassis applies query parameters to a ChassisCollection
func applyQueryParametersToChassis(collection *models.ChassisCollection, params *QueryParameters) *models.ChassisCollection {
	if params == nil {
		return collection
	}

	result := *collection // Create a copy

	// Apply $skip and $top for pagination
	totalMembers := len(result.Members)
	start := params.Skip
	if start > totalMembers {
		start = totalMembers
	}

	end := totalMembers
	if params.Top > 0 && start+params.Top < totalMembers {
		end = start + params.Top
	}

	result.Members = result.Members[start:end]
	result.MembersODataCount = len(result.Members)

	return &result
}

// applyQueryParametersToManagers applies query parameters to a ManagerCollection
func applyQueryParametersToManagers(collection *models.ManagerCollection, params *QueryParameters) *models.ManagerCollection {
	if params == nil {
		return collection
	}

	result := *collection // Create a copy

	// Apply $skip and $top for pagination
	totalMembers := len(result.Members)
	start := params.Skip
	if start > totalMembers {
		start = totalMembers
	}

	end := totalMembers
	if params.Top > 0 && start+params.Top < totalMembers {
		end = start + params.Top
	}

	result.Members = result.Members[start:end]
	result.MembersODataCount = len(result.Members)

	return &result
}

// applySelectToSystem applies $select filtering to a ComputerSystem
// For now, this validates the select parameters but returns the full object
// TODO: Implement actual property filtering
func applySelectToSystem(system *models.ComputerSystem, selectProps []string) *models.ComputerSystem {
	// Validate that requested properties exist on ComputerSystem
	validProps := map[string]bool{
		"@odata.context":     true,
		"@odata.id":          true,
		"@odata.type":        true,
		"Id":                 true,
		"Name":               true,
		"Description":        true,
		"SystemType":         true,
		"AssetTag":           true,
		"Manufacturer":       true,
		"Model":              true,
		"SKU":                true,
		"SerialNumber":       true,
		"PartNumber":         true,
		"UUID":               true,
		"HostName":           true,
		"Status":             true,
		"PowerState":         true,
		"Boot":               true,
		"BiosVersion":        true,
		"ProcessorSummary":   true,
		"MemorySummary":      true,
		"Storage":            true,
		"Processors":         true,
		"Memory":             true,
		"StorageControllers": true,
		"NetworkInterfaces":  true,
		"EthernetInterfaces": true,
		"LogServices":        true,
		"Links":              true,
		"Actions":            true,
		"Oem":                true,
	}

	for _, prop := range selectProps {
		if !validProps[prop] {
			// For now, ignore invalid properties rather than erroring
			// In a full implementation, this might return an error
		}
	}

	// Return the full system for now
	// TODO: Implement actual selective property marshaling
	return system
}

// applyExpandToSystem applies $expand to include related resources inline
func applyExpandToSystem(system *models.ComputerSystem, expandProps []string) *models.ComputerSystem {
	// Create a copy to avoid modifying the original
	result := *system

	// For each expand property, inline the related resource
	for _, prop := range expandProps {
		switch prop {
		case "Chassis":
			// Expand chassis information
			// In Redfish, expanded resources are typically added as new properties
			// For this demo, we'll just ensure the Links.Chassis points to expanded data
			result.Links.Chassis = []models.ODataID{"/redfish/v1/Chassis/1"}

		case "ManagedBy":
			// Expand manager information
			result.Links.ManagedBy = []models.ODataID{"/redfish/v1/Managers/1"}

		// Add more expandable properties as needed
		default:
			// Unknown expand property - ignore for now
		}
	}

	return &result
}
