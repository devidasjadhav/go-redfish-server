package auth

import (
	"testing"
)

func TestValidateBasicAuth(t *testing.T) {
	auth := NewAuthService()

	// Test valid credentials
	if !auth.ValidateBasicAuth("admin", "password") {
		t.Error("Valid admin credentials should be accepted")
	}

	if !auth.ValidateBasicAuth("operator", "password") {
		t.Error("Valid operator credentials should be accepted")
	}

	// Test invalid credentials
	if auth.ValidateBasicAuth("admin", "wrongpassword") {
		t.Error("Invalid password should be rejected")
	}

	if auth.ValidateBasicAuth("nonexistent", "password") {
		t.Error("Non-existent user should be rejected")
	}
}

func TestSessionManagement(t *testing.T) {
	auth := NewAuthService()

	// Create a session
	token, err := auth.CreateSession("admin")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if token == "" {
		t.Error("Session token should not be empty")
	}

	// Validate the session
	username, valid := auth.ValidateSessionToken(token)
	if !valid {
		t.Error("Valid session token should be accepted")
	}

	if username != "admin" {
		t.Errorf("Expected username 'admin', got '%s'", username)
	}

	// Test invalid token
	_, valid = auth.ValidateSessionToken("invalid-token")
	if valid {
		t.Error("Invalid token should be rejected")
	}

	// Delete session
	auth.DeleteSession(token)
	_, valid = auth.ValidateSessionToken(token)
	if valid {
		t.Error("Deleted session should be invalid")
	}
}

func TestGetUser(t *testing.T) {
	auth := NewAuthService()

	user, exists := auth.GetUser("admin")
	if !exists {
		t.Error("Admin user should exist")
	}

	if user.Username != "admin" {
		t.Errorf("Expected username 'admin', got '%s'", user.Username)
	}

	if user.Role != "Administrator" {
		t.Errorf("Expected role 'Administrator', got '%s'", user.Role)
	}

	_, exists = auth.GetUser("nonexistent")
	if exists {
		t.Error("Non-existent user should not exist")
	}
}

func TestListUsers(t *testing.T) {
	auth := NewAuthService()

	users := auth.ListUsers()
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	usernames := make(map[string]bool)
	for _, user := range users {
		usernames[user.Username] = true
	}

	if !usernames["admin"] || !usernames["operator"] {
		t.Error("Expected admin and operator users")
	}
}

func TestGlobalAuthService(t *testing.T) {
	auth1 := GetAuthService()
	auth2 := GetAuthService()

	if auth1 != auth2 {
		t.Error("Global auth service should be singleton")
	}

	// Test that it works
	if !auth1.ValidateBasicAuth("admin", "password") {
		t.Error("Global auth service should validate credentials")
	}
}
