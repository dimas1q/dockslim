package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/dimas1q/dockslim/backend/internal/auth"
	"github.com/dimas1q/dockslim/backend/internal/config"
	"github.com/dimas1q/dockslim/backend/internal/db"
	"github.com/dimas1q/dockslim/backend/internal/httpapi"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	database, err := db.Connect(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

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
