package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/user/redfish-server/internal/config"
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

	httpServer := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      mux,
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
	if s.config.TLS.Enabled {
		return s.httpServer.ListenAndServeTLS("", "")
	}
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
}

// healthHandler handles health check requests
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}

// serviceRootHandler handles the Redfish service root
func serviceRootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("OData-Version", "4.0")

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
