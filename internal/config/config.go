package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the Redfish server
type Config struct {
	Server ServerConfig
	TLS    TLSConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Address      string
	ReadTimeout  int // seconds
	WriteTimeout int // seconds
}

// TLSConfig holds TLS-specific configuration
type TLSConfig struct {
	Enabled  bool
	CertFile string
	KeyFile  string
}

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Address:      getEnv("SERVER_ADDRESS", ":8443"),
			ReadTimeout:  getEnvAsInt("SERVER_READ_TIMEOUT", 30),
			WriteTimeout: getEnvAsInt("SERVER_WRITE_TIMEOUT", 30),
		},
		TLS: TLSConfig{
			Enabled:  getEnvAsBool("TLS_ENABLED", true),
			CertFile: getEnv("TLS_CERT_FILE", "certs/server.crt"),
			KeyFile:  getEnv("TLS_KEY_FILE", "certs/server.key"),
		},
	}

	return cfg, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as int or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool gets an environment variable as bool or returns a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Address == "" {
		return fmt.Errorf("server address cannot be empty")
	}
	if c.TLS.Enabled {
		if c.TLS.CertFile == "" || c.TLS.KeyFile == "" {
			return fmt.Errorf("TLS cert and key files must be specified when TLS is enabled")
		}
	}
	return nil
}
