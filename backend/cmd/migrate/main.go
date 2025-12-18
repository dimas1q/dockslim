package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/dimas1q/dockslim/backend/internal/config"
	"github.com/dimas1q/dockslim/backend/internal/migrate"
)

func main() {
	migrationsPath := flag.String("path", "", "path to migrations directory")
	flag.Parse()

	cfg := config.Load()
	path := *migrationsPath
	if path == "" {
		path = cfg.MigrationsPath
	}

	resolvedPath, err := resolveMigrationsPath(path)
	if err != nil {
		log.Fatalf("failed to resolve migrations path: %v", err)
	}

	if cfg.PostgresDSN == "" {
		log.Println("POSTGRES_DSN is required")
		os.Exit(1)
	}

	runner, err := migrate.NewRunner(cfg.PostgresDSN, resolvedPath)
	if err != nil {
		log.Fatalf("failed to create migration runner: %v", err)
	}
	defer runner.Close()

	if err := runner.Up(); err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
	}

	log.Println("migrations applied successfully")
}

func resolveMigrationsPath(preferred string) (string, error) {
	defaultPath := "backend/migrations"
	path := preferred
	if path == "" {
		path = defaultPath
	}

	resolve := func(p string) (string, bool) {
		if filepath.IsAbs(p) {
			if _, err := os.Stat(p); err == nil {
				return p, true
			}
			return "", false
		}
		if _, err := os.Stat(p); err == nil {
			abs, err := filepath.Abs(p)
			if err != nil {
				return p, true
			}
			return abs, true
		}
		return "", false
	}

	if resolved, ok := resolve(path); ok {
		log.Printf("using migrations path: %s", resolved)
		return resolved, nil
	}

	if path != defaultPath {
		if resolved, ok := resolve(defaultPath); ok {
			log.Printf("using migrations path: %s", resolved)
			return resolved, nil
		}
	}

	return "", fmt.Errorf("migrations path not found: %s", path)
}
