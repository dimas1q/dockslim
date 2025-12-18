package config

import "os"

// Config holds configuration values for the backend API.
type Config struct {
	HTTPPort       string
	PostgresDSN    string
	AutoMigrate    bool
	MigrationsPath string
}

// Load reads configuration from environment variables with sane defaults for development.
func Load() Config {
	cfg := Config{
		HTTPPort:       getEnv("BACKEND_HTTP_PORT", "8080"),
		PostgresDSN:    getEnv("POSTGRES_DSN", "postgres://dockslim:dockslim@localhost:5432/dockslim?sslmode=disable"),
		AutoMigrate:    getEnvAsBool("AUTO_MIGRATE", true),
		MigrationsPath: getEnv("MIGRATIONS_PATH", "backend/migrations"),
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
