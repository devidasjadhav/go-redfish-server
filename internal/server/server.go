package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/user/redfish-server/internal/config"
	"github.com/user/redfish-server/internal/middleware"
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

	// Redfish root endpoint
	mux.HandleFunc("/redfish/v1/", serviceRootHandler)

	// Redfish metadata endpoint
	mux.HandleFunc("/redfish/v1/$metadata", metadataHandler)

	// Redfish OData service document
	mux.HandleFunc("/redfish/v1/odata", odataHandler)
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

	response := `{
		"@odata.context": "/redfish/v1/$metadata#ServiceRoot.ServiceRoot",
		"@odata.id": "/redfish/v1/",
		"@odata.type": "#ServiceRoot.v1_15_0.ServiceRoot",
		"Id": "RootService",
		"Name": "Root Service",
		"RedfishVersion": "1.15.0",
		"UUID": "00000000-0000-0000-0000-000000000000",
		"Systems": {
			"@odata.id": "/redfish/v1/Systems"
		},
		"Chassis": {
			"@odata.id": "/redfish/v1/Chassis"
		},
		"Managers": {
			"@odata.id": "/redfish/v1/Managers"
		},
		"Tasks": {
			"@odata.id": "/redfish/v1/TaskService"
		},
		"SessionService": {
			"@odata.id": "/redfish/v1/SessionService"
		},
		"AccountService": {
			"@odata.id": "/redfish/v1/AccountService"
		},
		"EventService": {
			"@odata.id": "/redfish/v1/EventService"
		},
		"Registries": {
			"@odata.id": "/redfish/v1/Registries"
		},
		"JsonSchemas": {
			"@odata.id": "/redfish/v1/JsonSchemas"
		},
		"Links": {
			"Sessions": {
				"@odata.id": "/redfish/v1/SessionService/Sessions"
			}
		}
	}`

	w.Write([]byte(response))
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

// setRedfishHeaders sets common Redfish headers
func setRedfishHeaders(w http.ResponseWriter) {
	w.Header().Set("OData-Version", "4.0")
	w.Header().Set("Cache-Control", "no-cache")
}
