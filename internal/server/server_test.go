package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/redfish-server/internal/config"
)

func TestHealthHandler(t *testing.T) {
	// Create a test server
	mux := http.NewServeMux()
	setupRoutes(mux)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Find the handler
	handler := mux
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	expected := `{"status": "ok", "service": "redfish-server"}`
	if w.Body.String() != expected {
		t.Errorf("Expected body %s, got %s", expected, w.Body.String())
	}

	// Check headers
	if w.Header().Get("OData-Version") != "4.0" {
		t.Errorf("Expected OData-Version header to be 4.0")
	}
}

func TestServiceRootHandler(t *testing.T) {
	// Create a test server
	mux := http.NewServeMux()
	setupRoutes(mux)

	req := httptest.NewRequest("GET", "/redfish/v1/", nil)
	w := httptest.NewRecorder()

	handler := mux
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check content type
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type to be application/json")
	}

	// Check OData headers
	if w.Header().Get("OData-Version") != "4.0" {
		t.Errorf("Expected OData-Version header to be 4.0")
	}
}

func TestServerCreation(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Address:      ":8443",
			ReadTimeout:  30,
			WriteTimeout: 30,
		},
		TLS: config.TLSConfig{
			Enabled:  false, // Disable for test
			CertFile: "certs/server.crt",
			KeyFile:  "certs/server.key",
		},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if server == nil {
		t.Fatal("Server is nil")
	}

	if server.config != cfg {
		t.Error("Server config not set correctly")
	}
}
