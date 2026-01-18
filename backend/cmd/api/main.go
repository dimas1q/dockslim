package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dimas1q/dockslim/backend/internal/analyses"
	"github.com/dimas1q/dockslim/backend/internal/auth"
	"github.com/dimas1q/dockslim/backend/internal/config"
	"github.com/dimas1q/dockslim/backend/internal/db"
	"github.com/dimas1q/dockslim/backend/internal/httpapi"
	"github.com/dimas1q/dockslim/backend/internal/migrate"
	"github.com/dimas1q/dockslim/backend/internal/projects"
	"github.com/dimas1q/dockslim/backend/internal/registries"
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

	authRepo := auth.NewRepository(database)
	projectRepo := projects.NewRepository(database)
	registryRepo := registries.NewRepository(database)
	analysisRepo := analyses.NewRepository(database)
	tokenManager, err := auth.NewTokenManager(ctx, authRepo, auth.DefaultAccessTokenTTL)
	if err != nil {
		log.Fatalf("failed to initialize token manager: %v", err)
	}
	authService := auth.NewService(authRepo, tokenManager)
	projectService := projects.NewService(projectRepo)
	activeKey, err := registries.EnsureActiveKey(ctx, registryRepo)
	if err != nil {
		log.Fatalf("failed to load registry encryption key: %v", err)
	}
	registryService := registries.NewService(registryRepo, projectRepo, activeKey)
	analysisService := analyses.NewService(analysisRepo, projectRepo, registryRepo)
	middleware := auth.NewMiddleware(tokenManager, authRepo)
	cookieSameSite, err := parseSameSite(cfg.CookieSameSite)
	if err != nil {
		log.Fatalf("invalid COOKIE_SAMESITE: %v", err)
	}
	if cookieSameSite == http.SameSiteNoneMode && !cfg.CookieSecure {
		log.Fatalf("COOKIE_SAMESITE=none requires COOKIE_SECURE=true")
	}
	authHandler := httpapi.NewAuthHandler(authService, auth.DefaultAccessTokenTTL, httpapi.CookieConfig{
		SameSite: cookieSameSite,
		Secure:   cfg.CookieSecure,
		Domain:   cfg.CookieDomain,
		Path:     cfg.CookiePath,
	})
	projectsHandler := httpapi.NewProjectsHandler(projectService)
	registriesHandler := httpapi.NewRegistriesHandler(registryService)
	analysesHandler := httpapi.NewAnalysesHandler(analysisService)

	router := httpapi.NewRouter(httpapi.Dependencies{
		AuthHandler:       authHandler,
		AuthMiddleware:    middleware,
		ProjectsHandler:   projectsHandler,
		RegistriesHandler: registriesHandler,
		AnalysesHandler:   analysesHandler,
		AllowedOrigins:    cfg.CORSAllowedOrigins,
	})

	addr := fmt.Sprintf(":%s", cfg.HTTPPort)
	log.Printf("Starting backend API on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func parseSameSite(value string) (http.SameSite, error) {
	switch value {
	case "lax":
		return http.SameSiteLaxMode, nil
	case "strict":
		return http.SameSiteStrictMode, nil
	case "none":
		return http.SameSiteNoneMode, nil
	default:
		return http.SameSiteDefaultMode, fmt.Errorf("unsupported value %q", value)
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
