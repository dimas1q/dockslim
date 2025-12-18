package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/dimas1q/dockslim/backend/internal/auth"
	"github.com/go-chi/chi/v5"
)

type Dependencies struct {
	AuthHandler    *AuthHandler
	AuthMiddleware *auth.Middleware
}

func NewRouter(deps Dependencies) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", healthHandler)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", deps.AuthHandler.Register)
			r.Post("/login", deps.AuthHandler.Login)
		})

		r.Group(func(r chi.Router) {
			r.Use(deps.AuthMiddleware.Authenticate)
			r.Get("/me", deps.AuthHandler.Me)
		})
	})

	return r
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]string{"status": "ok"}
	_ = json.NewEncoder(w).Encode(resp)
}
