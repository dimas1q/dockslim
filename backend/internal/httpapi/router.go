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
	BudgetsHandler    *BudgetsHandler
	CITokensHandler   *CITokensHandler
	CIHandler         *CIHandler
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
				r.Route("/{projectId}/budgets", func(r chi.Router) {
					r.Get("/", deps.BudgetsHandler.List)
					r.Put("/default", deps.BudgetsHandler.UpsertDefault)
					r.Post("/overrides", deps.BudgetsHandler.CreateOverride)
					r.Patch("/overrides/{budgetId}", deps.BudgetsHandler.UpdateOverride)
					r.Delete("/overrides/{budgetId}", deps.BudgetsHandler.DeleteOverride)
				})
				r.Route("/{projectId}/registries", func(r chi.Router) {
					r.Get("/", deps.RegistriesHandler.List)
					r.Post("/", deps.RegistriesHandler.Create)
					r.Patch("/{registryId}", deps.RegistriesHandler.Update)
					r.Delete("/{registryId}", deps.RegistriesHandler.Delete)
				})
				if deps.CITokensHandler != nil {
					r.Route("/{projectId}/ci-tokens", func(r chi.Router) {
						r.Post("/", deps.CITokensHandler.Create)
						r.Get("/", deps.CITokensHandler.List)
						r.Post("/{tokenId}/revoke", deps.CITokensHandler.Revoke)
					})
				}
				r.Route("/{projectId}/analyses", func(r chi.Router) {
					r.Get("/", deps.AnalysesHandler.List)
					r.Post("/", deps.AnalysesHandler.Create)
					r.Get("/compare", deps.AnalysesHandler.Compare)
					r.Get("/{analysisId}", deps.AnalysesHandler.Get)
					r.Delete("/{analysisId}", deps.AnalysesHandler.Delete)
					r.Post("/{analysisId}/rerun", deps.AnalysesHandler.Rerun)
				})
			})
		})

		if deps.CIHandler != nil {
			r.Route("/ci", func(r chi.Router) {
				r.Use(deps.AuthMiddleware.AuthenticateUserOrCIToken)
				r.Post("/reports/image", deps.CIHandler.CreateAnalysisReport)
				r.Post("/reports/compare", deps.CIHandler.CompareReport)
				r.Post("/comments", deps.CIHandler.PostComment)
			})
		}
	})

	return r
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]string{"status": "ok"}
	_ = json.NewEncoder(w).Encode(resp)
}
