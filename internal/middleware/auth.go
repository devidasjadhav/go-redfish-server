package middleware

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/user/redfish-server/internal/auth"
)

// AuthMiddleware handles authentication for protected endpoints
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if authentication is required for this endpoint
		if !requiresAuth(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Try Basic Authentication first
		if username, password, ok := r.BasicAuth(); ok {
			if auth.ValidateBasicAuth(username, password) {
				// Set user context for later use
				ctx := auth.SetUserContext(r.Context(), username, "Basic")
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
				return
			}
		}

		// Try Session Authentication (X-Auth-Token header)
		if token := r.Header.Get("X-Auth-Token"); token != "" {
			if username, ok := auth.ValidateSessionToken(token); ok {
				ctx := auth.SetUserContext(r.Context(), username, "Session")
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
				return
			}
		}

		// Authentication failed
		w.Header().Set("WWW-Authenticate", `Basic realm="Redfish Service"`)
		http.Error(w, `{"error": {"code": "Base.1.0.InsufficientPrivilege", "message": "Authentication required"}}`, http.StatusUnauthorized)
	})
}

// requiresAuth determines if authentication is required for the given path
func requiresAuth(path string) bool {
	// Public endpoints that don't require authentication
	publicPaths := []string{
		"/health",
		"/redfish/v1/",
		"/redfish/v1/$metadata",
		"/redfish/v1/odata",
		"/redfish/v1/SessionService",
		"/redfish/v1/SessionService/Sessions",
	}

	for _, publicPath := range publicPaths {
		if path == publicPath {
			return false
		}
	}

	// All other endpoints require authentication
	return true
}

// BasicAuthDecode decodes a base64 encoded username:password string
func BasicAuthDecode(encoded string) (username, password string, ok bool) {
	c, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", false
	}

	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return "", "", false
	}

	return cs[:s], cs[s+1:], true
}
