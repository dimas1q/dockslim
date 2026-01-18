package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/dimas1q/dockslim/backend/internal/auth"
	"github.com/go-chi/chi/v5"
)

type Dependencies struct {
	AuthHandler       *AuthHandler
	AuthMiddleware    *auth.Middleware
	ProjectsHandler   *ProjectsHandler
	RegistriesHandler *RegistriesHandler
	AnalysesHandler   *AnalysesHandler
	AllowedOrigins    []string
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
				r.Route("/{projectId}/registries", func(r chi.Router) {
					r.Get("/", deps.RegistriesHandler.List)
					r.Post("/", deps.RegistriesHandler.Create)
					r.Delete("/{registryId}", deps.RegistriesHandler.Delete)
				})
				r.Route("/{projectId}/analyses", func(r chi.Router) {
					r.Get("/", deps.AnalysesHandler.List)
					r.Post("/", deps.AnalysesHandler.Create)
					r.Get("/{analysisId}", deps.AnalysesHandler.Get)
					r.Delete("/{analysisId}", deps.AnalysesHandler.Delete)
					r.Post("/{analysisId}/rerun", deps.AnalysesHandler.Rerun)
				})
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
