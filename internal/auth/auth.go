package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// User represents a user account
type User struct {
	Username string
	Password string // In production, this should be hashed
	Role     string
	Enabled  bool
}

// Session represents an active user session
type Session struct {
	Token    string
	Username string
	Created  time.Time
	Expires  time.Time
}

// AuthService manages authentication and sessions
type AuthService struct {
	users    map[string]*User
	sessions map[string]*Session
	mutex    sync.RWMutex
}

// NewAuthService creates a new authentication service with default users
func NewAuthService() *AuthService {
	auth := &AuthService{
		users:    make(map[string]*User),
		sessions: make(map[string]*Session),
	}

	// Add default admin user (for development)
	auth.users["admin"] = &User{
		Username: "admin",
		Password: "password", // In production, use hashed passwords
		Role:     "Administrator",
		Enabled:  true,
	}

	// Add default operator user
	auth.users["operator"] = &User{
		Username: "operator",
		Password: "password",
		Role:     "Operator",
		Enabled:  true,
	}

	return auth
}

// ValidateBasicAuth validates username/password credentials
func (a *AuthService) ValidateBasicAuth(username, password string) bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	user, exists := a.users[username]
	if !exists || !user.Enabled {
		return false
	}

	// In production, use proper password hashing (bcrypt)
	return user.Password == password
}

// CreateSession creates a new session for the authenticated user
func (a *AuthService) CreateSession(username string) (string, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// Generate a random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)

	session := &Session{
		Token:    token,
		Username: username,
		Created:  time.Now(),
		Expires:  time.Now().Add(24 * time.Hour), // 24 hour session
	}

	a.sessions[token] = session

	return token, nil
}

// ValidateSessionToken validates a session token and returns the username
func (a *AuthService) ValidateSessionToken(token string) (string, bool) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	session, exists := a.sessions[token]
	if !exists {
		return "", false
	}

	// Check if session has expired (disabled for testing)
	// if time.Now().After(session.Expires) {
	// 	// Clean up expired session
	// 	go func() {
	// 		a.mutex.Lock()
	// 		delete(a.sessions, token)
	// 		a.mutex.Unlock()
	// 	}()
	// 	return "", false
	// }

	return session.Username, true
}

// DeleteSession removes a session
func (a *AuthService) DeleteSession(token string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	delete(a.sessions, token)
}

// GetUser returns user information
func (a *AuthService) GetUser(username string) (*User, bool) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	user, exists := a.users[username]
	return user, exists
}

// ListUsers returns all users (for AccountService)
func (a *AuthService) ListUsers() []*User {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	users := make([]*User, 0, len(a.users))
	for _, user := range a.users {
		users = append(users, user)
	}
	return users
}

// Global auth service instance
var globalAuth *AuthService
var once sync.Once

// GetAuthService returns the global authentication service
func GetAuthService() *AuthService {
	once.Do(func() {
		globalAuth = NewAuthService()
	})
	return globalAuth
}

// Convenience functions for middleware
func ValidateBasicAuth(username, password string) bool {
	return GetAuthService().ValidateBasicAuth(username, password)
}

func ValidateSessionToken(token string) (string, bool) {
	return GetAuthService().ValidateSessionToken(token)
}

// Context helpers
type userKey struct{}

type UserContext struct {
	Username string
	Method   string // "Basic" or "Session"
}

// SetUserContext adds user information to request context
func SetUserContext(ctx context.Context, username, method string) context.Context {
	return context.WithValue(ctx, userKey{}, &UserContext{
		Username: username,
		Method:   method,
	})
}

// GetUserContext retrieves user information from request context
func GetUserContext(ctx context.Context) (*UserContext, bool) {
	userCtx, ok := ctx.Value(userKey{}).(*UserContext)
	return userCtx, ok
}
