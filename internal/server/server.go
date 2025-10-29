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
	"sync"
	"time"

	"github.com/user/redfish-server/internal/auth"
	"github.com/user/redfish-server/internal/config"
	"github.com/user/redfish-server/internal/middleware"
	"github.com/user/redfish-server/internal/models"
)

// Global task storage for demo purposes
var (
	tasksMutex sync.RWMutex
	tasks      = make(map[string]*models.Task)
)

// Server represents the Redfish HTTP server
type Server struct {
	httpServer    *http.Server
	config        *config.Config
	subscriptions map[string]*models.EventSubscription // In-memory storage for demo
	tasks         map[string]*models.Task              // In-memory storage for demo
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
		httpServer:    httpServer,
		config:        cfg,
		subscriptions: make(map[string]*models.EventSubscription),
		tasks:         make(map[string]*models.Task),
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

// SendEvent sends an event to all matching subscribers
func (s *Server) SendEvent(event *models.Event) {
	// For now, just log the event
	fmt.Printf("Event sent: %+v\n", event)
	// In a real implementation, this would filter subscribers and send HTTP POSTs
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
	mux.HandleFunc("/redfish/v1/SessionService/Sessions/", sessionItemHandler)
	mux.HandleFunc("/redfish/v1/SessionService/Sessions", sessionsHandler)
	mux.HandleFunc("/redfish/v1/SessionService/Sessions/Members", sessionsHandler)
	mux.HandleFunc("/redfish/v1/SessionService", sessionServiceHandler)

	// Account service endpoints
	mux.HandleFunc("/redfish/v1/AccountService/Accounts/", accountHandler)
	mux.HandleFunc("/redfish/v1/AccountService/Accounts", accountsHandler)
	mux.HandleFunc("/redfish/v1/AccountService/Roles/", roleHandler)
	mux.HandleFunc("/redfish/v1/AccountService/Roles", rolesHandler)
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

	// Event service endpoints
	mux.HandleFunc("/redfish/v1/EventService/Subscriptions/", eventSubscriptionHandler)
	mux.HandleFunc("/redfish/v1/EventService/Subscriptions", eventSubscriptionsHandler)
	mux.HandleFunc("/redfish/v1/EventService/SSE", eventSSEHandler)
	mux.HandleFunc("/redfish/v1/EventService", eventServiceHandler)

	// Task service endpoints
	mux.HandleFunc("/redfish/v1/TaskService/Tasks/", taskHandler)
	mux.HandleFunc("/redfish/v1/TaskService/Tasks", tasksHandler)
	mux.HandleFunc("/redfish/v1/TaskService", taskServiceHandler)

	// Registry endpoints
	mux.HandleFunc("/redfish/v1/Registries/", registryHandler)
	mux.HandleFunc("/redfish/v1/Registries", registriesHandler)

	// OEM endpoints
	mux.HandleFunc("/redfish/v1/Oem/Contoso/CustomAction", oemCustomActionHandler)

	// OpenAPI endpoint
	mux.HandleFunc("/redfish/v1/openapi.yaml", openapiHandler)

	// Redfish root endpoint
	mux.HandleFunc("/redfish", redfishRootHandler)

	// Redfish v1 root endpoint - handle both /redfish/v1 and /redfish/v1/
	mux.HandleFunc("/redfish/v1", serviceRootHandler)
	mux.HandleFunc("/redfish/v1/", serviceRootHandler)
}

// healthHandler handles health check requests
func healthHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Allow", "GET")

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

// openapiHandler serves the OpenAPI specification
func openapiHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Allow", "GET")

	switch r.Method {
	case "GET":
		handleGetOpenAPI(w, r)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetOpenAPI returns the OpenAPI specification
func handleGetOpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")

	// Basic OpenAPI spec
	openapi := `openapi: 3.0.0
info:
  title: Redfish API
  version: 1.0.0
  description: Redfish API specification
paths:
  /redfish/v1/:
    get:
      summary: Get service root
      responses:
        '200':
          description: OK
`

	etag := generateETag(openapi)
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

	w.Write([]byte(openapi))
}

// serviceRootHandler handles the Redfish service root
// redfishRootHandler handles requests to /redfish
func redfishRootHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Allow", "GET, HEAD")

	switch r.Method {
	case "GET":
		handleGetRedfishRoot(w, r)
	case "HEAD":
		handleGetRedfishRoot(w, r)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetRedfishRoot returns the redfish root
func handleGetRedfishRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"v1": "/redfish/v1/"}`))
}

func serviceRootHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Allow", "GET, HEAD")

	switch r.Method {
	case "GET":
		handleGetServiceRoot(w, r)
	case "HEAD":
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
	w.Header().Set("Allow", "GET")

	switch r.Method {
	case "GET":
		handleGetMetadata(w, r)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetMetadata returns the OData metadata document
func handleGetMetadata(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Content-Type", "application/xml;charset=utf-8")

	// Basic metadata document with EntityContainer
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
      <EntityContainer Name="Service">
        <EntitySet Name="ServiceRoot" EntityType="Service.ServiceRoot" />
        <EntitySet Name="Systems" EntityType="ComputerSystemCollection.ComputerSystemCollection" />
        <EntitySet Name="Chassis" EntityType="ChassisCollection.ChassisCollection" />
        <EntitySet Name="Managers" EntityType="ManagerCollection.ManagerCollection" />
        <EntitySet Name="TaskService" EntityType="TaskService.TaskService" />
        <EntitySet Name="SessionService" EntityType="SessionService.SessionService" />
        <EntitySet Name="AccountService" EntityType="AccountService.AccountService" />
        <EntitySet Name="EventService" EntityType="EventService.EventService" />
        <EntitySet Name="Registries" EntityType="MessageRegistryFileCollection.MessageRegistryFileCollection" />
      </EntityContainer>
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
	w.Header().Set("Allow", "GET")

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
	w.Header().Set("Allow", "GET")

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
	w.Header().Set("Allow", "GET, POST")

	switch r.Method {
	case "GET":
		handleGetSessions(w, r)
	case "POST":
		handleCreateSession(w, r)
	default:
		methodNotAllowed(w, r)
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
	var username, password string
	var ok bool

	// Try Basic Auth first
	username, password, ok = r.BasicAuth()
	if !ok {
		// Try JSON body with UserName/Password
		var requestBody struct {
			UserName string `json:"UserName"`
			Password string `json:"Password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err == nil {
			username = requestBody.UserName
			password = requestBody.Password
			ok = true
		}
	}

	if !ok || username == "" || password == "" {
		w.Header().Set("WWW-Authenticate", `Basic realm="Redfish Service"`)
		http.Error(w, `{"error": {"code": "Base.1.0.InsufficientPrivilege", "message": "Authentication required"}}`, http.StatusUnauthorized)
		return
	}

	// Validate credentials
	authService := auth.GetAuthService()
	if !authService.ValidateBasicAuth(username, password) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Redfish Service"`)
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
	w.Header().Set("Location", "https://"+r.Host+"/redfish/v1/SessionService/Sessions/"+token)
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

