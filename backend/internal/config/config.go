package config

import (
	"fmt"
	"os"
)

// Config holds configuration values for the backend API.
type Config struct {
	HTTPPort    string
	PostgresDSN string
	RedisAddr   string
}

// Load reads configuration from environment variables with sane defaults for development.
func Load() Config {
	cfg := Config{
		HTTPPort:    getEnv("BACKEND_HTTP_PORT", "8080"),
		PostgresDSN: getEnv("BACKEND_POSTGRES_DSN", "postgres://dockslim:dockslim@localhost:5432/dockslim?sslmode=disable"),
		RedisAddr:   getEnv("BACKEND_REDIS_ADDR", "localhost:6379"),
	}

	return cfg
}

// Addr returns the full HTTP listen address.
func (c Config) Addr() string {
	return fmt.Sprintf(":%s", c.HTTPPort)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
