package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dimas1q/dockslim/backend/internal/auth"
	"github.com/dimas1q/dockslim/backend/internal/config"
	"github.com/dimas1q/dockslim/backend/internal/db"
	"github.com/dimas1q/dockslim/backend/internal/httpapi"
	"github.com/dimas1q/dockslim/backend/internal/migrate"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	database, err := db.ConnectWithRetry(ctx, cfg.PostgresDSN, 20, 500*time.Millisecond)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	if cfg.AutoMigrate {
		migrationsPath, err := resolveMigrationsPath(cfg.MigrationsPath)
		if err != nil {
			log.Fatalf("failed to resolve migrations path: %v", err)
		}

		runner, err := migrate.NewRunner(cfg.PostgresDSN, migrationsPath)
		if err != nil {
			log.Fatalf("failed to create migration runner: %v", err)
		}
		defer runner.Close()

		migrationCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		if err := runner.UpWithLock(migrationCtx, "dockslim_migrations"); err != nil {
			log.Fatalf("failed to apply migrations: %v", err)
		}
	}

	repo := auth.NewRepository(database)
	tokenManager, err := auth.NewTokenManager(ctx, repo, auth.DefaultAccessTokenTTL)
	if err != nil {
		log.Fatalf("failed to initialize token manager: %v", err)
	}
	service := auth.NewService(repo, tokenManager)
	middleware := auth.NewMiddleware(tokenManager, repo)
	handler := httpapi.NewAuthHandler(service, auth.DefaultAccessTokenTTL)

	router := httpapi.NewRouter(httpapi.Dependencies{
		AuthHandler:    handler,
		AuthMiddleware: middleware,
	})

	addr := fmt.Sprintf(":%s", cfg.HTTPPort)
	log.Printf("Starting backend API on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
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
