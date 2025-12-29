package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/dimas1q/dockslim/backend/internal/auth"
	"github.com/go-chi/chi/v5"
)

type Dependencies struct {
	AuthHandler     *AuthHandler
	AuthMiddleware  *auth.Middleware
	ProjectsHandler *ProjectsHandler
	AllowedOrigins  []string
}

func NewRouter(deps Dependencies) http.Handler {
	r := chi.NewRouter()

	if len(deps.AllowedOrigins) > 0 {
		r.Use(corsMiddleware(deps.AllowedOrigins))
	}
	r.Use(csrfMiddleware)

	r.Get("/health", healthHandler)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", deps.AuthHandler.Register)
			r.Post("/login", deps.AuthHandler.Login)
			r.Post("/logout", deps.AuthHandler.Logout)
		})

		r.Group(func(r chi.Router) {
			r.Use(deps.AuthMiddleware.Authenticate)
			r.Get("/me", deps.AuthHandler.Me)
			r.Route("/projects", func(r chi.Router) {
				r.Post("/", deps.ProjectsHandler.Create)
				r.Get("/", deps.ProjectsHandler.List)
				r.Get("/{id}", deps.ProjectsHandler.Get)
				r.Patch("/{id}", deps.ProjectsHandler.Update)
				r.Delete("/{id}", deps.ProjectsHandler.Delete)
			})
		})
	})

	return r
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]string{"status": "ok"}
	_ = json.NewEncoder(w).Encode(resp)
}