// sessionItemHandler handles individual session resources
func sessionItemHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Allow", "GET, DELETE")

	// Extract session ID from URL path
	sessionID := strings.TrimPrefix(r.URL.Path, "/redfish/v1/SessionService/Sessions/")

	// Validate session exists and authentication
	authService := auth.GetAuthService()
	_, sessionExists := authService.ValidateSessionToken(sessionID)
	if !sessionExists {
		sendRedfishError(w, "ResourceNotFound", "Session not found", http.StatusNotFound)
		return
	}

	// For session resources, the session ID in the URL serves as authentication
	// No additional token validation required

	switch r.Method {
	case "GET":
		handleGetSession(w, r, sessionID)
	case "DELETE":
		handleDeleteSession(w, r, sessionID)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetSession returns a specific session
func handleGetSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	// Session existence already validated in sessionItemHandler
	authService := auth.GetAuthService()
	username, _ := authService.ValidateSessionToken(sessionID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := fmt.Sprintf(`{
		"@odata.context": "/redfish/v1/$metadata#Session.Session",
		"@odata.id": "/redfish/v1/SessionService/Sessions/%s",
		"@odata.type": "#Session.v1_1_6.Session",
		"Id": "%s",
		"Name": "User Session",
		"UserName": "%s"
	}`, sessionID, sessionID, username)

	w.Write([]byte(response))
}

// handleDeleteSession terminates a session
func handleDeleteSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	authService := auth.GetAuthService()
	authService.DeleteSession(sessionID)
	w.WriteHeader(http.StatusNoContent)
}

