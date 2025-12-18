package config

import "os"

// Config holds configuration for the analyzer worker.
type Config struct {
	PostgresDSN string
	RedisAddr   string
}

// Load reads configuration from environment variables with sensible defaults for development.
func Load() Config {
	return Config{
		PostgresDSN: getEnv("ANALYZER_POSTGRES_DSN", "postgres://dockslim:dockslim@localhost:5432/dockslim?sslmode=disable"),
		RedisAddr:   getEnv("ANALYZER_REDIS_ADDR", "localhost:6379"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
