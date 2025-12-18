package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dimas1q/dockslim/backend/internal/config"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.Load()

	router := chi.NewRouter()
	router.Get("/health", healthHandler)

	log.Printf("Starting backend API on %s", cfg.Addr())
	if err := http.ListenAndServe(cfg.Addr(), router); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]string{"status": "ok"}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
	}
}
