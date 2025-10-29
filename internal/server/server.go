package server

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok", "service": "redfish-server"}`))
}

// serviceRootHandler handles the Redfish service root
func serviceRootHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Content-Type", "application/json")

	serviceRoot := models.NewServiceRoot()
	json.NewEncoder(w).Encode(serviceRoot)
}

// metadataHandler serves the OData metadata document
func metadataHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
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

	w.Write([]byte(metadata))
}

// odataHandler serves the OData service document
func odataHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
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

	w.Write([]byte(response))
}

// sessionServiceHandler handles the SessionService resource
func sessionServiceHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
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
	w.Header().Set("Content-Type", "application/json")

	accountService := models.NewAccountService()
	json.NewEncoder(w).Encode(accountService)
}

// accountsHandler handles the accounts collection
func accountsHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)

	switch r.Method {
	case "GET":
		handleGetAccounts(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetAccounts returns the accounts collection
func handleGetAccounts(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Content-Type", "application/json")

	accounts := models.NewManagerAccountCollection()
	json.NewEncoder(w).Encode(accounts)
}

// systemsHandler handles the computer systems collection
func systemsHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Content-Type", "application/json")

	systems := models.NewComputerSystemCollection()
	json.NewEncoder(w).Encode(systems)
}

// systemHandler handles individual computer system resources
func systemHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Content-Type", "application/json")

	// Extract system ID from URL path
	path := r.URL.Path
	id := path[len("/redfish/v1/Systems/"):]

	system := models.NewComputerSystem(id)
	json.NewEncoder(w).Encode(system)
}

// chassisHandler handles the chassis collection
func chassisHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Content-Type", "application/json")

	chassis := models.NewChassisCollection()
	json.NewEncoder(w).Encode(chassis)
}

// chassisItemHandler handles individual chassis resources
func chassisItemHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Content-Type", "application/json")

	// Extract chassis ID from URL path
	path := r.URL.Path
	id := path[len("/redfish/v1/Chassis/"):]

	chassis := models.NewChassis(id)
	json.NewEncoder(w).Encode(chassis)
}

// managersHandler handles the managers collection
func managersHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Content-Type", "application/json")

	managers := models.NewManagerCollection()
	json.NewEncoder(w).Encode(managers)
}

// managerHandler handles individual manager resources
func managerHandler(w http.ResponseWriter, r *http.Request) {
	setRedfishHeaders(w)
	w.Header().Set("Content-Type", "application/json")

	// Extract manager ID from URL path
	path := r.URL.Path
	id := path[len("/redfish/v1/Managers/"):]

	manager := models.NewManager(id)
	json.NewEncoder(w).Encode(manager)
}

// setRedfishHeaders sets common Redfish headers
func setRedfishHeaders(w http.ResponseWriter) {
	w.Header().Set("OData-Version", "4.0")
	w.Header().Set("Cache-Control", "no-cache")
}
