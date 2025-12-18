package config

import "os"

// Config holds configuration values for the backend API.
type Config struct {
	HTTPPort    string
	PostgresDSN string
}

// Load reads configuration from environment variables with sane defaults for development.
func Load() Config {
	cfg := Config{
		HTTPPort:    getEnv("BACKEND_HTTP_PORT", "8080"),
		PostgresDSN: getEnv("POSTGRES_DSN", "postgres://dockslim:dockslim@localhost:5432/dockslim?sslmode=disable"),
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
