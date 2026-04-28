package core

import (
	"os"
	"strconv"
	"strings"
	"sync"
)

// Config provides centralized, thread-safe access to application configuration.
// All configuration values are loaded from environment variables (via .env file).
type Config struct {
	mu sync.RWMutex
}

var configInstance *Config
var configOnce sync.Once

// GetConfig returns the global Config singleton
func GetConfig() *Config {
	configOnce.Do(func() {
		configInstance = &Config{}
	})
	return configInstance
}

// String reads a string environment variable, returning defaultVal if unset
func (c *Config) String(key string, defaultVal ...string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if val := os.Getenv(key); val != "" {
		return val
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return ""
}

// Int reads an integer environment variable, returning defaultVal if unset or invalid
func (c *Config) Int(key string, defaultVal ...int) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val := os.Getenv(key)
	if val == "" {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}
	return i
}

// Bool reads a boolean environment variable (true/1/yes vs false/0/no), returning defaultVal if unset
func (c *Config) Bool(key string, defaultVal ...bool) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val := strings.ToLower(os.Getenv(key))
	if val == "" {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return false
	}
	return val == "true" || val == "1" || val == "yes"
}

// DatabaseConfig returns database connection parameters
func (c *Config) DatabaseConfig() (host, port, user, password, dbname, sslmode string) {
	return c.String("DB_HOST", "localhost"),
		c.String("DB_PORT", "5432"),
		c.String("DB_USER", "postgres"),
		c.String("DB_PASSWORD", "postgres"),
		c.String("DB_NAME", "purecore"),
		c.String("DB_SSLMODE", "disable")
}

// AppPort returns the backend server port
func (c *Config) AppPort() string {
	return c.String("BACKEND_PORT", "9002")
}

// AppDebug returns whether debug mode is enabled
func (c *Config) AppDebug() bool {
	return c.Bool("APP_DEBUG", false)
}

// AdminRoutePrefix returns the admin route prefix
func (c *Config) AdminRoutePrefix() string {
	return c.String("ADMIN_ROUTE_PREFIX", "admin")
}

// IsProduction returns true if the app is running in production mode
func (c *Config) IsProduction() bool {
	return c.String("APP_ENV") == "production"
}

// IsDevelopment returns true if the app is running in development mode
func (c *Config) IsDevelopment() bool {
	return c.String("APP_ENV", "development") != "production"
}