// accountServiceHandler handles the AccountService resource
func accountServiceHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Allow", "GET, PATCH")

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
	w.Header().Set("Allow", "GET")

	switch r.Method {
	case "GET":
		handleGetAccounts(w, r)
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
	w.Header().Set("Allow", "GET, PATCH, PUT, DELETE")

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

// rolesHandler handles the roles collection
func rolesHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Allow", "GET, HEAD")

	switch r.Method {
	case "GET":
		handleGetRoles(w, r)
	case "HEAD":
		handleGetRoles(w, r)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetRoles returns the roles collection
func handleGetRoles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	roles := models.NewRoleCollection()
	etag := generateETag(roles)
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

	json.NewEncoder(w).Encode(roles)
}

// roleHandler handles individual role resources
func roleHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Allow", "GET, HEAD")

	// Extract role ID from URL path
	path := r.URL.Path
	id := path[len("/redfish/v1/AccountService/Roles/"):]

	switch r.Method {
	case "GET":
		handleGetRole(w, r, id)
	case "HEAD":
		handleGetRole(w, r, id)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetRole returns a specific role
func handleGetRole(w http.ResponseWriter, r *http.Request, id string) {
	w.Header().Set("Content-Type", "application/json")

	var role *models.Role
	switch id {
	case "Administrator":
		role = models.NewRole("Administrator", "Administrator", []string{"Login", "ConfigureManager", "ConfigureUsers", "ConfigureComponents", "ConfigureSelf"}, true)
	case "Operator":
		role = models.NewRole("Operator", "Operator", []string{"Login", "ConfigureComponents", "ConfigureSelf"}, true)
	case "ReadOnly":
		role = models.NewRole("ReadOnly", "ReadOnly", []string{"Login", "ConfigureSelf"}, true)
	default:
		sendRedfishError(w, "ResourceNotFound", "Role not found", http.StatusNotFound)
		return
	}

	etag := generateETag(role)
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

	json.NewEncoder(w).Encode(role)
}

// systemsHandler handles the computer systems collection
func systemsHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Allow", "GET")

	switch r.Method {
	case "GET":
		handleGetSystems(w, r)
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
	w.Header().Set("Allow", "GET, PATCH, PUT, DELETE")

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

	// Create a task for the reset operation
	id := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("reset-%s-%s-%s", systemId, resetType, time.Now().String()))))[:8]

	task := models.NewTask(id, "POST", fmt.Sprintf("/redfish/v1/Systems/%s/Actions/ComputerSystem.Reset", systemId))
	task.Payload.JsonBody = fmt.Sprintf(`{"ResetType": "%s"}`, resetType)

	// Simulate asynchronous reset operation
	go func() {
		time.Sleep(3 * time.Second) // Simulate reset time
		tasksMutex.Lock()
		task.UpdateTaskState("Completed")
		task.SetPercentComplete(100)
		task.AddMessage(models.Message{
			MessageID:  "Base.1.12.Success",
			Message:    fmt.Sprintf("Computer system %s reset (%s) completed successfully", systemId, resetType),
			Severity:   "OK",
			Resolution: "No action required",
		})
		tasksMutex.Unlock()
	}()

	tasksMutex.Lock()
	tasks[id] = task
	tasksMutex.Unlock()

	// Return the task location
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", string(task.ODataID))
	w.WriteHeader(http.StatusAccepted)

	response := map[string]interface{}{
		"@odata.id":   task.ODataID,
		"@odata.type": task.ODataType,
		"Id":          task.ID,
		"Name":        task.Name,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// chassisHandler handles the chassis collection
func chassisHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Allow", "GET")

	switch r.Method {
	case "GET":
		handleGetChassis(w, r)
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
	w.Header().Set("Allow", "GET, PATCH, PUT, DELETE")

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
	w.Header().Set("Allow", "GET")

	switch r.Method {
	case "GET":
		handleGetManagers(w, r)
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
	w.Header().Set("Allow", "GET, PATCH, PUT, DELETE")

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

	// Create a task for the manager reset operation
	id := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("mgr-reset-%s-%s-%s", managerId, resetType, time.Now().String()))))[:8]

	task := models.NewTask(id, "POST", fmt.Sprintf("/redfish/v1/Managers/%s/Actions/Manager.Reset", managerId))
	task.Payload.JsonBody = fmt.Sprintf(`{"ResetType": "%s"}`, resetType)

	// Simulate asynchronous manager reset operation
	go func() {
		time.Sleep(5 * time.Second) // Simulate longer reset time for manager
		tasksMutex.Lock()
		task.UpdateTaskState("Completed")
		task.SetPercentComplete(100)
		task.AddMessage(models.Message{
			MessageID:  "Base.1.12.Success",
			Message:    fmt.Sprintf("Manager %s reset (%s) completed successfully", managerId, resetType),
			Severity:   "OK",
			Resolution: "No action required",
		})
		tasksMutex.Unlock()
	}()

	tasksMutex.Lock()
	tasks[id] = task
	tasksMutex.Unlock()

	// Return the task location
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", string(task.ODataID))
	w.WriteHeader(http.StatusAccepted)

	response := map[string]interface{}{
		"@odata.id":   task.ODataID,
		"@odata.type": task.ODataType,
		"Id":          task.ID,
		"Name":        task.Name,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// setRedfishHeaders sets common Redfish headers
func setRedfishHeaders(w http.ResponseWriter) {
	w.Header().Set("OData-Version", "4.0")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Link", "</redfish/v1/$metadata>; rel=describedby")
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
		result.Members = []models.Link{}
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
			result.Links.Chassis = []models.Link{models.Link{ODataID: "/redfish/v1/Chassis/1"}}

		case "ManagedBy":
			// Expand manager information
			result.Links.ManagedBy = []models.Link{models.Link{ODataID: "/redfish/v1/Managers/1"}}

		// Add more expandable properties as needed
		default:
			// Unknown expand property - ignore for now
		}
	}

	return &result
}

