package main

import (
	"flag"
	"log"
	"os"

	"github.com/dimas1q/dockslim/backend/internal/config"
	"github.com/dimas1q/dockslim/backend/internal/migrate"
)

func main() {
	migrationsPath := flag.String("path", "./backend/migrations", "path to migrations directory")
	flag.Parse()

	cfg := config.Load()

	if cfg.PostgresDSN == "" {
		log.Println("POSTGRES_DSN is required")
		os.Exit(1)
	}

	runner, err := migrate.NewRunner(cfg.PostgresDSN, *migrationsPath)
	if err != nil {
		log.Fatalf("failed to create migration runner: %v", err)
	}
	defer runner.Close()

	if err := runner.Up(); err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
	}

	log.Println("migrations applied successfully")
}
