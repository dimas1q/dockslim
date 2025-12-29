package config

import (
	"os"
	"strings"
)

// Config holds configuration values for the backend API.
type Config struct {
	HTTPPort           string
	PostgresDSN        string
	AutoMigrate        bool
	MigrationsPath     string
	CORSAllowedOrigins []string
	CookieSecure       bool
}

// Load reads configuration from environment variables with sane defaults for development.
func Load() Config {
	cfg := Config{
		HTTPPort:           getEnv("BACKEND_HTTP_PORT", "8080"),
		PostgresDSN:        getEnv("POSTGRES_DSN", "postgres://dockslim:dockslim@localhost:5432/dockslim?sslmode=disable"),
		AutoMigrate:        getEnvAsBool("AUTO_MIGRATE", true),
		MigrationsPath:     getEnv("MIGRATIONS_PATH", "backend/migrations"),
		CORSAllowedOrigins: getEnvAsList("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173", "http://127.0.0.1:5173"}),
		CookieSecure:       getEnvAsBool("COOKIE_SECURE", false),
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	if val := os.Getenv(key); val != "" {
		if val == "1" || val == "true" || val == "TRUE" || val == "True" {
			return true
		}
		if val == "0" || val == "false" || val == "FALSE" || val == "False" {
			return false
		}
	}
	return fallback
}

func getEnvAsList(key string, fallback []string) []string {
	if val := os.Getenv(key); val != "" {
		parts := strings.Split(val, ",")
		var values []string
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed == "" {
				continue
			}
			values = append(values, trimmed)
		}
		if len(values) > 0 {
			return values
		}
	}
	return fallback
}