// eventServiceHandler handles EventService requests
func eventServiceHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Allow", "GET")

	switch r.Method {
	case "GET":
		handleGetEventService(w, r)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetEventService returns the EventService resource
func handleGetEventService(w http.ResponseWriter, r *http.Request) {
	eventService := models.NewEventService()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(eventService); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// eventSubscriptionsHandler handles EventService Subscriptions collection requests
func eventSubscriptionsHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Allow", "GET, POST")

	switch r.Method {
	case "GET":
		handleGetEventSubscriptions(w, r)
	case "POST":
		handlePostEventSubscription(w, r)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetEventSubscriptions returns the EventSubscriptions collection
func handleGetEventSubscriptions(w http.ResponseWriter, r *http.Request) {
	// For now, return empty collection
	collection := models.Collection{
		ODataContext:      "/redfish/v1/$metadata#EventDestinationCollection.EventDestinationCollection",
		ODataID:           "/redfish/v1/EventService/Subscriptions",
		ODataType:         "#EventDestinationCollection.EventDestinationCollection",
		Name:              "Event Subscriptions Collection",
		Members:           []models.Link{},
		MembersODataCount: 0,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(collection); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handlePostEventSubscription creates a new event subscription
func handlePostEventSubscription(w http.ResponseWriter, r *http.Request) {
	var subscription models.EventSubscription
	if err := json.NewDecoder(r.Body).Decode(&subscription); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if subscription.Destination == "" {
		http.Error(w, "Destination is required", http.StatusBadRequest)
		return
	}
	if subscription.Protocol == "" {
		subscription.Protocol = "Redfish" // Default
	}

	// Generate ID (in a real implementation, this would be stored)
	id := fmt.Sprintf("%x", md5.Sum([]byte(subscription.Destination+time.Now().String())))[:8]

	// Create the subscription
	newSubscription := models.NewEventSubscription(id, subscription.Destination, subscription.Protocol)
	if subscription.Context != "" {
		newSubscription.Context = subscription.Context
	}
	if len(subscription.RegistryPrefixes) > 0 {
		newSubscription.RegistryPrefixes = subscription.RegistryPrefixes
	}
	if len(subscription.ResourceTypes) > 0 {
		newSubscription.ResourceTypes = subscription.ResourceTypes
	}
	if len(subscription.Severities) > 0 {
		newSubscription.Severities = subscription.Severities
	}
	newSubscription.IncludeOriginOfCondition = subscription.IncludeOriginOfCondition
	newSubscription.SubordinateResources = subscription.SubordinateResources

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", string(newSubscription.ODataID))
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(newSubscription); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// eventSubscriptionHandler handles individual EventSubscription requests
func eventSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Allow", "GET, DELETE")

	// Extract subscription ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/redfish/v1/EventService/Subscriptions/")
	parts := strings.Split(path, "/")
	id := parts[0]

	if id == "" {
		http.Error(w, "Subscription ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		handleGetEventSubscription(w, r, id)
	case "DELETE":
		handleDeleteEventSubscription(w, r, id)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetEventSubscription returns a specific event subscription
func handleGetEventSubscription(w http.ResponseWriter, r *http.Request, id string) {
	// For now, return 404 as we don't persist subscriptions
	http.Error(w, "Subscription not found", http.StatusNotFound)
}

// handleDeleteEventSubscription deletes an event subscription
func handleDeleteEventSubscription(w http.ResponseWriter, r *http.Request, id string) {
	// For now, return 404 as we don't persist subscriptions
	http.Error(w, "Subscription not found", http.StatusNotFound)
}

// eventSSEHandler handles Server-Sent Events requests
func eventSSEHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Allow", "GET")

	switch r.Method {
	case "GET":
		handleGetEventSSE(w, r)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetEventSSE handles Server-Sent Events connections
func handleGetEventSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// For now, just send a test event and close
	// In a real implementation, this would maintain persistent connections
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Send a heartbeat event
	fmt.Fprintf(w, "event: heartbeat\n")
	fmt.Fprintf(w, "data: {\"EventType\": \"Heartbeat\", \"Message\": \"Connection established\"}\n\n")
	flusher.Flush()

	// Close the connection after a short time for demo purposes
	time.Sleep(1 * time.Second)
}

// registriesHandler handles Registries collection requests
func registriesHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Allow", "GET")

	switch r.Method {
	case "GET":
		handleGetRegistries(w, r)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetRegistries returns the Registries collection
func handleGetRegistries(w http.ResponseWriter, r *http.Request) {
	// Create sample registry files
	baseRegistry := models.NewMessageRegistryFile("Base.1.0.0", "Base.1.0")
	taskRegistry := models.NewMessageRegistryFile("Task.1.0.0", "Task.1.0")

	members := []models.Link{
		models.Link{ODataID: baseRegistry.ODataID},
		models.Link{ODataID: taskRegistry.ODataID},
	}

	collection := models.Collection{
		ODataContext:      "/redfish/v1/$metadata#MessageRegistryFileCollection.MessageRegistryFileCollection",
		ODataID:           "/redfish/v1/Registries",
		ODataType:         "#MessageRegistryFileCollection.MessageRegistryFileCollection",
		Name:              "Message Registry File Collection",
		Members:           members,
		MembersODataCount: len(members),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(collection); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// registryHandler handles individual Registry requests
func registryHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Allow", "GET")

	// Extract registry ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/redfish/v1/Registries/")
	parts := strings.Split(path, "/")
	id := parts[0]

	if id == "" {
		http.Error(w, "Registry ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		handleGetRegistry(w, r, id)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetRegistry returns a specific registry file
func handleGetRegistry(w http.ResponseWriter, r *http.Request, id string) {
	var registry *models.MessageRegistryFile

	switch id {
	case "Base.1.0.0":
		registry = models.NewMessageRegistryFile("Base.1.0.0", "Base.1.0")
	case "Task.1.0.0":
		registry = models.NewMessageRegistryFile("Task.1.0.0", "Task.1.0")
	default:
		http.Error(w, "Registry not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(registry); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// oemCustomActionHandler handles OEM custom action requests
func oemCustomActionHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Allow", "POST")

	switch r.Method {
	case "POST":
		handleOemCustomAction(w, r)
	default:
		methodNotAllowed(w, r)
	}
}

// handleOemCustomAction handles the OEM custom action
func handleOemCustomAction(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Action     string                 `json:"Action"`
		Parameters map[string]interface{} `json:"Parameters,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil && err.Error() != "EOF" {
		sendRedfishError(w, "MalformedJSON", "Invalid JSON in request body", http.StatusBadRequest)
		return
	}

	// Simulate OEM-specific action processing
	response := map[string]interface{}{
		"@odata.type": "#OemCustomAction.v1_0_0.Response",
		"Action":      requestBody.Action,
		"Status":      "Success",
		"Message":     "OEM custom action executed successfully",
		"Timestamp":   time.Now().Format(time.RFC3339),
	}

	if requestBody.Parameters != nil {
		response["Parameters"] = requestBody.Parameters
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// taskServiceHandler handles TaskService requests
func taskServiceHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Allow", "GET")

	switch r.Method {
	case "GET":
		handleGetTaskService(w, r)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetTaskService returns the TaskService resource
func handleGetTaskService(w http.ResponseWriter, r *http.Request) {
	taskService := models.NewTaskService()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(taskService); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// tasksHandler handles TaskService Tasks collection requests
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Allow", "GET, POST")

	switch r.Method {
	case "GET":
		handleGetTasks(w, r)
	case "POST":
		handlePostTask(w, r)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetTasks returns the Tasks collection
func handleGetTasks(w http.ResponseWriter, r *http.Request) {
	tasksMutex.RLock()
	defer tasksMutex.RUnlock()

	members := make([]models.Link, 0, len(tasks))
	for _, task := range tasks {
		members = append(members, models.Link{ODataID: task.ODataID})
	}

	collection := models.Collection{
		ODataContext:      "/redfish/v1/$metadata#TaskCollection.TaskCollection",
		ODataID:           "/redfish/v1/TaskService/Tasks",
		ODataType:         "#TaskCollection.TaskCollection",
		Name:              "Task Collection",
		Members:           members,
		MembersODataCount: len(members),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(collection); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handlePostTask creates a new task
func handlePostTask(w http.ResponseWriter, r *http.Request) {
	// For demo purposes, create a simple task
	// In a real implementation, this would parse task creation parameters
	id := fmt.Sprintf("%x", md5.Sum([]byte(time.Now().String())))[:8]

	task := models.NewTask(id, "POST", "/redfish/v1/TaskService/Tasks")

	// Simulate task execution
	go func() {
		time.Sleep(2 * time.Second) // Simulate work
		tasksMutex.Lock()
		task.UpdateTaskState("Running")
		task.SetPercentComplete(50)
		tasksMutex.Unlock()

		time.Sleep(2 * time.Second) // More work
		tasksMutex.Lock()
		task.UpdateTaskState("Completed")
		task.SetPercentComplete(100)
		tasksMutex.Unlock()
	}()

	tasksMutex.Lock()
	tasks[id] = task
	tasksMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", string(task.ODataID))
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(task); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// taskHandler handles individual Task requests
func taskHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Allow", "GET, DELETE")

	// Extract task ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/redfish/v1/TaskService/Tasks/")
	parts := strings.Split(path, "/")
	id := parts[0]

	if id == "" {
		http.Error(w, "Task ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		handleGetTask(w, r, id)
	case "DELETE":
		handleDeleteTask(w, r, id)
	default:
		methodNotAllowed(w, r)
	}
}

// handleGetTask returns a specific task
func handleGetTask(w http.ResponseWriter, r *http.Request, id string) {
	tasksMutex.RLock()
	task, exists := tasks[id]
	tasksMutex.RUnlock()

	if !exists {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(task); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleDeleteTask deletes a task
func handleDeleteTask(w http.ResponseWriter, r *http.Request, id string) {
	tasksMutex.Lock()
	_, exists := tasks[id]
	if exists {
		delete(tasks, id)
	}
	tasksMutex.Unlock()

	if !exists {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
