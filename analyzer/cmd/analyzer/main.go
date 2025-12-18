package main

import (
	"log"

	"github.com/dimas1q/dockslim/analyzer/internal/config"
)

func main() {
	cfg := config.Load()

	log.Printf("Analyzer worker started")
	log.Printf("Using Postgres DSN: %s", cfg.PostgresDSN)
	log.Printf("Using Redis address: %s", cfg.RedisAddr)
}
